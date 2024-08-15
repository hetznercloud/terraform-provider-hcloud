#!/usr/bin/env bash

set -eux

BASE="/home/jo/code/github.com/hetznercloud/terraform-provider-hcloud"
export PATH="$BASE/terraform/bin:$PATH"

build_terraform() {
  pushd "$BASE/terraform"
  go build -ldflags "-w -s -X 'github.com/hashicorp/terraform/version.dev=no'" -o bin/ .

  terraform version
  which terraform
  popd
}

run_test() {
  pushd "$BASE"
  export TF_LOG="trace"
  export TF_LOG_PATH_MASK="test-%s.log"
  make TEST="./internal/server" TESTARGS="-run TestServerResource_PrimaryIPTests" testacc
  popd
}

build_terraform || exit 125
run_test

# git bisect start
# git bisect bad v1.9.4
# git bisect good v1.8.5
# git bisect run ../debug/bisect-check.sh

# result:
# 460c7f3933115c3edf670caacd2ffa489ef4eeb8
# https://github.com/hashicorp/terraform/pull/35467
