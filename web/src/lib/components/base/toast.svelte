<script lang="ts">
    import { toast } from "$lib/stores/toast_store";

    $: getBackgroundColor = () => {
        if ($toast && $toast.type == "success") {
            return "green-200";
        }
        return "";
    };

    $: getBorderColor = () => {
        if ($toast && $toast.type == "success") {
            return "green-500";
        }
        return "";
    };
</script>

{#if $toast}
    <div
        class="fixed px-8 py-4 shadow-xl rounded-xl border bottom-4 right-4 bg-{getBackgroundColor()} border-{getBorderColor()}"
        style="z-index: 1000;"
    >
        <i class="fa fa-{$toast.icon} mr-2"></i>
        <span>{$toast.text}</span>
    </div>
{/if}
