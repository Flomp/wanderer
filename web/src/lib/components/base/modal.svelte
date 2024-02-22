<script lang="ts">
    export let id: string;
    export let title: string;
    export function openModal() {
        (document.getElementById(id) as HTMLDialogElement).showModal();
    }

    export function closeModal() {
        (document.getElementById(id) as HTMLDialogElement).close();
    }
</script>

<slot {openModal} />
<dialog
    {id}
    tabindex="-1"
    aria-hidden="true"
    class="relative w-full max-w-2xl max-h-full rounded-xl"
>
    <!-- Modal content -->
    <div class="relative bg-white shadow dark:bg-gray-700">
        <!-- Modal header -->
        <div
            class="flex items-center justify-between p-4 md:p-5 border-b rounded-t dark:border-gray-600"
        >
            <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
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
        <div
            class="p-4 md:p-5 border-t border-gray-200 rounded-b dark:border-gray-600"
        >
            <slot name="footer" {closeModal} />
        </div>
    </div>
</dialog>

<style>
</style>
