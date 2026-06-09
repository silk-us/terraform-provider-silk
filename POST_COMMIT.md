# Releasing a New Version

Once the provider is registered on the Terraform Registry (one-time setup, see
`PUBLISHING.md`), shipping a new version is just a **tag push**. You never touch
the registry web UI again — the webhook handles it.

## The mental model

**Tag = publish.** Everything downstream is automatic:

```
git push v1.2.7  ──▶  GitHub Actions (GoReleaser builds + signs + releases)
                 ──▶  Registry webhook ingests the release
                 ──▶  silk-us/silk shows 1.2.7 as latest (~1 min later)
```

No re-registration, no re-uploading the GPG key, no manual artifact upload. The
signing key and webhook are persistent one-time setup.

---

## Releasing v1.2.7 (step by step)

```powershell
cd C:\Users\Jar\Dropbox\_GitHub\terraform-provider-silk

# 1. make your code changes, then commit + push to master
git add -A
git commit -m "describe what changed"
git push origin master

# 2. (optional) bump VERSION in the Makefile to 1.2.7 to keep local dev builds
#    in sync, then commit that change too. Not required for the registry.

# 3. tag and push the tag — THIS is the only thing that triggers a release
git tag v1.2.7
git push origin v1.2.7

# 4. watch the build so you catch failures early
gh run watch
```

That's the whole flow.

---

## Rules and gotchas

- **Tag format must be `vX.Y.Z`** — leading `v`, semver. `git tag 1.2.7` (no `v`)
  will **not** be picked up as a release.

- **The Makefile `VERSION` is cosmetic for releases.** GoReleaser derives the
  version from the git tag (`{{.Version}}`), not the Makefile. `VERSION` only
  affects local `build-local.ps1` / `make` dev builds. Keep it in sync for
  tidiness, but it does not change what gets published.

- **A failed build publishes nothing.** If GitHub Actions fails (build break,
  bad config), the registry never sees a release for that tag — there's no
  half-published state. Fix the issue and re-tag.

- **Published versions are immutable.** Once a version is live on the registry,
  don't try to overwrite it. To ship a fix, bump to the next patch (e.g.
  `v1.2.8`) rather than re-cutting `v1.2.7`.

- **The manifest is automatic.** As of the `.goreleaser.yml` fix, every release
  attaches `terraform-provider-silk_<version>_manifest.json` alongside the zips.
  Nothing extra to do.

---

## If the SDK changed too

If your provider change depends on a **new feature in `silk-sdp-go-sdk`**, the
SDK must be released **first**, or the provider build will fail (it can't import
a method that isn't in a published SDK tag).

```powershell
# --- SDK repo first ---
cd C:\Users\Jar\Dropbox\_GitHub\silk-sdp-go-sdk
git add <changed files>
git commit -m "describe SDK change"
git push origin master
git tag v1.3.1            # bump the SDK version
git push origin v1.3.1

# --- then the provider ---
cd C:\Users\Jar\Dropbox\_GitHub\terraform-provider-silk
go get github.com/silk-us/silk-sdp-go-sdk@v1.3.1   # bump the dependency
go mod tidy
go build ./...           # verify it resolves and compiles
git add go.mod go.sum <your changes>
git commit -m "bump SDK to v1.3.1, <change>"
git push origin master
git tag v1.2.7
git push origin v1.2.7
```

> The local `replace` directive in `go.mod` stays **commented out** for releases
> — CI must build against the published SDK module, not your local checkout. Use
> `build-local.ps1` if you want to test against the local SDK during development.

---

## Re-cutting a bad tag (rare — prefer bumping instead)

Only if a tag was pushed by mistake and is **not yet live on the registry**:

```powershell
git push --delete origin v1.2.7   # remove the remote tag
git tag -d v1.2.7                 # remove the local tag
gh release delete v1.2.7 --yes    # remove the GitHub release
# fix, then re-tag and push again
```

Avoid this once a version has been ingested by the registry. Bump to the next
patch version instead.

---

## Quick verification after a release

```powershell
gh release view v1.2.7 --json assets --jq '.assets[].name'
```

A healthy release contains: per-platform `.zip` files, `_SHA256SUMS`,
`_SHA256SUMS.sig`, and `_manifest.json`. Then confirm the registry picked it up:

```
https://registry.terraform.io/providers/silk-us/silk/latest
```
