import {
    PluginInstanceCreateSchema,
} from "$lib/models/api/plugin_instance_schema";
import type { PluginInstance } from "$lib/models/plugin_instance";
import { Collection, handlePluginError as handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

/**
 * @swagger
 * /api/v1/plugin-instance:
 *   get:
 *     summary: List plugin instances
 *     description: Retrieves the authenticated user's configured plugin instances.
 *     tags:
 *       - Plugins
 *     parameters:
 *       - in: query
 *         name: page
 *         schema:
 *           type: integer
 *       - in: query
 *         name: perPage
 *         schema:
 *           type: integer
 *       - in: query
 *         name: sort
 *         schema:
 *           type: string
 *       - in: query
 *         name: filter
 *         schema:
 *           type: string
 *     responses:
 *       200:
 *         description: ListResult<PluginInstance>
 *       400:
 *         description: Bad Request
 *       401:
 *         description: Unauthorized
 *       500:
 *         description: Internal Server Error
 */
export async function GET(event: RequestEvent) {
    try {
        const r = await list<PluginInstance>(event, Collection.plugin_instances);
        return json(r);
    } catch (e) {
        return handleError(e);
    }
}

/**
 * @swagger
 * /api/v1/plugin-instance:
 *   put:
 *     summary: Create plugin instance
 *     description: Creates an owner-bound plugin instance through the host. Direct PocketBase collection creates return 403.
 *     tags:
 *       - Plugins
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - plugin_id
 *             properties:
 *               plugin_id:
 *                 type: string
 *               enabled:
 *                 type: boolean
 *               auth:
 *                 type: object
 *                 additionalProperties:
 *                   type: string
 *                   nullable: true
 *                 description: Editable credentials only. Empty secrets preserve stored material; null removes a secret.
 *               config:
 *                 type: object
 *                 additionalProperties: true
 *               authContext:
 *                 type: string
 *     responses:
 *       200:
 *         description: Plugin instance created
 *       400:
 *         description: Bad Request
 *       401:
 *         description: Unauthorized
 *       409:
 *         description: An instance already exists (data.operation.code is plugin_instance_exists).
 *       500:
 *         description: Internal Server Error
 */
export async function PUT(event: RequestEvent) {
    try {
        const body = PluginInstanceCreateSchema.parse(await event.request.json());
        const r = await event.locals.pb.send("/plugins/instances", {
            method: "POST", body, fetch: event.fetch,
        });
        return json(r);
    } catch (e) {
        return handleError(e);
    }
}
