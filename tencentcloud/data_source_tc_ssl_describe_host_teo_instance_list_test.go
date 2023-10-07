package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudSslDescribeHostTeoInstanceListDataSource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccSslDescribeHostTeoInstanceListDataSource,
				Check: resource.ComposeTestCheckFunc(testAccCheckTencentCloudDataSourceID("data.tencentcloud_ssl_describe_host_teo_instance_list.describe_host_teo_instance_list"),
					resource.TestCheckResourceAttr("data.tencentcloud_ssl_describe_host_teo_instance_list.describe_host_teo_instance_list", "certificate_id", "8u8DII0l"),
					resource.TestCheckResourceAttr("data.tencentcloud_ssl_describe_host_teo_instance_list.describe_host_teo_instance_list", "resource_type", "teo"),
				),
			},
		},
	})
}

const testAccSslDescribeHostTeoInstanceListDataSource = `

data "tencentcloud_ssl_describe_host_teo_instance_list" "describe_host_teo_instance_list" {
  certificate_id = "8u8DII0l"
  resource_type = "teo"
}
`