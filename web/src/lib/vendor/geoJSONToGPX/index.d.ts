import { Feature, FeatureCollection } from 'geojson';
declare interface Bounds {
    minlat: string;
    minlon: string;
    maxlat: string;
    maxlon: string;
}
declare interface Copyright {
    author?: string;
    year?: string;
    license?: string;
}
declare interface Link {
    href: string;
    text?: string;
    type?: string;
}
declare interface Person {
    name?: string;
    email?: string;
    link?: Link;
}
declare interface MetaData {
    name?: string;
    desc?: string;
    author?: Person;
    copyright?: Copyright;
    link?: Link;
    time?: string;
    keywords?: string;
    bounds?: Bounds;
}
export declare interface Options {
    creator?: string;
    version?: string;
    metadata?: MetaData;
}
export default function GeoJsonToGpx(geoJson: Feature | FeatureCollection, options?: Options, implementation?: DOMImplementation): XMLDocument;
export {};
