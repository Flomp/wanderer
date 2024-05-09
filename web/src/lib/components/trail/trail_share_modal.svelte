<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import Button from "../base/button.svelte";
    import Search, { type SearchItem } from "../base/search.svelte";
    import { users_search } from "$lib/stores/user_store";
    import type { User } from "$lib/models/user";
    import { show_toast } from "$lib/stores/toast_store";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    let copyButtonText = $_("copy-link");

    let searchItems: SearchItem[] = [];

    async function updateUsers(q: string) {
        try {
            const users: User[] = await users_search(q);
            searchItems = users.map((u) => ({
                text: u.username!,
                value: u.id,
                icon: "user"
            }))
        } catch (e) {
            console.error(e);
            show_toast({
                type: "error",
                icon: "close",
                text: "Error during search",
            });
        }
    }

    function copyURLToClipboard() {
        navigator.clipboard.writeText(window.location.href);

        copyButtonText = $_("link-copied");
        setTimeout(() => (copyButtonText = $_("copy-link")), 3000);
    }

    function close() {
        searchItems = [];
        dispatch("save");
        closeModal!();
    }
</script>

<Modal
    id="share-modal"
    title="Share this trail"
    size="max-w-lg"
    bind:openModal
    bind:closeModal
>
    <div slot="content">
        <Search
            on:update={(e) => updateUsers(e.detail)}
            placeholder={`${$_("username")}`}
            items={searchItems}
        ></Search>
    </div>
    <div slot="footer" class="flex justify-between items-center gap-4">
        <Button
            secondary={true}
            disabled={copyButtonText == $_("link-copied")}
            on:click={copyURLToClipboard}
        >
            <i class="fa fa-link mr-2"></i>
            {copyButtonText}
        </Button>
        <button class="btn-primary" on:click={close}>{$_("close")}</button>
    </div></Modal
>
