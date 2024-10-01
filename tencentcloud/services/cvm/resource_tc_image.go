package cvm

import (
	"context"
	"fmt"
	"log"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	svctag "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/tag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

func ResourceTencentCloudImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudImageCreate,
		Read:   resourceTencentCloudImageRead,
		Update: resourceTencentCloudImageUpdate,
		Delete: resourceTencentCloudImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"image_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Image name.",
			},
			"instance_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"snapshot_ids"},
				Description:  "Cloud server instance ID.",
			},
			"snapshot_ids": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"instance_id"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Cloud disk snapshot ID list; creating a mirror based on a snapshot must include a system disk snapshot. It cannot be passed in simultaneously with InstanceId.",
			},
			"image_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Image Description.",
			},
			"force_poweroff": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set whether to force shutdown during mirroring. The default value is `false`, when set to true, it means that the mirror will be made after shutdown.",
			},
			"sysprep": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Sysprep function under Windows. When creating a Windows image, you can select true or false to enable or disable the Syspre function.",
			},
			"data_disk_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Cloud disk ID list, When creating a whole machine image based on an instance, specify the data disk ID contained in the image.",
			},
			"image_family": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Set image family. Example value: `business-daily-update`.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags of the image.",
			},
		},
	}
}

func resourceTencentCloudImageCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_image.create")()
	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	cvmService := CvmService{
		client: meta.(tccommon.ProviderMeta).GetAPIV3Conn(),
	}

	request := cvm.NewCreateImageRequest()
	request.ImageName = helper.String(d.Get("image_name").(string))
	if d.Get("force_poweroff").(bool) {
		request.ForcePoweroff = helper.String(TRUE)
	} else {
		request.ForcePoweroff = helper.String(FALSE)
	}

	if v, ok := d.GetOk("image_description"); ok {
		request.ImageDescription = helper.String(v.(string))
	}
	if v, ok := d.GetOkExists("sysprep"); ok {
		value := v.(bool)
		if value {
			request.Sysprep = helper.String(TRUE)
		} else {
			request.Sysprep = helper.String(FALSE)
		}
	}
	if v, ok := d.GetOk("data_disk_ids"); ok {
		diskIds := v.(*schema.Set).List()
		diskArr := make([]*string, 0, len(diskIds))
		for _, id := range diskIds {
			diskArr = append(diskArr, helper.String(id.(string)))
		}
		request.DataDiskIds = diskArr
	}
	if v, ok := d.GetOk("instance_id"); ok {
		request.InstanceId = helper.String(v.(string))
	}
	if v, ok := d.GetOk("snapshot_ids"); ok {
		ids := v.(*schema.Set).List()
		snapshotIds := make([]*string, 0, len(ids))
		for _, v := range ids {
			snapshotIds = append(snapshotIds, helper.String(v.(string)))
		}
		request.SnapshotIds = snapshotIds
	}

	if len(request.SnapshotIds) > 0 && len(request.DataDiskIds) > 0 {
		return fmt.Errorf("`%s` and `%s` Can't appear in the profile China at the same time,The parameter `%s` depends on the pre_parameter `%s`",
			"snapshot_ids", "data_disk_ids", "data_disk_ids", "instance_id")
	}

	if v, ok := d.GetOk("image_family"); ok {
		request.ImageFamily = helper.String(v.(string))
	}

	if v := helper.GetTags(d, "tags"); len(v) > 0 {
		tags := make([]*cvm.Tag, 0)
		for tagKey, tagValue := range v {
			tag := cvm.Tag{
				Key:   helper.String(tagKey),
				Value: helper.String(tagValue),
			}
			tags = append(tags, &tag)
		}
		tagSpecification := cvm.TagSpecification{
			ResourceType: helper.String("image"),
			Tags:         tags,
		}
		request.TagSpecification = append(request.TagSpecification, &tagSpecification)
	}

	imageId := ""
	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		response, err := cvmService.client.UseCvmClient().CreateImage(request)
		if err != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), err.Error())
			return tccommon.RetryError(err)
		}
		imageId = *response.Response.ImageId
		return nil
	})
	if err != nil {
		return err
	}
	d.SetId(imageId)

	// Wait for the tags attached to the vm since tags attachment it's async while vm creation.
	if tags := helper.GetTags(d, "tags"); len(tags) > 0 {
		tcClient := meta.(tccommon.ProviderMeta).GetAPIV3Conn()
		tagService := svctag.NewTagService(tcClient)
		resourceName := tccommon.BuildTagResourceName("cvm", "image", tcClient.Region, imageId)
		if err := tagService.ModifyTags(ctx, resourceName, tags, nil); err != nil {
			// If tags attachment failed, the user will be notified, then plan/apply/update with terraform.
			return err
		}
	}

	// wait for status
	_, has, errRet := cvmService.DescribeImageById(ctx, imageId, false)
	if errRet != nil {
		return errRet
	}
	if !has {
		return fmt.Errorf("[CRITAL]%s creating cvm image failed, image doesn't exist", logId)
	}

	return resourceTencentCloudImageRead(d, meta)
}

func resourceTencentCloudImageRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_image.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	imageId := d.Id()
	cvmService := CvmService{
		client: meta.(tccommon.ProviderMeta).GetAPIV3Conn(),
	}

	image, has, errRet := cvmService.DescribeImageById(ctx, imageId, false)
	if errRet != nil {
		return errRet
	}
	if !has {
		d.SetId("")
		return nil
	}

	_ = d.Set("image_name", image.ImageName)
	if image.ImageDescription != nil && *image.ImageDescription != "" {
		_ = d.Set("image_description", image.ImageDescription)
	}

	if image.ImageFamily != nil {
		_ = d.Set("image_family", image.ImageFamily)
	}

	// Use the resource value when the instance_id in the resource is not empty.
	// the instance ID is not returned in the query response body.
	instanceId := ""
	if v, ok := d.GetOk("instance_id"); ok {
		instanceId = v.(string)
	}

	snapShotSysDisk := make([]interface{}, 0, len(image.SnapshotSet))
	for _, v := range image.SnapshotSet {
		snapShotSysDisk = append(snapShotSysDisk, v.SnapshotId)
	}

	if instanceId != "" {
		_ = d.Set("instance_id", helper.String(instanceId))
	} else {
		_ = d.Set("snapshot_ids", snapShotSysDisk)
	}

	client := meta.(tccommon.ProviderMeta).GetAPIV3Conn()
	tagService := svctag.NewTagService(client)

	tags, err := tagService.DescribeResourceTags(ctx, "cvm", "image", client.Region, d.Id())
	if err != nil {
		return err
	}
	_ = d.Set("tags", tags)
	return nil
}

func resourceTencentCloudImageUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_image.update")()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	instanceId := d.Id()
	cvmService := CvmService{
		client: meta.(tccommon.ProviderMeta).GetAPIV3Conn(),
	}

	if d.HasChange("image_name") || d.HasChange("image_description") || d.HasChange("image_family") {
		imageName := d.Get("image_name").(string)
		imageDesc := d.Get("image_description").(string)
		imageFamily := d.Get("image_family").(string)

		if err := cvmService.ModifyImage(ctx, instanceId, imageName, imageDesc, imageFamily); nil != err {
			return err
		}
	}

	if d.HasChange("tags") {
		oldInterface, newInterface := d.GetChange("tags")
		replaceTags, deleteTags := svctag.DiffTags(oldInterface.(map[string]interface{}), newInterface.(map[string]interface{}))
		tagService := svctag.NewTagService(meta.(tccommon.ProviderMeta).GetAPIV3Conn())
		region := meta.(tccommon.ProviderMeta).GetAPIV3Conn().Region
		resourceName := tccommon.BuildTagResourceName("cvm", "image", region, instanceId)
		err := tagService.ModifyTags(ctx, resourceName, replaceTags, deleteTags)
		if err != nil {
			return err
		}
	}

	return resourceTencentCloudImageRead(d, meta)
}

func resourceTencentCloudImageDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_image.delete")()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	cvmService := CvmService{
		client: meta.(tccommon.ProviderMeta).GetAPIV3Conn(),
	}

	imageId := d.Id()

	if err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		e := cvmService.DeleteImage(ctx, imageId)
		if e != nil {
			return tccommon.RetryError(e)
		}
		return nil
	}); nil != err {
		return err
	}

	//check image
	if err := resource.Retry(3*tccommon.ReadRetryTimeout, func() *resource.RetryError {
		_, has, err := cvmService.DescribeImageById(ctx, imageId, true)
		if err != nil {
			return tccommon.RetryError(err)
		}
		if has {
			return resource.RetryableError(fmt.Errorf("image exits error,image_id = %s", imageId))
		}
		return nil
	}); nil != err {
		return err
	}

	return nil
}
