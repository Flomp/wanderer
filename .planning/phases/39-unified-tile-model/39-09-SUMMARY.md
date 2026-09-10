---
phase: 39-unified-tile-model
plan: 09
subsystem: infra
tags: [flutter, maplibre, android, platform-channel, verification]

# Dependency graph
requires:
  - phase: 39-unified-tile-model (plan 01)
    provides: "the setConnected pulse mechanism this plan deletes, and the `## Risk gate outcome` verdict that justifies deleting it"
  - phase: 39-unified-tile-model (plan 08)
    provides: "the collapsed provider and renamed transform, so this plan verifies a finished phase rather than a half-migrated one"
provides:
  - "MainActivity.kt carrying only the MapLibre.setConnected(true) pin, with a comment that describes what the code actually does"
  - "device verification of ROADMAP criteria 1, 2, 5 and 6"
affects: []

# Tech tracking
tech-stack:
  removed: ["Flutter MethodChannel com.openwanderer.wanderer/maplibre_connectivity"]
---

## Accomplishments

Removed the rejected recovery mechanism and closed the phase with on-device verification.

Plan 01's risk gate returned `setstyle`: the `setConnected(false)` -> `setConnected(true)` pulse
fires correctly (2ms channel round trip, `result.success(null)`, no `PlatformException` and no
`MissingPluginException` in logcat) and MapLibre re-schedules nothing — 82 Connection-class tile
failures before the pulse, **zero** tile requests after. The developer then declined the
documented `setStyle` reload fallback as disproportionate (it rebuilds every source, layer and
image, dropping and re-adding the navigation screen's trail track and breadcrumb on every
connectivity regain) and accepted user-initiated recovery instead — CONTEXT.md D-12a.

This plan therefore deleted rather than wired. D-15's re-probe loop was descoped with it, since
it existed only to trigger the recovery.

## Task Commits

- `c09aa100` — feat(39-09): delete the rejected setConnected pulse mechanism

## Files Created/Modified

- `app/lib/services/maplibre_connectivity_pulse.dart` — **deleted**
- `app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt` — `MethodChannel`,
  `pulseConnected` handler, `configureFlutterEngine` override, `CHANNEL` companion object and two
  now-unused Flutter imports removed. `MapLibre.getInstance` and `MapLibre.setConnected(true)`
  retained unchanged — the loopback proxy still needs the pin to stay reachable when the radio
  reports no network. Comment rewritten to justify only the pin and to record the accepted
  consequence: the permanent pin means MapLibre never self-retries a Connection-failed tile, so
  recovery is user-initiated.
- `app/test/services/tile_proxy_spike_harness.dart` — pulse button, `_pulseConnectivity()`, its
  offline-refusal guard, test case (f) in the header comment, and two now-unused imports removed.

## Decisions Made

- D-12b honored: the mechanism is removed, not retained-and-marked-rejected. An unused channel
  whose doc comment describes a rejected mechanism is the exact hazard D-14 exists to prevent.
- The `setConnected(true)` pin survives. Only the post-startup mutation of it is gone.

## Deviations from Plan

**Executed inline by the orchestrator rather than by a `gsd-executor` subagent.** The dispatched
executor was terminated by a provider session rate limit after 4 tool calls, having written
nothing (tree clean at `e4494af8`, no partial commits, no SUMMARY). Task 1 is a fully specified
deletion, so the orchestrator completed it directly rather than waiting for the quota reset. No
scope change.

## Issues Encountered

A `git add` invocation aborted on the already-`git rm`'d pulse-service path, so the first commit
(`fbd35117`) captured only the deletion and left `MainActivity.kt` and the harness unstaged. Fixed
by staging both and amending; the amended commit `c09aa100` carries all three files. No
intermediate state was pushed.

## Device verification

Reported by the developer on a physical Android device. All four device-observable ROADMAP
criteria **PASS**.

| Criterion | Result | Notes |
|---|---|---|
| 1 — the recovery case | **PASS** | Online tiles appear around the downloaded region on the same screen after a pan or zoom, without reopening the trail. Automatic fill-in is not expected under D-12a. |
| 2 — no offline regression | **PASS** | Basemap, place-name labels, category icons and hillshade all render in airplane mode, at every zoom, with labels and icons now served through the proxy's glyph/sprite routes rather than a pre-warmed `file://` cache. |
| 5 — cache survives a cold start | **PASS** | Ground panned over online still renders after a force-stop and relaunch in airplane mode, confirming the persisted loopback port keeps MapLibre's ambient-cache keys stable across launches. |
| 6 — no thermal regression | **PASS** | Continuous panning feels unchanged, confirming the redirect design kept the extra loopback hop cheap — bytes never transit the root isolate. |

Criteria 3 and 4 are structural and were asserted by grep gates in Plans 03, 06, 07 and 08, then
re-verified at phase close: zero references to `widget.offline` / `widget.isOffline`, zero to
`offlineMapStyleJsonProvider` / `offlineGlyphSpritePaths`, zero to `rewriteStyleForOffline`.

## Incidental confirmation

The gate's logcat evidence also confirmed, on real hardware, that MapLibre Native **follows the
302 redirect from loopback to the CDN** on Android. The composed style only ever pointed at
`127.0.0.1`, so the `Unable to resolve host "api.protomaps.com"` failures could only arise by
following Plan 03's redirect. That was the single load-bearing assumption of the whole
architecture (RESEARCH.md §4.1) and had previously been verified only by reading MapLibre source.

## Known limitations

`trail_panel.dart` mounts `TrailMap(disabled: true, embedded: true, …)`, and `disabled` resolves
to `MapGestures.none()`. That map cannot be pan-recovered — its fixed camera never requests new
tile coordinates — so it recovers only when the widget remounts on navigation. Deliberately
accepted (CONTEXT.md D-12a); still strictly better than before the phase, where it never
recovered at all.

## Next Phase Readiness

Phase 39 is complete. No follow-on work is blocked on it.

Two items carried forward as documented, not as debt:
- Restoring true z15 online detail, traded away by the D-08 maxzoom pin.
- iOS binary verification of the redirect/delegate findings against the shipped `MapLibre ~> 6.25`
  pod — research was against MapLibre Native `main`. Android was verified against the shipped
  `13.0.3-pre0` AAR and now on hardware.

## Self-Check: PASSED

`flutter analyze` — 13 pre-existing info-level issues in vendor code, unchanged. `flutter test` —
1134 passed + 1 skip, exactly the post-Plan-08 baseline, no regressions. No `flutter build` or
`adb install` was run by the orchestrator; the developer built and installed for the device checks.
