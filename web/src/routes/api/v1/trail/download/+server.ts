import { handleError } from "$lib/util/api_util";
import { type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.json();

    try {
        const response = await event.fetch(data.url);
        const blob = await response.blob();

        const contentType = 'determine the content type here';

        return new Response(blob, {
            headers: {
                'Content-Type': contentType,
            },
        });
    } catch (e) {
        throw handleError(e)
    }
}
