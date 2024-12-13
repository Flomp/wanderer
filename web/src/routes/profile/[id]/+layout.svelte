<script lang="ts">
    import { page } from "$app/stores";
    import type { SelectItem } from "$lib/components/base/select.svelte";
    import { getFileURL } from "$lib/util/file_util";
    import { onMount } from "svelte";

    export let data;

    const profileLinks: SelectItem[] = [
        { text: "Profile", value: `/profile/${$page.params.id}` },
        { text: "Trails", value: `/profile/${$page.params.id}/trails` },
        { text: "Activities", value: `/profile/${$page.params.id}/activities` },
        { text: "Stats", value: `/profile/${$page.params.id}/stats` },
    ];

    $: activeIndex = profileLinks.findIndex(
        (l) => l.value === $page.url.pathname,
    );
</script>

<div
    class="grid grid-cols-1 lg:grid-cols-[356px_minmax(0,_1fr)] gap-4 max-w-6xl mx-auto items-start"
>
    <div class="border border-input-border rounded-xl space-y-8">
        {#if data.user}
            <div class="flex items-center gap-x-6 px-6 mt-6">
                <img
                    class="rounded-full w-16 aspect-square overflow-hidden"
                    src={getFileURL(data.user, data.user.avatar) ||
                        `https://api.dicebear.com/7.x/initials/svg?seed=${data.user.username}&backgroundType=gradientLinear`}
                    alt="avatar"
                />
                <div>
                    <h4 class="text-2xl font-semibold col-start-2">
                        {data.user.username}
                    </h4>
                    <p class="text-sm">
                        <span class="text-gray-500">Joined:</span>
                        {new Date(data.user.created ?? "").toLocaleDateString(
                            undefined,
                            {
                                month: "2-digit",
                                day: "2-digit",
                                year: "numeric",
                                timeZone: "UTC",
                            },
                        )}
                    </p>
                </div>
            </div>
        {/if}
        <div class="flex gap-x-6 text-sm px-6">
            <div>
                <p class="font-semibold">0</p>
                <p>Followers</p>
            </div>
            <div>
                <p class="font-semibold">0</p>
                <p>Following</p>
            </div>
        </div>
        <div>
            {#each profileLinks as link, i}
                <a
                    class="block mx-2 px-4 py-3 my-1 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors rounded-md"
                    class:bg-menu-item-background-hover={i == activeIndex}
                    href={link.value}>{link.text}</a
                >
            {/each}
        </div>
    </div>
    <slot></slot>
</div>
