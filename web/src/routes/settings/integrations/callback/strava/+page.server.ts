import { integrations_index, integrations_update } from "$lib/stores/integration_store";
import { error, redirect, type RequestEvent, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ url, fetch }) => {
    const oauthError = url.searchParams.get('error');
    if (oauthError) {
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

    const integrations = await integrations_index(fetch);

    if (!integrations.length || !integrations[0].strava) {
        return error(400, {
            message: "Missing integration record"
        });
    }

    const integration = integrations[0]

    const tokenResponse = await fetch('https://www.strava.com/oauth/token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            client_id: integration.strava?.clientId,
            client_secret: integration.strava?.clientSecret,
            code,
            grant_type: 'authorization_code'
        })
    });

    if (!tokenResponse.ok) {
        const r = await tokenResponse.json()
        console.error(r)
        return error(500, 'Failed to get access token');
    }

    const { access_token, refresh_token, expires_at } = await tokenResponse.json();

    integration.strava!.accessToken = access_token
    integration.strava!.refreshToken = refresh_token
    integration.strava!.expiresAt = expires_at
    integration.strava!.active = true

    await integrations_update(integration, fetch);

    return redirect(302, '/settings/integrations')
}