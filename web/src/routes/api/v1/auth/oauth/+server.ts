import { env } from "$env/dynamic/public";
import { env as private_env } from "$env/dynamic/private";

import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

const redirectURL = private_env.ORIGIN + "/login/redirect"

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('users').listAuthMethods();

        for (const provider of r.authProviders) {
            const imageURL = `${env.PUBLIC_POCKETBASE_URL}/_/images/oauth2/${provider.name}.svg`
            provider['img' as keyof typeof provider] = await imageUrlToBase64(imageURL, event.fetch);
            provider['url' as keyof typeof provider] = `${provider.authUrl}${redirectURL}`
        }
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }

}

export async function POST(event: RequestEvent) {
    const data = await event.request.json();
    try {
        const r = await pb.collection('users').authWithOAuth2Code(
            data.name,
            data.code,
            data.codeVerifier,
            redirectURL,
        )
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }

}

async function imageUrlToBase64(url: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response>) {
    try {
        const response = await f(url);
        if (!response.ok) {
            throw new Error(`Failed to fetch image: ${response.status} ${response.statusText}`);
        }
        const arrayBuffer = await response.arrayBuffer();
        const buffer = Buffer.from(arrayBuffer);
        const base64DataUrl = `data:${response.headers.get('content-type') || 'image/png'};base64,${buffer.toString('base64')}`;
        return base64DataUrl;
    } catch (error) {
        console.error('Error fetching image:', error);
        throw error;
    }
}