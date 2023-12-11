package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	sdkErrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	postgresql "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/connectivity"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

type PostgresqlService struct {
	client *connectivity.TencentCloudClient
}

func (me *PostgresqlService) CreatePostgresqlInstance(
	ctx context.Context,
	name, dbVersion, dbMajorVersion, dbKernelVersion, chargeType, specCode string, autoRenewFlag, projectId, period int, subnetId, vpcId, zone string,
	securityGroups []string,
	storage int,
	username, password, charset string,
	dbNodeSet []*postgresql.DBNode,
	needSupportTde int, kmsKeyId, kmsRegion string, autoVoucher int, voucherIds []*string,
) (instanceId string, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewCreateInstancesRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.Name = &name
	request.DBVersion = &dbVersion
	if dbMajorVersion != "" {
		request.DBMajorVersion = helper.String(dbMajorVersion)
	}
	if dbKernelVersion != "" {
		request.DBKernelVersion = helper.String(dbKernelVersion)
	}
	request.InstanceChargeType = &chargeType
	request.SpecCode = &specCode
	request.AutoRenewFlag = helper.IntInt64(autoRenewFlag)
	request.ProjectId = helper.Int64(int64(projectId))
	request.Period = helper.Int64Uint64(int64(period))
	request.SubnetId = &subnetId
	request.VpcId = &vpcId
	request.Storage = helper.IntUint64(storage)
	request.Zone = &zone
	request.InstanceCount = helper.Int64Uint64(1)
	request.AdminName = &username
	request.AdminPassword = &password
	request.Charset = &charset

	if needSupportTde == 1 {
		request.NeedSupportTDE = helper.IntUint64(1)
		if kmsKeyId != "" {
			request.KMSKeyId = helper.String(kmsKeyId)
		}
		if kmsRegion != "" {
			request.KMSRegion = helper.String(kmsRegion)
		}
	}

	if len(securityGroups) > 0 {
		request.SecurityGroupIds = make([]*string, 0, len(securityGroups))
		for v := range securityGroups {
			request.SecurityGroupIds = append(request.SecurityGroupIds, &securityGroups[v])
		}
	}

	if len(dbNodeSet) > 0 {
		request.DBNodeSet = dbNodeSet
	}

	if autoVoucher > 0 {
		request.AutoVoucher = helper.IntUint64(autoVoucher)
	}

	if len(voucherIds) > 0 {
		request.VoucherIds = voucherIds
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().CreateInstances(request)
	if err != nil {
		errRet = err
		return
	}
	if response == nil || response.Response == nil {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
	}
	if len(response.Response.DBInstanceIdSet) == 0 {
		if len(response.Response.DealNames) == 0 {
			errRet = errors.New("TencentCloud SDK returns empty postgresql ID and Deals")
			return
		}
		log.Printf("[WARN] No postgresql ID returns, requesting Deal Id")
		dealId := response.Response.DealNames[0]
		deals, err := me.DescribeOrders(ctx, []*string{dealId})
		if err != nil {
			errRet = err
			return
		}
		if len(deals) > 0 && len(deals[0].DBInstanceIdSet) > 0 {
			instanceId = *deals[0].DBInstanceIdSet[0]
		}
	} else if len(response.Response.DBInstanceIdSet) > 1 {
		errRet = errors.New("TencentCloud SDK returns more than one postgresql ID")
		return
	} else {
		instanceId = *response.Response.DBInstanceIdSet[0]
	}
	return
}

func (me *PostgresqlService) DescribeOrders(ctx context.Context, dealIds []*string) (deals []*postgresql.PgDeal, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeOrdersRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	request.DealNames = dealIds

	ratelimit.Check(request.GetAction())

	var response *postgresql.DescribeOrdersResponse
	err := resource.Retry(readRetryTimeout*5, func() *resource.RetryError {
		res, err := me.client.UsePostgresqlClient().DescribeOrders(request)
		if err != nil {
			return retryError(err)
		}
		if len(res.Response.Deals) == 0 {
			return resource.RetryableError(fmt.Errorf("waiting for deal return instance id"))
		}
		response = res
		return nil
	})

	if err != nil {
		errRet = err
		return
	}

	if response != nil {
		deals = response.Response.Deals
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) InitPostgresqlInstance(ctx context.Context, instanceId string, username string, password string, charset string) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewInitDBInstancesRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.Charset = &charset
	request.AdminName = &username
	request.AdminPassword = &password
	request.DBInstanceIdSet = []*string{&instanceId}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().InitDBInstances(request)
	if err != nil {
		return err
	}
	if response == nil || response.Response == nil {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
	}

	return
}

func (me *PostgresqlService) DescribeSpecinfos(ctx context.Context, zone string) (specCodeList []*postgresql.SpecItemInfo, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeProductConfigRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.Zone = &zone

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeProductConfig(request)
	if err != nil {
		errRet = err
		return
	}
	if response == nil || response.Response == nil || len(response.Response.SpecInfoList) == 0 {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		return
	}

	specCodeList = response.Response.SpecInfoList[0].SpecItemInfoList
	return
}

func (me *PostgresqlService) ModifyBackupPlan(ctx context.Context, request *postgresql.ModifyBackupPlanRequest) (errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().ModifyBackupPlan(request)

	if err != nil {
		errRet = err
		return
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) DescribeBackupPlans(ctx context.Context, request *postgresql.DescribeBackupPlansRequest) (result []*postgresql.BackupPlan, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeBackupPlans(request)

	if err != nil {
		errRet = err
		return
	}

	if len(response.Response.Plans) > 0 {
		result = response.Response.Plans
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) DescribeDBXlogs(ctx context.Context, request *postgresql.DescribeDBXlogsRequest) (xlogs []*postgresql.Xlog, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	offset := 0
	request.Limit = helper.IntInt64(100)
	request.Offset = helper.IntInt64(offset)
	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeDBXlogs(request)

	if err != nil {
		errRet = err
		return
	}

	if len(response.Response.XlogList) > 0 {
		xlogs = append(xlogs, response.Response.XlogList...)
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) ModifyPublicService(ctx context.Context, openInternet bool, instanceId string) (errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s-open:%t] fail,reason[%s]", logId, "modifyInternetService", openInternet, errRet.Error())
		}
	}()

	if openInternet {
		request := postgresql.NewOpenDBExtranetAccessRequest()
		request.DBInstanceId = &instanceId
		ratelimit.Check(request.GetAction())

		var response *postgresql.OpenDBExtranetAccessResponse
		err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			resp, err := me.client.UsePostgresqlClient().OpenDBExtranetAccess(request)
			if err != nil {
				return retryError(err, postgresql.OPERATIONDENIED_INSTANCESTATUSLIMITOPERROR)
			}
			response = resp
			return nil
		})

		if err != nil {
			errRet = err
			return
		}
		if response == nil || response.Response == nil {
			errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		}

		// Retry if status still persist after API invoked.
		startProgressRetries := 5

		// check open or not
		err = resource.Retry(3*readRetryTimeout, func() *resource.RetryError {
			instance, has, err := me.DescribePostgresqlInstanceById(ctx, instanceId)
			if err != nil {
				return retryError(err)
			}
			if !has {
				return resource.NonRetryableError(fmt.Errorf("check postgresql instance %s fail, instance is not exist", instanceId))
			}
			if len(instance.DBInstanceNetInfo) == 0 {
				return resource.NonRetryableError(fmt.Errorf("illegal net info of postgresql instance %s", instanceId))
			}
			for _, v := range instance.DBInstanceNetInfo {
				if *v.NetType != "public" {
					continue
				}
				if MatchAny(*v.Status, "opened", "2") {
					return nil
				}
				if MatchAny(*v.Status, "opening", "4") {
					startProgressRetries = 0
					return resource.RetryableError(fmt.Errorf("status %s, postgresql instance %s waiting", *v.Status, instanceId))
				}
				if startProgressRetries > 0 && MatchAny(*v.Status, "closed", "initing") {
					startProgressRetries -= 1
					return resource.RetryableError(fmt.Errorf("status still closed, retry remaining count: %d", startProgressRetries))
				}
				return resource.NonRetryableError(fmt.Errorf("status %s, postgresql instance %s open public service fail", *v.Status, instanceId))
			}
			// there is no public service yet
			return resource.RetryableError(fmt.Errorf("cannot find public status, postgresql instance %s watiting", instanceId))
		})
		if err != nil {
			return err
		}

	} else {
		request := postgresql.NewCloseDBExtranetAccessRequest()
		request.DBInstanceId = &instanceId
		ratelimit.Check(request.GetAction())

		var response *postgresql.CloseDBExtranetAccessResponse
		err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			resp, err := me.client.UsePostgresqlClient().CloseDBExtranetAccess(request)
			if err != nil {
				return retryError(err, postgresql.OPERATIONDENIED_INSTANCESTATUSLIMITOPERROR)
			}
			response = resp
			return nil
		})

		if err != nil {
			return err
		}
		if response == nil || response.Response == nil {
			errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		}
		startProgressRetries := 5
		// check close or not
		err = resource.Retry(3*readRetryTimeout, func() *resource.RetryError {
			instance, has, err := me.DescribePostgresqlInstanceById(ctx, instanceId)
			if err != nil {
				return retryError(err)
			}
			if !has {
				return resource.NonRetryableError(fmt.Errorf("check postgresql instance %s fail, instance is not exist", instanceId))
			}
			if len(instance.DBInstanceNetInfo) == 0 {
				return resource.NonRetryableError(fmt.Errorf("illegal net info of postgresql instance %s", instanceId))
			}
			for _, v := range instance.DBInstanceNetInfo {
				if *v.NetType != "public" {
					continue
				}
				if MatchAny(*v.Status, "closed", "3", "initing", "1") {
					return nil
				}
				if MatchAny(*v.Status, "closing", "4") {
					return resource.RetryableError(fmt.Errorf("status %s, postgresql instance %s waiting", *v.Status, instanceId))
				}
				if startProgressRetries > 0 && *v.Status == "opened" {
					startProgressRetries -= 1
					return resource.RetryableError(fmt.Errorf("status still opened, retry remaining count: %d", startProgressRetries))
				}
				return resource.NonRetryableError(fmt.Errorf("status %s, postgresql instance %s open public service fail", *v.Status, instanceId))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return
}

func (me *PostgresqlService) DescribePostgresqlInstanceById(ctx context.Context, instanceId string) (instance *postgresql.DBInstance, has bool, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeDBInstanceAttributeRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceAttribute(request)
	if err != nil {
		errRet = err
		return
	}

	if response == nil || response.Response == nil {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		return
	}
	instance = response.Response.DBInstance
	if instance != nil && *instance.DBInstanceStatus != "isolated" && *instance.DBInstanceStatus != "recycled" && *instance.DBInstanceStatus != "offline" {
		has = true
	}
	return
}

func (me *PostgresqlService) DescribePostgresqlInstanceHAConfigById(ctx context.Context, instanceId string) (haConfig *postgresql.DescribeDBInstanceHAConfigResponseParams, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeDBInstanceHAConfigRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceHAConfig(request)
	if err != nil {
		errRet = err
		return
	}

	if response == nil || response.Response == nil {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		return
	}

	haConfig = response.Response
	return
}

func (me *PostgresqlService) DescribePostgresqlInstances(ctx context.Context, filter []*postgresql.Filter) (instanceList []*postgresql.DBInstance, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeDBInstancesRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()

	var offset, limit uint64 = 0, 10

	for {
		request.Offset = &offset
		request.Limit = &limit
		request.Filters = filter
		ratelimit.Check(request.GetAction())
		response, err := me.client.UsePostgresqlClient().DescribeDBInstances(request)
		if err != nil {
			errRet = err
			return
		}
		if response == nil || response.Response == nil {
			errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		}
		instanceList = append(instanceList, response.Response.DBInstanceSet...)
		if len(response.Response.DBInstanceSet) < int(limit) {
			return
		}
		offset += limit
	}
}

func (me *PostgresqlService) ModifyPostgresqlInstanceName(ctx context.Context, instanceId string, name string) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewModifyDBInstanceNameRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId
	request.InstanceName = &name

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().ModifyDBInstanceName(request)
	return err
}

func (me *PostgresqlService) UpgradePostgresqlInstance(ctx context.Context, instanceId string, memory int, storage int) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewModifyDBInstanceSpecRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId
	request.Storage = helper.IntUint64(storage)
	request.Memory = helper.IntUint64(memory)

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().ModifyDBInstanceSpec(request)
	return err
}

func (me *PostgresqlService) ModifyPostgresqlInstanceProjectId(ctx context.Context, instanceId string, projectId int) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewModifyDBInstancesProjectRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceIdSet = []*string{&instanceId}
	request.ProjectId = helper.String(strconv.Itoa(projectId))

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().ModifyDBInstancesProject(request)
	return err
}

func (me *PostgresqlService) SetPostgresqlInstanceAutoRenewFlag(ctx context.Context, instanceId string, autoRenewFlag int) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewSetAutoRenewFlagRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceIdSet = []*string{&instanceId}
	request.AutoRenewFlag = helper.IntInt64(autoRenewFlag)

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().SetAutoRenewFlag(request)
	return err
}

func (me *PostgresqlService) IsolatePostgresqlInstance(ctx context.Context, instanceId string) (errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewIsolateDBInstancesRequest()
	request.DBInstanceIdSet = []*string{&instanceId}
	ratelimit.Check(request.GetAction())
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()

	errRet = resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		_, errRet = me.client.UsePostgresqlClient().IsolateDBInstances(request)
		if errRet != nil {
			log.Printf("[CRITAL]%s describe account failed, reason: %v", logId, errRet)
			return retryError(errRet)
		}
		return nil
	})
	return errRet
}

func (me *PostgresqlService) DeletePostgresqlInstance(ctx context.Context, instanceId string) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDestroyDBInstanceRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().DestroyDBInstance(request)
	return err
}

func (me *PostgresqlService) SetPostgresqlInstanceRootPassword(ctx context.Context, instanceId string, password string) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewResetAccountPasswordRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.DBInstanceId = &instanceId
	request.UserName = helper.String("root")
	request.Password = &password

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().ResetAccountPassword(request)
	return err
}

func (me *PostgresqlService) CheckDBInstanceStatus(ctx context.Context, instanceId string, retryMinutes ...int) error {

	var timeout = 2 * readRetryTimeout
	if len(retryMinutes) > 0 && retryMinutes[0] > 0 {
		times := retryMinutes[0]
		timeout = time.Minute * time.Duration(times)
	}
	// check status
	err := resource.Retry(timeout, func() *resource.RetryError {
		instance, has, err := me.DescribePostgresqlInstanceById(ctx, instanceId)
		if err != nil {
			return retryError(err)
		}
		if !has {
			return resource.NonRetryableError(fmt.Errorf("check postgresql instance %s fail", instanceId))
		}
		if *instance.DBInstanceStatus == POSTGRESQL_STAUTS_RUNNING {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("checking postgresql instance %s , status %s ", instanceId, *instance.DBInstanceStatus))
	})

	return err
}

func (me *PostgresqlService) DescribeRootUser(ctx context.Context, instanceId string) (accounts []*postgresql.AccountInfo, errRet error) {
	logId := getLogId(ctx)
	orderBy := "createTime"
	orderByType := "asc"

	request := postgresql.NewDescribeAccountsRequest()
	request.DBInstanceId = &instanceId
	request.OrderBy = &orderBy
	request.OrderByType = &orderByType
	var response *postgresql.DescribeAccountsResponse
	errRet = resource.Retry(2*readRetryTimeout, func() *resource.RetryError {
		response, errRet = me.client.UsePostgresqlClient().DescribeAccounts(request)
		if errRet != nil {
			log.Printf("[CRITAL]%s describe account failed, reason: %v", logId, errRet)
			return retryError(errRet)
		}
		return nil
	})
	if errRet != nil {
		return nil, errRet
	}
	if response == nil || response.Response == nil || response.Response.Details == nil {
		errRet = fmt.Errorf("TencentCloud SDK return nil response, %+v, %s", response, request.GetAction())
	} else {
		accounts = response.Response.Details
	}

	return accounts, errRet
}

func (me *PostgresqlService) ModifyPgParams(ctx context.Context, instanceId string, paramEntrys map[string]string) (err error) {
	logId := getLogId(ctx)
	request := postgresql.NewModifyDBInstanceParametersRequest()
	request.DBInstanceId = &instanceId
	request.ParamList = make([]*postgresql.ParamEntry, 0)

	for key, value := range paramEntrys {
		request.ParamList = append(request.ParamList, &postgresql.ParamEntry{
			Name:          helper.String(key),
			ExpectedValue: helper.String(value),
		})
	}
	_, err = me.client.UsePostgresqlClient().ModifyDBInstanceParameters(request)

	if err != nil {
		log.Printf("[CRITAL]%s modify pgInstance parameter failed, reason: %v", logId, err)
		return err
	}
	return nil
}

func (me *PostgresqlService) DescribePgParams(ctx context.Context, instanceId string) (params map[string]string, err error) {
	logId := getLogId(ctx)
	stateCheckRequest := postgresql.NewDescribeParamsEventRequest()
	stateCheckRequest.DBInstanceId = &instanceId

	// wait for param effective
	err = resource.Retry(readRetryTimeout, func() *resource.RetryError {
		stateCheckResponse, inErr := me.client.UsePostgresqlClient().DescribeParamsEvent(stateCheckRequest)
		if inErr != nil {
			ee, ok := inErr.(*sdkErrors.TencentCloudSDKError)
			if ok && ee.GetCode() == "ResourceNotFound.InstanceNotFoundError" {
				return nil
			}
			return retryError(inErr)
		}

		for _, eventItem := range stateCheckResponse.Response.EventItems {
			eventDetail := eventItem.EventDetail
			if eventDetail == nil {
				continue
			}
			for _, eventDetailItem := range eventDetail {
				if *(eventDetailItem.State) != "success" {
					return resource.RetryableError(
						fmt.Errorf("params is updating!"))
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	request := postgresql.NewDescribeDBInstanceParametersRequest()
	request.DBInstanceId = &instanceId

	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceParameters(request)

	if err != nil {
		log.Printf("[CRITAL]%s fetch pg instance parameter failed, reason: %v", logId, err)
		return
	}
	detail := response.Response.Detail
	params = make(map[string]string)
	for _, item := range detail {
		params[*item.Name] = *item.CurrentValue
	}

	return
}

func (me *PostgresqlService) ModifyDBInstanceDeployment(ctx context.Context, request *postgresql.ModifyDBInstanceDeploymentRequest) (errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().ModifyDBInstanceDeployment(request)

	if err != nil {
		errRet = err
		return
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) DescribeDBInstanceAttribute(ctx context.Context, request *postgresql.DescribeDBInstanceAttributeRequest) (ins *postgresql.DBInstance, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceAttribute(request)

	if err != nil {
		errRet = err
		return
	}

	ins = response.Response.DBInstance

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) DescribeDBEncryptionKeys(ctx context.Context, request *postgresql.DescribeEncryptionKeysRequest) (has bool, key *postgresql.EncryptionKey, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeEncryptionKeys(request)
	if err != nil {
		errRet = err
		return
	}

	keys := response.Response.EncryptionKeys
	if len(keys) < 1 {
		return
	}
	has = true
	key = keys[0]
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	return
}

func (me *PostgresqlService) DescribePostgresqlReadonlyGroups(ctx context.Context, filter []*postgresql.Filter) (instanceList []*postgresql.ReadOnlyGroup, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeReadOnlyGroupsRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()

	var offset, limit int64 = 1, 10

	for {
		request.PageNumber = &offset
		request.PageSize = &limit
		request.Filters = filter
		ratelimit.Check(request.GetAction())
		response, err := me.client.UsePostgresqlClient().DescribeReadOnlyGroups(request)
		if err != nil {
			errRet = err
			return
		}
		if response == nil || response.Response == nil {
			errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		}
		instanceList = append(instanceList, response.Response.ReadOnlyGroupList...)
		if len(response.Response.ReadOnlyGroupList) < int(limit) {
			return
		}
		offset += 1
	}
}

func (me *PostgresqlService) DescribePostgresqlReadOnlyGroupById(ctx context.Context, dbInstanceId string) (instanceList []*postgresql.ReadOnlyGroup, errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDescribeReadOnlyGroupsRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()

	request.Filters = append(
		request.Filters,
		&postgresql.Filter{
			Name:   helper.String("db-master-instance-id"),
			Values: []*string{&dbInstanceId},
		},
	)

	var offset, limit int64 = 1, 10

	for {
		request.PageNumber = &offset
		request.PageSize = &limit
		ratelimit.Check(request.GetAction())
		response, err := me.client.UsePostgresqlClient().DescribeReadOnlyGroups(request)
		if err != nil {
			errRet = err
			return
		}
		if response == nil || response.Response == nil {
			errRet = fmt.Errorf("TencentCloud SDK return nil response, %s", request.GetAction())
		}
		instanceList = append(instanceList, response.Response.ReadOnlyGroupList...)
		if len(response.Response.ReadOnlyGroupList) < int(limit) {
			return
		}
		offset += 1
	}
}

func (me *PostgresqlService) DeletePostgresqlReadOnlyGroupById(ctx context.Context, groupId string) (errRet error) {
	logId := getLogId(ctx)
	request := postgresql.NewDeleteReadOnlyGroupRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail,reason[%s]", logId, request.GetAction(), errRet.Error())
		}
	}()
	request.ReadOnlyGroupId = &groupId

	ratelimit.Check(request.GetAction())
	_, err := me.client.UsePostgresqlClient().DeleteReadOnlyGroup(request)
	return err
}

func (me *PostgresqlService) DescribePostgresqlParameterTemplatesByFilter(ctx context.Context, param map[string]interface{}) (ParameterTemplates []*postgresql.ParameterTemplate, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeParameterTemplatesRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "filters" {
			request.Filters = v.([]*postgresql.Filter)
		}
		if k == "order_by" {
			request.OrderBy = v.(*string)
		}
		if k == "order_by_type" {
			request.OrderByType = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset int64 = 0
		limit  int64 = 20
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UsePostgresqlClient().DescribeParameterTemplates(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.ParameterTemplateSet) < 1 {
			break
		}
		ParameterTemplates = append(ParameterTemplates, response.Response.ParameterTemplateSet...)
		if len(response.Response.ParameterTemplateSet) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *PostgresqlService) DescribePostgresqlParameterTemplateById(ctx context.Context, templateId string) (ParameterTemplate *postgresql.DescribeParameterTemplateAttributesResponseParams, errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDescribeParameterTemplateAttributesRequest()
	request.TemplateId = &templateId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeParameterTemplateAttributes(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	ParameterTemplate = response.Response
	return
}

func (me *PostgresqlService) DeletePostgresqlParameterTemplateById(ctx context.Context, templateId string) (errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDeleteParameterTemplateRequest()
	request.TemplateId = &templateId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DeleteParameterTemplate(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) PostgresqlUpgradeKernelVersionRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ctx := contextNil

		instance, _, err := me.DescribePostgresqlInstanceById(ctx, instanceId)

		if err != nil {
			return nil, "", err
		}

		return instance, *instance.DBInstanceStatus, nil
	}
}

func (me *PostgresqlService) DescribePostgresqlReadonlyGroupsByFilter(ctx context.Context, param map[string]interface{}) (ReadOnlyGroups []*postgresql.ReadOnlyGroup, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeReadOnlyGroupsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "Filters" {
			request.Filters = v.([]*postgresql.Filter)
		}
		if k == "OrderBy" {
			request.OrderBy = v.(*string)
		}
		if k == "OrderByType" {
			request.OrderByType = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeReadOnlyGroups(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.ReadOnlyGroupList) < 1 {
		return
	}
	ReadOnlyGroups = response.Response.ReadOnlyGroupList

	return
}

func (me *PostgresqlService) DescribePostgresqlBackupPlanConfigById(ctx context.Context, dBInstanceId string) (BackupPlanConfig *postgresql.BackupPlan, errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDescribeBackupPlansRequest()
	request.DBInstanceId = &dBInstanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeBackupPlans(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if len(response.Response.Plans) < 1 {
		return
	}

	BackupPlanConfig = response.Response.Plans[0]
	return
}

func (me *PostgresqlService) DescribePostgresqlBackupDownloadRestrictionConfigById(ctx context.Context, restrictionType string) (BackupDownloadRestrictionConfig *postgresql.DescribeBackupDownloadRestrictionResponseParams, errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDescribeBackupDownloadRestrictionRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeBackupDownloadRestriction(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	BackupDownloadRestrictionConfig = response.Response
	return
}

func (me *PostgresqlService) DescribePostgresqlSecurityGroupConfigById(ctx context.Context, dBInstanceId string, readOnlyGroupId string) (SecurityGroupConfigs []*postgresql.SecurityGroup, errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDescribeDBInstanceSecurityGroupsRequest()

	if dBInstanceId != "" {
		request.DBInstanceId = &dBInstanceId
	}
	if readOnlyGroupId != "" {
		request.ReadOnlyGroupId = &readOnlyGroupId
	}

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceSecurityGroups(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if len(response.Response.SecurityGroupSet) < 1 {
		return
	}

	SecurityGroupConfigs = response.Response.SecurityGroupSet
	return
}

func (me *PostgresqlService) PostgresqlDbInstanceOperationStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ctx := contextNil

		instance, _, err := me.DescribePostgresqlInstanceById(ctx, instanceId)

		if err != nil {
			return nil, "", err
		}

		return instance, *instance.DBInstanceStatus, nil
	}
}

func (me *PostgresqlService) PostgresqlDBInstanceStateRefreshFunc(dbInstanceId string, failStates []string) resource.StateRefreshFunc {
	logElapsed("PostgresqlDBInstanceStateRefreshFunc called")()
	return func() (interface{}, string, error) {
		ctx := contextNil

		object, has, err := me.DescribePostgresqlInstanceById(ctx, dbInstanceId)

		if err != nil {
			err, ok := err.(*sdkErrors.TencentCloudSDKError)
			if ok && err.GetCode() == "ResourceNotFound.InstanceNotFoundError" {
				// it is ok
				return nil, "", nil
			}
			return nil, "", err
		}
		if object == nil || !has {
			return &postgresql.DBInstance{}, "closed", nil
		}

		return object, helper.PString(object.DBInstanceStatus), nil
	}
}

func (me *PostgresqlService) DescribePostgresqlDBInstanceNetInfosById(ctx context.Context, dBInstanceId string) (netInfos []*postgresql.DBInstanceNetInfo, errRet error) {
	logElapsed("DescribePostgresqlDBInstanceNetInfosById called")()
	logId := getLogId(ctx)

	request := postgresql.NewDescribeDBInstanceAttributeRequest()
	request.DBInstanceId = &dBInstanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeDBInstanceAttribute(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	instances := response.Response.DBInstance
	if instances != nil {
		netInfos = instances.DBInstanceNetInfo
	}

	return
}

func (me *PostgresqlService) DeletePostgresqlDBInstanceNetworkAccessById(ctx context.Context, dBInstanceId, vpcId, subnetId, vip string) (errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDeleteDBInstanceNetworkAccessRequest()
	request.DBInstanceId = &dBInstanceId
	request.VpcId = &vpcId
	request.SubnetId = &subnetId
	request.Vip = &vip

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DeleteDBInstanceNetworkAccess(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) PostgresqlDBInstanceNetworkAccessStateRefreshFunc(dBInstanceId, vpcId, subnetId, oldVip, newVip string, failStates []string) resource.StateRefreshFunc {
	logElapsed("PostgresqlDBInstanceNetworkAccessStateRefreshFunc called")()
	return func() (interface{}, string, error) {
		ctx := contextNil

		netInfos, err := me.DescribePostgresqlDBInstanceNetInfosById(ctx, dBInstanceId)

		if err != nil {
			err, ok := err.(*sdkErrors.TencentCloudSDKError)
			if ok && err.GetCode() == "ResourceNotFound.InstanceNotFoundError" {
				// it is ok
				return nil, "", nil
			}
			return nil, "", err
		}

		var object *postgresql.DBInstanceNetInfo
		for _, info := range netInfos {
			if *info.NetType == "private" {
				if *info.VpcId == vpcId && *info.SubnetId == subnetId && (*info.Ip != oldVip || *info.Ip == newVip) {
					object = info
					break
				}
			}
		}

		if object == nil {
			return &postgresql.DBInstanceNetInfo{}, "closed", nil
		}

		return object, helper.PString(object.Status), nil
	}
}

func (me *PostgresqlService) DescribePostgresqlReadonlyGroupNetInfosById(ctx context.Context, dbInstanceId, roGroupId string) (netInfos []*postgresql.DBInstanceNetInfo, errRet error) {
	logElapsed("DescribePostgresqlReadonlyGroupNetInfoById called")()
	logId := getLogId(ctx)

	paramMap := map[string]interface{}{
		"Filters": []*postgresql.Filter{
			{
				Name:   helper.String("db-master-instance-id"),
				Values: []*string{helper.String(dbInstanceId)},
			},
		},
	}

	result, err := me.DescribePostgresqlReadonlyGroupsByFilter(ctx, paramMap)
	if err != nil {
		errRet = err
		return
	}

	roGroup := result[0]
	if roGroupId != "" {
		for _, group := range result {
			if *group.ReadOnlyGroupId == roGroupId {
				roGroup = group
				break
			}
		}
	}

	if roGroup != nil {
		netInfos = roGroup.DBInstanceNetInfo
	}

	log.Printf("[DEBUG]%s DescribePostgresqlReadonlyGroupNetworkAccessById dbInstanceId:[%s] roGroupId:[%s] success, result is roGroup:[%v], \n", logId, dbInstanceId, roGroupId, roGroup)
	return
}

func (me *PostgresqlService) DeletePostgresqlReadonlyGroupNetworkAccessById(ctx context.Context, readOnlyGroupId, vpcId, subnetId, vip string) (errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDeleteReadOnlyGroupNetworkAccessRequest()
	request.ReadOnlyGroupId = &readOnlyGroupId
	request.VpcId = &vpcId
	request.SubnetId = &subnetId
	request.Vip = &vip

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DeleteReadOnlyGroupNetworkAccess(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *PostgresqlService) PostgresqlReadonlyGroupNetworkAccessStateRefreshFunc(dbInstanceId, roGroupId, vpcId, subnetId, oldVip, newVip string, failStates []string) resource.StateRefreshFunc {
	logElapsed("PostgresqlReadonlyGroupNetworkAccessStateRefreshFunc called")()
	return func() (interface{}, string, error) {
		ctx := contextNil

		netInfos, err := me.DescribePostgresqlReadonlyGroupNetInfosById(ctx, dbInstanceId, roGroupId)

		if err != nil {
			err, ok := err.(*sdkErrors.TencentCloudSDKError)
			if ok && err.GetCode() == "ResourceNotFound.InstanceNotFoundError" {
				// it is ok
				return nil, "", nil
			}
			return nil, "", err
		}

		var object *postgresql.DBInstanceNetInfo
		for _, info := range netInfos {
			if *info.NetType == "private" {
				if *info.VpcId == vpcId && *info.SubnetId == subnetId && (*info.Ip != oldVip || *info.Ip == newVip) {
					object = info
					break
				}
			}
		}

		if object == nil {
			return &postgresql.DBInstanceNetInfo{}, "closed", nil
		}

		return object, helper.PString(object.Status), nil
	}
}

func (me *PostgresqlService) PostgresqlReadonlyGroupStateRefreshFunc(dbInstanceId, roGroupId string, failStates []string) resource.StateRefreshFunc {
	logElapsed("PostgresqlReadonlyGroupStateRefreshFunc called")()
	return func() (interface{}, string, error) {
		ctx := contextNil

		result, err := me.DescribePostgresqlReadOnlyGroupById(ctx, dbInstanceId)

		roGroup := result[0]
		if roGroupId != "" {
			for _, group := range result {
				if *group.ReadOnlyGroupId == roGroupId {
					roGroup = group
					break
				}
			}
		}

		if err != nil {
			err, ok := err.(*sdkErrors.TencentCloudSDKError)
			if ok && err.GetCode() == "ResourceNotFound.InstanceNotFoundError" {
				// it is ok
				return nil, "", nil
			}
			return nil, "", err
		}
		if roGroup == nil {
			return &postgresql.ReadOnlyGroup{}, "closed", nil
		}

		return roGroup, helper.PString(roGroup.Status), nil
	}
}

func (me *PostgresqlService) DescribePostgresqlBackupDownloadUrlsByFilter(ctx context.Context, param map[string]interface{}) (BackupDownloadUrl *string, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeBackupDownloadURLRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "DBInstanceId" {
			request.DBInstanceId = v.(*string)
		}
		if k == "BackupType" {
			request.BackupType = v.(*string)
		}
		if k == "BackupId" {
			request.BackupId = v.(*string)
		}
		if k == "URLExpireTime" {
			request.URLExpireTime = v.(*uint64)
		}
		if k == "BackupDownloadRestriction" {
			request.BackupDownloadRestriction = v.(*postgresql.BackupDownloadRestriction)
		}
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UsePostgresqlClient().DescribeBackupDownloadURL(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || response.Response.BackupDownloadURL == nil {
		return
	}
	BackupDownloadUrl = response.Response.BackupDownloadURL

	return
}

func (me *PostgresqlService) DescribePostgresqlBaseBackupsByFilter(ctx context.Context, param map[string]interface{}) (BaseBackups []*postgresql.BaseBackup, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeBaseBackupsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "MinFinishTime" {
			request.MinFinishTime = v.(*string)
		}
		if k == "MaxFinishTime" {
			request.MaxFinishTime = v.(*string)
		}
		if k == "Filters" {
			request.Filters = v.([]*postgresql.Filter)
		}
		if k == "OrderBy" {
			request.OrderBy = v.(*string)
		}
		if k == "OrderByType" {
			request.OrderByType = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset int64 = 0
		limit  int64 = 20
	)
	for {
		request.Offset = helper.Int64Uint64(offset)
		request.Limit = helper.Int64Uint64(limit)
		response, err := me.client.UsePostgresqlClient().DescribeBaseBackups(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.BaseBackupSet) < 1 {
			break
		}
		BaseBackups = append(BaseBackups, response.Response.BaseBackupSet...)
		if len(response.Response.BaseBackupSet) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *PostgresqlService) DescribePostgresqlLogBackupsByFilter(ctx context.Context, param map[string]interface{}) (LogBackups []*postgresql.LogBackup, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeLogBackupsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "MinFinishTime" {
			request.MinFinishTime = v.(*string)
		}
		if k == "MaxFinishTime" {
			request.MaxFinishTime = v.(*string)
		}
		if k == "Filters" {
			request.Filters = v.([]*postgresql.Filter)
		}
		if k == "OrderBy" {
			request.OrderBy = v.(*string)
		}
		if k == "OrderByType" {
			request.OrderByType = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset uint64 = 0
		limit  uint64 = 20
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UsePostgresqlClient().DescribeLogBackups(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.LogBackupSet) < 1 {
			break
		}
		LogBackups = append(LogBackups, response.Response.LogBackupSet...)
		if len(response.Response.LogBackupSet) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *PostgresqlService) DescribePostgresqlDbInstanceClassesByFilter(ctx context.Context, param map[string]interface{}) (DbInstanceClasses []*postgresql.ClassInfo, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeClassesRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "Zone" {
			request.Zone = v.(*string)
		}
		if k == "DBEngine" {
			request.DBEngine = v.(*string)
		}
		if k == "DBMajorVersion" {
			request.DBMajorVersion = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeClasses(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.ClassInfoSet) < 1 {
		return
	}
	DbInstanceClasses = response.Response.ClassInfoSet

	return
}

func (me *PostgresqlService) DescribePostgresqlDefaultParametersByFilter(ctx context.Context, param map[string]interface{}) (DefaultParameters []*postgresql.ParamInfo, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeDefaultParametersRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "DBMajorVersion" {
			request.DBMajorVersion = v.(*string)
		}
		if k == "DBEngine" {
			request.DBEngine = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeDefaultParameters(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.ParamInfoSet) < 1 {
		return
	}
	DefaultParameters = response.Response.ParamInfoSet

	return
}

func (me *PostgresqlService) DescribePostgresqlRecoveryTimeByFilter(ctx context.Context, param map[string]interface{}) (ret *postgresql.DescribeAvailableRecoveryTimeResponseParams, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeAvailableRecoveryTimeRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "DBInstanceId" {
			request.DBInstanceId = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeAvailableRecoveryTime(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil {
		return
	}
	ret = response.Response

	return
}

func (me *PostgresqlService) DescribePostgresqlRegionsByFilter(ctx context.Context) (Regions []*postgresql.RegionInfo, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeRegionsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeRegions(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.RegionSet) < 1 {
		return
	}
	Regions = response.Response.RegionSet

	return
}

func (me *PostgresqlService) DescribePostgresqlDbInstanceVersionsByFilter(ctx context.Context) (DbInstanceVersions []*postgresql.Version, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeDBVersionsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeDBVersions(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.VersionSet) < 1 {
		return
	}
	DbInstanceVersions = response.Response.VersionSet

	return
}

func (me *PostgresqlService) DescribePostgresqlZonesByFilter(ctx context.Context) (Zones []*postgresql.ZoneInfo, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = postgresql.NewDescribeZonesRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UsePostgresqlClient().DescribeZones(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.ZoneSet) < 1 {
		return
	}
	Zones = response.Response.ZoneSet

	return
}

func (me *PostgresqlService) DescribePostgresqlBaseBackupById(ctx context.Context, baseBackupId string) (BaseBackup *postgresql.BaseBackup, errRet error) {
	logId := getLogId(ctx)

	params := map[string]interface{}{
		"Filters": []*postgresql.Filter{
			{
				Name: helper.String("base-backup-id"),
				Values: []*string{
					helper.String(baseBackupId),
				},
			},
		},
	}

	backups, err := me.DescribePostgresqlBaseBackupsByFilter(ctx, params)
	if err != nil {
		errRet = err
		return
	}

	if len(backups) == 1 {
		BaseBackup = backups[0]
		log.Printf("[DEBUG]%s DescribePostgresqlBaseBackupById success, BaseBackupId:[%s]\n", logId, *BaseBackup.Id)
	}
	return
}

func (me *PostgresqlService) DeletePostgresqlBaseBackupById(ctx context.Context, dBInstanceId string, baseBackupId string) (errRet error) {
	logId := getLogId(ctx)

	request := postgresql.NewDeleteBaseBackupRequest()
	request.DBInstanceId = &dBInstanceId
	request.BaseBackupId = &baseBackupId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	errRet = resource.Retry(3*writeRetryTimeout, func() *resource.RetryError {
		response, e := me.client.UsePostgresqlClient().DeleteBaseBackup(request)
		if e != nil {
			tcErr := e.(*sdkErrors.TencentCloudSDKError)

			if tcErr.Code == "FailedOperation.FailedOperationError" {
				return resource.RetryableError(fmt.Errorf("state not ready, retry...: %v", e.Error()))
			}
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
		}
		return nil
	})

	return
}
