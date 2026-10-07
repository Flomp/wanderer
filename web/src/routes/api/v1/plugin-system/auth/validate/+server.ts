import { PluginSessionAuthValidateSchema } from "$lib/models/api/plugin_instance_schema";
import { handlePluginError as handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

/**
 * @swagger
 * /api/v1/plugin-system/auth/validate:
 *   post:
 *     summary: Validate plugin session auth
 *     description: Validates draft session credentials and effective configuration without saving either.
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
 *               - auth
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
 *               auth:
 *                 type: object
 *                 additionalProperties:
 *                   type: string
 *                   nullable: true
 *                 description: Draft credential patch. Empty secrets preserve stored values; null removes a secret for this validation.
 *               config:
 *                 type: object
 *                 description: Unsaved instance config draft, merged with installed defaults and stored overrides.
 *                 additionalProperties: true
 *     responses:
 *       200:
 *         description: Auth validation result
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 ok:
 *                   type: boolean
 *                 authContext:
 *                   type: string
 *       400:
 *         description: Bad Request
 *       401:
 *         description: Unauthorized
 *       409:
 *         description: Busy or connection conflict; inspect data.operation.code and retry only deliberately.
 *       500:
 *         description: Internal Server Error
 */
export async function POST(event: RequestEvent) {
    try {
        const body = PluginSessionAuthValidateSchema.parse(await event.request.json());
        const r = await event.locals.pb.send("/plugins/auth/validate", {
            method: "POST",
            body,
            fetch: event.fetch,
        });
        return json(r);
    } catch (e) {
        return handleError(e);
    }
}
