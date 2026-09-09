import 'dart:convert';

/// The app's sole style transform (D-01, D-17): [rewriteStyleForProxy]
/// rewrites a MapLibre style so every URL-bearing field — vector/DEM tiles,
/// glyphs, and sprite alike — resolves through the app's loopback tile proxy
/// (`tile_proxy_server.dart`). There is no separate "offline style" and no
/// filesystem path is ever emitted; see [rewriteStyleForProxy]'s own doc
/// comment for the field-by-field rewrite.
///
/// This file used to also contain a legacy N-cell transform that pointed a
/// downloaded trail's style at native on-disk archives — one archive (and
/// one duplicated per-cell style-layer set) per downloaded 0.5° grid cell,
/// guarded by a private path-safety validator rejecting any non-absolute,
/// traversal-carrying, or foreign-scheme archive path before it could enter
/// the style. That path was retired in Phase 39 (D-17): the proxy resolves
/// per-tile coverage server-side ([resolveRegionForTile]) rather than
/// requiring the style to enumerate every downloaded region's archive up
/// front, so the N-cell duplication and its path-safety validator had no
/// remaining purpose — the surviving transform below never builds a
/// filesystem path at all, so there is nothing left for that validator to
/// check. Deleting it is therefore not a loosening of D-07: the invariant
/// D-07 actually cares about — the loopback-only prefix check on
/// `proxyBaseUrl` — is asserted unchanged inside [rewriteStyleForProxy]
/// below.
///
/// The deepest zoom level actually present in a locally-extracted `.pmtiles`
/// cell. Must match `maxZoom` in `db/services/tiles/generator.go` (currently
/// 14) — the server runs `pmtiles extract --maxzoom=14`, which is shallower
/// than a naive online `maxzoom: 15` would be (inherited from the live
/// Protomaps CDN's own, deeper tile pyramid). If a source keeps that deeper
/// online `maxzoom`, MapLibre requests nonexistent z15+ tiles directly from
/// the local archive instead of overzooming the z14 tile — rendering blank
/// above z14. This pin applies to [rewriteStyleForProxy]'s single unified
/// style, online included (D-08): a little online sharpness is traded away
/// so a downloaded region never goes blank above its local depth, since the
/// proxy cannot fake overzoom (MVT coordinates are tile-local — serving a
/// z14 parent's `.pbf` at z15 would crush the parent's geometry into the
/// child tile rather than showing more detail).
const int _proxyVectorMaxZoom = 14;

/// The deepest zoom level actually present in a locally-extracted DEM
/// `.pmtiles` cell. MUST equal the Go `demMaxZoom` const in
/// `db/services/tiles/generator.go` (currently 12) — kept in lockstep for the
/// same reason as [_proxyVectorMaxZoom]: a mismatch means MapLibre requests
/// DEM tiles that were never extracted, leaving relief blank above the cap
/// instead of overzooming. Deliberately a separate, lower constant than the
/// vector basemap's z14 — hillshading doesn't need that much detail. This
/// pin applies to [rewriteStyleForProxy]'s single unified style, online
/// included (D-08), for the same "never blank above the local depth" reason
/// as [_proxyVectorMaxZoom].
const int _proxyDemMaxZoom = 12;

/// The single unconditional style transform (D-01): every composed style,
/// online and offline alike, is routed through the client-local loopback
/// tile proxy (`tile_proxy_server.dart`). There is no separate "offline
/// style" produced by this function — tiles, glyphs and sprite all resolve
/// through the proxy, so the transform emits no filesystem path and needs no
/// cache root (D-11).
///
/// Every URL-bearing field is rewritten to a `<proxyBaseUrl>/...` loopback
/// URL:
///
///  * vector sources get a static XYZ tiles template,
///    `<proxyBaseUrl>/vector/{z}/{x}/{y}.pbf`, pinned to [_proxyVectorMaxZoom]
///    (D-08);
///  * any `type: raster-dem` source (e.g. `hillshadeSource`) gets
///    `<proxyBaseUrl>/dem/{z}/{x}/{y}.png`, pinned to [_proxyDemMaxZoom], and
///    gains `encoding: terrarium` + `tileSize: 512`, which the online style
///    relies on Mapterhorn's tilejson to supply and therefore does not carry
///    itself;
///  * `glyphs` becomes `<proxyBaseUrl>/glyphs/{fontstack}/{range}.pbf` — the
///    literal `{fontstack}`/`{range}` tokens are preserved for native
///    runtime substitution, exactly as `{z}`/`{x}`/`{y}` already are;
///  * `sprite` becomes `<proxyBaseUrl>/sprite/<light|dark>` (via [dark]) —
///    no file suffix is appended, since MapLibre's sprite loader appends
///    `.json`/`.png`/`@2x.*` itself, matching the set
///    `map_cache_path.dart`'s `allowedSpriteFileNames` whitelists.
///
/// No `__cellN` source/layer cloning is produced (the retired legacy
/// transform's approach — see the file-level doc comment), because the proxy
/// resolves per-tile coverage server-side ([resolveRegionForTile]) rather
/// than the style needing to enumerate every downloaded region's archive up
/// front. A tile the proxy has no local coverage for is redirected to the
/// operator's upstream template rather than answered locally, so coverage
/// degrades per tile rather than per screen; glyphs/sprites are answered
/// local-first with write-through into `map_cache` — both decided inside
/// `tile_proxy_server.dart` rather than by the style.
///
/// [proxyBaseUrl] must start with `http://127.0.0.1:` — any other base is
/// rejected (D-07: defense in depth, this style must never point at a
/// non-loopback host).
///
/// The input [style] is deep-copied before any mutation, so the shared
/// online base JSON (from the `keepAlive` style provider) is never
/// corrupted.
Map<String, dynamic> rewriteStyleForProxy(
  Map<String, dynamic> style, {
  required String proxyBaseUrl,
  bool dark = false,
}) {
  if (!proxyBaseUrl.startsWith('http://127.0.0.1:')) {
    throw ArgumentError.value(
      proxyBaseUrl,
      'proxyBaseUrl',
      'must start with "http://127.0.0.1:" (loopback-only)',
    );
  }

  // Deep copy so the shared online base style is never mutated in place.
  final out = jsonDecode(jsonEncode(style)) as Map<String, dynamic>;

  out['glyphs'] = '$proxyBaseUrl/glyphs/{fontstack}/{range}.pbf';
  out['sprite'] = '$proxyBaseUrl/sprite/${dark ? 'dark' : 'light'}';

  final sources = out['sources'];
  if (sources is Map<String, dynamic>) {
    for (final key in sources.keys.toList()) {
      final source = sources[key];
      if (source is! Map) continue;
      final isTiled = source.containsKey('tiles') || source.containsKey('url');
      if (!isTiled) continue;

      final sourceMap = source as Map<String, dynamic>;
      sourceMap.remove('url');
      if (sourceMap['type'] == 'raster-dem') {
        sourceMap['tiles'] = ['$proxyBaseUrl/dem/{z}/{x}/{y}.png'];
        sourceMap['encoding'] = 'terrarium';
        sourceMap['tileSize'] = 512;
        sourceMap['maxzoom'] = _proxyDemMaxZoom;
      } else {
        sourceMap['tiles'] = ['$proxyBaseUrl/vector/{z}/{x}/{y}.pbf'];
        sourceMap['maxzoom'] = _proxyVectorMaxZoom;
      }
    }
  }

  return out;
}

