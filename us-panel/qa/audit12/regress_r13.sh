#!/usr/bin/env bash
set -eu
DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
exec python3 "$DIR/regress_r13.py" "$@"
