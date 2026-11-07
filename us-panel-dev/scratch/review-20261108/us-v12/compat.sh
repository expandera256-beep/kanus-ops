#!/bin/bash
b=/workspace/us-panel-dev/scratch/review-20261108/us-v12
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0 US_GPU_WAKE_DISABLE=1 US_TASK_STATE_KEY_FILE=$b/compat.key
cd "$b/round7/term" && PANEL_COMPAT_MODE=generate PANEL_COMPAT_OUT=$b/old-licenses.json go test -count=1 -p=1 -v -run '^TestPanelCompatGenerate$' ./licensesrv > "$b/compat-generate.log" 2>&1; echo "old-generate $?"
cd "$b/probes" && PANEL_COMPAT_MODE=roundtrip PANEL_COMPAT_IN=$b/old-licenses.json PANEL_COMPAT_OUT=$b/new-licenses.json go test -count=1 -p=1 -v -run '^TestPanelCompatRoundtrip$' ./licensesrv > "$b/compat-forward.log" 2>&1; echo "old-to-new $?"
cd "$b/round7/term" && PANEL_COMPAT_MODE=roundtrip PANEL_COMPAT_IN=$b/new-licenses.json PANEL_COMPAT_OUT=$b/back-licenses.json go test -count=1 -p=1 -v -run '^TestPanelCompatRoundtrip$' ./licensesrv > "$b/compat-back.log" 2>&1; echo "new-to-old $?"
