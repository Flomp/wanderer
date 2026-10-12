import { describe, expect, it, vi } from "vitest";
import GPX from "$lib/models/gpx/gpx";
import Track from "$lib/models/gpx/track";
import TrackSegment from "$lib/models/gpx/track-segment";
import GPXWaypoint from "$lib/models/gpx/waypoint";
import { Trail } from "$lib/models/trail";
import { trails_show } from "$lib/stores/trail_store";
import { joinContiguousTrackSegments, trail2gpx } from "./gpx_util";

vi.mock("$lib/stores/trail_store", () => ({ trails_show: vi.fn() }));

function pt(lat: number | string, lon: number | string) {
    return new GPXWaypoint({ $: { lat: lat as number, lon: lon as number } });
}

function segment(points: GPXWaypoint[], extensions?: string) {
    return new TrackSegment({ trkpt: points, extensions });
}

function coordinates(gpx: GPX) {
    return (gpx.trk ?? []).map((track) =>
        (track.trkseg ?? []).map((seg) =>
            (seg.trkpt ?? []).map((point) => [Number(point.$.lat), Number(point.$.lon)]),
        ),
    );
}

const multiLegGpx = `<gpx><trk>` +
    `<trkseg><trkpt lat="47.0" lon="11.0"/><trkpt lat="47.1" lon="11.1"/><trkpt lat="47.2" lon="11.2"/></trkseg>` +
    `<trkseg><trkpt lat="47.2" lon="11.2"/><trkpt lat="47.3" lon="11.3"/></trkseg>` +
    `<trkseg><trkpt lat="47.3" lon="11.3"/><trkpt lat="47.4" lon="11.4"/><trkpt lat="47.0" lon="11.0"/></trkseg>` +
    `</trk></gpx>`;

describe("joinContiguousTrackSegments", () => {
    it("merges meeting legs into one segment and keeps the closing point", () => {
        const gpx = new GPX({
            trk: [new Track({
                trkseg: [
                    segment([pt("47.0", "11.0"), pt(47.1, 11.1), pt(47.2, 11.2)]),
                    segment([pt("47.2", "11.2"), pt(47.3, 11.3)]),
                    segment([pt("47.30", "11.30"), pt(47.4, 11.4), pt(47, 11)]),
                ],
            })],
        });

        joinContiguousTrackSegments(gpx);

        expect(coordinates(gpx)).toEqual([[
            [[47, 11], [47.1, 11.1], [47.2, 11.2], [47.3, 11.3], [47.4, 11.4], [47, 11]],
        ]]);
    });

    it("keeps segments apart when their endpoints differ", () => {
        const gpx = new GPX({
            trk: [new Track({
                trkseg: [
                    segment([pt(1, 1), pt(2, 2)]),
                    segment([pt(3, 3), pt(4, 4)]),
                ],
            })],
        });

        joinContiguousTrackSegments(gpx);

        expect(coordinates(gpx)).toEqual([[[[1, 1], [2, 2]], [[3, 3], [4, 4]]]]);
    });

    it("leaves a single segment unchanged", () => {
        const gpx = new GPX({
            trk: [new Track({ trkseg: [segment([pt(1, 2), pt(3, 4), pt(5, 6)])] })],
        });
        const pointCount = gpx.trk?.[0].trkseg?.[0].trkpt?.length;

        joinContiguousTrackSegments(gpx);

        expect(gpx.trk?.[0].trkseg).toHaveLength(1);
        expect(gpx.trk?.[0].trkseg?.[0].trkpt).toHaveLength(pointCount!);
    });

    it("joins segments only inside the same track", () => {
        const gpx = new GPX({
            trk: [
                new Track({
                    name: "out",
                    trkseg: [
                        segment([pt(1, 1), pt(2, 2)]),
                        segment([pt("2", "2"), pt(3, 3)]),
                    ],
                }),
                new Track({
                    name: "back",
                    trkseg: [
                        segment([pt(3, 3), pt(4, 4)]),
                        segment([pt(4, 4), pt(5, 5)]),
                    ],
                }),
            ],
        });

        joinContiguousTrackSegments(gpx);

        expect(gpx.trk?.map((track) => track.name)).toEqual(["out", "back"]);
        expect(coordinates(gpx)).toEqual([
            [[[1, 1], [2, 2], [3, 3]]],
            [[[3, 3], [4, 4], [5, 5]]],
        ]);
    });

    it("does not absorb a segment that carries extensions", () => {
        const gpx = new GPX({
            trk: [new Track({
                trkseg: [
                    segment([pt(1, 1), pt(2, 2)]),
                    segment([pt(2, 2), pt(3, 3)], "<extensions><line>1</line></extensions>"),
                    segment([pt(3, 3), pt(4, 4)]),
                ],
            })],
        });

        joinContiguousTrackSegments(gpx);

        expect(coordinates(gpx)).toEqual([
            [[[1, 1], [2, 2]], [[2, 2], [3, 3]], [[3, 3], [4, 4]]],
        ]);
    });

    it("skips an empty segment between meeting legs", () => {
        const gpx = new GPX({
            trk: [new Track({
                trkseg: [
                    segment([pt(1, 1), pt(2, 2)]),
                    segment([]),
                    segment([pt("2.0", "2.0"), pt(3, 3)]),
                ],
            })],
        });

        joinContiguousTrackSegments(gpx);

        expect(coordinates(gpx)).toEqual([[[[1, 1], [2, 2], [3, 3]]]]);
    });
});

describe("trail2gpx", () => {
    it("exports a multi-leg route as one segment without fetching", async () => {
        const fetchSpy = vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("unexpected fetch"));
        const trail = new Trail("Alpine loop", { gpx_data: multiLegGpx });

        const exported = GPX.parse(await trail2gpx(trail));

        expect(exported.trk).toHaveLength(1);
        expect(exported.trk?.[0].name).toBe("Alpine loop");
        expect(exported.trk?.[0].trkseg).toHaveLength(1);
        expect(coordinates(exported)).toEqual([[
            [[47, 11], [47.1, 11.1], [47.2, 11.2], [47.3, 11.3], [47.4, 11.4], [47, 11]],
        ]]);
        expect(trails_show).not.toHaveBeenCalled();
        expect(fetchSpy).not.toHaveBeenCalled();
        fetchSpy.mockRestore();
    });
});
