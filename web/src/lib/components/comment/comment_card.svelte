<script lang="ts">
    import type { Comment } from "$lib/models/comment";
    import { getFileURL } from "$lib/util/file_util";
    import { createEventDispatcher } from "svelte";
    import TextField from "../base/text_field.svelte";
    import { fade } from "svelte/transition";
    import { formatTimeSince } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";

    interface Props {
        comment: Comment;
        mode?: "show" | "edit";
    }

    let { comment, mode = "show" }: Props = $props();

    const dispatch = createEventDispatcher();

    let editing: boolean = $state(false);

    let editedComment = $state(comment.text);

    $effect(() => {
        editedComment = comment.text;
    });

    let avatarSrc = $derived(
        comment.expand?.author.avatar
            ? getFileURL(comment.expand.author, comment.expand.author.avatar)
            : `https://api.dicebear.com/7.x/initials/svg?seed=${comment.expand?.author.username ?? ""}&backgroundType=gradientLinear`,
    );

    let timeSince = $derived(formatTimeSince(new Date(comment.created ?? "")));

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
    class="flex gap-4 items-start"
    in:fade={{ duration: 150 }}
    out:fade={{ duration: 150 }}
>
    {#if comment.expand?.author.private}
        <img
            class="rounded-full w-10 aspect-square shrink-0"
            src={avatarSrc}
            alt="avatar"
        />
    {:else}
        <a
            href="/profile/{comment.expand?.author.id}"
            class="text-sm font-semibold shrink-0"
        >
            <img
                class="rounded-full w-10 aspect-square"
                src={avatarSrc}
                alt="avatar"
            />
        </a>
    {/if}
    <div>
        <div class="flex items-center">
            <p class="">
                {#if comment.expand?.author.private}
                    <span class="text-sm font-semibold"
                        >{comment.expand?.author.username}</span
                    >
                {:else}
                    <a
                        href="/profile/{comment.expand?.author.id}"
                        class="text-sm font-semibold"
                        >{comment.expand?.author.username}</a
                    >
                {/if}
                <span class="text-xs text-gray-500 ml-2"
                    >{$_(`n-${timeSince.unit}-ago`, {
                        values: { n: timeSince.value },
                    })}
                    {comment.updated != comment.created
                        ? `(${$_("edited")})`
                        : ""}</span
                >
            </p>
            {#if mode == "edit"}
                <button
                    aria-label="Edit comment"
                    type="button"
                    class="btn-icon ml-2"
                    style="font-size: 0.75rem;"
                    onclick={toggleEdit}
                    ><i class="fa fa-{editing ? 'check' : 'pen'}"></i></button
                >
                <button
                    aria-label="Delete comment"
                    type="button"
                    class="btn-icon text-xs"
                    style="font-size: 0.75rem;"
                    onclick={deleteComment}><i class="fa fa-trash"></i></button
                >
            {/if}
        </div>
        {#if editing}
            <TextField extraClasses="mt-2" bind:value={editedComment}
            ></TextField>
        {:else}
            <p class="whitespace-pre-wrap text-sm">{comment.text}</p>
        {/if}
    </div>
</div>
