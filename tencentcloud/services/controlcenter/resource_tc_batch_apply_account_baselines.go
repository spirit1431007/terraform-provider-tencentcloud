// Code generated by iacg; DO NOT EDIT.
package controlcenter

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	controlcenterv20230110 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/controlcenter/v20230110"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudBatchApplyAccountBaselines() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudBatchApplyAccountBaselinesCreate,
		Read:   resourceTencentCloudBatchApplyAccountBaselinesRead,
		Delete: resourceTencentCloudBatchApplyAccountBaselinesDelete,
		Schema: map[string]*schema.Schema{
			"member_uin_list": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Description: "Member account UIN, which is also the UIN of the account to which the baseline is applied.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"baseline_config_items": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "List of baseline item configuration information.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A unique identifier for an Account Factory baseline item, which can only contain English letters, digits, and @,._[]-:()+=. It must be 2-128 characters long.Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"configuration": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Account Factory baseline item configuration. Different items have different parameters.Note: This field may return null, indicating that no valid values can be obtained.",
						},
					},
				},
			},
		},
	}
}

func resourceTencentCloudBatchApplyAccountBaselinesCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_batch_apply_account_baselines.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		logId         = tccommon.GetLogId(tccommon.ContextNil)
		ctx           = tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)
		request       = controlcenterv20230110.NewBatchApplyAccountBaselinesRequest()
		memberUinList []string
	)

	if v, ok := d.GetOk("member_uin_list"); ok {
		memberUinListSet := v.(*schema.Set).List()
		for i := range memberUinListSet {
			value := memberUinListSet[i].(int)
			valueStr := strconv.Itoa(value)
			request.MemberUinList = append(request.MemberUinList, helper.IntInt64(value))
			memberUinList = append(memberUinList, valueStr)
		}
	}

	if v, ok := d.GetOk("baseline_config_items"); ok {
		for _, item := range v.([]interface{}) {
			baselineConfigItemsMap := item.(map[string]interface{})
			baselineConfigItem := controlcenterv20230110.BaselineConfigItem{}
			if v, ok := baselineConfigItemsMap["identifier"]; ok {
				baselineConfigItem.Identifier = helper.String(v.(string))
			}

			if v, ok := baselineConfigItemsMap["configuration"]; ok {
				baselineConfigItem.Configuration = helper.String(v.(string))
			}

			request.BaselineConfigItems = append(request.BaselineConfigItems, &baselineConfigItem)
		}
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseControlcenterV20230110Client().BatchApplyAccountBaselinesWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}

		return nil
	})

	if err != nil {
		log.Printf("[CRITAL]%s create batch apply account baselines failed, reason:%+v", logId, err)
		return err
	}

	d.SetId(strings.Join(memberUinList, tccommon.FILED_SP))

	return resourceTencentCloudBatchApplyAccountBaselinesRead(d, meta)
}

func resourceTencentCloudBatchApplyAccountBaselinesRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_batch_apply_account_baselines.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	return nil
}

func resourceTencentCloudBatchApplyAccountBaselinesDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_batch_apply_account_baselines.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	return nil
}
