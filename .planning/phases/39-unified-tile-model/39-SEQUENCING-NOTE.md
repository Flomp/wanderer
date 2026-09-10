# Phase 39 — Sequencing deviation: Plan 03 runs before Plan 01's risk gate

**Recorded:** 2026-09-09, during execution
**Approved by:** developer, explicitly, mid-execution

## What changed

The standard wave barrier would hold Wave 2 (Plan 03) until Wave 1 (Plans 01 + 02) fully
completes, including Plan 01's blocking `checkpoint:decision`. **Plan 03 was executed before
that checkpoint was answered.** Plan 02 had already completed normally.

## Why

The risk gate is untestable in the pre-Plan-03 build, so running it as written would have
produced a false verdict.

`OnlineFileRequest::networkIsReachableAgain()` re-schedules **only** requests whose last
failure was `Reason::Connection` (`online_file_source.cpp:587`). In the pre-Plan-03 build,
every tile URL in the composed style points at `127.0.0.1`. Loopback is reachable even in
airplane mode — that is precisely why `MapLibre.setConnected(true)` is pinned
(`MainActivity.kt`). A coverage miss therefore returns **404** → `Reason::NotFound` →
`Duration::max()`, never retried, and persisted as an empty tile row.

Consequence: there is no `Reason::Connection` failure anywhere in the harness, so the gate's
step-6 question ("do the previously-blank areas fill in?") resolves to *no* whether or not the
pulse works. The gate would have recorded `setstyle` and wired the heavy fallback for no
reason.

Connection-failing tile requests only exist once the proxy redirects uncovered tiles upstream
— Plan 03's change. That is also the only configuration in which the load-bearing half of the
gate is observable at all, because it needs downloaded-region tiles (loopback) and uncovered
tiles (CDN) rendering in the same style simultaneously.

## Why this still honors the gate's intent

ROADMAP's risk gate exists to prove the recovery mechanism "before the phase invests in the
style/provider collapse". The collapse is Waves 4–6 (Plans 05, 06, 07, 08). Plan 03 is the
proxy itself, and is a prerequisite of the gate being answerable rather than part of the work
the gate protects. **Waves 4–6 remain behind the gate.**

## Current state

- Plan 01: tasks 1–3 committed (`193a9c81`, `aa59de04`, `0c298443`); checkpoint OPEN.
  No `39-01-SUMMARY.md` yet — deliberately, since Plan 09 Task 2 reads its
  `## Risk gate outcome` section and must not read a fabricated value.
- Plan 02: complete.
- Plan 03: executing ahead of the gate under this deviation.

## Open issue carried into the re-test

During the first gate attempt the harness rendered a black screen after Connect → Load map.
Not yet diagnosed. `map_source_persistence.dart` and the `objectBoxProvider` override were
both checked and ruled out. The discriminating signal is whether the harness log contains
`onStyleLoaded fired`:

- absent → style never finished loading (suspect `file://` sprite/glyph resolution)
- present → style loaded and every tile 404'd (no downloaded regions in that device's store)

Plan 03's redirect-on-miss should make the second case self-resolving, since uncovered tiles
will reach the CDN instead of 404ing.
