import { PUBLIC_MEILISEARCH_URL } from '$env/static/public'
import { PUBLIC_MEILISEARCH_API_KEY } from '$env/static/public'

import { MeiliSearch } from 'meilisearch'

export function createInstance() {
    return new MeiliSearch({ host: PUBLIC_MEILISEARCH_URL, apiKey: PUBLIC_MEILISEARCH_API_KEY })
}

export const ms = createInstance()