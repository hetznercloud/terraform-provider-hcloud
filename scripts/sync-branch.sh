#!/usr/bin/env bash
# Merges <source-ref> into the local `main` via a work branch, opens a PR
# for it with the GitHub CLI, and tries to land it right away - falling
# back to auto-merge, then to just leaving the PR open - so it never pushes
# to `main` directly and never drops a change silently. If the merge
# itself conflicts, opens an issue instead of a PR.
#
# Idempotent: safe to run repeatedly (e.g. from a daily workflow) - skips
# if <source-ref> is already merged into main, or if a PR/issue for it is
# already open.
#
# Usage: scripts/sync-branch.sh <source-ref> <work-branch> <pr-title> <body-file>
#
# Requires: git identity configured, GH_TOKEN/gh auth set up, run from the
# repo root with <source-ref> already fetched locally.
set -euo pipefail

if [[ $# -ne 4 ]]; then
  echo "usage: $0 <source-ref> <work-branch> <pr-title> <body-file>" >&2
  exit 1
fi

source_ref=$1
work_branch=$2
pr_title=$3
body_file=$4

if git merge-base --is-ancestor "$source_ref" main; then
  echo "main is already up to date with $source_ref."
  exit 0
fi

existing_pr=$(gh pr list --state open --head "$work_branch" --json number --jq 'length')
if [[ "$existing_pr" != "0" ]]; then
  echo "A PR is already open for $work_branch, skipping."
  exit 0
fi

git checkout -B "$work_branch" main

if ! git merge "$source_ref" --no-edit; then
  git merge --abort

  issue_title="Manual merge needed: $source_ref into main (conflict)"
  existing_issue=$(gh issue list --state open --search "in:title \"$issue_title\"" --json number --jq 'length')
  if [[ "$existing_issue" != "0" ]]; then
    echo "An issue is already open for the $source_ref conflict, skipping."
    exit 0
  fi

  conflict_body=$(mktemp)
  {
    echo "Automated merge of \`$source_ref\` into \`main\` hit a merge conflict and could not be merged automatically."
    echo
    echo 'Please merge manually:'
    echo '```sh'
    echo 'git checkout main && git pull'
    echo "git merge $source_ref"
    echo '```'
  } >"$conflict_body"

  gh issue create --title "$issue_title" --body-file "$conflict_body"
  exit 0
fi

git push origin "$work_branch" --force-with-lease

pr_url=$(gh pr create \
  --base main \
  --head "$work_branch" \
  --title "$pr_title" \
  --body-file "$body_file")
pr_number=${pr_url##*/}

# Try to land it in main right away. If that's blocked (branch
# protection, required reviews/checks, ...), fall back to enabling
# auto-merge so it lands as soon as requirements are met. If even that
# isn't possible, leave the PR open rather than failing silently - a
# human merges it manually.
if gh pr merge "$pr_number" --merge --delete-branch; then
  echo "Merged PR #$pr_number into main."
elif gh pr merge "$pr_number" --merge --auto --delete-branch; then
  echo "::warning::PR #$pr_number couldn't be merged immediately (branch protection?) - auto-merge enabled, it will merge once requirements are met."
else
  echo "::warning::Couldn't merge or enable auto-merge for PR #$pr_number - left open for manual review."
fi
