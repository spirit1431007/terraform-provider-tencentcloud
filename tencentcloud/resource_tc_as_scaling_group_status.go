/*
Provides a resource to set as scaling_group status

Example Usage

Deactivate Scaling Group

```hcl
data "tencentcloud_availability_zones_by_product" "zones" {
  product = "as"
}

data "tencentcloud_images" "image" {
  image_type = ["PUBLIC_IMAGE"]
  os_name    = "TencentOS Server 3.2 (Final)"
}

resource "tencentcloud_vpc" "vpc" {
  name       = "vpc-example"
  cidr_block = "10.0.0.0/16"
}

resource "tencentcloud_subnet" "subnet" {
  vpc_id            = tencentcloud_vpc.vpc.id
  name              = "subnet-example"
  cidr_block        = "10.0.0.0/16"
  availability_zone = data.tencentcloud_availability_zones_by_product.zones.zones.0.name
}

resource "tencentcloud_as_scaling_config" "example" {
  configuration_name = "tf-example"
  image_id           = data.tencentcloud_images.image.images.0.image_id
  instance_types     = ["SA1.SMALL1", "SA2.SMALL1", "SA2.SMALL2", "SA2.SMALL4"]
  instance_name_settings {
    instance_name = "test-ins-name"
  }
}

resource "tencentcloud_as_scaling_group" "example" {
  scaling_group_name = "tf-example"
  configuration_id   = tencentcloud_as_scaling_config.example.id
  max_size           = 1
  min_size           = 0
  vpc_id             = tencentcloud_vpc.vpc.id
  subnet_ids         = [tencentcloud_subnet.subnet.id]
}

resource "tencentcloud_as_scaling_group_status" "scaling_group_status" {
  auto_scaling_group_id = tencentcloud_as_scaling_group.example.id
  enable                = false
}
```

Enable Scaling Group

```hcl
resource "tencentcloud_as_scaling_group_status" "scaling_group_status" {
  auto_scaling_group_id = tencentcloud_as_scaling_group.example.id
  enable                = true
}
```

Import

as scaling_group_status can be imported using the id, e.g.

```
terraform import tencentcloud_as_scaling_group_status.scaling_group_status scaling_group_id
```
*/
package tencentcloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
)

func resourceTencentCloudAsScalingGroupStatus() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudAsScalingGroupStatusCreate,
		Read:   resourceTencentCloudAsScalingGroupStatusRead,
		Update: resourceTencentCloudAsScalingGroupStatusUpdate,
		Delete: resourceTencentCloudAsScalingGroupStatusDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"auto_scaling_group_id": {
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Description: "Scaling group ID.",
			},

			"enable": {
				Required:    true,
				Type:        schema.TypeBool,
				Description: "If enable auto scaling group.",
			},
		},
	}
}

func resourceTencentCloudAsScalingGroupStatusCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_as_scaling_group_status.create")()
	defer inconsistentCheck(d, meta)()

	autoScalingGroupId := d.Get("auto_scaling_group_id").(string)

	d.SetId(autoScalingGroupId)

	return resourceTencentCloudAsScalingGroupStatusUpdate(d, meta)
}

func resourceTencentCloudAsScalingGroupStatusRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_as_scaling_group_status.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := AsService{client: meta.(*TencentCloudClient).apiV3Conn}

	autoScalingGroupId := d.Id()

	scalingGroup, has, err := service.DescribeAutoScalingGroupById(ctx, autoScalingGroupId)
	if err != nil {
		return err
	}

	if has == 0 {
		d.SetId("")
		log.Printf("[WARN]%s resource `AsScalingGroupStatus` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if scalingGroup.AutoScalingGroupId != nil {
		_ = d.Set("auto_scaling_group_id", scalingGroup.AutoScalingGroupId)
	}

	if scalingGroup.EnabledStatus != nil {
		if *scalingGroup.EnabledStatus == "ENABLED" {
			_ = d.Set("enable", true)
		} else {
			_ = d.Set("enable", false)
		}
	}

	return nil
}

func resourceTencentCloudAsScalingGroupStatusUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_as_scaling_group_status.update")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	var (
		enable           bool
		enableAsRequest  = as.NewEnableAutoScalingGroupRequest()
		disableAsRequest = as.NewDisableAutoScalingGroupRequest()
	)

	autoScalingGroupId := d.Id()

	if v, ok := d.GetOkExists("enable"); ok {
		enable = v.(bool)
	}

	if enable {
		enableAsRequest.AutoScalingGroupId = &autoScalingGroupId
		err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			result, e := meta.(*TencentCloudClient).apiV3Conn.UseAsClient().EnableAutoScalingGroup(enableAsRequest)
			if e != nil {
				return retryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, enableAsRequest.GetAction(), enableAsRequest.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s enable vpc snapshotPolicyConfig failed, reason:%+v", logId, err)
			return err
		}
	} else {
		disableAsRequest.AutoScalingGroupId = &autoScalingGroupId
		err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			result, e := meta.(*TencentCloudClient).apiV3Conn.UseAsClient().DisableAutoScalingGroup(disableAsRequest)
			if e != nil {
				return retryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, enableAsRequest.GetAction(), enableAsRequest.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s disable vpc snapshotPolicyConfig failed, reason:%+v", logId, err)
			return err
		}
	}

	return resourceTencentCloudAsScalingGroupStatusRead(d, meta)
}

func resourceTencentCloudAsScalingGroupStatusDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_as_scaling_group_status.delete")()
	defer inconsistentCheck(d, meta)()

	return nil
}