import { Trail } from "$lib/models/trail";

export function getTrailDifficulty(t: Trail, modeOfTransport: "pedestrian" | "bicycle" | "auto", speed: number) {
    switch (modeOfTransport)
    {
        case "bicycle":
            return getBikingDifficulty(t, speed);
        case "pedestrian":
            return getHikingDifficulty(t, speed);
    }
}

function getBikingDifficulty(t: Trail, speed: number) : "easy" | "moderate" | "difficult" | undefined {
    if (!t.distance || !t.elevation_gain) return undefined;
    
    const duration : number = t.distance / speed / 1000;

    if (duration > 5) return "difficult";
    if (t.elevation_gain > 450) return "difficult";

    if (duration > 2) return "moderate";
    if (t.elevation_gain > 150) return "moderate";

    return "easy";
}

function getHikingDifficulty(t: Trail, speed: number) : "easy" | "moderate" | "difficult" | undefined {
    if (!t.distance || !t.elevation_gain) return undefined;
    
    const duration : number = t.distance / speed / 1000;

    if (duration > 5) return "difficult";
    if (t.elevation_gain > 900) return "difficult";

    if (duration > 2) return "moderate";
    if (t.elevation_gain > 300) return "moderate";

    return "easy";
}