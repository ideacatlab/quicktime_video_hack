#!/bin/bash
# QVH capture + HLS, RELIABLE. Retries (in the qvh binary's StartReading) until the AV
# session is up. Downscales the iPhone's native res to a web-friendly height so HLS is
# smooth and within H.264 levels. Video frames appear when the phone screen is awake.
UDID="$1"
[ -z "$UDID" ] && { echo "usage: capture.sh <udid>"; exit 1; }
FIFO=/root/qvh-stream/qvh.h264
SERVE=/root/qvh-stream/serve
WAV=/root/qvh-stream/qvh.wav
QVH=/root/qvh-test/qvh
mkdir -p "$SERVE"
while true; do
  rm -f "$FIFO" "$WAV"; mkfifo "$FIFO"
  rm -f "$SERVE"/seg_*.ts "$SERVE"/stream.m3u8
  ffmpeg -y -loglevel warning -fflags +genpts -r 30 -f h264 -i "$FIFO" -an \
    -vf "scale=-2:1024,format=yuv420p" \
    -c:v libx264 -preset veryfast -tune zerolatency -profile:v high -g 30 -keyint_min 30 \
    -f hls -hls_time 1 -hls_list_size 8 -hls_flags delete_segments+omit_endlist+independent_segments \
    -hls_segment_filename "$SERVE/seg_%05d.ts" "$SERVE/stream.m3u8" >>/root/qvh-stream/ffmpeg.log 2>&1 &
  FF=$!
  sleep 1
  "$QVH" record "$FIFO" "$WAV" --udid="$UDID" -v >>/root/qvh-stream/qvh.log 2>&1 &
  QV=$!
  ok=0
  for i in $(seq 1 8); do
    sleep 2
    [ "$(stat -c%s "$WAV" 2>/dev/null || echo 0)" -gt 2000 ] && { ok=1; break; }
    kill -0 "$QV" 2>/dev/null || break
  done
  if [ "$ok" = 1 ]; then
    echo "$(date -u) session UP udid=$UDID" >>/root/qvh-stream/loop.log
    wait "$QV"
  else
    echo "$(date -u) no PING, retrying" >>/root/qvh-stream/loop.log
    kill "$QV" 2>/dev/null; wait "$QV" 2>/dev/null
  fi
  kill "$FF" 2>/dev/null; wait "$FF" 2>/dev/null
  sleep 1
done
