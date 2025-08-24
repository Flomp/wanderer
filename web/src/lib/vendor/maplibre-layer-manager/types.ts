export interface OverpassResponse {
    elements: OverpassElement[];
    generator: string;
    osm3s: {
        copyright: string;
        timestamp_osm_base: string;
    };
    version: number;
}
export interface OverpassElement {
    id: number;
    center?: {
        lat: number,
        lon: number,
    }
    lat: number;
    lon: number;
    nodes?: number[];
    tags?: Record<string, string>;
    type: string;
}