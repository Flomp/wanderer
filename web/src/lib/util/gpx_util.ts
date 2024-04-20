import { Trail } from "$lib/models/trail";
import { Waypoint } from "$lib/models/waypoint";
import { kml, tcx } from "$lib/vendor/toGeoJSON/toGeoJSON"
import GeoJsonToGpx from "$lib/vendor/geoJSONToGPX"
import { browser } from "$app/environment";

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

export async function gpx2trail(gpx: string) {
    let xml;
    if (browser) {       
        const parser = new DOMParser();
        xml = parser.parseFromString(gpx, "application/xml")
    } else {        
        const JSDOM = (await import("jsdom")).JSDOM
        xml = new JSDOM(gpx).window.document
    }

    const trail = new Trail("");

    let name = xml.getElementsByTagName('name');
    if (name.length > 0) {
        trail.name = name[0].textContent ?? "";
    }
    let desc = xml.getElementsByTagName('desc');
    if (desc.length > 0) {
        trail.description = desc[0].textContent ?? "";
    }

    const el = xml.getElementsByTagName('wpt');
    const elLength = el.length
    for (let i = 0; i < elLength; i++) {
        const wp = new Waypoint(parseFloat(el[i].getAttribute('lat')!), parseFloat(el[i].getAttribute('lon')!));

        let nameEl = el[i].getElementsByTagName('name');
        wp.name = nameEl.length > 0 ? nameEl[0].textContent ?? "" : '';

        let descEl = el[i].getElementsByTagName('desc');
        wp.description = descEl.length > 0 ? descEl[0].textContent ?? "" : '';

        trail.expand.waypoints.push(wp);
    }

    let totalElevationGain = 0;
    let totalDuration = 0;
    let totalDistance = 0;

    const tracks = xml.getElementsByTagName("trk")
    for (const track of tracks) {
        const segments = track.getElementsByTagName("trkseg")
        for (const segment of segments) {
            const points = segment.getElementsByTagName("trkpt")

            if (points.length >= 2) {
                const startTimeString = points[0].querySelector('time')?.textContent;
                const startTime = startTimeString ? new Date(startTimeString) : undefined;
                const endTimeString = points[points.length - 1].querySelector('time')?.textContent;
                const endTime = endTimeString ? new Date(endTimeString) : undefined;

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
                const elevation = parseFloat(points[i].querySelector('ele')?.textContent || '0')
                const previousElevation = parseFloat(points[i - 1].querySelector('ele')?.textContent || '0')
                const elevationDiff = elevation - previousElevation;
                if (elevationDiff > 0) {
                    totalElevationGain += elevationDiff;
                }

                const prevPoint = points[i - 1];
                const point = points[i];
                const distance = calculateDistance(
                    parseFloat(prevPoint.getAttribute("lat") || '0'),
                    parseFloat(prevPoint.getAttribute("lon") || '0'),
                    parseFloat(point.getAttribute("lat") || '0'),
                    parseFloat(point.getAttribute("lon") || '0')
                );
                totalDistance += distance;
            }
        }
    }

    const startPoint = xml.getElementsByTagName("trk")[0].getElementsByTagName("trkseg")[0].getElementsByTagName("trkpt")[0];
    if (startPoint) {
        trail.lat = parseFloat(startPoint.getAttribute('lat')!)
        trail.lon = parseFloat(startPoint.getAttribute('lon')!)
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