---
quick_id: 260906-uwx
description: Tune tracelet position filters to stop stationary GPS jitter
date: 2026-09-06
status: complete
---

# Quick Task 260906-uwx — Summary

Applied the fix direction from `.planning/debug/recording-gps-jitter-still.md`.
One file changed: `app/lib/services/tracelet_position_source.dart`.

## What changed

Added `TraceletPositionSource._locationFilter`, a single shared
`tl.LocationFilter`, and applied it to **both** the foreground and background
profiles:

| Setting | Before | After |
|---|---|---|
| `filter.trackingAccuracyThreshold` | 100 m (tracelet default) | 30 m |
| `filter.maxImpliedSpeed` | 80 m/s / 288 km/h (default) | 15 m/s / 54 km/h |
| `filter.policy` | `adjust` (correct + record anyway) | `ignore` (drop) |
| `filter.rejectMockLocations` | true fg / false bg | true, both |
| `filter.useKalmanFilter` | true fg / **false bg** | true, both |
| `geo.distanceFilter` (foreground) | 0.0 | 3.0 |
| `geo.enableDeadReckoning` (foreground) | true (preset-inherited) | false |

Background `distanceFilter` stays 5.0 and `desiredAccuracy` stays high —
untouched, they were already battery-appropriate.

The filter is one shared constant rather than two inline literals specifically
because the presets disagree: `Config.highAccuracy()` enables the Kalman filter
and `Config.balanced()` disables it, which is how a screen-off recording came to
lose GPS smoothing entirely. Sharing the object prevents that divergence
recurring.

## Verification

- `flutter analyze lib/services/tracelet_position_source.dart` — No issues found.
- `flutter analyze` (whole app) — 13 issues, all pre-existing `info` lints in
  `vendor/tiptap_flutter/`; none in changed code.
- `flutter test` — 1124 passed, 1 skipped, 0 failed.
- **No on-device verification.** These are acquisition-time settings inside the
  native tracking SDK; nothing in the Dart test suite exercises them. Field
  testing is required.

## Constraints honored

Nothing in the CONV-05 / elevation constraint set was touched. `distanceFilter`
is an acquisition-time gate inside tracelet (last recorded fix vs. incoming
fix); the removed CONV-05 gate and the load-bearing elevation `thresholdXY_m` /
`GpxMetricsComputation(5, 5)` filter are measurement-time gates over an already
accumulated track. `elevation_profile.dart` and the vendored
`maplibre-elevation-profile` are untouched; elevation output is unchanged.

## Open risks for field test

1. **The accuracy ceiling is the weaker gate.** Multipath spikes often report
   optimistic accuracy, so 30 m will not catch all of them. `maxImpliedSpeed:
   15` is what should actually kill the tester's starburst. If jitter persists,
   that is the number to tighten, not the accuracy ceiling.
2. **`distanceFilter: 3.0` is an unvalidated guess.** Chosen below the preset's
   5.0 to reduce switchback chord-shortcutting risk, but with no raw GPX to
   measure against. Watch for thinned tracks on tight switchbacks or slow
   uphill walking. The regression concern is the same one behind
   `fixtures/gpx-corpus/12-dense-switchback`.
3. **Dead reckoning is now off in the foreground.** Genuine GPS dropouts
   (tunnels, deep gorges) will leave gaps in the track rather than
   IMU-estimated positions. If gap-free tracks matter more than jitter, this is
   the knob to revisit — possibly re-enabled with a non-zero
   `deadReckoningActivationDelay` instead of outright.
4. **Empty-track failure mode.** If conditions are bad enough that every fix
   breaches 30 m, `policy: ignore` records nothing rather than something noisy.
   Judged unlikely for non-urban outdoor use, but it is the trade being made.

## Follow-ups not done

- App-side gating in `Navigation.onPosition` / `session_gap_backfill.dart`.
  Deliberately skipped — the tracelet config is the primary lever and this
  would be at most a secondary defense.
- Mode-aware speed gating (pedestrian vs bicycle via `_recordingCosting`).
  Offered and declined in favour of one constant.
- Raw GPX from the tester's original session would let every threshold above be
  picked from measured values rather than estimated. Still worth requesting.
