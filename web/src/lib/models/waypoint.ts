import type { Marker } from "leaflet";
import { date, number, object, string } from "yup";

class Waypoint {
    id?: string;
    name?: string;
    description?: string;
    lat: number;
    lon: number;
    icon?: string;
    marker?: Marker;

    constructor(lat: number, lon: number, params?: {id?: string, name?: string, description?: string, icon?: string, marker?: Marker}) {
        this.id = params?.id;
        this.name = params?.name ?? "";
        this.description = params?.description ?? "";
        this.lat = lat;
        this.lon = lon;
        this.icon = params?.icon ?? "circle";
        this.marker = params?.marker;
    }
}

const waypointSchema = object<Waypoint>({
    id: string().optional(),
    name: string().optional(),
    description: string().optional(),
    lat: number().min(0).required('Required').typeError('Invalid latitude'),
    lon: number().min(0).required('Required').typeError('Invalid longitude'),
    icon: string().optional()
  });

export { Waypoint, waypointSchema }