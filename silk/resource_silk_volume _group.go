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

func resourceSilkVolumeGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkVolumeGroupCreate,
		ReadContext:   resourceSilkVolumeGroupRead,
		UpdateContext: resourceSilkVolumeGroupUpdate,
		DeleteContext: resourceSilkVolumeGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume Group..",
			},
			"quota_in_gb": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The size quota, in GB, of the Volume Group.",
			},
			"enable_deduplication": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "This value corresponds to 'Provisioning Type' in the UI. When set to true, the Provisioning Type will be 'thin provisioning with dedupe'.",
			},
			"description": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "A description of the Volume Group..",
			},
			"capacity_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default_vg_capacity_policy",
				Description: "The capacity threshold policy profile for the Volume Group.",
			},
			"host_mapping": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Hosts the Volume Group is mapped to.",
			},
			"host_group_mapping": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Host Groups the Volume Group is mapped to.",
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

func resourceSilkVolumeGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	quotaInGb := d.Get("quota_in_gb").(int)
	enableDeDuplication := d.Get("enable_deduplication").(bool)
	description := d.Get("description").(string)
	capacityPolicy := d.Get("capacity_policy").(string)
	hostMapping := d.Get("host_mapping").([]interface{})
	hostGroupMapping := d.Get("host_group_mapping").([]interface{})
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	volumeGroup, err := silk.CreateVolumeGroup(name, quotaInGb, enableDeDuplication, description, capacityPolicy, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(hostMapping) != 0 {
		for _, h := range hostMapping {
			_, err := silk.CreateHostVolumeGroupMapping(h.(interface{}).(string), name)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if len(hostGroupMapping) != 0 {
		for _, h := range hostGroupMapping {
			_, err := silk.CreateHostGroupVolumeGroupMapping(h.(interface{}).(string), name)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-volumeGroup-%d-%s", volumeGroup.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkVolumeGroupRead(ctx, d, m)
}

func resourceSilkVolumeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getVolumeGroup, err := silk.GetVolumeGroups(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, volumeGroup := range getVolumeGroup.Hits {
		if volumeGroup.Name == d.Get("name").(string) {

			// If the Volume Group has a capacity policy, parse the output for the capacity ID and then convert that to the
			// policy name
			if volumeGroup.CapacityPolicy != nil {
				capacityPolicy := volumeGroup.CapacityPolicy.(map[string]interface{})
				for _, value := range capacityPolicy {
					capacityPolicyID, _ := strconv.Atoi(strings.Replace(value.(string), "/vg_capacity_policies/", "", 1))
					// If an err is returned, we can assume the capacity policy is not present
					if err != nil {
						d.Set("capacity_policy", "")

					}

					capacityPolicyName, err := silk.GetCapacityPolicyName(capacityPolicyID, timeout)
					if err != nil {
						if strings.Contains(err.Error(), "The server does not contain") == true {
							d.Set("capacity_policy", "")
						}
						return diag.FromErr(err)
					}

					d.Set("capacity_policy", capacityPolicyName)
				}
			}

			if len(d.Get("host_mapping").([]interface{})) != 0 {

				// Get the current hosts mapped to the volume then set the TF host_mapping value with
				// those responses
				hostsMappedToVolumeGroup, err := silk.GetVolumeGroupHostMappings(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(hostsMappedToVolumeGroup, func(i, j int) bool {
					return hostsMappedToVolumeGroup[i] < hostsMappedToVolumeGroup[j]
				})

				d.Set("host_mapping", hostsMappedToVolumeGroup)

			}

			if len(d.Get("host_group_mapping").([]interface{})) != 0 {

				// Get the current hosts mapped to the volume then set the TF host_mapping value with
				// those responses
				hostGroupsMappedToVolumeGroup, err := silk.GetVolumeGroupHostGroupMappings(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(hostGroupsMappedToVolumeGroup, func(i, j int) bool {
					return hostGroupsMappedToVolumeGroup[i] < hostGroupsMappedToVolumeGroup[j]
				})

				d.Set("host_group_mapping", hostGroupsMappedToVolumeGroup)

			}

			d.Set("name", volumeGroup.Name)
			d.Set("quota_in_gb", volumeGroup.Quota.(float64)/1024/1024)
			d.Set("enable_deduplication", volumeGroup.IsDedup)
			d.Set("description", volumeGroup.Description)

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Volume Group was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkVolumeGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var volumeGroupName string

	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the volume
		currentVolumeGroupName, _ := d.GetChange("name")
		volumeGroupName = currentVolumeGroupName.(string)
	} else {
		volumeGroupName = d.Get("name").(string)
	}

	hostMappingToRemove := []string{}
	hostMappingToAdd := []string{}
	if d.HasChange("host_mapping") {

		// Get the current (c) and new (n) pwwn hosts
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

		// Mapping Hosts
		if len(current) < len(new) {
			// Find all pwwns that have been added to the pwwn slice
			for _, h := range new {
				_, found := find(current, h)
				if !found {
					hostMappingToAdd = append(hostMappingToAdd, h)
				}
			}

			// Add each Host to the Volume
			for _, h := range hostMappingToAdd {
				_, err := silk.CreateHostVolumeGroupMapping(h, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
			}

		} else {

			// Find all Hosts that have been removed from the host_mapping slice
			for _, h := range current {
				_, found := find(new, h)
				if !found {
					hostMappingToRemove = append(hostMappingToRemove, h)
				}
			}

			// Remove each Host from the Volume
			for _, h := range hostMappingToRemove {
				_, err := silk.DeleteHostVolumeGroupMapping(h, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

	}

	hostGroupMappingToRemove := []string{}
	hostGroupMappingToAdd := []string{}
	if d.HasChange("host_group_mapping") {

		// Get the current (c) and new (n) pwwn hosts
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

		// Mapping Host Group
		if len(current) < len(new) {
			// Find all Host Groups that have been added to the host_group_mapping slice
			for _, hg := range new {
				_, found := find(current, hg)
				if !found {
					hostGroupMappingToAdd = append(hostGroupMappingToAdd, hg)
				}
			}

			// Map each Host Group to the Volume
			for _, hg := range hostGroupMappingToAdd {
				_, err := silk.CreateHostGroupVolumeGroupMapping(hg, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
			}

		} else {

			// Find all Host Group that have been removed from the host_group_mapping slice
			for _, hg := range current {
				_, found := find(new, hg)
				if !found {
					hostGroupMappingToRemove = append(hostGroupMappingToRemove, hg)
				}
			}

			// Remove each Host Group mapping from the Volume
			for _, hg := range hostGroupMappingToRemove {
				_, err := silk.DeleteHostGroupVolumeGroupMapping(hg, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

			}
		}

	}

	if d.HasChange("quota_in_gb") {
		config["quota"] = d.Get("quota_in_gb").(int) * 1024 * 1024
	}

	if d.HasChange("enable_deduplication") {
		config["is_dedupe"] = d.Get("enable_deduplication").(bool)
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	if d.HasChange("capacity_policy") {
		config["capacityPolicy"] = d.Get("capacity_policy").(bool)
	}

	_, err := silk.UpdateVolumeGroup(volumeGroupName, config, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSilkVolumeGroupRead(ctx, d, m)
}

func resourceSilkVolumeGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	_, err := silk.DeleteVolumeGroup(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
