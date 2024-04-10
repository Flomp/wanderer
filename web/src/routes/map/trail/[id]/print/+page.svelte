<script lang="ts">
    import { page } from "$app/stores";
    import Button from "$lib/components/base/button.svelte";
    import Select from "$lib/components/base/select.svelte";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import MapWithElevation from "$lib/components/trail/map_with_elevation.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import {
        calculatePixelPerMeter,
        calculateScaleFactor,
    } from "$lib/util/leaflet_util";
    import { createRect, createText } from "$lib/util/svg_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import leafletImage from "$lib/vendor/leaflet-image/leaflet-image.js";
    import type { Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import QrCodeWithLogo from "qrcode-with-logos";
    import { onMount, tick } from "svelte";
    import { _ } from "svelte-i18n";

    let map: Map;

    const paperSizes: { text: string; value: keyof typeof paperDimensions }[] =
        [
            { text: "A4", value: "a4" },
            { text: "Letter", value: "letter" },
        ];
    const paperDimensions = {
        a4: {
            width: "21cm",
            height: "29.7cm",
        },
        letter: {
            width: "21.59cm",
            height: "27.94cm",
        },
    };
    let selectedPaperSize = paperSizes[0].value;

    const orientations: { text: string; value: "portrait" | "landscape" }[] = [
        { text: "Portrait", value: "portrait" },
        { text: "Landscape", value: "landscape" },
    ];
    let selectedOrientation: "portrait" | "landscape" = orientations[0].value;

    let scale = 1;

    onMount(() => {
        let qrcode = new QrCodeWithLogo({
            content: $page.url.href.replace("/print", ""),
            image: document.getElementById("qrcode") as HTMLImageElement,
            logo: {
                src: "/favicon.png",
            },
        });
    });

    function print() {
        leafletImage(map, function (err: string, canvas: HTMLCanvasElement) {
            var img = document.createElement("img");
            var dimensions = map.getSize();
            img.width = dimensions.x;
            img.height = dimensions.y;
            img.src = canvas.toDataURL();

            document.getElementById("images")!.innerHTML = "";
            document.getElementById("images")!.appendChild(img);
        });
    }

    async function updateMapSize() {
        if (!map) {
            return;
        }
        await tick();
        map.invalidateSize();
        updateScale(map);
    }

    async function updateScale(map: Map) {
        const svg = document.getElementById("ruler") as HTMLElement;
        if (!svg) {
            return;
        }
        const svgRect = svg.getBoundingClientRect();
        const height = svgRect.height;
        const width = svgRect.width;

        svg.replaceChildren();

        let oneUnitInPixels = calculatePixelPerMeter(
            map,
            $currentUser && $currentUser.unit == "imperial" ? 1609.34 : 1000,
        );
        let unitsInRuler = width / oneUnitInPixels;

        function createRulerSegment(
            x: number,
            width: number,
            label: string,
            fill: string,
        ) {
            const bar = createRect(x, height / 2, width - 4, 6, fill);
            const tick = createRect(x + width - 4, height / 2, 4, 6, "#000");
            const text = createText(
                label,
                x + width - 5 * label.length,
                height / 2 + 20,
                "0.75rem",
            );
            svg.appendChild(bar);
            svg.appendChild(tick);
            svg.append(text);
        }
        let multiplier = 1;

        if (unitsInRuler > 500) {
            multiplier = 250;
        } else if (unitsInRuler > 150) {
            multiplier = 50;
        } else if (unitsInRuler > 25) {
            multiplier = 10;
        } else if (unitsInRuler > 15) {
            multiplier = 5;
        } else if (unitsInRuler > 5) {
            multiplier = 2;
        } else if (unitsInRuler > 1.5) {
            multiplier = 1;
        } else if (unitsInRuler > 1) {
            multiplier = 0.5;
        } else if (unitsInRuler > 0.2) {
            multiplier = 0.1;
        } else if (unitsInRuler > 0) {
            multiplier = 0.05;
        }
        const segmentCount = Math.floor(unitsInRuler / multiplier);

        svg.appendChild(createText("0", 0, height / 2 + 20, "0.75rem"));
        for (let i = 0; i < segmentCount; i++) {
            createRulerSegment(
                oneUnitInPixels * i * multiplier,
                oneUnitInPixels * multiplier,
                `${multiplier < 1 ? ((i + 1) * multiplier).toFixed(1) : (i + 1) * multiplier}`,
                i % 2 == 0 ? "#7c7c7c" : "#5d5d5d",
            );
        }

        svg.appendChild(
            createText(
                $currentUser && $currentUser.unit == "imperial" ? "MI" : "KM",
                segmentCount * multiplier * oneUnitInPixels - 20,
                height / 2 - 7,
            ),
        );

        scale = calculateScaleFactor(map);
    }
</script>

<svelte:head>
    <title>{$_("print")} | wanderer</title>
</svelte:head>

<main class="grid grid-cols-1 md:grid-cols-[400px_1fr] gap-x-1 gap-y-4 items-start">
    <div
        class="print-details px-6 space-y-4 pb-4 border border-input-border rounded-3xl"
    >
        <div id="images"></div>
        <h1 class="text-4xl font-bold">Print Trail</h1>
        <Select
            bind:value={selectedPaperSize}
            items={paperSizes}
            label="Paper size"
            on:change={updateMapSize}
        ></Select>
        <Select
            bind:value={selectedOrientation}
            items={orientations}
            label="Orientation"
            on:change={updateMapSize}
        ></Select>

        <Button primary={true} on:click={print}>Print!</Button>
    </div>

    <div class="paper-container overflow-scroll">
        <div
            class="paper flex flex-col mx-auto shadow-xl mb-8 p-8 bg-white"
            style="width: {paperDimensions[selectedPaperSize][
                selectedOrientation == 'portrait' ? 'width' : 'height'
            ]}; height: {paperDimensions[selectedPaperSize][
                selectedOrientation == 'portrait' ? 'height' : 'width'
            ]}"
        >
            <div class="basis-full">
                <MapWithElevation
                    trail={$trail}
                    options={{
                        theme: "gray-theme",
                        slope: false,
                        speed: false,
                        legend: false,
                        time: false,
                        hotline: false,
                        graticule: true,
                        height: 150,
                    }}
                    on:zoomend={(e) => updateScale(e.detail)}
                    bind:map
                >
                    <img
                        class="w-[100px] h-[100px]"
                        id="qrcode"
                        alt="QR Code"
                    />
                    <div class="basis-full mx-4">
                        <div
                            class="flex mt-1 gap-4 text-sm text-gray-500 whitespace-nowrap"
                        >
                            <span
                                ><i class="fa fa-left-right mr-2"
                                ></i>{formatDistance($trail.distance)}</span
                            >
                            <span
                                ><i class="fa fa-up-down mr-2"
                                ></i>{formatElevation(
                                    $trail.elevation_gain,
                                )}</span
                            >
                            <span
                                ><i class="fa fa-clock mr-2"
                                ></i>{formatTimeHHMM($trail.duration)}</span
                            >
                        </div>
                        <svg width="100%" height="50" id="ruler"> </svg>
                        <span class="text-sm"
                            >Scale: &nbsp 1 : {Math.round(scale)}</span
                        >
                    </div>
                </MapWithElevation>
            </div>
            <div class="pt-4 -mt-4 border-t border-t-input-border z-10">
                <div class="float-right"><LogoText height={48}></LogoText></div>

                <h4 class="text-sm font-semibold">{$trail.name}</h4>
                <h5 class="text-sm">
                    <i class="fa fa-location-dot mr-2"></i>{$trail.location ||
                        "-"}
                </h5>
            </div>
        </div>
    </div>
</main>

<style>
    :global(#map-container) {
        height: 100%;
    }
</style>
