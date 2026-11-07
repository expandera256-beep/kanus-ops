#!/bin/bash
set -u
O=/workspace/us-panel-dev/scratch/review-20261108/global08b-unsigned
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0 TMPDIR="$O/tmp" US_TASK_STATE_KEY_FILE="$O/check-key"
mkdir -p "$TMPDIR" "$O/checks"
cd "$O/tree"
run(){ n="$1"; shift; echo START "$n"; "$@" >"$O/checks/$n.log" 2>&1; echo $? >"$O/checks/$n.exit"; echo END "$n"; }
run build go build ./...
run vet go vet ./...
run windows-vet env GOOS=windows go vet ./...
run windows-build env GOOS=windows go build ./...
run gofmt bash -c 'gofmt -l $(rg --files -g "*.go")'
go list ./... > "$O/packages.txt"
while read -r pkg; do
 n=$(echo "$pkg" | tr / _)
 run "race-$n" go test -race -p=1 -count=1 -json "$pkg"
done < "$O/packages.txt"
run s2 bash /workspace/us-panel-dev/scratch/review-20261108/run_s2_persistent.sh "$O/tree" "$O/checks/s2"
run check-ui node check_ui.mjs
run gui-lang node gui_lang_ui_test.mjs
run gui-pricing-order node gui_pricing_order_test.mjs
run flaky-new go test -race -p=1 -count=10 -json -run '^TestSubmitLeavesTheServerRunning$' .
cd "$O/pre08"
run flaky-base go test -race -p=1 -count=10 -json -run '^TestSubmitLeavesTheServerRunning$' .
