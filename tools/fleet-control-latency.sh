#!/bin/bash
# fleet-control-latency — measure tap latency across attached phones to find the SLOW ones
# (hung/degraded WDA = the real control-lag source; healthy phones are ~0.5s). Sequential,
# uses a harmless status-bar tap. Skips phones with recent user activity if the pod exposes it.
# Usage: fleet-control-latency.sh [max] [slow_threshold_s]   (env POD_ADMIN=http://127.0.0.1:7799)
set +e
POD="${POD_ADMIN:-http://127.0.0.1:7799}"
MAX="${1:-100}"; SLOW="${2:-3}"
mapfile -t ROWS < <(curl -s "$POD/state" | python3 -c "
import json,sys
d=json.load(sys.stdin); v=d.get('devices') or d; v=list(v.values()) if isinstance(v,dict) else v
for x in v:
  s=x.get('state') or {}
  if isinstance(x,dict) and x.get('udid') and s.get('kind')=='attached': print(x['udid'])
")
echo "=== fleet control-latency: ${#ROWS[@]} attached phones (slow>${SLOW}s) ==="
slowlist=""; nslow=0; nfast=0; i=0
for U in "${ROWS[@]}"; do
  i=$((i+1)); [ "$i" -gt "$MAX" ] && break
  T=$( { /usr/bin/time -f "%e" curl -s -o /dev/null --max-time 14 -X POST "$POD/admin/control/tap?udid=$U" -H 'content-type: application/json' -d '{"coords":{"normX":0.5,"normY":0.01}}' >/dev/null; } 2>&1 )
  bad=$(echo "$T>$SLOW" | bc -l 2>/dev/null)
  if [ "$bad" = "1" ]; then nslow=$((nslow+1)); slowlist="$slowlist $U"; flag="SLOW"; else nfast=$((nfast+1)); flag="ok"; fi
  printf '[%2d/%d] %s  %ss  %s\n' "$i" "${#ROWS[@]}" "${U:0:8}" "$T" "$flag"
done
echo "=== summary: $nfast fast, $nslow SLOW ==="
[ -n "$slowlist" ] && echo "SLOW udids:$slowlist"
