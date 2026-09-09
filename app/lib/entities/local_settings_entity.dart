import 'package:objectbox/objectbox.dart';

@Entity()
class LocalSettingsEntity {
  @Id()
  int obxId = 0;

  String themeMode;

  /// Whether the background-location disclosure has been shown once.
  ///
  /// Android 11+ cannot grant ACCESS_BACKGROUND_LOCATION from a runtime
  /// prompt — "Allow all the time" exists only in system settings — so the
  /// disclosure can never be answered in place. Without this flag it would
  /// reappear on every recording for anyone who has not gone to settings.
  bool backgroundLocationAsked;

  /// D-05: the loopback tile proxy's persisted port. `0` means "not yet
  /// minted". MapLibre's ambient cache keys on `resource.url` — the loopback
  /// URL — so a port that changes every launch orphans the whole tile cache
  /// on every cold start. Persisted (rather than recomputed) so the port
  /// stays stable across launches while remaining random per install.
  int tileProxyPort;

  /// D-06: the loopback tile proxy's per-install secret path segment.
  /// Empty string means "not yet minted". A per-launch secret would defeat
  /// D-05 by churning the cache key exactly as an ephemeral port does, so
  /// this is minted once and persisted alongside the port.
  String tileProxySecret;

  /// D-09: `jsonEncode` of the last successful `/map/style-sources`
  /// response. Empty string means "never fetched". Persisted so the proxy
  /// can build an upstream redirect target on an offline cold start, when no
  /// network call is available to re-fetch it.
  String mapStyleSourcesJson;

  /// D-09: the resolved hillshade DEM XYZ template. Empty string means
  /// "not yet resolved". Persisted for the same offline-cold-start reason as
  /// [mapStyleSourcesJson]; written by Plan 03.
  String demTileTemplate;

  LocalSettingsEntity({
    this.themeMode = 'system',
    this.backgroundLocationAsked = false,
    this.tileProxyPort = 0,
    this.tileProxySecret = '',
    this.mapStyleSourcesJson = '',
    this.demTileTemplate = '',
  });
}
