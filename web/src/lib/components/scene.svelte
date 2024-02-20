<script lang="ts">
    import { T, useTask } from "@threlte/core";
    import Earth from "./earth.svelte";

    let rotation = 0;
    useTask((delta) => {
        rotation += delta/4;
    });
</script>

<T.PerspectiveCamera
    makeDefault
    position={[5, 2, 0]}
    on:create={({ ref }) => {
        ref.lookAt(0, 0, -0.5);
    }}
/>

<T.AmbientLight color="#ffffff" intensity={1.1} />

<Earth rotation={rotation}></Earth>
