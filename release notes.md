# v1.2.7

## New resources

* **`silk_volume_group_snapshot`**: take a snapshot of a Volume Group.
* **`silk_volume_group_view`**: create a writable view of a snapshot.

Snapshots and views created by Terraform are always created with `is_auto_deleteable = false`, so a Retention Policy can't expire something Terraform is tracking. The Silk server doesn't support modifying snapshots or views, so changing any argument other than `timeout` replaces them. `silk_thin_clone` can now clone from a snapshot managed in the same configuration. See the resource docs for examples.

## Fixes

* **Arrays with more than 100 objects.** Lookups only ever returned the first 100 objects of each type. On larger arrays this made Terraform think objects past the first 100 had been deleted, and plans tried to re-create them. Every lookup now returns the full set.
* **Long object names.** Looking up an object by name could hang the API when the name was long. Names are now matched exactly.
* **Name lookups could match the wrong object.** A lookup for `vol1` could also match `vol10`. Matching is now exact.
* **`terraform import`.** Importing a volume, volume group, host or host group only worked if the object was among the first 100 of its type. Import now looks the object up directly.
* **Renames outside Terraform.** Resources now track their object by its SDP ID. If an object is renamed on the array, the next plan renames it back instead of treating it as deleted.
* **Thin clone snapshot names.** `source_snapshot_name` now accepts the short snapshot name as well as the full `{volume group}:{snapshot}` name.
* **Destroying a host that's already gone.** Destroy no longer crashes when the host was already removed outside Terraform.
* **Empty API responses.** These now return an error instead of crashing the provider.
* **Faster refresh.** Reads ask the API for just the object they need instead of pulling every object and searching through it. The fixed one-second pause that came after every API call now only happens after creates. Refreshing a volume mapped to two hosts and a host group goes from 9 API calls to 5.

## Upgrade notes

* Existing state files work as-is. On the first refresh, resources that didn't record their SDP ID yet (`obj_id`), such as retention policies, pick it up.
* `silk_capacity_policy`: set `snapshotoverheadthreshold` to a value from 1 to 97. If it is omitted, `0` is sent, and the Silk server rejects `0`.
* Object names on the Silk server are limited to 32 characters.
