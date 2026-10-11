import { handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

const assetPluginActions = [
    "candidates",
    "check",
    "import",
    "import-to-target",
    "import-to-waypoint",
    "materialize-all",
    "delete-remote-assets",
    "remote-assets-summary",
    "repair-remote-assets",
] as const;

type AssetPluginAction = (typeof assetPluginActions)[number];
type ProxyMethod = "GET" | "POST";
type BodyMode = "none" | "required-json" | "optional-json";

const actionConfig = {
    candidates: { method: "POST", body: "required-json" },
    check: { method: "POST", body: "required-json" },
    import: { method: "POST", body: "required-json" },
    "import-to-target": { method: "POST", body: "required-json" },
    "import-to-waypoint": { method: "POST", body: "required-json" },
    "materialize-all": { method: "POST", body: "optional-json" },
    "delete-remote-assets": { method: "POST", body: "none" },
    "remote-assets-summary": { method: "GET", body: "none" },
    "repair-remote-assets": { method: "POST", body: "none" },
} satisfies Record<AssetPluginAction, { method: ProxyMethod; body: BodyMode }>;

const paramsSchema = z.object({
    plugin: z.string().min(1).max(128).regex(/^[a-zA-Z0-9._-]+$/),
    action: z.string().min(1).max(128).regex(/^[a-zA-Z0-9._-]+$/),
});

export async function GET(event: RequestEvent) {
    return proxyPluginAssetAction(event, "GET");
}

/**
 * @swagger
 * /api/v1/plugins/assets/{plugin}/candidates:
 *   post:
 *     summary: Search a photo library plugin
 *     description: >
 *       Asks the user's enabled instance of the photo library plugin `plugin`, such as Immich, for photos near a
 *       trail's track, near a position or inside a map area. Results come in batches. While `hasMore` is true, send
 *       the same search with the returned `cursorId` to get the next batch. When `restartRequired` is true, the cursor
 *       has expired or no longer matches the search, and the search has to start again without `cursorId`.
 *     tags:
 *       - Plugins
 *     parameters:
 *       - in: path
 *         name: plugin
 *         required: true
 *         schema:
 *           type: string
 *         description: Plugin ID
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               trailId:
 *                 type: string
 *                 description: Trail to search along. The user must be its author.
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
 *                 description: Start of the time window to search.
 *               takenBefore:
 *                 type: string
 *                 format: date-time
 *                 description: End of the time window to search.
 *               doubleRadius:
 *                 type: boolean
 *                 description: Asks the plugin to search a wider area.
 *               cursorId:
 *                 type: string
 *                 description: Cursor from the previous response, to get the next batch.
 *     responses:
 *       200:
 *         description: A batch of matching photos
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
 *                 cursorId:
 *                   type: string
 *                   description: Cursor for the next batch while `hasMore` is true.
 *                 restartRequired:
 *                   type: boolean
 *                 hasTimestamps:
 *                   type: boolean
 *                   description: Whether the track had timestamps the plugin could use, as reported by the plugin.
 *                 takenAfter:
 *                   type: string
 *                   description: Start of the time window the plugin searched, if it used one.
 *       400:
 *         description: Bad Request, no enabled instance of the plugin, or an error reported by the plugin
 *       401:
 *         description: Unauthorized
 *       403:
 *         description: Insufficient permissions for trail
 *       404:
 *         description: Not Found
 *       500:
 *         description: Internal Server Error
 */
export async function POST(event: RequestEvent) {
    return proxyPluginAssetAction(event, "POST");
}

async function proxyPluginAssetAction(event: RequestEvent, method: ProxyMethod) {
    try {
        const params = paramsSchema.parse(event.params);
        if (!isAssetPluginAction(params.action)) {
            return json({ message: "not_found" }, { status: 404 });
        }

        const config = actionConfig[params.action];

        if (config.method !== method) {
            return json(
                { message: "method_not_allowed" },
                { status: 405, headers: { Allow: config.method } },
            );
        }

        const sendOptions: {
            method: ProxyMethod;
            body?: string;
            headers?: Record<string, string>;
        } = { method: config.method };

        if (config.body !== "none") {
            const data =
                config.body === "optional-json"
                    ? await event.request.json().catch(() => ({}))
                    : await event.request.json();
            sendOptions.body = JSON.stringify(data);
            sendOptions.headers = { "Content-Type": "application/json" };
        }

        const r = await event.locals.pb.send(
            `/plugins/assets/${encodeURIComponent(params.plugin)}/${encodeURIComponent(params.action)}`,
            sendOptions,
        );
        return json(r);
    } catch (e: any) {
        return handleError(e);
    }
}

function isAssetPluginAction(action: string): action is AssetPluginAction {
    return Object.prototype.hasOwnProperty.call(actionConfig, action);
}
