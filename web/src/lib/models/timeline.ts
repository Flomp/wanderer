interface TimelineItem {
    id: string;
    trail_id: string;
    trail_author_username: string;
    trail_author_domain: string;
    trail_iri: string;
    name: string,
    description: string;
    date: string;
    gpx: string;
    photos: string[];
    distance: number;
    duration: number;
    elevation_gain: number;
    elevation_loss: number;
    type: "trail" | "summit_log"
    author: string;
    created: string;
}

export { type TimelineItem }