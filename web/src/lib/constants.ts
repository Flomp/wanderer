import PocketBase from 'pocketbase';
import { env } from "$env/dynamic/public";

const pb = new PocketBase(env.PUBLIC_DB_HOST);

export { pb }