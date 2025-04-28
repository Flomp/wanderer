import { error, json, RequestEvent } from '@sveltejs/kit';


export async function GET(event: RequestEvent) {
    const resource = event.url.searchParams.get('resource');

    if (!resource || !resource.startsWith('acct:')) {
        return error(400, { message: 'Bad request' });

    }

    const [usernamePart, domainPart] = resource.replace('acct:', '').split('@');

    try {
        const user = await (event.locals as any).pb.collection('users_anonymous').getFirstListItem(`username="${usernamePart}"`);

        if (!user) {
            return error(404, { message: 'Not found' });
        }

        return json({
            "subject": `acct:${user.username}@${event.url.host}`,
            "links": [
                {
                    "rel": "self",
                    "type": "application/activity+json",
                    "href": `${event.url.origin}/users/@${user.username}`
                },
                {
                    "rel": "profile",
                    "type": "text/html",
                    "href": `${event.url.origin}/profile/${user.id}`
                }
            ]
        });
    } catch (err) {
        console.error('Error fetching user in WebFinger:', err);
        return error(500, { message: 'Internal Server Error' });

    }
}