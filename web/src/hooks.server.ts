import { env } from '$env/dynamic/private'
import { env as envPub } from '$env/dynamic/public'
import type { Settings } from '$lib/models/settings'

import { pb } from '$lib/pocketbase'
import { isRouteProtected } from '$lib/util/authorization_util'
import { json, redirect, text, type Handle } from '@sveltejs/kit'
import { sequence } from '@sveltejs/kit/hooks'
import { MeiliSearch } from 'meilisearch'
import { locale } from 'svelte-i18n'


function csrf(allowedPaths: string[]): Handle {
  return async ({ event, resolve }) => {
    const { request, url } = event;
    const forbidden =
      isFormContentType(request) &&
      (request.method === "POST" ||
        request.method === "PUT" ||
        request.method === "PATCH" ||
        request.method === "DELETE") &&
      request.headers.get("origin") !== url.origin &&
      !allowedPaths.some(p => url.pathname.startsWith(p));

    if (forbidden) {
      const message = `Cross-site ${request.method} form submissions are forbidden`;
      if (request.headers.get("accept") === "application/json") {
        return json({ message }, { status: 403 });
      }
      return text(message, { status: 403 });
    }

    return resolve(event);
  };
}

function isContentType(request: Request, ...types: string[]) {
  const type = request.headers.get("content-type")?.split(";", 1)[0].trim() ?? "";
  return types.includes(type.toLowerCase());
}
function isFormContentType(request: Request) {
  return isContentType(
    request,
    "application/x-www-form-urlencoded",
    "multipart/form-data",
    "text/plain",
  );
}


const auth: Handle = async ({ event, resolve }) => {
  // load the store data from the request cookie string
  pb.authStore.loadFromCookie(event.request.headers.get('cookie') || '')

  const url = new URL(event.request.url);


  // validate the user existence and if the path is acceesible
  if (!pb.authStore.model && isRouteProtected(url.pathname)) {
    throw redirect(302, '/login?r=' + url.pathname);
  } else if (pb.authStore.model && url.pathname === "/login") {
    throw redirect(302, '/');
  } else if (envPub.PUBLIC_DISABLE_SIGNUP === "true" && url.pathname === "/register") {
    throw redirect(302, '/');
  }

  try {
    // get an up-to-date auth store state by verifying and refreshing the loaded auth model (if any)
    if (pb.authStore.isValid) {
      await pb.collection('users').authRefresh({ requestKey: null })
    }
  } catch (_) {
    // clear the auth store on failed refresh
    pb.authStore.clear()
  }

  let meiliApiKey: string = "";
  let settings: Settings | undefined;
  if (pb.authStore.model) {
    meiliApiKey = pb.authStore.model.token
    settings = await pb.collection('settings').getFirstListItem<Settings>(`user="${pb.authStore.model.id}"`, { requestKey: null })
  } else {
    const r = await event.fetch(pb.buildUrl("/public/search/token"));
    const response = await r.json();
    meiliApiKey = response.token;
  }
  const ms = new MeiliSearch({ host: env.MEILI_URL, apiKey: meiliApiKey });

  event.locals.ms = ms
  event.locals.pb = pb
  event.locals.user = pb.authStore.model
  event.locals.settings = settings

  const lang = settings?.language ?? event.request.headers.get('accept-language')?.split(',')[0]

  if (lang) {
    locale.set(lang)
    if (pb.authStore.model) {
      pb.authStore.model!.language = lang;
    }
  }

  const response = await resolve(event)

  // send back the default 'pb_auth' cookie to the client with the latest store state
  response.headers.set(
    'set-cookie',
    pb.authStore.exportToCookie({ httpOnly: false, secure: event.url.protocol === "https:" || event.url.hostname === "localhost", sameSite: "none" })
  )

  return response
}

const removeLinkFromHeaders: Handle =
  async ({ event, resolve }) => {
    const response = await resolve(event);
    response.headers.delete('link');
    return response;
  }


export const handle = sequence(csrf(['/api/v1']), auth, removeLinkFromHeaders)