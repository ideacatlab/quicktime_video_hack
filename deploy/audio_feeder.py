#!/usr/bin/env python3
# audio_feeder: read qvh's growing .wav (44B header + s16le/48k/stereo PCM) and emit a
# CONTINUOUS real-time PCM stream (real audio when present, silence otherwise) to a fifo
# for a SEPARATE audio ffmpeg. It reads a FILE (never blocks qvh) so the video pipeline is
# fully decoupled and can never stall on audio.
import os, sys, time
wav, out = sys.argv[1], sys.argv[2]
CHUNK = 3840                 # 20 ms of s16le stereo @ 48000 Hz
SILENCE = b'\x00' * CHUNK
HDR = 44
for _ in range(600):
    if os.path.exists(wav):
        break
    time.sleep(0.05)
try:
    f = open(wav, 'rb')
except Exception:
    sys.exit(0)
f.seek(HDR)
try:
    fo = open(out, 'wb', buffering=0)   # blocks until the audio ffmpeg opens the read end
except Exception:
    sys.exit(0)
t = time.monotonic()
while True:
    try:
        sz = os.fstat(f.fileno()).st_size
    except Exception:
        sz = 0
    pos = f.tell()
    if pos > sz:                        # file truncated -> capture restarted
        f.seek(HDR)
    elif sz - pos > 192000:             # >1s behind live -> jump to ~0.4s behind
        f.seek(sz - 76800)
    data = f.read(CHUNK)
    if not data:
        data = SILENCE
    elif len(data) < CHUNK:
        data = data + SILENCE[len(data):]
    try:
        fo.write(data)
    except Exception:
        break
    t += 0.02
    d = t - time.monotonic()
    if d > 0:
        time.sleep(d)
    elif d < -0.1:
        t = time.monotonic()
