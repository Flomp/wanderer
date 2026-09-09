---
phase: 39-unified-tile-model
plan: 07
subsystem: infra
tags: [flutter, maplibre, riverpod, objectbox, tile-proxy, style-composition, go-router]

# Dependency graph
requires:
  - phase: 39-unified-tile-model (plan 05)
    provides: "rewriteStyleForProxy as the cache-root-free unconditional style transform, and NavigationScreen's _composeStyle already narrowed to a single-argument signature calling it"
provides:
  - "NavigationScreen with no isOffline parameter, no compose branch, and no ref.listen fork — one style source (mapStyleJsonProvider), one listener, one compose path (twin of TrailMap, Plan 06)"
  - "A live cloud_off maneuver-banner indicator reading ref.watch(onlineStatusProvider) instead of a value frozen at screen construction"
  - "The /trail/:id/navigate extra record as a 3-tuple (NavigateResponse, ActiveNavigationEntity?, geo.Position?) at all three construction sites, and the /record extra map with no isOffline key"
  - "ActiveNavigationEntity with no isOffline column — property UID 5220757162905877938 retired in objectbox-model.json"
affects: [39-08]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Deletion over defaulting (twin of Plan 06): a caller-supplied mode flag whose flip could not reach the mounted native view was removed outright, along with every upstream value that only ever fed it — a route-extra tuple slot, a persisted ObjectBox column, and two reachability probes that gated a push instead of merely refreshing app-wide status."
    - "Retiring an ObjectBox property is routine generated output: dart run build_runner build --delete-conflicting-outputs moves the field's UID into retiredPropertyUids automatically; no hand-written migration needed on an unshipped schema."

key-files:
  created: []
  modified:
    - app/lib/routes/navigation_screen.dart
    - app/lib/entities/active_navigation_entity.dart
    - app/lib/objectbox-model.json
    - app/lib/objectbox.g.dart
    - app/lib/provider/router_provider.dart
    - app/lib/actions/launch_navigation.dart
    - app/lib/main.dart
    - app/lib/routes/trail_source_select_screen.dart

key-decisions:
  - "Followed the plan's amendment: deleted ActiveNavigationEntity.isOffline outright (field + constructor param) rather than retaining it, since nothing reads it back after router_provider.dart's /record builder stops consulting resume?.isOffline (D-16a)."
  - "Reverted two unrelated .g.dart files (auth_provider.g.dart, tile_proxy_provider.g.dart) that dart run build_runner build regenerated as a side effect — their source drift predates this plan (stale doc-comment/hash sync from earlier plans) and both files are outside this plan's files_modified scope, so they were excluded from the commit."
  - "Both main.dart resume paths keep their onlineStatusProvider.notifier.refresh() call (now unawaited-equivalent in intent, though _pushNavigationResume still awaits it before the push) — the probe settles app-wide online status for other consumers (watchdog, sync drain) even though it no longer selects a style path."
  - "trail_source_select_screen.dart's probe was converted from an awaited gate to a genuinely fire-and-forget unawaited(...) call, per the plan's action step, requiring a new dart:async import."

requirements-completed: [D-01, D-16]

# Metrics
duration: ~25min
completed: 2026-09-10
---

# Phase 39 Plan 07: Delete NavigationScreen.isOffline Summary

**`NavigationScreen` now composes and watches exactly one style path (`mapStyleJsonProvider` → `rewriteStyleForProxy`) with no `isOffline` parameter, its maneuver-banner offline icon now tracks live connectivity, both `/navigate` and `/record` route payloads dropped their mode-selecting slot, and the persisted `ActiveNavigationEntity.isOffline` ObjectBox column is retired (not merely unused).**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-09-09T23:45:00Z
- **Completed:** 2026-09-10T00:10:00Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Deleted `NavigationScreen.isOffline`'s field declaration and constructor default outright — the twin of Plan 06's `TrailMap.offline` deletion. Collapsed `build()`'s `if (widget.isOffline) {...} else {...}` `ref.listen` fork to a single unconditional `ref.listen(mapStyleJsonProvider, (_, _) => _swapStyle())`, and its `baseAsync` ternary to a plain `ref.watch(mapStyleJsonProvider)`.
- Removed `_composeStyle`'s `if (!widget.isOffline) return baseJson;` early return so the decode → `rewriteStyleForProxy` → encode path runs for every style, online and offline alike (D-01) — verified by a sed-scoped grep for `return baseJson;` inside the method returning 0 matches. Collapsed `_swapStyle`'s read-ternary to the single `ref.read(mapStyleJsonProvider).value`.
- Made the maneuver-banner `cloud_off` indicator live: `if (widget.isOffline)` became `if (!ref.watch(onlineStatusProvider))`, so the icon now appears/disappears mid-session instead of being frozen at screen construction. Added the `online_status_provider.dart` import this required.
- Dropped the `isOffline:` argument from `_persistNow`'s `ActiveNavigationEntity(...)` construction with no replacement — the field itself is being deleted, so there was nothing to feed.
- Deleted `ActiveNavigationEntity.isOffline` outright (the `bool? isOffline;` field and its `this.isOffline,` constructor parameter) per the plan's amendment (D-16a), then regenerated ObjectBox bindings via `dart run build_runner build --delete-conflicting-outputs`: property UID `5220757162905877938` moved from `ActiveNavigationEntity`'s `properties` array into `retiredPropertyUids` in `objectbox-model.json`, and `objectbox.g.dart` dropped every reference to the field.
- Shrank the `/trail/:id/navigate` route's `extra` record from a 4-tuple to `(NavigateResponse, ActiveNavigationEntity?, geo.Position?)` at all three construction sites — `router_provider.dart`'s `is!` guard and destructure, `launch_navigation.dart`'s two pushes (fresh launch and cached-fallback), and `main.dart`'s `_pushNavigationResume` — and dropped the now-absent `isOffline:` argument from the `NavigationScreen(...)` construction.
- Deleted the `/record` route's `isOffline` local (`extra is Map ? extra['isOffline'] : resume?.isOffline`) from `router_provider.dart`, and the `'isOffline': isOffline` key from `trail_source_select_screen.dart`'s `/record` push. Converted that screen's reachability probe from an awaited gate (`final isOffline = await offlineFuture;`) to a fire-and-forget `unawaited(ref.read(onlineStatusProvider.notifier).refresh())`, adding the `dart:async` import it needed.
- Kept the reachability probes in `main.dart`'s `_pushNavigationResume`/`_pushRecordingResume` (still settle `onlineStatusProvider` for other consumers) but stopped threading their result into a style-selecting flag or the persisted row; rewrote both methods' doc comments to state the persisted flag is no longer consulted at all.
- Rewrote every stale doc comment whose premise D-01 deleted — `_composeStyle`'s class comment, the `build()` comment block ("Offline reads the network-free providers so no `/map/style-sources` call is ever made"), `_swapStyle`'s comment, `launch_navigation.dart`'s tuple-slot-naming comment, and `trail_source_select_screen.dart`'s probe comment — replacing each with the true single-path-through-the-proxy statement.

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove isOffline from NavigationScreen and from the persisted entity** - `fb8806d4` (feat)
2. **Task 2: Remove the isOffline slot from both route entry points** - `230189be` (feat)
3. **Task 3: Stop computing the recorder's offline flag** - `04aa0819` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `app/lib/routes/navigation_screen.dart` — `isOffline` field/constructor param deleted; `build()`'s listen fork and `baseAsync` ternary collapsed to one path; `_composeStyle`'s early-return branch deleted; `_swapStyle`'s read-ternary collapsed; `_persistNow` drops the `isOffline:` argument; maneuver-banner icon reads `onlineStatusProvider` live; added `online_status_provider.dart` import; doc comments rewritten
- `app/lib/entities/active_navigation_entity.dart` — `bool? isOffline;` field and `this.isOffline,` constructor param deleted
- `app/lib/objectbox-model.json` — property UID `5220757162905877938` moved into `retiredPropertyUids`
- `app/lib/objectbox.g.dart` — regenerated; all `isOffline` references removed
- `app/lib/provider/router_provider.dart` — `/record` builder's `isOffline` local and its `NavigationScreen` argument deleted; `/navigate` builder's `is!` guard and destructure shrunk to a 3-tuple, `isOffline:` argument dropped
- `app/lib/actions/launch_navigation.dart` — both `/navigate` pushes shrunk to a 3-tuple `extra`; slot-naming comment rewritten
- `app/lib/main.dart` — `_pushNavigationResume`'s `extra` shrunk to a 3-tuple, its `isOffline` local dropped; `_pushRecordingResume` no longer writes `row.isOffline`; both doc comments rewritten to state the flag is no longer consulted
- `app/lib/routes/trail_source_select_screen.dart` — `/record` push drops the `isOffline` key; reachability probe converted from an awaited gate to `unawaited(...)`; added `dart:async` import; stale comment rewritten

## Decisions Made

See `key-decisions` in the frontmatter. In summary: followed the plan's amendment to delete the persisted `isOffline` column outright rather than retain it; reverted two unrelated `.g.dart` files that `build_runner` regenerated as a side effect of stale pre-existing doc-comment/hash drift, since they sit outside this plan's `files_modified` scope; kept both `main.dart` resume-path reachability probes (they still settle app-wide online status for the watchdog and sync drain) while removing their role in selecting a style path.

## Deviations from Plan

None beyond the pre-approved amendment stated in the plan header (D-16a, deleting the persisted column outright). No Rule 1-4 deviation was needed during execution.

One incidental side effect handled without a rule invocation: `dart run build_runner build --delete-conflicting-outputs` (required by Task 1) also regenerated `lib/provider/auth_provider.g.dart` and `lib/provider/region/tile_proxy_provider.g.dart` to sync them with doc-comment/logic changes already committed by earlier plans but never regenerated. Neither file is in this plan's `files_modified` list and neither change was caused by this plan's edits, so both were reverted via `git checkout --` before committing, per the scope-boundary rule (only auto-fix issues directly caused by the current task's changes).

## Issues Encountered

None. All acceptance-criterion greps/seds in the plan passed as specified on first attempt; `flutter analyze` was clean on every touched file with zero new issues; the three targeted test files (64 tests) and the full suite (1153 passed + 1 skip) both passed with no regressions.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plan 08 can retire `rewriteStyleForOffline`/`offlineMapStyleJsonProvider` (D-17) now that both Plan 06 (`TrailMap`) and this plan (`NavigationScreen`) have stopped naming `offlineMapStyleJsonProvider` entirely (confirmed by grep — 0 occurrences in `navigation_screen.dart`). The provider itself is still defined in `map_style_json_provider.dart` pending Plan 08's deletion.
- `flutter test` baseline: 1153 passed + 1 skip, 0 failures — unchanged from Plan 05/06's baseline, no regressions.
- `flutter analyze` reports the same 13 pre-existing info-level issues as Plans 05/06, all outside this plan's scope (`lib/entities/actor_entity.dart`, `lib/entities/category_entity.dart`, `lib/provider/navigation_stats_provider.dart`, `lib/store/local_photo_store.dart`, `lib/util/local/id.dart`, `vendor/tiptap_flutter/...`); 0 issues in every file this plan touched.
- `git diff --stat` confirms `app/lib/components/base/trail_map.dart` is untouched by this plan (Plan 06's file) and no file outside `app/` was modified.
- The `NavigateResponse`/`ActiveNavigationEntity?`/`geo.Position?` 3-tuple shape is now consistent across all three `/navigate` producers and the one consumer — no stale 4-tuple call site remains anywhere in `lib/`.

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-10*

## Self-Check: PASSED

All modified files (`app/lib/routes/navigation_screen.dart`,
`app/lib/entities/active_navigation_entity.dart`, `app/lib/objectbox-model.json`,
`app/lib/objectbox.g.dart`, `app/lib/provider/router_provider.dart`,
`app/lib/actions/launch_navigation.dart`, `app/lib/main.dart`,
`app/lib/routes/trail_source_select_screen.dart`) and all three task commit
hashes (fb8806d4, 230189be, 04aa0819) verified present.
