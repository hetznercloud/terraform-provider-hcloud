#!/usr/bin/env bash

set -eux

export TF_LOG_CORE="TRACE"
export TF_LOG_PROVIDER="INFO"
export TF_LOG_PATH="$1/$1.log"
export HCLOUD_TOKEN="<REDACTED>"

terraform apply -backup="$1/$1.tfstate"