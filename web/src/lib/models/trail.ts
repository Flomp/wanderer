import type { Category } from "./category";
import type { Comment } from "./comment";
import type { SummitLog } from "./summit_log";
import type { TrailShare } from "./trail_share";
import type { Waypoint } from "./waypoint";

class Trail {
    id?: string;
    name: string;
    location?: string;
    date?: string;
    public: boolean;
    distance?: number;
    elevation_gain?: number;
    duration?: number;
    difficulty?: "easy" | "moderate" | "difficult"
    lat?: number;
    lon?: number;
    thumbnail: number;
    photos: string[];
    gpx?: string;
    created?: string;
    category?: string;
    waypoints: string[];
    summit_logs: string[];
    expand: {
        category?: Category;
        waypoints: Waypoint[]
        summit_logs: SummitLog[]
        comments_via_trail: Comment[]
        gpx_data?: string
        trail_share_via_trail?: TrailShare[]
    }
    tags?: string[];
    description?: string;
    author?: string;

    constructor(name: string,
        params?: {
            id?: string,
            location?: string,
            date?: string,
            public?: boolean,
            distance?: number,
            elevation_gain?: number,
            duration?: number,
            difficulty?: "easy" | "moderate" | "difficult",
            lat?: number,
            lon?: number,
            thumbnail?: number,
            photos?: string[],
            gpx?: string,
            category?: Category,
            waypoints?: Waypoint[],
            summit_logs?: SummitLog[],
            comments?: Comment[],
            shares?: TrailShare[],
            tags?: string[],
            description?: string
            created?: string
        }

    ) {
        this.id = params?.id;
        this.name = name;
        this.location = params?.location;
        this.date = params?.date ?? new Date().toISOString().split('T')[0];
        this.public = params?.public ?? false
        this.distance = params?.distance;
        this.elevation_gain = params?.elevation_gain;
        this.duration = params?.duration;
        this.difficulty = params?.difficulty ?? "easy";
        this.lat = params?.lat;
        this.lon = params?.lon;
        this.thumbnail = params?.thumbnail ?? 0;
        this.photos = params?.photos ?? [];
        this.waypoints = [];
        this.summit_logs = [];
        this.gpx = params?.gpx;
        this.expand = {
            category: params?.category,
            waypoints: params?.waypoints ?? [],
            summit_logs: params?.summit_logs ?? [],
            comments_via_trail: params?.comments ?? [],
            trail_share_via_trail: params?.shares ?? []
        }
        this.tags = params?.tags ?? []
        this.description = params?.description ?? "";
        this.created = params?.created;
    }
}

interface TrailFilter {
    q: string,
    category: string[],
    difficulty: ("easy" | "moderate" | "difficult")[]
    near: {
        lat?: number,
        lon?: number,
        radius: number
    }
    distanceMin: number,
    distanceMax: number,
    distanceLimit: number,
    elevationGainMin: number;
    elevationGainMax: number;
    elevationGainLimit: number;
    startDate?: string;
    endDate?: string;
    completed?: boolean;
    sort: "name" | "distance" | "elevation_gain" | "created";
    sortOrder: "+" | "-"
}

interface TrailFilterValues {
    min_distance: number,
    max_distance: number,
    min_elevation_gain: number,
    max_elevation_gain: number,
    min_durtation: number,
    max_duration: number,
}

interface TrailBoundingBox {
    max_lat: number,
    min_lat: number,
    max_lon: number,
    min_lon: number,
}

export { Trail };

export type { TrailFilter, TrailFilterValues, TrailBoundingBox };