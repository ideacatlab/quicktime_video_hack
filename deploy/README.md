# ALI / ideacatlab deploy — QVH → HLS → Cloudflare, with WDA coexistence

End-to-end pipeline proven live on a Linux iPhone farm (iOS 16.7, jailbroken),
streaming the phone screen as hardware H.264 over USB to a browser via Cloudflare,
WHILE WDA/Appium controls the same phone.

## Prereqs (per pod)
- `usbmuxd` (master build) with **`USBMUXD_DEFAULT_DEVICE_MODE=2`** so devices sit on
  the Valeria config (5) which carries BOTH usbmux/control (0xFE) and AV (0x2A).
  Drop-in: `/etc/systemd/system/usbmuxd.service.d/zz-qvh-devmode.conf` →
  `[Service]\nEnvironment=USBMUXD_DEFAULT_DEVICE_MODE=2`, then `systemctl restart usbmuxd`.
- `ffmpeg`, `go` (build qvh from this repo), `cloudflared`.

## Pieces
- `qvh` (this repo) `record <fifo> <wav> --udid=X` → H.264 to a fifo. The
  StartReading re-cycle retry re-arms the AV session reliably in mode 2.
- `capture.sh <udid>` → fifo → ffmpeg (downscale + HLS) → `serve/stream.m3u8`.
- `index.html` → hls.js player.
- `python3 -m http.server 8099` serves `serve/`; `cloudflared` tunnels a hostname → :8099.

## Notes
- Video frames only flow when the phone screen is AWAKE (encoder needs content).
- Target ONE phone by `--udid`; never flip global state on phones serving traffic.
- Recovery from a stuck device on a shared pod: `systemctl restart usbmuxd` (re-scans all).

## Multi-phone server (deploy/server.js + dashboard.html)
Node (built-ins only) server on :8099 behind the Cloudflare tunnel:
- serves `.ts` as **video/mp2t** + `.m3u8` correctly, all **no-cache** (fixes the
  python http.server bug that served segments as `text/vnd.trolltech.linguist` and
  broke playback).
- `GET /api/phones` (from the pod ledger :7799/state), `POST /api/start|stop|wake?udid=`.
- per-phone on-demand capture → `serve/<udid>/stream.m3u8`; gentle status-bar tap
  keep-awake; Wake button does a WDA home-press (wakes a slept display).
`dashboard.html` = a grid of all phones, click to stream, robust hls.js live player.

## Reliability analysis (the hard, still-open part)
Per-phone AV arming in mode-2 is finicky:
- A **fresh** device (right after a usbmuxd restart / first connect) arms cleanly:
  `PING → CWPA → HPD1 → CVRP → video` in ~5s. Proven repeatedly, end-to-end via Cloudflare.
- After a QVH session the device gets **stuck audio-only**: it answers `PING` and sends
  sparse `SYNC_TIME` but never a fresh `CWPA`, so no `HPD1`/video. The soft re-cycle
  (disable+enable 0x52) **cannot** clear this; only a real reset (usbmuxd restart /
  re-enumeration) does.
- Root cause: QVH's disconnect cleanup does `DisableQTConfig` + `SetConfig(usbmux/4)`,
  which **fails `busy [-6]`** in mode-2 (usbmuxd holds the iface). The device is left
  with a half-open AV session (and often drifts to config 4), so the next connect can't
  re-arm. Repeated re-cycling + direct libusb access on a busy pod also triggers
  `libusb timeout [-7]` (usbmuxd contention).
- **Proper fix (next):** implement a clean AV teardown in QVH (send the host stop/HPD0
  sequence so the device ends the session cleanly and re-arms next time) and/or a
  per-stream fresh-enumeration. Do this on a **dedicated 2–3 phone Linux rig** where
  usbmuxd can be stopped — NOT a production pod (contention defeats clean claims).
