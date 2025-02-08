import { pb } from "$lib/pocketbase";
import { integrations_index, integrations_update } from "$lib/stores/integration_store";
import { error, redirect, type RequestEvent, type ServerLoad } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export const load: ServerLoad = async ({ url, fetch }) => {
    const oauthError = url.searchParams.get('error');
    if (oauthError) {
        // user cancelled
        if (oauthError == "access_denied") {
            return redirect(302, '/settings/integrations')
        }
        return error(400, {
            message: oauthError
        });
    }
    const code = url.searchParams.get('code');
    if (!code) {
        return error(400, {
            message: "No code provided"
        });
    }

    try {
        await pb.send("/integration/strava/token", {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                code,
                grant_type: 'authorization_code'
            })
        });
    } catch (e) {
        console.error(e)

        if (e instanceof ClientResponseError) {
            return error(e.status, e.message);

        }
        throw e
    }
    return redirect(302, '/settings/integrations')
}