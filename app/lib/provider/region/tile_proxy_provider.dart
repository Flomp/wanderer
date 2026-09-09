import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'tile_proxy_provider.g.dart';

/// Exposes the resolved loopback base URL (e.g.
/// `http://127.0.0.1:54321/<32-hex-secret>`) of the app-wide
/// [TileProxyServer] — mirrors `objectbox_store_provider.dart`'s exact
/// `keepAlive` "overridden in main.dart" shape. The port and secret are
/// persisted (D-05/D-06) and stable **across launches**, not merely for the
/// process lifetime, so there is no `update*` method here, unlike
/// `api_provider.dart`'s mutable base URL. The URL carries the per-install
/// secret path segment and MUST NOT be logged or surfaced on any
/// user-visible surface.
@Riverpod(keepAlive: true)
class TileProxyBaseUrl extends _$TileProxyBaseUrl {
  @override
  String build() {
    // This will be overridden in main.dart
    throw UnimplementedError();
  }
}
