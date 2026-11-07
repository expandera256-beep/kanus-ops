#!/bin/bash
set -u
cd /workspace/us-panel-dev/f01/term || exit 1
export GOPROXY=off GOSUMDB=off US_SMTP_ENABLE=0 US_REAL_MODEL_E2E=0 US_REAL_MODEL_TEST=0 US_REAL_PROJECT_TEST=0 US_REA_NETWORK_TEST=0 US_DOOR_PROBE=0
go test -json -count=1 ./... > /workspace/us-panel-dev/f01/logs/resume_test.jsonl 2>&1 || exit $?
bash /workspace/us-panel-dev/f01/check_f01.sh "$PWD" /workspace/us-panel-dev/f01/logs/resume_check.log
bash /workspace/us-panel-dev/f01/s2_f01/run_s2_regression.sh "$PWD" /workspace/us-panel-dev/f01/logs/s2_resume
