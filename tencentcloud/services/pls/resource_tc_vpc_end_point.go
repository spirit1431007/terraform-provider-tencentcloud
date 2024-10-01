package pls

import (
	"context"
	"fmt"
	"log"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	svcvpc "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/vpc"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudVpcEndPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcEndPointCreate,
		Read:   resourceTencentCloudVpcEndPointRead,
		Update: resourceTencentCloudVpcEndPointUpdate,
		Delete: resourceTencentCloudVpcEndPointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "ID of vpc instance.",
			},

			"subnet_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "ID of subnet instance.",
			},

			"end_point_name": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Name of endpoint.",
			},

			"end_point_service_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "ID of endpoint service.",
			},

			"end_point_vip": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "VIP of endpoint ip.",
			},

			"security_groups_ids": {
				Optional:    true,
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Ordered security groups associated with the endpoint.",
			},

			"end_point_owner": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "APPID.",
			},

			"state": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "state of end point.",
			},

			"create_time": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Create Time.",
			},

			"cdc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CDC instance ID.",
			},
		},
	}
}

func resourceTencentCloudVpcEndPointCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_vpc_end_point.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	var (
		request    = vpc.NewCreateVpcEndPointRequest()
		response   = vpc.NewCreateVpcEndPointResponse()
		endPointId string
	)
	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = helper.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		request.SubnetId = helper.String(v.(string))
	}

	if v, ok := d.GetOk("end_point_name"); ok {
		request.EndPointName = helper.String(v.(string))
	}

	if v, ok := d.GetOk("end_point_service_id"); ok {
		request.EndPointServiceId = helper.String(v.(string))
	}

	if v, ok := d.GetOk("end_point_vip"); ok {
		request.EndPointVip = helper.String(v.(string))
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseVpcClient().CreateVpcEndPoint(request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s create vpc endPoint failed, reason:%+v", logId, err)
		return err
	}
	endPointId = *response.Response.EndPoint.EndPointId
	d.SetId(endPointId)

	if v, ok := d.GetOk("security_groups_ids"); ok {
		request := vpc.NewModifyVpcEndPointAttributeRequest()
		request.EndPointId = helper.String(endPointId)
		request.SecurityGroupIds = helper.InterfacesStringsPoint(v.([]interface{}))

		err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
			result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseVpcClient().ModifyVpcEndPointAttribute(request)
			if e != nil {
				return tccommon.RetryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s create vpc endPoint failed, reason:%+v", logId, err)
			return err
		}

	}
	return resourceTencentCloudVpcEndPointRead(d, meta)
}

func resourceTencentCloudVpcEndPointRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_vpc_end_point.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	service := svcvpc.NewVpcService(meta.(tccommon.ProviderMeta).GetAPIV3Conn())

	endPointId := d.Id()

	endPoint, err := service.DescribeVpcEndPointById(ctx, endPointId)
	if err != nil {
		return err
	}

	if endPoint == nil {
		d.SetId("")
		return fmt.Errorf("resource `track` %s does not exist", d.Id())
	}

	if endPoint.VpcId != nil {
		_ = d.Set("vpc_id", endPoint.VpcId)
	}

	if endPoint.SubnetId != nil {
		_ = d.Set("subnet_id", endPoint.SubnetId)
	}

	if endPoint.EndPointName != nil {
		_ = d.Set("end_point_name", endPoint.EndPointName)
	}

	if endPoint.EndPointServiceId != nil {
		_ = d.Set("end_point_service_id", endPoint.EndPointServiceId)
	}

	if endPoint.EndPointVip != nil {
		_ = d.Set("end_point_vip", endPoint.EndPointVip)
	}

	if endPoint.EndPointOwner != nil {
		_ = d.Set("end_point_owner", endPoint.EndPointOwner)
	}

	if endPoint.State != nil {
		_ = d.Set("state", endPoint.State)
	}

	if endPoint.CreateTime != nil {
		_ = d.Set("create_time", endPoint.CreateTime)
	}

	if endPoint.GroupSet != nil {
		_ = d.Set("security_groups_ids", endPoint.GroupSet)
	}

	if endPoint.CdcId != nil {
		_ = d.Set("cdc_id", endPoint.CdcId)
	}

	return nil
}

func resourceTencentCloudVpcEndPointUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_vpc_end_point.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	request := vpc.NewModifyVpcEndPointAttributeRequest()

	endPointId := d.Id()

	request.EndPointId = &endPointId

	unsupportedUpdateFields := []string{
		"vpc_id",
		"subnet_id",
		"end_point_service_id",
		"end_point_vip",
	}
	for _, field := range unsupportedUpdateFields {
		if d.HasChange(field) {
			return fmt.Errorf("tencentcloud_vpc_end_point update on %s is not support yet", field)
		}
	}

	if d.HasChange("end_point_name") || d.HasChange("security_groups_ids") {
		if v, ok := d.GetOk("end_point_name"); ok {
			request.EndPointName = helper.String(v.(string))
		}

		if v, ok := d.GetOk("security_groups_ids"); ok {
			request.SecurityGroupIds = helper.InterfacesStringsPoint(v.([]interface{}))
		}
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseVpcClient().ModifyVpcEndPointAttribute(request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s create vpc endPoint failed, reason:%+v", logId, err)
		return err
	}

	return resourceTencentCloudVpcEndPointRead(d, meta)
}

func resourceTencentCloudVpcEndPointDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_vpc_end_point.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	service := svcvpc.NewVpcService(meta.(tccommon.ProviderMeta).GetAPIV3Conn())
	endPointId := d.Id()

	if err := service.DeleteVpcEndPointById(ctx, endPointId); err != nil {
		return nil
	}

	return nil
}
