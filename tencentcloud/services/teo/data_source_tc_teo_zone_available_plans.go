// Code generated by iacg; DO NOT EDIT.
package teo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func DataSourceTencentCloudTeoZoneAvailablePlans() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudTeoZoneAvailablePlansRead,
		Schema: map[string]*schema.Schema{
			"plan_info_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Zone plans which current account can use.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Plan type.",
						},
						"currency": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Settlement Currency Type. Valid values: `CNY`, `USD`.",
						},
						"flux": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of fluxes included in the zone plan. Unit: Byte.",
						},
						"frequency": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Billing cycle. Valid values:\n- `y`: Billed by the year.\n- `m`: Billed by the month.\n- `h`: Billed by the hour.\n- `M`: Billed by the minute.\n- `s`: Billed by the second.",
						},
						"price": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Price of the plan. Unit: cent.",
						},
						"request": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of requests included in the zone plan.",
						},
						"site_number": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of zones this zone plan can bind.",
						},
						"area": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Acceleration area of the plan. Valid value: `mainland`, `overseas`.",
						},
					},
				},
			},

			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceTencentCloudTeoZoneAvailablePlansRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("data_source.tencentcloud_teo_zone_available_plans.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(nil)
	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	service := TeoService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	paramMap := make(map[string]interface{})
	var respData []*teo.PlanInfo
	err := resource.Retry(tccommon.ReadRetryTimeout, func() *resource.RetryError {
		result, e := service.DescribeTeoZoneAvailablePlansByFilter(ctx, paramMap)
		if e != nil {
			return tccommon.RetryError(e)
		}
		respData = result
		return nil
	})
	if err != nil {
		return err
	}

	planInfoId := make([]string, 0, len(respData))
	planInfoList := make([]map[string]interface{}, 0, len(respData))
	if respData != nil {
		for _, planInfo := range respData {
			planInfoMap := map[string]interface{}{}

			var planType string
			if planInfo.PlanType != nil {
				planInfoMap["plan_type"] = planInfo.PlanType
				planType = *planInfo.PlanType
			}

			if planInfo.Currency != nil {
				planInfoMap["currency"] = planInfo.Currency
			}

			if planInfo.Flux != nil {
				planInfoMap["flux"] = planInfo.Flux
			}

			if planInfo.Frequency != nil {
				planInfoMap["frequency"] = planInfo.Frequency
			}

			if planInfo.Price != nil {
				planInfoMap["price"] = planInfo.Price
			}

			if planInfo.Request != nil {
				planInfoMap["request"] = planInfo.Request
			}

			if planInfo.SiteNumber != nil {
				planInfoMap["site_number"] = planInfo.SiteNumber
			}

			if planInfo.Area != nil {
				planInfoMap["area"] = planInfo.Area
			}

			planInfoId = append(planInfoId, planType)
			planInfoList = append(planInfoList, planInfoMap)
		}

		_ = d.Set("plan_info_list", planInfoList)
	}

	d.SetId(helper.DataResourceIdsHash(planInfoId))

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := tccommon.WriteToFile(output.(string), planInfoList); e != nil {
			return e
		}
	}

	return nil
}
