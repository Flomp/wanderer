import type { Marker } from "leaflet";
import { number, object, string } from "yup";

class Waypoint {
    id?: string;
    name?: string;
    description?: string;
    lat: number;
    lon: number;
    icon?: string;
    marker?: Marker;
    photos: string[];
    _photos: File[];
    author?: string;

    constructor(lat: number, lon: number, params?: {
        id?: string, name?: string, description?: string, icon?: string, marker?: Marker, photos?: string[];
    }) {
        this.id = params?.id;
        this.name = params?.name ?? "";
        this.description = params?.description ?? "";
        this.lat = lat;
        this.lon = lon;
        this.icon = params?.icon ?? "circle";
        this.marker = params?.marker;
        this.photos = params?.photos ?? []
        this._photos = [];
    }
}

const waypointSchema = object<Waypoint>({
    id: string().optional(),
    name: string().optional(),
    description: string().optional(),
    lat: number().required('Required').typeError('Invalid latitude'),
    lon: number().required('Required').typeError('Invalid longitude'),
    icon: string().optional()
});

export { Waypoint, waypointSchema };
