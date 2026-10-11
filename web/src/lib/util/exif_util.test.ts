import { describe, expect, it } from "vitest";
import { extractGPSCoordinatesAndTime } from "./exif_util";

interface ExifEntry {
    tag: number;
    type: number;
    count: number;
    bytes: number[];
}

const ASCII = 2;
const LONG = 4;
const RATIONAL = 5;

function uint16(value: number) {
    return [value & 0xff, (value >> 8) & 0xff];
}

function uint32(value: number) {
    return [value & 0xff, (value >> 8) & 0xff, (value >> 16) & 0xff, (value >>> 24) & 0xff];
}

function ascii(tag: number, value: string): ExifEntry {
    const bytes = [...value].map((c) => c.charCodeAt(0)).concat(0);
    return { tag, type: ASCII, count: bytes.length, bytes };
}

function rationals(tag: number, values: [number, number][]): ExifEntry {
    return {
        tag,
        type: RATIONAL,
        count: values.length,
        bytes: values.flatMap(([numerator, denominator]) => [...uint32(numerator), ...uint32(denominator)]),
    };
}

function ifd(entries: ExifEntry[], offset: number) {
    const dataOffset = offset + 2 + entries.length * 12 + 4;
    const head = uint16(entries.length);
    const data: number[] = [];
    for (const entry of entries) {
        head.push(...uint16(entry.tag), ...uint16(entry.type), ...uint32(entry.count));
        if (entry.bytes.length <= 4) {
            head.push(...entry.bytes, ...new Array(4 - entry.bytes.length).fill(0));
        } else {
            head.push(...uint32(dataOffset + data.length));
            data.push(...entry.bytes);
        }
    }
    return [...head, ...uint32(0), ...data];
}

// A JPEG holding only an APP1 segment with the given Exif and GPS IFD entries.
function jpeg(exif: ExifEntry[], gps: ExifEntry[]) {
    const exifOffset = 8 + 2 + 2 * 12 + 4;
    const exifIfd = ifd(exif, exifOffset);
    const gpsOffset = exifOffset + exifIfd.length;
    const tiff = [
        0x49, 0x49, ...uint16(42), ...uint32(8),
        ...ifd([
            { tag: 0x8769, type: LONG, count: 1, bytes: uint32(exifOffset) },
            { tag: 0x8825, type: LONG, count: 1, bytes: uint32(gpsOffset) },
        ], 8),
        ...exifIfd,
        ...ifd(gps, gpsOffset),
    ];
    const app1 = [0x45, 0x78, 0x69, 0x66, 0, 0, ...tiff];
    const length = app1.length + 2;
    return new Uint8Array([0xff, 0xd8, 0xff, 0xe1, length >> 8, length & 0xff, ...app1, 0xff, 0xd9]).buffer;
}

const position = [
    ascii(0x0001, "N"),
    rationals(0x0002, [[46, 1], [30, 1], [0, 1]]),
    ascii(0x0003, "W"),
    rationals(0x0004, [[7, 1], [15, 1], [0, 1]]),
];

describe("extractGPSCoordinatesAndTime", () => {
    it("reads the position and prefers the GPS timestamp, which is UTC", async () => {
        const photo = jpeg(
            [ascii(0x9003, "2026:10:03 08:00:00"), ascii(0x9011, "+02:00")],
            [...position, rationals(0x0007, [[14, 1], [6, 1], [335, 10]]), ascii(0x001d, "2026:10:03")],
        );

        const result = await extractGPSCoordinatesAndTime(new Blob([photo]));

        expect(result?.lat).toBeCloseTo(46.5);
        expect(result?.lon).toBeCloseTo(-7.25);
        expect(result?.time?.toISOString()).toBe("2026-10-03T14:06:33.500Z");
    });

    it("falls back to the local time with its offset", async () => {
        const photo = jpeg([ascii(0x9003, "2026:10:03 10:06:33"), ascii(0x9011, "-04:00")], position);

        expect((await extractGPSCoordinatesAndTime(photo))?.time?.toISOString()).toBe("2026-10-03T14:06:33.000Z");
    });

    it("uses the offset of the other timestamps when OffsetTimeOriginal is missing", async () => {
        const photo = jpeg(
            [ascii(0x9003, "2026:10:03 01:30:00"), ascii(0x9010, "+05:30"), ascii(0x9012, "+01:00")],
            position,
        );

        expect((await extractGPSCoordinatesAndTime(photo))?.time?.toISOString()).toBe("2026-10-02T20:00:00.000Z");
    });

    it("leaves out a local time without an offset", async () => {
        const result = await extractGPSCoordinatesAndTime(jpeg([ascii(0x9003, "2026:10:03 10:06:33")], position));

        expect(result?.lat).toBeCloseTo(46.5);
        expect(result?.time).toBeUndefined();
    });

    it("ignores photos without a position", async () => {
        const photo = jpeg([ascii(0x9003, "2026:10:03 10:06:33"), ascii(0x9011, "-04:00")], []);

        expect(await extractGPSCoordinatesAndTime(photo)).toBeUndefined();
    });
});
