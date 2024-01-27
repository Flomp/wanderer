import type { Category } from "./category";
import type { SummitLog } from "./summit_log";
import type { Waypoint } from "./waypoint";

interface Trail {
    id: string;
    name: string;
    location: string;
    distance: number;
    elevation_gain: number;
    duration: number;
    thumbnail: string;
    photos: string[];
    gpx: string;
    expand: {
        category: Category;
        waypoints: Waypoint[]
        summit_logs: SummitLog[]
    }
    tags?: string[];
    description: string;
}

export type { Trail }