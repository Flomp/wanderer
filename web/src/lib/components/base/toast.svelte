<script lang="ts">
    import { closeToast, toastStore } from "$lib/stores/toast_store.svelte";
    import { fade } from "svelte/transition";
</script>

{#each toastStore.toasts as t, i}
    <div
        class="fixed px-4 py-3 shadow-xl rounded-xl border right-4 bg-white text-gray-500 flex items-center max-w-md"
        style="z-index: 1001; bottom: {16 + i * 72}px"
        in:fade={{ duration: 250 }}
        out:fade={{ duration: 250 }}
    >
        <i
            class="fa fa-{t.icon} p-3 mr-2 rounded-lg"
            class:success-toast={t.type == "success"}
            class:error-toast={t.type == "error"}
            class:warning-toast={t.type == "warning"}
        ></i>
        <p class="mr-4">
            {t.text}
        </p>
        <button
            type="button"
            class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm w-8 h-8 ms-auto inline-flex justify-center items-center dark:hover:bg-gray-600 dark:hover:text-white ml-6"
            onclick={() => closeToast(t)}
        >
            <i class="fa fa-close"></i>
            <span class="sr-only">Close modal</span>
        </button>
    </div>
{/each}

<style>
    .success-toast {
        color: #10b981;
        background-color: #a7f3d0;
    }
    .error-toast {
        color: #ef4444;
        background-color: #fecaca;
    }
    .warning-toast {
        color: #ffb900;
        background-color: #fef3c6;
    }
</style>
