import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// Method-channel name shared verbatim with `MainActivity.kt`'s
/// `configureFlutterEngine` handler. Keep the literal identical on both
/// sides -- there is no compile-time link between a Dart string and a
/// Kotlin `private const val`.
// dart format off
const String kMapLibreConnectivityChannel = 'com.openwanderer.wanderer/maplibre_connectivity';
// dart format on

/// Method name invoked on [kMapLibreConnectivityChannel] to trigger the
/// pulse. See [pulseMapLibreConnectivity].
const String kPulseConnectedMethod = 'pulseConnected';

const MethodChannel _channel = MethodChannel(kMapLibreConnectivityChannel);

/// Forces MapLibre Native on Android to re-`schedule()` every tile request
/// whose last failure was `Reason::Connection`.
///
/// Why this exists: `MainActivity.kt` permanently pins
/// `MapLibre.setConnected(true)` so the in-app loopback tile proxy stays
/// reachable even in airplane mode (every style is now always proxied --
/// there is no separate offline style). But MapLibre Native's own
/// connectivity retry only fires on a genuine `false` -> `true` edge
/// (`NetworkStatus::Set` -> `Reachable()` ->
/// `OnlineFileRequest::networkIsReachableAgain()`), and the app can never
/// rely on the system connectivity broadcast to produce that edge --
/// `ConnectivityReceiver.onReceive` early-returns for as long as the
/// override is non-null. So when the app itself detects connectivity has
/// returned, it must drive that edge deliberately: this function pulses the
/// override `false` then immediately back to `true`, on the same call, with
/// no delay in between.
///
/// If on-device testing ever finds that the brief `false` window drops
/// in-flight loopback requests, the documented fallback is a full `setStyle`
/// reload on the mounted map controller (heavier -- rebuilds every layer and
/// image) rather than this targeted pulse.
///
/// Android-only **by construction**, not by runtime configuration (D-13):
/// iOS's `MLNReachability` already calls `Reachable()` on regain, and no
/// Apple platform file ever calls `NetworkStatus::Set(Offline)` in the first
/// place, so iOS never suppresses the loopback proxy and needs no pulse. Do
/// not "fix" this asymmetry into symmetry -- it is deliberate and correct.
Future<void> pulseMapLibreConnectivity() async {
  if (!Platform.isAndroid) return;
  try {
    await _channel.invokeMethod<void>(kPulseConnectedMethod);
  } on PlatformException catch (e) {
    debugPrint('[maplibre_connectivity_pulse] pulse failed: $e');
  } on MissingPluginException catch (e) {
    debugPrint('[maplibre_connectivity_pulse] channel not registered: $e');
  }
}
