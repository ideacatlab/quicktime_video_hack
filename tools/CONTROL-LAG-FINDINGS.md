# Control-lag investigation — findings (2026-06-20, ascensionpod1, config-5 fleet)

Measured with `tools/control-latency.py` + `tools/fleet-control-latency.sh` on a live pod.

## Diagnosis: it's USB **concurrency contention**, not a per-phone or transport defect
- WDA `/status` = **0.35s**, raw iproxy connect = **instant** → usbmux transport is fine.
- A tap on a **single** phone = **~0.6s** → control itself is fine.
- BUT a fleet scan that taps 28 phones in rapid sequence shows **5-25% "slow" (8-14s),
  varying every run** — the *same* phone is fast, then slow, then fast. The scan's own
  concurrent tapping causes it. Adding a screenshot "warmer" on all phones made it WORSE
  (1 slow → 7 slow). **Concurrent USB load on config-5's shared usbmux bandwidth stalls
  commands.**
- So the lag the operator feels appears under **load** — many phones being controlled +
  streamed at once — not on a quiet pod.

## Why config-5 is implicated
config-5 ("Valeria") adds the AV interface alongside usbmux on one config. Even idle it can
consume usbmux attention; active QVH streaming definitely competes for the same bandwidth.
A/B config-4-vs-config-5 under identical concurrent load (maintenance window / rig) is the
clean way to quantify the config-5 tax.

## What did NOT help / made it worse
- **Keep-warm via periodic screenshots** (`control-health.sh` warmer): counterproductive —
  it adds exactly the concurrent USB load that causes the stalls. Left disabled.
- Per-phone "recovery" (detach/re-attach): phones self-recover; not the lever.

## Fix directions (ranked)
1. **Reduce concurrent USB load.** Don't screenshot/stream every phone continuously; stream
   only viewed phones (the pod already gates MJPEG on consumers — verify it's effective on
   config-5), and keep QVH captures to the few phones actually being watched.
2. **usbmuxd/usbmux fork optimization** — fairer scheduling / lower per-transfer overhead on
   config-5 devices so control isn't starved by AV/screenshot traffic. (ideacatlab/usbmuxd.)
3. **Spread phones across USB controllers/buses** so bandwidth isn't shared 28-wide.
4. **A/B config-4 vs config-5** under load to decide whether to keep config-5 fleet-wide or
   only switch a phone to config-5 on demand when streaming it (then back).

## Tools delivered
- `control-latency.py` — single-phone /status + warm/cold tap percentiles.
- `fleet-control-latency.sh` — fleet-wide tap latency, flags slow phones (also demonstrates
  the concurrency contention).
- `control-health.sh` — warmer + monitor + optional recovery (RECOVER=0 default). Use the
  monitor part for visibility; warming is NOT recommended (adds contention).
