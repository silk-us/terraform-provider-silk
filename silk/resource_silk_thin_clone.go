package silk

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

func resourceSilkThinClone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkThinCloneCreate,
		ReadContext:   resourceSilkThinCloneRead,
		UpdateContext: resourceSilkThinCloneUpdate,
		DeleteContext: resourceSilkThinCloneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkThinCloneImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Thin Clone.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of the Thin Clone (lives in /volumes).",
			},
			"volume_group_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Volume Group ID (used to detect Volume Group renames).",
			},
			"volume_group_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume Group the Thin Clone is created in. Must contain the source Volume.",
			},
			"source_snapshot_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Volume Group Snapshot to clone from. Changing this forces a new Thin Clone.",
			},
			"source_volume_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the source Volume within the snapshot. Changing this forces a new Thin Clone.",
			},
			"source_snap_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of the volsnap that backs this Thin Clone.",
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size, in GB, of the Thin Clone (inherited from the source Volume).",
			},
			"vmware": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "VMware support flag (inherited from the source Volume).",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A description of the Thin Clone. Defaults to the API-provided 'Clone of volume <name>'.",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Sets the Thin Clone to Read-Only when true.",
			},
			"allow_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When set to true, this value will allow the Thin Clone to be destroyed through Terraform.",
			},
			"host_mapping": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Hosts the Thin Clone is mapped to.",
			},
			"host_group_mapping": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Host Groups the Thin Clone is mapped to.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15,
				Description: "The number of seconds to wait to establish a connection the Silk server before returning a timeout error.",
			},
			"scsi_sn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The scsi serial number as string.",
			},
		},
	}
}

func resourceSilkThinCloneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	volumeGroupName := d.Get("volume_group_name").(string)
	sourceSnapshotName := d.Get("source_snapshot_name").(string)
	sourceVolumeName := d.Get("source_volume_name").(string)
	hostMapping := d.Get("host_mapping").([]interface{})
	hostGroupMapping := d.Get("host_group_mapping").([]interface{})
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	thinClone, err := silk.CreateThinClone(name, volumeGroupName, sourceSnapshotName, sourceVolumeName, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// description and read_only aren't part of the create payload, so apply them via UpdateVolume if set.
	postCreateConfig := map[string]interface{}{}
	if v, ok := d.GetOk("description"); ok {
		postCreateConfig["description"] = v.(string)
	}
	if d.Get("read_only").(bool) {
		postCreateConfig["read_only"] = true
	}
	if len(postCreateConfig) != 0 {
		if _, err := silk.UpdateVolume(name, postCreateConfig, timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(hostMapping) != 0 {
		for _, h := range hostMapping {
			if _, err := silk.CreateHostVolumeMapping(h.(string), name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if len(hostGroupMapping) != 0 {
		for _, h := range hostGroupMapping {
			if _, err := silk.CreateHostGroupVolumeMapping(h.(string), name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(fmt.Sprintf("silk-thin-clone-%d-%s", thinClone.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkThinCloneRead(ctx, d, m)
}

// pre v1.2.7, full pull + client side filter. kept for reference
// func resourceSilkThinCloneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//
// 	var diags diag.Diagnostics
//
// 	timeout := d.Get("timeout").(int)
// 	silk := m.(*silksdp.Credentials)
//
// 	getVolume, err := silk.GetVolumes(timeout)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}
//
// 	for _, volume := range getVolume.Hits {
// 		if volume.Name == d.Get("name").(string) {
//
// 			volumeGroupRefID, err := strconv.Atoi(strings.Replace(volume.VolumeGroup.Ref, "/volume_groups/", "", 1))
// 			if err != nil {
// 				d.Set("volume_group_name", "")
// 			}
//
// 			getVolumeGroups, err := silk.GetVolumeGroups(timeout)
// 			if err != nil {
// 				return diag.FromErr(err)
// 			}
//
// 			for _, volumeGroup := range getVolumeGroups.Hits {
// 				if volumeGroup.ID == volumeGroupRefID {
// 					d.Set("volume_group_id", volumeGroupRefID)
// 					d.Set("volume_group_name", volumeGroup.Name)
// 				}
// 			}
//
// 			if len(d.Get("host_mapping").([]interface{})) != 0 {
// 				hostsMappedToVolume, err := silk.GetVolumeHostMappings(d.Get("name").(string))
// 				if err != nil {
// 					return diag.FromErr(err)
// 				}
// 				sort.Slice(hostsMappedToVolume, func(i, j int) bool {
// 					return hostsMappedToVolume[i] < hostsMappedToVolume[j]
// 				})
// 				d.Set("host_mapping", hostsMappedToVolume)
// 			}
//
// 			if len(d.Get("host_group_mapping").([]interface{})) != 0 {
// 				hostGroupsMappedToVolume, err := silk.GetVolumeHostGroupMappings(d.Get("name").(string))
// 				if err != nil {
// 					return diag.FromErr(err)
// 				}
// 				sort.Slice(hostGroupsMappedToVolume, func(i, j int) bool {
// 					return hostGroupsMappedToVolume[i] < hostGroupsMappedToVolume[j]
// 				})
// 				d.Set("host_group_mapping", hostGroupsMappedToVolume)
// 			}
//
// 			d.Set("name", volume.Name)
// 			d.Set("obj_id", volume.ID)
// 			d.Set("size_in_gb", volume.Size/1024/1024)
// 			d.Set("vmware", volume.VmwareSupport)
// 			d.Set("description", volume.Description)
// 			d.Set("read_only", volume.ReadOnly)
// 			d.Set("allow_destroy", d.Get("allow_destroy").(bool))
// 			d.Set("scsi_sn", volume.ScsiSn)
//
// 			return diags
// 		}
// 	}
//
// 	d.SetId("")
// 	return diags
// }

func resourceSilkThinCloneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	// by obj_id when we have it so a rename on the array doesnt look like a delete
	var getVolume *silksdp.GetVolumesResponse
	var err error
	if id := objID(d); id != 0 {
		getVolume, err = silk.GetVolumeByID(id, timeout)
	} else {
		getVolume, err = silk.GetVolumeByName(d.Get("name").(string), timeout)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	for _, volume := range getVolume.Hits {
		if volume.ID != objID(d) && volume.Name != d.Get("name").(string) {
			continue
		}

		d.Set("volume_group_name", "")
		if volumeGroupRefID, err := strconv.Atoi(strings.Replace(volume.VolumeGroup.Ref, "/volume_groups/", "", 1)); err == nil {
			getVolumeGroup, err := silk.GetVolumeGroupByID(volumeGroupRefID, timeout)
			if err != nil {
				return diag.FromErr(err)
			}
			for _, volumeGroup := range getVolumeGroup.Hits {
				d.Set("volume_group_id", volumeGroupRefID)
				d.Set("volume_group_name", volumeGroup.Name)
			}
		}

		// one mappings call covers both lists
		wantHosts := len(d.Get("host_mapping").([]interface{})) != 0
		wantHostGroups := len(d.Get("host_group_mapping").([]interface{})) != 0
		if wantHosts || wantHostGroups {
			hostsMappedToVolume, hostGroupsMappedToVolume, err := silk.GetVolumeMappingsByID(volume.ID, timeout)
			if err != nil {
				return diag.FromErr(err)
			}
			// Sort to prevent any TF comparison issues
			if wantHosts {
				sort.Strings(hostsMappedToVolume)
				d.Set("host_mapping", hostsMappedToVolume)
			}
			if wantHostGroups {
				sort.Strings(hostGroupsMappedToVolume)
				d.Set("host_group_mapping", hostGroupsMappedToVolume)
			}
		}

		d.Set("name", volume.Name)
		d.Set("obj_id", volume.ID)
		d.Set("size_in_gb", volume.Size/1024/1024) // Convert to GB
		d.Set("vmware", volume.VmwareSupport)
		d.Set("description", volume.Description)
		d.Set("read_only", volume.ReadOnly)
		d.Set("allow_destroy", d.Get("allow_destroy").(bool))
		d.Set("scsi_sn", volume.ScsiSn)

		return diags
	}
	// Volume was not found on the server
	d.SetId("")

	return diags
}

func resourceSilkThinCloneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)
	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var currentName string

	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		oldName, _ := d.GetChange("name")
		currentName = oldName.(string)
	} else {
		currentName = d.Get("name").(string)
	}

	if d.HasChange("host_mapping") {
		c, n := d.GetChange("host_mapping")
		cReflect := reflect.ValueOf(c)
		nReflect := reflect.ValueOf(n)
		current := []string{}
		new := []string{}
		for i := 0; i < cReflect.Len(); i++ {
			current = append(current, cReflect.Index(i).Interface().(string))
		}
		for i := 0; i < nReflect.Len(); i++ {
			new = append(new, nReflect.Index(i).Interface().(string))
		}

		toAdd := []string{}
		toRemove := []string{}
		for _, h := range unique(append(current, new...)) {
			_, foundInNew := find(new, h)
			_, foundInCurrent := find(current, h)
			if foundInNew && !foundInCurrent {
				toAdd = append(toAdd, h)
			} else if !foundInNew && foundInCurrent {
				toRemove = append(toRemove, h)
			}
		}

		for _, h := range toAdd {
			if _, err := silk.CreateHostVolumeMapping(h, currentName); err != nil {
				return diag.FromErr(err)
			}
		}
		for _, h := range toRemove {
			if _, err := silk.DeleteHostVolumeMapping(h, currentName); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("host_group_mapping") {
		c, n := d.GetChange("host_group_mapping")
		cReflect := reflect.ValueOf(c)
		nReflect := reflect.ValueOf(n)
		current := []string{}
		new := []string{}
		for i := 0; i < cReflect.Len(); i++ {
			current = append(current, cReflect.Index(i).Interface().(string))
		}
		for i := 0; i < nReflect.Len(); i++ {
			new = append(new, nReflect.Index(i).Interface().(string))
		}

		toAdd := []string{}
		toRemove := []string{}
		for _, h := range unique(append(current, new...)) {
			_, foundInNew := find(new, h)
			_, foundInCurrent := find(current, h)
			if foundInNew && !foundInCurrent {
				toAdd = append(toAdd, h)
			} else if !foundInNew && foundInCurrent {
				toRemove = append(toRemove, h)
			}
		}

		for _, hg := range toAdd {
			if _, err := silk.CreateHostGroupVolumeMapping(hg, currentName); err != nil {
				return diag.FromErr(err)
			}
		}
		for _, hg := range toRemove {
			if _, err := silk.DeleteHostGroupVolumeMapping(hg, currentName); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("volume_group_name") {
		volumeGroupID, err := silk.GetVolumeGroupID(d.Get("volume_group_name").(string), timeout)
		if err != nil {
			return diag.FromErr(err)
		}
		cGroupIDInterface, _ := d.GetChange("volume_group_id")
		cGroupID := cGroupIDInterface.(int)
		if volumeGroupID != cGroupID {
			d.Set("volume_group_id", volumeGroupID)
			volumeGroupConfig := map[string]interface{}{}
			volumeGroupConfig["ref"] = fmt.Sprintf("/volume_groups/%d", volumeGroupID)
			config["volume_group"] = volumeGroupConfig
		}
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	if d.HasChange("read_only") {
		config["read_only"] = d.Get("read_only").(bool)
	}

	if len(config) != 0 {
		if _, err := silk.UpdateVolume(currentName, config, timeout); err != nil {
			d.Set("name", currentName)
			return diag.FromErr(err)
		}
	}

	return resourceSilkThinCloneRead(ctx, d, m)
}

func resourceSilkThinCloneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	if d.Get("allow_destroy") == false {
		return diag.Errorf("The `allow_destroy` value is set to false. The Thin Clone can not be destroyed through Terraform")
	}

	name := d.Get("name").(string)
	silk := m.(*silksdp.Credentials)

	currentHostMappings, _ := d.GetChange("host_mapping")
	currentHostMappingsReflect := reflect.ValueOf(currentHostMappings)
	for i := 0; i < currentHostMappingsReflect.Len(); i++ {
		h := currentHostMappingsReflect.Index(i).Interface().(string)
		if _, err := silk.DeleteHostVolumeMapping(h, name); err != nil {
			return diag.FromErr(err)
		}
	}

	currentHostGroupMappings, _ := d.GetChange("host_group_mapping")
	currentHostGroupMappingsReflect := reflect.ValueOf(currentHostGroupMappings)
	for i := 0; i < currentHostGroupMappingsReflect.Len(); i++ {
		hg := currentHostGroupMappingsReflect.Index(i).Interface().(string)
		if _, err := silk.DeleteHostGroupVolumeMapping(hg, name); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, err := silk.DeleteVolume(name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

func resourceSilkThinCloneImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	timeout := d.Get("timeout").(int)
	silk := m.(*silksdp.Credentials)

	getVolume, err := silk.GetVolumeByName(d.Id(), timeout)
	if err != nil {
		return nil, err
	}

	for _, volume := range getVolume.Hits {
		if volume.Name == d.Id() {
			volumeGroupRefID, err := strconv.Atoi(strings.Replace(volume.VolumeGroup.Ref, "/volume_groups/", "", 1))
			if err != nil {
				d.Set("volume_group_name", "")
			}

			// getVolumeGroups, err := silk.GetVolumeGroups(timeout)
			getVolumeGroups, err := silk.GetVolumeGroupByID(volumeGroupRefID, timeout)
			if err != nil {
				return nil, err
			}
			for _, vg := range getVolumeGroups.Hits {
				if vg.ID == volumeGroupRefID {
					d.Set("volume_group_name", vg.Name)
				}
			}

			if hostsMapped, err := silk.GetVolumeHostMappings(d.Id()); err == nil {
				sort.Slice(hostsMapped, func(i, j int) bool { return hostsMapped[i] < hostsMapped[j] })
				d.Set("host_mapping", hostsMapped)
			}
			if hgMapped, err := silk.GetVolumeHostGroupMappings(d.Id()); err == nil {
				sort.Slice(hgMapped, func(i, j int) bool { return hgMapped[i] < hgMapped[j] })
				d.Set("host_group_mapping", hgMapped)
			}

			d.Set("name", volume.Name)
			d.Set("obj_id", volume.ID)
			d.Set("size_in_gb", volume.Size/1024/1024)
			d.Set("vmware", volume.VmwareSupport)
			d.Set("description", volume.Description)
			d.Set("read_only", volume.ReadOnly)
			d.Set("allow_destroy", d.Get("allow_destroy").(bool))
			d.Set("scsi_sn", volume.ScsiSn)
			d.Set("timeout", 15)
			d.SetId(fmt.Sprintf("silk-thin-clone-%d-%s", volume.ID, strconv.FormatInt(time.Now().Unix(), 10)))
		}
	}

	return []*schema.ResourceData{d}, nil
}
