import * as M from "maplibre-gl";

class Waypoint {
    id?: string;
    name?: string;
    description?: string;
    lat: number;
    lon: number;
    icon?: string;
    marker?: M.Marker;
    photos: string[];
    _photos?: File[];
    author: string;

    constructor(lat: number, lon: number, params?: {
        id?: string, name?: string, description?: string, icon?: string, marker?: M.Marker, photos?: string[];
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
        this.author = "000000000000000"
    }
}

export { Waypoint };
