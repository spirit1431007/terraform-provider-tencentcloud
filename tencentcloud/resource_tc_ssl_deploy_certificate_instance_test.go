package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudSslDeployCertificateInstanceResource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccSslDeployCertificateInstance,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_ssl_deploy_certificate_instance.deploy_certificate_instance", "id"),
					resource.TestCheckResourceAttr("tencentcloud_ssl_deploy_certificate_instance.deploy_certificate_instance", "certificate_id", "8x1eUSSl"),
					resource.TestCheckResourceAttr("tencentcloud_ssl_deploy_certificate_instance.deploy_certificate_instance", "resource_type", "cdn"),
					resource.TestCheckResourceAttr("tencentcloud_ssl_deploy_certificate_instance.deploy_certificate_instance", "instance_id_list.0", "api1.ninghhuang.online|off"),
				),
			},
		},
	})
}

const testAccSslDeployCertificateInstance = `

resource "tencentcloud_ssl_deploy_certificate_instance" "deploy_certificate_instance" {
  certificate_id = "8x1eUSSl"
  resource_type = "cdn"
  instance_id_list =["api1.ninghhuang.online|off"]
}

`