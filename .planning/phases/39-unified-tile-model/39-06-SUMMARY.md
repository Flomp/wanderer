---
phase: 39-unified-tile-model
plan: 06
subsystem: infra
tags: [flutter, maplibre, riverpod, tile-proxy, style-composition]

# Dependency graph
requires:
  - phase: 39-unified-tile-model (plan 05)
    provides: "rewriteStyleForProxy as the cache-root-free unconditional style transform, and TrailMap's _composeStyle already narrowed to a single-argument signature calling it"
provides:
  - "TrailMap with no offline parameter, no compose branch, and no ref.listen fork — one style source (mapStyleJsonProvider), one listener, one compose path"
  - "Three call sites (trail_detail_map_screen.dart, trail_panel.dart, trail_create_screen.dart) that construct TrailMap without a mode argument"
affects: [39-07, 39-08]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Deletion over defaulting: a caller-supplied mode flag whose flip could not reach the mounted native view was removed outright rather than fixed in place, making the bug class (and the separate TrailMap(offline: trail.isOffline) conflation) a compile error instead of a runtime defect"

key-files:
  created: []
  modified:
    - app/lib/components/base/trail_map.dart
    - app/lib/routes/trail_detail_map_screen.dart
    - app/lib/components/trail/trail_panel.dart
    - app/lib/routes/trail_create_screen.dart

key-decisions:
  - "Left onlineStatusProvider-derived locals (isOnline in trail_panel.dart, isOffline in trail_create_screen.dart) and their imports in place — each still gates an unrelated onTap/app-bar consumer further down its file, so only the offline: TrailMap argument and its stale explanatory comment were removed."
  - "trail_detail_map_screen.dart's onlineStatusProvider import was removed — its only reference in that file was the deleted offline: argument."
  - "TrailMap's class/build doc comments were rewritten to state the new single-path contract and to record, for the next reader, exactly which shipped call shape (TrailMap(offline: trail.isOffline)) is now unrepresentable and why, rather than leaving that context only in RESEARCH.md."

requirements-completed: [D-01, D-16]

# Metrics
duration: ~15min
completed: 2026-09-09
---

# Phase 39 Plan 06: Delete TrailMap.offline Summary

**`TrailMap` now composes and watches exactly one style path (`mapStyleJsonProvider` → `rewriteStyleForProxy`) with no `offline` parameter, no compose branch, and no `ref.listen` fork — the three call sites that used to pass a mode no longer can.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-09-09T21:50:00Z
- **Completed:** 2026-09-09T22:05:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Deleted `TrailMap.offline`'s field declaration and constructor default outright — not defaulted, not deprecated. `TrailMap(offline: trail.isOffline)`, the shipped conflation of "downloaded" with "no connectivity" documented at `models/trail.dart:110-116`, is now a compile error.
- Collapsed `build()`'s `if (widget.offline) { ... } else { ... }` `ref.listen` fork to a single unconditional `ref.listen(mapStyleJsonProvider, (_, _) => _swapStyle())`, and its `baseAsync` ternary to a plain `ref.watch(mapStyleJsonProvider)`.
- Removed `_composeStyle`'s `if (!widget.offline) return baseJson;` early return, so the decode → `rewriteStyleForProxy` → encode path now runs for every style, online and offline alike (D-01) — verified by a sed-scoped grep for `return baseJson;` inside the method returning 0 matches.
- Collapsed `_swapStyle`'s read-ternary to the single `ref.read(mapStyleJsonProvider).value`.
- Rewrote `TrailMap`'s `_composeStyle` doc comment to describe the new single-path contract (coverage resolved per tile inside `tile_proxy_server.dart`, not by a widget-level mode) and to explicitly name what the deletion makes unrepresentable, per the plan's action step.
- Dropped the `offline:` argument, and its now-orphaned "Connectivity, NOT trail.isOffline" comment, from all three call sites: `trail_detail_map_screen.dart`, `trail_panel.dart`, `trail_create_screen.dart`.
- `trail_detail_map_screen.dart`'s `onlineStatusProvider` import was removed as unused (its only reference was the deleted argument). `trail_panel.dart`'s `isOnline` local/import and `trail_create_screen.dart`'s `isOffline` local/import were both retained — each still has a live consumer elsewhere in its file (an `onTap` guard and an app-bar action, respectively).
- Confirmed `app/lib/models/trail.dart` and `app/lib/routes/navigation_screen.dart` are untouched (`git diff --stat` empty for both) — the model is read-only reference material for this plan, and `navigation_screen.dart` is Plan 07's file in this same wave.

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove the offline parameter and both forks from TrailMap** - `acd9a167` (feat)
2. **Task 2: Drop the offline argument at all three call sites** - `5f52e76d` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `app/lib/components/base/trail_map.dart` — `offline` field/constructor param deleted; `build()`'s listen fork and `baseAsync` ternary collapsed to one path; `_composeStyle`'s early-return branch deleted so it always rewrites; `_swapStyle`'s read-ternary collapsed; class/method doc comments rewritten
- `app/lib/routes/trail_detail_map_screen.dart` — `offline: !ref.watch(onlineStatusProvider),` argument and its comment removed; now-unused `onlineStatusProvider` import removed
- `app/lib/components/trail/trail_panel.dart` — `offline: !isOnline,` argument and its comment removed; `isOnline` local/import kept (still used by an `onTap` guard at line ~340)
- `app/lib/routes/trail_create_screen.dart` — `offline: isOffline,` argument removed; `isOffline` local/import kept (still gates an app-bar route-planner action)

## Decisions Made

See `key-decisions` in the frontmatter. In summary: retained the two `onlineStatusProvider`-derived locals that have consumers beyond the deleted `TrailMap` argument, removed the one import that had no other reference, and used the doc-comment rewrite required by the plan to record both the new single-path contract and the specific call shape it makes unrepresentable.

## Deviations from Plan

None — plan executed exactly as written. All acceptance-criterion greps/seds pass as specified in the plan (see Issues Encountered for one grep whose literal-zero expectation was not met by design, not by omission).

## Issues Encountered

The plan's Task 2 acceptance criterion `grep -rn "offline:" lib/ --include='*.dart' | grep -v ... | wc -l` returns 0 was checked and returns 6, not 0. All six are prose-comment matches on the substring `offline:` inside sentences (`lib/util/connectivity.dart:52`, `lib/provider/auth_provider.dart:94`, `lib/models/trail.dart:111,116`, `lib/routes/profile_screen.dart:587`, and `lib/components/base/trail_map.dart:127`) — none is a `TrailMap`/`NavigationScreen` code argument. The `trail_map.dart:127` hit is the doc comment the plan's own action step required ("record why the parameter was deleted... `TrailMap(offline: trail.isOffline)` conflated..."), so a literal zero here would conflict with that instruction. The plan's more precise Task 2 check — `grep -rn 'offline:' ... | grep -c 'TrailMap'` across the three call-site files — returns 0 as required. Treated as an imprecise sanity grep in the plan text, not a defect; no code change made.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plan 07 can proceed with the identical deletion in `NavigationScreen` (`isOffline`, same compose/listen fork shape) — confirmed no edit surface overlap, `navigation_screen.dart` untouched by this plan.
- Plan 08 can retire `rewriteStyleForOffline`/`offlineMapStyleJsonProvider` (D-17) once both Plan 06 and Plan 07 land — `trail_map.dart` no longer names `offlineMapStyleJsonProvider` at all (confirmed by grep), but the provider itself is still defined in `map_style_json_provider.dart` pending Plan 08's deletion, and `navigation_screen.dart` still calls it until Plan 07 lands.
- `flutter test` baseline: 1153 passed + 1 skip, 0 failures — unchanged from Plan 05's baseline, no regressions.
- `flutter analyze` reports the same 13 pre-existing info-level issues as Plan 05, all outside this plan's scope (`lib/entities/actor_entity.dart`, `lib/entities/category_entity.dart`, `lib/provider/navigation_stats_provider.dart`, `lib/store/local_photo_store.dart`, `lib/util/local/id.dart`, `vendor/tiptap_flutter/...`); 0 issues in every file this plan touched.

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-09*

## Self-Check: PASSED

All modified files (`app/lib/components/base/trail_map.dart`,
`app/lib/routes/trail_detail_map_screen.dart`,
`app/lib/components/trail/trail_panel.dart`,
`app/lib/routes/trail_create_screen.dart`) and both task commit hashes
(acd9a167, 5f52e76d) verified present.
