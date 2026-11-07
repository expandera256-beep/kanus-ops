#!/bin/bash
# Resume notes: server part done (licensesrv/chat_settings.go + hook in model_proxy.go).
# Remaining: tests, client chat_settings.go, gui.html panel, privacy line, diff, report.
cd /workspace/us-panel-dev/chatset/term && for p in . ./licensesrv ./cmd/...; do go test -p=1 -count=1 $p; done
