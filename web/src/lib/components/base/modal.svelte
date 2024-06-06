<script lang="ts">
    export let id: string;
    export let title: string;
    export let size: string = "max-w-2xl";
    export function openModal() {
        document.body.style.position = "fixed";
        document.body.style.width = "100%";
        document.body.style.top = `-${window.scrollY}px`;

        (document.getElementById(id) as HTMLDialogElement).showModal();
    }

    export function closeModal() {
        const scrollY = document.body.style.top;
        document.body.style.position = "";
        document.body.style.top = "";
        window.scrollTo(0, parseInt(scrollY || "0") * -1);

        (document.getElementById(id) as HTMLDialogElement).close();
    }
</script>

<slot {openModal} />
<dialog
    {id}
    tabindex="-1"
    aria-hidden="true"
    class="relative w-full {size} max-h-full rounded-xl text-content"
>
    <!-- Modal content -->
    <div class="relative bg-background shadow rounded-xl">
        <!-- Modal header -->
        <div
            class="flex items-center justify-between p-4 md:p-5 border-b border-separator rounded-t"
        >
            <h3 class="text-xl font-semibold">
                {title}
            </h3>
            <button
                type="button"
                class="rounded-full btn-icon"
                on:click={closeModal}
            >
                <i class="fa fa-close"></i>
                <span class="sr-only">Close modal</span>
            </button>
        </div>
        <!-- Modal body -->
        <div class="p-4 md:p-5 space-y-4">
            <slot name="content" />
        </div>
        <!-- Modal footer -->
        <div class="p-4 md:p-5 border-t border-separator rounded-b">
            <slot name="footer" {closeModal} />
        </div>
    </div>
</dialog>

<style>
    dialog::backdrop {
        @apply bg-gray-500/50;
    }
</style>
