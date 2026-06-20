# ALI QuickTime-over-USB live streaming — master POC report

**Goal:** stream a real iPhone's screen as **hardware H.264 over USB on a Linux pod**
(no Mac, no MJPEG), **coexisting with WDA/Appium control on the same phone**, delivered to
a browser through a **Cloudflare tunnel**. Replace the WDA-MJPEG screen feed (the chronic
"loading"/lag source) with the phone's own hardware encoder.

Date: 2026-06-20. Test bed: ALI `ascensionpod1` (dev), 29× jailbroken iPhone @ iOS 16.7.11,
`usbmuxd-master` (= libimobiledevice/usbmuxd `1.1.1-72-g3ded00c`).

## Verdict
- **PROVEN possible and demonstrated end-to-end multiple times** — iPhone hardware H.264
  over USB on Linux, with WDA control alive on the same phone, played live in a browser via
  `qvh.turbocat.dk`. No public project does this on Linux; this is the breakthrough.
- **NOT yet reliable per-phone.** Arming the AV session is fragile on a contended prod pod
  (details below). The fix is well-scoped (a clean SET_MODE reset, ideally in usbmuxd) and
  belongs on a **dedicated rig**, not a control-serving pod.

## Repos (all on github.com/ideacatlab)
| repo | what | branch |
|---|---|---|
| `quicktime_video_hack` | fork of danielpaulus/QVH + our patches, `deploy/`, `tools/`, this report | `mode2-wda-coexistence` |
| `usbmuxd` | fork of libimobiledevice/usbmuxd @ `3ded00c` + `FIX_PLAN_ALI.md` | `valeria-ali` |

## Architecture (the pipeline)
```
iPhone (iOS16.7, config 5 "Valeria")
  ├─ iface 5.1 usbmux/control (0xFE) ── usbmuxd ── iproxy ── WDA/Appium ── dashboard control
  └─ iface 5.2 AV/Valeria   (0x2A) ── qvh record ── H.264 fifo ── ffmpeg (HLS, scale 1024, no-cache)
                                                                      └─ Node server :8099 (deploy/server.js)
                                                                           └─ cloudflared ── qvh.turbocat.dk ── browser (hls.js, deploy/dashboard.html)
```
- **usbmuxd device-mode 2** (`USBMUXD_DEFAULT_DEVICE_MODE=2`) puts the phone on **config 5**,
  which carries BOTH the control iface and the AV iface → coexistence.
- **qvh** (this fork) does the QuickTime AV USB protocol, writes Annex-B H.264 + PCM audio.
- **ffmpeg** repackages to HLS (`video/mp2t`, no-cache — critical, see bugs).
- **Node server** lists phones (pod ledger), starts/stops/wakes per-phone captures, serves HLS.
- **cloudflared** tunnels it to a public HTTPS host.

## How the device side actually works (the RE)
- Apple modes via vendor control transfers (`usbmuxd src/usb.c`): `GET_MODE=0x45`,
  **`SET_MODE=0x52`**; mode **1 = initial (config 4)**, **2 = Valeria (config 5 / H.26x AV)**.
- **QVH's `EnableQTConfig` is literally `SET_MODE 0x52`** — same command usbmuxd uses, hence
  the conflict.
- QuickTime AV handshake (`screencapture/messageprocessor.go`): device `PING` → host `PING`;
  device `SYNC_CWPA` (audio clock) → host replies **+ sends `HPD1`** (video-enable) → device
  `SYNC_CVRP` (video format, e.g. 1126×2436 AVC-1) → host `NEED` → device `FEED` (H.264).
  **`CWPA` is the trigger for the whole video chain.**

## What WORKED (verified)
1. QVH builds + runs on Linux (Go 1.22, libusb 1.0.27, glib/gstreamer dev headers).
2. Mode-2 → all phones on config 5; pod kept 28/29 attached (control unaffected).
3. **Coexistence**: control + QuickTime AV session live on the same phone simultaneously.
4. **CWPA-gated re-arm** (`usbadapter.go`): re-cycle the QT config *while reading*, counting
   the session armed only on a real `CWPA` (not leftover audio `asyn`/`SYNC_TIME`) → made the
   fresh-device handshake reliable.
5. **End-to-end video** via Cloudflare, browser-played, multiple times.
6. **Cloudflare tunnel** as stable systemd services (survives SSH drops).

## What did NOT work / the bugs (so we don't repeat them)
1. **Playback was black despite live segments** — the python `http.server` served `.ts` as
   `text/vnd.trolltech.linguist`. Fix = serve `video/mp2t` + no-cache (`deploy/server.js`).
2. **Soft re-cycle `disable(0)+enable(2)` cannot reset a stuck session** — `wIndex 0` is not a
   valid Apple mode, so it is not a real transition. A real **`SET_MODE 1 → 2`** transition is
   needed (now in QVH behind `QVH_RESET_MODE_TRANSITION=1`; long-term belongs in usbmuxd).
3. **Stuck "audio-only"**: after a session the device answers PING + `SYNC_TIME` but never a
   fresh `CWPA` → no `HPD1`/video. Root cause: QVH's teardown (`SetConfig(usbmux)`) fails
   `busy[-6]` in mode-2 (usbmuxd holds the iface) → half-open AV session; device drifts to
   config 4.
4. **Screen must be awake** to encode frames — a *slept* display ignores synthetic taps; only
   a **WDA home-press** wakes it. Keep-awake = gentle status-bar tap (awake) + Wake button
   (home-press for slept).
5. **Contention degrades the fleet** — repeated probing/usbmuxd-restarts → `libusb timeout[-7]`,
   `guess_mode` fails so usbmuxd stops applying mode-2, and the WDA control path lags. Hard
   data from `tools/fleet-av-scan.sh`: whole fleet → `NO-AV-IFACE` (0% armable) after heavy RE.
   **This is why the reliability work must move to a dedicated rig.**

## The reliability fix (next, on a rig)
Implement a **clean per-device AV reset inside usbmuxd** (`SET_MODE 1 → 2`), exposed on
demand (a usbmux `ResetDeviceMode` request), so the capture manager arms each phone fresh
without a fleet-wide restart and without contending for the device. Full design +
code-locations in `ideacatlab/usbmuxd` → `FIX_PLAN_ALI.md`. QVH already has the transition
behind a flag for A/B testing (`QVH_RESET_MODE_TRANSITION=1`).

## Reproduce / iterate
1. Pod prereqs: usbmuxd mode-2 drop-in, `qvh` built, `deploy/server.js` (qvh-server), cloudflared.
2. `tools/fleet-av-scan.sh` → baseline. `tools/qvh-reliability.sh <udid> 10 <port>` on a phone.
3. Toggle `QVH_RESET_MODE_TRANSITION=1`, rebuild, re-measure. Confirm with `tools/playwright-verify.js`.
4. Log deltas in `FINDINGS_IDEACATLAB.md`. See `tools/README.md` + `deploy/README.md`.

## Business framing
Hits the on-USP goal (kill MJPEG "loading"/lag with the phone's own H.264 encoder) for the
iOS≤16 jailbroken segment. The iOS-17+ RSD-tunnel wall does NOT apply to these devices.
Strategic side-benefit: owning a usbmuxd fork also lets us fix the fleet's chronic usbmuxd
pain (blind state, contention, watchdog hacks).
