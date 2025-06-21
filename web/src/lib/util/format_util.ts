import { browser } from "$app/environment";
import { page } from "$app/state";
import { get } from "svelte/store";

export function formatTimeHHMM(minutes?: number) {
    if (!minutes) {
        return "-";
    }
    const m = minutes % 60;

    const h = (minutes - m) / 60;

    return (h < 10 ? "0" : "") + h.toString() + "h " + (Math.round(m) < 10 ? "0" : "") + Math.round(m).toString() + "m";
}

export function formatDistance(meters?: number) {
    if (meters === undefined) {
        return "-";
    }

    const unit = page.data.settings?.unit ?? "metric";

    if (unit == "metric") {
        if (meters >= 1000) {
            return `${(meters / 1000).toFixed(2)} km`
        } else {
            return meters % 1 == 0 ? `${meters} m` : `${Math.round(meters)} m`;
        }
    } else {
        const miles = meters * 0.000621371;
        const roundedMiles = miles.toFixed(2);

        return `${roundedMiles} mi`;
    }
}

export function formatElevation(meters?: number) {
    if (meters === undefined) {
        return "-";
    }

    const unit = page.data.settings?.unit ?? "metric";

    if (unit == "metric") {
        return `${Math.round(meters)} m`
    } else {
        const feet = meters * 3.28084;

        return `${Math.round(feet)} ft`;
    }
}

export function formatSpeed(speed?: number) {
    if (speed === undefined) {
        return "-";
    }

    const unit = page.data.settings?.unit ?? "metric";

    if (unit == "metric") {
        return `${(speed * 3.6).toFixed(2)} km/h`
    } else {
        const mph = speed * 3.6 * 0.621371;

        return `${Math.round(mph)} mp/h`;
    }
}

export function formatTimeSince(date: Date) {
    const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);

    let interval = seconds / 31536000;
    if (interval > 1) {
        return { unit: "years", value: Math.floor(interval) };
    }
    interval = seconds / 2592000;
    if (interval > 1) {
        return { unit: "months", value: Math.floor(interval) };
    }
    interval = seconds / 86400;
    if (interval > 1) {
        return { unit: "days", value: Math.floor(interval) };
    }
    interval = seconds / 3600;
    if (interval > 1) {
        return { unit: "hours", value: Math.floor(interval) };
    }
    interval = seconds / 60;
    if (interval > 1) {
        return { unit: "minutes", value: Math.floor(interval) };
    }
    return { unit: "seconds", value: seconds };
}

export function formatHTMLAsText(html?: string) {
    if(!html || !browser) {
        return ""
    }
    // Create a temporary DOM element
    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;

    // Replace <br> with newlines to preserve line breaks
    tempDiv.querySelectorAll("br").forEach((br) => br.replaceWith("\n"));

    // Replace block-level elements with newlines before and after
    const blockTags = new Set([
        "DIV",
        "P",
        "LI",
        "SECTION",
        "ARTICLE",
        "HEADER",
        "FOOTER",
        "ASIDE",
        "MAIN",
        "NAV",
        "FIGURE",
        "TABLE",
        "TR",
        "TD",
        "TH",
        "UL",
        "OL",
        "PRE",
    ]);
    tempDiv.querySelectorAll("*").forEach((el) => {
        if (blockTags.has(el.tagName)) {
            el.insertAdjacentText("beforebegin", "\n");
            el.insertAdjacentText("afterend", "\n");
        }
    });

    // Extract the text content
    let text = tempDiv.textContent;

    if (!text) {
        return "";
    }

    // Replace multiple spaces and newlines with a single one if needed
    // Optional: collapse excessive blank lines to max two
    text = text
        .replace(/[ \t]+\n/g, "\n") // trailing spaces
        .replace(/\n[ \t]+/g, "\n") // leading spaces
        .replace(/\n{3,}/g, "\n\n") // collapse 3+ newlines
        .replace(/[ \t]{2,}/g, "  "); // collapse multiple spaces to two

    // Trim the result
    return text.trim();
}