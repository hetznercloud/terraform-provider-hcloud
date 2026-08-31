#!/usr/bin/env bash
# Cuts a fork release for a given upstream tag: merges the fork's feature
# branch ($FEATURE_BRANCH) onto that tag and pushes the result as a new
# tag whose patch number is upstream's patch plus $PATCH_OFFSET, e.g.
# upstream v1.69.0 -> fork v1.69.900 (v1.69.901, ... if re-cut later under
# the same base). This keeps fork tags valid, ordinary-looking semver -
# required for the Terraform Registry to pick them up at all, unlike the
# build-metadata (`+n`) suffix this used to use - while the large offset
# keeps them out of the range Hetzner would plausibly ever use for a real
# patch release on the same minor version.
#
# The exact base upstream tag a fork tag was cut from is recorded in the
# fork tag's annotation ("Fork release of vX.Y.Z"), since the tag name
# alone doesn't preserve it once re-cuts increment past the offset. See
# cut-fork-release.yml for where that's read back.
#
# The merge commit is also pushed to a rolling branch ($RELEASE_BRANCH
# below, default "fork-release") so it's inspectable on GitHub, but that
# branch is reset (force-pushed) on every run - only the tags are
# permanent history.
#
# Pushing the tag triggers .github/workflows/release.yml (GoReleaser),
# which builds and publishes it exactly like an upstream release, just
# with the fork's feature merged in.
#
# On a merge conflict, opens an issue instead of failing loudly, and
# skips. Idempotent: safe to rerun - does nothing if the fork tag for
# this upstream tag already exists or an issue for it is already open.
#
# Usage: scripts/cut-fork-tag.sh <upstream-tag>
#
# Requires: git identity configured, GH_TOKEN/gh auth set up, run from the
# repo root with <upstream-tag> and $FEATURE_BRANCH already fetched
# locally (as refs/upstream-tags/<upstream-tag> or a plain tag - see
# cut-fork-release.yml for why the workflow uses the former).
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <upstream-tag>" >&2
  exit 1
fi

upstream_tag=$1
: "${FEATURE_BRANCH:?FEATURE_BRANCH env var must be set}"
release_branch=${RELEASE_BRANCH:-fork-release}
patch_offset=${PATCH_OFFSET:-900}

if [[ ! "$upstream_tag" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "error: $upstream_tag doesn't look like a plain vX.Y.Z tag" >&2
  exit 1
fi
major_minor="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"
upstream_patch="${BASH_REMATCH[3]}"

if git rev-parse -q --verify "refs/upstream-tags/$upstream_tag" >/dev/null; then
  base_ref="refs/upstream-tags/$upstream_tag"
elif git rev-parse -q --verify "refs/tags/$upstream_tag" >/dev/null; then
  base_ref="refs/tags/$upstream_tag"
else
  echo "error: tag $upstream_tag not found (did you fetch upstream --tags?)" >&2
  exit 1
fi

# Next available patch number under this base tag's offset range - starts
# at upstream_patch + offset, and skips past any already used by a fork
# tag annotated as cut from this same upstream tag (e.g. a previous
# conflict was resolved manually as .900, so a later automated retry
# becomes .901).
last_patch=$(git for-each-ref "refs/tags/v${major_minor}.*" --format='%(refname:short)|%(contents:subject)' \
  | grep -F "|Fork release of ${upstream_tag}" \
  | cut -d'|' -f1 \
  | sed -E "s/^v${major_minor}\.//" \
  | sort -n | tail -1)
next_patch=$(( ${last_patch:-$(( upstream_patch + patch_offset - 1 ))} + 1 ))
fork_tag="v${major_minor}.${next_patch}"

if git rev-parse -q --verify "refs/tags/$fork_tag" >/dev/null; then
  echo "$fork_tag already exists, skipping."
  exit 0
fi

git checkout -B "$release_branch" "$base_ref"

if ! git merge "origin/$FEATURE_BRANCH" --no-edit; then
  git merge --abort

  issue_title="Manual fork release needed: $FEATURE_BRANCH conflicts with $upstream_tag"
  existing_issue=$(gh issue list --state open --search "in:title \"$issue_title\"" --json number --jq 'length')
  if [[ "$existing_issue" != "0" ]]; then
    echo "An issue is already open for the $upstream_tag conflict, skipping."
    exit 0
  fi

  body_file=$(mktemp)
  {
    echo "Automated merge of \`$FEATURE_BRANCH\` onto upstream tag \`$upstream_tag\` hit a merge conflict - could not cut a fork release automatically."
    echo
    echo 'Please resolve and cut the release manually:'
    echo '```sh'
    echo "git fetch upstream --tags && git fetch origin $FEATURE_BRANCH"
    echo "git checkout -B $release_branch $upstream_tag"
    echo "git merge origin/$FEATURE_BRANCH"
    echo '# resolve conflicts, then:'
    echo "git push origin $release_branch --force-with-lease"
    echo "git tag -a ${fork_tag} -m \"Fork release of ${upstream_tag}\""
    echo "git push origin ${fork_tag}"
    echo '```'
  } >"$body_file"

  gh issue create --title "$issue_title" --body-file "$body_file"
  exit 0
fi

git push origin "$release_branch" --force-with-lease
git tag -a "$fork_tag" -m "Fork release of $upstream_tag"
git push origin "$fork_tag"

echo "Cut fork release $fork_tag."
