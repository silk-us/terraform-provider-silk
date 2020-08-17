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

func resourceSilkRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkRetentionPolicyCreate,
		ReadContext:   resourceSilkRetentionPolicyRead,
		UpdateContext: resourceSilkRetentionPolicyUpdate,
		DeleteContext: resourceSilkRetentionPolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Volume Group.",
			},
			"num_snapshots": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"weeks": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Number of weeks to retain the snapshot.",
			},
			"days": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Number of days to retain the snapshot.",
			},
			"hours": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Number of hours to retain the snapshot.",
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

func resourceSilkRetentionPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	numSnapshots := d.Get("num_snapshots").(string)
	weeks := d.Get("weeks").(string)
	days := d.Get("days").(string)
	hours := d.Get("hours").(string)
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	RetentionPolicy, err := silk.CreateRetentionPolicy(name, numSnapshots, weeks, days, hours, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-RetentionPolicy-%d-%s", RetentionPolicy.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkRetentionPolicyRead(ctx, d, m)
}

func resourceSilkRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getRetentionPolicy, err := silk.GetRetentionPolicy(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, RetentionPolicy := range getRetentionPolicy.Hits {
		if RetentionPolicy.Name == d.Get("name").(string) {

			d.Set("name", RetentionPolicy.Name)
			d.Set("num_snapshots", RetentionPolicy.num_snapshots)
			d.Set("weeks", RetentionPolicy.weeks)
			d.Set("days", RetentionPolicy.days)
			d.Set("hours", RetentionPolicy.hours)

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Retention Policy was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkRetentionPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var RetentionPolicyName string

	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the volume
		currentRetentionPolicyName, _ := d.GetChange("name")
		RetentionPolicyName = currentRetentionPolicyName.(string)
	} else {
		RetentionPolicyName = d.Get("name").(string)
	}

	if d.HasChange("num_snapshots") {
		config["num_snapshots"] = d.Get("num_snapshots").(string)
	}

	if d.HasChange("weeks") {
		config["weeks"] = d.Get("weeks").(string)
	}

	if d.HasChange("days") {
		config["days"] = d.Get("days").(string)
	}

	if d.HasChange("hours") {
		config["hours"] = d.Get("hours").(string)
	}

	_, err := silk.UpdateRetentionPolicy(RetentionPolicyName, config, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSilkRetentionPolicyRead(ctx, d, m)
}

func resourceSilkRetentionPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	_, err := silk.DeleteRetentionPolicy(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
