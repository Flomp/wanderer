import { describe, expect, it } from "vitest";
import type { Asset } from "$lib/models/asset";
import { Trail } from "$lib/models/trail";
import {
    assetPhotoURL,
    enrichTrailAssetExpands,
    isGeneratedRoutePreviewAsset,
    linkableAssetLinks,
} from "./asset_link_util";

function asset(values: Partial<Asset> = {}): Asset {
    return {
        id: "asset0000000001",
        collectionId: "assets",
        collectionName: "assets",
        type: "photo",
        author: "actor",
        ...values,
    };
}

// A route preview that a dev version stored as a trail photo, as the asset
// migration copies it.
const migratedPreview = asset({
    file: "route_abcdef1234.webp",
    metadata: {
        source_collection: "trails",
        source_record: "trail000000001",
        source_file: "route_abcdef1234.webp",
    },
});

describe("isGeneratedRoutePreviewAsset", () => {
    it("recognises route previews like the server does", () => {
        const cases: [string, Asset | undefined, boolean][] = [
            ["marked by the editor", asset({ file: "photo.jpg", metadata: { generated: { kind: "route-preview" } } }), true],
            ["migrated from a dev version", migratedPreview, true],
            ["legacy name only in the metadata", asset({ file: "photo_x1.webp", metadata: { source_file: "route_abcdef1234.webp" } }), true],
            ["legacy editor name", asset({ file: "wanderer-route-preview.webp" }), true],
            ["legacy name from a provider", { ...migratedPreview, external_provider: "immich" }, false],
            ["legacy name with a capture time", { ...migratedPreview, taken_at: "2026-01-01 10:00:00.000Z" }, false],
            ["regular photo", asset({ file: "route_to_the_summit.jpg" }), false],
            ["no asset", undefined, false],
        ];
        for (const [name, value, want] of cases) {
            expect(isGeneratedRoutePreviewAsset(value), name).toBe(want);
        }
    });

    it("leaves a migrated preview out where previews are excluded", () => {
        const photo = asset({ id: "asset0000000002", file: "photo.jpg" });
        const links = [
            { asset: migratedPreview.id, expand: { asset: migratedPreview } },
            { asset: photo.id, expand: { asset: photo } },
        ];

        expect(linkableAssetLinks(links, { includeGeneratedRoutePreview: false }).map((link) => link.asset))
            .toEqual([photo.id]);
    });
});

describe("enrichTrailAssetExpands", () => {
    const first = asset({ id: "asset0000000002", file: "first.jpg" });
    const second = asset({ id: "asset0000000003", file: "second.jpg" });
    const preview = asset({ id: "asset0000000004", file: "photo_x1.webp", metadata: { generated: { kind: "route-preview" } } });
    const unavailable = asset({ id: "asset0000000005" });

    // Search picks the same photo (TrailThumbnailURLs in db/util/assets.go).
    it("picks the thumbnail like search does", () => {
        const cases: [string, Asset[], Asset[], Asset][] = [
            ["marked photo", [preview, first, second], [second], second],
            ["first photo, not the stored index, without a mark", [preview, first, second], [], first],
            ["marked photo over a marked preview", [preview, first, second], [preview, second], second],
            ["first photo over a marked preview", [migratedPreview, first, second], [migratedPreview], first],
            ["first photo that can be shown", [unavailable, preview, first], [unavailable], first],
            ["marked preview without other photos", [migratedPreview, preview], [preview], preview],
            ["first preview without other photos", [migratedPreview, preview], [], migratedPreview],
        ];
        for (const [name, assets, marked, want] of cases) {
            const trail = new Trail("Ridge walk", { thumbnail: 2 });
            trail.expand!.trail_assets_via_trail = assets.map((value) => ({
                asset: value.id,
                is_thumbnail: marked.includes(value),
                expand: { asset: value },
            }));

            enrichTrailAssetExpands(trail);

            expect(trail.photos[trail.thumbnail!], name).toBe(assetPhotoURL(want));
        }
    });
});
