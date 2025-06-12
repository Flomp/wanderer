<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { TrailLike } from "$lib/models/trail_like";
    import {
        trail_like_create,
        trail_like_delete,
    } from "$lib/stores/trail_like_store";
    import { currentUser } from "$lib/stores/user_store";

    interface Props {
        trail: Trail;
        large?: boolean;
    }

    let { trail, large = false }: Props = $props();

    let likeCount = $state(trail.expand?.trail_like_via_trail?.length ?? 0);

    async function like(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        if (!$currentUser || !trail.id) {
            return;
        }

        if (trailLike) {
            await trail_like_delete(trailLike);
            likeCount -= 1;
            trailLike = undefined;
        } else {
            trailLike = await trail_like_create(
                new TrailLike($currentUser.actor, trail.id),
            );
            likeCount += 1;
        }
    }

    let trailLike = $state(
        trail.expand?.trail_like_via_trail?.find(
            (l) => l.actor === $currentUser?.actor,
        ),
    );
</script>

<button
    class="btn-icon bg-background"
    aria-label="like"
    type="button"
    onclick={like}
>
    {#if likeCount > 0}
        <div
            class="absolute pointer-events-none -top-[2px] left-6 text-xs rounded-full bg-content text-content-inverse px-1 text-center"
        >
            {likeCount}
        </div>
    {/if}
    <i
        class="{trailLike ? 'fa' : 'fa-regular'} fa-heart"
        class:text-3xl={large}
        class:text-red-500={trailLike !== undefined}
    ></i>
</button>
