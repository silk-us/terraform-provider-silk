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

func resourceSilkVolumeGroupSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkVolumeGroupSnapshotCreate,
		ReadContext:   resourceSilkVolumeGroupSnapshotRead,
		UpdateContext: resourceSilkVolumeGroupSnapshotUpdate,
		DeleteContext: resourceSilkVolumeGroupSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkVolumeGroupSnapshotImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Short name of the snapshot. The SDP names it volume_group_name:name. Changing this forces a new Snapshot.",
			},
			"volume_group_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Volume Group to snapshot. Changing this forces a new Snapshot.",
			},
			"retention_policy": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Retention Policy attached to the snapshot. /snapshots has no PATCH, so changing this forces a new Snapshot.",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full volume_group_name:name the SDP assigns.",
			},
			"is_auto_deleteable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Always created false so the Retention Policy can not delete the snapshot out from under Terraform.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of the Snapshot.",
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

func resourceSilkVolumeGroupSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	volumeGroupName := d.Get("volume_group_name").(string)
	retentionPolicy := d.Get("retention_policy").(string)
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	// never auto deleteable, policy expiring it would poison the state
	snapshot, err := silk.CreateVolumeGroupSnapshot(name, volumeGroupName, retentionPolicy, false, false, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("silk-snapshot-%d-%s", snapshot.ID, strconv.FormatInt(time.Now().Unix(), 10)))
	d.Set("obj_id", snapshot.ID)

	return resourceSilkVolumeGroupSnapshotRead(ctx, d, m)
}

func resourceSilkVolumeGroupSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	fullName := d.Get("volume_group_name").(string) + ":" + d.Get("name").(string)
	snapshot, err := lookupSnapshot(silk, d, fullName, timeout)
	if err != nil {
		return diag.FromErr(err)
	}
	if snapshot == nil {
		// Snapshot was not found on the server
		d.SetId("")
		return diags
	}

	retentionPolicy, err := retentionPolicyName(silk, snapshot.RetentionPolicy.Ref, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", snapshot.ShortName)
	d.Set("volume_group_name", snapshot.Name[:strings.LastIndex(snapshot.Name, ":")])
	d.Set("retention_policy", retentionPolicy)
	d.Set("full_name", snapshot.Name)
	d.Set("is_auto_deleteable", snapshot.IsAutoDeleteable)
	d.Set("obj_id", snapshot.ID)

	if snapshot.IsAutoDeleteable {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Snapshot %s is auto deleteable", snapshot.Name),
			Detail:   "Someone turned on is_auto_deleteable outside Terraform. The retention policy can now delete it. Taint and re-apply to replace it.",
		})
	}

	return diags
}

// pre PATCH check, /snapshots doesnt support it
// func resourceSilkVolumeGroupSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
// 	return resourceSilkVolumeGroupSnapshotRead(ctx, d, m)
// }

// no PATCH on /snapshots, only timeout can change in place
func resourceSilkVolumeGroupSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceSilkVolumeGroupSnapshotRead(ctx, d, m)
}

func resourceSilkVolumeGroupSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	if _, err := silk.DeleteVolumeGroupSnapshotByID(objID(d), timeout); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

// import id is the full volumegroup:snapshot name
func resourceSilkVolumeGroupSnapshotImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	silk := m.(*silksdp.Credentials)

	snaps, err := silk.GetVolumeGroupSnapshotByName(d.Id(), 15)
	if err != nil {
		return nil, err
	}
	if len(snaps.Hits) == 0 {
		return nil, fmt.Errorf("No Snapshot named '%s', import with the full volumegroup:snapshot name", d.Id())
	}

	snapshot := snaps.Hits[0]
	d.Set("obj_id", snapshot.ID)
	d.Set("timeout", 15)
	d.SetId(fmt.Sprintf("silk-snapshot-%d-%s", snapshot.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return []*schema.ResourceData{d}, nil
}
