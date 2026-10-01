# CRUD test for the provider. Same idea as Test-SDPCrud.ps1 in the PS sdk.
#
#   cd Test
#   terraform init
#   terraform apply -var stage=1      # create
#   terraform plan  -var stage=1      # must say No changes, proves Read
#   terraform apply -var stage=2      # update everything updatable, incl a volume rename
#   terraform plan  -var stage=2      # must say No changes again
#
#   # import check, should come back clean
#   terraform state rm silk_volume.vol02
#   terraform import -var stage=2 silk_volume.vol02 <tag>-vol02-<suffix>   (suffix is in the outputs)
#   terraform plan  -var stage=2
#
#   # snapshot import uses the full vg:snap name (snapshot_full_name output)
#   terraform state rm silk_volume_group_snapshot.snap
#   terraform import -var stage=2 silk_volume_group_snapshot.snap <tag>-vg-<suffix>:snap01-<suffix>
#   terraform plan  -var stage=2
#
#   terraform destroy -var stage=2
#
# Snapshot, view and thin clone all hang off the vg this plan makes. They stay put in stage 2,
# /snapshots has no PATCH so theres nothing to update in place.

terraform {
  required_providers {
    silk = {
      source  = "localdomain/provider/silk"
      version = "1.2.7"
    }
    random = {
      source = "hashicorp/random"
    }
  }
}

provider "silk" {
  server   = "10.10.10.10"
  username = "admin"
  password = "password"
}

# ---------------------------------------------------------------- vars

variable "stage" {
  description = "1 = create, 2 = update"
  type        = number
  default     = 1
}

variable "tag" {
  description = "Prefix on everything this test makes, so its easy to find and clean up"
  type        = string
  default     = "tf127"
}

variable "test_long_name" {
  description = "Make a host with a name right at the 32 char limit. This is what used to hang name__contains"
  type        = bool
  default     = true
}

variable "test_fc" {
  description = "Make an FC host with pwwns. Off by default, the lab array is iscsi only"
  type        = bool
  default     = false
}

# variable "clone_volume_group" {
#   description = "VG that owns the snapshot to thin clone from. Blank skips the thin clone test"
#   type        = string
#   default     = ""
# }

# variable "clone_snapshot" {
#   description = "Snapshot to clone from, short name or vg:snap"
#   type        = string
#   default     = ""
# }

# variable "clone_source_volume" {
#   description = "Volume inside that snapshot to clone"
#   type        = string
#   default     = ""
# }

# random bit on the end of every name so reruns dont collide with leftovers. lives in state so it holds across stages
resource "random_string" "sfx" {
  length  = 4
  upper   = false
  special = false
}

locals {
  s2  = var.stage >= 2
  sfx = random_string.sfx.result
  # long_name = "${var.tag}-host-with-a-deliberately-long-name-01"
  # array caps names at 32 chars, so go right up to it
  # long_name = substr("${var.tag}-host-with-a-long-name-0000000000000000", 0, 32)
  long_name = "${substr("${var.tag}-host-with-a-long-name-0000000000000000", 0, 27)}-${local.sfx}"
}

# ---------------------------------------------------------------- policies

resource "silk_retention_policy" "rp" {
  name          = "${var.tag}-rp-${local.sfx}"
  num_snapshots = local.s2 ? "7" : "5"
  weeks         = "0"
  days          = local.s2 ? "2" : "1"
  hours         = "0"
}

# was for the snapshot PATCH test, still made so theres a 2nd policy on the array
resource "silk_retention_policy" "rp2" {
  name          = "${var.tag}-rp2-${local.sfx}"
  num_snapshots = "3"
  weeks         = "0"
  days          = "1"
  hours         = "0"
}

resource "silk_capacity_policy" "cp" {
  name              = "${var.tag}-cp-${local.sfx}"
  warningthreshold  = 70
  errorthreshold    = 80
  criticalthreshold = 90
  # leaving this out sends 0, array wants 1-97
  snapshotoverheadthreshold = 50
  # fullthreshold is pinned to 100 by the array and everything else here is ForceNew, so no stage 2 change
}

# ---------------------------------------------------------------- volume group

resource "silk_volume_group" "vg" {
  name            = "${var.tag}-vg-${local.sfx}"
  description     = local.s2 ? "tf crud test, updated" : "tf crud test"
  quota_in_gb     = local.s2 ? 200 : 100
  capacity_policy = silk_capacity_policy.cp.name
}

# ---------------------------------------------------------------- hosts

resource "silk_host" "host01" {
  name      = "${var.tag}-host01-${local.sfx}"
  host_type = "Linux"
  iqn       = local.s2 ? "iqn.2026-10.com.silk:${var.tag}-host01-${local.sfx}-b" : "iqn.2026-10.com.silk:${var.tag}-host01-${local.sfx}"
}

# resource "silk_host" "host02" {
#   name      = "${var.tag}-host02-${local.sfx}"
#   host_type = "Linux"
#   pwwn      = local.s2 ? ["20:00:00:25:b5:7f:00:01", "20:00:00:25:b5:7f:00:02"] : ["20:00:00:25:b5:7f:00:01"]
# }
# lab array has no FC, host02 is iscsi now and the pwwn test moved to fchost
resource "silk_host" "host02" {
  name      = "${var.tag}-host02-${local.sfx}"
  host_type = "Windows"
  iqn       = "iqn.1991-05.com.microsoft:${var.tag}-host02-${local.sfx}"
}

resource "silk_host" "fchost" {
  count     = var.test_fc ? 1 : 0
  name      = "${var.tag}-fchost-${local.sfx}"
  host_type = "Linux"
  pwwn      = local.s2 ? ["20:00:00:25:b5:7f:00:01", "20:00:00:25:b5:7f:00:02"] : ["20:00:00:25:b5:7f:00:01"]
}

resource "silk_host" "longname" {
  count     = var.test_long_name ? 1 : 0
  name      = local.long_name
  host_type = "Linux"
}

resource "silk_host_group" "hg" {
  name        = "${var.tag}-hg-${local.sfx}"
  description = local.s2 ? "tf crud test, updated" : "tf crud test"
  # stage 2 adds the long name host, goes thru GetHostByName on the way in
  host_mapping = local.s2 && var.test_long_name ? sort([silk_host.host02.name, silk_host.longname[0].name]) : [silk_host.host02.name]
}

# ---------------------------------------------------------------- volumes

resource "silk_volume" "vol01" {
  # rename in stage 2, read by obj_id should follow it
  name              = local.s2 ? "${var.tag}-vol01r-${local.sfx}" : "${var.tag}-vol01-${local.sfx}"
  description       = local.s2 ? "tf crud test, updated" : "tf crud test"
  size_in_gb        = local.s2 ? 20 : 10
  volume_group_name = silk_volume_group.vg.name
  host_mapping      = [silk_host.host01.name]
  allow_destroy     = true
}

resource "silk_volume" "vol02" {
  name               = "${var.tag}-vol02-${local.sfx}"
  description        = "tf crud test"
  size_in_gb         = 10
  volume_group_name  = silk_volume_group.vg.name
  host_group_mapping = [silk_host_group.hg.name]
  # stage 2 adds a host mapping next to the host group one
  host_mapping  = local.s2 ? [silk_host.host01.name] : []
  allow_destroy = true
}

# # ---------------------------------------------------------------- thin clone (optional)
#
# resource "silk_thin_clone" "clone" {
#   count                = var.clone_snapshot != "" ? 1 : 0
#   name                 = "${var.tag}-clone01-${local.sfx}"
#   volume_group_name    = var.clone_volume_group
#   source_snapshot_name = var.clone_snapshot
#   source_volume_name   = var.clone_source_volume
#   host_mapping         = [silk_host.host01.name]
#   allow_destroy        = true
# }

# ---------------------------------------------------------------- snapshot, view, thin clone

# always created with is_auto_deleteable = false so the policy cant expire it under us
resource "silk_volume_group_snapshot" "snap" {
  name              = "snap01-${local.sfx}"
  volume_group_name = silk_volume_group.vg.name
  # retention_policy  = local.s2 ? silk_retention_policy.rp2.name : silk_retention_policy.rp.name
  # no PATCH on /snapshots, a policy change would replace the snap + view + clone
  retention_policy = silk_retention_policy.rp.name
  # snap the vg once the volumes are in it
  depends_on = [silk_volume.vol01, silk_volume.vol02]
}

resource "silk_volume_group_view" "view" {
  name          = "view01-${local.sfx}"
  snapshot_name = silk_volume_group_snapshot.snap.full_name
  # retention_policy = local.s2 ? silk_retention_policy.rp2.name : silk_retention_policy.rp.name
  retention_policy = silk_retention_policy.rp.name
}

# clones vol02 since vol01 gets renamed in stage 2 and source_volume_name is ForceNew
resource "silk_thin_clone" "clone" {
  name                 = "${var.tag}-clone01-${local.sfx}"
  volume_group_name    = silk_volume_group.vg.name
  source_snapshot_name = silk_volume_group_snapshot.snap.name
  source_volume_name   = silk_volume.vol02.name
  host_mapping         = [silk_host.host01.name]
  allow_destroy        = true
}

# ---------------------------------------------------------------- what read back

output "snapshot_full_name" {
  value = silk_volume_group_snapshot.snap.full_name
}

output "view_full_name" {
  value = silk_volume_group_view.view.full_name
}

output "suffix" {
  value = local.sfx
}

output "obj_ids" {
  value = {
    rp     = silk_retention_policy.rp.obj_id
    cp     = silk_capacity_policy.cp.obj_id
    vg     = silk_volume_group.vg.obj_id
    host01 = silk_host.host01.obj_id
    host02 = silk_host.host02.obj_id
    hg     = silk_host_group.hg.obj_id
    vol01  = silk_volume.vol01.obj_id
    vol02  = silk_volume.vol02.obj_id
    snap   = silk_volume_group_snapshot.snap.obj_id
    view   = silk_volume_group_view.view.obj_id
    clone  = silk_thin_clone.clone.obj_id
  }
}

output "vol01" {
  value = {
    name         = silk_volume.vol01.name
    size_in_gb   = silk_volume.vol01.size_in_gb
    scsi_sn      = silk_volume.vol01.scsi_sn
    volume_group = silk_volume.vol01.volume_group_name
  }
}
