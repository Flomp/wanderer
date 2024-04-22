import GPX from "$lib/models/gpx/gpx";
import { Trail } from "$lib/models/trail";
import { Waypoint } from "$lib/models/waypoint";
import GeoJsonToGpx from "$lib/vendor/geoJSONToGPX";
import { kml, tcx } from "$lib/vendor/toGeoJSON/toGeoJSON";

function calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    const R = 6371; // Radius of the Earth in km
    const dLat = (lat2 - lat1) * (Math.PI / 180); // Convert degrees to radians
    const dLon = (lon2 - lon1) * (Math.PI / 180);
    const a =
        Math.sin(dLat / 2) * Math.sin(dLat / 2) +
        Math.cos((lat1 * (Math.PI / 180))) * Math.cos((lat2 * (Math.PI / 180))) *
        Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    const distance = R * c * 1000; // Distance in km
    return distance;
}

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
        wp.name = wp.name
        wp.description = wp.description;
        trail.expand.waypoints.push(wp);
    }

    let totalElevationGain = 0;
    let totalDuration = 0;
    let totalDistance = 0;

    for (const track of gpx.trk ?? []) {
        for (const segment of track.trkseg ?? []) {
            const points = segment.trkpt ?? [];

            if (points.length >= 2) {
                const startTime = points[0].time;
                const endTime = points[points.length - 1].time

                if (startTime && endTime) {
                    totalDuration += endTime.getTime() - startTime.getTime();
                    if (!trail.date) {
                        trail.date = startTime.toISOString()
                            .substring(0, 10);
                    }
                }
            }

            const pointLength = points.length
            for (let i = 1; i < pointLength; i++) {
                const prevPoint = points[i - 1];
                const point = points[i];
                const elevation = point.ele ?? 0
                const previousElevation = prevPoint.ele ?? 0
                const elevationDiff = elevation - previousElevation;
                if (elevationDiff > 0) {
                    totalElevationGain += elevationDiff;
                }

                const distance = calculateDistance(
                    prevPoint.$.lat ?? 0,
                    prevPoint.$.lon ?? 0,
                    point.$.lat ?? 0,
                    point.$.lon ?? 0,
                );
                totalDistance += distance;
            }
        }
    }

    const startPoint = gpx.trk?.at(0)?.trkseg?.at(0)?.trkpt?.at(0);
    if (startPoint) {
        trail.lat = startPoint.$.lat
        trail.lon = startPoint.$.lon
    }

    trail.duration = totalDuration / 1000 / 60
    trail.elevation_gain = totalElevationGain;
    trail.distance = totalDistance

    return trail
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