---
phase: 39-unified-tile-model
plan: 01
subsystem: infra
tags: [flutter, android, kotlin, maplibre, methodchannel, connectivity]

# Dependency graph
requires: []
provides:
  - "Android MethodChannel `com.openwanderer.wanderer/maplibre_connectivity` / `pulseConnected` (dead code as of this SUMMARY — see Risk gate outcome)"
  - "pulseMapLibreConnectivity() Dart entry point in app/lib/services/maplibre_connectivity_pulse.dart (dead code as of this SUMMARY — see Risk gate outcome)"
  - "Retired ROADMAP risk gate: the recovery mechanism for Connection-failed tiles is `setStyle` reload, not the `setConnected` pulse"
affects: [39-09]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "First MethodChannel/configureFlutterEngine precedent in the repo's only Kotlin file (MainActivity.kt) — superseded by 39-09's removal, but the pattern (single no-arg method, result.notImplemented() default branch, no call.arguments read) is the template if a future channel is ever needed"

key-files:
  created:
    - app/lib/services/maplibre_connectivity_pulse.dart
  modified:
    - app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt
    - app/test/services/tile_proxy_spike_harness.dart

key-decisions:
  - "Risk gate verdict: setstyle. The setConnected(false)->true pulse does NOT recover Connection-failed tiles on a physical Android device; Plan 39-09 wires the setStyle reload fallback instead."
  - "Follow-on decision (developer, post-gate): the pulse channel is now dead code. Plan 39-09 will REMOVE maplibre_connectivity_pulse.dart, the Kotlin MethodChannel handler + configureFlutterEngine override, the harness button, and revert MainActivity.kt's comment to describe only the pin."
  - "Harness hardening: a1df4ddf added an online-probe guard so the harness refuses to fire a pulse while the backend is unreachable, after a first gate attempt produced an indeterminate false-negative (pulse fired 16s into an offline period, no in-flight requests existed to revive)."

patterns-established:
  - "On-device risk gates that depend on transient network state need a guard against firing the probe during the wrong window (see a1df4ddf) — an invalid gate attempt is worse than a slow one, since it can be silently misread as a negative result."

requirements-completed: [D-12, D-13, D-14]

# Metrics
duration: ~2min (Tasks 1-3 automatable work) + on-device verification wait + ~5min closing
completed: 2026-09-09
---

# Phase 39 Plan 01: setConnected Pulse Risk Gate Summary

**Built the Android setConnected pulse behind a MethodChannel (D-12/13/14) to test it as the ROADMAP recovery mechanism for Connection-failed tiles — physical-device testing proved it does not work, retiring the risk gate in favor of the documented setStyle reload fallback, which Plan 39-09 will wire while Plan 39-09 also removes this plan's now-dead pulse code.**

## Performance

- **Duration:** ~2 min for Tasks 1-3's automatable work (22:12:21–22:13:56 CEST), plus a discarded first gate attempt, a harness hardening commit (`a1df4ddf`, 22:55:50), an on-device verification wait, and this closing pass
- **Started:** 2026-09-09T22:12:21+02:00
- **Completed:** 2026-09-09 (this closing pass)
- **Tasks:** 3 (Task 3 is the blocking `checkpoint:decision`, now answered)
- **Files modified:** 3 (1 created, 2 modified)

## Accomplishments
- `MainActivity.kt` gained a `configureFlutterEngine` override registering a single-method, no-argument `pulseConnected` handler that runs `MapLibre.setConnected(false)` immediately followed by `MapLibre.setConnected(true)` on the same main-thread turn, and its `setConnected(true)` pin comment was rewritten to state the true post-D-01 justification (the pin keeps the loopback proxy reachable in airplane mode; the pulse is the only way `networkIsReachableAgain()` ever fires while it stays pinned).
- `app/lib/services/maplibre_connectivity_pulse.dart` added `pulseMapLibreConnectivity()`, Android-only by construction (`!Platform.isAndroid` early return, D-13), swallowing `PlatformException`/`MissingPluginException` so a failed pulse never propagates to a connectivity-listener caller.
- `tile_proxy_spike_harness.dart` gained a `Pulse connectivity (Android)` control wired to the real production entry point, plus (in `a1df4ddf`) an online-probe guard refusing to fire a pulse while the backend is unreachable — added after a first, indeterminate gate attempt.
- The ROADMAP risk gate was run on a physical Android device against live loopback tile traffic and retired with a recorded, evidence-backed verdict (below), rather than an assumption.

## Task Commits

Each task was committed atomically:

1. **Task 1: Kotlin pulse handler and rewritten pin comment** - `193a9c81` (feat)
2. **Task 2: Dart pulse service, Android-only by construction** - `aa59de04` (feat)
3. **Task 3: Risk gate harness control** - `0c298443` (feat)

Landed after the checkpoint was raised, before it was answered:

- **Harness hardening (online-probe guard)** - `a1df4ddf` (test) — orchestrator-side fix after a first, invalid gate attempt (pulse fired while offline, producing a result indistinguishable from failure).

**Plan metadata:** (this commit)

## Files Created/Modified
- `app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt` - `configureFlutterEngine` MethodChannel handler for `pulseConnected`; rewritten pin comment
- `app/lib/services/maplibre_connectivity_pulse.dart` - `pulseMapLibreConnectivity()`, Android-only, never throws
- `app/test/services/tile_proxy_spike_harness.dart` - `Pulse connectivity (Android)` control, test case `(f)` procedure, online-probe guard

## Risk gate outcome

**Verdict: `setstyle`.**

The `setConnected(false)` -> `setConnected(true)` pulse was tested on a physical Android device
and does **NOT** recover Connection-failed tiles. Plan 39-09 must implement the `setStyle`
reload variant documented as the fallback in `39-RESEARCH.md` section 4.3, not the pulse.

### Evidence (device logcat, 2026-09-09, device clock CEST)

**First attempt — INVALID, discarded:**
- Pulse #1 at 22:49:17.749 fired while WiFi was still connected; no failed requests existed to
  revive.
- Pulse #2 at 22:51:36.934 fired 16s into an offline period; WiFi never came back in the buffer.
  A pulse fired offline re-schedules requests into an immediate second failure, which is
  indistinguishable from the mechanism not working.
- This false-negative risk is why `a1df4ddf` added an online-probe guard to the harness that
  refuses to fire a pulse while the backend is unreachable.

**Second attempt — VALID, and the basis for this verdict:**
- `22:57:14` proxy started at `http://127.0.0.1:62875/<32-hex-secret>` (Plan 03's stable port +
  secret).
- `22:57:27` style composed via `rewriteStyleForProxy`; `onStyleLoaded fired`.
- `22:57:40`–`22:58:04`: **82** tile failures, all `Unable to resolve host` — 71 for
  `api.protomaps.com` (source `protomaps`), 11 for `tiles.mapterhorn.com` (source
  `hillshadeSource`).
- `22:58:36.116` `online confirmed -- pulse ARMED` (the new guard verified backend reachability
  first).
- `22:58:36.118` `pulseMapLibreConnectivity() -- returned` — a 2ms channel round trip with no
  `PlatformException` and no `MissingPluginException` in logcat, proving the call genuinely
  reached Kotlin and got `result.success(null)`.
- **After the pulse: zero tile requests.** Not failures — no requests at all. The developer
  confirmed tiles did not reappear until a manual zoom.

### Likely mechanism (hypothesis, evidence-consistent, NOT confirmed)

`networkIsReachableAgain()` iterates *live* `OnlineFileRequest` objects. By pulse time those
requests were already gone — logcat shows `Request failed due to a permanent error: Canceled`
at `22:57:53` and `22:58:04`, both BEFORE the pulse. The failure is retained in the source's
tile pyramid as an errored tile rather than as a pending request, so a connectivity edge has
nothing to act on. This explains why a manual zoom works (new tile coordinates create fresh
requests) and why a `setStyle` reload works (`Style::Impl::loadJSON` does `sources.clear()`,
discarding the errored pyramid).

### Caveat (record honestly)

It could NOT be verified that `UnknownHostException` maps to `Reason::Connection` in this build.
The three `org/maplibre/android/module/http/*` classes in the shipped
`android-sdk-opengl-13.0.3-pre0.aar` contain no reference to `UnknownHostException`,
`NoRouteToHostException`, `SocketException` or `SSLException`, so the classification happens
C++-side and was not confirmed from the artifact. If DNS failures land in a class other than
`Connection`, `networkIsReachableAgain()` would be correctly skipping them — meaning the pulse
is the wrong tool for this failure mode rather than a broken one. The verdict is unaffected;
the D-12 rationale in `39-CONTEXT.md` is weaker than it currently reads.

### Follow-on decision (developer, made after the gate)

The pulse channel is now dead code. Plan 39-09 will **REMOVE** it:
`maplibre_connectivity_pulse.dart`, the Kotlin `MethodChannel` handler and
`configureFlutterEngine` override, the harness button, and `MainActivity.kt`'s comment
(reverting it to describe only the pin, since the pulse justification for keeping the
comment's second half no longer applies). This removal was explicitly deferred to Plan
39-09 and was NOT performed by this closing pass.

## Decisions Made
- Risk gate verdict is `setstyle`, recorded verbatim above per the plan's `<output>` contract — Plan 39-09 Task 2 reads this section.
- The pulse implementation (Tasks 1-3 of this plan) is retained in the tree, unmodified, as historical record of the gate; its removal is explicitly Plan 39-09's responsibility, not this closing pass's.
- The harness's online-probe guard (`a1df4ddf`) stays — it is generally useful hardening for any future on-device gate, independent of the pulse's fate.

## Deviations from Plan

None beyond what the plan itself anticipated (the blocking checkpoint). The harness hardening
commit (`a1df4ddf`) is a Rule 1 auto-fix (the first gate attempt was invalid due to a timing
race between the pulse and the offline window) applied by the orchestrator between Task 3's
commit and the checkpoint being answered; it is recorded here for completeness rather than as
a new deviation.

## Issues Encountered
- First on-device gate attempt was invalid (pulse fired during an offline window, producing an
  indeterminate result) — resolved by adding an online-probe guard (`a1df4ddf`) before re-running
  the procedure. See Risk gate outcome above for the full evidence trail.
- The `UnknownHostException` -> `Reason::Connection` classification could not be confirmed from
  the shipped `.aar` (C++-side classification, not present in the decompiled Kotlin/Java
  classes) — documented as a caveat rather than resolved, since it does not change the verdict.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Plan 39-09 has an unambiguous, evidence-backed verdict to read from this SUMMARY's
  `## Risk gate outcome` section: wire `setStyle` reload, not the pulse.
- Plan 39-09 additionally owns removing this plan's now-dead pulse code (Dart service, Kotlin
  handler, harness button, and reverting the `MainActivity.kt` comment) — noted as a forward
  reference here, not performed by this closing pass.
- No render-path production behavior changed by this plan; the pulse code is inert on every
  path except the harness's dedicated button.

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-09*

## Self-Check: PASSED

All four task/harness commit hashes (193a9c81, aa59de04, 0c298443, a1df4ddf) verified present in
git log. SUMMARY.md file confirmed on disk.
