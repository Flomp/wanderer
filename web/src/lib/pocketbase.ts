import { browser } from '$app/environment';
import { env } from '$env/dynamic/public'
import PocketBase from 'pocketbase'

let pb: PocketBase;

export function getPb(): PocketBase {
  if(!browser) {
    throw Error("Only available client side!")
  }
  if (!pb) {
    pb = new PocketBase(env.PUBLIC_POCKETBASE_URL);
  }
  return pb;
}