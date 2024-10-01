package vpc

import (
	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	svctag "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/tag"

	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func DataSourceTencentCloudVpcSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudVpcSubnetsRead,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be queried.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the subnet to be queried.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the subnet to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone of the subnet to be queried.",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter default or no default subnets.",
			},
			"tag_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter if subnet has this tag.",
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter subnet with this CIDR.",
			},
			"cdc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of CDC instance.",
			},
			"is_remote_vpc_snat": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter the VPC SNAT address pool subnet.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags of the subnet to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},

			// Computed values
			"instance_list": {Type: schema.TypeList,
				Computed:    true,
				Description: "List of subnets.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The availability zone of the subnet.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VPC.",
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the subnet.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subnet.",
						},
						"cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A network address block of the subnet.",
						},
						"cdc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of CDC instance.",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether it is the default subnet of the VPC for this region.",
						},
						"is_multicast": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether multicast is enabled.",
						},
						"route_table_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the routing table.",
						},
						"available_ip_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of available IPs.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the subnet resource.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Tags of the subnet resource.",
						},
					},
				},
			},
		},
	}
}

func dataSourceTencentCloudVpcSubnetsRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("data_source.tencentcloud_vpc_subnets.read")()

	var (
		logId            = tccommon.GetLogId(tccommon.ContextNil)
		ctx              = context.WithValue(context.TODO(), tccommon.LogIdKey, logId)
		vpcService       = VpcService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}
		tagService       = svctag.NewTagService(meta.(tccommon.ProviderMeta).GetAPIV3Conn())
		region           = meta.(tccommon.ProviderMeta).GetAPIV3Conn().Region
		vpcId            string
		subnetId         string
		name             string
		availabilityZone string
		isDefault        *bool
		isRemoteVpcSNAT  *bool
		tagKey           string
		cidrBlock        string
		cdcId            string
	)

	if temp, ok := d.GetOk("vpc_id"); ok {
		vpcId = temp.(string)
	}

	if temp, ok := d.GetOk("subnet_id"); ok {
		subnetId = temp.(string)
	}

	if temp, ok := d.GetOk("name"); ok {
		name = temp.(string)
	}

	if temp, ok := d.GetOk("availability_zone"); ok {
		availabilityZone = temp.(string)
	}

	if temp, ok := d.GetOkExists("is_default"); ok {
		isDefault = helper.Bool(temp.(bool))
	}

	if temp, ok := d.GetOkExists("is_remote_vpc_snat"); ok {
		isRemoteVpcSNAT = helper.Bool(temp.(bool))
	}

	if temp, ok := d.GetOk("tag_key"); ok {
		tagKey = temp.(string)
	}

	if temp, ok := d.GetOk("cidr_block"); ok {
		cidrBlock = temp.(string)
	}

	if temp, ok := d.GetOk("cdc_id"); ok {
		cdcId = temp.(string)
	}

	var (
		tags  = helper.GetTags(d, "tags")
		infos []VpcSubnetBasicInfo
		err   error
	)

	err = resource.Retry(tccommon.ReadRetryTimeout, func() *resource.RetryError {
		infos, err = vpcService.DescribeSubnets(ctx, subnetId, vpcId,
			name, availabilityZone, tags,
			isDefault, isRemoteVpcSNAT, tagKey,
			cidrBlock, cdcId)

		if err != nil {
			return tccommon.RetryError(err, tccommon.InternalError)
		}

		return nil
	})

	if err != nil {
		return err
	}

	var infoList = make([]map[string]interface{}, 0, len(infos))
	for _, item := range infos {
		respTags, err := tagService.DescribeResourceTags(ctx, "vpc", "subnet", region, item.subnetId)
		if err != nil {
			return err
		}

		var infoMap = make(map[string]interface{})
		infoMap["availability_zone"] = item.zone
		infoMap["vpc_id"] = item.vpcId
		infoMap["subnet_id"] = item.subnetId
		infoMap["name"] = item.name
		infoMap["cidr_block"] = item.cidr
		infoMap["cdc_id"] = item.cdcId
		infoMap["is_default"] = item.isDefault
		infoMap["is_multicast"] = item.isMulticast
		infoMap["route_table_id"] = item.routeTableId
		infoMap["available_ip_count"] = item.availableIpCount
		infoMap["create_time"] = item.createTime
		infoMap["tags"] = respTags
		infoList = append(infoList, infoMap)
	}

	if err := d.Set("instance_list", infoList); err != nil {
		log.Printf("[CRITAL]%s provider set  subnet instances fail, reason:%s\n ", logId, err.Error())
		return err
	}

	idBytes, err := json.Marshal(map[string]interface{}{
		"vpcId":            vpcId,
		"subnetId":         subnetId,
		"availabilityZone": availabilityZone,
		"name":             name,
		"isDefault":        isDefault,
		"tagKey":           tagKey,
		"isRemoteVpcSnat":  isRemoteVpcSNAT,
		"cidrBlock":        cidrBlock,
		"tags":             tags,
	})

	if err != nil {
		log.Printf("[CRITAL]%s create data source id error, reason:%s\n ", logId, err.Error())
		return err
	}

	md := md5.New()
	_, _ = md.Write(idBytes)
	id := fmt.Sprintf("%x", md.Sum(nil))
	d.SetId(id)

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := tccommon.WriteToFile(output.(string), infoList); err != nil {
			log.Printf("[CRITAL]%s output file[%s] fail, reason[%s]\n",
				logId, output.(string), err.Error())
			return err
		}
	}

	return nil
}
