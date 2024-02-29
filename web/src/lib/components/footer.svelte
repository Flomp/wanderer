<script>
    import LogoTextLight from "./logo/logo_text_light.svelte";
    import { _ } from "svelte-i18n";

</script>

<footer class="bg-footer-background text-white w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-y-8 p-12 mt-12 rounded-t-3xl">
    <LogoTextLight></LogoTextLight>
    <div>
        <h5 class="font-semibold">Resources</h5>
        <ul></ul>
    </div>
    <div>
        <h5 class="font-semibold">wanderer</h5>
        <ul class="mt-4 text-sm">
            <li>{$_('about')}</li>
            <li>{$_('features')}</li>
            <li>{$_('changelog')}</li>
            <li><a href="https://github.com/Flomp/wanderer/blob/main/LICENSE">{$_('license')}</a></li>
        </ul>
    </div>
    <div>
        <h5 class="font-semibold">Community</h5>
        <ul class="mt-4 text-sm">
            <li><a href="https://github.com/Flomp/wanderer">GitHub</a></li>
            <li><a href="https://github.com/Flomp/wanderer/issues">Issues</a></li>
            <li><a href="https://github.com/Flomp/wanderer/pulls">{$_('contribute')}</a></li>
        </ul>
    </div>
</footer>

<style>
    footer li:not(:last-child) {
        margin-bottom: 4px;
    }
</style>
