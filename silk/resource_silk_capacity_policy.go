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

func resourceSilkCapacityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkCapacityPolicyCreate,
		ReadContext:   resourceSilkCapacityPolicyRead,
		UpdateContext: resourceSilkCapacityPolicyUpdate,
		DeleteContext: resourceSilkCapacityPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkCapacityPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Volume Group.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of Capacity Policy.",
			},
			"warningthreshold": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Percentage of used capacity required to trigger a 'warning'.",
			},
			"errorthreshold": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Percentage of used capacity required to trigger an 'error'.",
			},
			"criticalthreshold": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Percentage of used capacity required to trigger a 'critical' alert.",
			},
			/* despite this being a valid parameter, it is always set for 100 despite the provided value, this plays hell with the resource state so I'm forcing it to 100 in the create function.
			"fullthreshold": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Percentage of used capacity required to trigger a 'full' alert.",
			},
			*/
			"snapshotoverheadthreshold": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Percentage of capacity used by snapshots to generate an alert.",
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

func resourceSilkCapacityPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	warningthreshold := d.Get("warningthreshold").(int)
	errorthreshold := d.Get("errorthreshold").(int)
	criticalthreshold := d.Get("criticalthreshold").(int)
	fullthreshold := 100
	snapshotoverheadthreshold := d.Get("snapshotoverheadthreshold").(int)
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	CapacityPolicy, err := silk.CreateCapacityPolicy(name, warningthreshold, errorthreshold, criticalthreshold, fullthreshold, snapshotoverheadthreshold, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-CapacityPolicy-%d-%s", CapacityPolicy.ID, strconv.FormatInt(time.Now().Unix(), 10)))

	return resourceSilkCapacityPolicyRead(ctx, d, m)
}

// resourceSilkCapacityPolicyRead Reads the decllared capacity policy
func resourceSilkCapacityPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getCapacityPolicy, err := silk.GetCapacityPolicy(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, CapacityPolicy := range getCapacityPolicy.Hits {
		if CapacityPolicy.Name == d.Get("name").(string) {

			d.Set("name", CapacityPolicy.Name)
			d.Set("obj_id", CapacityPolicy.ID)
			d.Set("warningthreshold", CapacityPolicy.WarningThreshold)
			d.Set("errorthreshold", CapacityPolicy.ErrorThreshold)
			d.Set("criticalthreshold", CapacityPolicy.CriticalThreshold)
			d.Set("fullthreshold", CapacityPolicy.FullThreshold)
			d.Set("snapshotoverheadthreshold", CapacityPolicy.SnapshotOverheadThreshold)

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Retention Policy was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkCapacityPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	/* This endpoint does not provide a PATCH method, so this was written in waste.

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var CapacityPolicyName string

	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the object
		currentCapacityPolicyName, _ := d.GetChange("name")
		CapacityPolicyName = currentCapacityPolicyName.(string)
	} else {
		CapacityPolicyName = d.Get("name").(string)
	}

	if d.HasChange("warningthreshold") {
		config["warningthreshold"] = d.Get("warningthreshold").(int)
	}

	if d.HasChange("errorthreshold") {
		config["errorthreshold"] = d.Get("errorthreshold").(int)
	}

	if d.HasChange("criticalthreshold") {
		config["criticalthreshold"] = d.Get("criticalthreshold").(int)
	}

	if d.HasChange("fullthreshold") {
		config["fullthreshold"] = d.Get("fullthreshold").(int)
	}

	if d.HasChange("snapshotoverheadthreshold") {
		config["snapshotoverheadthreshold"] = d.Get("snapshotoverheadthreshold").(int)
	}

	_, err := silk.UpdateCapacityPolicy(CapacityPolicyName, config, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	*/

	return resourceSilkCapacityPolicyRead(ctx, d, m)
}

func resourceSilkCapacityPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	_, err := silk.DeleteCapacityPolicy(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
func resourceSilkCapacityPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getCapacityPolicy, err := silk.GetCapacityPolicy(timeout)
	if err != nil {
		return nil, err
	}

	for _, CapacityPolicy := range getCapacityPolicy.Hits {
		if CapacityPolicy.Name == d.Id() {

			d.Set("name", CapacityPolicy.Name)
			d.Set("obj_id", CapacityPolicy.ID)
			d.Set("warningthreshold", CapacityPolicy.WarningThreshold)
			d.Set("errorthreshold", CapacityPolicy.ErrorThreshold)
			d.Set("criticalthreshold", CapacityPolicy.CriticalThreshold)
			d.Set("fullthreshold", CapacityPolicy.FullThreshold)
			d.Set("snapshotoverheadthreshold", CapacityPolicy.SnapshotOverheadThreshold)
			d.Set("timeout", 15)
			d.SetId(fmt.Sprintf("silk-CapacityPolicy-%d-%s", CapacityPolicy.ID, strconv.FormatInt(time.Now().Unix(), 10)))

		}
	}

	return []*schema.ResourceData{d}, nil

}
