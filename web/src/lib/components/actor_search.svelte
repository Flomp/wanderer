<script lang="ts">
    import type { Actor } from "$lib/models/activitypub/actor";
    import type { User } from "$lib/models/user";
    import { actors_search } from "$lib/stores/actor_store";
    import { show_toast } from "$lib/stores/toast_store.svelte";
    import { users_search } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import Search, { type SearchItem } from "./base/search.svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        label?: string;
        value?: string;
        includeSelf?: boolean;
        clearAfterSelect?: boolean;
        onclear?: () => void;
        onclick?: (item: SearchItem) => void;
    }

    let {
        label = "",
        value = $bindable(""),
        includeSelf = true,
        clearAfterSelect = true,
        onclear,
        onclick,
    }: Props = $props();

    let searchItems: SearchItem[] = $state([]);

    async function updateActors(q: string) {
        if (!q.length) {
            searchItems = [];
            onclear?.();
            return;
        }
        try {
            const actors: Actor[] = await actors_search(q, includeSelf);
            searchItems = actors.map((a) => ({
                text: a.preferred_username!,
                description: `@${a.username}${a.isLocal ? "" : "@" + a.domain}`,
                value: a,
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
    onupdate={(q) => updateActors(q)}
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
            src={item.value.icon ||
                `https://api.dicebear.com/7.x/initials/svg?seed=${item.value.username}&backgroundType=gradientLinear`}
            alt="avatar"
        />
    {/snippet}
</Search>
