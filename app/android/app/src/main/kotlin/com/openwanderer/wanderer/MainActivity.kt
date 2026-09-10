package com.openwanderer.wanderer

import android.os.Bundle
import io.flutter.embedding.android.FlutterActivity
import org.maplibre.android.MapLibre

class MainActivity : FlutterActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // MapLibre suppresses every online-file-source request when its
        // ConnectivityReceiver reports no network -- including requests to our
        // loopback tile proxy. Every style routes through that proxy, so this
        // must stay pinned `true` or the proxy becomes unreachable the moment
        // the radio drops.
        //
        // Accepted consequence: a permanent pin means MapLibre never sees the
        // false->true edge networkIsReachableAgain() needs, so it never
        // retries a Connection-failed tile on its own. Recovery is
        // user-initiated -- a pan or zoom issues fresh requests at new
        // coordinates. Driving that edge deliberately was tried on a device
        // and does not work: MapLibre re-schedules nothing.
        MapLibre.getInstance(applicationContext)
        MapLibre.setConnected(true)
    }
}
