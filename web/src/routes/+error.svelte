<script>
    import { page } from "$app/stores";
</script>

<div
    class="flex flex-col gap-y-4 items-center justify-center relative"
    style="height: calc(100dvh - 120px); font-size: 4rem"
>
    <img
        class="absolute top-1/2 left-1/2 aspect-square rounded-xl object-cover"
        width="500"
        src="/imgs/error.webp"
        alt="Default error img"
    />
    <div
        class="z-10 rounded-full p-8 bg-background flex flex-col gap-y-4 items-center"
        style="max-width: 400px"
    >
        <h1 class="font-bold" style="line-height: 4rem;">
            {$page.status}
        </h1>
        <p class="text-lg text-center">
            {$page.error?.message}
        </p>
    </div>
</div>
