package teo

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"

	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

func resourceTencentCloudTeoCertificateConfigReadPostHandleResponse0(ctx context.Context, resp *teo.AccelerationDomain) error {
	d := tccommon.ResourceDataFromContext(ctx)
	meta := tccommon.ProviderMetaFromContext(ctx)
	service := TeoService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	var (
		zoneId string
	)
	if v, ok := d.GetOk("zone_id"); ok {
		zoneId = v.(string)
	}

	accelerationDomain := resp
	if accelerationDomain.Certificate != nil {
		certificate := accelerationDomain.Certificate
		zone, err := service.DescribeTeoZone(ctx, zoneId)
		if err != nil {
			return err
		}

		serverCertInfoList := []interface{}{}
		for _, serverCertInfo := range certificate.List {
			serverCertInfoMap := map[string]interface{}{}

			if serverCertInfo.CertId != nil {
				serverCertInfoMap["cert_id"] = serverCertInfo.CertId
			}

			if serverCertInfo.Alias != nil {
				serverCertInfoMap["alias"] = serverCertInfo.Alias
			}

			if serverCertInfo.Type != nil {
				serverCertInfoMap["type"] = serverCertInfo.Type
			}

			if serverCertInfo.ExpireTime != nil {
				serverCertInfoMap["expire_time"] = serverCertInfo.ExpireTime
			}

			if serverCertInfo.DeployTime != nil {
				serverCertInfoMap["deploy_time"] = serverCertInfo.DeployTime
			}

			if serverCertInfo.SignAlgo != nil {
				serverCertInfoMap["sign_algo"] = serverCertInfo.SignAlgo
			}

			if zone.ZoneName != nil {
				serverCertInfoMap["common_name"] = zone.ZoneName
			}

			serverCertInfoList = append(serverCertInfoList, serverCertInfoMap)
		}

		_ = d.Set("server_cert_info", serverCertInfoList)

		if certificate.Mode != nil {
			_ = d.Set("mode", certificate.Mode)
		}
	}

	return nil
}

func resourceTencentCloudTeoCertificateConfigUpdateOnExit(ctx context.Context) error {
	d := tccommon.ResourceDataFromContext(ctx)
	meta := tccommon.ProviderMetaFromContext(ctx)
	service := TeoService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

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

	err := service.CheckAccelerationDomainStatus(ctx, zoneId, host, "")
	if err != nil {
		return err
	}

	return nil
}

func resourceTencentCloudTeoCertificateConfigUpdateOnStart(ctx context.Context) error {
	d := tccommon.ResourceDataFromContext(ctx)
	meta := tccommon.ProviderMetaFromContext(ctx)

	logId := ctx.Value(tccommon.LogIdKey)

	request := teo.NewModifyHostsCertificateRequest()

	idSplit := strings.Split(d.Id(), tccommon.FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	zoneId := idSplit[0]
	host := idSplit[1]

	request.ZoneId = &zoneId
	request.Hosts = []*string{&host}

	if v, ok := d.GetOk("server_cert_info"); ok {
		for _, item := range v.([]interface{}) {
			dMap := item.(map[string]interface{})
			serverCertInfo := teo.ServerCertInfo{}
			if v, ok := dMap["cert_id"]; ok {
				serverCertInfo.CertId = helper.String(v.(string))
			}
			if v, ok := dMap["alias"]; ok && v.(string) != "" {
				serverCertInfo.Alias = helper.String(v.(string))
			}
			if v, ok := dMap["type"]; ok && v.(string) != "" {
				serverCertInfo.Type = helper.String(v.(string))
			}
			if v, ok := dMap["expire_time"]; ok && v.(string) != "" {
				serverCertInfo.ExpireTime = helper.String(v.(string))
			} else {
				serverCertInfo.ExpireTime = nil
			}
			if v, ok := dMap["deploy_time"]; ok && v.(string) != "" {
				serverCertInfo.DeployTime = helper.String(v.(string))
			} else {
				serverCertInfo.DeployTime = nil
			}
			if v, ok := dMap["sign_algo"]; ok && v.(string) != "" {
				serverCertInfo.SignAlgo = helper.String(v.(string))
			}
			if v, ok := dMap["common_name"]; ok && v.(string) != "" {
				serverCertInfo.CommonName = helper.String(v.(string))
			}
			request.ServerCertInfo = append(request.ServerCertInfo, &serverCertInfo)
		}
	}

	if v, ok := d.GetOk("mode"); ok {
		request.Mode = helper.String(v.(string))
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseTeoClient().ModifyHostsCertificate(request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s update teo certificate failed, reason:%+v", logId, err)
		return err
	}

	return nil
}
