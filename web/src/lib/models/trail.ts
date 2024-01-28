import type { Category } from "./category";
import type { SummitLog } from "./summit_log";
import type { Waypoint } from "./waypoint";

class Trail {
    id?: string;
    name: string;
    location?: string;
    distance?: number;
    elevation_gain?: number;
    duration?: number;
    thumbnail?: string;
    photos: string[];
    gpx?: string;
    expand: {
        category?: Category;
        waypoints: Waypoint[]
        summit_logs: SummitLog[]
    }
    tags?: string[];
    description?: string;

    constructor(name: string,
        id?: string,
        location?: string,
        distance?: number,
        elevation_gain?: number,
        duration?: number,
        thumbnail?: string,
        photos?: string[],
        gpx?: string,
        category?: Category,
        waypoints?: Waypoint[],
        summit_logs?: SummitLog[],
        tags?: string[],
        description?: string
        ) {
        this.id = id;
        this.name = name;
        this.location = location;
        this.distance = distance;
        this.elevation_gain = elevation_gain;
        this.duration = duration;
        this.thumbnail = thumbnail;
        this.photos = photos ?? [];
        this.gpx = gpx;
        this.expand = {
            category: category,
            waypoints: waypoints ?? [],
            summit_logs: summit_logs ??  []
        }
        this.tags = tags ?? []
        this.description = description ?? "";
    }
}

export { Trail };
