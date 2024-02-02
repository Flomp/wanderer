import { pb } from '$lib/pocketbase'
import { currentUser } from '$lib/stores/user_store'

pb.authStore.loadFromCookie(document.cookie)
pb.authStore.onChange(() => {
  currentUser.set(pb.authStore.model)
  document.cookie = pb.authStore.exportToCookie({ httpOnly: false })
}, true)