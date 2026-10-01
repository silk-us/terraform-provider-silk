package silk

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

func resourceSilkVolumeGroupView() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkVolumeGroupViewCreate,
		ReadContext:   resourceSilkVolumeGroupViewRead,
		UpdateContext: resourceSilkVolumeGroupViewUpdate,
		DeleteContext: resourceSilkVolumeGroupViewDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkVolumeGroupViewImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Short name of the view. The SDP names it volume_group:name. Changing this forces a new View.",
			},
			"snapshot_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Snapshot to make the view from, volumegroup:snapshot or just the short name. Changing this forces a new View.",
			},
			"retention_policy": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Retention Policy attached to the view. /snapshots has no PATCH, so changing this forces a new View.",
			},
			"volume_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Volume Group the view belongs to.",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full volume_group:name the SDP assigns.",
			},
			"is_auto_deleteable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Always created false so the Retention Policy can not delete the view out from under Terraform.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of the View.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15,
				Description: "The number of seconds to wait to establish a connection the Silk server before returning a timeout error.",
			},
		},
	}
}

func resourceSilkVolumeGroupViewCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	snapshotName := d.Get("snapshot_name").(string)
	retentionPolicy := d.Get("retention_policy").(string)
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	// never auto deleteable, same reason as snapshots
	view, err := silk.CreateVolumeGroupView(name, snapshotName, retentionPolicy, false, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("silk-view-%d-%s", view.ID, strconv.FormatInt(time.Now().Unix(), 10)))
	d.Set("obj_id", view.ID)

	return resourceSilkVolumeGroupViewRead(ctx, d, m)
}

func resourceSilkVolumeGroupViewRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	view, err := lookupSnapshot(silk, d, d.Get("full_name").(string), timeout)
	if err != nil {
		return diag.FromErr(err)
	}
	if view == nil {
		// View was not found on the server
		d.SetId("")
		return diags
	}

	// parent snapshot, only overwrite snapshot_name if it doesnt match either form of the parent name
	parent, err := silk.GetVolumeGroupSnapshotByID(refID(view.Source.Ref), timeout)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(parent.Hits) > 0 {
		current := d.Get("snapshot_name").(string)
		if current != parent.Hits[0].Name && current != parent.Hits[0].ShortName {
			d.Set("snapshot_name", parent.Hits[0].Name)
		}
	}

	retentionPolicy, err := retentionPolicyName(silk, view.RetentionPolicy.Ref, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", view.ShortName)
	d.Set("volume_group_name", view.Name[:strings.LastIndex(view.Name, ":")])
	d.Set("retention_policy", retentionPolicy)
	d.Set("full_name", view.Name)
	d.Set("is_auto_deleteable", view.IsAutoDeleteable)
	d.Set("obj_id", view.ID)

	if view.IsAutoDeleteable {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("View %s is auto deleteable", view.Name),
			Detail:   "Someone turned on is_auto_deleteable outside Terraform. The retention policy can now delete it. Taint and re-apply to replace it.",
		})
	}

	return diags
}

// pre PATCH check, /snapshots doesnt support it
// func resourceSilkVolumeGroupViewUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//
// 	timeout := d.Get("timeout").(int)
// 	silk := m.(*silksdp.Credentials)
//
// 	config := map[string]interface{}{}
// 	if d.HasChange("retention_policy") {
// 		config["retention_policy"] = d.Get("retention_policy").(string)
// 	}
// 	if d.Get("is_auto_deleteable").(bool) {
// 		config["is_auto_deleteable"] = false
// 	}
//
// 	if len(config) != 0 {
// 		if _, err := silk.UpdateVolumeGroupSnapshotByID(objID(d), config, timeout); err != nil {
// 			return diag.FromErr(err)
// 		}
// 	}
//
// 	return resourceSilkVolumeGroupViewRead(ctx, d, m)
// }

// no PATCH on /snapshots, only timeout can change in place
func resourceSilkVolumeGroupViewUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceSilkVolumeGroupViewRead(ctx, d, m)
}

func resourceSilkVolumeGroupViewDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	if _, err := silk.DeleteVolumeGroupSnapshotByID(objID(d), timeout); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

// import id is the full volumegroup:view name
func resourceSilkVolumeGroupViewImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	silk := m.(*silksdp.Credentials)

	views, err := silk.GetVolumeGroupSnapshotByName(d.Id(), 15)
	if err != nil {
		return nil, err
	}
	if len(views.Hits) == 0 || !views.Hits[0].IsExposable {
		return nil, fmt.Errorf("No View named '%s', import with the full volumegroup:view name", d.Id())
	}

	view := views.Hits[0]
	d.Set("obj_id", view.ID)
	d.Set("timeout", 15)
	d.SetId(fmt.Sprintf("silk-view-%d-%s", view.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return []*schema.ResourceData{d}, nil
}
