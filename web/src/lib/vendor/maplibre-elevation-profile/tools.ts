import { haversineDistance } from "$lib/models/gpx/utils";
import type {
    Position
} from "geojson";


export function haversineCumulatedDistanceWgs84(path: Position[]): number[] {
    if (path.length < 2) {
        return [];
    }

    let totalDistance = 0;
    const distances: number[] = [0];    

    for (let i = 0; i < path.length - 1; i++) {
        const [lon1, lat1] = path[i];
        const [lon2, lat2] = path[i + 1];

        totalDistance += haversineDistance(lat1, lon1, lat2, lon2);
        distances.push(totalDistance);
    }    

    return distances; // Array of cumulative distances in meters
}

export function smoothElevations(positions: Position[], windowSize: number): Position[] {
    // Ensure windowSize is valid (at least 1)
    if (windowSize < 1) throw new Error("Window size must be at least 1.");

    // Create a new array with smoothed elevations
    return positions.map((pos, i, arr) => {
        const start = Math.max(0, i - Math.floor(windowSize / 2)); // Start index for the window
        const end = Math.min(arr.length, i + Math.floor(windowSize / 2) + 1); // End index for the window
        const segment = arr.slice(start, end); // Extract the positions in the window

        // Calculate the weighted moving average of elevations
        const weights = segment.map((_, idx) => idx + 1); // Increasing weights: 1, 2, 3...
        const elevations = segment.map(p => p[2]); // Extract elevations
        const weightedSum = elevations.reduce((sum, elevation, idx) => sum + elevation * weights[idx], 0);
        const weightTotal = weights.reduce((sum, weight) => sum + weight, 0);

        const smoothedElevation = weightedSum / weightTotal; // Weighted average elevation       

        // Return a new Position with the smoothed elevation
        return [pos[0], pos[1], smoothedElevation] as Position;
    });
}