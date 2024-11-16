<script lang="ts">
	import type { SummitLog } from "$lib/models/summit_log";
	import { createEventDispatcher } from "svelte";
	import { isSameDay, isToday } from "../../util/date_util";
	import { _ } from "svelte-i18n";
	export let logs: SummitLog[] = [];
	export let colorMap: Record<string, string> = {};

	const dispatch = createEventDispatcher<{
		forward: {
			start: Date;
			end: Date;
		};
		backward: {
			start: Date;
			end: Date;
		};
		click: Date;
	}>();

	const weekdays = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];
	const months = [
		"Januar",
		"Februar",
		"März",
		"April",
		"Mai",
		"Juni",
		"Juli",
		"August",
		"September",
		"Oktober",
		"November",
		"Dezember",
	];
	const today = new Date();
	let currentMonth = today.getMonth();
	let currentYear = today.getFullYear();
	let currentMonthArray: ({
		date: Date | undefined;
		today: boolean;
		log?: SummitLog;
	} | null)[];
	$: currentMonthArray = generateMonthArray(currentYear, currentMonth, logs);

	function calculateFirstDayOfMonthDayOfWeek(year: number, month: number) {
		const date = new Date(year, month, 1);
		const day = date.getDay();

		return day == 0 ? 6 : day - 1;
	}

	function daysInMonth(year: number, month: number) {
		const days = new Date(year, month + 1, 0).getDate();
		return days;
	}

	function generateMonthArray(
		year: number,
		month: number,
		logs: SummitLog[],
	) {
		const a: ({ date: Date; today: boolean; log?: SummitLog } | null)[] =
			[];
		const firstDay = calculateFirstDayOfMonthDayOfWeek(year, month);
		const totalDays = daysInMonth(year, month);

		for (let i = 0; i < 42; i++) {
			if (i < firstDay || i - firstDay >= totalDays) {
				a.push(null);
			} else {
				const date = new Date(
					currentYear,
					currentMonth,
					i + 1 - firstDay,
				);
				const today = isToday(date);

				const logAtDate = logs.find((l) =>
					isSameDay(date, new Date(l.date)),
				);

				a.push({ date: date, today: today, log: logAtDate });
			}
		}

		return a;
	}

	function monthPlus() {
		if (currentMonth == 11) {
			currentYear++;
			currentMonth = 0;
		} else {
			currentMonth++;
		}
		dispatch("forward", {
			start: new Date(currentYear, currentMonth, 1),
			end: new Date(currentYear, currentMonth + 1, 0),
		});
	}

	function monthMinus() {
		if (currentMonth == 0) {
			currentYear--;
			currentMonth = 11;
		} else {
			currentMonth--;
		}
		dispatch("backward", {
			start: new Date(currentYear, currentMonth, 1),
			end: new Date(currentYear, currentMonth + 1, 0),
		});
	}

	function colorKey(a: typeof currentMonthArray, i: number) {
		return $_(
			a[i]?.log?.expand?.trails_via_summit_logs?.at(0)?.expand?.category
				?.name ?? "",
		);
	}

	function handleDateClick(date?: Date) {
		if (!date) {
			return;
		}
		dispatch("click", date);
	}
</script>

<div class="calendar-header w-full flex items-center justify-between mb-6">
	<div class="calendar-month-year basis-full">
		<span class="text-lg">{months[currentMonth]}</span>
		<span>{currentYear}</span>
	</div>
	<button class="btn-icon mr-2" on:click={monthMinus}
		><i class="fa fa-caret-left"></i></button
	>
	<button class="btn-icon" on:click={monthPlus}
		><i class="fa fa-caret-right"></i></button
	>
</div>
<div class="calendar-body">
	<div class="grid grid-cols-7">
		{#each weekdays as weekday, i}
			<div
				class="calendar-weekday flex items-center justify-center h-10 text-gray-500"
			>
				{weekday}
			</div>
		{/each}
	</div>
	<div
		class="grid grid-cols-7 grid-rows-6 gap-1"
		style="aspect-ratio: 1.17/1"
	>
		{#each { length: 42 } as _, i}
			<button
				class="calendar-day flex items-center justify-center rounded-xl"
				on:click={() => handleDateClick(currentMonthArray[i]?.date)}
				class:today={currentMonthArray[i]?.today}
				style="background-color: {colorMap[
					colorKey(currentMonthArray, i)
				] ?? ''}"
			>
				{currentMonthArray[i]?.date?.getDate() ?? ""}
			</button>
		{/each}
	</div>
</div>

<style>
	.calendar-weekday {
		font-weight: 600;
	}
	.calendar-month-year span {
		font-weight: 600;
	}
	.calendar-day.today {
		@apply border border-input-border;
	}
</style>
