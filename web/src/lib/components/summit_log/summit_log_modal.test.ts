import { describe, expect, it } from "vitest";
import type { PhotoLibraryCandidate } from "$lib/models/photo_library";
import type { SummitLog } from "$lib/models/summit_log";
import { applyPhotoLibrarySelection, removePendingPicks } from "./summit_log_modal.svelte";

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

describe("summit log dialog", () => {
    it("keeps the copied photos of a duplicated entry when library photos are added", () => {
        // A duplicated entry links its copied photos by ID.
        const log: Pick<SummitLog, "photos" | "_assetLinks" | "_assetPluginLinks"> = {
            photos: [photo("copied")], _assetLinks: [id("copied")],
        };

        applyPhotoLibrarySelection(log, [candidate("library"), candidate("remote", "immich")]);

        expect(log._assetLinks).toEqual([id("copied"), id("library")]);
        expect(log.photos).toEqual([photo("copied"), photo("library")]);
        expect(log._assetPluginLinks).toEqual([{ pluginId: "immich", assetIds: ["remote"] }]);
    });

    it("keeps earlier picks of an entry that is not saved yet when more are added", () => {
        const log: Pick<SummitLog, "photos" | "_assetLinks" | "_assetPluginLinks"> = {
            photos: [], _assetPluginLinks: [{ pluginId: "immich", assetIds: ["earlier"] }],
        };

        applyPhotoLibrarySelection(log, [candidate("added", "immich"), candidate("earlier", "immich")]);

        expect(log._assetPluginLinks).toEqual([{ pluginId: "immich", assetIds: ["earlier", "added"] }]);
    });

    it("leaves picks that are still pending after a failed save to their cards", () => {
        const log: Pick<SummitLog, "photos" | "_assetPluginLinks"> = {
            photos: [photo("copied"), photo("library")],
            _assetPluginLinks: [{ pluginId: "immich", assetIds: ["earlier", "remote"] }],
        };

        removePendingPicks(log, [candidate("library"), candidate("remote", "immich")]);

        expect(log.photos).toEqual([photo("copied")]);
        expect(log._assetPluginLinks).toEqual([{ pluginId: "immich", assetIds: ["earlier"] }]);

        removePendingPicks(log, [candidate("earlier", "immich")]);

        expect(log._assetPluginLinks).toBeUndefined();
    });
});
