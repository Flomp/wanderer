<script>
    import Avatar from "./avatar.svelte";
    import LogoText from "./logo_text.svelte";
</script>

<nav class="flex justify-between items-center p-6">
    <a href="/"><LogoText></LogoText></a>
    <menu class="flex gap-8">
        <a class="font-semibold" href="">Trails</a>
        <a class="font-semibold" href="">Map</a>
        <a class="font-semibold" href="">Categories</a>
        <a class="font-semibold" href="">Favorites</a>
    </menu>
    <div class="flex gap-6 items-center">
        <button class="btn-primary"
            ><i class="fa fa-plus mr-2"></i>New Trail</button
        >
        <Avatar></Avatar>
    </div>
</nav>