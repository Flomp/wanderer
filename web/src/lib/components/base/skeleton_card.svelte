<script lang="ts">
    export let fullWidth: boolean = false;
</script>


<div
    class="skeleton-card relative rounded-2xl animate-pulse flex flex-col {fullWidth ? "w-full" : "sm:w-72 w-80"} p-4 bg-menu-background border border-input-border"
    aria-hidden="true"
>
    <!-- Image placeholder -->
    <div class="h-48 bg-menu-item-background-focus rounded-md"></div>

    <!-- Text placeholders -->
    <div class="mt-4 space-y-3">
        <div class="h-4 bg-menu-item-background-focus rounded w-3/4"></div>
        <div class="h-4 bg-menu-item-background-focus rounded w-full"></div>
        <div class="h-4 bg-menu-item-background-focus rounded w-1/2"></div>
    </div>
</div>
