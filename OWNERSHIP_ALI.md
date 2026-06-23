# quicktime_video_hack (ideacatlab fork) — the QVH AV-capture core of the ALI iOS toolchain

Fork base: **danielpaulus/quicktime_video_hack**. Canonical branch: **`main`** (all ALI work lands here;
experiment branches are kept but not merged). This is a **core, heavily-modified library** for ALI — we
capture QuickTime-over-USB H.264+PCM ("Valeria") off iPhones at fleet scale, alongside usbmuxd/WDA, and we
expect to keep changing it.

## Why we own it
The pod runs QVH on real fleets (20+ iPhones per pod) **at the same time as WDA control + MJPEG**, on a
usbmuxd that holds every device in config-5 (mode 2). Upstream qvh assumes it is the *only* process touching
the device and that it drives the device mode itself. Both assumptions are wrong for us and cause fleet
outages, so we patch the source.

## ALI modifications on `main` (the load-bearing ones)
1. **`fix(usb): open ONLY the target device, not the whole fleet`** (`screencapture/discovery.go`).
   `OpenDevice` used `ctx.OpenDevices(func{return true})` — a `libusb_open` **on every phone on the host**
   plus an EP0 serial-descriptor read on each, *per arm*. So one qvh-start did N opens + N EP0 reads across
   the whole fleet; with usbmuxd holding those devices, that contends every phone's EP0 at once and yields
   `Error opening usb device: no device [-4]`. A lone arm survives; arming many in succession wedges usbmuxd
   → phones fail attach → fleet destabilizes. **Fix:** open only the target phone by **bus/address** (now
   carried on `IosDevice`), so each arm is exactly one open + one EP0 read. (Falls back to the legacy scan
   only if bus/address are unknown.)
2. **`fix(arming): default to just-listen`** (`screencapture/usbadapter.go`).
   Default to NOT re-cycling the QT config (no `SET_MODE 0x52` from qvh) and not driving the mode back on
   teardown. On the mode-2 fleet usbmuxd OWNS config 5; a competing `SET_MODE`/`SetConfiguration` from qvh
   forces the iPhone to **re-enumerate** (the device-global, fleet-disrupting event) and fights usbmuxd.
   Legacy re-cycle is behind `QVH_FORCE_RECYCLE=1`. `EnableQTConfig` also early-returns when the AV config
   is already live, so qvh never issues the enable on a config-5 device.

## The root-cause research (2026-06-23) — verdict: no libusb fork needed
Three independent source-level investigations (libusb/kernel, gousb/qvh, ecosystem) agreed:
- The two device-global disruptors are **(a)** qvh opening the whole fleet (`OpenDevice`, fixed in #1) and
  **(b)** the `SET_MODE 0x52` control transfer that re-enumerates the device (gated off in #2). `Config(5)`
  is **not** a disruptor — gousb skips `set_configuration` when the config is unchanged, and the kernel's
  `proc_setconfig` `-EBUSY` guard blocks it anyway while usbmuxd holds an interface. `claim_interface` and
  `clear_halt` are per-interface/endpoint and safe.
- `LIBUSB_ERROR_NO_DEVICE (-4)` is `libusb_open` on a device momentarily mid-transition / re-enumerating —
  a *symptom*, not fixable by any host-side library. **A libusb fork cannot help**; the fix is qvh-side
  (above), which it now is.
- Clean ownership model (Apple/macOS + upstream usbmuxd `6d0183dd`): **usbmuxd owns the device mode/config
  switch and claims only the usbmux interface; the AV consumer claims only the Valeria interface, never the
  config.** Our two fixes implement exactly the qvh half of this.
- Full write-up: `ALI-pod/docs/research/qvh-usbmuxd-coexistence-userspace-vs-single-owner-2026-06-23.md`,
  plus the protocol/usbmuxd-port specs captured in cross-session memory.

## Roadmap
- **Now:** validate the `OpenDevice` fix at fleet scale (single arm proven streaming 5.4 MB; the fix removes
  the N×N open contention that wedged the fleet during multi-phone arming).
- **Endgame (b-full):** port the AV reader INTO usbmuxd (`ideacatlab/usbmuxd`) so ONE process owns config 5,
  the usbmux interface, AND the Valeria reader — zero cross-process contention, the CWPA caught at the
  config-5 transition usbmuxd already performs. Step-1 (claim AV interface + log CWPA) is drafted in the
  usbmuxd fork.

## Branches
- `main` — canonical ALI state (the fixes above + the prior mode-2/HLS/multi-phone work, consolidated here).
- Kept experiments (unmerged, reviewable, NOT deleted): `extra-dict-value-types` (AV-protocol docs),
  `explore/valeria-wifi-mirroring`, `replaceMessageProcessor{,-add-resume-feature,-unbreak-gstreamer-somehow}`,
  `windows/support`. Merge selectively if/when a piece is needed; none are required for QVH-over-USB on Linux.

## The ALI iOS toolchain (owned, github.com/ideacatlab)
`usbmuxd` (daemon/Valeria/re-arm) · `libusbmuxd` (client+iproxy) · `libimobiledevice` (pairing/info/diag) ·
**`quicktime_video_hack` (THIS — QVH AV capture)** · `qvh-webrtc-media-server`. Deps kept upstream: libusb,
libplist, gousb (no fork needed — confirmed by the research above).
