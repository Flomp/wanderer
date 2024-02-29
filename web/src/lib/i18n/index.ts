import { browser } from '$app/environment'
import { currentUser } from '$lib/stores/user_store'
import { init, register } from 'svelte-i18n'
import { get } from 'svelte/store'

const defaultLocale = 'en'

register('en', () => import('./locales/en.json'))
register('de', () => import('./locales/de.json'))

init({
    fallbackLocale: defaultLocale,
    initialLocale: browser ? get(currentUser)?.language ?? window.navigator.language : defaultLocale,
})