import { env } from '$env/dynamic/public'

import { MeiliSearch } from 'meilisearch'
import { pb } from './pocketbase'

export function createInstance() {    
    if (pb.authStore.model) {
        return new MeiliSearch({ host: env.PUBLIC_MEILISEARCH_URL, apiKey: pb.authStore.model.token })
    }
    return new MeiliSearch({ host: env.PUBLIC_MEILISEARCH_URL, apiKey: env.PUBLIC_MEILISEARCH_API_TOKEN })
}

export function regenerateInstance() {
    ms = createInstance();
}

export let ms = createInstance()