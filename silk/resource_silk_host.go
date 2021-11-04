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

func resourceSilkHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSilkHostCreate,
		ReadContext:   resourceSilkHostRead,
		UpdateContext: resourceSilkHostUpdate,
		DeleteContext: resourceSilkHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSilkHostImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Host.",
			},
			"obj_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SDP ID of Host.",
			},
			"host_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of Host.",
			},
			"pwwn": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An optional list of PWWNs that are mapped to the Host.",
			},
			"iqn": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "The IQN that is mapped to the Host.",
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

func resourceSilkHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Read in the resource schema arguments for easier assignment
	name := d.Get("name").(string)
	hostType := d.Get("host_type").(string)
	pwwn := d.Get("pwwn").([]interface{})
	iqn := d.Get("iqn").(string)
	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	host, err := silk.CreateHost(name, hostType, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(pwwn) != 0 {
		for _, p := range pwwn {
			_, err := silk.CreateHostPWWN(name, p.(interface{}).(string))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if iqn != "" {
		_, err := silk.CreateHostIQN(name, iqn)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("silk-host-%d-%s", host.ID, strconv.FormatInt(time.Now().Unix(), 10))) // <-- maybe simply make this the the name or host.ID?

	return resourceSilkHostRead(ctx, d, m)
}

func resourceSilkHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getHost, err := silk.GetHosts(timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gimme an id as an int <-- this breaks creation...
	// hostID, err := strconv.Atoi(d.Id())
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	for _, host := range getHost.Hits {
		if host.Name == d.Get("name").(string) {

			d.Set("name", host.Name)
			d.Set("host_type", host.Type)
			d.Set("obj_id", host.ID)

			if len(d.Get("pwwn").([]interface{})) != 0 {

				// Get the current PWWNs on the host and then set the TF pwwn value with
				// those responses
				pwwns := []string{}
				getPwwn, err := silk.GetHostPWWN(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
				for _, value := range getPwwn {
					pwwns = append(pwwns, value.Pwwn)
				}

				// Sort the new slice to prevent any TF comparison issues
				sort.Slice(pwwns, func(i, j int) bool {
					return pwwns[i] < pwwns[j]
				})

				d.Set("pwwn", pwwns)

			}

			if d.Get("iqn").(string) != "" {

				// Get the current IQNs on the host and then set the TF IQN value with
				// those responses
				iqns := []string{}
				getIQN, err := silk.GetHostIQN(d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
				for _, value := range getIQN {
					iqns = append(iqns, value.Iqn)
				}

				if len(iqns) == 0 {
					d.Set("iqn", "")

				} else {
					d.Set("iqn", iqns[0])
				}

			}

			// Stop the loop and return a nil err
			return diags
		}
	}
	// Volume was not found on the server
	d.SetId("")

	return diags

}

func resourceSilkHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	silk := m.(*silksdp.Credentials)

	timeout := d.Get("timeout").(int)

	config := map[string]interface{}{}
	var hostName string
	if d.HasChange("name") {
		config["name"] = d.Get("name").(string)
		// If the name changed in Terraform, we need to look up the "original" name (i.e what is currently is on the Silk server)
		// to push the new name change to the volume
		currentHostName, _ := d.GetChange("name")
		hostName = currentHostName.(string)
	} else {
		hostName = d.Get("name").(string)
	}

	pwwnToRemove := []string{}
	pwwnToAdd := []string{}
	if d.HasChange("pwwn") {

		// Get the current (c) and new (n) pwwn hosts
		c, n := d.GetChange("pwwn")
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

		// Adding PWWNs
		if len(current) < len(new) {
			// Find all pwwns that have been added to the pwwn slice
			for _, p := range new {
				_, found := find(current, p)
				if !found {
					pwwnToAdd = append(pwwnToAdd, p)
				}
			}

			// Add each PWWN to the Host
			for _, p := range pwwnToAdd {
				_, err := silk.CreateHostPWWN(d.Get("name").(string), p)
				if err != nil {
					return diag.FromErr(err)
				}
			}

		} else {

			// Find all pwwns that have been removed from the pwwn slice
			for _, p := range current {
				_, found := find(new, p)
				if !found {
					pwwnToRemove = append(pwwnToRemove, p)
				}
			}

			// Remove each PWWN from the Host

			for _, p := range pwwnToRemove {
				_, err := silk.DeleteHostIndividualPWWN(d.Get("name").(string), p)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

	}

	if d.HasChange("iqn") {

		c, n := d.GetChange("iqn")

		// There was a change the current value is blank so we know an IQN needs to be added
		if c.(string) == "" {
			_, err := silk.CreateHostIQN(d.Get("name").(string), d.Get("iqn").(string), timeout)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			// There was change and the current IQN is not blank so know we need to either remove
			// or changed the IQN

			// No mater the situation we are going to need to remove
			// the IQN from the Host (i.e in order to change we have to first remove)
			_, err := silk.DeleteHostIQN(d.Get("name").(string), timeout)
			if err != nil {
				return diag.FromErr(err)
			}

			// The new value is not blank so we know its an 'update' scenario
			if n.(string) != "" {
				_, err := silk.CreateHostIQN(d.Get("name").(string), d.Get("iqn").(string), timeout)
				if err != nil {
					return diag.FromErr(err)
				}
			}

		}

	}

	if d.HasChange("host_type") {
		config["type"] = d.Get("host_type").(string)
	}

	// If only the PWWN or IQN changes the config map won't be populated
	// and will throw an error
	if len(config) != 0 {
		_, err := silk.UpdateHost(hostName, config, timeout)
		if err != nil {
			return diag.FromErr(err)
		}

	}

	return resourceSilkHostRead(ctx, d, m)
}

func resourceSilkHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	silk := m.(*silksdp.Credentials)

	_, err := silk.DeleteHost(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceSilkHostImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	timeout := d.Get("timeout").(int)

	silk := m.(*silksdp.Credentials)

	getHost, err := silk.GetHosts(timeout)
	if err != nil {
		return nil, err
	}

	for _, host := range getHost.Hits {
		if host.Name == d.Id() {
			// Set the base data
			d.Set("name", host.Name)
			d.Set("obj_id", host.ID)
			d.Set("host_type", host.Type)
			d.Set("timeout", 15)

			// Check for IQNs
			iqns := []string{}
			getIQN, err := silk.GetHostIQN(d.Get("name").(string))
			if err != nil {
				return nil, err
			}
			for _, value := range getIQN {
				iqns = append(iqns, value.Iqn)
			}

			if len(iqns) == 0 {
				d.Set("iqn", "")

			} else {
				d.Set("iqn", iqns[0])
			}

			// Check for pwwns
			pwwns := []string{}
			getPwwn, err := silk.GetHostPWWN(d.Get("name").(string))
			if err != nil {
				return nil, err
			}
			for _, value := range getPwwn {
				pwwns = append(pwwns, value.Pwwn)
			}

			// Sort the new slice to prevent any TF comparison issues
			sort.Slice(pwwns, func(i, j int) bool {
				return pwwns[i] < pwwns[j]
			})

			d.Set("pwwn", pwwns)

			// Set the ID
			d.SetId(fmt.Sprintf("silk-host-%d-%s", host.ID, strconv.FormatInt(time.Now().Unix(), 10)))
		}
	}

	return []*schema.ResourceData{d}, nil
}
