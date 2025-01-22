import { page } from "$app/state";
import { get } from "svelte/store";

export function formatTimeHHMM(minutes?: number) {
    if (!minutes) {
        return "-";
    }
    const m = minutes % 60;

    const h = (minutes - m) / 60;

    return (h < 10 ? "0" : "") + h.toString() + "h " + (Math.round(m) < 10 ? "0" : "") + Math.round(m).toString() + "m";
}

export function formatDistance(meters?: number) {
    if (meters === undefined) {
        return "-";
    }

    const unit = page.data.settings?.unit ?? "metric";

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

    const unit = page.data.settings?.unit ?? "metric";

    if (unit == "metric") {
        return `${Math.round(meters)} m`
    } else {
        const feet = meters * 3.28084;

        return `${Math.round(feet)} ft`;
    }
}

export function formatSpeed(speed?: number) {
    if (speed === undefined) {
        return "-";
    }

    const unit = page.data.settings?.unit ?? "metric";

    if (unit == "metric") {
        return `${(speed * 3.6).toFixed(2)} km/h`
    } else {
        const mph = speed * 3.6 * 0.621371;

        return `${Math.round(mph)} mp/h`;
    }
}

export function formatTimeSince(date: Date) {
    const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);

    let interval = seconds / 31536000;
    if (interval > 1) {
        return { unit: "years", value: Math.floor(interval) };
    }
    interval = seconds / 2592000;
    if (interval > 1) {
        return { unit: "months", value: Math.floor(interval) };
    }
    interval = seconds / 86400;
    if (interval > 1) {
        return { unit: "days", value: Math.floor(interval) };
    }
    interval = seconds / 3600;
    if (interval > 1) {
        return { unit: "hours", value: Math.floor(interval) };
    }
    interval = seconds / 60;
    if (interval > 1) {
        return { unit: "minutes", value: Math.floor(interval) };
    }
    return { unit: "seconds", value: seconds };
}