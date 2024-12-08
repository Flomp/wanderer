<script lang="ts">
    import { _ } from "svelte-i18n";
    export let data;
</script>

<div class="mx-8 space-y-6">
    <div class="space-y-4">
        <h4 class="text-xl font-semibold">
            {$_("about")}
            {data.user.username}
            {#if data.isOwnProfile && data.user.bio?.length}
                <a class="ml-4" href="/settings/profile"
                    ><i class="fa fa-pen text-base"></i></a
                >
            {/if}
        </h4>
        {#if data.user.bio?.length}
            <p class="whitespace-pre-wrap text-sm">{data.user.bio}</p>
        {:else if data.isOwnProfile}
            <a class="btn-primary inline-block" href="/settings/profile"
                >+ Add Bio</a
            >
        {/if}
    </div>
    <div>
        <h4 class="text-xl font-semibold">
            {$_("list", { values: { n: 2 } })}
        </h4>
    </div>
    <div>
        <h4 class="text-xl font-semibold">Timeline</h4>
    </div>
</div>
