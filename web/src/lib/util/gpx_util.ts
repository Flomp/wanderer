import GPX from "$lib/models/gpx/gpx";
import { Trail } from "$lib/models/trail";
import { Waypoint } from "$lib/models/waypoint";
import GeoJsonToGpx from "$lib/vendor/geoJSONToGPX";
import { kml, tcx } from "$lib/vendor/toGeoJSON/toGeoJSON";
import cryptoRandomString from "crypto-random-string";

export async function gpx2trail(gpxString: string) {
    const gpx = await GPX.parse(gpxString);

    if (gpx instanceof Error) {
        throw gpx;
    }

    const trail = new Trail("");

    trail.name = gpx.metadata?.name ?? "";

    trail.description = gpx.metadata?.desc;

    for (const wpt of gpx.wpt ?? []) {               
        const wp = new Waypoint(wpt.$.lat ?? 0, wpt.$.lon ?? 0);
        wp.id = cryptoRandomString({ length: 15 });
        wp.name = wpt.name
        wp.description = wpt.desc;
        trail.expand.waypoints.push(wp);
    }

    const totals = gpx.getTotals()

    const points = gpx.trk?.at(0)?.trkseg?.at(0)?.trkpt
    const startPoint = points?.at(0);
    if (startPoint) {
        trail.lat = startPoint.$.lat
        trail.lon = startPoint.$.lon
    }

    const startTime = points?.at(0)?.time;
    const endTime = points?.at((points?.length ?? 1) - 1)?.time

    if (startTime && endTime && !trail.date) {
        trail.date = startTime.toISOString()
            .substring(0, 10);
    }

    trail.duration = totals.duration / 1000 / 60
    trail.elevation_gain = totals.elevationGain;
    trail.distance = totals.distance

    return {gpx: gpx, trail: trail}
}

export function fromKML(kmlData: string) {
    const geojson = kml(new DOMParser().parseFromString(kmlData, "text/xml"))
    const options = {
        metadata: {
            name: '',
            author: {
                name: '',
                link: {
                    href: ''
                }
            }
        }
    }
    const gpx = GeoJsonToGpx(geojson as any, options)

    return new XMLSerializer().serializeToString(gpx)
}

export function fromTCX(tcxData: string) {
    const geojson = tcx(new DOMParser().parseFromString(tcxData, "text/xml"))
    const options = {
        metadata: {
            name: '',
            author: {
                name: '',
                link: {
                    href: ''
                }
            }
        }
    }
    const gpx = GeoJsonToGpx(geojson as any, options)
    return new XMLSerializer().serializeToString(gpx)
}