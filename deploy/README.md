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
