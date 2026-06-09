# Publishing `terraform-provider-silk` to the Terraform Registry

Step-by-step plan to publish this provider to the **public** registry at
`registry.terraform.io` under the **`silk-us`** namespace. The published source
address will be:

```
silk-us/silk
```

so users consume it as:

```hcl
terraform {
  required_providers {
    silk = {
      source  = "silk-us/silk"
      version = "~> 1.2"
    }
  }
}
```

> Namespace rule: the registry namespace **is** the GitHub org/owner. Repo must
> be `terraform-provider-{NAME}` → `terraform-provider-silk` (already correct).
> The Go module path (`github.com/silk-us/silk-terraform-provider`) does **not**
> need to match — only the repo name does.

---

## Prerequisites (one-time)

- [ ] Admin access to the `silk-us` GitHub org *(confirmed)*.
- [ ] Repo is **public** on GitHub (registry will not index a private repo).
- [ ] `LICENSE` file at repo root *(present)*.
- [ ] `gpg`, `go`, and `git` installed locally.
- [ ] A HashiCorp Cloud / registry account you can log into at
      <https://registry.terraform.io> with the `silk-us` GitHub identity.

---

## Phase 1 — Repo hygiene (do before any release)

1. **Remove the committed binary.** There's an ~18 MB
   `terraform-provider-silk_1.2.4_linux_amd64` at the repo root. Delete it and
   make sure `bin/` and built binaries are git-ignored.
   ```powershell
   git rm terraform-provider-silk_1.2.4_linux_amd64
   ```
   Confirm `.gitignore` covers `bin/`, `*.bak`, and `terraform-provider-silk*`
   build artifacts.

2. **Decide the release version.** Makefile says `VERSION=1.2.6`. The registry
   keys releases off **git tags** of the form `vX.Y.Z` (note the leading `v`).
   Pick the next tag, e.g. `v1.2.6` or `v1.3.0`.

---

## Phase 2 — Registry-format documentation

The registry renders docs from a specific layout. The current `docs/` folder
uses the old `localdomain/provider/silk` source and a flat file naming that the
registry won't pick up cleanly. Two options:

- **Recommended:** generate docs with
  [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs), which
  produces the exact `docs/` structure the registry expects.

  ```powershell
  go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
  tfplugindocs generate
  ```

  Expected layout afterward:
  ```
  docs/
    index.md                       # provider overview (from templates/ or schema)
    resources/
      volume.md
      volume_group.md
      host.md
      host_group.md
      retention_policy.md
      capacity_policy.md
      thin_clone.md
  ```

- **Manual:** restructure the existing `docs/*.md` into `docs/index.md` +
  `docs/resources/<name>.md` yourself.

3. Update every `source` in the docs from `localdomain/provider/silk` to
   `silk-us/silk`.
4. (Optional but nice) add `examples/` with runnable `.tf` snippets;
   `tfplugindocs` will fold them into the rendered pages.

---

## Phase 3 — GoReleaser config

The registry expects a specific set of release artifacts: per-platform zips, a
`SHA256SUMS` file, and a **GPG-signed** `SHA256SUMS.sig`, plus a
`manifest.json`. GoReleaser produces all of this. There is no `.goreleaser.yml`
yet — create one at the repo root.

5. Create `.goreleaser.yml` (use HashiCorp's
   [scaffold template](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.goreleaser.yml)
   as the base). Key requirements:
   - `builds`: `CGO_ENABLED=0`, ldflags `-s -w -X main.version={{.Version}}`,
     and the same `goos`/`goarch` matrix as the Makefile (darwin, linux,
     freebsd, openbsd, solaris amd64, windows).
   - `archives`: `format: zip`, name
     `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}`.
   - `checksum`: `name_template: '{{ .ProjectName }}_{{ .Version }}_SHA256SUMS'`,
     `algorithm: sha256`.
   - `signs`: detach-sign the checksums with GPG (`${signature}` / `${artifact}`,
     `--local-user {{ .Env.GPG_FINGERPRINT }}`).
   - A `terraform-registry-manifest.json` artifact declaring
     `metadata.protocol_versions: ["5.0"]` (this provider uses
     terraform-plugin-sdk/v2 → protocol 5).

6. Add `terraform-registry-manifest.json` at the repo root:
   ```json
   {
     "version": 1,
     "metadata": { "protocol_versions": ["5.0"] }
   }
   ```

---

## Phase 4 — GPG signing key

The registry verifies releases against a GPG public key you upload once.

7. Generate (or reuse) a signing key:
   ```powershell
   gpg --full-generate-key          # RSA 4096, no expiry or a long one
   gpg --list-secret-keys --keyid-format=long
   ```
   Note the **fingerprint** (the long hex string).

8. Export the keys:
   ```powershell
   gpg --armor --export <FINGERPRINT>        > silk-public.gpg   # upload to registry
   gpg --armor --export-secret-keys <FINGERPRINT> > silk-private.gpg # for CI secret
   ```
   Keep the private key out of git.

---

## Phase 5 — GitHub Actions release workflow

9. Create `.github/workflows/release.yml` based on HashiCorp's
   [recommended release workflow](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.github/workflows/release.yml).
   It triggers on tag push `v*`, imports the GPG key, and runs GoReleaser.

10. Add the two repo **secrets** under
    `silk-us/terraform-provider-silk → Settings → Secrets and variables → Actions`:
    - `GPG_PRIVATE_KEY` = contents of `silk-private.gpg`
    - `PASSPHRASE` = the key's passphrase (omit if you made a passphrase-less key)

    And expose the fingerprint to GoReleaser. The standard workflow imports the
    key with `crazy-max/ghaction-import-gpg` and sets `GPG_FINGERPRINT` from its
    output, so no separate secret is needed for the fingerprint.

---

## Phase 6 — Cut the first release

11. Commit Phases 1–5, push to `main`.
12. Tag and push:
    ```powershell
    git tag v1.2.6
    git push origin v1.2.6
    ```
13. Watch the Actions run. On success the GitHub Release for `v1.2.6` should
    contain: per-OS `.zip` files, `_SHA256SUMS`, `_SHA256SUMS.sig`, and the
    `manifest.json`. **If any of those are missing the registry import will
    fail**, so verify the asset list before moving on.

---

## Phase 7 — Register the provider (one-time)

14. Go to <https://registry.terraform.io>, sign in with the `silk-us` GitHub
    identity → **Publish → Provider**.
15. Select the `silk-us/terraform-provider-silk` repo. This adds a GitHub
    webhook so future tags auto-import.
16. Upload the **GPG public key** (`silk-public.gpg`) when prompted (or under
    **User/Org Settings → Signing Keys**). The fingerprint must match what
    signed the release.
17. The registry ingests the existing `v1.2.6` release. After a minute the
    provider appears at:
    ```
    https://registry.terraform.io/providers/silk-us/silk/latest
    ```

---

## Phase 8 — Verify end-to-end

18. In a scratch dir:
    ```hcl
    terraform {
      required_providers {
        silk = {
          source  = "silk-us/silk"
          version = "1.2.6"
        }
      }
    }
    ```
    ```powershell
    terraform init
    ```
    Init should download the signed provider from the registry with no
    `localdomain` override and no manual mirror.

---

## Ongoing releases

Once Phase 7 is done, every new version is just:

```powershell
# bump VERSION in Makefile, commit
git tag v1.3.0
git push origin v1.3.0
```

The webhook + Actions workflow handle the build, sign, and registry import
automatically.

---

## Gotchas specific to this repo

- **Stray binary + bin/** must be gitignored, or GoReleaser's clean-tree check
  (`--clean`) may complain and the repo bloats.
- **Module path mismatch** (`silk-terraform-provider` vs repo
  `terraform-provider-silk`) is fine for the registry but means `main.go`'s
  import path stays `github.com/silk-us/silk-terraform-provider/silk` — don't
  "fix" it to match the repo name or the build breaks.
- **Protocol version is 5**, not 6 — this provider is on
  terraform-plugin-sdk/v2. Setting `protocol_versions: ["6.0"]` in the manifest
  would make `terraform init` reject it.
- **`go 1.14` in go.mod** is ancient; GoReleaser/CI runners default to modern
  Go. Bump the `go` directive (e.g. `go 1.21`) and run `go mod tidy` before the
  first release to avoid toolchain surprises.
- The local `build-local.ps1` / `build-local.sh` and the `localdomain/...`
  install flow are for **dev/testing only** — they're independent of registry
  publishing and can stay as-is.
```
