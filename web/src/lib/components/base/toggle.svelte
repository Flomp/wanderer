<script lang="ts">
    export let name: string = "";
    export let value: boolean = false;
    export let label: string = "";
    export let error: string = "";
</script>

<label class="relative my-2 inline-flex items-center cursor-pointer">
    <input {name} bind:checked={value} type="checkbox" class="sr-only peer" value="1"/>
    <div
        class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-gray-400 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary"
    ></div>
    <span class="ms-3 text-sm font-medium">{label}</span>
</label>

<span class="toggle-error text-xs text-red-400">
    {error}
</span>
