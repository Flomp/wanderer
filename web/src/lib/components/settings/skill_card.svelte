<script lang="ts">
    import Select from "$lib/components/base/select.svelte";
    import type { SelectItem } from "$lib/components/base/select.svelte";
    import { _ } from "svelte-i18n";
    import { Settings } from "$lib/models/settings";
    import type { Category } from "$lib/models/category.js";
  import { onMount } from "svelte";

    interface Props {
        settings: Settings,
        category: Category,
        onchange?: (value: any) => void
    }

    let {
        settings = $bindable(),
        category,
        onchange,
    }: Props = $props();
   
    let currentAlgorithm = $state("none");
    let currentSpeed = $state("normal");

    onMount(() => {
        let currentSkill = settings.skills?.find((s) => s.category == category.name);
        if (currentSkill) {
            currentAlgorithm = currentSkill.algorithm;
            currentSpeed = currentSkill.speed;
            console.log(currentAlgorithm)
        }
    })

    const difficultyModeItems: SelectItem[] = [ 
        { text: "", value: "none" }, 
        { text: $_("hiking"), value: "pedestrian" }, 
        { text: $_("cycling"), value: "bicycle" }, 
        { text: $_("driving"), value: "auto" } 
    ];

    const speedItems: SelectItem[] = [
        { text: $_("relaxed"), value: "0" },
        { text: $_("moderate"), value: "1" },
        { text: $_("normal"), value: "2" },
        { text: $_("fast"), value: "3" },
        { text: $_("super-fast"), value: "4" },
    ];

    async function handleDifficultyModeSelection(value: string) {
        if (!settings.skills) {
            settings.skills = new Array();
            settings.skills.push({ category: category.id, algorithm: value, speed: "" })
        } else if (settings.skills?.find((skill) => skill.category == category.id)) {
            settings.skills.find((skill) => skill.category == category.id)!.algorithm = value;
        } else {
            settings.skills.push({ category: category.id, algorithm: value, speed: "" })
        }


        onChange(category)
    }

    async function handleSpeedItemSelection(value: string) {
        if (!settings.skills) {
            settings.skills = new Array();
            settings.skills.push({ category: category.id, algorithm: "", speed: value })
        } else if (settings.skills?.find((skill) => skill.category == category.id)) {
            settings.skills.find((skill) => skill.category == category.id)!.speed = value;
        } else {
            settings.skills.push({ category: category.id, algorithm: "", speed: value })
        }

        onChange(category)
    }

    function onChange(target: any) {
        onchange?.(target?.value);
    }

</script>
<div class="flex items-top justify-between px-4 py-2 border border-input-border rounded-xl mb-2">
    <h3 class="text-x1 font-medium mb-2">{$_(category.name)}</h3>
    <div class="flex items-top justify-between py-2 mb-2">
        <div class="px-4">
            <h4 class="text-xs font-small mb-2">{$_("difficulty-algorithm")}</h4>
            <Select
                items={difficultyModeItems}
                bind:value={currentAlgorithm}
                onchange={handleDifficultyModeSelection}
            ></Select>
        </div>
        <div>
            <h4 class="text-xs font-small mb-2">{$_("speed")}</h4>
            <Select
                items={speedItems}
                bind:value={currentSpeed}
                onchange={handleSpeedItemSelection}
            ></Select>
        </div>
    </div>
</div>
