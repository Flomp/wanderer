---
slug: current-fix-times-out-stationary
status: awaiting_human_verify
trigger: "currentFix(timeout: 20s) on foregroundPositionStreamProvider frequently times out in trail_source_select_screen.dart:103-105, even when location was already established elsewhere (e.g. map_screen). Hypothesis: the position stream only emits on movement, so when stationary no new fix is pushed and currentFix waits forever on a stream that never emits — a cached/last-known fix is likely not being used."
created: 2026-09-09
updated: 2026-09-09
---

# Debug: currentFix times out while stationary

## Symptoms

- **Expected:** Tapping the recorder entry in `trail_source_select_screen` resolves a GPS
  fix promptly and pushes `/record`, especially when a fix was already acquired earlier in
  the session (e.g. the user came from `map_screen`, where the location marker was showing).
- **Actual:** `currentFix(timeout: 20s)` returns `null` a large fraction of the time; the
  screen shows `l10n.location_unavailable` and the recorder never opens. Reproduces even
  though the device demonstrably had a fix moments earlier on another screen.
- **Errors:** No exception surfaced — `currentFix` swallows the `TimeoutException` and
  returns `null` by design (`foreground_position_stream_provider.dart:185`).
- **Timeline:** Present after the GPS reference-counting rework that replaced the always-on
  ghost subscription with `acquire()`/`release()` (see the thermal-profile work).
- **Reproduction:** Sit still. Open `map_screen`, let the blue dot appear. Navigate back to
  the trail source select screen, tap the recorder option. ~20s spinner, then
  "location unavailable".

## Current Focus

hypothesis: |
  `currentFix` awaits a *future* emission on a broadcast stream that, by design, emits
  nothing while the device is stationary — and it has no access to the fix already
  received. Three code facts compound:

  1. `_settings` uses `distanceFilter: 10` metres, and the code comment states outright
     that "a stationary device produces no callbacks at all"
     (`foreground_position_stream_provider.dart:98-100`). On iOS this is worse:
     `pauseLocationUpdatesAutomatically: true` + `ActivityType.fitness` lets Core Location
     park the receiver entirely when the user is still.
  2. `_controller` is a plain `StreamController.broadcast` — no last-value replay. A
     subscriber that attaches after the last emission sees nothing, so the fix `map_screen`
     already displayed is unreachable to `currentFix`
     (`foreground_position_stream_provider.dart:184, 258`).
  3. `acquire()` only calls `_scheduleStart()` when `_consumers++ == 0`
     (`foreground_position_stream_provider.dart:147-149`). If `map_screen` is still mounted
     in the back stack it holds its `acquire()` (acquired at `map_screen.dart:116`, released
     only in dispose at `map_screen.dart:217`), so `currentFix` piggybacks on the existing,
     stationary-silent subscription instead of starting a fresh `getPositionStream` — and a
     fresh subscription is the one path that reliably yields an immediate first fix. This
     explains why the failure correlates with *having already been on the map screen*,
     which is otherwise counterintuitive.

  Nothing anywhere calls `Geolocator.getLastKnownPosition()`.

test: |
  Instrument or reason through: (a) confirm `_consumers > 0` at the moment
  `trail_source_select_screen` calls `currentFix` when arriving from `map_screen`;
  (b) confirm no emission reaches the controller during the 20s window while stationary;
  (c) confirm `Geolocator.getLastKnownPosition()` returns a usable fix in the same window.

expecting: |
  A stationary device with map_screen in the back stack times out at 20s while
  `getLastKnownPosition()` would have returned a valid recent fix immediately.

next_action: |
  Fix applied and self-verified (analyze clean, full test suite green). Awaiting on-device
  confirmation from the user — they build/install and re-run the original repro (sit still,
  open map_screen until the blue dot appears, navigate to trail source select, tap
  recorder). Report back "confirmed fixed" or what's still failing.

reasoning_checkpoint:
  hypothesis: |
    `currentFix()` fails to return an already-available fix because it only ever awaits a
    *future* stream emission, and the underlying `getPositionStream` (distanceFilter: 10m,
    iOS pauseLocationUpdatesAutomatically) produces no callback at all while the device is
    stationary. When map_screen sits in the back stack, `LocationMarkerLayer`'s continuous
    `ref.watch(liveLocationProvider)` keeps `_consumers > 0`, so `currentFix`'s `acquire()`
    piggybacks on that already-silent subscription instead of starting a fresh one — the one
    scenario (fresh `getPositionStream` call) that reliably delivers an immediate first fix.
  confirming_evidence:
    - "Code comment explicitly documents the stationary-silence behaviour as intentional (foreground_position_stream_provider.dart:98-100)."
    - "acquire() only calls _scheduleStart() on the 0→1 transition (line 147-149); LocationMarkerLayer's permanent watch of liveLocationProvider keeps consumers >0 while MapScreen is merely covered, not disposed, by a pushed route."
    - "launch_navigation.dart independently diagnosed and fixed the identical symptom at its own call site using Geolocator.getLastKnownPosition() as primary source, with an inline comment describing the exact same stationary/distanceFilter failure mode."
  falsification_test: |
    If `Geolocator.getLastKnownPosition()` reliably returned null immediately after map_screen
    displayed a fix (i.e. the OS cache were empty/unpopulated even though a fix was just
    shown), the proposed fix would not help and the hypothesis would need revisiting. This
    is contradicted by launch_navigation.dart's shipped behavior, which relies on exactly
    this call succeeding in the equivalent scenario.
  fix_rationale: |
    Query the OS's cached last-known position first (near-instant, no new GPS callback
    required) and only fall back to the existing stream-wait/timeout when no cached fix
    exists. This addresses the root cause (waiting on a stream that structurally cannot emit
    while stationary) rather than symptom-patching via a longer timeout. Centralizing in
    currentFix() (rather than duplicating launch_navigation.dart's inline pattern) fixes all
    three call sites — including the two shorter-timeout ones flagged as likely
    under-affected — with one change.
  blind_spots: |
    Not yet device-verified: how stale `getLastKnownPosition()` can be in practice on this
    fleet of test devices, and whether Android/iOS ever return a permission error rather than
    null when permission is technically granted but not yet "fresh" (mitigated by wrapping in
    try/catch and falling through to the existing stream path, mirroring launch_navigation.dart).
    No unit test harness exists for this Notifier (direct `Geolocator` static calls, not
    injected), so verification is analyze/test (compile+lint) plus human on-device check, not
    an automated regression test.

## Evidence

- timestamp: 2026-09-09 (continuation — code inspection)
  observation: |
    Correction to contributor 3: `map_screen.dart`'s one-shot seed `acquire()` (initState,
    line 116) does NOT hold the consumer until dispose as previously recorded. Its listener
    calls `_cancelGpsSeed()` — which calls `release()` — on the *first* non-null position
    (`if (position == null || _resolvedGpsCenter != null) return; _cancelGpsSeed();`), i.e.
    the moment the blue dot first appears. So by the time the repro says "let the blue dot
    appear", this particular acquire is already released.
  file: app/lib/routes/map_screen.dart:117-124, 209-218

- timestamp: 2026-09-09 (continuation — code inspection)
  observation: |
    The actual persistent hold comes from a different consumer: `LocationMarkerLayer`
    (unconditionally rendered inside `MapScreen`'s body) calls `ref.watch(liveLocationProvider)`
    every build. `liveLocationProvider` is `Provider.autoDispose` and calls
    `notifier.acquire()` on creation / `release()` in `ref.onDispose` — i.e. it stays acquired
    for as long as the widget is watching it. Because `context.push()` keeps the previous
    route (`MapScreen`) mounted underneath the new route rather than disposing it,
    `LocationMarkerLayer` keeps watching and `_consumers` never drops to 0 while `map_screen`
    sits in the back stack. This confirms contributor 3 with a different, stronger mechanism
    than originally logged (a continuous live-marker watch, not the one-shot seed).
  file: app/lib/components/map/location_marker_layer.dart:119-150, app/lib/routes/map_screen.dart:611-616 (LocationMarkerLayer in build), foreground_position_stream_provider.dart:65-71

- timestamp: 2026-09-09 (continuation — code inspection, strong corroboration)
  observation: |
    `launch_navigation.dart`'s `_openRecorder`-equivalent path already independently
    discovered and fixed this exact defect for its own seed fetch. It calls
    `Geolocator.getLastKnownPosition()` first, and only falls back to
    `currentFix(timeout: 3s)` if that returns null, with an inline comment stating: "a
    stationary device ... never produces a new callback at all under the stream's 10m
    distance filter, so that wait almost always just burned its timeout." This is the same
    root cause independently diagnosed and worked around at one call site, but never
    centralized into `currentFix()` itself — so `_openRecorder` and `_resolveInitialCenter`
    in `trail_source_select_screen.dart` never got the fix.
  file: app/lib/actions/launch_navigation.dart:148-173

- timestamp: 2026-09-09 (orchestrator pre-read, code inspection only — not yet device-verified)
  observation: |
    `currentFix` body is `acquire(); await state.firstWhere((p) => p != null).timeout(timeout)`.
    `state` is the broadcast stream with no replay buffer; `firstWhere` therefore only ever
    sees emissions that occur strictly after subscription.
  file: app/lib/provider/foreground_position_stream_provider.dart:179-190

- timestamp: 2026-09-09 (orchestrator pre-read, code inspection only)
  observation: |
    `static LocationSettings get _settings` sets `distanceFilter = 10` on all platforms and
    documents the stationary-silence behaviour as intentional (it is what fixed the Play
    Services location churn). iOS additionally sets `pauseLocationUpdatesAutomatically: true`.
  file: app/lib/provider/foreground_position_stream_provider.dart:98-130

- timestamp: 2026-09-09 (orchestrator pre-read, code inspection only)
  observation: |
    `acquire()` starts the receiver only on the 0→1 transition. `map_screen` acquires on init
    and releases on dispose, so it holds a consumer while sitting in the back stack.
  file: app/lib/provider/foreground_position_stream_provider.dart:147-149, app/lib/routes/map_screen.dart:116,217

- timestamp: 2026-09-09 (orchestrator pre-read, code inspection only)
  observation: |
    Second caller in the same file uses a 4s timeout for the planner path
    (`_resolveInitialCenter`), gated on `allowAutoGeolocate`; `launch_navigation.dart:170-171`
    uses 3s. Both share the same defect, with a much smaller window — worth checking whether
    they silently fall back more often than intended.
  file: app/lib/routes/trail_source_select_screen.dart:157-159, app/lib/actions/launch_navigation.dart:170-171

## Eliminated

- hypothesis: |
    "map_screen holds its acquire() until dispose because it's the one-shot seed acquired at
    initState and released only in map_screen.dart:217" (original contributor-3 wording).
  evidence: |
    `_cancelGpsSeed()` (which calls `release()`) is invoked from the seed listener itself on
    the first non-null position, not only from `dispose()`. Since the repro explicitly waits
    for the blue dot to appear, this acquire is already released by the time the user
    navigates away. The real persistent hold is `LocationMarkerLayer`'s continuous
    `ref.watch(liveLocationProvider)` (see Evidence). Net effect on the hypothesis is
    unchanged (map_screen in the back stack still holds a consumer) but the mechanism was
    misattributed.
  timestamp: 2026-09-09

## Constraints

- Do not reintroduce a continuously-running foreground GPS receiver — the acquire/release
  reference counting exists because the receiver previously ran at 100% duty cycle for the
  whole app session (measured).
- Preserve the at-most-once-per-session "Turn on GPS?" dialog contract (`_promptedForService`).
- Recording (tracelet) owns a separate native session and must stay unaffected.
- `allowAutoGeolocate` must keep gating the planner entry point.

## Resolution

root_cause: |
  `ForegroundPositionStream.currentFix()` only awaits a future emission on the shared
  position stream. `getPositionStream` is configured with `distanceFilter: 10` (and, on iOS,
  `pauseLocationUpdatesAutomatically: true`), so it produces no callback while the device is
  stationary — documented as intentional in the file's own comments. When map_screen sits in
  the back stack, `LocationMarkerLayer`'s continuous `ref.watch(liveLocationProvider)` keeps
  the receiver's consumer count above zero, so `currentFix`'s `acquire()` attaches to that
  already-silent subscription instead of starting a fresh one, and `firstWhere` never
  resolves until the timeout. `currentFix()` never consulted
  `Geolocator.getLastKnownPosition()`, even though a usable fix already existed (the one the
  user just saw on the map). `launch_navigation.dart` had already discovered and worked
  around this identical defect locally, but the fix was never centralized into `currentFix()`
  itself.
fix: |
  `currentFix()` now checks `Geolocator.getLastKnownPosition()` first (mirroring the pattern
  already proven in `launch_navigation.dart`), returning it immediately when present AND no
  older than `_cachedFixMaxAge` (2 minutes) — without touching the receiver's acquire/release
  lifecycle at all in that fast path. A null or too-stale cached fix falls through to the
  existing acquire() + stream-wait + timeout path unchanged. Errors from
  `getLastKnownPosition()` are swallowed and also fall through. The staleness bound is a
  deliberate refinement beyond launch_navigation.dart's unconditional trust: that call site's
  fix is a best-effort live-marker seed that self-corrects the instant tracelet's own GPS
  reports, whereas `currentFix()`'s result can become the actual initial center for a
  recording session or route plan (`_openRecorder`, `_resolveInitialCenter`), so an old cached
  fix needs a bound rather than being trusted outright. This benefits all three callers
  without changing any of their timeouts or call signatures.

  Note: this file was edited concurrently on disk during the fix/verify step (observed via
  file-changed system reminders) — the version landed with the staleness bound is a refined
  superset of my initial change (which used getLastKnownPosition unconditionally, no
  staleness check). Verified it is self-consistent (`flutter analyze` clean, no unused
  fields) and addresses the same root cause before accepting it as-is, per instruction not to
  revert a concurrent change that looks correct.
verification: |
  `flutter analyze lib/provider/foreground_position_stream_provider.dart` — no issues found.
  `flutter test` (full suite, app/) — 1124 passed, 0 failed (no existing coverage of this
  Notifier specifically; it isn't unit-testable as written since it calls static `Geolocator`
  methods directly rather than through an injected seam).
  On-device confirmation pending — see CHECKPOINT REACHED.
files_changed:
  - app/lib/provider/foreground_position_stream_provider.dart
