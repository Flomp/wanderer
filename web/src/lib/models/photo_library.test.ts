import type { Position } from "geojson";
import { describe, expect, it } from "vitest";
import { haversineCumulatedDistanceWgs84, waypointTrackIndex } from "$lib/vendor/maplibre-elevation-profile/tools";
import { earliestPhotoLibraryCaptureTime, photoLibraryCaptureTime, type PhotoLibraryCandidate } from "./photo_library";

function candidate(values: Partial<PhotoLibraryCandidate> = {}): PhotoLibraryCandidate {
    return {
        assetId: "asset",
        originalFileName: "photo.jpg",
        takenAt: "2026-10-03T10:30:00Z",
        lat: 46,
        lon: 8,
        distance: 0,
        distanceFromStart: 0,
        ...values,
    };
}

describe("photo library capture times", () => {
    it("only counts the capture time, not takenAt", () => {
        expect(photoLibraryCaptureTime(candidate({ capturedAt: "2026-10-03T10:45:00Z" }))).toBe(
            Date.parse("2026-10-03T10:45:00Z"),
        );
        expect(photoLibraryCaptureTime(candidate())).toBeUndefined();
        expect(photoLibraryCaptureTime(candidate({ capturedAt: "not a time" }))).toBeUndefined();
        expect(earliestPhotoLibraryCaptureTime([candidate()])).toBeUndefined();
    });

    it("puts a waypoint from photos on the way back on the return pass", () => {
        // 1 km north and back from 10:00 to 11:00, passing the spot 500 m from
        // the start at 10:15 and at 10:45.
        const positions: Position[] = [];
        const times: Date[] = [];
        for (let i = 0; i <= 20; i++) {
            positions.push([8, 46 + 0.0009 * (i > 10 ? 20 - i : i)]);
            times.push(new Date(Date.UTC(2026, 9, 3, 10, 3 * i)));
        }
        const cumulatedDistance = haversineCumulatedDistanceWgs84(positions);
        const [lon, lat] = positions[5];

        const _time = earliestPhotoLibraryCaptureTime([
            candidate({ capturedAt: "2026-10-03T10:46:00Z" }),
            candidate({ capturedAt: "2026-10-03T10:45:00Z" }),
            candidate(),
        ]);
        expect(_time?.toISOString()).toBe("2026-10-03T10:45:00.000Z");

        // The candidates' distance from start is the nearest point, on the way out.
        const waypoint = { lat, lon, distance_from_start: cumulatedDistance[5] };
        expect(waypointTrackIndex(positions, times, cumulatedDistance, waypoint)).toBe(5);
        expect(waypointTrackIndex(positions, times, cumulatedDistance, { ...waypoint, _time })).toBe(15);
    });
});
