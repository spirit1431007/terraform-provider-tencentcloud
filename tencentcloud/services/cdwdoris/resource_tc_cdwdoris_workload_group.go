// Code generated by iacg; DO NOT EDIT.
package cdwdoris

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cdwdorisv20211228 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdwdoris/v20211228"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudCdwdorisWorkloadGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudCdwdorisWorkloadGroupCreate,
		Read:   resourceTencentCloudCdwdorisWorkloadGroupRead,
		Update: resourceTencentCloudCdwdorisWorkloadGroupUpdate,
		Delete: resourceTencentCloudCdwdorisWorkloadGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Instance id.",
			},
			"workload_group": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Resource group configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workload_group_name": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Workload group name. Note: This field may return null, indicating that no valid value can be obtained.",
						},
						"cpu_share": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "CPU weight. Note: This field may return null, indicating that no valid value can be obtained.",
						},
						"memory_limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Memory limit, the sum of the memory limit values of all resource groups should be less than or equal to 100. Note: This field may return null, indicating that no valid value can be obtained.",
						},
						"enable_memory_over_commit": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to allow over-allocation. Note: This field may return null, indicating that no valid value can be obtained.",
						},
						"cpu_hard_limit": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Cpu hard limit. Note: This field may return null, indicating that no valid value can be obtained.",
						},
					},
				},
			},
		},
	}
}

func resourceTencentCloudCdwdorisWorkloadGroupCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cdwdoris_workload_group.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		logId             = tccommon.GetLogId(tccommon.ContextNil)
		ctx               = tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)
		request           = cdwdorisv20211228.NewCreateWorkloadGroupRequest()
		instanceId        string
		workloadGroupName string
	)

	if v, ok := d.GetOk("instance_id"); ok {
		request.InstanceId = helper.String(v.(string))
		instanceId = v.(string)
	}

	if workloadGroupMap, ok := helper.InterfacesHeadMap(d, "workload_group"); ok {
		workloadGroupConfig := cdwdorisv20211228.WorkloadGroupConfig{}
		if v, ok := workloadGroupMap["workload_group_name"]; ok {
			workloadGroupConfig.WorkloadGroupName = helper.String(v.(string))
			workloadGroupName = v.(string)
		}

		if v, ok := workloadGroupMap["cpu_share"]; ok {
			workloadGroupConfig.CpuShare = helper.IntInt64(v.(int))
		}

		if v, ok := workloadGroupMap["memory_limit"]; ok {
			workloadGroupConfig.MemoryLimit = helper.IntInt64(v.(int))
		}

		if v, ok := workloadGroupMap["enable_memory_over_commit"]; ok {
			workloadGroupConfig.EnableMemoryOverCommit = helper.Bool(v.(bool))
		}

		if v, ok := workloadGroupMap["cpu_hard_limit"]; ok {
			workloadGroupConfig.CpuHardLimit = helper.String(v.(string))
		}

		request.WorkloadGroup = &workloadGroupConfig
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseCdwdorisV20211228Client().CreateWorkloadGroupWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]. ", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}

		return nil
	})

	if err != nil {
		log.Printf("[CRITAL]%s create cdwdoris workload group failed, reason:%+v", logId, err)
		return err
	}

	d.SetId(strings.Join([]string{instanceId, workloadGroupName}, tccommon.FILED_SP))

	return resourceTencentCloudCdwdorisWorkloadGroupRead(d, meta)
}

func resourceTencentCloudCdwdorisWorkloadGroupRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cdwdoris_workload_group.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)
		service = CdwdorisService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}
	)

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}

	instanceId := idSplit[0]
	workloadGroupName := idSplit[1]

	respData, err := service.DescribeCdwdorisWorkloadGroupById(ctx, instanceId, workloadGroupName)
	if err != nil {
		return err
	}

	if respData == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `cdwdoris_workload_group` [%s] not found, please check if it has been deleted.. ", logId, d.Id())
		return nil
	}

	_ = d.Set("instance_id", instanceId)
	workloadGroupsList := make([]map[string]interface{}, 0)
	workloadGroupsMap := map[string]interface{}{}
	if respData.WorkloadGroupName != nil {
		workloadGroupsMap["workload_group_name"] = respData.WorkloadGroupName
	}

	if respData.CpuShare != nil {
		workloadGroupsMap["cpu_share"] = respData.CpuShare
	}

	if respData.MemoryLimit != nil {
		workloadGroupsMap["memory_limit"] = respData.MemoryLimit
	}

	if respData.EnableMemoryOverCommit != nil {
		workloadGroupsMap["enable_memory_over_commit"] = respData.EnableMemoryOverCommit
	}

	if respData.CpuHardLimit != nil {
		workloadGroupsMap["cpu_hard_limit"] = respData.CpuHardLimit
	}

	workloadGroupsList = append(workloadGroupsList, workloadGroupsMap)
	_ = d.Set("workload_group", workloadGroupsList)

	return nil
}

func resourceTencentCloudCdwdorisWorkloadGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cdwdoris_workload_group.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)
		request = cdwdorisv20211228.NewModifyWorkloadGroupRequest()
	)

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}

	instanceId := idSplit[0]
	workloadGroupName := idSplit[1]
	request.InstanceId = &instanceId
	if workloadGroupMap, ok := helper.InterfacesHeadMap(d, "workload_group"); ok {
		workloadGroupConfig := cdwdorisv20211228.WorkloadGroupConfig{}
		workloadGroupConfig.WorkloadGroupName = helper.String(workloadGroupName)
		if v, ok := workloadGroupMap["cpu_share"]; ok {
			workloadGroupConfig.CpuShare = helper.IntInt64(v.(int))
		}

		if v, ok := workloadGroupMap["memory_limit"]; ok {
			workloadGroupConfig.MemoryLimit = helper.IntInt64(v.(int))
		}

		if v, ok := workloadGroupMap["enable_memory_over_commit"]; ok {
			workloadGroupConfig.EnableMemoryOverCommit = helper.Bool(v.(bool))
		}

		if v, ok := workloadGroupMap["cpu_hard_limit"]; ok {
			workloadGroupConfig.CpuHardLimit = helper.String(v.(string))
		}

		request.WorkloadGroup = &workloadGroupConfig
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseCdwdorisV20211228Client().ModifyWorkloadGroupWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]. ", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}

		return nil
	})

	if err != nil {
		log.Printf("[CRITAL]%s update cdwdoris workload group failed, reason:%+v", logId, err)
		return err
	}

	return resourceTencentCloudCdwdorisWorkloadGroupRead(d, meta)
}

func resourceTencentCloudCdwdorisWorkloadGroupDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cdwdoris_workload_group.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)
		request = cdwdorisv20211228.NewDeleteWorkloadGroupRequest()
	)

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}

	instanceId := idSplit[0]
	workloadGroupName := idSplit[1]
	request.InstanceId = helper.String(instanceId)
	request.WorkloadGroupName = helper.String(workloadGroupName)
	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseCdwdorisV20211228Client().DeleteWorkloadGroupWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]. ", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}

		return nil
	})

	if err != nil {
		log.Printf("[CRITAL]%s delete cdwdoris workload group failed, reason:%+v", logId, err)
		return err
	}

	return nil
}
