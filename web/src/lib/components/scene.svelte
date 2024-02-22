<script lang="ts">
    import { theme } from "$lib/stores/theme_store";
    import { T, useTask } from "@threlte/core";
    import { tweened } from "svelte/motion";
    import Earth from "./earth.svelte";
    import { onMount } from "svelte";
    import * as THREE from "three";

    let rotation = 0;
    useTask((delta) => {
        rotation += delta / 4;
    });

    const ambientIntensitiy = tweened($theme == "light" ? 1.2 : 0, {
        duration: 500,
    });

    $: ambientColor = $theme == "light" ? "#ffffff" : "#4c7fe6";
    $: ambientIntensitiy.set($theme == "light" ? 1.2 : 0);

    let starsGeometry!: THREE.BufferGeometry;
    let starsMaterial!: THREE.PointsMaterial;
    onMount(() => {
        const stars = new Array(0);
        for (var i = 0; i < 500; i++) {
            let x = THREE.MathUtils.randFloatSpread(500);
            let y = THREE.MathUtils.randFloatSpread(500);
            let z = THREE.MathUtils.randFloatSpread(500);
            stars.push(x, y, z);
        }
        starsGeometry = new THREE.BufferGeometry();
        starsGeometry.setAttribute(
            "position",
            new THREE.Float32BufferAttribute(stars, 3),
        );
        starsMaterial = new THREE.PointsMaterial({ color: 0xffffff });
    });
</script>

<T.PerspectiveCamera
    makeDefault
    position={[5, 2, 0]}
    on:create={({ ref }) => {
        ref.lookAt(0, 0, -0.5);
    }}
/>

{#if starsGeometry && starsMaterial}
    <T.Points geometry={starsGeometry} material={starsMaterial}></T.Points>
{/if}

<T.AmbientLight color={ambientColor} intensity={$ambientIntensitiy} />

<Earth {rotation}></Earth>
