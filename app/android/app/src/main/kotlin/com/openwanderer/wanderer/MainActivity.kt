package com.openwanderer.wanderer

import android.os.Bundle
import io.flutter.embedding.android.FlutterActivity
import org.maplibre.android.MapLibre

class MainActivity : FlutterActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // MapLibre Native suppresses ALL online-file-source HTTP requests when
        // its ConnectivityReceiver reports no network (e.g. airplane mode) --
        // including requests to our in-app loopback tile proxy
        // (http://127.0.0.1). Every style is now always proxied through it
        // (there is no separate "offline style" that carries no online URL),
        // so this override must stay pinned `true` or the proxy itself becomes
        // unreachable the moment the radio reports no network.
        //
        // Consequence, deliberately accepted: because the pin is permanent,
        // MapLibre never sees the false->true connectivity edge that
        // OnlineFileRequest::networkIsReachableAgain() needs, so it never
        // retries a Connection-failed tile on its own. Recovery after service
        // returns is therefore user-initiated -- a pan or zoom requests tiles
        // at new coordinates, which are fresh requests and succeed once the
        // radio is back. See CONTEXT.md decision D-12a in
        // .planning/phases/39-unified-tile-model/. An app-driven alternative
        // was built, tested on a physical device and rejected; the evidence is
        // in 39-01-SUMMARY.md under "Risk gate outcome".
        MapLibre.getInstance(applicationContext)
        MapLibre.setConnected(true)
    }
}
