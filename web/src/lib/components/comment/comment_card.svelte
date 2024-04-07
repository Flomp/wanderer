<script lang="ts">
    import type { Comment } from "$lib/models/comment";
    import { getFileURL } from "$lib/util/file_util";
    import { createEventDispatcher } from "svelte";
    import TextField from "../base/text_field.svelte";
    import { fade } from "svelte/transition";
    import { formatTimeSince } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";

    export let comment: Comment;
    export let mode: "show" | "edit" = "show";

    const dispatch = createEventDispatcher();

    let editing: boolean = false;

    let editedComment = comment.text;

    $: avatarSrc = comment.expand?.author.avatar
        ? getFileURL(comment.expand.author, comment.expand.author.avatar)
        : `https://api.dicebear.com/7.x/initials/svg?seed=${comment.expand?.author.username ?? ""}&backgroundType=gradientLinear`;

    $: timeSince = formatTimeSince(new Date(comment.created ?? ""));

    function deleteComment() {
        dispatch("delete", comment);
    }

    function toggleEdit() {
        if (editing) {
            dispatch("edit", { comment: comment, text: editedComment });
        }
        editing = !editing;
    }
</script>

<div
    class="flex gap-4 items-center"
    in:fade={{ duration: 150 }}
    out:fade={{ duration: 150 }}
>
    <img class="rounded-full w-10 aspect-square" src={avatarSrc} alt="avatar" />
    <div>
        <div class="flex items-center">
            <p class="">
                <span class="text-sm font-semibold"
                    >{comment.expand?.author.username}</span
                >
                <span class="text-xs text-gray-500 ml-2"
                    >{$_(`n-${timeSince.unit}-ago`, {
                        values: { n: timeSince.value },
                    })}</span
                >
            </p>
            {#if mode == "edit"}
                <button
                    type="button"
                    class="btn-icon ml-2"
                    style="font-size: 0.75rem;"
                    on:click={toggleEdit}
                    ><i class="fa fa-{editing ? 'check' : 'pen'}"></i></button
                >
                <button
                    type="button"
                    class="btn-icon text-xs"
                    style="font-size: 0.75rem;"
                    on:click={deleteComment}><i class="fa fa-trash"></i></button
                >
            {/if}
        </div>
        {#if editing}
            <TextField extraClasses="mt-2" bind:value={editedComment}
            ></TextField>
        {:else}
            <p>{comment.text}</p>
        {/if}
    </div>
</div>
