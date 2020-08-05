package silk

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

func resourceSilkHostGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkHostGroupCreate,
		ReadContext:   resourceSilkHostGroupRead,
		UpdateContext: resourceSilkHostGroupUpdate,
		DeleteContext: resourceSilkHostGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Host Group.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A description of the Host Group.",
			},
			"allow_different_host_types": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     false,
				Description: "Corresponds to the 'Enable mixed host OS types' checkbox in the UI.",
			},
			"host_mapping": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of Hosts that belong to the Host Group.",
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

func resourceSilkHostGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	allowDifferentHostTypes := d.Get("allow_different_host_types").(bool)
	hostMapping := d.Get("host_mapping").([]interface{})
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	hostGroup, err := silk.CreateHostGroup(name, description, allowDifferentHostTypes, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(hostMapping) != 0 {
		for _, h := range hostMapping {
			_, err := silk.CreateHostHostGroupMapping(h.(interface{}).(string), name)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-host-group-%d-%s", hostGroup.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkHostGroupRead(ctx, d, m)
}

func resourceSilkHostGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getHostGroups, err := silk.GetHostGroups(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, hostGroup := range getHostGroups.Hits {
		if hostGroup.Name == d.Get("name").(string) {

			if len(d.Get("host_mapping").([]interface{})) != 0 {

				// Get the hosts in the host group and then set the TF host_mapping value with
				// those responses
				hostsInHostGroup, err := silk.GetHostGroupHosts(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(hostsInHostGroup, func(i, j int) bool {
					return hostsInHostGroup[i] < hostsInHostGroup[j]
				})

				d.Set("host_mapping", hostsInHostGroup)

			}

			d.Set("name", hostGroup.Name)
			d.Set("description", hostGroup.Description)
			d.Set("allow_different_host_types", hostGroup.AllowDifferentHostTypes)

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Volume was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkHostGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	if d.HasChange("name") {
		return diag.Errorf("Host Group names can not be changed.")
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

		// Mapping Hosts
		if len(current) < len(new) {
			// Find all hosts that have been added to the host slice
			for _, h := range new {
				_, found := find(current, h)
				if !found {
					hostMappingToAdd = append(hostMappingToAdd, h)
				}
			}

			// Add each Host to the Volume
			for _, h := range hostMappingToAdd {
				_, err := silk.CreateHostHostGroupMapping(h, d.Get("name").(string))
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
				_, err := silk.DeleteHostHostGroupMapping(h, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

	}

	if d.HasChange("allow_different_host_types") {
		config["allow_different_host_types"] = d.Get("allow_different_host_types").(bool)
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	if len(config) != 0 {
		_, err := silk.UpdateHostGroup(d.Get("name").(string), config, timeout)
		if err != nil {
			return diag.FromErr(err)
		}

	}

	return resourceSilkHostGroupRead(ctx, d, m)
}

func resourceSilkHostGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	_, err := silk.DeleteHostGroup(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
