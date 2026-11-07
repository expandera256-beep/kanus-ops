#!/bin/bash
b=/workspace/us-panel-dev/scratch/review-20261108/uyum-v41
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0 US_GPU_WAKE_DISABLE=1
r3='^TestReview(BadConfirm|WrongPassword|OriginAllowlist|OriginUnlisted|RetentionOverflow|ParentSymlink|RenameFailure|OpenHelper|MailMatching)'
for v in 11 12;do
 cd "$b/probe$v" || exit 1
 go test -race -json -count=1 -p=1 -timeout 30m ./internal/privacy ./licensesrv ./cmd/us-license-hook -run "$r3" > "$b/old-r3-$v.jsonl" 2>&1; echo "r3-$v $?" >> "$b/old-probe-status.txt"
 go test -race -json -count=1 -p=1 -timeout 30m ./internal/privacy ./licensesrv ./cmd/us-license-hook -run '^TestReviewV4' > "$b/old-35-$v.jsonl" 2>&1; echo "35-original-$v $?" >> "$b/old-probe-status.txt"
 go test -race -json -count=1 -p=1 -timeout 30m ./internal/privacy -run '^TestReviewInvalidOriginPort$' > "$b/old-ports-$v.jsonl" 2>&1; echo "ports-$v $?" >> "$b/old-probe-status.txt"
done
