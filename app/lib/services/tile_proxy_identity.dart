/// Mint / read / persist the loopback tile proxy's stable-but-random port
/// (D-05) and per-install secret path segment (D-06).
///
/// Split into a pure half (minting/validation, no I/O) and a store-backed
/// half (read-or-mint-and-persist), mirroring
/// `util/region/map_cache_path.dart`'s validate-before-use shape so the pure
/// half stays unit-testable without a live ObjectBox [Store].
///
/// `TileProxyServer.start` runs in `main()` before `ProviderScope` exists, so
/// this module takes a raw [Store] rather than a Riverpod provider.
library;

import 'dart:math';

import 'package:wanderer/entities/local_settings_entity.dart';
import 'package:wanderer/objectbox.g.dart';

/// Lower bound of the IANA dynamic/private port range. Deliberately the same
/// range an ephemeral bind (`HttpServer.bind(..., 0)`) would have drawn from
/// — persisting a value inside this range is no more guessable than today's
/// ephemeral-port behavior, it is simply drawn once and remembered instead
/// of redrawn every launch (D-05).
const int kTileProxyPortMin = 49152;

/// Upper bound of the IANA dynamic/private port range (see
/// [kTileProxyPortMin]).
const int kTileProxyPortMax = 65535;

/// A stable loopback tile proxy identity: the port the server binds to and
/// the per-install secret required as the first path segment of every
/// request.
class TileProxyIdentity {
  const TileProxyIdentity({required this.port, required this.secret});

  final int port;
  final String secret;

  @override
  bool operator ==(Object other) =>
      other is TileProxyIdentity &&
      other.port == port &&
      other.secret == secret;

  @override
  int get hashCode => Object.hash(port, secret);
}

/// Pattern a persisted/minted secret must satisfy: exactly 32 lowercase hex
/// characters, anchored at both ends. Anchoring at both ends is the point —
/// it is what stops a corrupted or truncated persisted value (or one
/// containing `/` or `..`) from ever becoming a live route prefix.
final RegExp _secretPattern = RegExp(r'^[0-9a-f]{32}$');

/// Whether [secret] is a valid 32-lowercase-hex-character proxy secret. This
/// is the gate the proxy uses before accepting a persisted secret and before
/// comparing a request's first path segment against it.
bool isValidProxySecret(String secret) => _secretPattern.hasMatch(secret);

/// Mints a fresh [TileProxyIdentity]: a random port in
/// `[kTileProxyPortMin, kTileProxyPortMax]` and a random 128-bit secret
/// rendered as 32 lowercase hex characters.
///
/// [random] defaults to `Random.secure()`, which is MANDATORY in
/// production — the port and secret must be unguessable to a co-resident
/// app (D-05, D-06). A non-secure `Random` (e.g. a seeded `Random(1234)`) is
/// accepted ONLY so tests can assert deterministic output; never pass one
/// outside a test.
TileProxyIdentity mintTileProxyIdentity({Random? random}) {
  final rng = random ?? Random.secure();
  final port =
      kTileProxyPortMin + rng.nextInt(kTileProxyPortMax - kTileProxyPortMin + 1);
  final buffer = StringBuffer();
  for (var i = 0; i < 16; i++) {
    buffer.write(rng.nextInt(256).toRadixString(16).padLeft(2, '0'));
  }
  return TileProxyIdentity(port: port, secret: buffer.toString());
}

/// Reads the persisted proxy identity from [store], minting and persisting a
/// fresh one if the persisted row is missing or invalid.
///
/// A valid persisted identity (port inside range AND secret matching
/// [isValidProxySecret]) is returned UNCHANGED — this is the D-05/D-06
/// stability guarantee that keeps MapLibre's ambient tile cache alive across
/// cold starts. An invalid row is never partially repaired: port and secret
/// are always re-minted together, since a half-valid row is replaced
/// wholesale.
TileProxyIdentity resolveTileProxyIdentity(Store store) {
  final box = store.box<LocalSettingsEntity>();
  final entity = box.getAll().firstOrNull ?? LocalSettingsEntity();

  final persistedPort = entity.tileProxyPort;
  final persistedSecret = entity.tileProxySecret;
  if (persistedPort >= kTileProxyPortMin &&
      persistedPort <= kTileProxyPortMax &&
      isValidProxySecret(persistedSecret)) {
    return TileProxyIdentity(port: persistedPort, secret: persistedSecret);
  }

  final minted = mintTileProxyIdentity();
  entity.tileProxyPort = minted.port;
  entity.tileProxySecret = minted.secret;
  box.put(entity);
  return minted;
}

/// Overwrites only the persisted port, leaving the secret untouched.
///
/// Used by the bind-retry path (Plan 03) when the persisted port turns out
/// to already be occupied at bind time. The secret must survive a rebind
/// unchanged — re-minting it would churn the glyph/sprite cache keys for no
/// reason, the same problem D-06 exists to avoid.
void persistTileProxyPort(Store store, int port) {
  final box = store.box<LocalSettingsEntity>();
  final entity = box.getAll().firstOrNull ?? LocalSettingsEntity();
  entity.tileProxyPort = port;
  box.put(entity);
}
