package silk

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

// objID is the array id from state, 0 if we dont have one yet (first read, older state)
func objID(d *schema.ResourceData) int {
	id, _ := d.Get("obj_id").(int)
	return id
}

// refID pulls the id off a ref like /retention_policies/4
func refID(r string) int {
	id, _ := strconv.Atoi(r[strings.LastIndex(r, "/")+1:])
	return id
}

// lookupSnapshot finds a snapshot or view by obj_id, or by full vg:name before we have one.
// nil means its gone
func lookupSnapshot(silk *silksdp.Credentials, d *schema.ResourceData, fullName string, timeout int) (*silksdp.VolumeGroupSnapshot, error) {
	var snaps *silksdp.GetVolumeGroupSnapshotResponse
	var err error
	if id := objID(d); id != 0 {
		snaps, err = silk.GetVolumeGroupSnapshotByID(id, timeout)
	} else {
		snaps, err = silk.GetVolumeGroupSnapshotByName(fullName, timeout)
	}
	if err != nil {
		return nil, err
	}
	for i := range snaps.Hits {
		// deleted ones can hang around briefly
		if !snaps.Hits[i].IsDeleted {
			return &snaps.Hits[i], nil
		}
	}
	return nil, nil
}

// retentionPolicyName turns a /retention_policies/N ref into the policy name
func retentionPolicyName(silk *silksdp.Credentials, r string, timeout int) (string, error) {
	if r == "" {
		return "", nil
	}
	policies, err := silk.GetRetentionPolicyByID(refID(r), timeout)
	if err != nil || len(policies.Hits) == 0 {
		return "", err
	}
	return policies.Hits[0].Name, nil
}
