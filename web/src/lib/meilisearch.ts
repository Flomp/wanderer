import { env } from '$env/dynamic/private'

import { MeiliSearch } from 'meilisearch'
import { pb } from './pocketbase'

export function createInstance() {
    if (pb.authStore.model) {
        return new MeiliSearch({ host: env.MEILI_URL, apiKey: pb.authStore.model.token })
    }
    return new MeiliSearch({ host: env.MEILI_URL, apiKey: env.MEILI_API_TOKEN })
}

export function regenerateInstance() {
    ms = createInstance();
}

export let ms = createInstance()