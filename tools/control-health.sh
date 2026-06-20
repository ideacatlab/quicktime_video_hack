#!/bin/bash
# control-health — keep WDA/testmanagerd WARM on attached phones so control taps stay fast,
# monitor tap latency, and recover phones that are persistently slow. Non-invasive: warming is
# a screenshot (same RPC the MJPEG loop uses); recovery is the pod's own detach->re-attach,
# and ONLY after N consecutive slow samples (conservative — VAs are live).
# Why: idle phones with no MJPEG viewer let testmanagerd go cold -> first command stalls ~11s;
# a periodic screenshot keeps it warm so taps stay ~0.5s. Measured: fleet 27/28 already fast.
# Env: POD_ADMIN, WARM_INTERVAL (s), PROBE_EVERY (cycles), SLOW_S, RECOVER_AFTER, RECOVER(0/1)
set +e
POD="${POD_ADMIN:-http://127.0.0.1:7799}"
WARM_INTERVAL="${WARM_INTERVAL:-22}"
PROBE_EVERY="${PROBE_EVERY:-5}"      # probe latency every Nth warm cycle
SLOW_S="${SLOW_S:-3}"
RECOVER_AFTER="${RECOVER_AFTER:-3}"  # consecutive slow probes before recovery
RECOVER="${RECOVER:-0}"              # 0 = log only (safe default), 1 = auto detach/re-attach
LOG=/root/qvh-stream/control-health.log
declare -A slowstreak
cycle=0
udids(){ curl -s "$POD/state" | python3 -c "
import json,sys
d=json.load(sys.stdin); v=d.get('devices') or d; v=list(v.values()) if isinstance(v,dict) else v
for x in v:
  s=x.get('state') or {}
  if isinstance(x,dict) and x.get('udid') and s.get('kind')=='attached': print(x['udid'])"; }
while true; do
  cycle=$((cycle+1)); warmed=0
  for U in $(udids); do
    curl -s -o /dev/null --max-time 10 "$POD/admin/screenshot?udid=$U" && warmed=$((warmed+1))
  done
  msg="$(date -u +%H:%M:%S) cycle=$cycle warmed=$warmed"
  if [ $((cycle % PROBE_EVERY)) -eq 0 ]; then
    slow=""
    for U in $(udids); do
      T=$( { /usr/bin/time -f "%e" curl -s -o /dev/null --max-time 14 -X POST "$POD/admin/control/tap?udid=$U" -H 'content-type: application/json' -d '{"coords":{"normX":0.5,"normY":0.01}}' >/dev/null; } 2>&1 )
      if [ "$(echo "$T>$SLOW_S" | bc -l 2>/dev/null)" = "1" ]; then
        slow="$slow ${U:0:8}(${T}s)"; slowstreak[$U]=$(( ${slowstreak[$U]:-0} + 1 ))
        if [ "$RECOVER" = "1" ] && [ "${slowstreak[$U]}" -ge "$RECOVER_AFTER" ]; then
          curl -s -o /dev/null -X POST "$POD/admin/detach?udid=$U"; echo "$(date -u +%H:%M:%S) RECOVER $U (slow x${slowstreak[$U]})" >> "$LOG"; slowstreak[$U]=0
        fi
      else slowstreak[$U]=0; fi
    done
    msg="$msg probe:[${slow:- all-fast}]"
  fi
  echo "$msg" >> "$LOG"
  sleep "$WARM_INTERVAL"
done
