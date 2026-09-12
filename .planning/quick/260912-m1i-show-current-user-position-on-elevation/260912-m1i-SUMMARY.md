---
phase: quick-260912-m1i
plan: 01
subsystem: app
tags: [flutter, navigation, elevation-profile, gps-matching]

requires: []
provides:
  - TrackPositionMatcher (app/lib/util/route/track_position_matcher.dart)
  - ElevationProfile.livePositionMeters overlay
affects:
  - app/lib/routes/navigation_screen.dart

tech-stack:
  added: []
  patterns:
    - "Second RouteMapMatcher instance run over a track's own raw polyline/axis, rather than reusing an existing matcher tuned to a different coordinate system"
    - "ValueNotifier<double?> threaded from a hot-path GPS listener into a leaf widget's ValueListenableBuilder, to avoid rebuilding the parent screen per fix"

key-files:
  created:
    - app/lib/util/route/track_position_matcher.dart
    - app/test/util/route/track_position_matcher_test.dart
  modified:
    - app/lib/components/trail/elevation_profile.dart
    - app/test/components/trail/elevation_profile_test.dart
    - app/lib/routes/navigation_screen.dart

decisions:
  - "buildRawTrackPoints lost its @visibleForTesting annotation (Task 3 deviation) — navigation_screen.dart is a real production caller of it now, not just a test"
  - "elevationAtDistance test tolerance for _simplifyTrackPoints's target count loosened from an exact 250 cap to <=260 — the bucketed min/max-extremes algorithm can emit a couple of points over its nominal target (first + up to 2 per bucket + last), which is pre-existing, unmodified behavior"
  - "unitProvider override in the new widget test uses the codebase's established overrideWithValue('metric') convention rather than the plan's literal overrideWith((ref) => 'metric') wording — functionally identical, matches sibling tests"

metrics:
  duration: ~25min
  completed: 2026-09-12
---

# Phase quick-260912-m1i Plan 01: Show current user position on elevation Summary

Live "you are here" marker (vertical guide + dot) on the elevation profile chart while navigating a trail in the Flutter app, driven by the same GPS fixes that move the map's location marker, hidden whenever the user is more than 50 m off the track or no fix has landed yet.

## What was built

1. **`TrackPositionMatcher`** (`app/lib/util/route/track_position_matcher.dart`) — wraps the existing `RouteMapMatcher` (HMM map-matcher) over a GPX track's own raw polyline and cumulative-metre axis (not the Valhalla-snapped shape `Navigation` already matches against). `update()` returns the along-track distance in chart x-axis units, or `null` when the fix is more than `onTrackThresholdMeters` (default 50 m) from the matcher's committed point — which also covers the "matcher is still confirming a leg jump" uncertain case. `pointAtAlongTrack()` interpolates a position along the polyline, guarding zero-length segments.

2. **`ElevationProfile.livePositionMeters`** (`app/lib/components/trail/elevation_profile.dart`) — new opt-in nullable `ValueListenable<double?>` parameter. When non-null, a `ValueListenableBuilder` scoped to just the `LineChart` widget adds a solid `VerticalLine` plus a single-spot, radius-5 dot `LineChartBarData` at the live x-coordinate (matching the scrub indicator's look). `buildRawTrackPoints` was extracted from `buildElevationTrackPoints` (unsimplified walk, same raw `distanceM` axis) so a live matcher can be built over exactly the points the chart itself derives from. `elevationAtDistance` does a binary-search linear interpolation for the marker's y. Scrub touch/indicator callbacks were updated to discriminate the marker bar (`barIndex`/`dotData.show`) from the profile line so scrubbing is unaffected.

3. **Wiring** (`app/lib/routes/navigation_screen.dart`) — `_trackMatcher` is built once, post-frame, when the trail model resolves (alongside the existing notification-text trail-name lookup), from `buildRawTrackPoints(trail.expand?.gpx)`. The `_sub` GPS listener (documented zero-provider-lookup hot path) now also calls `_trackMatcher?.update(fix, ...)` per fix and publishes the result to `_liveTrackMeters` (a `ValueNotifier<double?>`), which is passed to `ElevationProfile(livePositionMeters: _liveTrackMeters)` in the non-recording branch of `_buildElevationPage`. Recording mode is untouched — no matcher is built there.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `buildRawTrackPoints` incorrectly marked `@visibleForTesting`**
- **Found during:** Task 3
- **Issue:** Task 2 marked the extracted `buildRawTrackPoints` as `@visibleForTesting` per the plan's Task 2 action text. Task 3's wiring in `navigation_screen.dart` calls it directly as production code, which `flutter analyze` correctly flagged as a visibility violation.
- **Fix:** Removed the `@visibleForTesting` annotation and updated the doc comment to note the real caller.
- **Files modified:** `app/lib/components/trail/elevation_profile.dart`
- **Commit:** `60245b83`

**2. [Rule 1 - Bug] Test assertion for `_simplifyTrackPoints`'s output count was too strict**
- **Found during:** Task 2
- **Issue:** The plan's behavior bullet said `buildElevationTrackPoints(sameGpx, 1)` "returns at most 250" points. The actual (unmodified, pre-existing) `_simplifyTrackPoints` bucketing algorithm can emit up to `1 (first) + 2*numBuckets + 1 (last)` = 252 points for a target of 250.
- **Fix:** Loosened the test's upper bound to 260 and added `lessThan(raw.length)` as the meaningful assertion (simplification did occur), with a comment explaining the ceiling isn't exact.
- **Files modified:** `app/test/components/trail/elevation_profile_test.dart`
- **Commit:** `957abe8f`

**3. [Rule 3 - Blocking] `flutter/foundation.dart` import required for `ValueListenable`/`ValueNotifier`**
- **Found during:** Task 2
- **Issue:** The plan noted `ValueListenable` needs `package:flutter/foundation.dart`, but the analyzer initially reported it as an unnecessary import before the type was actually referenced in code (a red herring from edit ordering) — then reported `ValueListenable` as undefined once the import was removed.
- **Fix:** Kept `flutter/foundation.dart` imported in both `elevation_profile.dart` and `track_position_matcher_test.dart` was NOT needed (material.dart already re-exports `ValueNotifier` for the test file), so it was removed there; kept in `elevation_profile.dart` where it's genuinely needed.
- **Files modified:** `app/lib/components/trail/elevation_profile.dart`, `app/test/components/trail/elevation_profile_test.dart`
- **Commit:** `957abe8f`

## Known Stubs

None — all three artifacts (`TrackPositionMatcher`, `ElevationProfile.livePositionMeters`, the `navigation_screen.dart` wiring) are fully wired and tested; no placeholder data paths.

## Threat Flags

None — no new network endpoints, auth paths, or schema changes. The GPS-fix trust boundary (T-m1i-01) and per-fix cost boundary (T-m1i-02) from the plan's threat model are both implemented as specified (on-track gate hides a misplaced marker; matcher construction happens once outside the hot path; `ValueNotifier` suppresses no-change notifications).

## Verification

- `dart format --output=none --set-exit-if-changed` on all 5 touched/created files: clean.
- `flutter analyze` (whole project): 13 pre-existing issues, none in files touched by this plan.
- `flutter analyze lib/util/route/track_position_matcher.dart test/util/route/track_position_matcher_test.dart`: no issues.
- `flutter analyze lib/components/trail test/components/trail`: no issues.
- `flutter analyze lib/routes/navigation_screen.dart`: no issues.
- `flutter test test/util/route/track_position_matcher_test.dart test/components/trail/elevation_profile_test.dart test/util/route/map_matcher_test.dart`: 30/30 passed.
- `git diff --stat -- lib/util/route/map_matcher.dart`: no output (file untouched, per the plan's explicit constraint).
- Not run: `flutter build`/`adb install`/on-device walk test — per project convention, the user builds, installs, and verifies on-device.

## Self-Check: PASSED

- `app/lib/util/route/track_position_matcher.dart` — FOUND
- `app/test/util/route/track_position_matcher_test.dart` — FOUND
- `app/lib/components/trail/elevation_profile.dart` — FOUND (modified)
- `app/test/components/trail/elevation_profile_test.dart` — FOUND (modified)
- `app/lib/routes/navigation_screen.dart` — FOUND (modified)
- Commit `aafe3544` — FOUND in `git log`
- Commit `957abe8f` — FOUND in `git log`
- Commit `60245b83` — FOUND in `git log`
