---
phase: 39-unified-tile-model
plan: 02
subsystem: infra
tags: [flutter, objectbox, riverpod, maplibre, tile-proxy, persistence]

# Dependency graph
requires: []
provides:
  - "LocalSettingsEntity.tileProxyPort / tileProxySecret / mapStyleSourcesJson / demTileTemplate persisted fields (additive ObjectBox migration)"
  - "tile_proxy_identity.dart: mintTileProxyIdentity, isValidProxySecret, resolveTileProxyIdentity, persistTileProxyPort"
  - "map_source_persistence.dart: readPersistedMapStyleSources, writePersistedMapStyleSources, readPersistedDemTileTemplate, writePersistedDemTileTemplate"
  - "MapStyleSourcesNotifier persists on fetch success and falls back to the persisted copy on fetch failure"
affects: [39-03, 39-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Pure-mint/validate half + store-backed read-or-mint-and-persist half, split the same way map_cache_path.dart splits validation from use"
    - "Small settings persist via a single-row LocalSettingsEntity read via box.getAll().firstOrNull ?? LocalSettingsEntity(), never raw files"
    - "Shared persistence module (map_source_persistence.dart) taking a raw Store so both a Riverpod provider and a pre-ProviderScope service (Plan 03's TileProxyServer.start) can read/write the same data"

key-files:
  created:
    - app/lib/services/tile_proxy_identity.dart
    - app/test/services/tile_proxy_identity_test.dart
    - app/lib/services/map_source_persistence.dart
  modified:
    - app/lib/entities/local_settings_entity.dart
    - app/lib/objectbox-model.json
    - app/lib/objectbox.g.dart
    - app/lib/provider/map_style_sources_provider.dart
    - app/lib/provider/map_style_sources_provider.g.dart

key-decisions:
  - "Port and secret are re-minted together, never independently, when either half of the persisted identity is invalid — a half-valid row is replaced wholesale (avoids a corrupted port paired with a stale-but-valid-looking secret or vice versa)."
  - "persistTileProxyPort only ever touches tileProxyPort — the secret must survive a bind-retry rebind unchanged so glyph/sprite cache keys don't churn for no reason."
  - "readPersistedMapStyleSources/readPersistedDemTileTemplate return null rather than throwing on any decode failure, since persisted data is never trusted to be well-formed."
  - "MapStyleSourcesNotifier rethrows the original network error when there is no persisted fallback available (first-ever offline run), rather than fabricating a value."

requirements-completed: [D-05, D-06, D-09]

# Metrics
duration: ~25min
completed: 2026-09-09
---

# Phase 39 Plan 02: Tile Proxy Identity + Style-Sources Persistence Summary

**Stable random loopback port + per-install secret minting/persistence (D-05/D-06), plus a disk-backed fallback for `/map/style-sources` (D-09), built on a new `LocalSettingsEntity` migration.**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-09-09T20:00:00Z (approx.)
- **Completed:** 2026-09-09T20:09:00Z
- **Tasks:** 3
- **Files modified:** 8 (3 created, 5 modified)

## Accomplishments
- `LocalSettingsEntity` gained four defaulted fields (`tileProxyPort`, `tileProxySecret`, `mapStyleSourcesJson`, `demTileTemplate`) via an additive ObjectBox migration — no property renamed or dropped, verified via schema diff.
- `tile_proxy_identity.dart` mints a stable random port (IANA dynamic range, `Random.secure()` by default) and a 128-bit per-install secret, with a store-backed resolve/persist cycle that returns a valid persisted identity unchanged and re-mints wholesale otherwise.
- `map_source_persistence.dart` gives both the Riverpod side and Plan 03's raw-`Store` proxy a shared, decode-failure-safe read/write surface for the operator's upstream templates.
- `MapStyleSourcesNotifier` now writes through to disk on a successful `/map/style-sources` fetch and falls back to the persisted copy on failure, rethrowing only when no persisted copy exists.

## Task Commits

Each task was committed atomically:

1. **Task 1: Persisted fields on LocalSettingsEntity** - `fc029a70` (feat)
2. **Task 2: tile_proxy_identity.dart — mint, validate, resolve, persist** - `30e7c86e` (feat)
3. **Task 3: Persist /map/style-sources and fall back to it offline** - `3af49398` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified
- `app/lib/entities/local_settings_entity.dart` - four new defaulted fields with D-05/D-06/D-09 doc comments
- `app/lib/objectbox-model.json` / `app/lib/objectbox.g.dart` - regenerated additive ObjectBox model
- `app/lib/services/tile_proxy_identity.dart` - `TileProxyIdentity`, `mintTileProxyIdentity`, `isValidProxySecret`, `resolveTileProxyIdentity`, `persistTileProxyPort`, `kTileProxyPortMin`/`kTileProxyPortMax`
- `app/test/services/tile_proxy_identity_test.dart` - 11 unit tests covering the pure mint/validate half
- `app/lib/services/map_source_persistence.dart` - `readPersistedMapStyleSources`, `writePersistedMapStyleSources`, `readPersistedDemTileTemplate`, `writePersistedDemTileTemplate`
- `app/lib/provider/map_style_sources_provider.dart` / `.g.dart` - `MapStyleSourcesNotifier` write-through on success, read-fallback on failure

## Decisions Made
- Port/secret re-minting is atomic (never independent) to avoid a half-valid persisted row.
- `persistTileProxyPort` never touches the secret, so Plan 03's bind-retry path can't accidentally churn glyph/sprite cache keys.
- Persistence reads are exception-safe (`try/catch` returning `null`) since ObjectBox-stored JSON/strings are never assumed well-formed.
- `MapStyleSourcesNotifier` rethrows on a fetch failure with no persisted fallback, preserving the existing error-surfacing contract for a genuinely offline first run.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- The first `dart run build_runner build --delete-conflicting-outputs` run regenerated hash comments in two unrelated generated files (`api_provider.g.dart`, `auth_provider.g.dart`) not listed in this plan's `files_modified`. These were out of scope and reverted with `git checkout --` before committing, keeping each commit scoped to its task's files.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Plan 03 can now call `resolveTileProxyIdentity`/`persistTileProxyPort` from `TileProxyServer.start`'s bind logic and read the persisted upstream templates via `map_source_persistence.dart` to build redirect targets — both new modules take a raw `Store`, matching the pre-`ProviderScope` constraint Plan 03 operates under.
- `demTileTemplate` field and its persistence functions exist but are not yet written by anything; Plan 03 is expected to populate them.
- No render-path behavior changed in this plan (by design) — nothing to verify visually yet.

---
*Phase: 39-unified-tile-model*
*Completed: 2026-09-09*

## Self-Check: PASSED

All created/modified files and all four commit hashes (fc029a70, 30e7c86e, 3af49398, ff2b84e8) verified present.
