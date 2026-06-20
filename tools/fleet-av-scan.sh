#!/bin/bash
# fleet-av-scan — scan EVERY attached phone on a pod and report its AV arming state.
# Sequential (never hammer usbmuxd in parallel). Gives a fleet-wide picture: how many
# phones can stream fresh vs are stuck vs not in mode-2. Read-mostly.
# Usage: fleet-av-scan.sh [max_phones]  (env: POD_ADMIN=http://127.0.0.1:7799)
set +e
DIR="$(cd "$(dirname "$0")" && pwd)"
POD="${POD_ADMIN:-http://127.0.0.1:7799}"
MAX="${1:-100}"
mapfile -t ROWS < <(curl -s "$POD/state" | python3 -c "
import json,sys
d=json.load(sys.stdin); v=d.get('devices') or d; v=list(v.values()) if isinstance(v,dict) else v
for x in v:
  s=x.get('state') or {}
  if isinstance(x,dict) and x.get('udid') and s.get('kind')=='attached':
    print(x['udid'], s.get('clientPort',''))
")
echo "=== fleet AV scan: ${#ROWS[@]} attached phones ==="
declare -A tally; i=0
for row in "${ROWS[@]}"; do
  i=$((i+1)); [ "$i" -gt "$MAX" ] && break
  set -- $row; U="$1"; P="$2"
  [ -n "$P" ] && curl -s --max-time 4 -X POST "http://127.0.0.1:$P/wda/homescreen" -H 'content-type: application/json' -d '{}' >/dev/null 2>&1
  R=$("$DIR/qvh-probe.sh" "$U" 14)
  S=$(echo "$R" | sed -E 's/.*state=([A-Za-z0-9/()?-]+).*/\1/')
  tally[$S]=$(( ${tally[$S]:-0} + 1 ))
  printf '[%2d/%d] %s\n' "$i" "${#ROWS[@]}" "$R"
done
echo "=== fleet summary ==="
for k in "${!tally[@]}"; do printf '  %-18s %s\n' "$k" "${tally[$k]}"; done
