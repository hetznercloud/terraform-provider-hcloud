# Maintaining this fork

This is a fork of [hetznercloud/terraform-provider-hcloud](https://github.com/hetznercloud/terraform-provider-hcloud)
that carries an extra feature not in upstream, developed on the
`allow-rebuilding-with-image-or-user-data` branch. This document describes
how the fork is kept in sync with upstream and how releases are versioned
and cut.

## Remotes

```
origin    git@github.com:svanrossem/terraform-provider-hcloud.git   (this fork)
upstream  https://github.com/hetznercloud/terraform-provider-hcloud.git
```

If `upstream` isn't configured:

```sh
git remote add upstream https://github.com/hetznercloud/terraform-provider-hcloud.git
```

## Syncing with upstream

`.github/workflows/sync-upstream.yml` merges `upstream/main` into `main`
automatically every day (and on manual dispatch). It opens a PR via the
GitHub CLI (`scripts/sync-branch.sh`) and tries to merge it right away;
if that's blocked (e.g. branch protection) it enables auto-merge instead,
and if even that isn't possible it just leaves the PR open - either way
it never pushes to `main` directly or drops a change silently. A merge
conflict opens an issue instead of a PR.

To do it by hand instead:

```sh
git fetch upstream --tags
git checkout main
git merge upstream/main   # or upstream/vX.Y.Z for a specific tag
```

`main` should always be a fast-forward from upstream - it never carries
the fork's own feature branch (`allow-rebuilding-with-image-or-user-data`
below); that only gets merged onto release tags, as described next.

## Cutting fork releases automatically

`.github/workflows/cut-fork-release.yml` watches upstream's tags (polled
daily, since GitHub Actions can't watch another repo's tags directly). For
every upstream release tag it hasn't seen yet, it merges the fork's
feature branch onto that tag and pushes the result as a new tag whose
patch number is upstream's patch plus a fixed offset:

```
upstream tag: v1.69.0
fork tag:     v1.69.900
```

If the fork's feature branch is later re-merged onto the same upstream
tag (e.g. after resolving a conflict by hand), the next one becomes
`v1.69.901`, and so on - see `scripts/cut-fork-tag.sh`. On first run
(no fork tags exist yet), it only bootstraps from the latest upstream tag
rather than backfilling this fork's entire release history.

The merge commit is also pushed to a rolling `fork-release` branch so
it's inspectable on GitHub, but that branch is force-pushed on every run
- only the tags are permanent. A merge conflict opens an issue instead of
failing silently; resolve it manually per the issue's instructions.

Pushing the fork tag triggers `.github/workflows/release.yml` (GoReleaser)
exactly like a real upstream tag would, publishing a GitHub Release with
the binaries, checksums, signature, and `terraform-registry-manifest.json`.
Fork builds are Linux-only (`amd64`/`arm64`) - see `.goreleaser.yml` - since
that's the only platform this fork is consumed on; adjust `goos`/`goarch`
there if that changes.

### Versioning scheme

Fork tags are ordinary `vX.Y.Z` tags - the patch number is upstream's
patch plus `PATCH_OFFSET` (900, set in both `cut-fork-release.yml` and
defaulted in `cut-fork-tag.sh`), e.g. upstream `v1.69.0` becomes fork
`v1.69.900`. Unlike the semver build-metadata suffix (`+n`) this used to
use, this is a real, strictly-increasing version bump, which is required
for the Terraform Registry to recognize the tag as a release at all (it
doesn't parse `+build` metadata as a valid provider version) and for
`>=` version constraints to resolve to it as newer than the upstream base
it's built from.

The large offset is what keeps a fork tag from ever colliding with a
genuine upstream release: Hetzner would need to ship 900 patch releases
under the same minor version before `v1.69.900` became a real upstream
tag. Because the tag name itself no longer encodes which upstream tag it
was cut from once re-cuts increment past the offset (`v1.69.901`, ...),
`cut-fork-tag.sh` records that in the tag's annotation instead
(`Fork release of v1.69.0`) - `cut-fork-release.yml` reads that back to
know which upstream tags are already forked.

### Cutting a release by hand

```sh
git fetch upstream --tags
git fetch origin allow-rebuilding-with-image-or-user-data
scripts/cut-fork-tag.sh v1.69.0   # replace with the upstream tag to build on
```

This merges the feature branch onto the given upstream tag, and pushes
both the `fork-release` branch and the resulting offset-patch tag - the
same thing the workflow above does automatically.

Add a short entry to `FORK_CHANGELOG.md` before tagging, summarizing what
changed since the last fork release (upstream's own `CHANGELOG.md` is left
untouched by upstream merges and continues to describe upstream's
history).

## One-time setup still needed (not automatable from here)

- **`SYNC_PAT` secret**: both `sync-upstream.yml` and
  `cut-fork-release.yml` fall back to the default `GITHUB_TOKEN` if this
  isn't set, but GitHub deliberately does not run other workflows off of
  pushes/tags made with the default token (to avoid recursive workflow
  chains). That's merely inconvenient for `sync-upstream.yml` (the sync
  PR's own CI checks won't run), but it's a hard blocker for
  `cut-fork-release.yml`: the fork release tag it pushes would silently
  never trigger `release.yml`/GoReleaser. Add a personal access token
  (classic, `repo` scope, or a fine-grained token scoped to this repo with
  contents/pull-requests/issues read-write) as the `SYNC_PAT` secret so
  tag pushes actually trigger releases.
- **GPG signing key**: generate a key dedicated to this fork and add it as
  the `GPG_PRIVATE_KEY` / `GPG_PASSPHRASE` secrets on this repo (Settings →
  Secrets and variables → Actions). Upstream's key is Hetzner's and isn't
  usable here.
- **Terraform Registry publishing**: sign in to
  [registry.terraform.io](https://registry.terraform.io) with GitHub and
  publish this repo as a provider. `main.go`'s registry address is already
  set to `registry.terraform.io/svanrossem/hcloud` (update both together
  if you ever rename). The registry picks up new tags automatically after
  that; no further action needed per release. Note that only Linux
  binaries are published (see "Cutting fork releases automatically"
  above), so this provider is only installable on Linux.
- `.github/workflows/releaser-pleaser.yml` is upstream's automated
  changelog/version-bump bot; it's gated to only run on
  `hetznercloud/terraform-provider-hcloud` and is a no-op here on purpose
  since it doesn't know about the fork's versioning scheme. Leave it
  disabled rather than re-enabling it.
