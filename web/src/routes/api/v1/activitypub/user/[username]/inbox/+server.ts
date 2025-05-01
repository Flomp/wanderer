// src/routes/inbox/+server.ts
import type { UserAnonymous } from '$lib/models/user';
import { error, json } from '@sveltejs/kit';
import type { RequestEvent } from './$types';
import { ClientResponseError } from 'pocketbase';

export async function GET(event: RequestEvent) {
    const activity = await event.request.json();

    let { username } = event.params;

    if (username?.startsWith('@')) {
        username = username.substring(1)
    }
    try {
        const user: UserAnonymous = await event.locals.pb.collection('users_anonymous').getFirstListItem(`username="${username}"`);
        const key: { public_key: string } = await event.locals.pb.send(`/public/user/${user.username}/key`, { method: "GET", fetch: event.fetch })

        // Signature verification step would go here (see below)
        // Check if the signature matches the one in the incoming request

        // For simplicity, we accept the activity without verification here
        console.log("Received Activity:", activity);

        // Process the activity (e.g., save to DB, follow user, create post)
        // For example, if it's a follow activity:
        if (activity.type === 'Follow') {
            // Handle following logic here (e.g., add to "followers" list)
        }

        return new Response(null, { status: 202 }); // Accepted
    } catch (e) {
        if (e instanceof ClientResponseError) {
            return error(e.status, { message: e.message });
        }
        throw e;
    }
};
