import { APIError } from "$lib/util/api_util";

export async function profile_show(username: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let r = await f('/api/v1/profile/' + username, {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response = await r.json()
    return response;

}