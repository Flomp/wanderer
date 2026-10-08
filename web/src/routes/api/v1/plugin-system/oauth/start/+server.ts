import { handlePluginError as handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

/**
 * @swagger
 * /api/v1/plugin-system/oauth/start:
 *   post:
 *     summary: Start plugin OAuth flow
 *     description: Creates an OAuth authorization URL for a plugin instance.
 *     tags:
 *       - Plugins
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - pluginId
 *               - instanceId
 *               - redirectUri
 *             properties:
 *               pluginId:
 *                 type: string
 *               instanceId:
 *                 type: string
 *               expected_connection_generation:
 *                 type: string
 *                 description: Required for an existing instance; use the latest owner-read connection_generation.
 *               authContext:
 *                 type: string
 *               activate_on_success:
 *                 type: boolean
 *                 default: false
 *                 description: True requests activation after successful authorization; false preserves activation. Later explicit enable/disable overrides the request.
 *               redirectUri:
 *                 type: string
 *                 format: uri
 *     responses:
 *       200:
 *         description: OAuth authorization URL
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 url:
 *                   type: string
 *                   format: uri
 *                 state:
 *                   type: string
 *                 instanceId:
 *                   type: string
 *                 connection_generation:
 *                   type: string
 *                   description: The generation of this flow's own commit, unchanged by start.
 *                 updated:
 *                   type: string
 *                   description: The timestamp of this flow's own commit; later lookups may already describe another write.
 *       400:
 *         description: Bad Request
 *       401:
 *         description: Unauthorized
 *       409:
 *         description: Connection conflict; reload and review before beginning a new flow.
 *       500:
 *         description: Internal Server Error
 */
export async function POST(event: RequestEvent) {
    try {
        const body = await event.request.json();
        const r = await event.locals.pb.send("/plugins/oauth/start", {
            method: "POST",
            body,
            fetch: event.fetch,
        });
        return json(r);
    } catch (e) {
        return handleError(e);
    }
}
