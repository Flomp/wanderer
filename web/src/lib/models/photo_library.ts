export interface PhotoLibraryBounds {
    west: number;
    south: number;
    east: number;
    north: number;
}

export interface PhotoLibraryCandidate {
    source?: "plugin" | "wanderer";
    providerId?: string;
    pluginId?: string;
    externalProvider?: string;
    externalId?: string;
    assetId: string;
    originalFileName: string;
    takenAt: string;
    // Only set when the time comes from the photo's own metadata. Unlike
    // takenAt, which can be a file or upload time, it can tell apart the
    // passes of a track that covers the same ground twice.
    capturedAt?: string;
    lat: number;
    lon: number;
    pointLat?: number;
    pointLon?: number;
    distance: number;
    distanceFromStart: number;
    city?: string;
    country?: string;
    thumbnailUrl?: string;
}

export interface PhotoLibraryPluginLink {
    pluginId: string;
    assetIds: string[];
}

export function photoLibraryCandidateKey(candidate: PhotoLibraryCandidate): string {
    return `${candidate.source ?? "plugin"}:${candidate.providerId ?? candidate.pluginId ?? ""}:${candidate.assetId}`;
}

export function photoLibraryCaptureTime(candidate: PhotoLibraryCandidate): number | undefined {
    const time = candidate.capturedAt ? Date.parse(candidate.capturedAt) : NaN;
    return Number.isFinite(time) ? time : undefined;
}

// The earliest capture time of the photos of a new waypoint becomes its time,
// which picks the pass of the track it lies on.
export function earliestPhotoLibraryCaptureTime(candidates: PhotoLibraryCandidate[]): Date | undefined {
    const times = candidates
        .map(photoLibraryCaptureTime)
        .filter((time): time is number => time !== undefined);
    return times.length ? new Date(Math.min(...times)) : undefined;
}

export function photoLibraryPluginLinks(candidates: PhotoLibraryCandidate[]): PhotoLibraryPluginLink[] {
    const grouped = new Map<string, string[]>();
    for (const candidate of candidates) {
        const pluginId = candidate.pluginId ?? candidate.providerId;
        if (!pluginId) {
            continue;
        }
        grouped.set(pluginId, Array.from(new Set([
            ...(grouped.get(pluginId) ?? []),
            candidate.assetId,
        ])));
    }
    return Array.from(grouped.entries()).map(([pluginId, assetIds]) => ({
        pluginId,
        assetIds,
    }));
}

export function mergePhotoLibraryPluginLinks(
    existing: PhotoLibraryPluginLink[] | undefined,
    incoming: PhotoLibraryPluginLink[] | undefined,
): PhotoLibraryPluginLink[] | undefined {
    const grouped = new Map<string, string[]>();
    for (const link of [...(existing ?? []), ...(incoming ?? [])]) {
        grouped.set(link.pluginId, Array.from(new Set([
            ...(grouped.get(link.pluginId) ?? []),
            ...link.assetIds,
        ])));
    }
    const merged = Array.from(grouped.entries())
        .map(([pluginId, assetIds]) => ({ pluginId, assetIds }))
        .filter((link) => link.pluginId && link.assetIds.length);
    return merged.length ? merged : undefined;
}
