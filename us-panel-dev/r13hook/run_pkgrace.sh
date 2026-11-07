#!/bin/bash
L=/workspace/us-panel-dev/r13hook/logs; while [ ! -f $L/all.done ]; do sleep 5; done
cd /workspace/us-panel-dev/r13hook/term
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0
: > $L/race_per_pkg.txt
for p in $(go list ./...); do s=$(date +%s); go test -race -count=1 -json $p > $L/race_pkg_$(echo $p|tr / _).json 2>&1; e=$?; echo "$p exit=$e sec=$(( $(date +%s)-s ))" >> $L/race_per_pkg.txt; done
echo DONE >> $L/race_per_pkg.txt
