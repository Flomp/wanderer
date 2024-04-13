import type { User } from '$lib/models/user'
import { pb } from '$lib/pocketbase'
import { settings_show } from '$lib/stores/settings_store';
import { currentUser } from '$lib/stores/user_store'
import { get } from 'svelte/store';

pb.authStore.loadFromCookie(document.cookie)
pb.authStore.onChange(() => {
  currentUser.set(pb.authStore.model as User)
  document.cookie = pb.authStore.exportToCookie({ httpOnly: false })
}, true)