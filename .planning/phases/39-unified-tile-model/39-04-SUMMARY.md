---
phase: 39-unified-tile-model
plan: 04
subsystem: infra
tags: [flutter, tile-proxy, http, glyphs, sprites, cache, write-through, security]

# Dependency graph
requires:
  - phase: 39-unified-tile-model (plan 03)
    provides: "TileProxyServer's secret-prefixed routing, isSafeRedirectTarget, _redirectOrUnavailable, and the 302/503 miss path this plan extends"
provides:
  - "allowedSpriteFileNames / isAllowedSpriteFileName / spriteCacheFilePath in map_cache_path.dart — an 8-entry sprite filename whitelist matching the existing fontstack whitelist's validate-then-join discipline"
  - "/<secret>/glyphs/{fontstack}/{range}.pbf and /<secret>/sprite/<fileName> proxy routes — local-first serving from map_cache, reverse-proxied with write-through on a miss"
  - "_serveCachedAsset / _fetchAndCacheAsset / _downloadAndWriteThrough — the shared local-first-then-reverse-proxy-with-write-through primitive, in-flight-deduplicated per cache path"
affects: [39-05, 39-06, 39-07, 39-08]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Local-first-then-reverse-proxy-with-write-through (_serveCachedAsset): read map_cache first, fall back to a deduplicated upstream fetch that writes through atomically (.part + rename) before serving — the one place in this proxy that reverse-proxies bytes, sanctioned only for glyphs/sprites (D-11) because tiles must stay redirect-only (D-02)"
    - "Whitelist-then-join path safety (spriteCacheFilePath), mirroring glyphCacheFilePath exactly: membership in a fixed list makes traversal structurally unrepresentable before any p.join runs"

key-files:
  created: []
  modified:
    - app/lib/util/region/map_cache_path.dart
    - app/lib/services/tile_proxy_server.dart
    - app/test/util/region/map_cache_path_test.dart

key-decisions:
  - "Reordered _handle so the literal comparison 'kind == \\'vector\\'' appears before 'kind == \\'glyphs\\'' in the source text (tile branch checked first, glyphs/sprite dispatch second) — required for the plan's own sed-range acceptance check (kind == 'vector' .. kind == 'glyphs' contains no _serveCachedAsset/HttpClient) to be able to isolate the tile-serving block at all; also reads naturally as hot-path-first."
  - "Range stripped via the same 'segments[N].split('.').first' idiom the existing y-coordinate parsing already uses, rather than an explicit '.pbf' suffix check — glyphCacheFilePath's ^\\d+-\\d+$ regex is the real validation gate, so a second ad hoc extension check would be redundant."
  - "_serveCachedAsset dedupes the fetch (not the disk write) via _inFlightAssetFetches keyed by localPath: every concurrent caller for the same asset shares one Future<List<int>?>, so N simultaneous requests for one glyph range or sprite file produce exactly one upstream HttpClient call and one write-through, with every caller then serving from the same in-memory bytes rather than re-reading the file."
  - "BytesBuilder is imported from dart:typed_data directly (not the deprecated dart:io re-export) to keep flutter analyze clean, per the same discipline the file already follows for HttpClient/debugPrint wording."

requirements-completed: [D-11, D-07]

# Metrics
duration: ~15min
completed: 2026-09-09
---

# Phase 39 Plan 04: Glyph/Sprite Proxy Routes Summary

**The loopback tile proxy now serves `/glyphs/{fontstack}/{range}.pbf` and `/sprite/<fileName>` local-first from `map_cache`, reverse-proxying and write-through-caching a miss instead of ever leaving a first-run-offline map without labels or icons.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-09-09T21:20:48Z
- **Completed:** 2026-09-09T21:30:16Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- `map_cache_path.dart` gained `allowedSpriteFileNames` (the 8 filenames MapLibre's sprite loader can request — `light`/`dark` crossed with `.json`/`.png`/`@2x.json`/`@2x.png`), `isAllowedSpriteFileName`, and `spriteCacheFilePath`, all following the module's existing "validate against a known-good set before any `p.join`" discipline — membership in a fixed 8-element list makes `..`, `/`, a percent-encoded separator, or an absolute path structurally unrepresentable, matching the guarantee `allowedFontstacks` already gives the glyph path.
- `TileProxyServer.start` now resolves the shared `map_cache` root once via `resolveGlyphSpriteCachePaths()` and stores it as `_cacheRoot`, so the proxy, the (soon-to-be-retired) cache warm, and the offline render path all agree on the same on-disk layout. `main.dart`'s `start(Store store)` call site is unchanged (verified via `git diff --stat`).
- `_handle` gained two new route branches: `kind == 'glyphs'` (4 segments: fontstack + `<range>.pbf`) and `kind == 'sprite'` (3 segments: filename). Both validate their path via the whitelisted builders inside a `try`/`on ArgumentError` — a non-whitelisted fontstack/range/filename answers 404 (genuinely permanent, no filesystem touched), everything else reaches `_serveCachedAsset`.
- `_serveCachedAsset` implements local-first-then-reverse-proxy-with-write-through: serve straight from `map_cache` if the file exists; otherwise validate the upstream URL via the same `isSafeRedirectTarget` Plan 03 introduced, fetch it (deduplicated per cache path via `_inFlightAssetFetches` so N concurrent requests for one asset produce exactly one upstream `HttpClient` call), write the bytes through atomically (`<path>.part` then `File.rename`), and serve them. Any miss or failure answers 503 — never 404 — per D-04, which this plan's constraints extend explicitly to glyphs and sprites.
- The tile path (`kind == 'vector' || kind == 'dem'`) is structurally unchanged and still contains zero `HttpClient`/`_serveCachedAsset` calls — D-02's redirect-only invariant for tiles is preserved and verified by a grep over the isolated tile-serving block.
- The upstream glyph URL is built with the exact `{fontstack}`-encoded, `{range}`-literal substitution `glyph_sprite_cache_provider.dart` already uses; the upstream sprite URL is `<spriteUrl>/<fileName>` with no separately-appended variant, avoiding the "theme-agnostic base + already-variant-carrying filename" double-append trap `map_style_json_provider.dart` documents.

## Task Commits

Each task was committed atomically:

1. **Task 1: Sprite filename whitelist in map_cache_path.dart** - `95dcf262` (feat)
2. **Task 2: Glyph and sprite routes with local-first serving and write-through** - `d449c5cb` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `app/lib/util/region/map_cache_path.dart` — `allowedSpriteFileNames`, `isAllowedSpriteFileName`, `spriteCacheFilePath`
- `app/test/util/region/map_cache_path_test.dart` — accept/reject coverage for the 8 whitelisted sprite filenames, plus a length-8 assertion
- `app/lib/services/tile_proxy_server.dart` — `_cacheRoot` (resolved in `start`), `_inFlightAssetFetches`, `_handleGlyphRequest`, `_handleSpriteRequest`, `_serveCachedAsset`, `_fetchAndCacheAsset`, `_downloadAndWriteThrough`; class and `_handle` doc comments updated to describe all four route families and the D-02-vs-D-11 split

## Decisions Made

See `key-decisions` in the frontmatter. In summary: reordered the tile-vs-glyph/sprite branches in `_handle` so the file's literal text satisfies the plan's own sed-range verification (`kind == 'vector'` precedes `kind == 'glyphs'`); reused the existing `.split('.').first` extension-stripping idiom for the glyph range rather than adding a redundant `.pbf` suffix check; deduplication is per-fetch (shared `Future`) rather than per-disk-write; `BytesBuilder` imported directly from `dart:typed_data` to avoid a deprecation info from the indirect `dart:io` re-export.

## Deviations from Plan

None — plan executed exactly as written. All acceptance-criterion greps/seds pass as specified in the plan (verified below), and no Rule 1-4 deviation was needed.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plans 05-08 (`trail_map.dart`, `navigation_screen.dart`, the style providers) remain gated behind Plan 01's now-closed risk gate; this plan touched neither surface, staying inside its own scope fence.
- Plan 05/07 can now delete the explicit cache warm at `trail_map.dart:104-107` — the proxy's write-through on first glyph/sprite request makes that warm redundant, per D-11's own rationale.
- `flutter test` baseline: 1151 → 1154 (3 new `spriteCacheFilePath`/`allowedSpriteFileNames` tests), 0 failures, no regressions.

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-09*

## Self-Check: PASSED

All modified files (`app/lib/util/region/map_cache_path.dart`,
`app/lib/services/tile_proxy_server.dart`,
`app/test/util/region/map_cache_path_test.dart`) and both task commit
hashes (95dcf262, d449c5cb) verified present.
