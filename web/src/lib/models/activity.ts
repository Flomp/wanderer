interface Activity {
    id: string;
    trail_id: string;
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
}

export { type Activity }