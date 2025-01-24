<script lang="ts">
    import type { User } from "$lib/models/user";
    import { show_toast } from "$lib/stores/toast_store";
    import { users_search } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import Search, { type SearchItem } from "./base/search.svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        label?: string;
        value?: string;
        includeSelf?: boolean;
        clearAfterSelect?: boolean;
        onclear?: () => void
        onclick?: (item: SearchItem) => void
    }

    let {
        label = "",
        value = $bindable(""),
        includeSelf = true,
        clearAfterSelect = true,
        onclear,
        onclick
    }: Props = $props();

    let searchItems: SearchItem[] = $state([]);

    async function updateUsers(q: string) {
        if (!q.length) {
            searchItems = [];
            onclear?.();
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

    function onClick(item: SearchItem) {
        value = item.value.username ?? value;

        onclick?.(item);
        searchItems = [];
    }
</script>

<Search
    onupdate={(q) => updateUsers(q)}
    onclick={(item) => onClick(item)}
    placeholder={`${$_("username")}...`}
    items={searchItems}
    {clearAfterSelect}
    {label}
    bind:value
>
    {#snippet prepend({ item })}
        <img
            
            
            class="rounded-full w-8 aspect-square mr-2"
            src={getFileURL(item.value, item.value.avatar) ||
                `https://api.dicebear.com/7.x/initials/svg?seed=${item.value.username}&backgroundType=gradientLinear`}
            alt="avatar"
        />
    {/snippet}
</Search>
