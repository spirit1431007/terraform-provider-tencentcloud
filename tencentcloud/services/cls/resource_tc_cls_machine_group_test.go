package cls_test

import (
	tcacctest "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/acctest"
	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	localcls "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/cls"

	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("tencentcloud_cls_machine_group", &resource.Sweeper{
		Name: "tencentcloud_cls_machine_group",
		F:    testSweepMachineGroup,
	})
}

func testSweepMachineGroup(region string) error {
	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	sharedClient, err := tcacctest.SharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("getting tencentcloud client error: %s", err.Error())
	}
	client := sharedClient.(tccommon.ProviderMeta).GetAPIV3Conn()

	clsService := localcls.NewClsService(client)

	instances, err := clsService.DescribeClsMachineGroupByFilter(ctx, nil)
	if err != nil {
		return fmt.Errorf("get instance list error: %s", err.Error())
	}

	// add scanning resources
	var resources, nonKeepResources []*tccommon.ResourceInstance
	for _, v := range instances {
		if !tccommon.CheckResourcePersist(*v.GroupName, *v.CreateTime) {
			nonKeepResources = append(nonKeepResources, &tccommon.ResourceInstance{
				Id:   *v.GroupId,
				Name: *v.GroupName,
			})
		}
		resources = append(resources, &tccommon.ResourceInstance{
			Id:         *v.GroupId,
			Name:       *v.GroupName,
			CreateTime: *v.CreateTime,
		})
	}
	tccommon.ProcessScanCloudResources(client, resources, nonKeepResources, "CreateMachineGroup")

	for _, v := range instances {
		instanceId := v.GroupId
		instanceName := v.GroupName

		now := time.Now()

		createTime := tccommon.StringToTime(*v.CreateTime)
		interval := now.Sub(createTime).Minutes()
		if strings.HasPrefix(*instanceName, tcacctest.KeepResource) || strings.HasPrefix(*instanceName, tcacctest.DefaultResource) {
			continue
		}
		// less than 30 minute, not delete
		if tccommon.NeedProtect == 1 && int64(interval) < 30 {
			continue
		}

		if err = clsService.DeleteClsMachineGroup(ctx, *instanceId); err != nil {
			log.Printf("[ERROR] sweep instance %s error: %s", *instanceId, err.Error())
		}
	}

	return nil
}

func TestAccTencentCloudClsMachineGroup_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { tcacctest.AccPreCheck(t) },
		Providers: tcacctest.AccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClsMachineGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClsMachineGroupExists("tencentcloud_cls_machine_group.group"),
					resource.TestCheckResourceAttr("tencentcloud_cls_machine_group.group", "group_name", "tf-basic-group"),
				),
			},
			{
				ResourceName:      "tencentcloud_cls_machine_group.group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckClsMachineGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		logId := tccommon.GetLogId(tccommon.ContextNil)
		ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("[CHECK][CLS machine group][Exists] check: CLS machine group %s is not found", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("[CHECK][CLS machine group][Exists] check: CLS machine group id is not set")
		}
		clsService := localcls.NewClsService(tcacctest.AccProvider.Meta().(tccommon.ProviderMeta).GetAPIV3Conn())
		instance, err := clsService.DescribeClsMachineGroupById(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if instance == nil {
			return fmt.Errorf("[CHECK][CLS machine group][Exists] id %s is not exist", rs.Primary.ID)
		}
		return nil
	}
}

const testAccClsMachineGroup = `
resource "tencentcloud_cls_machine_group" "group" {
  group_name        = "tf-basic-group"
  service_logging   = true
  auto_update       = true
  update_end_time   = "19:05:00"
  update_start_time = "17:05:00"

  machine_group_type {
    type   = "ip"
    values = [
      "192.168.1.1",
      "192.168.1.2",
    ]
  }
}
`
