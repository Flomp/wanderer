<script lang="ts">
    import { page } from "$app/stores";
    import Calendar from "$lib/components/base/calendar.svelte";
    import Datepicker from "$lib/components/base/datepicker.svelte";
    import MultiSelect from "$lib/components/base/multi_select.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import SummitLogTable from "$lib/components/summit_log/summit_log_table.svelte";
    import { categories } from "$lib/stores/category_store.js";
    import {
        summit_logs_index,
        summitLogs,
    } from "$lib/stores/summit_log_store";
    import { currentUser } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatSpeed,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import {
        ArcElement,
        BarElement,
        CategoryScale,
        Chart as ChartJS,
        Legend,
        LinearScale,
        Title,
        Tooltip,
    } from "chart.js";
    import { Bar, Pie } from "svelte-chartjs";
    import { _ } from "svelte-i18n";

    export let data;

    ChartJS.register(
        Title,
        Tooltip,
        Legend,
        ArcElement,
        CategoryScale,
        LinearScale,
        BarElement,
    );

    const filter = data.filter;

    const categorySelectItems: SelectItem[] = $categories.map((c) => ({
        value: c.id,
        text: c.name,
    }));

    const barChartSelectItems: SelectItem[] = [
        {
            text: $_("distance"),
            value: "distance",
        },
        {
            text: $_("duration"),
            value: "duration",
        },
        {
            text: $_("elevation-gain"),
            value: "elevation_gain",
        },
        {
            text: $_("elevation-loss"),
            value: "elevation_loss",
        },
    ];

    let barChartSelectedValue = barChartSelectItems[0].value;

    const categoryColors = [
        "#fb8500",
        "#ffb703",
        "#06d6a0",
        "#219ebc",
        "#8ecae6",
        "#ffafcc",
    ];

    const conversionFactors = {
        distance: 0.621371,
        elevation_gain: 3.28084,
        elevation_loss: 3.28084,
    };

    $: logCategories = $summitLogs.reduce(
        (acc, log) => {            
            const cat =
                log.expand?.trails_via_summit_logs?.at(0)?.expand?.category
                    ?.name ?? "unknown";
            acc[$_(cat)] = (acc[$_(cat)] || 0) + 1;
            return acc;
        },
        {} as Record<string, number>,
    );

    $: categoryLabels = Object.keys(logCategories).sort();
    $: categoryValues = Object.values(logCategories);
    $: categoryColorMap = Object.fromEntries(
        categoryLabels.map((label, index) => [
            label,
            categoryColors[index % categoryColors.length],
        ]),
    );

    $: categoryChartData = {
        labels: categoryLabels,
        datasets: [
            {
                data: categoryValues,
                backgroundColor: categoryColors,
            },
        ],
    };

    function barChartUnit() {
        const unit = $page.data.settings?.unit ?? "metric";

        switch (barChartSelectedValue) {
            case "duration":
                return "h";
            case "elevation_gain":
            case "elevation_loss":
                return unit == "metric" ? "m" : "ft";
            case "distance":
                return unit == "metric" ? "km" : "mi";
        }
    }

    $: barChartDataByDate = $summitLogs.reduce(
        (acc, log) => {
            const date = new Date(log.date).toLocaleDateString(undefined, {
                month: "2-digit",
                day: "2-digit",
                year: "numeric",
                timeZone: "UTC",
            });
            acc[date] =
                (acc[date] || 0) + ((log as any)[barChartSelectedValue] ?? 0);
            if (barChartSelectedValue === "distance") {
                acc[date] = acc[date] / 1000;
            } else if (barChartSelectedValue === "duration") {
                acc[date] = acc[date] / 60 / 60;
            }
            if ($page.data.settings?.unit !== "metric") {
                acc[date] =
                    acc[date] *
                    (conversionFactors as any)[barChartSelectedValue];
            }
            return acc;
        },
        {} as Record<string, number>,
    );

    $: barChartLabels = Object.keys(barChartDataByDate); // Dates as labels
    $: barChartValues = Object.values(barChartDataByDate);

    $: barChartData = {
        labels: barChartLabels,
        datasets: [
            {
                label: barChartSelectItems.find(
                    (i) => i.value === barChartSelectedValue,
                )?.text,
                data: barChartValues,
                backgroundColor: "#3388ff",
                borderRadius: 15,
            },
        ],
    };

    $: totalDistance = $summitLogs.reduce(
        (sum, log) => sum + (log.distance ?? 0),
        0,
    );

    $: totalDuration = $summitLogs.reduce(
        (sum, log) => sum + (log.duration ?? 0),
        0,
    );

    $: totalElevationGain = $summitLogs.reduce(
        (sum, log) => sum + (log.elevation_gain ?? 0),
        0,
    );

    $: totalElevationLoss = $summitLogs.reduce(
        (sum, log) => sum + (log.elevation_loss ?? 0),
        0,
    );

    $: averageSpeed =
        totalDuration > 0
            ? $summitLogs.reduce(
                  (sum, log) =>
                      sum + (log.distance && log.duration ? log.distance : 0),
                  0,
              ) / totalDuration
            : undefined;

    function updateFilterCategory(categories: SelectItem[]) {
        filter.category = categories.map((c) => c.value);
        loadSummitLogs();
    }

    function handleDateClick(date: Date) {
        const datePlusN = date;
        datePlusN.setDate(date.getDate() + 1);
        filter.startDate = datePlusN.toISOString().slice(0, 10);
        datePlusN.setDate(date.getDate() + 1);
        filter.endDate = datePlusN.toISOString().slice(0, 10);
        loadSummitLogs();
    }

    async function loadSummitLogs() {
        const logs = await summit_logs_index($page.params.id, filter);
        summitLogs.set(logs);
    }
</script>

<svelte:head>
    <title>{$_("profile")} | wanderer</title>
</svelte:head>

<div
    class="grid grid-cols-1 sm:grid-cols-[272px_minmax(0,_1fr)] md:grid-cols-[356px_minmax(0,_1fr)] gap-y-4 max-w-6xl mx-auto"
>
    {#if data.user}
        <div class="flex items-center gap-x-6 h-full px-4">
            <img
                class="rounded-full w-16 aspect-square overflow-hidden"
                src={getFileURL(data.user, data.user.avatar) ||
                    `https://api.dicebear.com/7.x/initials/svg?seed=${data.user.username}&backgroundType=gradientLinear`}
                alt="avatar"
            />
            <div>
                <h4 class="text-2xl font-semibold col-start-2">
                    {data.user.username}
                </h4>
                <p class="text-sm">
                    <span class="text-gray-500">Joined:</span>
                    {new Date(data.user.created ?? "").toLocaleDateString(
                        undefined,
                        {
                            month: "2-digit",
                            day: "2-digit",
                            year: "numeric",
                            timeZone: "UTC",
                        },
                    )}
                </p>
            </div>
        </div>
    {/if}
    <div
        class="flex flex-wrap md:flex-nowrap col-start-1 sm:col-start-2 gap-x-4 justify-end"
    >
        <MultiSelect
            on:change={(e) => updateFilterCategory(e.detail)}
            label={$_("categories")}
            items={categorySelectItems}
            placeholder={`${$_("filter-categories")}...`}
        ></MultiSelect>
        <Datepicker
            on:change={loadSummitLogs}
            bind:value={filter.startDate}
            label={$_("after")}
        ></Datepicker>
        <Datepicker
            on:change={loadSummitLogs}
            bind:value={filter.endDate}
            label={$_("before")}
        ></Datepicker>
    </div>
    <div class="space-y-4 grow-0 sm:mr-4">
        <div class="border border-input-border rounded-xl p-6">
            <Calendar
                on:click={(e) => handleDateClick(e.detail)}
                logs={$summitLogs}
                colorMap={categoryColorMap}
            ></Calendar>
        </div>
        <div
            class="border border-input-border rounded-xl p-6 space-y-4"
        >
            <span class="text-gray-500 font-semibold text-lg"
                ><i class="fa fa-person-hiking mr-3"></i>{$_(
                    "categories",
                )}</span
            >
            <Pie
                data={categoryChartData}
                options={{
                    responsive: true,
                    plugins: {
                        legend: {
                            position: "bottom",
                        },
                    },
                }}
            />
        </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-hashtag mr-3"></i>{$_("activity", {
                    values: { n: 2 },
                })}</span
            >
            <p class="text-3xl font-bold">{$summitLogs.length}</p>
        </div>
        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-left-right mr-3"></i>{$_("distance")}</span
            >
            <p class="text-3xl font-bold">{formatDistance(totalDistance)}</p>
        </div>

        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-clock mr-3"></i>{$_("duration")}</span
            >
            <p class="text-3xl font-bold">
                {formatTimeHHMM(totalDuration / 60)}
            </p>
        </div>
        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-arrow-trend-up mr-3"></i>{$_(
                    "elevation-gain",
                )}</span
            >
            <p class="text-3xl font-bold">
                {formatElevation(totalElevationGain)}
            </p>
        </div>
        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-arrow-trend-down mr-3"></i>{$_(
                    "elevation-loss",
                )}</span
            >
            <p class="text-3xl font-bold">
                {formatElevation(totalElevationLoss)}
            </p>
        </div>
        <div
            class="flex flex-col items-center gap-4 border border-input-border rounded-xl p-6"
        >
            <span class="text-gray-500 font-semibold text-lg self-start"
                ><i class="fa fa-arrow-trend-down mr-3"></i>{$_(
                    "average-speed",
                )}</span
            >
            <p class="text-3xl font-bold">{formatSpeed(averageSpeed)}</p>
        </div>
        <div
            class="h-full sm:col-span-2 lg:col-span-3 space-y-2 border border-input-border rounded-xl p-6"
        >
            <div class="flex justify-between">
                <span class="text-gray-500 font-semibold text-lg"
                    ><i class="fa fa-calendar mr-3"></i>{$_("activity", {
                        values: { n: 1 },
                    })}</span
                >
                <Select
                    bind:value={barChartSelectedValue}
                    items={barChartSelectItems}
                ></Select>
            </div>

            <Bar
                data={barChartData}
                options={{
                    plugins: {
                        legend: {
                            display: false,
                        },
                        tooltip: {
                            callbacks: {
                                label: (item) =>
                                    `${item.dataset.label}: ${item.formattedValue} ${barChartUnit()}`,
                            },
                        },
                    },
                    scales: {
                        y: {
                            ticks: {
                                callback: function (value, index, ticks) {
                                    return value + " " + barChartUnit();
                                },
                            },
                        },
                    },
                }}
            ></Bar>
        </div>
    </div>

    <div
        class="col-span-1 sm:col-span-2 border border-input-border rounded-xl p-6 space-y-6"
    >
        <span class="text-gray-500 font-semibold text-lg"
            ><i class="fa fa-table mr-3"></i>{$_("all-activities")}</span
        >
        <div class=" overflow-x-auto">
            <SummitLogTable summitLogs={$summitLogs} showCategory showTrail
            ></SummitLogTable>
        </div>
    </div>
</div>
