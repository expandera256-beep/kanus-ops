#!/bin/bash
# Adapted from qa/audit12/run.py (same overlay+test pattern); SOURCE param, logs to qa12/logs, does not overwrite QA qa-raw.
SRC=$1; CASE=$2; TAG=$3
case $CASE in ssrf) P='^TestQAAuditRepairRedirectIsolation$';; file_isolation) P='^TestQAAuditExecSiblingIsolation$';; esac
OV=/workspace/us-panel-dev/qa12/overlay-$TAG.json
echo "{\"Replace\":{\"$SRC/zz_qa_audit12_test.go\":\"/workspace/us-panel/qa/audit12/panel_audit_test.go\"}}" > $OV
cd $SRC && env -i PATH=$PATH HOME=$HOME GOCACHE=$(go env GOCACHE) GOMODCACHE=$(go env GOMODCACHE) GOPROXY=off GOSUMDB=off GOFLAGS=-mod=mod US_SMTP_ENABLE=0 US_REAL_MODEL_E2E=0 US_REAL_MODEL_TEST=0 US_REAL_PROJECT_TEST=0 US_REA_NETWORK_TEST=0 US_DOOR_PROBE=0 \
 go test -overlay $OV . -run "$P" -count=1 -v -timeout=90s > /workspace/us-panel-dev/qa12/logs/$CASE-$TAG.log 2>&1; rc=$?
echo "exit_code=$rc" >> /workspace/us-panel-dev/qa12/logs/$CASE-$TAG.log; grep -E "synthetic|redirect_contacted|^(ok|FAIL|---)" /workspace/us-panel-dev/qa12/logs/$CASE-$TAG.log; echo rc=$rc
