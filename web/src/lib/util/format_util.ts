
export function formatTimeHHMM(minutes?: number) {
    if (!minutes) {
        return "-";
    }
    const m = minutes % 60;

    const h = (minutes - m) / 60;

    return (h < 10 ? "0" : "") + h.toString() + "h " + (m < 10 ? "0" : "") + m.toString() + "m";
}

export function formatMeters(meters?: number) {
    if (!meters) {
        return "-";
    }
    if (meters % 1 === 0) {
        return meters >= 1000 ? `${(meters / 1000)} km` : `${meters} m`;
    } else {
        return meters >= 1000 ? `${(meters / 1000).toFixed(2)} km` : `${meters.toFixed(2)} m`
    }

}
