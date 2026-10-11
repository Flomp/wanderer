import type { Asset } from "$lib/models/asset";
import { RecordIdSchema, RecordIdValueSchema } from "$lib/models/api/base_schema";
import { Collection, handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

const AssetMetadataPatchSchema = z.object({
    lat: z.number().min(-90).max(90).optional(),
    lon: z.number().min(-180).max(180).optional(),
    takenAt: z.string().optional(),
    replaceCoordinates: z.boolean().optional(),
    markChecked: z.boolean().optional(),
}).refine((data) => (data.lat === undefined) === (data.lon === undefined), {
    message: "lat and lon must be provided together",
});

/**
 * @swagger
 * /api/v1/assets/{id}:
 *   delete:
 *     summary: Delete asset or remove it from a target
 *     description: >
 *       Without a target, deletes an asset the user is the author of, together with its links.
 *       With `trail`, `waypoint` or `summit_log`, removes the asset from the given target instead, which needs the same
 *       permission as attaching photos to it. The asset itself is then deleted only when it has no links left and the
 *       user is its author; `assetDeleted` says whether that happened.
 *     tags:
 *       - Assets
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *       - in: query
 *         name: trail
 *         schema:
 *           type: string
 *         description: Trail to remove the asset from. Ignored together with `waypoint` or `summit_log`.
 *       - in: query
 *         name: waypoint
 *         schema:
 *           type: string
 *         description: Waypoint to remove the asset from.
 *       - in: query
 *         name: summit_log
 *         schema:
 *           type: string
 *         description: Summit log to remove the asset from.
 *     responses:
 *       200:
 *         description: Asset deleted or removed from the target
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 acknowledged:
 *                   type: boolean
 *                 assetDeleted:
 *                   type: boolean
 *                   description: Only with a target. Whether the asset itself was deleted.
 *       400:
 *         description: Invalid asset or target id
 *       401:
 *         description: Unauthorized
 *       403:
 *         description: Without a target, the user is not the asset's author; otherwise insufficient permissions for the target
 *       404:
 *         description: Asset or target not found
 *       500:
 *         description: Internal Server Error
 */
export async function DELETE(event: RequestEvent) {
    try {
        if (!event.locals.user) {
            return json({ message: "Unauthorized" }, { status: 401 });
        }

        const safeParams = RecordIdSchema.parse(event.params);

        const targets = z.object({
            trail: RecordIdValueSchema.optional(),
            waypoint: RecordIdValueSchema.optional(),
            summit_log: RecordIdValueSchema.optional(),
        }).parse(Object.fromEntries(event.url.searchParams));

        const linkTargets = assetLinkTargets(targets);

        const searchParams = new URLSearchParams();
        for (const target of linkTargets) {
            searchParams.set(target.field, target.id);
        }
        const query = searchParams.toString();
        const response = await event.locals.pb.send(`/assets/${safeParams.id}${query ? `?${query}` : ""}`, {
            method: "DELETE",
            fetch: event.fetch,
        });
        return json(response);
    } catch (e: any) {
        return handleError(e);
    }
}

/**
 * @swagger
 * /api/v1/assets/{id}:
 *   patch:
 *     summary: Fill in missing photo metadata
 *     description: >
 *       Fills in what a photo asset is missing, as the photo metadata maintenance does with the photo's own EXIF data
 *       or values the user enters. Only the asset's author can change it. Fields that do not apply are ignored, and
 *       when nothing applies the asset is returned unchanged.
 *     tags:
 *       - Assets
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               lat:
 *                 type: number
 *                 minimum: -90
 *                 maximum: 90
 *                 description: >
 *                   Latitude, sent together with `lon`. Both are written only when the asset has no coordinates yet
 *                   (`lat` and `lon` both 0), or with `replaceCoordinates` on a migrated asset.
 *               lon:
 *                 type: number
 *                 minimum: -180
 *                 maximum: 180
 *                 description: Longitude, sent together with `lat`.
 *               takenAt:
 *                 type: string
 *                 format: date-time
 *                 description: >
 *                   Time the photo was taken, written only when the asset's `taken_at` is empty and stored in UTC.
 *                   A value that cannot be read as a date is ignored. The time gets no `metadata.capture_time`
 *                   mark, so it does not count as the photo's own capture time, for example as `capturedAt` in
 *                   `POST /api/v1/assets/library`.
 *               replaceCoordinates:
 *                 type: boolean
 *                 description: >
 *                   Replace existing coordinates with `lat` and `lon`. Applies only to an asset migrated from the
 *                   photos of an earlier version (`metadata.source_collection` is set), which got the coordinates of
 *                   its trail or waypoint rather than its own.
 *               markChecked:
 *                 type: boolean
 *                 description: >
 *                   Record the check in `metadata.exif_backfill.checked_at`, keeping the other metadata. The photo
 *                   metadata maintenance then no longer lists the asset.
 *     responses:
 *       200:
 *         description: The asset, updated or unchanged
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/Asset'
 *       400:
 *         description: Invalid asset id or body, or only one of `lat` and `lon`
 *       401:
 *         description: Unauthorized
 *       404:
 *         description: Asset not found, or a change to an asset the user is not the author of
 *       500:
 *         description: Internal Server Error
 */
export async function PATCH(event: RequestEvent) {
    try {
        if (!event.locals.user) {
            return json({ message: "Unauthorized" }, { status: 401 });
        }

        const safeParams = RecordIdSchema.parse(event.params);
        const data = AssetMetadataPatchSchema.parse(await event.request.json());

        const asset = await event.locals.pb.collection(Collection.assets).getOne<Asset>(safeParams.id);
        const update: Record<string, unknown> = {};

        if (
            data.lat !== undefined &&
            data.lon !== undefined &&
            (
                !assetHasCoordinates(asset) ||
                (data.replaceCoordinates && isLegacyMigratedAsset(asset))
            )
        ) {
            update.lat = data.lat;
            update.lon = data.lon;
        }

        if (data.takenAt && !asset.taken_at) {
            const takenAt = new Date(data.takenAt);
            if (!Number.isNaN(takenAt.getTime())) {
                update.taken_at = takenAt.toISOString();
            }
        }

        if (data.markChecked) {
            update.metadata = metadataWithExifBackfillMarker(asset.metadata);
        }

        if (Object.keys(update).length === 0) {
            return json(asset);
        }

        const updated = await event.locals.pb.collection(Collection.assets).update(
            safeParams.id,
            update,
            { requestKey: null },
        );
        return json(updated);
    } catch (e: any) {
        return handleError(e);
    }
}

function assetLinkTargets(targets: { trail?: string; waypoint?: string; summit_log?: string }) {
    const trailIsDirectTarget = targets.trail && !targets.waypoint && !targets.summit_log;
    return [
        { field: "trail", id: trailIsDirectTarget ? targets.trail : undefined },
        { field: "waypoint", id: targets.waypoint },
        { field: "summit_log", id: targets.summit_log },
    ].filter((target): target is { field: string; id: string } => Boolean(target.id));
}

function assetHasCoordinates(asset: Pick<Asset, "lat" | "lon">): boolean {
    return (
        typeof asset.lat === "number" &&
        typeof asset.lon === "number" &&
        Number.isFinite(asset.lat) &&
        Number.isFinite(asset.lon) &&
        (asset.lat !== 0 || asset.lon !== 0)
    );
}

function isLegacyMigratedAsset(asset: Pick<Asset, "metadata">): boolean {
    return typeof asset.metadata?.source_collection === "string";
}

function metadataWithExifBackfillMarker(metadata: Asset["metadata"]): Record<string, unknown> {
    const base = metadata && typeof metadata === "object" && !Array.isArray(metadata)
        ? metadata
        : {};
    return {
        ...base,
        exif_backfill: {
            checked_at: new Date().toISOString(),
        },
    };
}
