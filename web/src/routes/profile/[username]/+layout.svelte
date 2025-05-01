<script lang="ts">
    import { invalidateAll } from "$app/navigation";
    import { page } from "$app/state";
    import Button from "$lib/components/base/button.svelte";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ProfileShareModal from "$lib/components/profile/profile_share_modal.svelte";
    import { follows_create, follows_delete } from "$lib/stores/follow_store";
    import { currentUser } from "$lib/stores/user_store";
    import { _ } from "svelte-i18n";

    let { data, children } = $props();

    let profileShareModal: ProfileShareModal;

    let followLoading: boolean = $state(false);

    const profileLinks: DropdownItem[] = [
        {
            text: $_("profile"),
            value: `/profile/${page.params.username}`,
            icon: "user",
        },
        {
            text: $_("trail", { values: { n: 2 } }),
            value: `/profile/${page.params.username}/trails`,
            icon: "route",
        },
        {
            text: $_("statistics"),
            value: `/profile/${page.params.username}/stats`,
            icon: "chart-pie",
        },
    ];

    let activeIndex = $derived(
        profileLinks.findIndex((l) => l.value === page.url.pathname),
    );

    async function follow() {
        if (!$currentUser) {
            return;
        }
        followLoading = true;

        if (data.follow) {
            await follows_delete(data.follow);
        } else {
            await follows_create($currentUser.id, data.profile.acct);
        }
        await invalidateAll();
        followLoading = false;
    }
</script>

<div
    class="grid grid-cols-1 md:grid-cols-[356px_minmax(0,_1fr)] gap-6 max-w-6xl mx-4 md:mx-auto items-start"
>
    <div class="border border-input-border rounded-xl md:sticky top-8 md:ml-6">
        {#if data.profile}
            <div class="flex items-center gap-x-6 px-6 mt-6 mb-4">
                <img
                    class="rounded-full w-16 aspect-square overflow-hidden"
                    src={data.profile.icon ||
                        `https://api.dicebear.com/7.x/initials/svg?seed=${data.profile.username}&backgroundType=gradientLinear`}
                    alt="avatar"
                />
                <div>
                    <h4 class="text-2xl font-semibold col-start-2">
                        {data.profile.username}
                    </h4>
                    <p class="text-sm text-gray-500 mb-4">
                        {data.profile.acct}
                    </p>
                </div>
            </div>
        {/if}
        <div class="px-6 mb-4">
            <p class="text-sm">
                <span class="text-gray-500">{$_("joined")}:</span>
                {new Date(data.profile.createdAt ?? "").toLocaleDateString(
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
        <hr class="mb-4 border-input-border" />
        <div class="flex gap-x-6 items-center text-sm px-6">
            <a
                class:font-bold={page.url.pathname.endsWith("followers")}
                href="/profile/{data.profile.acct}/users/followers"
            >
                <p class="font-semibold">{data.profile.followers}</p>
                <p>{$_("followers")}</p>
            </a>
            <a
                class:font-bold={page.url.pathname.endsWith("following")}
                href="/profile/{data.profile.acct}/users/following"
            >
                <p class="font-semibold">{data.profile.following}</p>
                <p>{$_("following")}</p>
            </a>
            {#if !data.isOwnProfile}
                <div class="px-6 mt-2">
                    <Button
                        loading={followLoading}
                        disabled={followLoading}
                        primary={!data.follow}
                        secondary={!!data.follow}
                        icon={data.follow ? "check" : ""}
                        onclick={() => follow()}
                    >
                        {data.follow ? $_("following") : $_("follow")}</Button
                    >
                </div>
            {/if}
        </div>

        <hr class="mt-4 border-input-border" />

        <div class="mt-6 mb-4">
            {#each profileLinks as link, i}
                <a
                    class="block mx-2 px-4 py-3 my-1 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors rounded-md"
                    class:bg-menu-item-background-hover={i == activeIndex}
                    href={link.value}
                    ><i class="fa fa-{link.icon} mr-2"></i>{link.text}</a
                >
            {/each}
        </div>
        {#if data.isOwnProfile}
            <div class="px-6 mb-4 flex flex-col gap-2">
                <button
                    class="btn-secondary basis-full"
                    onclick={() => profileShareModal.openModal()}
                    >{$_("share-profile")}</button
                >
                <a
                    class="btn-secondary text-center basis-full"
                    href="/settings/profile">{$_("settings")}</a
                >
            </div>
            <ProfileShareModal bind:this={profileShareModal}
            ></ProfileShareModal>
        {/if}
    </div>
    <div class="md:mr-6">
        {@render children?.()}
    </div>
</div>
