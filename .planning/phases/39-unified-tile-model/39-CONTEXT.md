# Phase 39: Unified Tile Model — Context

**Gathered:** 2026-09-09
**Status:** Ready for planning
**Source:** `/gsd-explore` session (2026-09-09) — decisions locked interactively with the developer

<domain>
## Phase Boundary

**In scope.** Collapse the app's binary offline/online map model into a single always-proxied
tile path, so map coverage degrades and recovers per tile rather than per screen. Concretely:
the loopback tile proxy gains a network-fallback miss path; the duplicated style/glyph/tile
providers collapse to one; the `offline` / `isOffline` widget parameters are deleted; the
loopback port becomes stable so MapLibre's ambient cache survives a cold start; and Android
gains a mechanism to force a tile retry when connectivity returns.

**Out of scope.** Capturing CDN-fetched tiles for durable offline use (see `<deferred>`).
Anything touching region download, region selection, or the regions settings UI. Any change to
what "downloaded" means to the user. Web (SvelteKit) and Go backend are untouched — this is a
Flutter/`app/` phase plus one Kotlin file.

**Defect being retired.** A map opened without service can never pick up online tiles for the
life of the widget: `widget.offline` flipping recomposes `_lastStyleJson` in `build()` but
`_buildMap` returns cached `_mapOptions`, and `_swapStyle`'s `json != _lastStyleJson` guard is
already satisfied — so even a later theme toggle cannot rescue it. Full trace in
`39-RESEARCH.md` §1.

</domain>

<decisions>
## Implementation Decisions

### Architecture

- **D-01** — One style, always proxied. Every map style routes its tile sources through the
  loopback `TileProxyServer`, online and offline alike. There is no "offline style" and no
  "online style"; `rewriteStyleForProxy` becomes the single unconditional transform.

- **D-02** — The proxy answers a coverage miss by redirecting (HTTP 302) to the operator's
  upstream CDN tile URL. It MUST NOT fetch the tile itself and pipe the bytes back. Rationale:
  online tile bytes must never transit the root isolate, which is already contended during
  panning (~2 cores; `TextureViewRend` at 76%). Reverse-proxying is the documented fallback if
  redirect proves unusable in practice — not a co-equal option to pick from.

- **D-03** — The proxy is connectivity-blind. It does not read `onlineStatusProvider` or any
  network state to decide between redirecting and 404ing. It always redirects on a miss and
  lets the failure surface as MapLibre's `Reason::Connection`.

### Tile proxy behavior

- **D-04** — Never answer a retryable tile with 404. MapLibre persists `noContent` as an empty
  tile row (`offline_database.cpp:678`) and backs `NotFound` off to `Duration::max()`, so a 404
  becomes a permanently cached blank tile that is never re-requested. 404 is reserved for the
  genuinely permanent case. This decision reverses an earlier draft of the design; see
  `39-RESEARCH.md` §4.2.

- **D-05** — The loopback port is chosen once and persisted, not assigned per launch.
  MapLibre's ambient cache keys on `resource.url` — the loopback URL — so today's ephemeral
  bind orphans the whole tile cache on every cold start. Unguessability (the property
  `tile_proxy_server.dart:29-32` deliberately chose) is preserved by persisting a *random*
  port rather than a fixed one, plus a per-install secret path segment.

- **D-06** — The secret path segment is per-install stable. A per-launch secret would defeat
  D-05 by churning the cache key just as an ephemeral port does.

- **D-07** — Path-traversal and loopback-only invariants are preserved exactly as they stand.
  Archive paths continue to come only from a region's own `localFilePath`, never assembled from
  the request path; the server continues to bind `InternetAddress.loopbackIPv4`;
  `rewriteStyleForProxy` continues to reject any `proxyBaseUrl` that is not `http://127.0.0.1:`.

### Style composition

- **D-08** — Vector source `maxzoom` is pinned to 14 and `raster-dem` to 12 — the local pmtiles
  depths — for the unified style, online included. A z15 source would blank downloaded regions
  above z14 while offline, and the proxy cannot fake overzoom: MVT coordinates are tile-local,
  so a z14 parent's `.pbf` served at z15 renders the parent's geometry crushed into the child
  tile. Costs a little online sharpness; MapLibre overzooms.

- **D-09** — `/map/style-sources` is persisted to disk. The proxy needs the operator's upstream
  tile template to build redirect targets, and an offline cold start must know it.

- **D-10** — `mapStyleJsonProvider` and `offlineMapStyleJsonProvider` collapse into one
  provider. The `offlineSentinelPlaceholder` / `fillOfflineStyleSentinels` pair goes away with
  them.

- **D-11** — Glyphs and sprites route through the proxy under the same local-first rule
  (`/glyphs/…`, `/sprite/…`), with write-through into `map_cache`. Volume is a handful of
  requests per style load, so reverse-proxying those bytes is acceptable here — D-02's
  constraint is about per-tile volume. This retires the explicit cache warm at
  `trail_map.dart:104-107` and fixes first-run-offline rendering with no labels.

### Connectivity and recovery

- **D-12** — ~~Recovery on Android is a `setConnected(false)` → `setConnected(true)` pulse.~~
  **REJECTED 2026-09-09 by on-device test** (see `39-01-SUMMARY.md` `## Risk gate outcome`).
  The pulse fires correctly — channel round trip 2ms, `result.success(null)`, no exceptions —
  and MapLibre re-schedules **nothing**: 82 Connection-class tile failures before the pulse,
  **zero** tile requests after. The failed tiles are not pending requests waiting on a
  connectivity edge; they are errored entries in the source's tile pyramid, already torn down
  (`Request failed due to a permanent error: Canceled` appears *before* the pulse). There is
  nothing for `networkIsReachableAgain()` to act on.

  Note the D-12 rationale as originally written overstated its evidence: it could NOT be
  confirmed that `UnknownHostException` maps to `Reason::Connection` in this build — the three
  `org/maplibre/android/module/http/*` classes in the shipped
  `android-sdk-opengl-13.0.3-pre0.aar` reference no exception types, so the classification is
  C++-side and unverified from the artifact.

- **D-12a** *(replaces D-12, decided 2026-09-09)* — **There is no automatic recovery
  mechanism.** Recovery is user-initiated: panning or zooming requests tiles at new
  coordinates, which are fresh requests and succeed immediately once service is back. The
  developer accepted this explicitly. The `setStyle` reload fallback is NOT implemented — it
  rebuilds every source, layer and image, dropping and re-adding the navigation screen's trail
  track and breadcrumb on every connectivity regain, which is disproportionate to the problem.

  This still kills the reported defect. The original complaint was that *nothing* on the screen
  could bring online tiles back, because the offline style pointed every tile at the proxy with
  no upstream to fall through to — panning into uncovered area meant blank forever. D-02's
  redirect fixes that on its own.

  **Known limitation, accepted:** `trail_panel.dart` mounts `TrailMap(disabled: true,
  embedded: true, …)`, and `disabled` resolves to `MapGestures.none()`
  (`trail_map.dart:225`). That map cannot be pan-recovered; its fixed camera never requests new
  tile coordinates, so it recovers only when the widget remounts on navigation. Strictly better
  than today (where it never recovers), and deliberately not mitigated.

- **D-12b** — The now-dead pulse code is **removed**, not retained: `maplibre_connectivity_pulse.dart`,
  the Kotlin `MethodChannel` handler and its `configureFlutterEngine` override, and the spike
  harness's pulse control. `MainActivity.kt`'s comment reverts to justifying only the
  `setConnected(true)` pin. Leaving an unused channel whose doc comment describes a rejected
  mechanism is the exact hazard D-14 exists to prevent.

- **D-13** — iOS gets no connectivity code. `MLNReachability` already fires `Reachable()` on
  regain, and Apple platform code contains no `NetworkStatus::Set(Offline)` call, so iOS never
  suppresses loopback requests. Do not add a `setConnected` counterpart there.

- **D-14** — `MainActivity.kt` keeps its `setConnected(true)` pin, but its comment MUST be
  rewritten. The comment currently justifies the pin with "offline styles never carry an online
  URL to (fail to) reach" — a premise D-01 deletes. Leaving it standing would mislead the next
  reader into thinking the pin is still safe for the stated reason.

- **D-15** — ~~Something must re-probe connectivity when the radio regains service.~~
  **DESCOPED 2026-09-09.** This existed only to trigger D-12's recovery. With D-12a there is
  nothing to trigger, and the map no longer consults `onlineStatusProvider` at all — D-03 makes
  the proxy connectivity-blind, so tile rendering is fully decoupled from that provider.
  `onlineStatusProvider`'s optimistic-`true` staleness remains exactly as it was before this
  phase, affecting the same non-map consumers (`guard_online`, `resolve_track_save_options`,
  `import_trail_file`, `trail_sync_provider`, the offline banner). Pre-existing, not this
  phase's to fix.

### Deletions

- **D-16** — `TrailMap.offline` and `NavigationScreen.isOffline` are deleted outright, along
  with both compose branches and both `ref.listen` forks in each. Callers at
  `trail_detail_map_screen.dart:122`, `trail_panel.dart:159` and `trail_create_screen.dart:1232`
  drop the argument. This makes `TrailMap(offline: trail.isOffline)` — the "downloaded" vs "no
  connectivity" conflation documented at `trail.dart:110-116` — unrepresentable rather than
  merely fixed.

- **D-16a** *(added 2026-09-09, after planning)* — the deletion extends to storage: the persisted
  `ActiveNavigationEntity.isOffline` ObjectBox column is dropped too, not retained. The developer
  confirmed the app is not in production, but that is the weaker reason. The stronger one: the
  column only ever passed a value between `main.dart`'s `_pushRecordingResume` and
  `router_provider.dart`'s `/record` builder one call later in the same process, and D-16 deletes
  both ends. `main.dart` already documents that the persisted value "is deliberately not trusted"
  and re-probes on every resume, so it was never durable state. Retaining it would leave a write
  with no reader. Cost is one generated `retiredPropertyUids` entry — `lib/objectbox-model.json`
  already carries 41.

- **D-17** — The legacy `rewriteStyleForOffline` N-cell duplication path is retired in this
  phase. Its own doc comment already marks it as pending cleanup, and D-01 leaves it with no
  production caller.

### Claude's Discretion

- Where the persisted port and secret live (local settings vs. a dedicated file) and how they
  are generated.
- The shape of the Android platform channel for the D-12 pulse, and which layer owns firing it.
- How connectivity regain is detected for D-15 — a new dependency, a periodic probe, or
  piggybacking existing traffic. Weigh against the app's existing no-`connectivity_plus` stance.
- Whether `TileProxyServer`'s upstream template is injected at construction or read from a
  provider per request.
- Test strategy and how much of the proxy's new miss path is covered by the existing
  `tile_repository_manager_harness.dart` / `tile_proxy_spike_harness.dart` scaffolding.
- Plan/wave decomposition.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### This phase's own research
- `.planning/phases/39-unified-tile-model/39-RESEARCH.md` — the complete design and the
  MapLibre Native findings behind it, each marked CONFIRMED or UNCERTAIN with file+symbol
  citations. §5 lists the three open questions the planner must close.

### Primary implementation surfaces
- `app/lib/services/tile_proxy_server.dart` — the loopback proxy; `_handle` is where D-02/D-04
  land, the `_ArchiveCache` and the ephemeral bind at line 75 are where D-05 lands.
- `app/lib/util/region/offline_style_rewriter.dart` — `rewriteStyleForProxy` (D-01, D-08) and
  the legacy `rewriteStyleForOffline` to retire (D-17).
- `app/lib/provider/map_style_json_provider.dart` — the provider pair to collapse (D-09, D-10).
- `app/lib/components/base/trail_map.dart` — `offline` deletion, dual compose/listen branches
  (D-16); the cache warm at 104-107 (D-11).
- `app/lib/routes/navigation_screen.dart` — same, for `isOffline` (compose at 1112-1160, gates
  at 1306-1341, 1373-1386, 1742).
- `app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt` — D-12, D-14.

### Behavioral contracts that must not regress
- `app/lib/provider/online_status_provider.dart` + `app/lib/util/connectivity.dart` — the
  reachability model D-15 extends. Note `isConnectionFailure`'s deliberate exclusions.
- `app/lib/provider/region/tile_proxy_provider.dart` — the `keepAlive` provider whose
  `baseUrl` is baked into every composed style; D-05 changes what that URL looks like.
- `app/test/services/tile_repository_manager_test.dart`,
  `app/test/util/region/offline_style_rewriter_test.dart`,
  `app/test/provider/offline_map_style_json_test.dart` — existing coverage that this phase
  will invalidate or must keep green.

### Project conventions
- `CLAUDE.md` — project instructions.
- Memory: the user builds and installs the app; never run `flutter build` or `adb install`.
  Hand off after `flutter analyze` + `flutter test`.

</canonical_refs>

<specifics>
## Specific Ideas

- The whole change to the proxy's decision logic is one branch: `tile_proxy_server.dart:158`,
  `if (region == null) { … 404 … }` becomes the redirect path.
- `TileProxyServer` already resolves coverage fresh per request
  (`resolveRegionForTile`, line 147) — chosen originally to kill a reconcile race. That
  per-request resolve is exactly the hook the hybrid model needs; no new resolution
  machinery is required.
- The `Cache-Control: public, max-age=86400` at lines 203-206 currently only helps within a
  single app run because of the ephemeral port. D-05 makes it do what it was written to do.
- iOS/Android asymmetry is expected and correct (D-12 vs D-13) — do not "fix" it into symmetry.

</specifics>

<deferred>
## Deferred Ideas

- **Opportunistic tile capture.** CDN-fetched tiles are explicitly NOT captured for durable
  offline use. Downloaded regions stay the only promised offline coverage, so "downloaded"
  keeps meaning exactly one thing in the regions UI. Doing otherwise needs a writable tile
  store (the `pmtiles` 1.2.0 Dart package is read-only), an eviction policy, and a
  storage-usage surface in settings. Decided out of scope in the exploration session.
- **Restoring true z15 online detail.** D-08 trades it away. Revisitable by serving z15 from
  the CDN only, once there is a way to express per-zoom source routing.
- **iOS binary verification** of the redirect/delegate findings against the shipped
  `MapLibre ~> 6.25` pod (research was against `main`). Tracked as open question 3 in
  `39-RESEARCH.md` §5.

</deferred>

---

*Phase: 39-unified-tile-model*
*Context gathered: 2026-09-09 via /gsd-explore session*
