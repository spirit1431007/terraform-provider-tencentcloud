// Code generated by iacg; DO NOT EDIT.
package cvm

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudCvmLaunchTemplateDefaultVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudCvmLaunchTemplateDefaultVersionCreate,
		Read:   resourceTencentCloudCvmLaunchTemplateDefaultVersionRead,
		Update: resourceTencentCloudCvmLaunchTemplateDefaultVersionUpdate,
		Delete: resourceTencentCloudCvmLaunchTemplateDefaultVersionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"launch_template_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Instance launch template ID.",
			},

			"default_version": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The number of the version that you want to set as the default version.",
			},
		},
	}
}

func resourceTencentCloudCvmLaunchTemplateDefaultVersionCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cvm_launch_template_default_version.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		launchTemplateId string
	)
	if v, ok := d.GetOk("launch_template_id"); ok {
		launchTemplateId = v.(string)
	}

	d.SetId(launchTemplateId)

	return resourceTencentCloudCvmLaunchTemplateDefaultVersionUpdate(d, meta)
}

func resourceTencentCloudCvmLaunchTemplateDefaultVersionRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cvm_launch_template_default_version.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	service := CvmService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	launchTemplateId := d.Id()

	_ = d.Set("launch_template_id", launchTemplateId)

	respData, err := service.DescribeCvmLaunchTemplateDefaultVersionById(ctx, launchTemplateId)
	if err != nil {
		return err
	}

	if respData == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `cvm_launch_template_default_version` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}
	if respData.LaunchTemplateVersion != nil {
		_ = d.Set("launch_template_version", respData.LaunchTemplateVersion)
	}

	return nil
}

func resourceTencentCloudCvmLaunchTemplateDefaultVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cvm_launch_template_default_version.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	launchTemplateId := d.Id()

	needChange := false
	mutableArgs := []string{"default_version"}
	for _, v := range mutableArgs {
		if d.HasChange(v) {
			needChange = true
			break
		}
	}

	if needChange {
		request := cvm.NewModifyLaunchTemplateDefaultVersionRequest()

		request.LaunchTemplateId = helper.String(launchTemplateId)

		if v, ok := d.GetOkExists("default_version"); ok {
			request.DefaultVersion = helper.IntInt64(v.(int))
		}

		err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
			result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseCvmClient().ModifyLaunchTemplateDefaultVersionWithContext(ctx, request)
			if e != nil {
				return tccommon.RetryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s update cvm launch template default version failed, reason:%+v", logId, err)
			return err
		}
	}

	return resourceTencentCloudCvmLaunchTemplateDefaultVersionRead(d, meta)
}

func resourceTencentCloudCvmLaunchTemplateDefaultVersionDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cvm_launch_template_default_version.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	return nil
}
