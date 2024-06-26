// Code generated by iacg; DO NOT EDIT.
package teo

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
)

func ResourceTencentCloudTeoCertificateConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudTeoCertificateConfigCreate,
		Read:   resourceTencentCloudTeoCertificateConfigRead,
		Update: resourceTencentCloudTeoCertificateConfigUpdate,
		Delete: resourceTencentCloudTeoCertificateConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Site ID.",
			},

			"host": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Acceleration domain name that needs to modify the certificate configuration.",
			},

			"server_cert_info": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "SSL certificate configuration, this parameter takes effect only when mode = sslcert, just enter the corresponding CertId. You can go to the SSL certificate list to view the CertId.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cert_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the server certificate.Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"alias": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Alias of the certificate.Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Type of the certificate. Values: `default`: Default certificate; `upload`: Specified certificate; `managed`: Tencent Cloud-managed certificate. Note: This field may return `null`, indicating that no valid value can be obtained.",
						},
						"expire_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Time when the certificate expires. Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"deploy_time": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Time when the certificate is deployed. Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"sign_algo": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Signature algorithm. Note: This field may return null, indicating that no valid values can be obtained.",
						},
						"common_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Domain name of the certificate. Note: This field may return `null`, indicating that no valid value can be obtained.",
						},
					},
				},
			},

			"mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Mode of configuring the certificate, the values are: `disable`: Do not configure the certificate; `eofreecert`: Configure EdgeOne free certificate; `sslcert`: Configure SSL certificate. If not filled in, the default value is `disable`.",
			},
		},
	}
}

func resourceTencentCloudTeoCertificateConfigCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_teo_certificate_config.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	var (
		zoneId string
		host   string
	)
	if v, ok := d.GetOk("zone_id"); ok {
		zoneId = v.(string)
	}
	if v, ok := d.GetOk("host"); ok {
		host = v.(string)
	}

	d.SetId(strings.Join([]string{zoneId, host}, tccommon.FILED_SP))

	return resourceTencentCloudTeoCertificateConfigUpdate(d, meta)
}

func resourceTencentCloudTeoCertificateConfigRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_teo_certificate_config.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	service := TeoService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	zoneId := idSplit[0]
	host := idSplit[1]

	_ = d.Set("zone_id", zoneId)

	_ = d.Set("host", host)

	respData, err := service.DescribeTeoCertificateConfigById(ctx, zoneId, host)
	if err != nil {
		return err
	}

	if respData == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `teo_certificate_config` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if err := resourceTencentCloudTeoCertificateConfigReadPostHandleResponse0(ctx, respData); err != nil {
		return err
	}

	return nil
}

func resourceTencentCloudTeoCertificateConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_teo_certificate_config.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	zoneId := idSplit[0]
	host := idSplit[1]

	if err := resourceTencentCloudTeoCertificateConfigUpdateOnStart(ctx); err != nil {
		return err
	}

	if err := resourceTencentCloudTeoCertificateConfigUpdateOnExit(ctx); err != nil {
		return err
	}

	_ = zoneId
	_ = host
	return resourceTencentCloudTeoCertificateConfigRead(d, meta)
}

func resourceTencentCloudTeoCertificateConfigDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_teo_certificate_config.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	return nil
}
