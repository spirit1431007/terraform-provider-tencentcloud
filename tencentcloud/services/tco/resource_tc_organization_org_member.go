package tco

import (
	"context"
	"fmt"
	"log"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	svctag "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/tag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	organization "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/organization/v20210331"

	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudOrganizationOrgMember() *schema.Resource {
	return &schema.Resource{
		Read:   resourceTencentCloudOrganizationOrgMemberRead,
		Create: resourceTencentCloudOrganizationOrgMemberCreate,
		Update: resourceTencentCloudOrganizationOrgMemberUpdate,
		Delete: resourceTencentCloudOrganizationOrgMemberDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Member name.",
			},

			"policy_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization policy type.- `Financial`: Financial management policy.",
			},

			"permission_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Required:    true,
				Description: "Financial management permission IDs.Valid values:- `1`: View bill.- `2`: Check balance.- `3`: Fund transfer.- `4`: Combine bill.- `5`: Issue an invoice.- `6`: Inherit discount.- `7`: Pay on behalf.value 1,2 is required.",
			},

			"node_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Organization node ID.",
			},

			"remark": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Notes.",
			},

			"record_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Create member record ID.When create failed and needs to be recreated, is required.",
			},

			"pay_uin": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The uin which is payment account on behalf.When `PermissionIds` contains 7, is required.",
			},

			"node_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization node name.",
			},

			"member_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Member Type.Valid values:- `Invite`: The member is invited.- `Create`: The member is created.",
			},

			"org_policy_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization policy name.",
			},

			"org_permission": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Financial management permissions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Permissions ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Permissions name.",
						},
					},
				},
			},

			"is_allow_quit": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether to allow member to leave the organization.Valid values:- `Allow`.- `Denied`.",
			},

			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Member creation time.",
			},

			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Member update time.",
			},

			"pay_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The member name which is payment account on behalf.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tag description list.",
			},
		},
	}
}

func resourceTencentCloudOrganizationOrgMemberCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_organization_org_member.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	var (
		request  = organization.NewCreateOrganizationMemberRequest()
		response *organization.CreateOrganizationMemberResponse
		uin      int64
	)

	if v, ok := d.GetOk("name"); ok {
		request.Name = helper.String(v.(string))
		request.AccountName = helper.String(v.(string))
	}

	if v, ok := d.GetOk("policy_type"); ok {
		request.PolicyType = helper.String(v.(string))
	}

	if v, ok := d.GetOk("permission_ids"); ok {
		permissionIdsSet := v.(*schema.Set).List()
		for i := range permissionIdsSet {
			permissionIds := permissionIdsSet[i].(int)
			request.PermissionIds = append(request.PermissionIds, helper.IntUint64(permissionIds))
		}
	}

	if v, _ := d.GetOk("node_id"); v != nil {
		request.NodeId = helper.IntInt64(v.(int))
	}

	if v, ok := d.GetOk("remark"); ok {
		request.Remark = helper.String(v.(string))
	}

	if v, _ := d.GetOk("record_id"); v != nil {
		request.RecordId = helper.IntInt64(v.(int))
	}

	if v, ok := d.GetOk("pay_uin"); ok {
		request.PayUin = helper.String(v.(string))
	}

	if tags := helper.GetTags(d, "tags"); len(tags) > 0 {
		for k, v := range tags {
			tmpKey := k
			tmpValue := v
			request.Tags = append(request.Tags, &organization.Tag{
				TagKey:   &tmpKey,
				TagValue: &tmpValue,
			})
		}
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseOrganizationClient().CreateOrganizationMember(request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
				logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})

	if err != nil {
		log.Printf("[CRITAL]%s create organization orgMember failed, reason:%+v", logId, err)
		return err
	}

	uin = *response.Response.Uin

	d.SetId(helper.Int64ToStr(uin))
	return resourceTencentCloudOrganizationOrgMemberRead(d, meta)
}

func resourceTencentCloudOrganizationOrgMemberRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_organization_org_member.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	service := OrganizationService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	orgMemberId := d.Id()

	orgMember, err := service.DescribeOrganizationOrgMember(ctx, orgMemberId)

	if err != nil {
		return err
	}

	if orgMember == nil {
		d.SetId("")
		return fmt.Errorf("resource `orgMember` %s does not exist", orgMemberId)
	}

	if orgMember.Name != nil {
		_ = d.Set("name", orgMember.Name)
	}

	if orgMember.OrgPolicyType != nil {
		_ = d.Set("policy_type", orgMember.OrgPolicyType)
	}

	if orgMember.OrgPermission != nil {
		orgPermissionIds := []uint64{}
		for _, orgPermission := range orgMember.OrgPermission {
			if orgPermission.Id != nil {
				orgPermissionIds = append(orgPermissionIds, *orgPermission.Id)
			}
		}
		_ = d.Set("permission_ids", orgPermissionIds)
	}

	if orgMember.NodeId != nil {
		_ = d.Set("node_id", orgMember.NodeId)
	}

	if orgMember.Remark != nil {
		_ = d.Set("remark", orgMember.Remark)
	}

	if orgMember.PayUin != nil {
		_ = d.Set("pay_uin", orgMember.PayUin)
	}

	if orgMember.NodeName != nil {
		_ = d.Set("node_name", orgMember.NodeName)
	}

	if orgMember.MemberType != nil {
		_ = d.Set("member_type", orgMember.MemberType)
	}

	if orgMember.OrgPolicyName != nil {
		_ = d.Set("org_policy_name", orgMember.OrgPolicyName)
	}

	if orgMember.OrgPermission != nil {
		orgPermissionList := []interface{}{}
		for _, orgPermission := range orgMember.OrgPermission {
			orgPermissionMap := map[string]interface{}{}
			if orgPermission.Id != nil {
				orgPermissionMap["id"] = orgPermission.Id
			}
			if orgPermission.Name != nil {
				orgPermissionMap["name"] = orgPermission.Name
			}

			orgPermissionList = append(orgPermissionList, orgPermissionMap)
		}
		_ = d.Set("org_permission", orgPermissionList)
	}

	if orgMember.IsAllowQuit != nil {
		_ = d.Set("is_allow_quit", orgMember.IsAllowQuit)
	}

	if orgMember.CreateTime != nil {
		_ = d.Set("create_time", orgMember.CreateTime)
	}

	if orgMember.UpdateTime != nil {
		_ = d.Set("update_time", orgMember.UpdateTime)
	}

	if orgMember.PayName != nil {
		_ = d.Set("pay_name", orgMember.PayName)
	}

	if len(orgMember.Tags) != 0 {
		tags := make(map[string]string, len(orgMember.Tags))
		for _, tag := range orgMember.Tags {
			tags[*tag.TagKey] = *tag.TagValue
		}
		_ = d.Set("tags", tags)
	}

	return nil
}

func resourceTencentCloudOrganizationOrgMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_organization_org_member.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	request := organization.NewMoveOrganizationNodeMembersRequest()
	updateRequest := organization.NewUpdateOrganizationMemberRequest()

	orgMemberId := d.Id()

	request.MemberUin = []*int64{helper.Int64(helper.StrToInt64(orgMemberId))}
	updateRequest.MemberUin = helper.Uint64(helper.StrToUInt64(orgMemberId))
	if d.HasChange("node_id") {
		if v, _ := d.GetOk("node_id"); v != nil {
			request.NodeId = helper.IntInt64(v.(int))
		}
		err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
			result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseOrganizationClient().MoveOrganizationNodeMembers(request)
			if e != nil {
				return tccommon.RetryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
					logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})

		if err != nil {
			log.Printf("[CRITAL]%s create organization orgMember failed, reason:%+v", logId, err)
			return err
		}
	}

	if d.HasChange("name") {
		if v, _ := d.GetOk("name"); v != nil {
			updateRequest.Name = helper.String(v.(string))
		}
	}

	if d.HasChange("remark") {
		if v, _ := d.GetOk("remark"); v != nil {
			updateRequest.Remark = helper.String(v.(string))
		}
	}

	if d.HasChange("policy_type") {
		if v, _ := d.GetOk("policy_type"); v != nil {
			updateRequest.PolicyType = helper.String(v.(string))
		}
		if v, _ := d.GetOk("permission_ids"); v != nil {
			ids := v.(*schema.Set).List()
			for i := range ids {
				id := ids[i].(int)
				updateRequest.PermissionIds = append(updateRequest.PermissionIds, helper.IntUint64(id))
			}
		}
	}

	if d.HasChange("permission_ids") {
		if v, _ := d.GetOk("permission_ids"); v != nil {
			ids := v.(*schema.Set).List()
			for i := range ids {
				id := ids[i].(int)
				updateRequest.PermissionIds = append(updateRequest.PermissionIds, helper.IntUint64(id))
			}
		}
		if v, _ := d.GetOk("policy_type"); v != nil {
			updateRequest.PolicyType = helper.String(v.(string))
		}
	}

	if d.HasChange("is_allow_quit") {
		if v, _ := d.GetOk("is_allow_quit"); v != nil {
			updateRequest.IsAllowQuit = helper.String(v.(string))
		}
	}

	if d.HasChange("record_id") {
		return fmt.Errorf("`record_id` do not support change now.")
	}

	if d.HasChange("pay_uin") {
		if v, _ := d.GetOk("pay_uin"); v != nil {
			updateRequest.PayUin = helper.String(v.(string))
		}
	}

	UpdateErr := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseOrganizationClient().UpdateOrganizationMember(updateRequest)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, updateRequest.GetAction(), updateRequest.ToJsonString(), result.ToJsonString())
		}
		return nil
	})
	if UpdateErr != nil {
		log.Printf("[CRITAL]%s update organization orgMember failed, reason:%+v", logId, UpdateErr)
		return UpdateErr
	}

	if d.HasChange("tags") {
		tcClient := meta.(tccommon.ProviderMeta).GetAPIV3Conn()
		tagService := svctag.NewTagService(tcClient)
		oldTags, newTags := d.GetChange("tags")
		replaceTags, deleteTags := svctag.DiffTags(oldTags.(map[string]interface{}), newTags.(map[string]interface{}))
		resourceName := tccommon.BuildTagResourceName("organization", "member", tcClient.Region, orgMemberId)
		if err := tagService.ModifyTags(ctx, resourceName, replaceTags, deleteTags); err != nil {
			return err
		}
	}
	return resourceTencentCloudOrganizationOrgMemberRead(d, meta)
}

func resourceTencentCloudOrganizationOrgMemberDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_organization_org_member.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := context.WithValue(context.TODO(), tccommon.LogIdKey, logId)

	service := OrganizationService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	orgMemberId := d.Id()

	if err := service.DeleteOrganizationOrgMemberById(ctx, orgMemberId); err != nil {
		return err
	}

	return nil
}
