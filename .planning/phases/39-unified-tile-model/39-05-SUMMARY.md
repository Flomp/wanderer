---
phase: 39-unified-tile-model
plan: 05
subsystem: infra
tags: [flutter, maplibre, tile-proxy, glyphs, sprites, style-composition]

# Dependency graph
requires:
  - phase: 39-unified-tile-model (plan 04)
    provides: "the /<secret>/glyphs/{fontstack}/{range}.pbf and /<secret>/sprite/<fileName> proxy routes this plan's rewritten transform must target"
  - phase: 39-unified-tile-model (plan 03)
    provides: "TileProxyServer's secret-prefixed baseUrl shape (http://127.0.0.1:$port/$secret) that proxyBaseUrl carries into every emitted URL"
provides:
  - "rewriteStyleForProxy(style, {required proxyBaseUrl, dark}) — the single unconditional style transform (D-01), cache-root-free (D-11), emitting only http://127.0.0.1: URLs for tiles, glyphs and sprite"
  - "_proxyVectorMaxZoom (14) / _proxyDemMaxZoom (12) — renamed maxzoom pins, now documented as applying to the online style too (D-08)"
  - "TrailMap and NavigationScreen composing their style from base JSON alone, with the map-open glyph/sprite cache warm retired (D-11)"
affects: [39-06, 39-07, 39-08]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Cache-root-free style transform: rewriteStyleForProxy no longer takes or validates a filesystem path — every URL-bearing field is rewritten to a loopback proxy URL, so path-traversal safety no longer needs to be reasoned about inside the style layer at all"

key-files:
  created: []
  modified:
    - app/lib/util/region/offline_style_rewriter.dart
    - app/lib/components/base/trail_map.dart
    - app/lib/routes/navigation_screen.dart
    - app/test/util/region/offline_style_rewriter_test.dart
    - app/test/provider/offline_map_style_json_test.dart
    - app/test/services/tile_proxy_spike_harness.dart

key-decisions:
  - "Renamed _offlinePmtilesMaxZoom/_offlineDemMaxZoom to _proxyVectorMaxZoom/_proxyDemMaxZoom as a true rename (single definition site), not an alias pair, and updated the legacy _pointSourceAtCell/_pointDemSourceAtCell call sites plus every doc-comment cross-reference to match — the plan's acceptance grep for the old names required zero occurrences anywhere in the file, including comments."
  - "Left rewriteStyleForOffline's own N-cell cellPaths/cacheRoot machinery completely untouched — only rewriteStyleForProxy's signature and body changed. rewriteStyleForOffline is Plan 08's deletion target (D-17), not this plan's."
  - "trail_map.dart's error resolution collapsed to `final error = baseAsync.error;` (no longer `Object? error` reassigned from a cache-path AsyncValue) since there is no second async input left to merge once the cache-paths provider dependency is gone."

requirements-completed: [D-01, D-08, D-11]

# Metrics
duration: ~20min
completed: 2026-09-09
---

# Phase 39 Plan 05: Cache-Root-Free Unified Style Transform Summary

**`rewriteStyleForProxy` now routes every URL-bearing style field — vector/DEM tiles, glyphs, and sprite alike — through the loopback tile proxy with no cache root and no `file://` URL, and both map hosts (`TrailMap`, `NavigationScreen`) call it with the new signature and no longer kick a glyph/sprite warm on map open.**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-09-09T21:20:00Z
- **Completed:** 2026-09-09T21:42:06Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- `rewriteStyleForProxy`'s signature dropped `cacheRoot` entirely (and its `_assertSafePath` call); `glyphs` now resolves to `<proxyBaseUrl>/glyphs/{fontstack}/{range}.pbf` and `sprite` to `<proxyBaseUrl>/sprite/<light|dark>` — matching exactly the four route families Plan 04 added to `TileProxyServer`. No `file://` URL is emitted anywhere in the function body (verified by a sed-scoped grep).
- `_offlinePmtilesMaxZoom`/`_offlineDemMaxZoom` were renamed in place to `_proxyVectorMaxZoom`/`_proxyDemMaxZoom` (values unchanged: 14/12), with doc comments extended to state D-08's online-included rationale; the legacy `_pointSourceAtCell`/`_pointDemSourceAtCell` helpers (still used by `rewriteStyleForOffline`) were updated to reference the renamed constants so the file keeps analyzing clean pending their Plan 08 retirement.
- `rewriteStyleForProxy`'s doc comment was rewritten to describe it as the single unconditional transform applied to every composed style (D-01), with coverage decisions (redirect-on-miss for tiles, local-first-with-write-through for glyphs/sprite) explicitly attributed to `tile_proxy_server.dart` rather than the style.
- `TrailMap` and `NavigationScreen` both dropped their `_cacheWarmed` field and the map-open `ref.read(glyphSpriteCacheProvider.future).ignore()` warm block (D-11) — the proxy's write-through now populates the same `map_cache` directory on first request. Both `_composeStyle` methods now take only `String? baseJson` and call the cache-root-free `rewriteStyleForProxy`; both dropped their `offlineGlyphSpritePathsProvider` watch/listen and unused `GlyphSpriteCachePaths`/`glyph_sprite_cache_provider.dart` imports.
- `NavigationScreen`'s identity-keyed compose memo collapsed from a two-field guard (`_composeBaseInput` + `_composeCacheInput`) to a single `identical(baseJson, _composeBaseInput)` check — the heavy decode/rewrite/encode round-trip it protects is unchanged in cost, just no longer needing a second input to key on.
- Both test-tree callers (`offline_map_style_json_test.dart`, `tile_proxy_spike_harness.dart`) were updated to the new call signature, with `file://` assertions replaced by `http://127.0.0.1:` ones and the harness's now-unneeded `glyphSpriteCacheProvider` warm removed along with its import.
- `trail_download_state_provider.dart` (the download-time glyph/sprite warm, which is *not* retired by D-11) was confirmed untouched via `git diff --stat` returning empty.

## Task Commits

Each task was committed atomically:

1. **Task 1: rewriteStyleForProxy emits only loopback URLs** - `dc87089b` (feat)
2. **Task 2: TrailMap composes without a cache root and without the map-open warm** - `87c08d28` (feat)
3. **Task 3: NavigationScreen call site and the two spike/provider tests** - `27bfcc5f` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `app/lib/util/region/offline_style_rewriter.dart` — `rewriteStyleForProxy` cache-root-free, `_proxyVectorMaxZoom`/`_proxyDemMaxZoom` rename, doc comment rewrite
- `app/test/util/region/offline_style_rewriter_test.dart` — `rewriteStyleForProxy` group updated: no `cacheRoot`, `file://` assertions replaced with `http://127.0.0.1:` ones, `{fontstack}`/`{range}` token-survival assertion added, cacheRoot-traversal case deleted
- `app/lib/components/base/trail_map.dart` — `_cacheWarmed` field and map-open warm deleted, `_composeStyle(String? baseJson)`, `offlineGlyphSpritePathsProvider` dependency dropped, unused imports removed
- `app/lib/routes/navigation_screen.dart` — same shape as `trail_map.dart`, adapted to the `_composeBaseInput`/`_composeOutput` memo (dropped `_composeCacheInput`)
- `app/test/provider/offline_map_style_json_test.dart` — dropped `cacheRoot:`, `file://` → `http://127.0.0.1:` assertions
- `app/test/services/tile_proxy_spike_harness.dart` — dropped `cacheRoot:` and the now-unneeded `glyphSpriteCacheProvider` warm/import

## Decisions Made

See `key-decisions` in the frontmatter. In summary: the maxzoom constants were a true rename (not an alias pair) so the plan's zero-occurrence grep for the old names passed including doc comments; `rewriteStyleForOffline`'s own N-cell machinery was left completely alone as Plan 08's scope; `trail_map.dart`'s local `error` variable collapsed to a plain `final` since there's only one async input left to merge.

## Deviations from Plan

None — plan executed exactly as written. All acceptance-criterion greps/seds pass as specified in the plan (verified below), and no Rule 1-4 deviation was needed.

## Issues Encountered

None. The first edit pass to the maxzoom constants used an alias approach (`const int _proxyVectorMaxZoom = _offlinePmtilesMaxZoom;`) that left the old names' doc-comment references and declarations in place; corrected in the same task to a full rename before verification, so no separate deviation was needed — this was caught by the task's own acceptance-criterion grep before committing.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plans 06 and 07 can now delete `TrailMap.offline`/`NavigationScreen.isOffline` and their remaining compose/listen forks (D-16) — every intermediate commit in this plan stayed analyzer-clean and behaviourally self-consistent, per the plan's own design.
- Plan 08 can retire `rewriteStyleForOffline` (D-17) — it now has no styling behavior in common with `rewriteStyleForProxy` beyond the renamed maxzoom constants it still reads, and no production caller.
- `flutter test` baseline: 1154 → 1154 (1153 passed + 1 pre-existing skip), 0 failures, no regressions.
- `flutter analyze` reports 0 issues in every file this plan touched; the 13 remaining info/warning-level issues package-wide are pre-existing and outside this plan's scope (`lib/entities/actor_entity.dart`, `lib/entities/category_entity.dart`, `lib/provider/navigation_stats_provider.dart`, `lib/store/local_photo_store.dart`, `lib/util/local/id.dart`, `vendor/tiptap_flutter/...`).

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-09*

## Self-Check: PASSED

All modified files (`app/lib/util/region/offline_style_rewriter.dart`,
`app/lib/components/base/trail_map.dart`, `app/lib/routes/navigation_screen.dart`,
`app/test/util/region/offline_style_rewriter_test.dart`,
`app/test/provider/offline_map_style_json_test.dart`,
`app/test/services/tile_proxy_spike_harness.dart`) and all three task commit
hashes (dc87089b, 87c08d28, 27bfcc5f) verified present.
