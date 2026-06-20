#!/bin/bash
# qvh-probe — diagnose ONE phone's QuickTime AV arming state (read-mostly, ~18s).
# Runs qvh record briefly and classifies the result so iteration has a fast signal.
# Usage: qvh-probe.sh <udid> [seconds] [qvh_path]
# States:
#   FRESH-VIDEO        h264 frames flowing (CVRP+FEED)            -> healthy, streamable
#   AUDIO-ONLY-stuck   CWPA/audio but no video                   -> stuck (mode-2 half-open)
#   TIME-ONLY-stuck    only PING + SYNC_TIME, no CWPA            -> stuck (partial arm)
#   NO-AV-IFACE        device on config 4 / no Valeria interface -> not in mode 2
#   NO-PING/dead       no response                               -> dead/contended
set +e
UDID="$1"; SECS="${2:-18}"; QVH="${3:-/root/qvh-test/qvh}"
[ -z "$UDID" ] && { echo "usage: qvh-probe.sh <udid> [seconds]"; exit 1; }
LOG=$(mktemp); H=/tmp/probe_$$.h264; W=/tmp/probe_$$.wav
timeout "$SECS" "$QVH" record "$H" "$W" --udid="$UDID" -v 2>"$LOG"
PING=$(grep -c "PING received" "$LOG"); CWPA=$(grep -c SYNC_CWPA "$LOG")
CVRP=$(grep -c CVRP "$LOG"); TIME=$(grep -c SYNC_TIME "$LOG"); RECYC=$(grep -c "re-cycling\|re-arm" "$LOG")
H264=$(stat -c%s "$H" 2>/dev/null || echo 0); WAV=$(stat -c%s "$W" 2>/dev/null || echo 0)
NOIFACE=$(grep -cE "not activated|Could not retrieve config|MuxConfig 4|device not activated" "$LOG")
if   [ "${H264:-0}" -gt 1000 ];               then STATE="FRESH-VIDEO"
elif [ "${CWPA:-0}" -gt 0 ] || [ "${WAV:-0}" -gt 5000 ]; then STATE="AUDIO-ONLY-stuck"
elif [ "${TIME:-0}" -gt 0 ];                  then STATE="TIME-ONLY-stuck"
elif [ "${NOIFACE:-0}" -gt 0 ];               then STATE="NO-AV-IFACE"
else                                               STATE="NO-PING/dead"; fi
printf 'udid=%s state=%-18s PING=%s CWPA=%s CVRP=%s TIME=%s recyc=%s h264=%s wav=%s\n' \
  "${UDID:0:8}" "$STATE" "$PING" "$CWPA" "$CVRP" "$TIME" "$RECYC" "$H264" "$WAV"
rm -f "$H" "$W" "$LOG"
