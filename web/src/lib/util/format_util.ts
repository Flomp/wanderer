export function formatTimeHHMM(minutes: number) {
    const m = minutes % 60;

    const h = (minutes - m) / 60;

    return (h < 10 ? "0" : "") + h.toString() + "h " + (m < 10 ? "0" : "") + m.toString() + "m";
}

export function formatMeters(meters: number) {
    if (meters % 1 === 0) {
        return meters >= 1000 ? `${(meters / 1000)} km` : `${meters} m`;
    } else {
        return meters >= 1000 ? `${(meters / 1000).toFixed(2)} km` : `${meters.toFixed(2)} m`
    }

}

export function formatISODate(isoTimestamp: string) {
    const date = new Date(isoTimestamp);
    
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0'); // Month is zero-based
    const year = date.getFullYear();

    return `${day}.${month}.${year}`;
}