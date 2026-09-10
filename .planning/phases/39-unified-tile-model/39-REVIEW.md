---
phase: 39-unified-tile-model
reviewed: 2026-09-10T00:00:00Z
depth: standard
files_reviewed: 28
files_reviewed_list:
  - app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt
  - app/lib/actions/launch_navigation.dart
  - app/lib/components/base/trail_map.dart
  - app/lib/components/trail/trail_panel.dart
  - app/lib/entities/active_navigation_entity.dart
  - app/lib/entities/local_settings_entity.dart
  - app/lib/main.dart
  - app/lib/provider/glyph_sprite_cache_provider.dart
  - app/lib/provider/map_style_json_provider.dart
  - app/lib/provider/map_style_sources_provider.dart
  - app/lib/provider/region/tile_proxy_provider.dart
  - app/lib/provider/router_provider.dart
  - app/lib/routes/navigation_screen.dart
  - app/lib/routes/trail_create_screen.dart
  - app/lib/routes/trail_detail_map_screen.dart
  - app/lib/routes/trail_source_select_screen.dart
  - app/lib/services/map_source_persistence.dart
  - app/lib/services/tile_proxy_identity.dart
  - app/lib/services/tile_proxy_server.dart
  - app/lib/services/tile_repository_manager.dart
  - app/lib/util/region/map_cache_path.dart
  - app/lib/util/region/proxy_style_rewriter.dart
  - app/test/provider/map_style_json_test.dart
  - app/test/services/tile_proxy_identity_test.dart
  - app/test/services/tile_proxy_redirect_test.dart
  - app/test/services/tile_proxy_spike_harness.dart
  - app/test/util/region/map_cache_path_test.dart
  - app/test/util/region/proxy_style_rewriter_test.dart
findings:
  critical: 1
  warning: 4
  info: 0
  total: 5
status: resolved
resolved: 5
resolved_in: 39-unified-tile-model (post-review fix pass)
---

# Phase 39: Unified Tile Model — Code Review Report

**Reviewed:** 2026-09-10
**Depth:** standard
**Files Reviewed:** 28
**Status:** issues_found

## Summary

This phase collapses the app's binary offline/online tile model into one always-proxied path,
centered on `tile_proxy_server.dart`. The D-04 "never 404 a retryable resource" invariant was
traced through every response path in `_handle`, `_handleGlyphRequest`, `_handleSpriteRequest`,
`_serveCachedAsset` and `_redirectOrUnavailable` and holds: every retryable miss (uncovered
tile, missing package, vanished archive file, tile absent from a covering archive, glyph/sprite
cache miss with no or unsafe upstream, download failure) answers 302 or 503, never 404/204. The
D-07 loopback-only invariant and the path-traversal defenses in `map_cache_path.dart` also hold —
every on-disk path is built from a small fixed whitelist or a DB-derived, previously-validated
`localFilePath`, never from request-path segments. The `TileMap`/`NavigationScreen` D-16 deletion
of `offline`/`isOffline` (and D-16a's `ActiveNavigationEntity.isOffline` column) is complete and
consistent across every call site checked.

The one finding that must be fixed before this ships is a genuine, easily-reproduced race in
`_ArchiveCache.forPath`: concurrent tile requests for a not-yet-cached archive path each open
their own `PmTilesArchive` handle, and the loser is silently discarded without ever being closed
— an unbounded file-handle leak that also undermines the class's own documented 8-entry cap. The
remaining findings are robustness gaps in the proxy's startup/serve-loop error handling and one
narrow gap in the redirect-target safety check.

## Critical Issues

### CR-01: `_ArchiveCache.forPath` race leaks pmtiles file handles and can close a handle still in use

**File:** `app/lib/services/tile_proxy_server.dart:740-762`

**Issue:** `forPath` is not reentrant-safe for a path that is not yet cached:

```dart
Future<PmTilesArchive?> forPath(String path) async {
  if (!File(path).existsSync()) { ... }
  final cached = _open[path];
  if (cached != null) return cached;
  if (_open.length >= _capacity) {
    final oldestPath = _open.keys.first;
    await _evict(oldestPath);
  }
  final archive = await PmTilesArchive.fromFile(File(path));  // await point
  _open[path] = archive;                                       // last writer wins
  return archive;
}
```

`PmTilesArchive.fromFile` is awaited, which yields to the event loop. `TileProxyServer._handle`
is dispatched per-request via `unawaited(_handle(request)...)` inside `_serve`'s `await for`
loop, so nothing serializes calls to `forPath`. Whenever two or more tile requests for the same
not-yet-cached archive land close together — the normal case: MapLibre requests a burst of
vector/DEM tiles across a viewport the moment a region newly resolves, or a fast pan re-enters a
region that just aged out of the 8-entry cache — both calls see `cached == null`, both open their
own file handle, and both eventually execute `_open[path] = archive`. The handle from the losing
call is never referenced again and therefore never closed by `_evict`/`closeAll` — it leaks for
the life of the process. Repeated across a session this can grow well past the documented
8-handle cap and, over enough distinct-region panning, risks exhausting the process's file
descriptor limit.

The same unguarded window also lets `_evict` (triggered by the capacity check in a *different*
concurrent call) close an archive that another in-flight request already retrieved from `_open`
and is actively calling `.tile()` on, producing a spurious failure on that request (caught only
by `_serve`'s top-level 500 handler).

**Fix:** Dedupe concurrent opens per path the same way `_inFlightAssetFetches` already dedupes
concurrent asset fetches:

```dart
class _ArchiveCache {
  static const int _capacity = 8;
  final Map<String, PmTilesArchive> _open = {};
  final Map<String, Future<PmTilesArchive>> _opening = {};

  Future<PmTilesArchive?> forPath(String path) async {
    if (!File(path).existsSync()) {
      await _evict(path);
      return null;
    }
    final cached = _open[path];
    if (cached != null) return cached;

    final inFlight = _opening[path];
    if (inFlight != null) return inFlight;

    final future = _openAndCache(path);
    _opening[path] = future;
    try {
      return await future;
    } finally {
      _opening.remove(path);
    }
  }

  Future<PmTilesArchive> _openAndCache(String path) async {
    if (_open.length >= _capacity) {
      final oldestPath = _open.keys.first;
      await _evict(oldestPath);
    }
    final archive = await PmTilesArchive.fromFile(File(path));
    _open[path] = archive;
    return archive;
  }
}
```

## Warnings

### WR-01: `_ArchiveCache` is FIFO, not LRU as documented

**File:** `app/lib/services/tile_proxy_server.dart:725-757`

**Issue:** The class doc comment calls this a "Bounded LRU-style cache", but eviction picks
`_open.keys.first` — the earliest-*inserted* entry — and a cache hit (`final cached = _open[path]; if (cached != null) return cached;`) never re-inserts/promotes the entry. Dart's default `Map` preserves insertion order, so this is a FIFO cache: an archive that is hit on every request (e.g. the region the user is currently standing in) will still be evicted once 8 *other* distinct paths have been opened after it, even though it is by far the most active.

**Fix:** Either rename the doc comment to describe FIFO behavior accurately, or make it actually LRU by re-inserting on hit (`_open.remove(path); _open[path] = cached;` before returning) so `keys.first` is genuinely the least-recently-used entry.

### WR-02: `isSafeRedirectTarget` does not reject `0.0.0.0` as an unsafe redirect host

**File:** `app/lib/services/tile_proxy_server.dart:664-675`

**Issue:** The function's stated purpose is to stop "a tampered or stale persisted template from turning the proxy into a relay to another local listener" (T-39-11), and it explicitly rejects `localhost` and any `isLoopback`/`isLinkLocal` literal. `InternetAddress.tryParse('0.0.0.0').isLoopback` is `false`, so a template host of `0.0.0.0` passes this check — yet on several platform network stacks (notably Linux, which Android's kernel is derived from) a socket connect to `0.0.0.0` is treated equivalently to `127.0.0.1`. This is the same class of target the function otherwise defends against.

**Fix:** Add an explicit check for the unspecified address alongside the existing loopback/link-local checks:

```dart
if (address != null &&
    (address.isLoopback || address.isLinkLocal || address.isAnyLocal)) {
  return false;
}
```

### WR-03: `TileProxyServer._serve()`'s request loop has no protection against a stream-level failure

**File:** `app/lib/services/tile_proxy_server.dart:194-203`

**Issue:** Every *per-request* failure is caught (`_handle(request).catchError(...)` on line 197), but the `await for (final request in _server)` loop itself is not wrapped in try/catch. If the underlying `HttpServer` stream ever emits an error event (rather than a per-connection failure, which `dart:io` normally isolates) — for example from a severely malformed request at the socket level — that error propagates out of `_serve()`, the `await for` loop exits, and the server stops accepting any further connections. Because `_serve()` is started via `unawaited(proxy._serve())` in `start()` with no supervising restart, this would silently and permanently kill map tile/glyph/sprite serving for the remainder of the app session; the class's own doc comment promises the server "keeps the server alive for the process lifetime," which this doesn't actually guarantee.

**Fix:** Wrap the loop body/iteration in a way that survives a stream-level error, e.g. restart the listen loop on error instead of letting the `Future` complete with an exception:

```dart
Future<void> _serve() async {
  while (true) {
    try {
      await for (final request in _server) {
        unawaited(_handle(request).catchError((Object _) {
          request.response.statusCode = HttpStatus.internalServerError;
          return request.response.close();
        }));
      }
      return; // server was closed deliberately (stop())
    } catch (e) {
      debugPrint('TileProxyServer: serve loop error, restarting — $e');
    }
  }
}
```

### WR-04: Bind-retry loop only handles `SocketException`; any other bind failure crashes startup

**File:** `app/lib/services/tile_proxy_server.dart:161-192`

**Issue:** `TileProxyServer.start` retries the persisted port up to 4 times and explicitly documents a last-resort fallback to an OS-assigned port (`HttpServer.bind(InternetAddress.loopbackIPv4, 0)`) "so the app can still start" — but the retry loop's `catch` clause only matches `on SocketException`. Any other exception thrown by `HttpServer.bind` (a `FileSystemException`, a platform-specific `OSError` subtype, etc.) is not caught, propagates out of `start()`, and since `main.dart` does `final proxyServer = await TileProxyServer.start(store);` before `runApp`, an uncaught exception here prevents the app from launching at all — exactly the outcome the OS-assigned-port fallback exists to avoid.

**Fix:** Broaden the catch to the documented intent, or add a second guard around the whole attempt loop:

```dart
} catch (e) {
  if (attempt == maxBindAttempts) break;
  port = mintTileProxyIdentity().port;
  persistTileProxyPort(store, port);
}
```

---

_Reviewed: 2026-09-10_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_


---

## Resolution

All five findings were fixed in a single post-review pass on
`app/lib/services/tile_proxy_server.dart`. `flutter analyze` clean; `flutter test`
1134 passed + 1 skip, unchanged, plus 3 new cases in `tile_proxy_redirect_test.dart`.

| Finding | Fix |
|---|---|
| **Critical** — fd leak in `_ArchiveCache.forPath` | Added an `_opening` in-flight memo keyed by path, mirroring the server's existing `_inFlightAssetFetches` pattern, so concurrent misses share one open instead of each opening a handle. Added a defensive never-overwrite-a-live-handle branch that closes the loser rather than orphaning it. Eviction moved to *after* a successful open, so a failed open never costs a live handle. |
| **W1** — cache documented LRU but behaved FIFO | `forPath` now removes and re-inserts on a hit, so `_open.keys.first` is genuinely least-recently-*used*. Doc comment corrected to state the invariant rather than hedge with "LRU-style". |
| **W2** — `isSafeRedirectTarget` accepted `0.0.0.0` | Explicit rejection of `InternetAddress.anyIPv4` and `anyIPv6`. Covered by 3 new tests (`0.0.0.0`, `0.0.0.0:8080`, `[::]`). |
| **W3** — `_serve()` died silently on a stream error | Switched from `await for` (which terminates on the first error, killing the proxy for the process lifetime) to `Stream.listen` with `cancelOnError: false` and an `onError` that logs and continues. `_serve()` is now `void`; its call site dropped `unawaited`. |
| **W4** — bind retry caught only `SocketException` | Widened to `catch (e)` with a diagnostic `debugPrint`, so any bind failure still reaches the documented OS-assigned-port fallback instead of escaping `start()` and crashing before `runApp`. |

### Residual, knowingly accepted

**Capacity eviction can still close a handle an in-flight request holds.** Fully
eliminating it needs refcounting on `PmTilesArchive`, which is disproportionate: eviction
only triggers past 8 distinct archives in one session, true-LRU ordering now makes the
victim the least-recently-used entry, and a lost race surfaces as the proxy's existing
HTTP 500 — `Reason::Server` in MapLibre, which retries with backoff. It degrades to a
retried tile, not a permanently blank one.

**The Critical fix has no direct unit test.** `_ArchiveCache` is private and exercising
the concurrent-open path needs a live `HttpServer` plus real `.pmtiles` archives. The
change is reasoned and analyzer-clean, but it is not proven by a regression test — worth
knowing if this code is touched again.

### Provenance note

`_ArchiveCache` was introduced in Phase 25.1 (`71643353`); Phase 39 did not change
`forPath`'s logic. The leak predates this phase. Phase 39 made it materially worse by
turning the proxy from an offline-only path into the single always-on tile path for every
map in the app, so the burst-on-region-entry case now happens constantly.
