import { render } from "svelte/server";
import { addMessages, init } from "svelte-i18n";
import { beforeAll, describe, expect, it } from "vitest";
import type { StatisticActivity } from "$lib/models/statistic_activity";
import Calendar from "./calendar.svelte";

beforeAll(() => {
    addMessages("en", {
        activity: "{n, plural, =1 {Activity} other {Activities}}",
        calendar: {
            weekdays: [
                "Monday",
                "Tuesday",
                "Wednesday",
                "Thursday",
                "Friday",
                "Saturday",
                "Sunday",
            ],
        },
    });
    init({
        fallbackLocale: "en",
        initialLocale: "en",
        formats: {
            date: { monthName: { month: "long" } },
            number: {},
            time: {},
        },
    });
});

describe("Calendar", () => {
    it("announces activity counts on interactive days", () => {
        const activity = {
            date: "2026-09-15",
        } as StatisticActivity;

        const { body } = render(Calendar, {
            props: {
                month: "2026-09-01",
                activities: [activity],
                onclick: () => {},
            },
        });

        expect(body).toContain(
            'aria-label="Tuesday, September 15, 2026: 1 Activity"',
        );
    });
});
