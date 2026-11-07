#!/bin/bash
# usage: check.sh <tree> <outfile>
cd "$1" || exit 1
export US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0 GOPROXY=off GOSUMDB=off
{
echo "## $(date '+%F %T %Z')  tree=$1"
changed=$(cd "$1"; for f in $(find . -name '*.go'); do cmp -s "$f" "/workspace/us-panel-dev/baseline_tree/$f" || echo "$f"; done)
echo "### gofmt -l (changed/new files): "; gofmt -l $changed; echo "gofmt_done"
echo "### go vet ./..."; go vet ./... 2>&1; echo "vet_exit=$?"
echo "### GOOS=windows go vet ."; GOOS=windows go vet . 2>&1; echo "winvet_exit=$?"
echo "### go test -race -count=1 ./..."; go test -race -count=1 -v ./... > /workspace/us-panel-dev/f01/logs/check_testv.log 2>&1; rc=$?; grep -E '^(ok|FAIL|\?|---)' /workspace/us-panel-dev/f01/logs/check_testv.log | grep -v -- '--- PASS' ; echo "test_exit=$rc"
echo "PASS_count=$(grep -c -- '--- PASS' /workspace/us-panel-dev/f01/logs/check_testv.log) FAIL_count=$(grep -c -- '--- FAIL' /workspace/us-panel-dev/f01/logs/check_testv.log) SKIP_count=$(grep -c -- '--- SKIP' /workspace/us-panel-dev/f01/logs/check_testv.log)"
echo "### GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build"; GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /workspace/us-panel-dev/f01/check.exe . 2>&1; echo "winbuild_exit=$?"
} > "$2" 2>&1
cat "$2"
