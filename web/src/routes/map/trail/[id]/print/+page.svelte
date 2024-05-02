<script lang="ts">
    import { page } from "$app/stores";
    import "$lib/assets/fonts/IBMPlexSans-Regular-normal";
    import "$lib/assets/fonts/IBMPlexSans-SemiBold-bold";
    import "$lib/assets/fonts/fa-solid-900-normal";
    import Button from "$lib/components/base/button.svelte";
    import Select from "$lib/components/base/select.svelte";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import MapWithElevation from "$lib/components/trail/map_with_elevation.svelte";
    import type { Settings } from "$lib/models/settings";
    import { show_toast } from "$lib/stores/toast_store";
    import { trail } from "$lib/stores/trail_store";
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
    import type AutoGraticule from "$lib/vendor/leaflet-graticule/leaflet-auto-graticule";
    import leafletImage from "$lib/vendor/leaflet-image/leaflet-image.js";
    import { Canvg } from "canvg";
    import { jsPDF } from "jspdf";
    import type { Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import QrCodeWithLogo from "$lib/vendor/qr-code-with-logos/index";
    import { onMount, tick } from "svelte";
    import { _ } from "svelte-i18n";

    let map: Map;
    let graticule: AutoGraticule;

    const settings: Settings = $page.data.settings;

    const paperSizes: { text: string; value: keyof typeof paperDimensions }[] =
        [
            { text: "A4", value: "a4" },
            { text: "Letter", value: "letter" },
        ];
    const paperDimensions = {
        a4: {
            css: {
                width: "21cm",
                height: "29.7cm",
            },
            pdf: {
                width: 210,
                height: 297,
            },
        },
        letter: {
            css: {
                width: "21.59cm",
                height: "27.94cm",
            },
            pdf: {
                width: 215.9,
                height: 279.4,
            },
        },
    };
    let selectedPaperSize = paperSizes[0].value;

    const orientations: { text: string; value: "portrait" | "landscape" }[] = [
        { text: "Portrait", value: "portrait" },
        { text: "Landscape", value: "landscape" },
    ];
    let selectedOrientation: "portrait" | "landscape" = orientations[0].value;

    const gridOptions = [
        { text: "Degrees", value: "degree" },
        { text: "No Grid", value: "off" },
    ];
    let selectedGrid = gridOptions[0].value;

    let scale: number = 1;

    let printLoading: boolean = false;

    let includeDescription: boolean = false;

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
        printLoading = true;
        const doc = new jsPDF(selectedOrientation, "mm", [
            paperDimensions[selectedPaperSize].pdf[
                selectedOrientation == "portrait" ? "width" : "height"
            ],
            paperDimensions[selectedPaperSize].pdf[
                selectedOrientation == "portrait" ? "height" : "width"
            ],
        ]);
        doc.setFont("IBMPlexSans-Regular", "normal");
        var width = doc.internal.pageSize.getWidth();
        var height = doc.internal.pageSize.getHeight();
        let currentHeight = 0;
        const rulerElement = document.getElementById("ruler") as HTMLElement;
        const rulerSegments = rulerElement.getElementsByTagName("rect");
        const rulerLabels = rulerElement.getElementsByTagName("text");

        const pageWidth = document
            .getElementById("paper")
            ?.getBoundingClientRect().width;

        try {
            leafletImage(
                map,
                async function (err: string, canvas: HTMLCanvasElement) {
                    // Map
                    var img = document.createElement("img");
                    var dimensions = map.getSize();
                    img.width = dimensions.x;
                    img.height = dimensions.y;
                    const ratio = img.height / img.width;
                    img.src = canvas.toDataURL();
                    doc.addImage(img.src, "png", 8, 8, 0, (width - 16) * ratio);
                    currentHeight += (width - 16) * ratio + 11;
                    const elevationProfileHeight = currentHeight - 4;

                    // QR Code
                    doc.addImage(
                        document.getElementById("qrcode") as HTMLImageElement,
                        "png",
                        8,
                        currentHeight,
                        24,
                        24,
                    );

                    // Distance
                    currentHeight += 5;
                    let currentWidth = 8 + 24 + 4;
                    doc.setFontSize(9);
                    doc.setTextColor(107, 114, 128);
                    doc.setFont("fa-solid-900", "normal");
                    doc.text("\uf337", currentWidth, currentHeight);
                    currentWidth += doc.getTextWidth("\uf337") + 3;
                    doc.setFont("IBMPlexSans-Regular", "normal");
                    doc.text(
                        formatDistance($trail.distance),
                        currentWidth,
                        currentHeight,
                    );
                    currentWidth +=
                        doc.getTextWidth(formatDistance($trail.distance)) + 4;

                    // Elevation gain
                    doc.setFont("fa-solid-900", "normal");
                    doc.text("\uf338", currentWidth, currentHeight);
                    currentWidth += doc.getTextWidth("\uf338") + 3;
                    doc.setFont("IBMPlexSans-Regular", "normal");
                    doc.text(
                        formatElevation($trail.elevation_gain),
                        currentWidth,
                        currentHeight,
                    );
                    currentWidth +=
                        doc.getTextWidth(
                            formatDistance($trail.elevation_gain),
                        ) + 4;

                    // Duration
                    doc.setFont("fa-solid-900", "normal");
                    doc.text("\uf017", currentWidth, currentHeight);
                    currentWidth += doc.getTextWidth("\uf337") + 3;
                    doc.setFont("IBMPlexSans-Regular", "normal");
                    doc.text(
                        formatTimeHHMM($trail.duration),
                        currentWidth,
                        currentHeight,
                    );

                    // Ruler
                    currentHeight += 6;
                    currentWidth = 8 + 24 + 4;
                    doc.setTextColor(0, 0, 0);
                    for (let i = 0; i < rulerSegments.length; i++) {
                        const segment = rulerSegments[i];
                        const label = rulerLabels[i].textContent ?? "";
                        const segmentWidth =
                            segment.getBoundingClientRect().width;
                        const segmentWidthPercent =
                            segmentWidth / (pageWidth ?? 0);
                        const segmentPDFWidth = width * segmentWidthPercent;
                        let textWidth = doc.getTextWidth(label);

                        if (i % 2 == 0) {
                            doc.setFillColor(255, 255, 255);
                        } else {
                            doc.setFillColor(93, 93, 93);
                        }

                        doc.rect(
                            currentWidth,
                            currentHeight,
                            segmentPDFWidth,
                            2,
                            "FD",
                        );
                        doc.text(
                            label,
                            currentWidth - textWidth / 2,
                            currentHeight + 6,
                        );
                        currentWidth += segmentPDFWidth;
                        if (i == rulerSegments.length - 1) {
                            const endLabel =
                                rulerLabels[i + 1].textContent ?? "";
                            textWidth = doc.getTextWidth(endLabel);
                            doc.text(
                                endLabel,
                                currentWidth - textWidth / 2,
                                currentHeight + 6,
                            );
                            textWidth = doc.getTextWidth("KM");
                            doc.text(
                                "KM",
                                currentWidth - textWidth,
                                currentHeight - 2,
                            );
                        }
                    }

                    // Scale
                    currentHeight += 12;
                    doc.setFillColor(255, 255, 255);
                    doc.text(
                        `Scale:  1 : ${Math.round(scale)}`,
                        8 + 24 + 4,
                        currentHeight,
                    );

                    // Elevation profile
                    const gridLines = document.querySelectorAll(
                        "#elevation .grid line",
                    );
                    gridLines.forEach((l) =>
                        l.setAttribute("stroke", "#d1d5db"),
                    );
                    const gridBorders = document.querySelectorAll(
                        "#elevation .grid path",
                    );
                    gridBorders.forEach((l) =>
                        l.setAttribute("stroke", "#fff"),
                    );
                    const tooltip = document.querySelector(
                        "#elevation .tooltip",
                    ) as HTMLElement;
                    tooltip.style.display = "none";
                    const units = document.querySelectorAll(
                        "#elevation .axis text",
                    );
                    units.forEach((l) => l.setAttribute("fill", "#000"));
                    const elevationCanvas = document.createElement("canvas");
                    const ctx = elevationCanvas.getContext("2d")!;
                    const v = await Canvg.from(
                        ctx,
                        document.querySelector("#elevation svg")!.outerHTML,
                    );
                    v.start();
                    const profileRatio =
                        elevationCanvas.height / elevationCanvas.width;
                    doc.addImage(
                        elevationCanvas.toDataURL("image/png"),
                        width - 79.3 - 8,
                        elevationProfileHeight,
                        79.3,
                        79.3 * profileRatio,
                    );
                    const elevationPlot = (
                        document.querySelector(
                            ".canvas-plot",
                        ) as HTMLCanvasElement
                    ).toDataURL("image/png");
                    doc.addImage(
                        elevationPlot,
                        width - 77.3,
                        elevationProfileHeight + 8,
                        0,
                        0,
                    );
                    tooltip.style.display = "block";

                    // Separator
                    currentHeight += 5;
                    doc.setDrawColor(232, 234, 237);
                    doc.line(8, currentHeight, width - 8, currentHeight);
                    currentHeight += 6;

                    // Trail name & location
                    doc.setTextColor(0, 0, 0);
                    doc.setFont("IBMPlexSans-SemiBold", "bold");
                    doc.text($trail.name, 8, currentHeight);
                    doc.setFont("fa-solid-900", "normal");
                    currentHeight += 5;
                    doc.text("\uf3c5", 8, currentHeight);
                    doc.setFont("IBMPlexSans-Regular", "normal");
                    doc.text($trail.location || "-", 12, currentHeight);

                    // Logo
                    const logo = new Image();
                    logo.src = "/imgs/logo_text_dark.png";
                    const logoRatio = 64 / 212;
                    const logoWidth = 32;
                    const logoHeight = logoWidth * logoRatio;
                    doc.addImage(
                        logo,
                        width - 8 - logoWidth,
                        height - 4 - logoHeight,
                        logoWidth,
                        logoHeight,
                    );

                    if (includeDescription && $trail.description) {
                        doc.addPage();
                        doc.text($trail.description, 16, 16, {
                            maxWidth: width - 32,
                        });
                        doc.addImage(
                            logo,
                            width - 8 - logoWidth,
                            height - 4 - logoHeight,
                            logoWidth,
                            logoHeight,
                        );
                    }

                    doc.save($trail.name + ".pdf");
                    printLoading = false;
                },
            );
        } catch (e) {
            show_toast({
                icon: "close",
                type: "error",
                text: $_("error-printing-map"),
            });
            printLoading = false;
        }
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
            settings && settings.unit == "imperial" ? 1609.34 : 1000,
        );
        let unitsInRuler = width / oneUnitInPixels;

        function createRulerSegment(
            x: number,
            width: number,
            label: string,
            fill: string,
        ) {
            const bar = createRect(x, height / 2, width, 6, fill, "#000");
            const text = createText(
                label,
                x + width - 5 * label.length,
                height / 2 + 20,
                "0.75rem",
            );
            svg.appendChild(bar);
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
                i % 2 == 0 ? "#fff" : "#5d5d5d",
            );
        }

        svg.appendChild(
            createText(
                settings && settings.unit == "imperial" ? "MI" : "KM",
                segmentCount * multiplier * oneUnitInPixels - 20,
                height / 2 - 7,
            ),
        );

        scale = calculateScaleFactor(map);
    }

    function toggleGrid() {
        if (selectedGrid == "off") {
            graticule?.remove();
        } else {
            graticule.addTo(map);
        }
    }
</script>

<svelte:head>
    <title>{$_("print")} | wanderer</title>
</svelte:head>

<main
    class="grid grid-cols-1 md:grid-cols-[400px_1fr] gap-x-1 gap-y-4 items-start"
>
    <div
        class="print-details md:sticky w-80 md:top-8 mx-auto md:ml-8 p-6 flex flex-col gap-y-4 border border-input-border rounded-3xl"
    >
        <h1 class="text-4xl font-bold">Print Trail</h1>
        <Select
            bind:value={selectedPaperSize}
            items={paperSizes}
            label={$_("paper-size")}
            on:change={updateMapSize}
        ></Select>
        <Select
            bind:value={selectedOrientation}
            items={orientations}
            label={$_("orientation")}
            on:change={updateMapSize}
        ></Select>
        <Select
            bind:value={selectedGrid}
            items={gridOptions}
            label={$_("grid")}
            on:change={toggleGrid}
        ></Select>
        <div>
            <input
                id="description-checkbox"
                type="checkbox"
                bind:value={includeDescription}
                class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
            />
            <label for="description-checkbox" class="ms-2 text-sm"
                >{$_("include-description")}</label
            >
        </div>

        <Button
            loading={printLoading}
            extraClasses="mt-2"
            primary={true}
            on:click={print}>{$_("print")}!</Button
        >
    </div>

    <div class="paper-container overflow-scroll">
        <div
            id="paper"
            class="flex flex-col mx-auto shadow-xl mb-8 p-8 bg-white"
            style="width: {paperDimensions[selectedPaperSize].css[
                selectedOrientation == 'portrait' ? 'width' : 'height'
            ]}; height: {paperDimensions[selectedPaperSize].css[
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
                    bind:graticule
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
                        <svg
                            width="100%"
                            height="50"
                            id="ruler"
                            xmlns="http://www.w3.org/2000/svg"
                            version="1.1"
                        >
                        </svg>
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
