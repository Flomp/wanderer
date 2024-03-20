import { currentUser } from "$lib/stores/user_store";
import { get } from "svelte/store";

export function formatTimeHHMM(minutes?: number) {
    if (!minutes) {
        return "-";
    }
    const m = minutes % 60;

    const h = (minutes - m) / 60;

    return (h < 10 ? "0" : "") + h.toString() + "h " + (m < 10 ? "0" : "") + Math.round(m).toString() + "m";
}

export function formatDistance(meters?: number) {
    if (meters === undefined) {
        return "-";
    }

    const unit = get(currentUser)?.unit ?? "metric";

    if (unit == "metric") {
        if (meters >= 1000) {
            return `${(meters / 1000).toFixed(2)} km`
        } else {
            return meters % 1 == 0 ? `${meters} m` : `${Math.round(meters)} m`;
        }
    } else {
        const miles = meters * 0.000621371;
        const roundedMiles = miles.toFixed(2);

        return `${roundedMiles} mi`;
    }
}

export function formatElevation(meters?: number) {
    if (meters === undefined) {
        return "-";
    }

    const unit = get(currentUser)?.unit ?? "metric";

    if (unit == "metric") {
        return `${Math.round(meters)} m`
    } else {
        const feet = meters * 3.28084;

        return `${Math.round(feet)} ft`;
    }
}
