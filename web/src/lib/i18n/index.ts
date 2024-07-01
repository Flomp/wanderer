import { browser } from '$app/environment'
import { currentUser } from '$lib/stores/user_store'
import { init, register } from 'svelte-i18n'
import { get } from 'svelte/store'

const defaultLocale = 'en'

register('en', () => import('./locales/en.json'))
register('de', () => import('./locales/de.json'))
register('fr', () => import('./locales/fr.json'))
register('hu', () => import('./locales/hu.json'))
register('it', () => import('./locales/it.json'))
register('nl', () => import('./locales/nl.json'))
register('pl', () => import('./locales/pl.json'))
register('pt', () => import('./locales/pt.json'))
register('zh', () => import('./locales/zh.json'))

init({
    fallbackLocale: defaultLocale,
    initialLocale: browser ? get(currentUser)?.language ?? window.navigator.language : defaultLocale,
})