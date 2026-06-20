#!/bin/bash
# qvh-reliability — measure AV arming reliability of ONE phone over N attempts.
# Optionally home-presses (wake) before each probe. Reports the state distribution +
# the FRESH-VIDEO success rate. This is the metric to drive iteration on the arming fix.
# Usage: qvh-reliability.sh <udid> [N] [clientPort_for_wake]
set +e
UDID="$1"; N="${2:-5}"; PORT="$3"
DIR="$(cd "$(dirname "$0")" && pwd)"
[ -z "$UDID" ] && { echo "usage: qvh-reliability.sh <udid> [N] [clientPort]"; exit 1; }
declare -A counts; ok=0
echo "=== reliability: ${UDID:0:8}, $N runs ==="
for i in $(seq 1 "$N"); do
  [ -n "$PORT" ] && curl -s --max-time 5 -X POST "http://127.0.0.1:$PORT/wda/homescreen" -H 'content-type: application/json' -d '{}' >/dev/null 2>&1
  R=$("$DIR/qvh-probe.sh" "$UDID" 16)
  S=$(echo "$R" | sed -E 's/.*state=([A-Za-z0-9/()?-]+).*/\1/')
  counts[$S]=$(( ${counts[$S]:-0} + 1 ))
  [ "$S" = "FRESH-VIDEO" ] && ok=$((ok+1))
  echo "  run $i/$N: $R"
  sleep 2
done
echo "=== distribution ==="
for k in "${!counts[@]}"; do printf '  %-18s %s/%s\n' "$k" "${counts[$k]}" "$N"; done
printf '=== FRESH-VIDEO success rate: %s/%s (%d%%) ===\n' "$ok" "$N" $(( ok * 100 / N ))
