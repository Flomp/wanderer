import { trails_show } from "$lib/stores/trail_store";
import { APIError } from "$lib/util/api_util";
import { gpx2trail } from "$lib/util/gpx_util";
import { error, type Load, type NumericRange } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    try {
        const trail = await trails_show(params.id!, params.handle, true, fetch);
        // Collect alternative trails from summit logs
        const alternativeTrails = [];
        for (const log of trail.expand?.summit_logs_via_trail ?? []) {
            if (log.expand?.gpx_data) {
                try {
                    const parsed = gpx2trail(log.expand.gpx_data, log.id);
                    // Mark as alternative for coloring in UI
                    parsed.trail.id = log.id;
                    parsed.trail.name = log.text || `Summit log ${log.id}`;
                    parsed.trail.isAlternative = true;
                    alternativeTrails.push(parsed.trail);
                } catch (e) {
                    console.warn("Failed to parse summit log GPX", log.id, e);
                }
            }
        }
        return { trail, alternativeTrails };
    } catch (e) {
        if (e instanceof APIError) {
            error(e.status as NumericRange<400, 599>, {
                message: e.status == 404 ? 'Not found' : e.message
            });
        }
        console.error(e);

    }
};