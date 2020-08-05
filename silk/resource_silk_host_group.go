package silk

import (
	"context"
	"fmt"
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
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	hostGroup, err := silk.CreateHostGroup(name, description, allowDifferentHostTypes, timeout)
	if err != nil {
		return diag.FromErr(err)
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

	if d.HasChange("allow_different_host_types") {
		config["allow_different_host_types"] = d.Get("allow_different_host_types").(bool)
	}

	if d.HasChange("description") {
		config["description"] = d.Get("description").(string)
	}

	_, err := silk.UpdateHostGroup(d.Get("name").(string), config, timeout)
	if err != nil {
		return diag.FromErr(err)
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
