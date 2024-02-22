<script lang="ts">
    import { theme } from "$lib/stores/theme_store";
    import { T, useTask } from "@threlte/core";
    import { tweened } from "svelte/motion";
    import Earth from "./earth.svelte";

    let rotation = 0;
    useTask((delta) => {
        rotation += delta / 4;
    });

    const ambientIntensitiy = tweened(1.2, {
        duration: 500,
    });

    $: ambientColor = $theme == "light" ? "#ffffff" : "#4c7fe6";
    $: ambientIntensitiy.set($theme == "light" ? 1.2 : 0);
</script>

<T.PerspectiveCamera
    makeDefault
    position={[5, 2, 0]}
    on:create={({ ref }) => {
        ref.lookAt(0, 0, -0.5);
    }}
/>

<T.AmbientLight color={ambientColor} intensity={$ambientIntensitiy} />

<Earth {rotation}></Earth>
