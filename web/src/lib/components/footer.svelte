<script>
    import LogoTextLight from "./logo/logo_text_light.svelte";
</script>

<footer class="bg-primary text-white w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-y-8 p-12 mt-12 rounded-t-3xl">
    <LogoTextLight></LogoTextLight>
    <div>
        <h5 class="font-semibold">Resources</h5>
        <ul></ul>
    </div>
    <div>
        <h5 class="font-semibold">wanderer</h5>
        <ul class="mt-4">
            <li>About</li>
            <li>Features</li>
            <li>Changelog</li>
            <li>License</li>
        </ul>
    </div>
    <div>
        <h5 class="font-semibold">Community</h5>
        <ul class="mt-4">
            <li>GitHub</li>
            <li>Issues</li>
            <li>Contribute</li>
        </ul>
    </div>
</footer>

<style>
    footer li:not(:last-child) {
        margin-bottom: 4px;
    }
</style>
