# Phase 39: Unified Tile Model - Pattern Map

**Mapped:** 2026-09-09
**Files analyzed:** 12 (9 modified, 3 new/new-surface)
**Analogs found:** 10 / 12 (1 genuine gap: platform channel; 1 partial: connectivity re-probe)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `app/lib/services/tile_proxy_server.dart` (modify: 302 miss path, no 404 on retryable, stable port) | service (HTTP handler) | request-response | itself — precedent lines within same file | exact (self-precedent) |
| `app/lib/util/region/offline_style_rewriter.dart` (modify: `rewriteStyleForProxy` unconditional, delete `rewriteStyleForOffline`) | transform/utility | transform | itself — `rewriteStyleForProxy` already the target shape | exact (self-precedent) |
| `app/lib/provider/map_style_json_provider.dart` (collapse two providers → one, persist style-sources) | provider | CRUD (cache read/write) | `app/lib/provider/map_style_sources_provider.dart` (network fetch to persist-and-cache) | role-match |
| new: local settings fields for proxy port + secret path | model/config | CRUD | `app/lib/entities/local_settings_entity.dart` + `app/lib/provider/local_settings_provider.dart` | exact |
| `app/lib/components/base/trail_map.dart` (delete `offline` param, both branches, both listens) | component | event-driven (Riverpod listen → native `setStyle`) | itself — see `navigation_screen.dart` for the twin fork being deleted in lockstep | exact (twin file) |
| `app/lib/routes/navigation_screen.dart` (delete `isOffline` param, same shape) | route/screen | event-driven | itself — twin of `trail_map.dart` | exact (twin file) |
| `app/lib/routes/trail_detail_map_screen.dart`, `app/lib/components/trail/trail_panel.dart`, `app/lib/routes/trail_create_screen.dart` (drop `offline:`/`isOffline:` argument) | component (caller) | request-response | `trail_map.dart` / `navigation_screen.dart` constructors being called | exact |
| `app/android/.../MainActivity.kt` (rewrite comment; add pulse hook) | platform entry point | event-driven | itself — only Kotlin file in the repo | exact (sole file) |
| new: Flutter↔Android `MethodChannel` for `setConnected` pulse | platform channel | event-driven | **none found in repo** | no analog — genuine gap |
| new/modified: connectivity-regain re-probe (drives D-15 → fires the pulse) | provider/service | event-driven | `app/lib/provider/online_status_provider.dart` (`OnlineStatus.refresh()`) + `app/lib/util/connectivity.dart` | partial (probe exists; "fire on regain" trigger does not) |
| `app/test/util/region/offline_style_rewriter_test.dart` (extend for unconditional `rewriteStyleForProxy`, maxzoom pin) | test | transform | itself — existing `_onlineStyle()` fixture pattern | exact |
| `app/test/services/tile_proxy_spike_harness.dart`, `app/test/services/tile_repository_manager_harness.dart` (extend/adapt for 302 path, stable port) | test (device harness) | request-response | itself — on-device spike harness convention | exact |

## Pattern Assignments

### `app/lib/services/tile_proxy_server.dart` (service, request-response)

**Analog:** itself, `_handle` (lines 105-206) plus the `start`/bind block (lines 71-80).

**Current miss path to replace** (lines 158-162, the whole D-02/D-04 change):
```dart
if (region == null) {
  request.response.statusCode = HttpStatus.notFound;
  return request.response.close();
}
```
Becomes a 302 redirect built from the persisted upstream tile template (D-09's `/map/style-sources`), never a 404 — a genuinely missing tile at that z/x/y (`TileNotFoundException`, lines ~193-197) still 404s; only "no region covers this tile" changes to redirect.

**Existing non-200 response pattern to match for the new redirect path** (repeated 4x in this file, e.g. lines 108-111, 118-121, 130-142):
```dart
request.response.statusCode = HttpStatus.notFound;
return request.response.close();
```
The 302 path should follow the same `statusCode` + `close()` shape, adding a `Location` header via `request.response.headers.set(HttpHeaders.locationHeader, redirectUrl)` before `close()` — mirrors the existing `Cache-Control` header-then-close pattern at lines 199-206.

**Ephemeral bind to replace (D-05)** (lines 74-77):
```dart
static Future<TileProxyServer> start(Store store) async {
  final server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
  final proxy = TileProxyServer._(server, store);
```
Port `0` (OS-assigned ephemeral) must become a persisted random port read from local settings (falling back to picking-and-persisting one on first run), keeping `InternetAddress.loopbackIPv4` unchanged (D-07 invariant). The per-install secret path segment (D-06) is a new URL-prefix check added to `_handle`'s segment parsing (currently `segments.length != 4` at line 106) — becomes `segments.length != 5` with `segments[0]` checked against the secret before falling through to `kind = segments[1]`.

**Class doc precedent for new invariants** — the file's existing doc comment block (lines 1-30) already documents the loopback-only/ephemeral-port rationale; any rewrite of D-05/D-06 must update this doc comment in place rather than leaving it describing the old ephemeral behavior.

---

### `app/lib/util/region/offline_style_rewriter.dart` (transform, transform)

**Analog:** itself — `rewriteStyleForProxy` (lines ~265-310) is already the exact function D-01 promotes to unconditional/sole use; `rewriteStyleForOffline` (lines 63-107) is the N-cell path D-17 retires.

**Core pattern already correct, reuse as-is** (lines 265-282):
```dart
Map<String, dynamic> rewriteStyleForProxy(
  Map<String, dynamic> style, {
  required String cacheRoot,
  required String proxyBaseUrl,
  bool dark = false,
}) {
  _assertSafePath(cacheRoot, 'cacheRoot');
  if (!proxyBaseUrl.startsWith('http://127.0.0.1:')) {
    throw ArgumentError.value(
      proxyBaseUrl,
      'proxyBaseUrl',
      'must start with "http://127.0.0.1:" (loopback-only)',
    );
  }
```
This loopback-prefix guard is D-07's "must not accept a non-loopback `proxyBaseUrl`" invariant, already implemented — no change needed there, only its doc comment ("The legacy `rewriteStyleForOffline`... path is left intact pending a separate cleanup", lines 261-264) needs updating once D-17 actually deletes that function.

**maxzoom pin (D-08)** already present verbatim (lines 305-309):
```dart
sourceMap['tiles'] = ['$proxyBaseUrl/dem/{z}/{x}/{y}.png'];
...
sourceMap['tiles'] = ['$proxyBaseUrl/vector/{z}/{x}/{y}.pbf'];
sourceMap['maxzoom'] = _offlinePmtilesMaxZoom;
```
using module constants `_offlinePmtilesMaxZoom = 14` / `_offlineDemMaxZoom = 12` (defined near lines 225, 234) — D-08 is "apply this same rewrite to the online path too," not new constant values.

**Deletion target (D-17):** `rewriteStyleForOffline` (starts line 63) and its two private helpers `_rewriteSourcesAndLayers` / `_rewriteSourceGroup` (lines ~117-230) — confirm no production caller before deleting (RESEARCH.md and the doc comment both say this is already the case; `app/test/util/region/offline_style_rewriter_test.dart` will need its `rewriteStyleForOffline`-specific test cases removed too).

---

### `app/lib/provider/map_style_json_provider.dart` (provider, CRUD/cache)

**Analog:** `app/lib/provider/map_style_sources_provider.dart` (network-fetch-and-hold `keepAlive` provider) for the *fetch* half; **no existing disk-persistence provider precedent** for the *persist* half — closest is `app/lib/provider/local_settings_provider.dart`'s ObjectBox-entity read/write pattern, which is the shape to imitate for persisting the style-sources JSON (an entity field or a small settings-style entity, not a raw file — matches the project's "persist small local settings via ObjectBox" convention rather than introducing ad hoc file I/O).

**Provider-fetch pattern to imitate (`map_style_sources_provider.dart`, full file, 9 lines):**
```dart
@Riverpod(keepAlive: true)
class MapStyleSourcesNotifier extends _$MapStyleSourcesNotifier {
  @override
  Future<MapStyleSources> build() async {
    final api = ref.watch(apiProvider);
    final response = await api.get('/map/style-sources');
    return MapStyleSources.fromJson(response.data);
  }
}
```
D-09 needs this to fall back to a persisted copy when the network fetch fails (cold start offline) — no existing provider in this codebase does "try network, fall back to persisted cache" for a single value; this is new machinery, closest shape is `LocalSettingsNotifier.build()`'s `box.getAll().firstOrNull ?? Default()` fallback pattern (`local_settings_provider.dart` lines 17-19).

**Deletion targets (D-10):** `offlineMapStyleJson` provider (lines ~65-77 of `map_style_json_provider.dart`), `offlineSentinelPlaceholder` constant (line 49), `fillOfflineStyleSentinels` (lines 51-60) — all three retired together since `mapStyleJson` (lines 22-39) becomes the sole provider, unconditionally sourcing from the now-persisted `mapStyleSourcesProvider`.

---

### Local settings persistence for proxy port + secret (D-05, D-06)

**Analog:** `app/lib/entities/local_settings_entity.dart` (full file) + `app/lib/provider/local_settings_provider.dart`.

**Entity pattern to extend** (add `int? tileProxyPort` and `String? tileProxySecret` fields alongside the existing ones):
```dart
@Entity()
class LocalSettingsEntity {
  @Id()
  int obxId = 0;

  String themeMode;

  bool backgroundLocationAsked;

  LocalSettingsEntity({
    this.themeMode = 'system',
    this.backgroundLocationAsked = false,
  });
}
```

**Read/write-with-invalidate pattern to copy** (`local_settings_provider.dart` lines 24-34, the `markBackgroundLocationAsked` shape — a get-default-set-invalidate mutator is the exact shape a `ensureTileProxyIdentity()` mutator needs, generating port+secret once and persisting):
```dart
Future<void> markBackgroundLocationAsked() async {
  final entity = _box.getAll().firstOrNull ?? LocalSettingsEntity();
  if (entity.backgroundLocationAsked) return;
  entity.backgroundLocationAsked = true;
  _box.put(entity);
  ref.invalidateSelf();
}
```

**Consumer/override pattern** (`app/lib/provider/region/tile_proxy_provider.dart`, full file, and `app/lib/main.dart` lines 48/69-71):
```dart
@Riverpod(keepAlive: true)
class TileProxyBaseUrl extends _$TileProxyBaseUrl {
  @override
  String build() {
    // This will be overridden in main.dart
    throw UnimplementedError();
  }
}
```
```dart
final proxyServer = await TileProxyServer.start(store);
...
tileProxyBaseUrlProvider.overrideWithValue(proxyServer.baseUrl),
```
`TileProxyServer.start(store)` (currently binding port 0) is the exact call site to change to read the persisted port/secret before binding — `main.dart`'s override wiring itself needs no change since `baseUrl` stays a computed getter.

---

### `app/lib/components/base/trail_map.dart` and `app/lib/routes/navigation_screen.dart` (component/route, event-driven)

**Analog:** each other — these are a matched pair implementing the identical fork, per RESEARCH.md §2. Delete in lockstep.

**`trail_map.dart` fork to delete** (param, lines 33/63; branches at 104, 114, 121, 130, 165, 186, 189):
```dart
final bool offline;
...
this.offline = false,
...
if (!_cacheWarmed && !widget.offline) { ... }   // line 104 — cache warm, also retired by D-11
...
if (widget.offline) { ... }                      // line 114
...
final baseAsync = widget.offline ? ... : ...;    // line 121
...
if (!widget.offline) return baseJson;            // line 165 — inside the rewrite-memo function
```

**`navigation_screen.dart` fork to delete** (param, lines 89/126; branches at 842, 1126, 1130-1137, 1152-1156, 1306-1309, 1339-1341, 1373-1386, 1742):
```dart
final bool isOffline;
...
this.isOffline = false,
...
isOffline: widget.isOffline,                     // line 842, passed to a child
...
if (!widget.isOffline) return baseJson;          // line 1126
...
final offlineStyle = rewriteStyleForProxy(...);  // lines 1130-1137 — keep the CALL, drop the guard
...
if (widget.isOffline) {
  ref.listen(offlineMapStyleJsonProvider, (_, _) => _swapStyle());   // line 1340
  ref.listen(offlineGlyphSpritePathsProvider, (_, _) => _swapStyle()); // line 1341
}
```
After D-01/D-10, `rewriteStyleForProxy` is called unconditionally on the single collapsed `mapStyleJsonProvider` output — the `ref.listen` calls at 1340-1341 collapse into the single always-registered listener the online branch already has (whatever that currently is — read the `else` arm alongside these lines during planning), and `_swapStyle`'s existing `json != _lastStyleJson` guard (RESEARCH.md §1 step 4) is the mechanism that now correctly reacts to a proxy-served style changing from miss-redirect to hit, with no `widget.offline`/`widget.isOffline` branch left to desync it.

---

### `app/android/app/src/main/kotlin/com/openwanderer/wanderer/MainActivity.kt` (platform entry point, event-driven)

**Analog:** itself — only Kotlin file in the repository, no other platform-channel precedent exists anywhere in the codebase (`app/android/`, `app/ios/` both searched — zero `MethodChannel`/`EventChannel`/`FlutterPlugin` hits).

**Current file, comment MUST be rewritten (D-14), pin MUST stay (full file, 20 lines):**
```kotlin
package com.openwanderer.wanderer

import android.os.Bundle
import io.flutter.embedding.android.FlutterActivity
import org.maplibre.android.MapLibre

class MainActivity : FlutterActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // MapLibre Native suppresses ALL online-file-source HTTP requests when
        // its ConnectivityReceiver reports no network (e.g. airplane mode) —
        // including requests to our in-app loopback tile proxy
        // (http://127.0.0.1). Forcing the connectivity override to `true`
        // disables that gate so the proxy is always queried regardless of
        // radio state. Our own offline gate (trail.isOffline) already decides
        // when the style points at the loopback proxy vs. real online tiles,
        // so this is safe offline — offline styles never carry an online URL
        // to (fail to) reach.
        MapLibre.getInstance(applicationContext)
        MapLibre.setConnected(true)
    }
}
```
The comment's final sentence ("Our own offline gate ... offline styles never carry an online URL to (fail to) reach") is the exact premise D-01 deletes (RESEARCH.md §4.4 last paragraph) — must be replaced with wording that reflects "every style is always proxied; the pin exists so the loopback proxy is reachable even in airplane mode, and the app now drives a deliberate false→true pulse on connectivity regain via a platform channel" (D-12).

**GAP — no platform-channel precedent.** Flutter's standard `MethodChannel` is not yet used anywhere in this app. The planner must design this from Flutter/Android framework conventions, not a repo analog: a `MethodChannel(name).setMethodCallHandler` in `MainActivity.onCreate` (or a small overridden `configureFlutterEngine`) reacting to a `"pulseConnectivity"` call by doing `MapLibre.setConnected(false); MapLibre.setConnected(true)`, paired with a Dart-side `MethodChannel` invocation from wherever D-15's regain-detection lives. Recommend introducing a small `lib/services/tile_connectivity_pulse.dart` (or similarly named) service as the one new call site, analogous in *shape* (a thin wrapper service, `keepAlive`-adjacent) to `TileProxyServer` itself but with no existing code to copy from.

---

### Connectivity-regain re-probe (D-15)

**Analog:** `app/lib/provider/online_status_provider.dart`'s `OnlineStatus` notifier (full pattern, lines 22-46) and `app/lib/util/connectivity.dart`'s `isConnectionFailure`/`isBackendReachable`.

**Existing optimistic-state + explicit-refresh pattern to extend, not replace:**
```dart
@Riverpod(keepAlive: true)
class OnlineStatus extends _$OnlineStatus {
  @override
  bool build() => true;

  void markOnline() {
    if (state != true) state = true;
  }

  void markOffline() {
    if (state != false) state = false;
  }

  Future<bool> refresh() async {
    if (!ref.read(apiProvider.notifier).isConfigured) return state;
    final api = ref.read(apiProvider);
    final result = await isBackendReachable(api);
    state = result;
    return result;
  }
}
```
D-15 needs a *trigger* for `refresh()`/`markOnline()` that doesn't exist today — "Claude's Discretion" explicitly leaves open whether this is `connectivity_plus` (a new dependency the project has deliberately avoided so far — zero references in `pubspec.yaml` per grep) or a periodic probe or piggybacking existing API traffic. No in-repo analog for a periodic timer or a platform connectivity listener exists; the closest *shape* precedent for "something that watches a stream and reacts" is the `ref.listen` pattern already used throughout `trail_map.dart`/`navigation_screen.dart` for style swaps (e.g. the D-16-doomed lines 1340-1341 above), which is the idiom to reuse for wiring "on markOnline() transition, fire the Android pulse method channel."

---

## Shared Patterns

### Loopback-only / path-safety invariant (D-07)
**Source:** `app/lib/services/tile_proxy_server.dart` (bind at loopbackIPv4, archive path sourced only from `region.vectorPackage.target?.localFilePath`/`region.demPackage.target?.localFilePath`, never from the request path) and `app/lib/util/region/offline_style_rewriter.dart`'s `_assertSafePath` + the `proxyBaseUrl.startsWith('http://127.0.0.1:')` guard.
**Apply to:** Every change to `tile_proxy_server.dart`'s new redirect/port/secret logic and to the style-rewriter's persisted-style-sources injection — none of these invariants may be loosened by this phase (explicit in D-07).

### `keepAlive` provider + `main.dart` override
**Source:** `app/lib/provider/region/tile_proxy_provider.dart` (9-line file) + `app/lib/main.dart` lines 48, 69-71.
**Apply to:** Any new provider this phase introduces that wraps process-lifetime state (persisted port/secret accessor, the collapsed `mapStyleJsonProvider`, a connectivity-pulse service handle) — all should be `@Riverpod(keepAlive: true)`, sourced/overridden in `main.dart` alongside the existing `objectBoxProvider`/`tileProxyBaseUrlProvider`/`cookieJarProvider` overrides, not ad hoc singletons.

### ObjectBox-entity local settings persistence
**Source:** `app/lib/entities/local_settings_entity.dart` + `app/lib/provider/local_settings_provider.dart`'s get-default-mutate-put-invalidate cycle.
**Apply to:** D-05/D-06 (port + secret) and D-09 (persisted style-sources) — both should live as fields/entities read the same way `LocalSettingsNotifier` reads `LocalSettingsEntity`, not as raw files on disk (no raw-file-persistence precedent exists in `app/lib/` outside the ObjectBox/ archive-file domains, which are for large binary payloads, not small settings).

### Non-200 response shape in `tile_proxy_server.dart`
**Source:** repeated 4x pattern, e.g. lines 108-111:
```dart
request.response.statusCode = HttpStatus.notFound;
return request.response.close();
```
**Apply to:** The new 302 path — set `statusCode = HttpStatus.movedTemporarily` (or `HttpStatus.found`), set the `Location` header, `return request.response.close();` — no body needed, matching the file's existing terse close-without-body convention on every non-200 branch.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| Flutter↔Android `MethodChannel` for the `setConnected` pulse (D-12) | platform channel | event-driven | Zero `MethodChannel`/`EventChannel`/`FlutterPlugin` usage anywhere in `app/lib` or `app/android`; `MainActivity.kt` is the only native Kotlin file and currently has no Dart-facing channel at all. Planner must design this from Flutter framework convention, not a repo precedent — flagged explicitly per the task brief. |
| Connectivity-regain detection mechanism (D-15's trigger) | provider/service | event-driven | No periodic-timer, no `connectivity_plus` (absent from `pubspec.yaml`), no OS-level connectivity listener exists today; `OnlineStatus.refresh()` is reactive-on-demand only, never self-triggering. This is new machinery; `Claude's Discretion` in CONTEXT.md leaves the exact approach open. |

## Metadata

**Analog search scope:** `app/lib/services/`, `app/lib/provider/`, `app/lib/provider/region/`, `app/lib/util/region/`, `app/lib/util/`, `app/lib/entities/`, `app/lib/components/base/`, `app/lib/routes/`, `app/android/app/src/main/kotlin/com/openwanderer/wanderer/`, `app/test/services/`, `app/test/util/region/`, `app/test/provider/`
**Files scanned:** 12 target files + 9 analog files read in full/targeted excerpt (`tile_proxy_server.dart`, `tile_proxy_provider.dart`, `local_settings_provider.dart` + entity, `map_style_sources_provider.dart`, `map_style_json_provider.dart`, `offline_style_rewriter.dart`, `online_status_provider.dart`, `connectivity.dart`, `MainActivity.kt`) plus grep sweeps for `MethodChannel`/`EventChannel`/`FlutterPlugin` (zero hits) and `offline`/`isOffline` call sites in `trail_map.dart`/`navigation_screen.dart`
**Pattern extraction date:** 2026-09-09
