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

func resourceSilkVolumeGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkVolumeGroupCreate,
		ReadContext:   resourceSilkVolumeGroupRead,
		UpdateContext: resourceSilkVolumeGroupUpdate,
		DeleteContext: resourceSilkVolumeGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkVolumeGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume Group.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of Volume Group.",
			},
			"quota_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The size quota, in GB, of the Volume Group. The default option of 0 corresponds to an Unlimited Quota.",
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
				Description: "A description of the Volume Group.",
			},
			"capacity_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default_vg_capacity_policy",
				Description: "The capacity threshold policy profile for the Volume Group.",
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
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	volumeGroup, err := silk.CreateVolumeGroup(name, quotaInGb, enableDeDuplication, description, capacityPolicy, timeout)
	if err != nil {
		return diag.FromErr(err)
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

	// name := d.Get("name").(string)

	// getVolumeGroup, err := silk.GetVolumeGroupByName(name, timeout)
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

			d.Set("name", volumeGroup.Name)
			d.Set("obj_id", volumeGroup.ID)

			// When the Volume Group is set to an unlimated quota
			// the API will return a nil interface which causes
			// an error to be thrown when trying to convert from
			//  the usual float64
			if fmt.Sprintf("%T", volumeGroup.Quota) == "float64" {
				d.Set("quota_in_gb", volumeGroup.Quota.(float64)/1024/1024)
			} else {
				d.Set("quota_in_gb", 0)
			}

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
	var currentVolumeGroupName string

	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the volume
		oldVolumeGroupName, _ := d.GetChange("name")
		currentVolumeGroupName = oldVolumeGroupName.(string)
	} else {
		currentVolumeGroupName = d.Get("name").(string)
	}

	if d.HasChange("quota_in_gb") {
		config["quota"] = d.Get("quota_in_gb").(int) * 1024 * 1024
	}

	if d.HasChange("enable_deduplication") {
		return diag.Errorf("enable_deduplication can not be updated after the initial configuration")
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	if d.HasChange("capacity_policy") {
		config["capacityPolicy"] = d.Get("capacity_policy").(bool)
	}

	_, err := silk.UpdateVolumeGroup(currentVolumeGroupName, config, timeout)
	if err != nil {
		d.Set("name", currentVolumeGroupName)
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
func resourceSilkVolumeGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	name := d.Get("name").(string)

	getVolumeGroup, err := silk.GetVolumeGroupByName(name, timeout)
	if err != nil {
		return nil, err
	}

	for _, volumeGroup := range getVolumeGroup.Hits {
		if volumeGroup.Name == d.Id() {

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
						return nil, err
					}

					d.Set("capacity_policy", capacityPolicyName)
				}
			} else {
				d.Set("capacity_policy", "default_vg_capacity_policy")
			}

			d.Set("name", volumeGroup.Name)
			d.Set("obj_id", volumeGroup.ID)

			// When the Volume Group is set to an unlimated quota
			// the API will return a nil interface which causes
			// an error to be thrown when trying to convert from
			//  the usual float64
			if fmt.Sprintf("%T", volumeGroup.Quota) == "float64" {
				d.Set("quota_in_gb", volumeGroup.Quota.(float64)/1024/1024)
			} else {
				d.Set("quota_in_gb", 0)
			}

			d.Set("enable_deduplication", volumeGroup.IsDedup)
			d.Set("description", volumeGroup.Description)
			d.Set("timeout", 15)
			d.SetId(fmt.Sprintf("silk-volumeGroup-%d-%s", volumeGroup.ID, strconv.FormatInt(time.Now().Unix(), 10)))

			// Stop the loop and return a nil err
		}
	}

	return []*schema.ResourceData{d}, nil

}
