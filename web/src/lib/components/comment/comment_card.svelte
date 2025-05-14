<script lang="ts">
    import type { Comment } from "$lib/models/comment";
    import { formatTimeSince } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import { fade } from "svelte/transition";
    import TextField from "../base/text_field.svelte";

    interface Props {
        comment: Comment;
        mode?: "show" | "edit";
        ondelete?: (comment: Comment) => void;
        onedit?: (data: { comment: Comment; text: string }) => void;
    }

    let { comment, mode = "show", ondelete, onedit }: Props = $props();

    let editing: boolean = $state(false);

    let editedComment = $state(comment.text);

    $effect(() => {
        editedComment = comment.text;
    });

    let avatarSrc = $derived(
        comment.expand?.author.icon
            ? comment.expand?.author.icon
            : `https://api.dicebear.com/7.x/initials/svg?seed=${comment.expand?.author.username ?? ""}&backgroundType=gradientLinear`,
    );

    let timeSince = $derived(formatTimeSince(new Date(comment.created ?? "")));

    function deleteComment() {
        ondelete?.(comment);
    }

    function toggleEdit() {
        if (editing) {
            onedit?.({ comment: comment, text: editedComment });
        }
        editing = !editing;
    }
</script>

<div
    class="flex gap-4 items-start"
    in:fade={{ duration: 150 }}
    out:fade={{ duration: 150 }}
>
    <a
        href="/profile/@{comment.expand?.author.username.toLowerCase()}"
        class="text-sm font-semibold shrink-0"
    >
        <img
            class="rounded-full w-10 aspect-square"
            src={avatarSrc}
            alt="avatar"
        />
    </a>
    <div>
        <div class="flex items-center">
            <p class="">
                <a
                    href="/profile/@{comment.expand?.author.username.toLowerCase()}{comment
                        .expand?.author.isLocal
                        ? ''
                        : '@' + comment.expand?.author.domain}"
                    class="text-sm font-semibold"
                    >{comment.expand?.author.username}</a
                >
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
