import { Trail } from "$lib/models/trail";
import type { Threshold } from "$lib/models/difficulty_algorithms";


export function getTrailDifficulty(t: Trail, thresholds: Threshold[]) : "easy" | "moderate" | "difficult" | undefined {
    if (!t.distance || !t.elevation_gain) return undefined;

    if (doGetTrailDifficulty(t, thresholds, "difficult") != undefined)
        return "difficult";
    if (doGetTrailDifficulty(t, thresholds, "moderate") != undefined)
        return "moderate";
    
    return "easy";
}

function doGetTrailDifficulty(t: Trail, thresholds: Threshold[], difficulty: "moderate" | "difficult") : "moderate" | "difficult" | undefined {
    if (!t.distance || !t.elevation_gain) return undefined;

    let threshDistance = thresholds.find((thresh) => thresh.difficulty == difficulty && thresh.type == "distance");
    if (threshDistance && t.distance > threshDistance.limit)
        return difficulty;
    let threshElevation = thresholds.find((thresh) => thresh.difficulty == difficulty && thresh.type == "elevation");
    if (threshElevation && t.elevation_gain > threshElevation.limit)
        return difficulty;

    return undefined;
}