---
slug: recording-gps-jitter-still
status: diagnosed
trigger: "A tester reports a lot of jitter while the app was sitting still or with little motion during a recording session (1st image). He provided the other images as comparison of the same activity recorded by komoot and another app."
goal: find_root_cause_only
created: 2026-09-06
updated: 2026-09-06
---

# Debug Session: recording-gps-jitter-still

## Symptoms

**Expected behavior**
While the device is stationary or moving slowly, the recorded track should stay
compact — a small cluster or a short stub, the way the two reference apps render
the same walk. Slow/stopped motion should not accumulate spurious track geometry.

**Actual behavior**
Wanderer's recorded track for the same walk is a dense starburst of criss-crossing
segments spanning roughly a half-block. Direction arrows point every which way,
indicating many recorded points with large frame-to-frame displacement while the
tester was standing still or barely moving. The same physical activity recorded
by komoot and by a third app on the same outing produces clean, sparse polylines
over the same streets — the reference tracks show the *real* route (a few passes
along Karl-Heine-Straße and into the courtyard) with no starburst.

Location context from the screenshots: Leipzig, Karl-Heine-Straße / Hähnelstraße
(Schaubühne Lindenfels area) — a dense urban street canyon with multi-storey
buildings and a courtyard, i.e. exactly the environment that produces multipath
GPS error. All three apps recorded the same outing, so the raw satellite
conditions were comparable; the difference is in what each app *kept*.

**Error messages**
None reported. No crash, no user-visible error — the track is simply wrong.

**Timeline**
Not established. Unknown whether this is a regression or has always been present.
Note two adjacent recent changes to the recording pipeline that are worth checking
for a timeline link:
- 2026-08-01 quick task `260801-opr` superseded CONV-05's smoothed distance; the
  reported distance is now the raw haversine accumulator. Distance smoothing was
  deliberately removed — see the constraint below.
- 2026-08-09 two recording-elevation fixes landed unverified in the field
  (`.planning/debug/recording-elevation-gain-jump.md`, left unarchived on purpose),
  including a change to how synthetic seed fixes from `TraceletPositionSource`
  are treated.

**Reproduction**
Start a recording in the Wanderer app in a dense urban environment and stand still
(or move slowly / stop-and-go) for several minutes. Not yet reproduced in-house.

## Available Evidence

- Three map screenshots supplied by the tester, all covering the same block:
  1. Wanderer's recording — heavy starburst jitter over the courtyard/street.
  2. A second app's recording — clean, sparse polyline, ~10 straight segments.
  3. komoot's recording — clean polyline with sparse direction markers.
- **No raw GPX / track record and no logs are available yet.** The tester can be
  asked for an export if the investigation needs point-level data (timestamps,
  reported accuracy, speed, provider). Flag this at a checkpoint rather than
  assuming point data.
- Device and OS are unknown — the tester was not asked. Treat a typical Android
  phone as the working assumption but do not build the diagnosis on it; check the
  iOS path too where the pipeline differs.

## Investigation Pointers

Recording pipeline files (Flutter app, `app/lib/`):
- `services/tracelet_position_source.dart` — position source, seeding
- `provider/foreground_position_stream_provider.dart` — stream setup, `LocationSettings`
- `provider/navigation_provider.dart` — breadcrumb accumulation
- `provider/navigation_stats_provider.dart` — live metrics
- `services/session_gap_backfill.dart` — gap backfill on resume
- `entities/active_navigation_entity.dart` — persisted recording state
- `routes/navigation_screen.dart` — recording UI
- `util/trail/gpx_conversion_util.dart` — GPX metrics computation

Questions worth answering from source:
- What `LocationSettings` (accuracy, `distanceFilter`, interval) does the app
  request, foreground and background, per platform? A `distanceFilter` of 0 with
  high-accuracy polling accepts every noisy fix.
- Is any fix ever *rejected* before being appended to the breadcrumb — on
  `position.accuracy`, on implied speed between consecutive fixes, on age/staleness,
  or on `isMocked`? If the answer is "no filter at all", that is the likely root
  cause and it explains the contrast with komoot directly.
- Does `session_gap_backfill.dart` inject points that could compound the effect?
- Do synthetic seed positions (`seedPositionFrom`) enter the breadcrumb?

## Hard Constraint — do not "fix" this by reintroducing distance smoothing

The 5 m XY gate in the *distance* path was removed deliberately on 2026-08-01 and
validated against FIT ground truth (raw +0.54% vs. gated −3.29% over a 10.9 km
track); it chord-shortcuts switchbacks. Do not reintroduce it.

`thresholdXY_m` and both `GpxMetricsComputation(5, 5)` call sites are load-bearing
for the *elevation* noise filter and must not be touched — elevation output must
stay bit-identical. Do not modify `elevation_profile.dart` or the vendored
`maplibre-elevation-profile`.

The correct place for a fix (if one is proposed later) is **rejecting bad fixes at
ingest**, not smoothing an already-accumulated track. This session is
diagnose-only regardless — produce the Root Cause Report, do not apply a fix.

## Current Focus

reasoning_checkpoint:
  hypothesis: "The recording pipeline appends every GPS fix delivered while
    tracelet's coarse isMoving/isStationary state is 'moving' or 'slowing',
    with zero per-fix quality validation anywhere in the ingest path (no
    accuracy threshold, no implied-speed/displacement outlier gate, no
    isMocked/staleness check). In the tester's dense street-canyon location,
    individual fixes carry gross multipath positional error while the person
    is stationary or moving slowly; since 'moving' per tracelet's classifier
    is the ONLY admission criterion (not per-point quality), every noisy fix
    is drawn into the breadcrumb verbatim, producing the crisscrossing
    starburst."
  confirming_evidence:
    - "tracelet_position_source.dart `_onLocation` forwards every native fix
      into the shared stream unfiltered (no accuracy/speed check)."
    - "navigation_provider.dart `Navigation.onPosition()` appends
      unconditionally to `_breadcrumb` whenever `recordBreadcrumb` is true —
      the only gate is the caller-supplied boolean, never fix content."
    - "navigation_screen.dart's per-fix listener sets
      `recordBreadcrumb: !_frozen`, and `_frozen` mirrors
      `stats.isPaused || stats.isStationary`; `isStationary` is driven solely
      by `TraceletPositionSource.isMovingStream`, itself sourced from
      tracelet's native `MotionDetectionMode.speed` state machine — a
      binary/aggregate signal, not a per-fix check."
    - "session_gap_backfill.dart (the dead-app gap-splice path) mirrors the
      exact same single gate — `if (!location.isMoving) continue;` — with no
      independent accuracy/outlier check either, confirming this is a
      systemic design choice across the whole pipeline, not an isolated
      oversight in one code path."
    - "`_foregroundConfig()` sets `distanceFilter: 0.0` while moving, so
      nothing throttles fix cadence/displacement either — every fix tracelet
      hands over while 'moving' is accepted."
    - "Tester's own wording — 'sitting still or with little motion' — and the
      screenshot location (Karl-Heine-Straße/Hähnelstraße street canyon +
      courtyard, Leipzig) match the exact condition needed: real (if slow)
      motion keeps tracelet's classifier legitimately in 'moving', so the
      one existing gate correctly stays open, while dense multi-storey
      buildings are a textbook multipath/NLOS environment that degrades
      individual fix accuracy independent of any classifier behavior."
  falsification_test: "Point-level data (accuracy, speed, timestamp per fix)
    from the tester's recording would falsify this if the starburst points
    turn out to carry tight accuracy figures (e.g. <10 m) despite large
    frame-to-frame displacement — that would instead implicate a downstream
    coordinate bug (e.g. stale/duplicate timestamp handling, unit mixup)
    rather than raw multipath noise passing an absent filter. No such data is
    available in this session (screenshots only); flagged as the one gap a
    checkpoint could close if raw GPX/logs become available."
  fix_rationale: "The fix belongs at ingest — reject a fix before it is
    appended to the breadcrumb (e.g. an `accuracy` ceiling and/or an
    implied-speed-between-consecutive-ACCEPTED-fixes outlier gate in
    `Navigation.onPosition` / the per-fix listener in navigation_screen.dart)
    — not by smoothing or post-processing the already-accumulated track. This
    matches the hard constraint: the removed 5 m XY distance-smoothing gate
    and the load-bearing elevation `thresholdXY_m` gate are both
    post-hoc/accumulated-track mechanisms and must stay untouched; the gap
    being diagnosed here is a missing *admission* check, a different
    mechanism entirely."
  blind_spots: "tracelet's own speed-motion classifier
    (`com.ikolvi:tracelet-sdk`, a closed-source native Rust core fetched from
    Maven Central, not vendored in this repo) is opaque from source — cannot
    confirm whether it derives 'speed' from a hardware Doppler reading
    (more multipath-robust) or from raw position deltas (equally vulnerable
    to the same jitter, which would additionally explain why the starburst
    persisted for 'several minutes' rather than a brief ~10 s grace window
    around a stop). This is a compounding/severity question, not a
    prerequisite for the root cause: the confirmed absence of any per-fix
    quality gate is independently sufficient to explain the reported defect,
    since the tester describes real (if slow) motion, not a pure stop, for at
    least part of the session. Device/OS was never confirmed by the tester
    either (assumed Android based on typical usage, not verified) — the iOS
    path shares the same unfiltered `Navigation.onPosition`/
    `TraceletPositionSource._onLocation` code, so the diagnosis holds on both
    platforms regardless."
- next_action: Report revised root cause (diagnose-only). Tracelet config is the primary lever — see Revision section.

## Evidence

- timestamp: 2026-09-06
  checked: app/lib/services/tracelet_position_source.dart (`_onLocation`,
    `_foregroundConfig`, `_onSpeedMotionChange`)
  found: `_onLocation` converts every `tl.Location` the native engine emits
    directly into a `geo.Position` and pushes it to the shared broadcast
    stream — no accuracy, speed, or plausibility check of any kind.
    `_foregroundConfig()` sets `distanceFilter: 0.0` for the "moving"
    profile, so cadence/displacement is not throttled either. The only motion
    signal exposed (`isMovingStream`) is a coarse boolean derived from
    tracelet's native `MotionDetectionMode.speed` state machine
    (`speedMovingThreshold: 0.4 m/s`, `speedStationaryDelay: 10s`), which is
    an aggregate "is the user moving" classification, not a per-fix quality
    gate.
  implication: The bridge from the native location engine into the app has
    zero fix-level filtering; every native fix reaches the rest of the app
    verbatim while considered "moving".

- timestamp: 2026-09-06
  checked: app/lib/provider/navigation_provider.dart (`Navigation.onPosition`)
  found: `onPosition` appends unconditionally to the in-place `_breadcrumb`
    list whenever the caller passes `recordBreadcrumb: true` — no check on
    `accuracy`, implied speed/displacement from the previous breadcrumb
    point, altitude plausibility, or fix age exists in this function.
  implication: The breadcrumb — the data persisted to disk and exported as
    the saved trail's GPX — has no defense against an individual bad fix
    beyond the caller's binary recordBreadcrumb flag.

- timestamp: 2026-09-06
  checked: app/lib/routes/navigation_screen.dart (per-fix stream listener,
    `_movingSub`, `_frozen`)
  found: The hot-path listener (`_sub = _positionStream.listen(...)`) passes
    every fix straight to `_navNotifier.onPosition(...)` with
    `recordBreadcrumb: !_frozen`. `_frozen` mirrors
    `stats.isPaused || stats.isStationary`; `isStationary` is set exclusively
    from `_positionSource.isMovingStream` (`_movingSub`), i.e. from tracelet's
    native motion classifier — the SAME coarse signal noted above. No other
    gate exists between the raw position stream and the breadcrumb append.
  implication: The single existing "is this fix legitimate" gate in the live
    recording path is the binary moving/stationary classification — there is
    no independent per-fix quality check anywhere downstream of it either.

- timestamp: 2026-09-06
  checked: app/lib/services/session_gap_backfill.dart (`backfillSessionGap`)
  found: The dead-app gap-splice path (recovers points tracelet recorded
    while the process was killed) uses the exact same single gate:
    `if (!location.isMoving) { ...; continue; }` — no accuracy/outlier check
    independent of that flag either.
  implication: This is a systemic, pipeline-wide design choice (motion
    state is the only admission criterion for the breadcrumb), not an
    isolated oversight in one code path — reinforces confidence that no
    fix-quality gate exists anywhere in this app's recording code, on either
    the live or backfill path.

- timestamp: 2026-09-06
  checked: app/pubspec.lock + ~/.pub-cache tracelet-3.5.0 / tracelet_android-3.5.0
    source tree (locked version)
  found: `tracelet_android`'s own Kotlin source ships only plugin/wrapper
    code (pigeon API, plugin registration, headless task service) — the
    actual location engine and speed-motion state machine live in
    `com.ikolvi:tracelet-sdk:3.5.0`, a separate closed-source dependency
    resolved from Maven Central (backed by a Rust core per the Gradle
    `jna.library.path`/`system.loadLibrary("tracelet_core")` wiring), not
    vendored in this repo.
  implication: Whether tracelet's own "moving" classification is itself
    robust to multipath-driven position jitter (e.g. Doppler-based speed vs.
    position-delta-based speed) cannot be verified from source in this repo.
    This is a blind spot affecting only the SEVERITY/DURATION explanation,
    not the core finding — the confirmed absence of any app-level per-fix
    quality gate is independently sufficient to explain the reported
    starburst.

## Eliminated

(none — investigation converged on the first well-evidenced hypothesis
without needing to test and discard alternatives; the "no filter at ingest"
question was the single highest-value question flagged in the original
Investigation Pointers, and source inspection confirmed it directly)

## Revision — 2026-09-06 (user steer: "tracelet also has position filters based on distance")

The first pass concluded "no filter exists anywhere" and recommended building an
ingest gate by hand. That was incomplete. **tracelet ships exactly those gates
already**, in `GeoConfig` and `GeoConfig.filter` (`LocationFilter`) — the app is
either explicitly turning them off or leaving them at vehicle-scale defaults.
The earlier claim that the tracking SDK is opaque was also wrong in the part that
mattered: the config surface is fully readable in the Dart package.

Verified against the **locked** version, `tracelet 3.5.0` (`app/pubspec.lock`),
not just the newest cached 3.6.4 — the profile JSON and every default cited below
are byte-identical in both.

### What tracelet offers (source: `tracelet-3.5.0/lib/src/models/config.dart`)

`GeoConfig`:
- `distanceFilter` — "minimum distance (m) the device must move horizontally
  before a new location update is recorded", default `10.0`
- `stationaryRadius` — radius (m) treated as stationary, default `25.0`
- `disableElasticity` / `elasticityMultiplier` — speed-based distance-filter scaling
- `enableSparseUpdates` / `sparseDistanceThreshold` / `sparseMaxIdleSeconds` —
  persistence-layer dedup, drops fixes within N m of the last recorded position

`GeoConfig.filter` (`LocationFilter`), all left at defaults by this app:
- `trackingAccuracyThreshold` — "reject locations with accuracy worse than this
  value (meters)", default **100**
- `maxImpliedSpeed` — "reject locations that imply a speed greater than this value
  (m/s)", default **80** (= 288 km/h)
- `odometerAccuracyThreshold` — default `50`
- `policy` (`LocationFilterPolicy`) — default **`adjust`** = "smooth/correct rejected
  locations before recording them". The alternatives are `ignore` ("silently
  discard") and `discard` ("discard and emit an error event").
- `rejectMockLocations`, `mockDetectionLevel`
- `useKalmanFilter` — Extended Kalman smoothing, "eliminate jitter, produce cleaner
  tracks" (README), runs natively on both platforms

Also relevant: `enableDeadReckoning` + `deadReckoningActivationDelay` (default `0`)
+ `deadReckoningMaxDuration` (default `0` = unlimited). Per the tracelet CHANGELOG,
dead reckoning is IMU sensor fusion that "activates automatically on GPS loss,
deactivates on GPS recovery", and the README names **urban canyons** as a target
environment.

### What the presets set (`Config._profilesJson`, identical in 3.5.0 and 3.6.4)

`Config.highAccuracy()`:
`distanceFilter: 5.0`, `stationaryRadius: 25.0`, `disableElasticity: true`,
`enableAdaptiveMode: false`, **`enableDeadReckoning: true`**,
`filter: {useKalmanFilter: true, rejectMockLocations: true}`

`Config.balanced()`:
`distanceFilter: 20.0`, `stationaryRadius: 50.0`, `enableAdaptiveMode: true`,
`disableElasticity: false`, `filter: {useKalmanFilter: false}`

Partial `filter` maps are merged over `LocationFilter`'s own defaults by
`LocationFilter.fromMap`, so both presets inherit
`trackingAccuracyThreshold: 100`, `maxImpliedSpeed: 80`, `policy: adjust`.

### What the app actually ends up with (`tracelet_position_source.dart`)

`_foregroundConfig()` = `Config.highAccuracy()` with `geo.copyWith(distanceFilter: 0.0)`:
- `distanceFilter` **0.0** — the preset's own 5 m displacement gate is explicitly
  deleted. This is the single most direct cause of stationary jitter accumulating:
  with no minimum displacement, every multipath-perturbed fix is recorded.
- `trackingAccuracyThreshold` **100 m** — a fix reported at 90 m accuracy is accepted.
  For pedestrian recording this is meaningless; ~20–30 m is the realistic ceiling.
- `maxImpliedSpeed` **80 m/s** — 288 km/h. No walking-speed outlier can ever trip it.
  For hiking ~5 m/s (or ~10 m/s with headroom) is the useful value.
- `policy` **adjust** — even a fix that DOES breach a threshold is corrected and
  still recorded, never dropped. `ignore`/`discard` are the drop semantics.
- `useKalmanFilter` **true** (inherited) — so the foreground profile does smooth,
  which is why the starburst is jitter-with-smoothing rather than worse.
- `enableDeadReckoning` **true** (inherited), activation delay **0 s**, max duration
  **unlimited** — in a street canyon this can inject IMU-estimated positions
  the moment GPS degrades. Pedestrian dead reckoning with a handheld phone
  while standing still is itself a jitter source. Never a deliberate choice
  here; it arrived with the preset.

`_backgroundConfig()` = `Config.balanced()` with
`geo.copyWith(desiredAccuracy: high, distanceFilter: 5.0)`:
- `useKalmanFilter` **false** — the balanced preset turns Kalman off and the
  `copyWith` does not restore it. **A recording with the screen off / app
  backgrounded loses GPS smoothing entirely**, while gaining a 5 m distance
  filter. The two profiles are filtering asymmetrically, and neither was
  deliberately tuned for pedestrian recording.
- `rejectMockLocations` falls back to **false** here (highAccuracy sets it true).
- `enableAdaptiveMode` **true** — auto-scales `distanceFilter` behind your back.

### Why this reframes the root cause

The defect is not "the app forgot to filter". It is **the app disabled the
displacement filter its own preset provides, and never tuned the quality
thresholds down from tracelet's vehicle-oriented defaults**. The correct fix is
almost entirely configuration in `_foregroundConfig()`/`_backgroundConfig()`,
not new hand-rolled gating code in `Navigation.onPosition`.

Note on `distanceFilter`: it was set to `0.0` deliberately (the doc comment says
"continuous tracking with no distance filter while moving"), presumably so slow
walking still produces points. Raising it trades jitter suppression against
sampling density on switchbacks — the same trade-off that got the CONV-05 5 m
distance gate removed (see Hard Constraint). The difference is that
`distanceFilter` gates at *acquisition* against the last recorded fix, whereas
CONV-05 gated at *measurement* over an already-recorded track; but the switchback
concern is real and a value here should be validated, not guessed. The
accuracy/implied-speed/policy knobs carry no such trade-off and are the safer
first move.

## Resolution

- root_cause: The app's tracelet configuration in
  `app/lib/services/tracelet_position_source.dart` disables and under-tunes the
  position filtering tracelet already provides:
  (1) `_foregroundConfig()` overrides the `Config.highAccuracy()` preset's
  `distanceFilter: 5.0` with `distanceFilter: 0.0`, removing the minimum-
  displacement gate so every fix is recorded regardless of how little the device
  moved; (2) `GeoConfig.filter` is never configured, leaving
  `trackingAccuracyThreshold` at 100 m and `maxImpliedSpeed` at 80 m/s — both
  vehicle-scale defaults that no pedestrian multipath spike can breach;
  (3) `LocationFilterPolicy` stays at `adjust`, which corrects rather than drops
  a rejected fix, so nothing is ever discarded; (4) the foreground and background
  profiles filter asymmetrically — `useKalmanFilter` is true in foreground
  (inherited from highAccuracy) but false in background (the balanced preset sets
  it false and the app's `copyWith` does not restore it), so a backgrounded
  recording loses smoothing entirely; (5) `enableDeadReckoning: true` is
  inherited from the highAccuracy preset with a 0 s activation delay and
  unlimited duration, which in a GPS-degraded street canyon can inject
  IMU-estimated motion while the user stands still.
  In the tester's dense street canyon (Karl-Heine-Straße/Hähnelstraße, Leipzig),
  multipath-perturbed fixes therefore pass straight through into the breadcrumb
  and produce the starburst. The downstream finding from the first pass still
  holds and compounds this — `Navigation.onPosition` appends unconditionally and
  `session_gap_backfill.dart` gates only on `isMoving` — but the primary lever is
  the tracelet config, not new app-side gating code.
- fix: n/a — diagnose-only session, no fix applied. Fix direction, cheapest first:
  1. Configure `GeoConfig.filter` on BOTH profiles with pedestrian-scale values —
     `trackingAccuracyThreshold` ~20–30 m, `maxImpliedSpeed` ~5–10 m/s, and
     `policy: LocationFilterPolicy.ignore` (or `discard`) so breaching fixes are
     dropped rather than "adjusted" into the track. No switchback trade-off.
  2. Set `useKalmanFilter: true` explicitly on the background profile so a
     screen-off recording keeps the smoothing the foreground profile has.
  3. Re-evaluate `distanceFilter: 0.0` in `_foregroundConfig()`. A small non-zero
     value restores the preset's gate, but validate against the
     `fixtures/gpx-corpus/12-dense-switchback` concern before picking a number.
  4. Consider `enableDeadReckoning: false` for the foreground profile unless
     tunnel/canyon dead reckoning is actually wanted — it was inherited, not chosen.
  None of this touches the removed CONV-05 distance smoothing or the load-bearing
  elevation `thresholdXY_m` / `GpxMetricsComputation(5, 5)` filter (Hard Constraint
  above): these are acquisition-time settings in the tracking SDK, not
  measurement-time gates over an accumulated track.
- verification: n/a — diagnose-only. Field verification will need a re-record in a
  comparable urban canyon; raw GPX from the tester's original session would let the
  thresholds be picked from measured accuracy values rather than estimated.
- files_changed: (none — diagnose-only session)

## Corrections to the first pass

- "No filter exists anywhere in the pipeline" — wrong at the SDK layer. tracelet
  provides `LocationFilter` (accuracy, implied-speed, mock, Kalman) and
  `GeoConfig.distanceFilter`. The app leaves them at defaults or turns them off.
- "The tracking SDK is closed-source and opaque" — the enforcement is native, but
  the full config surface, every default, and all four preset profiles are
  readable in the Dart package and were read for this revision.
- "The fix belongs in `Navigation.onPosition`" — superseded. The primary fix is
  configuration in `tracelet_position_source.dart`. App-side gating is at most a
  secondary defense.
