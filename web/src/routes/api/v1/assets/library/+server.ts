import { RecordIdValueSchema } from "$lib/models/api/base_schema";
import { handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

const LibraryRequestSchema = z.object({
    trailId: RecordIdValueSchema.optional(),
    waypointId: RecordIdValueSchema.optional(),
    summitLogId: RecordIdValueSchema.optional(),
    lat: z.number().optional(),
    lon: z.number().optional(),
    bounds: z.object({
        west: z.number().finite().min(-180).max(180),
        south: z.number().finite().min(-90).max(90),
        east: z.number().finite().min(-180).max(180),
        north: z.number().finite().min(-90).max(90),
    }).refine((bounds) => bounds.south <= bounds.north).optional(),
    trailData: z.string().optional(),
    takenAfter: z.string().datetime().optional(),
    takenBefore: z.string().datetime().optional(),
    doubleRadius: z.boolean().optional(),
    page: z.number().int().min(1).optional(),
    perPage: z.number().int().min(1).max(250).optional(),
});

/**
 * @swagger
 * /api/v1/assets/library:
 *   post:
 *     summary: Search the user's photos
 *     description: >
 *       Searches the user's own photos in wanderer, to attach them through the `assetIds` field of
 *       `PUT /api/v1/assets`. Only photos with coordinates are listed: inside `bounds` when given, otherwise within
 *       1 km (2 km with `doubleRadius`) of the track of `trailId` or `trailData`, or else of `lat`/`lon`. Without any
 *       of these, all such photos are listed. Generated route previews and photos already linked to the given trail,
 *       waypoint or summit log are left out.
 *     tags:
 *       - Assets
 *     requestBody:
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               trailId:
 *                 type: string
 *                 description: Trail to search along. The user must be its author or have an edit share.
 *               waypointId:
 *                 type: string
 *                 description: Waypoint whose photos are left out. The user must be allowed to attach photos to it.
 *               summitLogId:
 *                 type: string
 *                 description: Summit log whose photos are left out. The user must be allowed to attach photos to it.
 *               trailData:
 *                 type: string
 *                 description: GPX content to search along, for a trail that is not saved yet.
 *               lat:
 *                 type: number
 *                 description: Latitude of a position to search around, such as a waypoint's.
 *               lon:
 *                 type: number
 *                 description: Longitude of a position to search around.
 *               bounds:
 *                 $ref: '#/components/schemas/PhotoLibraryBounds'
 *               takenAfter:
 *                 type: string
 *                 format: date-time
 *                 description: Only photos whose `taken_at` is at or after this UTC time, written with `Z`.
 *               takenBefore:
 *                 type: string
 *                 format: date-time
 *                 description: Only photos whose `taken_at` is at or before this UTC time, written with `Z`.
 *               doubleRadius:
 *                 type: boolean
 *               page:
 *                 type: integer
 *                 minimum: 1
 *               perPage:
 *                 type: integer
 *                 minimum: 1
 *                 maximum: 250
 *                 description: >
 *                   Page size, 100 by default. Only searches by `bounds` or without a track or position are paged;
 *                   the others return all matches at once.
 *     responses:
 *       200:
 *         description: Matching photos
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 candidates:
 *                   type: array
 *                   items:
 *                     $ref: '#/components/schemas/AssetPluginCandidate'
 *                 hasMore:
 *                   type: boolean
 *                   description: Whether there are more pages.
 *                 existingExternalRefs:
 *                   type: array
 *                   description: >
 *                     Provider and provider ID of each of the user's assets that came from a plugin, to recognize
 *                     plugin candidates that are already in wanderer.
 *                   items:
 *                     type: object
 *                     properties:
 *                       provider:
 *                         type: string
 *                       id:
 *                         type: string
 *       400:
 *         description: Bad Request
 *       401:
 *         description: Unauthorized
 *       403:
 *         description: Insufficient permissions for the trail, waypoint or summit log
 *       404:
 *         description: Trail, waypoint or summit log not found
 *       500:
 *         description: Internal Server Error
 */
export async function POST(event: RequestEvent) {
    try {
        const data = LibraryRequestSchema.parse(await event.request.json().catch(() => ({})));
        const response = await event.locals.pb.send("/assets/library", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
            fetch: event.fetch,
        });

        return json(response);
    } catch (e: any) {
        return handleError(e);
    }
}
