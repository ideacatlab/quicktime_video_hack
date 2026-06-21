#!/usr/bin/env python3
# Tail-follow QVH's growing audio .wav (s16le/48k/stereo, 44-byte RIFF header) and emit a
# CONTINUOUS real-time s16le stream to FED (silence-padded when no new audio). Reads the FILE
# with raw os.read (follows appends past EOF — no buffered-EOF caching), starting at the live
# end so it stays in sync. Separate process: if FED's reader (audio ffmpeg) dies, only audio stops.
import os, sys, time
WAV, FED = sys.argv[1], sys.argv[2]
RATE, CH, BPS = 48000, 2, 2
TICK = 0.02                                  # 20ms
CHUNK = int(RATE * CH * BPS * TICK)          # 3840 bytes / 20ms
SILENCE = b'\x00' * CHUNK
MAXBACKLOG = CHUNK * 10                       # cap ~200ms to avoid drift

while not os.path.exists(WAV):
    time.sleep(0.05)
fd = os.open(WAV, os.O_RDONLY)
size = os.fstat(fd).st_size
start = max(44, (size // 4) * 4)             # start at the LIVE end, 4-byte aligned, past header
os.lseek(fd, start, os.SEEK_SET)

fed = open(FED, 'wb', buffering=0)           # blocks until the audio-ffmpeg opens the read end
buf = b''
next_t = time.monotonic()
while True:
    try:
        while True:
            d = os.read(fd, 65536)
            if not d:
                break
            buf += d
    except OSError:
        pass
    if len(buf) >= CHUNK:
        out, buf = buf[:CHUNK], buf[CHUNK:]
        if len(buf) > MAXBACKLOG:            # stay live: drop stale backlog
            buf = buf[-MAXBACKLOG:]
    else:
        out, buf = buf + SILENCE[:CHUNK - len(buf)], b''
    try:
        fed.write(out)
    except (BrokenPipeError, OSError):
        break
    next_t += TICK
    s = next_t - time.monotonic()
    if s > 0:
        time.sleep(s)
    else:
        next_t = time.monotonic()
