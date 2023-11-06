package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// go test -i; go test -test.run TestAccTencentCloudMysqlRoGroupLoadOperationResource_basic -v -timeout=0
func TestAccTencentCloudMysqlRoGroupLoadOperationResource_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMysqlRoGroupLoadOperation,
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_mysql_ro_group_load_operation.ro_group_load_operation", "id")),
			},
		},
	})
}

const testAccMysqlRoGroupLoadOperation = `
variable "security_groups" {
	default = "` + defaultCrsSecurityGroups + `"
}

variable "availability_zone" {
  default = "ap-guangzhou-4"
}

resource "tencentcloud_vpc" "vpc" {
  name       = "vpc-mysql"
  cidr_block = "10.0.0.0/16"
}

resource "tencentcloud_subnet" "subnet" {
  availability_zone = var.availability_zone
  name              = "subnet-mysql"
  vpc_id            = tencentcloud_vpc.vpc.id
  cidr_block        = "10.0.0.0/16"
  is_multicast      = false
}

resource "tencentcloud_mysql_instance" "example" {
  internet_service  = 1
  engine_version    = "5.7"
  charge_type       = "POSTPAID"
  root_password     = "PassWord123"
  slave_deploy_mode = 0
  availability_zone = var.availability_zone
  slave_sync_mode   = 1
  instance_name     = "tf-example-mysql"
  mem_size          = 2000
  volume_size       = 200
  vpc_id            = tencentcloud_vpc.vpc.id
  subnet_id         = tencentcloud_subnet.subnet.id
  intranet_port     = 3306
  security_groups   = [var.security_groups]

  tags = {
    name = "test"
  }

  parameters = {
    character_set_server = "utf8"
    max_connections      = "1000"
  }
}

resource "tencentcloud_mysql_readonly_instance" "foo" {
  master_instance_id = tencentcloud_mysql_instance.example.id
  instance_name      = "tf-mysql"
  mem_size           = 2000
  volume_size        = 200
  vpc_id             = tencentcloud_vpc.vpc.id
  subnet_id          = tencentcloud_subnet.subnet.id
  intranet_port      = 3306
  security_groups    = [var.security_groups]

  tags = {
    createBy = "terraform"
  }
}

data "tencentcloud_mysql_instance" "foo" {
  mysql_id = tencentcloud_mysql_instance.example.id

  depends_on = [tencentcloud_mysql_readonly_instance.foo]
}

resource "tencentcloud_mysql_ro_group_load_operation" "ro_group_load_operation" {
  ro_group_id = data.tencentcloud_mysql_instance.foo.instance_list.0.ro_groups.0.group_id
}

`