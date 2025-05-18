<div class="animate-pulse space-y-4 p-4">
    <!-- Table Header Skeleton -->
    <div class="flex items-center justify-between gap-4 px-2">
        <div class="h-5 w-20 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-16 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-24 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-24 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-16 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-12 bg-menu-item-background-focus rounded"></div>
        <div class="h-5 w-8 bg-menu-item-background-focus rounded"></div>
    </div>

    <!-- Repeatable Skeleton Rows -->
    <div class="space-y-2">
        {#each { length: 5 } as _, index}
            <!-- Row 1 -->
            <div class="flex items-center justify-between gap-4 px-2">
                <!-- Image placeholder -->
                <div
                    class="h-10 w-10 bg-menu-item-background-focus rounded-lg"
                ></div>
                <!-- Date -->
                <div
                    class="h-4 w-20 bg-menu-item-background-focus rounded"
                ></div>
                <!-- Distance -->
                <div
                    class="h-4 w-16 bg-menu-item-background-focus rounded"
                ></div>
                <!-- Elevation gain -->
                <div
                    class="h-4 w-24 bg-menu-item-background-focus rounded"
                ></div>
                <!-- Elevation loss -->
                <div
                    class="h-4 w-24 bg-menu-item-background-focus rounded"
                ></div>
                <!-- Duration -->
                <div
                    class="h-4 w-12 bg-menu-item-background-focus rounded"
                ></div>
                <!-- Author -->
                <div
                    class="h-6 w-6 bg-menu-item-background-focus rounded-full"
                ></div>
                <!-- Map icon -->
                <div
                    class="h-5 w-5 bg-menu-item-background-focus rounded-md"
                ></div>
            </div>
        {/each}
    </div>
</div>
