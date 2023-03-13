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

func resourceSilkVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkVolumeCreate,
		ReadContext:   resourceSilkVolumeRead,
		UpdateContext: resourceSilkVolumeUpdate,
		DeleteContext: resourceSilkVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkVolumeImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of Volume.",
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The size, in GB, of the Volume.",
			},
			"volume_group_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "VolumeGroup ID (only used when renaming volumegroup name)",
			},
			"volume_group_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume Group that the Volume should be added to.",
			},
			"vmware": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "This value corresponds to the 'VMware support' checkbox in the UI and specifies whether to enable VMFS.",
			},
			"description": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "A description of the Volume.",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "This value corresponds to the 'Exposure Type' radio button in the UI and specifies whether the volume should be 'Read/Write' or 'Read Only'.",
			},
			"allow_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When set to true, this value will prevent the volume from being destroyed through Terraform.",
			},
			"host_mapping": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Hosts the Volume is mapped to.",
			},
			"host_group_mapping": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Host Groups the Volume is mapped to.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15,
				Description: "The number of seconds to wait to establish a connection the Silk server before returning a timeout error.",
			},
			"scsi_sn": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "The scsi serial number as string.",
			},
		},
	}

}

func resourceSilkVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	sizeInGb := d.Get("size_in_gb").(int)
	volumeGroupName := d.Get("volume_group_name").(string)
	vmware := d.Get("vmware").(bool)
	description := d.Get("description").(string)
	readOnly := d.Get("read_only").(bool)
	hostMapping := d.Get("host_mapping").([]interface{})
	hostGroupMapping := d.Get("host_group_mapping").([]interface{})
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	volume, err := silk.CreateVolume(name, sizeInGb, volumeGroupName, vmware, description, readOnly, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(hostMapping) != 0 {
		for _, h := range hostMapping {
			_, err := silk.CreateHostVolumeMapping(h.(interface{}).(string), name)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if len(hostGroupMapping) != 0 {
		for _, h := range hostGroupMapping {
			_, err := silk.CreateHostGroupVolumeMapping(h.(interface{}).(string), name)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-volume-%d-%s", volume.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkVolumeRead(ctx, d, m)
}

func resourceSilkVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	// name := d.Get("name").(string)

	// getVolume, err := silk.GetVolumeByName(name, timeout)
	getVolume, err := silk.GetVolumes(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, volume := range getVolume.Hits {
		if volume.Name == d.Get("name").(string) {
			// Since the API shows the Volume Group as an ID, we have to strip the ID from the provided ref and then
			// look up the volume name based off of that ID. From there we can run d.Set("volume_group_name")
			volumeGroupRefID, err := strconv.Atoi(strings.Replace(volume.VolumeGroup.Ref, "/volume_groups/", "", 1))
			// If any error occured while getting the volumes volume group id, set the volume group id to blank since we can
			// assume there is not one present
			if err != nil {
				d.Set("volume_group_name", "")
			}

			// Get the current volume groups on the server
			getVolumeGroups, err := silk.GetVolumeGroups(timeout)
			if err != nil {
				return diag.FromErr(err)
			}

			for _, volumeGroup := range getVolumeGroups.Hits {
				if volumeGroup.ID == volumeGroupRefID {
					d.Set("volume_group_id", volumeGroupRefID)
					d.Set("volume_group_name", volumeGroup.Name)
				}
			}

			if len(d.Get("host_mapping").([]interface{})) != 0 {

				// Get the current hosts mapped to the volume then set the TF host_mapping value with
				// those responses
				hostsMappedToVolume, err := silk.GetVolumeHostMappings(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(hostsMappedToVolume, func(i, j int) bool {
					return hostsMappedToVolume[i] < hostsMappedToVolume[j]
				})

				d.Set("host_mapping", hostsMappedToVolume)

			}

			if len(d.Get("host_group_mapping").([]interface{})) != 0 {

				// Get the current hosts mapped to the volume then set the TF host_mapping value with
				// those responses
				hostGroupsMappedToVolume, err := silk.GetVolumeHostGroupMappings(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(hostGroupsMappedToVolume, func(i, j int) bool {
					return hostGroupsMappedToVolume[i] < hostGroupsMappedToVolume[j]
				})

				d.Set("host_group_mapping", hostGroupsMappedToVolume)

			}

			d.Set("name", volume.Name)
			d.Set("obj_id", volume.ID)
			d.Set("size_in_gb", volume.Size/1024/1024) // Convert to GB

			d.Set("vmware", volume.VmwareSupport)
			d.Set("description", volume.Description)
			d.Set("read_only", volume.ReadOnly)
			d.Set("allow_destroy", d.Get("allow_destroy").(bool))
			d.Set("scsi_sn", volume.ScsiSn)

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Volume was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var currentVolumeName string

	if d.HasChange("name") {
		// if d.HasChange("volume_group_name") {
		// 	return diag.Errorf("The volume name cannot be changed while moving to another Volume Group")
		// }

		config["name"] = d.Get("name").(string)

		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the volume
		oldVolumeName, _ := d.GetChange("name")
		currentVolumeName = oldVolumeName.(string)

		// // If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// // to push the new name change to the volume
		// currentVolumeName, _ := d.GetChange("name")
		// volumeName = currentVolumeName.(string)
		// nameChangeConfig := map[string]interface{}{
		// 	"name":d.Get("name").(string),
		// }
		// _, err := silk.UpdateVolume(volumeName, nameChangeConfig, timeout)
		// if err != nil {
		// 	d.Set("name",volumeName)
		// 	return diag.FromErr(err)
		// }
	} else {
		currentVolumeName = d.Get("name").(string)
	}

	hostMappingToRemove := []string{}
	hostMappingToAdd := []string{}
	if d.HasChange("host_mapping") {

		// Get the current (c) and new (n) host mappings
		c, n := d.GetChange("host_mapping")
		// Use reflection to convert c and n into values that can be ranged through
		// without crashing Terraform
		cReflect := reflect.ValueOf(c)
		nReflect := reflect.ValueOf(n)
		// Create new []string{}, that holds the values of c, n, that can be ranged
		// through
		current := []string{}
		new := []string{}
		// Loop through cReflect and append values to current
		for i := 0; i < cReflect.Len(); i++ {
			current = append(current, cReflect.Index(i).Interface().(string))
		}
		// Loop through nReflect and append values to new
		for i := 0; i < nReflect.Len(); i++ {
			new = append(new, nReflect.Index(i).Interface().(string))
		}

		union := unique(append(current, new...))
		for _, h := range union {
			_, foundInNew := find(new, h)
			_, foundInCurrent := find(current, h)
			if foundInNew && !foundInCurrent {
				hostMappingToAdd = append(hostMappingToAdd, h)
			} else if !foundInNew && foundInCurrent {
				hostMappingToRemove = append(hostMappingToRemove, h)
			}
		}

		// Add each Host to the Volume
		for _, h := range hostMappingToAdd {
			_, err := silk.CreateHostVolumeMapping(h, currentVolumeName)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Remove each Host from the Volume
		for _, h := range hostMappingToRemove {
			_, err := silk.DeleteHostVolumeMapping(h, currentVolumeName)
			if err != nil {
				return diag.FromErr(err)
			}
		}

	}

	hostGroupMappingToRemove := []string{}
	hostGroupMappingToAdd := []string{}
	if d.HasChange("host_group_mapping") {

		// Get the current (c) and new (n) host group mappings
		c, n := d.GetChange("host_group_mapping")
		// Use reflection to convert c and n into values that can be ranged through
		// without crashing Terraform
		cReflect := reflect.ValueOf(c)
		nReflect := reflect.ValueOf(n)
		// Create new []string{}, that holds the values of c, n, that can be ranged
		// through
		current := []string{}
		new := []string{}
		// Loop through cReflect and append values to current
		for i := 0; i < cReflect.Len(); i++ {
			current = append(current, cReflect.Index(i).Interface().(string))
		}
		// Loop through nReflect and append values to new
		for i := 0; i < nReflect.Len(); i++ {
			new = append(new, nReflect.Index(i).Interface().(string))
		}

		union := unique(append(current, new...))
		for _, h := range union {
			_, foundInNew := find(new, h)
			_, foundInCurrent := find(current, h)
			if foundInNew && !foundInCurrent {
				hostGroupMappingToAdd = append(hostGroupMappingToAdd, h)
			} else if !foundInNew && foundInCurrent {
				hostGroupMappingToRemove = append(hostGroupMappingToRemove, h)
			}
		}

		// Map each Host Group to the Volume
		for _, hg := range hostGroupMappingToAdd {
			_, err := silk.CreateHostGroupVolumeMapping(hg, currentVolumeName)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Remove each Host Group mapping from the Volume
		for _, hg := range hostGroupMappingToRemove {
			_, err := silk.DeleteHostGroupVolumeMapping(hg, currentVolumeName)
			if err != nil {
				return diag.FromErr(err)
			}
		}

	}

	if d.HasChange("size_in_gb") {
		config["size"] = d.Get("size_in_gb").(int) * 1024 * 1024
	}

	if d.HasChange("volume_group_name") {
		volumeGroupID, err := silk.GetVolumeGroupID(d.Get("volume_group_name").(string), timeout)
		if err != nil {
			return diag.FromErr(err)
		}

		cGroupID_Interface, _ := d.GetChange("volume_group_id")
		cGroupID := cGroupID_Interface.(int)
		if volumeGroupID != cGroupID {
			d.Set("volume_group_id", volumeGroupID)
			volumeGroupConfig := map[string]interface{}{}
			volumeGroupConfig["ref"] = fmt.Sprintf("/volume_groups/%d", volumeGroupID)
			config["volume_group"] = volumeGroupConfig
		}
	}

	if d.HasChange("vmware") {
		// config["vmware_support"] = d.Get("vmware").(bool)
		return diag.Errorf("Updating the `vmware` field is not supported")
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	if d.HasChange("read_only") {
		config["read_only"] = d.Get("read_only").(bool)
	}

	// The update function can be triggered when only the allow_destroy option has been changed.
	// Since that does not require and update to Silk, skip the UpdateVolume() call.
	if len(config) != 0 {
		_, err := silk.UpdateVolume(currentVolumeName, config, timeout)
		if err != nil {
			d.Set("name", currentVolumeName)
			return diag.FromErr(err)
		}

	}

	return resourceSilkVolumeRead(ctx, d, m)
}

func resourceSilkVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.Get("allow_destroy") == false {
		return diag.Errorf("The `allow_destroy` value is set to false. The volume can not be destroyed through Terraform")
	}

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	// Delete host_mappings before remove volume
	currentHostMappings, _ := d.GetChange("host_mapping")
	currentHostMappingsReflect := reflect.ValueOf(currentHostMappings)
	for i := 0; i < currentHostMappingsReflect.Len(); i++ {
		hostMappingToRemove := currentHostMappingsReflect.Index(i).Interface().(string)
		_, err := silk.DeleteHostVolumeMapping(hostMappingToRemove, d.Get("name").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Delete host_group_mappings before remove volume
	currentHostGroupMappings, _ := d.GetChange("host_group_mapping")
	currentHostGroupMappingsReflect := reflect.ValueOf(currentHostGroupMappings)
	for i := 0; i < currentHostGroupMappingsReflect.Len(); i++ {
		hostGroupMappingToRemove := currentHostGroupMappingsReflect.Index(i).Interface().(string)
		_, err := silk.DeleteHostGroupVolumeMapping(hostGroupMappingToRemove, d.Get("name").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := silk.DeleteVolume(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceSilkVolumeImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	name := d.Get("name").(string)

	getVolume, err := silk.GetVolumeByName(name, timeout)
	if err != nil {
		return nil, err
	}

	for _, volume := range getVolume.Hits {
		if volume.Name == d.Id() {
			// Since the API shows the Volume Group as an ID, we have to strip the ID from the provided ref and then
			// look up the volume name based off of that ID. From there we can run d.Set("volume_group_name")
			volumeGroupRefID, err := strconv.Atoi(strings.Replace(volume.VolumeGroup.Ref, "/volume_groups/", "", 1))
			// If any error occured while getting the volumes volume group id, set the volume group id to blank since we can
			// assume there is not one present
			if err != nil {
				d.Set("volume_group_name", "")
			}

			// Get the current volume groups on the server
			getVolumeGroups, err := silk.GetVolumeGroups(timeout)
			if err != nil {
				return nil, err
			}

			for _, volumeGroup := range getVolumeGroups.Hits {
				if volumeGroup.ID == volumeGroupRefID {
					d.Set("volume_group_name", volumeGroup.Name)
				}
			}

			// Get the current hosts mapped to the volume then set the TF host_mapping value with
			// those responses
			hostsMappedToVolume, err := silk.GetVolumeHostMappings(d.Id())
			if err == nil {
				sort.Slice(hostsMappedToVolume, func(i, j int) bool {
					return hostsMappedToVolume[i] < hostsMappedToVolume[j]
				})

				d.Set("host_mapping", hostsMappedToVolume)
			}

			// Get the current hosts mapped to the volume then set the TF host_mapping value with
			// those responses
			hostGroupsMappedToVolume, err := silk.GetVolumeHostGroupMappings(d.Id())
			if err == nil {
				sort.Slice(hostGroupsMappedToVolume, func(i, j int) bool {
					return hostGroupsMappedToVolume[i] < hostGroupsMappedToVolume[j]
				})

				d.Set("host_group_mapping", hostGroupsMappedToVolume)
			}
			// Sort the new slice to prevent any TF comparison issues

			d.Set("name", volume.Name)
			d.Set("obj_id", volume.ID)
			d.Set("size_in_gb", volume.Size/1024/1024) // Convert to GB

			d.Set("vmware", volume.VmwareSupport)
			d.Set("description", volume.Description)
			d.Set("read_only", volume.ReadOnly)
			d.Set("allow_destroy", d.Get("allow_destroy").(bool))
			d.Set("scsi_sn", volume.ScsiSn)
			d.Set("timeout", 15)
			d.SetId(fmt.Sprintf("silk-volume-%d-%s", volume.ID, strconv.FormatInt(time.Now().Unix(), 10)))

			// Stop the loop and return a nil err
		}
	}

	return []*schema.ResourceData{d}, nil
}
