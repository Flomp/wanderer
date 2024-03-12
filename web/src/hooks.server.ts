import { env } from '$env/dynamic/private'
import { pb } from '$lib/pocketbase'
import { isRouteProtected } from '$lib/util/authorization_util'
import { redirect, type Handle } from '@sveltejs/kit'
import { MeiliSearch } from 'meilisearch'
import { locale } from 'svelte-i18n'

export const handle: Handle = async ({ event, resolve }) => {
  // load the store data from the request cookie string
  pb.authStore.loadFromCookie(event.request.headers.get('cookie') || '')

  const url = new URL(event.request.url);

  // validate the user existence and if the path is acceesible
  if (!pb.authStore.model && isRouteProtected(url.pathname)) {
    throw redirect(302, '/login?r=' + url.pathname);
  } else if (pb.authStore.model && url.pathname === "/login") {
    throw redirect(302, '/');
  }

  try {
    // get an up-to-date auth store state by verifying and refreshing the loaded auth model (if any)
    if (pb.authStore.isValid) {
      await pb.collection('users').authRefresh()
    }
  } catch (_) {
    // clear the auth store on failed refresh
    pb.authStore.clear()
  }

  let meiliApiKey: string = "";
  if (pb.authStore.model) {
    meiliApiKey = pb.authStore.model.token
  } else {
    const r = await event.fetch(pb.buildUrl("/public/search/token"));
    const response = await r.json();
    meiliApiKey = response.token;    
  }
  const ms = new MeiliSearch({ host: env.MEILI_URL, apiKey: meiliApiKey });

  event.locals.ms = ms
  event.locals.pb = pb
  event.locals.user = pb.authStore.model

  const lang = pb.authStore.model?.language ?? event.request.headers.get('accept-language')?.split(',')[0]

  if (lang) {
    locale.set(lang)
  }

  const response = await resolve(event)

  // send back the default 'pb_auth' cookie to the client with the latest store state
  response.headers.set(
    'set-cookie',
    pb.authStore.exportToCookie({ httpOnly: false })
  )

  return response
}