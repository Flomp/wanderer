import { env } from '$env/dynamic/public'
import PocketBase from 'pocketbase'

export function createInstance() {
  return new PocketBase(env.PUBLIC_POCKETBASE_URL)
}

export const pb = createInstance()