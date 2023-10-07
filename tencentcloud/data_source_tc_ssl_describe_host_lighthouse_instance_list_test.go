package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudSslDescribeHostLighthouseInstanceListDataSource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccSslDescribeHostLighthouseInstanceListDataSource,
				Check: resource.ComposeTestCheckFunc(testAccCheckTencentCloudDataSourceID("data.tencentcloud_ssl_describe_host_lighthouse_instance_list.describe_host_lighthouse_instance_list"),
					resource.TestCheckResourceAttr("data.tencentcloud_ssl_describe_host_lighthouse_instance_list.describe_host_lighthouse_instance_list", "certificate_id", "8hUkH3xC"),
					resource.TestCheckResourceAttr("data.tencentcloud_ssl_describe_host_lighthouse_instance_list.describe_host_lighthouse_instance_list", "resource_type", "lighthouse"),
				),
			},
		},
	})
}

const testAccSslDescribeHostLighthouseInstanceListDataSource = `

data "tencentcloud_ssl_describe_host_lighthouse_instance_list" "describe_host_lighthouse_instance_list" {
  certificate_id = "8hUkH3xC"
  resource_type = "lighthouse"
}
`
