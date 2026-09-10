// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'map_style_sources_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning
/// D-09: the proxy needs the operator's upstream tile/glyph/sprite templates
/// to build redirect targets, and an offline cold start must know them with
/// no network call available. This notifier fetches `/map/style-sources` as
/// its primary path and writes the result through to app-private ObjectBox
/// storage on success — the same trust class as the value the app already
/// embeds in every composed style today, just persisted rather than
/// recomputed. On a failed fetch it falls back to the persisted copy; a
/// first-ever run with no network and no persisted copy has genuinely no
/// sources, so the original error is rethrown rather than fabricating one.

@ProviderFor(MapStyleSourcesNotifier)
final mapStyleSourcesProvider = MapStyleSourcesNotifierProvider._();

/// D-09: the proxy needs the operator's upstream tile/glyph/sprite templates
/// to build redirect targets, and an offline cold start must know them with
/// no network call available. This notifier fetches `/map/style-sources` as
/// its primary path and writes the result through to app-private ObjectBox
/// storage on success — the same trust class as the value the app already
/// embeds in every composed style today, just persisted rather than
/// recomputed. On a failed fetch it falls back to the persisted copy; a
/// first-ever run with no network and no persisted copy has genuinely no
/// sources, so the original error is rethrown rather than fabricating one.
final class MapStyleSourcesNotifierProvider
    extends $AsyncNotifierProvider<MapStyleSourcesNotifier, MapStyleSources> {
  /// D-09: the proxy needs the operator's upstream tile/glyph/sprite templates
  /// to build redirect targets, and an offline cold start must know them with
  /// no network call available. This notifier fetches `/map/style-sources` as
  /// its primary path and writes the result through to app-private ObjectBox
  /// storage on success — the same trust class as the value the app already
  /// embeds in every composed style today, just persisted rather than
  /// recomputed. On a failed fetch it falls back to the persisted copy; a
  /// first-ever run with no network and no persisted copy has genuinely no
  /// sources, so the original error is rethrown rather than fabricating one.
  MapStyleSourcesNotifierProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'mapStyleSourcesProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$mapStyleSourcesNotifierHash();

  @$internal
  @override
  MapStyleSourcesNotifier create() => MapStyleSourcesNotifier();
}

String _$mapStyleSourcesNotifierHash() =>
    r'a81959bf89f1cc3daf318d83bdf0170295209489';

/// D-09: the proxy needs the operator's upstream tile/glyph/sprite templates
/// to build redirect targets, and an offline cold start must know them with
/// no network call available. This notifier fetches `/map/style-sources` as
/// its primary path and writes the result through to app-private ObjectBox
/// storage on success — the same trust class as the value the app already
/// embeds in every composed style today, just persisted rather than
/// recomputed. On a failed fetch it falls back to the persisted copy; a
/// first-ever run with no network and no persisted copy has genuinely no
/// sources, so the original error is rethrown rather than fabricating one.

abstract class _$MapStyleSourcesNotifier
    extends $AsyncNotifier<MapStyleSources> {
  FutureOr<MapStyleSources> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<MapStyleSources>, MapStyleSources>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<MapStyleSources>, MapStyleSources>,
              AsyncValue<MapStyleSources>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
