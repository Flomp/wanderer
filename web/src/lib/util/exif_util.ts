import ExifReader from "exifreader";
import type { ExpandedTags } from "exifreader";
import { convertDateTimeOriginalToDate, convertGPSTimestampToDate } from "./date_util";

export interface GPSCoordinates {
    lat: number;
    lon: number;
}

export interface GPSCoordinatesWithTime extends GPSCoordinates {
    time?: Date;
}

export interface PhotoExifMetadata {
    coordinates?: GPSCoordinates;
    takenAt?: string;
    // "exif" when takenAt is when the photo was taken by its own metadata,
    // not when it was scanned or last edited.
    takenAtSource?: "exif";
    // The day of the photo's local time, as YYYY-MM-DD.
    takenOn?: string;
}

type ExifLoadOptions = {
    expanded: true;
    includeTags?: {
        gps?: true;
        exif?: string[];
    };
    excludeTags?: {
        mpf?: true;
    };
};

const GPS_OPTIONS: ExifLoadOptions = {
    expanded: true,
    includeTags: { gps: true },
    excludeTags: { mpf: true },
};

const PHOTO_METADATA_OPTIONS: ExifLoadOptions = {
    expanded: true,
    includeTags: {
        gps: true,
        exif: [
            "GPSDateStamp",
            "GPSTimeStamp",
            "DateTimeOriginal",
            "DateTimeDigitized",
            "DateTime",
            "OffsetTimeOriginal",
            "OffsetTimeDigitized",
            "OffsetTime",
        ],
    },
    excludeTags: { mpf: true },
};

const GPS_TIME_OPTIONS: ExifLoadOptions = {
    expanded: true,
    includeTags: {
        gps: true,
        exif: [
            "GPSDateStamp",
            "GPSTimeStamp",
            "DateTimeOriginal",
            "OffsetTimeOriginal",
            "OffsetTime",
            "OffsetTimeDigitized",
        ],
    },
    excludeTags: { mpf: true },
};

// EXIF dates are the camera's local time, "YYYY:MM:DD HH:MM:SS".
const EXIF_DATE_TIME = /^(\d{4}):(\d{2}):(\d{2})[ T](\d{2}):(\d{2}):(\d{2})/;

export async function extractGPSCoordinates(
    source: Blob | string | ArrayBuffer | SharedArrayBuffer,
): Promise<GPSCoordinates | undefined> {
    try {
        const tags = await loadExifTags(source, GPS_OPTIONS);
        return gpsCoordinatesFromExpandedTags(tags);
    } catch {
        return undefined;
    }
}

// Unlike takenAt, the time is only set when it is exact, since it picks the
// pass of a track that covers the same ground twice.
export async function extractGPSCoordinatesAndTime(
    source: Blob | string | ArrayBuffer | SharedArrayBuffer,
): Promise<GPSCoordinatesWithTime | undefined> {
    try {
        const tags = await loadExifTags(source, GPS_TIME_OPTIONS);
        const coordinates = gpsCoordinatesFromExpandedTags(tags);
        if (!coordinates) {
            return undefined;
        }
        return { ...coordinates, time: exactTimeFromExpandedTags(tags) };
    } catch {
        return undefined;
    }
}

export async function extractPhotoExifMetadata(
    source: Blob | string | ArrayBuffer | SharedArrayBuffer,
): Promise<PhotoExifMetadata> {
    try {
        const tags = await loadExifTags(source, PHOTO_METADATA_OPTIONS);
        return {
            coordinates: gpsCoordinatesFromExpandedTags(tags),
            ...takenAtFromExpandedTags(tags),
        };
    } catch {
        return {};
    }
}

async function loadExifTags(
    source: Blob | string | ArrayBuffer | SharedArrayBuffer,
    options: ExifLoadOptions,
): Promise<ExpandedTags> {
    if (typeof source === "string") {
        return ExifReader.load(source, options);
    }
    if (isBlobLike(source)) {
        return ExifReader.load(await source.arrayBuffer(), options);
    }
    return ExifReader.load(source, options);
}

function gpsCoordinatesFromExpandedTags(tags: ExpandedTags): GPSCoordinates | undefined {
    const lat = tags.gps?.Latitude;
    const lon = tags.gps?.Longitude;
    if (
        typeof lat !== "number" ||
        typeof lon !== "number" ||
        !Number.isFinite(lat) ||
        !Number.isFinite(lon)
    ) {
        return undefined;
    }
    return { lat, lon };
}

// For photos added to the library, the time they were taken is the GPS time,
// which is UTC, else DateTimeOriginal in its offset or, without one, as UTC.
// The other dates can be when the photo was scanned or last edited, so they
// only stand in for it.
function takenAtFromExpandedTags(tags: ExpandedTags): Omit<PhotoExifMetadata, "coordinates"> {
    const original = tagString(tags.exif?.DateTimeOriginal);
    const value =
        original ??
        tagString(tags.exif?.DateTimeDigitized) ??
        tagString(tags.exif?.DateTime);
    const offset =
        tagString(tags.exif?.OffsetTimeOriginal) ??
        tagString(tags.exif?.OffsetTimeDigitized) ??
        tagString(tags.exif?.OffsetTime);
    const takenOn = value ? exifDay(value) : undefined;

    const capturedAt =
        gpsTimeFromExpandedTags(tags) ??
        (original && EXIF_DATE_TIME.test(original) ? parseExifDate(original, offset) : undefined);
    if (capturedAt) {
        return { takenAt: capturedAt, takenAtSource: "exif", takenOn };
    }
    return { takenAt: value ? parseExifDate(value, offset) : undefined, takenOn };
}

// Devices without a GPS fix may leave the date zero.
function gpsTimeFromExpandedTags(tags: ExpandedTags): string | undefined {
    const dateStamp = tagString(tags.exif?.GPSDateStamp);
    const date = dateStamp?.match(/^(\d{4}):(\d{2}):(\d{2})$/);
    if (!date || !isExifDay(date[1], date[2], date[3])) {
        return undefined;
    }
    return convertGPSTimestampToDate(dateStamp, gpsTimeStampValues(tags.exif?.GPSTimeStamp))?.toISOString();
}

function exactTimeFromExpandedTags(tags: ExpandedTags): Date | undefined {
    return (
        convertGPSTimestampToDate(
            tagString(tags.exif?.GPSDateStamp),
            gpsTimeStampValues(tags.exif?.GPSTimeStamp),
        ) ??
        convertDateTimeOriginalToDate(
            tagString(tags.exif?.DateTimeOriginal),
            // some devices only write the offset of the other timestamps, all
            // in one zone
            tagString(tags.exif?.OffsetTimeOriginal) ??
                tagString(tags.exif?.OffsetTime) ??
                tagString(tags.exif?.OffsetTimeDigitized),
        )
    );
}

// exifreader reads the hours, minutes and seconds of GPSTimeStamp as
// [numerator, denominator] pairs.
function gpsTimeStampValues(tag: { value?: unknown } | undefined): number[] | undefined {
    const value = tag?.value;
    if (!Array.isArray(value)) {
        return undefined;
    }
    return value.map((part: unknown) =>
        Array.isArray(part) ? Number(part[0]) / Number(part[1]) : Number(part),
    );
}

function tagString(tag: { value?: unknown; description?: string } | undefined): string | undefined {
    if (!tag) {
        return undefined;
    }
    const value = Array.isArray(tag.value) ? tag.value[0] : tag.value;
    if (typeof value === "string" && value.trim()) {
        return value.trim();
    }
    if (typeof tag.description === "string" && tag.description.trim()) {
        return tag.description.trim();
    }
    return undefined;
}

function parseExifDate(value: string, offset?: string): string | undefined {
    const match = value.match(EXIF_DATE_TIME);
    if (!match) {
        const fallback = new Date(value);
        return Number.isNaN(fallback.getTime()) ? undefined : fallback.toISOString();
    }

    const [, year, month, day, hour, minute, second] = match;
    if (!isExifDay(year, month, day)) {
        return undefined;
    }

    // Without an offset the local time counts as UTC, so that a photo gets the
    // same time wherever it is read.
    const isoDate = `${year}-${month}-${day}T${hour}:${minute}:${second}`;
    const date = new Date(`${isoDate}${normalizeExifOffset(offset) ?? "Z"}`);
    return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}

function exifDay(value: string): string | undefined {
    const match = value.match(EXIF_DATE_TIME);
    return match && isExifDay(match[1], match[2], match[3]) ? `${match[1]}-${match[2]}-${match[3]}` : undefined;
}

// Cameras without a set clock write zeros.
function isExifDay(year: string, month: string, day: string): boolean {
    return year !== "0000" && month !== "00" && day !== "00";
}

function normalizeExifOffset(offset?: string): string | undefined {
    if (!offset) {
        return undefined;
    }
    const match = offset.trim().match(/^([+-])(\d{2}):?(\d{2})$/);
    if (!match) {
        return undefined;
    }
    return `${match[1]}${match[2]}:${match[3]}`;
}

function isBlobLike(source: unknown): source is Blob {
    return (
        typeof source === "object" &&
        source !== null &&
        "arrayBuffer" in source &&
        typeof source.arrayBuffer === "function"
    );
}
