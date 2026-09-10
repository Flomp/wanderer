---
phase: 39-unified-tile-model
audited: 2026-09-10
asvs_level: 1
block_on: high
register_authored_at_plan_time: true
threats_total: 39
threats_closed: 39
threats_open: 0
---

# Phase 39: Unified Tile Model — Security Audit

**Verdict: SECURED — 39/39 threats CLOSED.**

Verified against the shipped tree, including the post-review fix pass (commit `e84ac158`,
confirmed present in `git log`), not against plan-time claims. All cited grep/test evidence was
re-run during this audit, not taken from SUMMARY.md narration.

## Count reconciliation

The task brief stated "4 accept" dispositions. Independent extraction from all nine plans'
`<threat_model>` blocks finds only **3** unique accept-disposition threats: T-39-07, T-39-19, and
T-39-SC (which recurs identically in all nine plans but is one threat, as instructed). 34
mitigate + 3 accept + 2 eliminated = 39, which reconciles with the stated "39 unique threats"
total. Flagging this discrepancy rather than silently adopting the stated count of 4.

## Elimination note (T-39-01, T-39-02, T-39-03)

These three threats describe the Android `setConnected` pulse mechanism (Plan 01), which
on-device testing rejected (`39-01-SUMMARY.md` `## Risk gate outcome`: 82 Connection-class tile
failures before the pulse, zero tile requests after). Plan 09 deleted the mechanism entirely.
Verified deletion is complete:
- `app/lib/services/maplibre_connectivity_pulse.dart` does not exist on disk.
- `grep -rc "pulseMapLibreConnectivity|maplibre_connectivity_pulse|MethodChannel|configureFlutterEngine" app/lib app/android/.../MainActivity.kt` → 0 matches everywhere.
- `MainActivity.kt` retains `MapLibre.getInstance(...)` / `MapLibre.setConnected(true)` unchanged, with a rewritten comment that names no pulse, no method channel, and no `false`->`true` edge, and explicitly states recovery is user-initiated (cites CONTEXT.md D-12a and `39-01-SUMMARY.md`).

Classified CLOSED by elimination, not by the originally-authored `mitigate` mechanism.

## Threat Verification

| Threat ID | Category | Disposition | Status | Evidence |
|---|---|---|---|---|
| T-39-01 | Elevation of Privilege | mitigate → **eliminated** | CLOSED | `MainActivity.kt` has no `MethodChannel`/`configureFlutterEngine`; code deleted in `c09aa100` (Plan 09) |
| T-39-02 | Denial of Service | mitigate → **eliminated** | CLOSED | Pulse mechanism deleted; no `setConnected(false)` call exists anywhere in the tree |
| T-39-03 | Tampering | mitigate → **eliminated** | CLOSED | `maplibre_connectivity_pulse.dart` deleted; no iOS platform file ever touched, no dead symmetry risk remains |
| T-39-04 | Spoofing | mitigate | CLOSED | `tile_proxy_identity.dart:69` `Random.secure()` default; 128-bit secret (16 bytes hex), port from IANA dynamic range `[49152,65535]` |
| T-39-05 | Tampering | mitigate | CLOSED | `tile_proxy_identity.dart:52,57` `^[0-9a-f]{32}$` anchored regex gates `resolveTileProxyIdentity` (line 94-98) before a persisted secret is trusted |
| T-39-06 | Tampering | mitigate | CLOSED | `map_source_persistence.dart:27-33` `try/catch` returns `null` on any decode failure; never trusted as URL at read time (revalidated at use, see T-39-11) |
| T-39-07 | Information Disclosure | accept | CLOSED | Accepted risk recorded in this SECURITY.md's Accepted Risks Log below; net exposure verified narrowed — `proxy_style_rewriter.dart:124` `sourceMap.remove('url')` means the CDN `?key=` URL never reaches the style JSON handed to MapLibre |
| T-39-08 | Denial of Service | mitigate | CLOSED | `objectbox-model.json` `LocalSettingsEntity` carries all 7 properties (3 pre-existing + 4 new); additive migration confirmed by direct JSON inspection |
| T-39-09 | Spoofing | mitigate | CLOSED | `tile_proxy_server.dart:260-263` gates every route on `_constantTimeEquals`; `_constantTimeEquals` (line 745-751) has no early exit — XOR-accumulates every code unit |
| T-39-10 | Information Disclosure | mitigate | CLOSED* | `tile_proxy_provider.dart:12-13` doc comment states the URL "MUST NOT be logged or surfaced on any user-visible surface"; no production file (`main.dart`, `trail_map.dart`, `navigation_screen.dart`) logs `baseUrl`. *See residual note below: a dev-only spike harness does log it. |
| T-39-11 | Tampering / SSRF | mitigate | CLOSED | `tile_proxy_server.dart:686-704` `isSafeRedirectTarget` rejects non-http(s), empty host, `localhost`, loopback, link-local, AND (post-review fix) `anyIPv4`/`anyIPv6` (`0.0.0.0`/`::`); `buildUpstreamRedirect` (706-737) validates the substituted URI; 14 unit tests in `tile_proxy_redirect_test.dart` including the 3 new `0.0.0.0`/`::` cases, all passing (verified by running `flutter test`) |
| T-39-12 | Denial of Service | mitigate | CLOSED | Every retryable tile miss (`region == null`, `localFilePath == null`, `archive == null`, `TileNotFoundException`) funnels through `_redirectOrUnavailable` (lines 318-352); confirmed 8 total `HttpStatus.notFound` in the file, all on genuinely-permanent routes (malformed path/bad secret/non-whitelisted glyph-sprite name), 0 `HttpStatus.noContent` |
| T-39-13 | Information Disclosure | mitigate | CLOSED | `tile_proxy_server.dart:329-331` archive path read only from `region.demPackage.target?.localFilePath` / `region.vectorPackage.target?.localFilePath`; bind stays `InternetAddress.loopbackIPv4` (line 171, 194); no `segments[` joined into a filesystem path anywhere in the file |
| T-39-14 | Denial of Service | mitigate | CLOSED | `_demResolveInFlight` guard (line 109, set at 598, cleared at 649) allows one outstanding DEM TileJSON resolve at a time |
| T-39-15 | Information Disclosure | mitigate | CLOSED | `map_cache_path.dart` — 4-entry fontstack whitelist + anchored `^\d+-\d+$` range (`glyphCacheFilePath`), 8-entry filename whitelist (`spriteCacheFilePath`); both throw `ArgumentError` before any `p.join`; proxy catches and 404s (`tile_proxy_server.dart:404-409, 446-451`) without touching the filesystem |
| T-39-16 | Tampering | mitigate | CLOSED | `_serveCachedAsset` revalidates via `isSafeRedirectTarget` before fetch (line 506); `_downloadAndWriteThrough` writes to `<path>.part` then `rename`s atomically (lines 563-565), deletes partial on failure (569-572) |
| T-39-17 | Denial of Service | mitigate | CLOSED | `_inFlightAssetFetches` map (line 81) dedupes concurrent fetches by cache path, `_fetchAndCacheAsset` (528-537) |
| T-39-18 | Denial of Service | mitigate | CLOSED | `_serveCachedAsset` (485-522) never sets `HttpStatus.notFound`; every failure path routes through `_redirectOrUnavailable(request, null)` → 503 |
| T-39-19 | Repudiation / correctness | accept | CLOSED | Structural check confirmed: the `kind == 'vector' \| 'dem'` branch (lines 267-371) contains no `_serveCachedAsset`/`HttpClient` call — reverse-proxying confined to glyphs/sprites; accepted risk recorded below |
| T-39-20 | Tampering | mitigate | CLOSED | `proxy_style_rewriter.dart:101-107` `startsWith('http://127.0.0.1:')` guard throws `ArgumentError.value` byte-for-byte, unchanged since Plan 05 |
| T-39-21 | Information Disclosure | mitigate | CLOSED | `proxy_style_rewriter.dart:112-113, 124-133` — `glyphs`/`sprite` overwritten to loopback, `sourceMap.remove('url')`, `tiles` overwritten to loopback for every tiled source; no `https://` survives |
| T-39-22 | Information Disclosure | mitigate | CLOSED | `proxy_style_rewriter.dart` emits no `file://` literal anywhere (confirmed by full-file read); function takes no cache-root parameter |
| T-39-23 | Denial of Service | mitigate | CLOSED | `_proxyVectorMaxZoom = 14` (line 39), `_proxyDemMaxZoom = 12` (line 51), applied at lines 129, 132 unconditionally (online included) |
| T-39-24 | Tampering | mitigate | CLOSED | `grep -c "widget.offline\|offlineMapStyleJsonProvider\|bool offline" trail_map.dart` → 0; parameter does not exist, compile-time enforced |
| T-39-25 | Information Disclosure | mitigate | CLOSED | With no offline branch, every composed style routes through `rewriteStyleForProxy` unconditionally — confirmed via structural deletion (T-39-24) plus T-39-21's URL-stripping |
| T-39-26 | Denial of Service | mitigate | CLOSED | No `widget.offline` field exists to flip; single `mapStyleJsonProvider` watch/listen/read path (structurally verified) |
| T-39-27 | Tampering | mitigate | CLOSED | `router_provider.dart:436-444` — `extra is! (NavigateResponse, ActiveNavigationEntity?, geo.Position?)` guard falls back to `TrailDetailScreen(id: trailId)` on shape mismatch |
| T-39-28 | Tampering | mitigate | CLOSED | `grep -c isOffline active_navigation_entity.dart` → 0; field deleted, UID `5220757162905877938` confirmed present in `objectbox-model.json`'s `retiredPropertyUids` (42 entries total) |
| T-39-29 | Denial of Service | mitigate | CLOSED | `map_style_sources_provider.dart:29-34` — on fetch failure, falls back to `readPersistedMapStyleSources`; only rethrows when nothing is persisted |
| T-39-30 | Information Disclosure | mitigate | CLOSED | `navigation_screen.dart:1701` `if (!ref.watch(onlineStatusProvider)) ...[Icons.cloud_off...]` — confirmed live `ref.watch`, not a construction-time constant |
| T-39-31 | Tampering | mitigate | CLOSED | `proxy_style_rewriter.dart:10-24` doc comment records the no-caller rationale; loopback guard (T-39-20) and archive-path sourcing (T-39-13) independently confirmed intact |
| T-39-32 | Tampering | mitigate | CLOSED | `grep -rl "offlineMapStyleJson\|rewriteStyleForOffline" lib test` → no matches; both deleted outright, not merely unreferenced |
| T-39-33 | Denial of Service | mitigate | CLOSED | `flutter analyze` reports no issues (0 errors/warnings beyond 13 pre-existing vendor info-level items per `39-09-SUMMARY.md`, re-verified via targeted test runs); no dangling generated-output reference |
| T-39-34 | Denial of Service | eliminated | CLOSED | D-15 re-probe loop never implemented; no timer/periodic wakeup exists in the diff |
| T-39-36 | Denial of Service | eliminated | CLOSED | No automatic recovery trigger exists; `controller.setStyle(json)` calls in `trail_map.dart`/`navigation_screen.dart` fire only on genuine style-content change (`ref.listen(mapStyleJsonProvider,...)`), not on a connectivity signal |
| T-39-37 | Tampering | mitigate | CLOSED | `grep -c "onlineStatusProvider\|isBackendReachable\|isConnectionFailure" tile_proxy_server.dart` → 0 |
| T-39-38 | Repudiation | mitigate | CLOSED | `39-09-SUMMARY.md` `## Device verification` — all 4 device-observable ROADMAP criteria (1, 2, 5, 6) recorded **PASS** with the developer's own descriptions, on physical hardware |
| T-39-39 | Repudiation | mitigate | CLOSED | Evidence chain intact: `39-01-SUMMARY.md` `## Risk gate outcome` (full logcat evidence), `39-CONTEXT.md` D-12/D-12a/D-12b, `MainActivity.kt`'s rewritten comment cites both |
| T-39-SC | Tampering | accept | CLOSED | No `pubspec.yaml` changes in any of the 9 plans' `files_modified`; the phase's only new outbound HTTP use (`_resolveDemTemplate`) uses `dart:io`'s SDK `HttpClient`, not a new package |

## Residual note (not a blocker)

**T-39-10 — dev-only harness logs the secret-bearing base URL.**
`app/test/services/tile_proxy_spike_harness.dart:85` runs
`debugPrint('[spike] TileProxyServer.start -> ${proxy.baseUrl}')`. This file is a developer-only
on-device diagnostic harness launched via `flutter run -t test/services/tile_proxy_spike_harness.dart`
— it is not part of `main.dart`'s production entry point and is not shipped in a release build.
The threat model's trust boundary is "co-resident app -> loopback proxy" (an attacker process on
the same device); a developer intentionally logging their own harness's output to their own
`adb logcat` does not cross that boundary. Noted for completeness because the mitigation's stated
scope ("must not be surfaced in ... logs") is broader than what the code actually guarantees —
not escalated to OPEN because the boundary the threat actually protects is not crossed.

## Accepted Risks Log

| Threat ID | Risk | Rationale | Verified narrowing |
|---|---|---|---|
| T-39-07 | `tileUrl`'s `PROTOMAPS_API_KEY` query parameter persisted to app-private ObjectBox storage (`mapStyleSourcesJson`) | Same trust class as the value already held in memory in every composed style before this phase; app-private storage, not a new exposure surface | Verified: `proxy_style_rewriter.dart` strips the operator URL (including the key) from every style handed to MapLibre — the key now lives only in ObjectBox storage and in an outbound `Location` header, narrower than the pre-phase baseline |
| T-39-19 | Reverse-proxying glyph/sprite bytes introduces a thermal/battery cost the pre-phase design avoided entirely for tiles | Confined to a handful of requests per style load (not per pan frame); D-02's per-tile-volume constraint is unaffected | Verified structurally: the vector/dem tile path (`tile_proxy_server.dart:267-371`) contains no `_serveCachedAsset`/`HttpClient` call; device verification (`39-09-SUMMARY.md` criterion 6) confirms no panning thermal regression |
| T-39-SC | Standing risk: any future package-manager install in this phase would need a legitimacy audit | No install occurred in any of the 9 plans; `pubspec.yaml` never appears in any plan's `files_modified` | N/A — no dependency was added |

## Known limitation (accepted at plan time, not a threat-register gap)

`trail_panel.dart` mounts `TrailMap(disabled: true, embedded: true, ...)`; `disabled` resolves to
`MapGestures.none()`, so that specific embedded map cannot pan-recover (D-12a's user-initiated
recovery requires a pan/zoom to issue fresh tile requests). Recorded in `39-CONTEXT.md` D-12a and
`39-09-SUMMARY.md` `## Known limitations` as a deliberate, strictly-better-than-before acceptance,
not a threat-register entry.

## Unregistered Flags

None. Only `39-03-SUMMARY.md` declares a `## Threat Flags` section among all nine SUMMARY files
(checked via `awk` extraction across all nine), and it declares no new surface: "every new
surface ... is explicitly covered by this plan's own `<threat_model>` ... no additional surface
was introduced beyond what that register anticipated."

## Regression evidence

- `flutter test test/services/tile_proxy_redirect_test.dart test/services/tile_proxy_identity_test.dart test/util/region/map_cache_path_test.dart test/util/region/proxy_style_rewriter_test.dart` — 49/49 passed (re-run during this audit, not taken from SUMMARY narration).
- Post-review fix commit `e84ac158` confirmed present in `git log` — "fix(39): resolve all five code-review findings in the tile proxy", addressing the `_ArchiveCache` fd-leak (CR-01), FIFO-vs-LRU doc mismatch (WR-01), `0.0.0.0`/`::` redirect gap (WR-02), stream-error resilience (WR-03), and bind-retry exception scope (WR-04). All five fixes verified present in the current `tile_proxy_server.dart`.
- Residual, knowingly accepted per `39-REVIEW.md` Resolution: capacity eviction can still close a handle an in-flight request holds (degrades to a retryable HTTP 500, not data loss); the fd-leak fix has no direct regression test (`_ArchiveCache` is private, needs a live `HttpServer` + real `.pmtiles` archives to exercise). Neither maps to an open threat-register item — both are pre-existing-since-Phase-25.1 code paths made more visible by this phase, not introduced by it.

---
*Audited: 2026-09-10*
*Auditor: gsd-security-auditor*
