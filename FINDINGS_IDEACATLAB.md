# QVH on Linux + WDA coexistence — ideacatlab RE findings (2026-06-20)

Live reverse-engineering of `quicktime_video_hack` (QVH) on a **Linux** iPhone farm
(ALI `ascensionpod1`, 29× jailbroken **iPhone, iOS 16.7.11**, `usbmuxd-master`), to
stream the iPhone screen as **hardware H.264 over USB** while keeping **WDA/Appium
control** running on the same phones. Goal: video → Cloudflare → operator's screen.

## TL;DR

- **QVH activates QuickTime AV mode on iOS 16.7 on Linux** with `usbmuxd-master`
  running (config 4 → 5), targeting a single device by `--udid` (no fleet impact).
- **WDA-control + QuickTime-video CAN coexist on the same phone.** The key is usbmuxd
  **device mode 2** (`USBMUXD_DEFAULT_DEVICE_MODE=2`): it puts every device on
  **config 5 ("Valeria"), which carries BOTH the usbmux/control interface (subclass
  `0xFE`) AND the AV interface (subclass `0x2A`)** — so control runs on `5.1` and
  video on `5.2` simultaneously. Verified: all 29 phones moved to config 5 and the pod
  kept **28 attached** (control unaffected).
- **The mode-2 fix in this fork** (`screencapture/activator.go`): when the device is
  already on the QT config, QVH used to *skip* activation, so the device never (re)armed
  the AV session and never sent `PING`. We now **re-cycle the AV session via
  disable+enable control transfers** — this produces `PING received` + the full audio
  handshake (`SYNC_CWPA`, `SYNC_AFMT`, ~3 MB audio) **while the phone stays `attached`
  to WDA.** Coexistence at the session level: working.

## Config 5 interface map (iOS 16.7, iPhone `idProduct 0x12a8`)

```
config 5 (Valeria / AV mode):
  iface 5.0  class 06 sub 01   Imaging (PTP)
  iface 5.1  class ff sub FE   Apple USB Multiplexor (usbmux) <- WDA/iproxy CONTROL
  iface 5.2  class ff sub 2A   Valeria / AV                   <- QVH H.264 VIDEO
  iface 5.3  class 02 sub 0D   CDC Communications
  iface 5.4  class 0A sub 00   CDC Data
```
`usbmuxd-master` log confirms it: `Found Valeria and Apple USB Multiplexor in device %i-%i configuration 5`.

## What works (verified live)
1. Build + run on Ubuntu 24.04 / Go 1.22 / libusb 1.0.27 (needs glib + gstreamer dev headers).
2. `qvh activate --udid=X` → config 5, `screen_mirroring_enabled:true`.
3. `usbmuxd` device-mode 2 → all devices on config 5, control intact (28/29 attached).
4. This fork's recycle patch → `PING received` → audio handshake, **phone stays attached.**

## Open issues (remaining work)
1. **Video sub-negotiation (`CVRP`/`FEED`) never completes** — audio streams, but the
   device never returns the H.264 format/frames after the host sends `HPD1`. Strong
   hypothesis: **the farm phones' screens are asleep** (no frames to encode). Needs a
   reliable software wake (these are jailbroken — SSH-to-device wake, or a WDA/Appium
   tap that lands) then re-test. Alternatively a video-handshake timing fix.
2. **PING reliability in mode 2 is racy** — the disable+enable re-cycle sometimes lands
   after QVH starts reading (no PING). Needs **retry/poll**: if no PING within N ms,
   re-cycle and re-read. (Add to `StartReading` / `EnableQTConfig`.)
3. **Deactivate on a mode-2 device fails** `set config 4: busy [-6]` (usbmuxd holds it).
   Recovery from a manual USB reset requires `systemctl restart usbmuxd` (re-scans all)
   — phones default to **config 4 = usbmux**; mode-2 default is config 5.

## Operational notes / gotchas
- Target ONE phone with `--udid`; never flip global state on phones serving traffic.
- A manually USB-reset phone drops off `usbmuxd` (`--no-preflight` won't re-register it);
  `systemctl restart usbmuxd` cleanly re-scans and recovers the whole fleet.
- The broken test phone had a **dead display** (audio streamed, no video) — not a QVH bug.
- iOS ≤ 16 (jailbroken) is QVH's working range; the iOS-17+ RSD-tunnel coexistence wall
  does **not** apply here (these devices use plain usbmux).

## Next steps
- [ ] Reliable screen-wake (jailbroken: SSH `uiopen`/SpringBoard, or land a tap) → confirm `CVRP`/`FEED` → real H.264.
- [ ] PING retry loop in mode 2 for reliability.
- [ ] `qvh gstreamer` → low-latency H.264 web stream (WebCodecs/MSE) → `cloudflared` tunnel → operator screen.
- [ ] Integrate as the pod video source replacing WDA-MJPEG (flag-gated, MJPEG fallback).
