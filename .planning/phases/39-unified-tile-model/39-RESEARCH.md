# Phase 39 — Research Source: Unified Tile Model

Captured 2026-09-09 from a `/gsd-explore` session. Everything a planner needs is here;
no prior conversation context is required.

---

## 1. The bug that started this

A hiker opens a trail map with no service. `TrailMap` renders from downloaded `.pmtiles`.
They walk back into coverage — and the map **never** picks up online tiles, for the rest of
that widget's life.

`TrailMap.offline` *is* reactive (`trail_detail_map_screen.dart` passes
`offline: !ref.watch(onlineStatusProvider)`), but the flip cannot reach the native map:

1. `widget.offline` flips false → `build()` recomposes the online style and **assigns it to
   `_lastStyleJson`** (`trail_map.dart:136-137`).
2. `_buildMap` returns the **cached `_mapOptions`**, rebuilt only when `disabled` flips
   (`trail_map.dart:219`). `initStyle` still carries the offline style, and it is the same
   object instance, so the plugin's own `didUpdateWidget` early-outs too.
3. `setStyle` is only ever called from `_swapStyle`, which is wired to `ref.listen` on the
   style/cache providers — never to `widget.offline`.
4. `_swapStyle` guards on `json != _lastStyleJson`, which step 1 already updated. **So even a
   later theme toggle cannot rescue it.** The map is pinned to whatever mode it was born in.

Same shape in `navigation_screen.dart:1339-1341`.

Secondary defect: `onlineStatusProvider` is optimistic-`true` and only moves on API traffic or
an explicit `refresh()`. Nothing re-probes when the radio regains service, and there is no
`connectivity_plus` dependency. Even a fixed `TrailMap` would sit offline until some unrelated
request happened to succeed.

## 2. The duality is forked in three places

| Concern | Online | Offline |
|---|---|---|
| Style JSON | `mapStyleJsonProvider` → network `/map/style-sources` | `offlineMapStyleJsonProvider` → `about:blank` sentinels |
| Glyphs/sprite | operator CDN URLs | `file://<map_cache>/…` |
| Tiles | operator XYZ CDN | loopback proxy → pmtiles, per-tile `resolveRegionForTile` |

Only the third is genuinely binary per tile. The other two are "use the cache, refresh when you
can" and are forked for no structural reason.

## 3. Decision: collapse to one always-proxied style

A MapLibre style gives each source exactly one `tiles` template, resolved at style load. There
is no "try local, then network" primitive in the style spec and no per-tile hook in MapLibre
Native's file source. The fallback must therefore live behind a single URL that MapLibre treats
as dumb XYZ.

`TileProxyServer` already is that thing — it resolves coverage fresh, per request
(`tile_proxy_server.dart:147`). Today it discards the answer on a miss and 404s
(`tile_proxy_server.dart:158`). **That line is the whole change:** on a miss, redirect (302) to
the operator's upstream CDN tile URL.

Consequences: there is no offline mode. One style, always proxied. Coverage degrades per tile,
per area, instantly. No `offline` parameter, no style swap on connectivity, no stuck map — the
bug class above becomes unrepresentable rather than merely fixed.

### Why redirect rather than reverse-proxy the bytes

Capability-equivalent; this is a cost decision.

| | Redirect (302 → CDN) | Reverse-proxy (Dart fetches, pipes bytes) |
|---|---|---|
| Per online tile, Dart does | parse path, memoized region lookup, write ~200-byte header | all that **plus** open HTTP client, buffer 50–200 KB into the Dart heap, copy to response |
| Who fetches | MapLibre's platform stack — own thread pool, connection pool, ambient disk cache | the root isolate's event loop |
| Online tile cost vs. today | ~unchanged | every tile in normal use transits the UI isolate |

Today online tiles never touch Dart at all. That isolate is already contended: map panning was
measured at roughly two cores, `TextureViewRend` alone at 76% while panning
(`trail_map.dart:230-241`). Piping tens of tile bodies per pan through it is the one way this
rework could end up worse than the bug it fixes.

**Reverse-proxy is the fallback if the redirect ever proves unusable — not a second design.**

### Scope decision: no opportunistic tile capture

CDN-fetched tiles are **not** captured for durable offline use. Downloaded regions remain the
only promised offline coverage; MapLibre's ambient cache is an unpromised bonus. This avoids a
writable tile store (the `pmtiles` 1.2.0 Dart package is read-only —
`offline_style_rewriter.dart:36-38`), an eviction policy, and a storage-usage story in settings.
It also keeps "downloaded" meaning exactly one thing in the regions UI.

---

## 4. MapLibre Native research findings

Researched against MapLibre Native `main`, plus the shipped Android artifact. Each finding is
marked CONFIRMED or UNCERTAIN.

### 4.1 Does MapLibre Native follow 3xx on tile requests? **CONFIRMED — yes, both platforms**

- **Android.** `HttpRequestImpl.java:205` builds
  `new OkHttpClient.Builder().dispatcher(getDispatcher()).build()` — the only builder
  configuration, with no `followRedirects()` / `followSslRedirects()` call, so both stay at
  OkHttp's default `true` (20-hop limit). `followSslRedirects` governs the cross-scheme
  `http://127.0.0.1 → https://cdn` hop, and it too defaults true. Verified against the actual
  shipped artifact — `android-sdk-opengl-13.0.3-pre0.aar`, class
  `org/maplibre/android/module/http/HttpRequestImpl.class`: its constant pool references only
  `okhttp3/OkHttpClient$Builder` and `okhttp3/Dispatcher`, no redirect symbol.
- **iOS.** `platform/darwin/core/http_file_source.mm:86` —
  `session = [NSURLSession sessionWithConfiguration:sessionConfig]`, created with **no
  delegate**, request issued via `dataTaskWithRequest:completionHandler:`. With no
  `URLSession:task:willPerformHTTPRedirection:` delegate, NSURLSession follows redirects
  automatically. `MLNNetworkConfiguration.mm:84` even asserts delegate-supplied sessions must
  not conform to `NSURLSessionDataDelegate`.
- **Status mapping sees only the final response** (`darwin/core/http_file_source.mm:381-408`,
  `android/src/cpp/http_file_source.cpp:164-193`). The 302 is invisible to MapLibre's error
  handling, and its `Cache-Control` is never read.

### 4.2 A 404 permanently poisons the tile — **CONFIRMED, and it dictates the design**

- `Reason::NotFound` gets a backoff of `Duration::max()` — **never retried**
  (`src/mln/util/http_timeout.cpp`; `online_file_source.cpp:475` returns on `Duration::max()`).
- `noContent` (204, or 404 on a tile resource) is **persisted as an empty tile row**
  (`offline_database.cpp:678`). Errors are never cached (`putInternal` returns early
  `if (response.error)`, `offline_database.cpp:313`) — but `noContent` is not an error.

**Rule: never 404 a tile that might succeed later. Always redirect on a coverage miss.** When
the device is offline the redirected CDN fetch fails with `Reason::Connection`, which backs off
at `2^(n-1)` s *and* is the one failure class retried on a connectivity edge (4.3). The proxy
therefore needs **no connectivity awareness at all**. Reserve 404 for the genuinely permanent
case.

> An earlier draft of this design had the proxy 404 when the app knew it was offline. That would
> have reproduced the original bug in a more durable form — permanently cached blank tiles.
> Do not reintroduce it.

### 4.3 Retry and recovery — **CONFIRMED**

- **Backoff** (`http_timeout.cpp`): Connection → `2^(n-1)` s; Server (5xx) → 1 s ×3 then
  exponential; RateLimit → `Retry-After`; everything else including NotFound and `Reason::Other`
  → `Duration::max()`, never retried.
- **`setStyle` with identical JSON does re-request.** `Style::Impl::loadJSON` → `parse()`
  (`src/mln/style/style_impl.cpp:46,86`) unconditionally does `sources.clear(); layers.clear();`
  and re-adds; no identical-JSON short circuit, and neither wrapper adds one
  (`MLNMapView.mm:628`, `MapLibreMap.java:1140`). Works, but heavy-handed — full style reload,
  layers and images rebuilt.
- **Connectivity edge is the targeted mechanism.** `NetworkStatus::Set`
  (`src/mln/storage/network_status.cpp:24`) fires `Reachable()` **only on a genuine false→true
  edge**, reaching `OnlineFileRequest::networkIsReachableAgain()`
  (`online_file_source.cpp:587`), which does `schedule(Duration::zero())` **only for requests
  whose last failure was `Reason::Connection`**.

**Android gets no free retries today.** `MainActivity.kt:21` pins `MapLibre.setConnected(true)`
permanently, so the false→true edge never occurs — and `ConnectivityReceiver.onReceive`
early-returns while the override is non-null (`ConnectivityReceiver.java:95`), so system
broadcasts cannot produce one either. **Fix: keep the pin (loopback still needs it in airplane
mode — see 4.4) and add a deliberate `setConnected(false)` → `setConnected(true)` pulse on
connectivity regain.** The offline gate sits at `schedule()` (`online_file_source.cpp:500`), so
the `false` window must be brief.

**iOS gets recovery free.** `MLNReachability` regaining a connection already calls
`Reachable()`, retrying Connection-failed requests.

### 4.4 iOS needs no `setConnected` counterpart — **CONFIRMED**

The suppression gate is core C++ and URL-agnostic (`online_file_source.cpp:150` and `:500`) — it
would block `127.0.0.1` too. But `NetworkStatus::online` initialises `true`
(`network_status.cpp:11`), and a full grep of `platform/ios` + `platform/darwin` finds exactly
one `mln::NetworkStatus` call site: `MLNMapView.mm:1008 → NetworkStatus::Reachable()` in
`reachabilityChanged:`. There is **no** `NetworkStatus::Set(Offline)` and no `setConnected`
equivalent in Apple platform code. Android's gate exists solely because
`connectivity_listener.cpp:19` drives it from `ConnectivityReceiver`. iOS stays permanently
"online" and never suppresses loopback requests.

The `MainActivity.kt:11-19` comment justifies the pin with *"offline styles never carry an
online URL to (fail to) reach."* **This phase deletes that premise** — the comment must be
rewritten, not just left standing.

### 4.5 The ambient cache keys on the loopback URL — **CONFIRMED; adds real work**

The ambient-cache key is `resource.url`, i.e. the **loopback** URL, not the CDN URL
(`offline_database.cpp:21-23`). `TileProxyServer` binds an **ephemeral** port
(`tile_proxy_server.dart:75`), so the port changes every launch and the entire tile cache is
orphaned on every cold start.

Today this is minor latent waste — the `max-age=86400` at `tile_proxy_server.dart:203-206` only
ever helps within a single app run, never across launches. **Under unification it becomes a
genuine regression:** online tiles currently cache under stable `https://` CDN URLs; routing
them through the proxy moves them onto churning keys, so every cold start re-downloads
everything the user looks at.

**A stable port is therefore a prerequisite of this phase, not a nice-to-have.** It collides
with the deliberate security choice documented at `tile_proxy_server.dart:29-32` — ephemeral
precisely so "a co-resident app must not be able to pre-target a guessable port." Both
properties are obtainable: **pick the port once at install and persist it** (stable but
unguessable), plus optionally a per-install secret path segment. Cache keys stay stable across
launches; the port stays unpredictable to a co-resident app.

### 4.6 maxzoom must be pinned to the local depth — **design decision**

One source, one `maxzoom`. Local pmtiles stop at z14, DEM at z12; the online style says z15
(`offline_style_rewriter.dart:222-238`).

Setting 15 means an offline hiker inside a downloaded region gets blank tiles above z14 — the
navigation case, the one that matters most. The proxy cannot paper over it: serving a z14
parent's `.pbf` at z15 coordinates renders the parent's geometry crushed into the child tile,
because MVT coordinates are tile-local with a fixed extent.

**Pin vector to 14 and DEM to 12** and let MapLibre overzoom online. Vector overzoom is visually
mild; DEM overzoom is invisible. Trades a little online sharpness for correctness in both
states, and avoids generating a whole class of guaranteed-to-fail z15 requests while offline.
One constant each; reversible.

### 4.7 Glyphs and sprites

Always-`file://` breaks a first-run-offline user — no warm cache, no labels. The consistent
answer is to route glyphs and sprites through the proxy too (`/glyphs/…`, `/sprite/…`) under the
same local-first rule. Volume is a handful of requests per style load, so reverse-proxying those
bytes is free, and the proxy can write through into `map_cache` — which retires the explicit
warm step at `trail_map.dart:104-107`.

Optional for a first cut; it is what makes the model uniform.

### 4.8 Prerequisite: persist `/map/style-sources`

The proxy needs the operator's upstream tile template to build redirect targets, and an offline
cold start has to know it. Persisting the style-sources response also collapses
`mapStyleJsonProvider` / `offlineMapStyleJsonProvider` into one provider.

---

## 5. Open questions the planner must close

1. **`setConnected` pulse on a real device.** Does a brief `false` → `true` window drop
   **in-flight loopback requests**? The gate is at `schedule()`, so a synchronous pulse *should*
   be safe, but this is untested. If it is not safe, fall back to the `setStyle` reload (4.3).
   Verify on a physical Android device, airplane mode → service restored.
2. **`Range` header × redirect.** MapLibre sets a `Range` header for `resource.dataRange`
   requests (PMTiles range reads). Whether NSURLSession's automatic redirect-follow preserves or
   mangles `Range` across the hop was **not tested**. Only matters if a ranged request can hit
   the redirect path.
3. **iOS binary confirmation.** The iOS findings in 4.1 and 4.4 come from MapLibre Native
   `main`, **not** from the shipped pod (`MapLibre ~> 6.25`). The redirect/delegate code there is
   long-stable, but it is unverified for that exact tag. Android *was* verified against the
   shipped `13.0.3-pre0` AAR.
4. No upstream issue/PR search was performed for known redirect bugs; 4.1 is source-derived only.

---

## 6. Surfaces this phase touches

| File | Change |
|---|---|
| `app/lib/services/tile_proxy_server.dart` | miss path → 302 to upstream CDN, **never** 404 for a retryable tile; inject upstream tile template; stable persisted port + secret path segment |
| `app/lib/util/region/offline_style_rewriter.dart` | `rewriteStyleForProxy` becomes *the* transform, applied unconditionally; retire the legacy `rewriteStyleForOffline` cell-duplication path its own docs already mark as pending cleanup |
| `app/lib/provider/map_style_json_provider.dart` | collapse to one provider; persist style-sources; sentinels always resolve to proxy URLs |
| `app/lib/components/base/trail_map.dart` | delete `offline`, both compose branches, both `ref.listen` forks; `_swapStyle` handles theme only; drop the `_cacheWarmed` network warm |
| `app/lib/routes/navigation_screen.dart` | same, for `isOffline` |
| `app/lib/routes/trail_detail_map_screen.dart:122`, `app/lib/components/trail/trail_panel.dart:159`, `app/lib/routes/trail_create_screen.dart:1232` | drop the argument |
| `app/android/.../MainActivity.kt` | keep `setConnected(true)`; **rewrite the comment** (its stated justification stops being true); expose a pulse hook over a platform channel |
| new | connectivity-regain re-probe (no `connectivity_plus` dependency today) → pulse `setConnected` on Android; iOS needs nothing |

**Free win:** `app/lib/models/trail.dart:110-116` records a shipped bug where
`TrailMap(offline: trail.isOffline)` conflated "downloaded" with "no connectivity." Deleting the
parameter makes that bug unrepresentable rather than merely fixed.
