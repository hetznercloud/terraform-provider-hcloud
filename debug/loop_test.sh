#!/usr/bin/env bash

export TF_LOG=trace

TEST="./internal/server"
TESTARGS="-run TestServerResource_PrimaryIPTests"

run() {
  declare -i counter
  counter=1

  testcase() {
    export TF_LOG_PATH_MASK=test-${counter}-%s.log
    make TEST="${TEST}" TESTARGS="${TESTARGS}" testacc
  }

  while testcase; do
    echo "[${counter}] Test executed successfully"
    counter+=1
  done

  echo "[${counter}] Failure! Stopping execution. Logs: logs-${counter}.txt"
}

run
