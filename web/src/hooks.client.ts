import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { currentUser } from '$lib/stores/user_store';

pb.authStore.loadFromCookie(document.cookie)
pb.authStore.onChange(() => {
  currentUser.set(pb.authStore.model as User)
  document.cookie = pb.authStore.exportToCookie({ httpOnly: false, secure: location.protocol === "https:" })
}, true)