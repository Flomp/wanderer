import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:wanderer/models/map_style_sources.dart';
import 'package:wanderer/provider/api_provider.dart';
import 'package:wanderer/provider/objectbox_store_provider.dart';
import 'package:wanderer/services/map_source_persistence.dart';

part 'map_style_sources_provider.g.dart';

/// D-09: the proxy needs the operator's upstream tile/glyph/sprite templates
/// to build redirect targets, and an offline cold start must know them with
/// no network call available. This notifier fetches `/map/style-sources` as
/// its primary path and writes the result through to app-private ObjectBox
/// storage on success — the same trust class as the value the app already
/// embeds in every composed style today, just persisted rather than
/// recomputed. On a failed fetch it falls back to the persisted copy; a
/// first-ever run with no network and no persisted copy has genuinely no
/// sources, so the original error is rethrown rather than fabricating one.
@Riverpod(keepAlive: true)
class MapStyleSourcesNotifier extends _$MapStyleSourcesNotifier {
  @override
  Future<MapStyleSources> build() async {
    final api = ref.watch(apiProvider);
    try {
      final response = await api.get('/map/style-sources');
      final sources = MapStyleSources.fromJson(response.data);
      final store = ref.read(objectBoxProvider);
      writePersistedMapStyleSources(store, sources);
      return sources;
    } catch (error) {
      final store = ref.read(objectBoxProvider);
      final persisted = readPersistedMapStyleSources(store);
      if (persisted != null) return persisted;
      rethrow;
    }
  }
}
