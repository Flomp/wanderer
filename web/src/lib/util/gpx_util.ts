import { Trail } from "$lib/models/trail";
import { Waypoint } from "$lib/models/waypoint";
import { kml, tcx } from "$lib/vendor/toGeoJSON/toGeoJSON"
import GeoJsonToGpx from "$lib/vendor/geoJSONToGPX"

export async function gpx2trail(gpx: string) {
    const JSDOM = (await import("jsdom")).JSDOM
    const xml = new JSDOM(gpx).window.document
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
    for (let i = 0; i < el.length; i++) {
        const wp = new Waypoint(parseFloat(el[i].getAttribute('lat')!), parseFloat(el[i].getAttribute('lon')!));

        let nameEl = el[i].getElementsByTagName('name');
        wp.name = nameEl.length > 0 ? nameEl[0].textContent ?? "" : '';

        let descEl = el[i].getElementsByTagName('desc');
        wp.description = descEl.length > 0 ? descEl[0].textContent ?? "" : '';

        trail.expand.waypoints.push(wp);
    }

    const startPoint = xml.getElementsByTagName("trk")[0].getElementsByTagName("trkseg")[0].getElementsByTagName("trkpt")[0];
    if (startPoint) {
        trail.lat = parseFloat(startPoint.getAttribute('lat')!)
        trail.lon = parseFloat(startPoint.getAttribute('lon')!)
    }

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