---
quick_id: 260906-uwx
description: Tune tracelet position filters to stop stationary GPS jitter
date: 2026-09-06
status: planned
---

# Quick Task 260906-uwx — Tune tracelet position filters

Implements the fix direction from the `recording-gps-jitter-still` debug session.
A tester's recording in a Leipzig street canyon produced a starburst of
criss-crossing segments while standing still; komoot and a third app recorded
the same walk cleanly. Root cause: the app disables and under-tunes the position
filtering tracelet already provides.

All changes are confined to `app/lib/services/tracelet_position_source.dart`.

## Decisions (user, 2026-09-06)

| Knob | Before | After | Who decided |
|---|---|---|---|
| `filter.maxImpliedSpeed` | 80 m/s (288 km/h) | **15 m/s** (54 km/h), single value for all modes | user |
| `geo.distanceFilter` (foreground) | 0.0 | **3.0** | user |
| `geo.enableDeadReckoning` (foreground) | true (inherited) | **false** | user |
| `filter.trackingAccuracyThreshold` | 100 m | **30 m** | delegated to Claude |
| `filter.policy` | `adjust` | **`ignore`** | delegated to Claude |
| `filter.useKalmanFilter` (background) | false | **true** | from diagnosis |
| `filter.rejectMockLocations` (background) | false | **true** | consistency with foreground |

Mode-aware speed gating (pedestrian vs bicycle via `_recordingCosting`) was
offered and declined in favour of one constant — no plumbing changes.

`trackingAccuracyThreshold: 30` (not 25) because the user expects most activity
outdoors in non-urban terrain: under conifer canopy or in a gorge Android
routinely reports 15–25 m accuracy, and a 25 m ceiling would start dropping
legitimate fixes there. 30 m keeps those while still rejecting gross spikes.

## Tasks

### Task 1 — Add a shared pedestrian-tuned LocationFilter and apply it to both profiles

- files: `app/lib/services/tracelet_position_source.dart`
- action:
  - Add a single `static const tl.LocationFilter _locationFilter` with
    `trackingAccuracyThreshold: 30`, `maxImpliedSpeed: 15`,
    `policy: tl.LocationFilterPolicy.ignore`, `rejectMockLocations: true`,
    `useKalmanFilter: true`. One constant so the two profiles cannot drift
    apart again — the background/foreground Kalman asymmetry is exactly the
    bug being fixed.
  - `_foregroundConfig()`: on the chained `geo.copyWith`, set
    `distanceFilter: 3.0`, `enableDeadReckoning: false`, `filter: _locationFilter`.
  - `_backgroundConfig()`: on the chained `geo.copyWith`, add
    `filter: _locationFilter` (keeping its existing `distanceFilter: 5.0` and
    `desiredAccuracy: high`).
  - Update the doc comments so they describe the new behaviour rather than
    "no distance filter while moving".
- verify: `flutter analyze` clean; `flutter test` green
- done: both profiles carry the same `LocationFilter`; foreground no longer
  sets `distanceFilter: 0.0` and no longer inherits `enableDeadReckoning: true`

## must_haves

- truths:
  - Every fix implying >15 m/s from the previous accepted fix is dropped, not adjusted.
  - Fixes reporting worse than 30 m accuracy are dropped, not adjusted.
  - Kalman smoothing is active in BOTH the foreground and background profiles.
  - Dead reckoning is off in the foreground profile.
- artifacts:
  - `app/lib/services/tracelet_position_source.dart`
- key_links:
  - `.planning/debug/recording-gps-jitter-still.md`

## Out of scope

- Any change to `Navigation.onPosition`, `session_gap_backfill.dart`, or other
  app-side gating. The tracelet config is the primary lever; app-side gating
  would be at most a secondary defense and is not needed for this fix.
- The removed CONV-05 distance smoothing and the elevation
  `thresholdXY_m` / `GpxMetricsComputation(5, 5)` noise filter. Both are
  measurement-time gates over an accumulated track; everything here is
  acquisition-time SDK configuration. Elevation output must stay bit-identical.
- On-device verification. Static checks only; the user builds and field-tests.
