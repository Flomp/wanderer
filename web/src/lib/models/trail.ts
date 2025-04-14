import type { Category } from "./category";
import type { Comment } from "./comment";
import type { SummitLog } from "./summit_log";
import type { Tag } from "./tag";
import type { TrailShare } from "./trail_share";
import type { UserAnonymous } from "./user";
import type { Waypoint } from "./waypoint";

class Trail {
    id?: string;
    name: string;
    location?: string;
    date?: string;
    public: boolean;
    distance?: number;
    elevation_gain?: number;
    elevation_loss?: number;
    duration?: number;
    difficulty?: "easy" | "moderate" | "difficult"
    lat?: number;
    lon?: number;
    thumbnail?: number;
    photos: string[];
    gpx?: string;
    created?: string;
    category?: string;
    tags: string[];
    waypoints: string[];
    summit_logs: string[];
    polyline?: string;
    expand?: {
        tags?: Tag[]
        category?: Category;
        waypoints?: Waypoint[]
        summit_logs?: SummitLog[]
        author?: UserAnonymous
        comments_via_trail?: Comment[]
        gpx_data?: string
        trail_share_via_trail?: TrailShare[]
    }
    description?: string;
    author: string;

    constructor(name: string,
        params?: {
            id?: string,
            location?: string,
            date?: string,
            public?: boolean,
            distance?: number,
            elevation_gain?: number,
            elevation_loss?: number,
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
        this.distance = params?.distance ?? 0;
        this.elevation_gain = params?.elevation_gain ?? 0;
        this.elevation_loss = params?.elevation_loss ?? 0;
        this.duration = params?.duration ?? 0;
        this.difficulty = params?.difficulty ?? "easy";
        this.lat = params?.lat;
        this.lon = params?.lon;
        this.thumbnail = params?.thumbnail ?? 0;
        this.photos = params?.photos ?? [];
        this.waypoints = [];
        this.summit_logs = [];
        this.tags = []
        this.gpx = params?.gpx;
        this.expand = {
            category: params?.category,
            waypoints: params?.waypoints ?? [],
            summit_logs: params?.summit_logs ?? [],
            comments_via_trail: params?.comments ?? [],
            trail_share_via_trail: params?.shares ?? []
        }
        this.description = params?.description ?? "";
        this.created = params?.created;
        this.author = "000000000000000"
    }
}

interface TrailFilter {
    q: string,
    category: string[],
    tags: string[],
    difficulty: ("easy" | "moderate" | "difficult")[]
    author?: string;
    public?: boolean;
    shared?: boolean;
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
    elevationLossMin: number;
    elevationLossMax: number;
    elevationLossLimit: number;
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
    min_elevation_loss: number,
    max_elevation_loss: number,
    min_durtation: number,
    max_duration: number,
}

interface TrailBoundingBox {
    max_lat: number,
    min_lat: number,
    max_lon: number,
    min_lon: number,
}


interface TrailSearchResult {
    id: string;
    author: string;
    author_name: string;
    author_avatar: string;
    name: string;
    description: string;
    location: string;
    distance: number;
    elevation_gain: number;
    elevation_loss: number;
    duration: number;
    difficulty: "easy" | "moderate" | "difficult";
    category: string;
    completed: boolean;
    date: number;
    created: number;
    public: boolean;
    thumbnail: string;
    polyline?: string;
    shares?: string[];
    tags?: string[]
    gpx: string;
    _geo: {
        lat: number,
        lng: number
    };
}

export const defaultTrailSearchAttributes = [
    "id",
    "author",
    "author_name",
    "author_avatar",
    "name",
    "description",
    "location",
    "distance",
    "elevation_gain",
    "elevation_loss",
    "duration",
    "difficulty",
    "category",
    "completed",
    "date",
    "created",
    "public",
    "thumbnail",
    "gpx",
    "tags",
    "_geo",]


export { Trail };

export type { TrailBoundingBox, TrailFilter, TrailFilterValues, TrailSearchResult };
