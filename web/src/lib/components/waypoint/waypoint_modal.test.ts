import { describe, expect, it } from "vitest";
import type { PhotoLibraryCandidate } from "$lib/models/photo_library";
import { Waypoint } from "$lib/models/waypoint";
import { applyPhotoLibrarySelection } from "./waypoint_modal.svelte";

const id = (name: string) => name.padEnd(15, "0");
const photo = (name: string) => `/api/v1/assets/${id(name)}/file`;

function candidate(name: string, pluginId?: string): PhotoLibraryCandidate {
    return {
        source: pluginId ? "plugin" : "wanderer", pluginId,
        assetId: pluginId ? name : id(name), thumbnailUrl: pluginId ? undefined : photo(name),
        originalFileName: `${name}.jpg`, takenAt: "2026-09-13T10:00:00Z",
        lat: 47, lon: 8, distance: 0, distanceFromStart: 0,
    };
}

describe("waypoint dialog", () => {
    it("keeps the copied photos of a duplicated waypoint when library photos are added", () => {
        // A duplicated waypoint links its copied photos by ID.
        const waypoint = new Waypoint(47, 8, { photos: [photo("copied")] });
        waypoint._assetLinks = [id("copied")];

        applyPhotoLibrarySelection(waypoint, [candidate("library"), candidate("remote", "immich")]);

        expect(waypoint._assetLinks).toEqual([id("copied"), id("library")]);
        expect(waypoint.photos).toEqual([photo("copied"), photo("library")]);
        expect(waypoint._assetPluginLinks).toEqual([{ pluginId: "immich", assetIds: ["remote"] }]);
    });

    it("keeps earlier picks of a waypoint that is not saved yet when more are added", () => {
        const waypoint = new Waypoint(47, 8, {});
        waypoint._assetCandidates = [{ pluginId: "immich", assetId: "earlier", lat: 47, lon: 8 }];
        waypoint._assetPluginLinks = [{ pluginId: "immich", assetIds: ["earlier"] }];

        applyPhotoLibrarySelection(waypoint, [candidate("added", "immich"), candidate("earlier", "immich")]);

        expect(waypoint._assetCandidates?.map((c) => c.assetId)).toEqual(["earlier", "added"]);
        expect(waypoint._assetPluginLinks).toEqual([{ pluginId: "immich", assetIds: ["earlier", "added"] }]);
    });
});
