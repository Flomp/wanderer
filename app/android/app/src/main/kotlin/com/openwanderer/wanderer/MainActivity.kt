package com.openwanderer.wanderer

import android.os.Bundle
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import org.maplibre.android.MapLibre

class MainActivity : FlutterActivity() {
    companion object {
        private const val CHANNEL = "com.openwanderer.wanderer/maplibre_connectivity"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // MapLibre Native suppresses ALL online-file-source HTTP requests when
        // its ConnectivityReceiver reports no network (e.g. airplane mode) --
        // including requests to our in-app loopback tile proxy
        // (http://127.0.0.1). Every style is now always proxied through it
        // (there is no separate "offline style" that carries no online URL),
        // so this override must stay pinned `true` or the proxy itself becomes
        // unreachable the moment the radio reports no network. Because the
        // pin is permanent, MapLibre's own false->true connectivity edge --
        // the only thing that makes OnlineFileRequest::networkIsReachableAgain()
        // retry a Connection-failed request -- never happens on its own. We
        // drive it deliberately: when connectivity returns, the Dart side
        // calls the "pulseConnected" method below, which pulses the override
        // false then immediately back to true.
        MapLibre.getInstance(applicationContext)
        MapLibre.setConnected(true)
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "pulseConnected" -> {
                        MapLibre.setConnected(false)
                        MapLibre.setConnected(true)
                        result.success(null)
                    }
                    else -> result.notImplemented()
                }
            }
    }
}
