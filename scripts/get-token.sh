#!/bin/bash

set -e
resp=$(curl --header 'Authorization: Bearer '"$TTS_TOKEN"'' -X POST https://tt-service.hetzner.cloud/token -o resp.json)
if grep -q Unauthorized "resp.json"
then
    echo "Wrong token"
    exit 1

else
    token=$(cat resp.json | jq -r '.token')
    echo $token
fi
