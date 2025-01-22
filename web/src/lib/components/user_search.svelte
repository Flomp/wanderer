<script lang="ts">
    import type { User } from "$lib/models/user";
    import { show_toast } from "$lib/stores/toast_store";
    import { users_search } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import { createEventDispatcher, tick } from "svelte";
    import Search, { type SearchItem } from "./base/search.svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        label?: string;
        value?: string;
        includeSelf?: boolean;
        clearAfterSelect?: boolean;
    }

    let {
        label = "",
        value = $bindable(""),
        includeSelf = true,
        clearAfterSelect = true
    }: Props = $props();

    let searchItems: SearchItem[] = $state([]);

    const dispatch = createEventDispatcher();

    async function updateUsers(q: string) {
        if (!q.length) {
            searchItems = [];
            dispatch("clear");
            return;
        }
        try {
            const users: User[] = await users_search(q, includeSelf);
            searchItems = users.map((u) => ({
                text: u.username!,
                value: u,
                icon: "user",
            }));
        } catch (e) {
            console.error(e);
            show_toast({
                type: "error",
                icon: "close",
                text: "Error during search",
            });
        }
    }

    function onClick(e: CustomEvent<SearchItem>) {
        value = e.detail.value.username ?? value;

        dispatch("click", e.detail);
        searchItems = [];
    }
</script>

<Search
    on:update={(e) => updateUsers(e.detail)}
    on:click={(e) => onClick(e)}
    placeholder={`${$_("username")}...`}
    items={searchItems}
    {clearAfterSelect}
    {label}
    bind:value
>
    <!-- @migration-task: migrate this slot by hand, `item-header` is an invalid identifier -->
    <img
        slot="item-header"
        let:item
        class="rounded-full w-8 aspect-square mr-2"
        src={getFileURL(item.value, item.value.avatar) ||
            `https://api.dicebear.com/7.x/initials/svg?seed=${item.value.username}&backgroundType=gradientLinear`}
        alt="avatar"
    />
</Search>
