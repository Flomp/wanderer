<script lang="ts">
    import Toggle from "$lib/components/base/toggle.svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        onclick: () => void;
        ontoggle: (value: boolean) => void;
        active: boolean;
        disabled: boolean;
        img: string;
        title: string;
        description: string;
    }

    let {
        onclick,
        ontoggle,
        active = $bindable(),
        disabled,
        img,
        title,
        description,
    }: Props = $props();
</script>
<div class="border border-input-border rounded-lg p-4 space-y-4 flex flex-col justify-between">
    <img class="h-20" src={img} alt="integration logo"/>
    <div>
        <h5 class="text-xl font-semibold">{title}</h5>
        <p class="text-sm text-gray-500">
            {description}
        </p>
    </div>
    <div class="flex items-center justify-between">
        <button class="btn-secondary" {onclick}
            ><i class="fa fa-cogs mr-2"></i>{$_("settings")}</button
        >
        <Toggle bind:value={active} onchange={ontoggle} {disabled}></Toggle>
    </div>
</div>
