import { browser } from '$app/environment';
import { getPb } from '$lib/pocketbase';
import { init, register } from 'svelte-i18n';

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
    initialLocale: browser ? getPb().authStore.record?.language ?? window.navigator.language : defaultLocale,
})