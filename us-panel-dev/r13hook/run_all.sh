#!/bin/bash
cd /workspace/us-panel-dev/r13hook/term; L=/workspace/us-panel-dev/r13hook/logs
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0
go test -p=1 -count=1 -json ./... > $L/test.json 2>&1; echo TEST_EXIT=$? > $L/test.exit
go test -race -p=1 -count=1 -json ./... > $L/race.json 2>&1; echo RACE_EXIT=$? > $L/race.exit
go test -race ./licensesrv -run 'TestR13Burst|TestR12ConcurrentWakeCoalescesStatus' -count=20 -v > $L/burst_count20.txt 2>&1; echo BURST_EXIT=$? > $L/burst.exit
bash /workspace/us-panel-dev/panel-evidence/review/regression/run_s2_regression.sh /workspace/us-panel-dev/r13hook/term $L/s2 > $L/s2.log 2>&1; echo S2_EXIT=$? > $L/s2.exit
echo ALLDONE > $L/all.done
