import type { Asset } from "$lib/models/asset";
import { RecordIdValueSchema } from "$lib/models/api/base_schema";
import {
    extractPhotoExifMetadata,
    type GPSCoordinates,
    type PhotoExifMetadata,
} from "$lib/util/exif_util";
import { Collection, handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

interface AssetLinkTarget {
    trail?: string;
    waypoint?: string;
    summit_log?: string;
}

/**
 * @swagger
 * /api/v1/assets:
 *   put:
 *     summary: Upload or link photo assets
 *     description: >
 *       Uploads photo files as assets, links photos the user already has in wanderer, or both, and attaches them to a
 *       trail, waypoint or summit log. The user must be the author of the target or of its trail, or have an edit
 *       share for that trail.
 *     tags:
 *       - Assets
 *     requestBody:
 *       required: true
 *       content:
 *         multipart/form-data:
 *           schema:
 *             type: object
 *             properties:
 *               files:
 *                 type: array
 *                 items:
 *                   type: string
 *                   format: binary
 *                 description: Photo files to upload.
 *               assetIds:
 *                 type: array
 *                 items:
 *                   type: string
 *                 description: >
 *                   IDs of the user's own assets to link instead of uploading them again, such as photos found through
 *                   `POST /api/v1/assets/library`. Repeat the field for several IDs. Values that are not record IDs are
 *                   ignored, and an asset already linked to the target is not linked twice.
 *               trail:
 *                 type: string
 *                 description: >
 *                   Trail id to attach assets to. Together with `waypoint` or `summit_log` it must be that record's
 *                   trail, and the assets are not linked to the trail itself.
 *               waypoint:
 *                 type: string
 *                 description: Waypoint id to attach assets to
 *               summit_log:
 *                 type: string
 *                 description: Summit log id to attach assets to
 *               metadata:
 *                 type: string
 *                 description: JSON object stored as `metadata` on every asset created from `files`.
 *               fileCoordinates:
 *                 type: string
 *                 description: JSON array of optional per-file coordinates aligned with files.
 *               fileMetadata:
 *                 type: string
 *                 description: >
 *                   JSON array aligned with `files`. Each entry is `null` or an object with the optional fields
 *                   `coordinates` (`{"lat": 47.37, "lon": 8.54}`), `takenAt` (date-time) and `takenAtSource`, whose
 *                   values replace what the server reads from the file. Set `takenAtSource` to `exif` only when
 *                   `takenAt` is the capture time from the photo's own metadata, such as its EXIF capture date, and not
 *                   a file or upload time. The server then marks the time in the asset's `metadata.capture_time`, and
 *                   `POST /api/v1/assets/library` offers marked times as `capturedAt` to tell apart the passes of a
 *                   track that covers the same ground twice. Without `takenAtSource`, a `takenAt` is stored unmarked;
 *                   an entry without `takenAt` takes the time and its mark from the file.
 *     responses:
 *       200:
 *         description: The assets linked from `assetIds`, followed by the assets created from `files`
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 $ref: '#/components/schemas/Asset'
 *       400:
 *         description: Missing or invalid target relation, or a waypoint or summit log that is not on `trail`
 *       401:
 *         description: Unauthorized
 *       403:
 *         description: Insufficient permissions for target trail
 *       404:
 *         description: Target or linked asset not found
 *       500:
 *         description: Internal Server Error
 */
export async function PUT(event: RequestEvent) {
    try {
        if (!event.locals.user) {
            return json({ message: "Unauthorized" }, { status: 401 });
        }
        if (!event.locals.user.actor) {
            return json({ message: "Actor not found" }, { status: 401 });
        }

        const data = await event.request.formData();
        const files = data.getAll("files").filter(isUploadedFile);
        const assetIds = data
            .getAll("assetIds")
            .filter((id): id is string => typeof id === "string" && RecordIdValueSchema.safeParse(id).success);
        const target = await validateTarget(event, data);
        const metadata = metadataValue(data.get("metadata"));
        const fileCoordinates = fileCoordinatesValue(data.get("fileCoordinates"), files.length);
        const fileMetadata = fileMetadataValue(data.get("fileMetadata"), files.length, fileCoordinates);
        const created: Asset[] = [];

        for (const assetId of assetIds) {
            const record = await event.locals.pb.collection(Collection.assets).getOne<Asset>(assetId);
            await createAssetLinks(event, record.id, target);
            created.push(record);
        }

        for (const [index, file] of files.entries()) {
            const asset = new FormData();
            asset.append("type", "photo");
            asset.append("storage_mode", "copy");
            asset.append("remote_status", "available");
            asset.append("author", event.locals.user.actor);
            asset.append("file", file);
            await appendAssetExifMetadata(asset, file, fileMetadata[index], metadata);

            const record = await event.locals.pb.collection(Collection.assets).create<Asset>(asset, { requestKey: null });
            await createAssetLinks(event, record.id, target);
            created.push(record);
        }

        return json(created);
    } catch (e: any) {
        return handleError(e);
    }
}

async function createAssetLinks(
    event: RequestEvent,
    assetId: string,
    target: AssetLinkTarget,
) {
    const links = [
        { collection: Collection.trail_assets, field: "trail", id: target.trail },
        { collection: Collection.waypoint_assets, field: "waypoint", id: target.waypoint },
        { collection: Collection.summit_log_assets, field: "summit_log", id: target.summit_log },
    ];

    for (const link of links) {
        if (!link.id) {
            continue;
        }
        const existing = await event.locals.pb.collection(link.collection).getFullList({
            filter: event.locals.pb.filter(`asset = {:asset} && ${link.field} = {:target}`, {
                asset: assetId,
                target: link.id,
            }),
        });
        if (existing.length) {
            continue;
        }
        await event.locals.pb.collection(link.collection).create(
            {
                asset: assetId,
                [link.field]: link.id,
            },
            { requestKey: null },
        );
    }
}

async function validateTarget(event: RequestEvent, data: FormData) {
    const user = event.locals.user!;
    const requestedTarget = {
        trail: recordIdValue(data.get("trail")),
        waypoint: recordIdValue(data.get("waypoint")),
        summit_log: recordIdValue(data.get("summit_log")),
    };
    let trailContext = requestedTarget.trail;
    let authorized = false;

    if (!requestedTarget.trail && !requestedTarget.waypoint && !requestedTarget.summit_log) {
        throw new ClientResponseError({
            status: 400,
            response: { message: "Asset target is required" },
        });
    }

    if (requestedTarget.waypoint) {
        const waypoint = await event.locals.pb.collection(Collection.waypoints).getOne(requestedTarget.waypoint);
        if (!trailContext) {
            trailContext = waypoint.trail;
        } else if (waypoint.trail !== trailContext) {
            throw new ClientResponseError({
                status: 400,
                response: { message: "Waypoint does not belong to trail" },
            });
        }
        if (waypoint.author === user.actor) {
            authorized = true;
        }
    }

    if (requestedTarget.summit_log) {
        const summitLog = await event.locals.pb.collection(Collection.summit_logs).getOne(requestedTarget.summit_log);
        if (!trailContext) {
            trailContext = summitLog.trail;
        } else if (summitLog.trail !== trailContext) {
            throw new ClientResponseError({
                status: 400,
                response: { message: "Summit log does not belong to trail" },
            });
        }
        if (summitLog.author === user.actor) {
            authorized = true;
        }
    }

    if (trailContext) {
        const trail = await event.locals.pb.collection(Collection.trails).getOne(trailContext);
        if (trail.author === user.actor) {
            authorized = true;
        }
        if (!authorized) {
            authorized = await hasTrailEditShare(event, trailContext, user.actor);
        }
    }

    if (!authorized) {
        throw new ClientResponseError({
            status: 403,
            response: { message: "Insufficient permissions for trail" },
        });
    }

    return {
        trail: requestedTarget.waypoint || requestedTarget.summit_log ? undefined : requestedTarget.trail,
        waypoint: requestedTarget.waypoint,
        summit_log: requestedTarget.summit_log,
    };
}

async function hasTrailEditShare(event: RequestEvent, trailId: string, actorId: string): Promise<boolean> {
    const shares = await event.locals.pb.collection(Collection.trail_share).getFullList({
        filter: event.locals.pb.filter("trail = {:trail} && actor = {:actor} && permission = {:permission}", {
            trail: trailId,
            actor: actorId,
            permission: "edit",
        }),
    });
    return shares.length > 0;
}

function stringValue(value: FormDataEntryValue | null): string | undefined {
    return typeof value === "string" && value.length ? value : undefined;
}

function recordIdValue(value: FormDataEntryValue | null): string | undefined {
    const raw = stringValue(value);
    return raw ? RecordIdValueSchema.parse(raw) : undefined;
}

function isUploadedFile(value: FormDataEntryValue): value is File {
    return (
        typeof value === "object" &&
        value !== null &&
        "arrayBuffer" in value &&
        typeof value.arrayBuffer === "function" &&
        "name" in value
    );
}

async function appendAssetExifMetadata(
    asset: FormData,
    file: File,
    fileMetadata: PhotoExifMetadata | undefined,
    metadata: Record<string, unknown> | undefined,
) {
    const extracted = await extractPhotoExifMetadata(file);
    const coordinates = fileMetadata?.coordinates ?? extracted.coordinates;
    const time = fileMetadata?.takenAt ? fileMetadata : extracted;
    const takenAt = time.takenAt ? new Date(time.takenAt) : undefined;
    const hasTakenAt = takenAt !== undefined && !Number.isNaN(takenAt.getTime());

    if (coordinates) {
        asset.append("lat", coordinates.lat.toString());
        asset.append("lon", coordinates.lon.toString());
    }

    if (hasTakenAt) {
        asset.append("taken_at", takenAt.toISOString());
    }

    // The photo library offers a marked taken_at to tell apart the passes of a
    // track that covers the same ground twice. The mark holds the time it
    // vouches for, so it no longer counts once taken_at changes.
    const assetMetadata = hasTakenAt && time.takenAtSource === "exif"
        ? { ...metadata, capture_time: { source: "exif", taken_at: takenAt.toISOString() } }
        : metadata;
    if (assetMetadata) {
        asset.append("metadata", JSON.stringify(assetMetadata));
    }
}

function fileCoordinatesValue(value: FormDataEntryValue | null, expectedLength: number): Array<GPSCoordinates | undefined> {
    const empty = Array<GPSCoordinates | undefined>(expectedLength).fill(undefined);
    if (typeof value !== "string" || !value.length) {
        return empty;
    }

    let parsed: unknown;
    try {
        parsed = JSON.parse(value);
    } catch {
        return empty;
    }
    if (!Array.isArray(parsed)) {
        return empty;
    }

    return empty.map((_, index) => gpsCoordinatesValue(parsed[index]));
}

function gpsCoordinatesValue(value: unknown): GPSCoordinates | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }
    const coordinates = value as { lat?: unknown; lon?: unknown };
    const lat = Number(coordinates.lat);
    const lon = Number(coordinates.lon);
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
        return undefined;
    }
    return { lat, lon };
}

function fileMetadataValue(
    value: FormDataEntryValue | null,
    expectedLength: number,
    fallbackCoordinates: Array<GPSCoordinates | undefined>,
): Array<PhotoExifMetadata | undefined> {
    const empty = fallbackCoordinates.map((coordinates) => (
        coordinates ? { coordinates } : undefined
    ));
    if (typeof value !== "string" || !value.length) {
        return empty;
    }

    let parsed: unknown;
    try {
        parsed = JSON.parse(value);
    } catch {
        return empty;
    }
    if (!Array.isArray(parsed)) {
        return empty;
    }

    return Array.from({ length: expectedLength }, (_, index) => {
        const metadata = photoExifMetadataValue(parsed[index]);
        if (!metadata) {
            return empty[index];
        }
        return {
            coordinates: metadata.coordinates ?? empty[index]?.coordinates,
            takenAt: metadata.takenAt,
            takenAtSource: metadata.takenAtSource,
        };
    });
}

function photoExifMetadataValue(value: unknown): PhotoExifMetadata | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }
    const raw = value as { coordinates?: unknown; takenAt?: unknown; takenAtSource?: unknown };
    const coordinates = gpsCoordinatesValue(raw.coordinates);
    const takenAt = typeof raw.takenAt === "string" && !Number.isNaN(new Date(raw.takenAt).getTime())
        ? new Date(raw.takenAt).toISOString()
        : undefined;
    if (!coordinates && !takenAt) {
        return undefined;
    }
    const takenAtSource = takenAt && raw.takenAtSource === "exif" ? "exif" : undefined;
    return { coordinates, takenAt, takenAtSource };
}

function metadataValue(value: FormDataEntryValue | null): Record<string, unknown> | undefined {
    if (typeof value !== "string" || !value.length) {
        return undefined;
    }
    let parsed: unknown;
    try {
        parsed = JSON.parse(value);
    } catch {
        return undefined;
    }
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
        return undefined;
    }
    return parsed as Record<string, unknown>;
}
