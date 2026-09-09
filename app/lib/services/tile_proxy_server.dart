import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart'
    show debugPrint, visibleForTesting;
import 'package:maplibre/maplibre.dart' show LngLatBounds;
import 'package:pmtiles/pmtiles.dart';
import 'package:wanderer/entities/region_entity.dart';
import 'package:wanderer/objectbox.g.dart';
import 'package:wanderer/services/tile_proxy_identity.dart';
import 'package:wanderer/services/tile_repository_manager.dart';
import 'package:wanderer/util/geo/xyz_tile_bounds.dart';

/// Loopback-only `HttpServer` that serves vector/DEM map tiles from a
/// downloaded region's `.pmtiles` archive, falling back to a redirect to the
/// operator's upstream CDN when no downloaded region covers a requested
/// tile.
///
/// Both `TrailMap` and `navigation_screen` bake a single STATIC
/// `tiles: ['<baseUrl>/vector/{z}/{x}/{y}.pbf']` /
/// `['<baseUrl>/dem/{z}/{x}/{y}.png']` XYZ source into every composed style
/// (see `offline_style_rewriter.dart`'s `rewriteStyleForProxy`) instead
/// of incrementally `addSource`/`removeSource`-ing per-region `pmtiles://`
/// archives. That incremental reconcile (an earlier
/// `_reconcileRegionComposition` in `navigation_screen.dart`) had no
/// reentrancy guard and desynced its tracking sets from the real native style
/// under overlapping camera-idle events. Routing every tile request
/// through this single static-source server structurally eliminates that bug
/// class: there is no reconcile call left to race, because MapLibre Native's
/// own viewport tracking decides which tiles to request, and every request
/// resolves its winning region fresh, per-request, via [resolveRegionForTile].
///
/// Binds `InternetAddress.loopbackIPv4` (never the wildcard-bind address) on
/// a port drawn once at install from the IANA dynamic range and persisted
/// (`tile_proxy_identity.dart`), because MapLibre's ambient cache keys on the
/// loopback URL and a port that changes every launch orphans the whole tile
/// cache on every cold start (RESEARCH.md 4.5). Unguessability is preserved
/// by the port being random rather than fixed, and by every route requiring
/// a per-install 128-bit secret path segment (`/$secret/vector/{z}/{x}/{y}.pbf`)
/// that a co-resident app on the same device cannot predict.
class TileProxyServer {
  final HttpServer _server;
  final Store _store;
  final String _secret;
  final _ArchiveCache _archiveCache = _ArchiveCache();

  /// Short-TTL memo of the region table. Every tile request used to run a
  /// full `RegionEntity` `getAll()` — during a pan that's an ObjectBox table
  /// scan per tile. Regions change on the order of user actions (download /
  /// delete), so a few seconds of staleness is imperceptible: a deleted
  /// region's file-vanished path already 404s via [_ArchiveCache.forPath],
  /// and a fresh download's tiles appear within the TTL.
  List<RegionEntity>? _regionCache;
  DateTime? _regionCacheAt;
  static const _regionCacheTtl = Duration(seconds: 5);

  List<RegionEntity> _regions() {
    final now = DateTime.now();
    final cachedAt = _regionCacheAt;
    final cached = _regionCache;
    if (cached != null &&
        cachedAt != null &&
        now.difference(cachedAt) < _regionCacheTtl) {
      return cached;
    }
    final fresh = _store.box<RegionEntity>().getAll();
    _regionCache = fresh;
    _regionCacheAt = now;
    return fresh;
  }

  TileProxyServer._(this._server, this._store, this._secret);

  /// The resolved loopback base URL, e.g.
  /// `http://127.0.0.1:54321/<32-hex-secret>` — exposed to the widget tree
  /// via the [tile_proxy_provider.dart] `keepAlive` provider, overridden in
  /// `main.dart` immediately after this starts. Carries the per-install
  /// secret (D-06) and must never be logged or surfaced in user-visible UI.
  String get baseUrl => 'http://127.0.0.1:${_server.port}/$_secret';

  /// Binds the loopback server and starts serving. Must be called after
  /// `openStore()` (the request handler queries [RegionEntity] rows) and
  /// before the first composed style is built, since the resolved [baseUrl]
  /// is baked into that style's tile source templates.
  ///
  /// Binds the persisted port (D-05) so the ambient tile cache survives a
  /// cold start. If that port is occupied by another process, mints and
  /// persists a fresh port (never re-minting the secret — D-06) and retries,
  /// up to 4 total bind attempts. As a last resort, if every attempt fails,
  /// binds an OS-assigned port (port `0`) so the app can still start; that
  /// only orphans the ambient tile cache for this one launch, since the
  /// still-persisted identity is retried on the next cold start.
  static Future<TileProxyServer> start(Store store) async {
    final identity = resolveTileProxyIdentity(store);
    final secret = identity.secret;
    var port = identity.port;

    HttpServer? server;
    const maxBindAttempts = 4;
    for (var attempt = 1; attempt <= maxBindAttempts; attempt++) {
      try {
        server = await HttpServer.bind(InternetAddress.loopbackIPv4, port);
        break;
      } on SocketException {
        if (attempt == maxBindAttempts) break;
        port = mintTileProxyIdentity().port;
        persistTileProxyPort(store, port);
      }
    }

    if (server == null) {
      debugPrint(
        'TileProxyServer: exhausted $maxBindAttempts bind attempts on the '
        'persisted port; falling back to an OS-assigned port. The ambient '
        'tile cache will be orphaned this launch only.',
      );
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    }

    final proxy = TileProxyServer._(server, store, secret);
    unawaited(proxy._serve());
    return proxy;
  }

  Future<void> _serve() async {
    await for (final request in _server) {
      unawaited(
        _handle(request).catchError((Object _) {
          request.response.statusCode = HttpStatus.internalServerError;
          return request.response.close();
        }),
      );
    }
  }

  /// Closes the server and every cached archive handle. Used by tests /
  /// defensive shutdown — production keeps the server alive for the process
  /// lifetime (matches the `keepAlive` ObjectBox [Store] precedent; no
  /// app-lifecycle-driven start/stop, since loopback binding has no
  /// battery/network cost while idle).
  Future<void> stop() async {
    await _archiveCache.closeAll();
    await _server.close(force: true);
  }

  /// Routes `/<secret>/vector/{z}/{x}/{y}.pbf` and
  /// `/<secret>/dem/{z}/{x}/{y}.png`.
  ///
  /// An empty path, a wrong/absent secret, an unknown route shape or kind,
  /// or an out-of-range z/x/y answer HTTP 404/400 — these are genuinely
  /// permanent: a request MapLibre would never legitimately repeat. A
  /// covered tile is served from the winning region's archive; an uncovered
  /// tile, a region with no package path, a vanished archive file and a tile
  /// absent from a covering archive all redirect (302) to the operator's
  /// upstream template; a request the proxy cannot yet name an upstream
  /// target for answers 503 (retryable). See [_redirectOrUnavailable].
  Future<void> _handle(HttpRequest request) async {
    final segments = request.uri.pathSegments;
    if (segments.isEmpty) {
      request.response.statusCode = HttpStatus.notFound;
      return request.response.close();
    }
    if (!_constantTimeEquals(segments[0], _secret)) {
      request.response.statusCode = HttpStatus.notFound;
      return request.response.close();
    }

    if (segments.length != 5 ||
        (segments[1] != 'vector' && segments[1] != 'dem')) {
      request.response.statusCode = HttpStatus.notFound;
      return request.response.close();
    }
    final kind = segments[1];

    final z = int.tryParse(segments[2]);
    final x = int.tryParse(segments[3]);
    final y = int.tryParse(segments[4].split('.').first);

    // Explicit bounds check BEFORE constructing a ZXY — never rely on ZXY's
    // constructor `assert`, which is stripped in release builds
    // — trusting that assert is an anti-pattern.
    if (z == null ||
        x == null ||
        y == null ||
        z < 0 ||
        z > 26 ||
        x < 0 ||
        x >= (1 << z) ||
        y < 0 ||
        y >= (1 << z)) {
      request.response.statusCode = HttpStatus.badRequest;
      return request.response.close();
    }

    // Computed once, reused by every retryable-miss branch below
    // (_redirectOrUnavailable): an uncovered tile, a region with no package
    // path, a vanished archive file, and a tile absent from a covering
    // archive all redirect to the same resolved upstream target.
    final target = _upstreamRedirectTargetFor(kind, z: z, x: x, y: y);

    final tileBounds = tileToBounds(z, x, y);
    // resolveRegionForTile is @visibleForTesting so its pure-function shape
    // stays unit-testable without a live Store (matches bboxOverlaps'/
    // splitRegionTilePaths' precedent) — this proxy handler is its one
    // sanctioned production caller, mirroring the pmtiles package's own
    // `fromReadAt` cross-file @visibleForTesting usage
    // (pmtiles-1.2.0/lib/src/archive.dart).
    // ignore: invalid_use_of_visible_for_testing_member
    final region = resolveRegionForTile(
      _regions(),
      LngLatBounds(
        longitudeWest: tileBounds.west,
        longitudeEast: tileBounds.east,
        latitudeSouth: tileBounds.south,
        latitudeNorth: tileBounds.north,
      ),
      dem: kind == 'dem',
    );

    if (region == null) {
      // No downloaded region covers this tile — redirect upstream (D-02,
      // D-04), never 404: this tile could still succeed via the CDN.
      return _redirectOrUnavailable(request, target);
    }

    // The archive path is read ONLY from the winning region's own
    // DownloadedTilePackageEntity.localFilePath — a DB-derived value already
    // validated at write time via util/region/file_path.dart's
    // assertValidRegionPath, NEVER assembled from the request path. This
    // structurally eliminates path traversal.
    final localFilePath = kind == 'dem'
        ? region.demPackage.target?.localFilePath
        : region.vectorPackage.target?.localFilePath;
    if (localFilePath == null) {
      return _redirectOrUnavailable(request, target);
    }

    final archive = await _archiveCache.forPath(localFilePath);
    if (archive == null) {
      // The winning region's file no longer exists on disk (mid-session
      // delete) — same retryable-miss outcome as no coverage.
      return _redirectOrUnavailable(request, target);
    }

    final tile = await archive.tile(ZXY(z, x, y).toTileId());
    List<int> bytes;
    try {
      bytes = tile.bytes();
    } on TileNotFoundException {
      // Present in the covering archive's index but absent from its data —
      // still retryable: the operator's upstream copy may have it.
      return _redirectOrUnavailable(request, target);
    }

    // Serve decompressed bytes with NO Content-Encoding header (safer
    // default than serving compressed bytes + Content-Encoding: gzip).
    request.response.headers.contentType = ContentType.parse(
      tile.type.mimeType(),
    );
    // Without a Cache-Control header MapLibre treats every offline tile as
    // immediately expired and re-runs the whole HTTP → region-resolve →
    // pmtiles read per pan revisit; with one, its ambient cache serves
    // revisits directly. One day balances that against a re-downloaded
    // region's updated tiles (same URLs) becoming visible.
    request.response.headers.set(
      HttpHeaders.cacheControlHeader,
      'public, max-age=86400',
    );
    request.response.add(bytes);
    return request.response.close();
  }

  /// Resolves this request's upstream redirect target from the persisted
  /// vector/DEM templates (D-09). Template resolution is wired in by Task 3
  /// — until then this always answers `null`, which correctly means "no
  /// upstream target known yet" and routes every retryable miss to 503.
  String? _upstreamRedirectTargetFor(
    String kind, {
    required int z,
    required int x,
    required int y,
  }) {
    return null;
  }

  /// Answers a retryable tile miss with either a 302 redirect to [target]
  /// (when the proxy has a validated upstream target) or a 503 (when it does
  /// not). This is the single shared answer for every retryable miss — an
  /// uncovered tile, a region with no package path, a vanished archive file,
  /// and a tile absent from a covering archive all reach this.
  ///
  /// The choice between 302 and 503 is deliberate and load-bearing: 503
  /// (`Reason::Server`) backs off 1s ×3 then exponentially and IS retried by
  /// MapLibre Native, whereas 404 (`Reason::NotFound`) backs off to
  /// `Duration::max()` and a 204 (`noContent`) is persisted as an empty tile
  /// row — both terminal. Never answer a tile that might succeed later with
  /// either (D-04).
  Future<void> _redirectOrUnavailable(HttpRequest request, String? target) {
    if (target != null) {
      request.response.statusCode = HttpStatus.found;
      request.response.headers.set(HttpHeaders.locationHeader, target);
      return request.response.close();
    }
    request.response.statusCode = HttpStatus.serviceUnavailable;
    return request.response.close();
  }
}

/// Whether [uri] is safe to use as a redirect `Location` target: an absolute
/// `http`/`https` URI with a non-empty host that is neither `localhost` nor
/// a loopback/link-local address.
///
/// Rejecting loopback and link-local targets is what prevents the proxy
/// from redirecting a request back into itself — an infinite redirect loop
/// that would exhaust OkHttp's 20-hop limit and burn the request — and is
/// the guard that keeps a tampered or stale persisted template from turning
/// the proxy into a relay to another local listener (T-39-11).
@visibleForTesting
bool isSafeRedirectTarget(Uri uri) {
  if (!uri.hasScheme || (uri.scheme != 'http' && uri.scheme != 'https')) {
    return false;
  }
  if (uri.host.isEmpty || uri.host == 'localhost') return false;
  final address = InternetAddress.tryParse(uri.host);
  if (address != null && (address.isLoopback || address.isLinkLocal)) {
    return false;
  }
  return true;
}

/// Substitutes `{z}`/`{x}`/`{y}` into [template] and validates the result as
/// a safe absolute redirect target, returning `null` on any failure:
/// [template] is `null`/empty, is missing one of the three tokens, fails to
/// parse as a URI once substituted, or is rejected by
/// [isSafeRedirectTarget].
///
/// Validation runs against the SUBSTITUTED URI, not the raw template, so a
/// template whose host is only malformed once tokens are filled is still
/// caught. The upstream vector template legitimately carries a query string
/// (the operator's `?key=` parameter) — substitution does not touch it.
@visibleForTesting
String? buildUpstreamRedirect(
  String? template, {
  required int z,
  required int x,
  required int y,
}) {
  if (template == null || template.isEmpty) return null;
  if (!template.contains('{z}') ||
      !template.contains('{x}') ||
      !template.contains('{y}')) {
    return null;
  }
  final substituted = template
      .replaceAll('{z}', '$z')
      .replaceAll('{x}', '$x')
      .replaceAll('{y}', '$y');
  final uri = Uri.tryParse(substituted);
  if (uri == null) return null;
  if (!isSafeRedirectTarget(uri)) return null;
  return substituted;
}

/// Constant-time equality, used by [TileProxyServer._handle] to compare a
/// request's secret path segment against the per-install secret (D-06).
/// Compares lengths once, then XOR-accumulates every code unit with no
/// early exit on the first mismatched character — an early return would let
/// a co-resident process on the same device time-oracle the secret one
/// character at a time.
bool _constantTimeEquals(String a, String b) {
  if (a.length != b.length) return false;
  var accumulator = 0;
  for (var i = 0; i < a.length; i++) {
    accumulator |= a.codeUnitAt(i) ^ b.codeUnitAt(i);
  }
  return accumulator == 0;
}

/// Bounded LRU-style cache of open [PmTilesArchive] handles, keyed by
/// absolute file path — never reopen `PmTilesArchive.fromFile` per request
/// (each open re-reads/parses the archive header and root directory).
/// Capped at a small fixed size so an unbounded number of distinct regions
/// visited in one session can't leave an ever-growing set of open file
/// handles. Evicts (and treats as a cache miss) any path whose
/// backing file no longer exists on disk, covering a mid-session region
/// delete.
class _ArchiveCache {
  static const int _capacity = 8;

  final Map<String, PmTilesArchive> _open = {};

  /// Returns the archive at [path], opening (and caching) it if not already
  /// cached. Returns `null` when [path] no longer exists on disk.
  Future<PmTilesArchive?> forPath(String path) async {
    if (!File(path).existsSync()) {
      await _evict(path);
      return null;
    }

    final cached = _open[path];
    if (cached != null) return cached;

    if (_open.length >= _capacity) {
      final oldestPath = _open.keys.first;
      await _evict(oldestPath);
    }

    final archive = await PmTilesArchive.fromFile(File(path));
    _open[path] = archive;
    return archive;
  }

  Future<void> _evict(String path) async {
    final archive = _open.remove(path);
    await archive?.close();
  }

  Future<void> closeAll() async {
    for (final archive in _open.values) {
      await archive.close();
    }
    _open.clear();
  }
}
