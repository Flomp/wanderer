import { afterEach, describe, expect, it, vi } from 'vitest';
import { MeilisearchApiError } from 'meilisearch';
import { getHTTPErrorStatus, handleError } from './api_util';
import { sanitizeTrailSort, sanitizeTrailSortOrder, trailFilterDateBoundary } from './trail_filter_util';

describe('search input and error boundaries', () => {
    afterEach(() => vi.unstubAllEnvs());

    it.each([
        ['2026-03-29', '2026-03-28T23:00:00Z', '2026-03-29T22:00:00Z'],
        ['2026-10-25', '2026-10-24T22:00:00Z', '2026-10-25T23:00:00Z'],
        ['2026-09-07', '2026-09-06T22:00:00Z', '2026-09-07T22:00:00Z'],
    ])('includes the entire local day %s', (day, start, end) => {
        vi.stubEnv('TZ', 'Europe/Zurich');
        expect(trailFilterDateBoundary(day)).toBe(Date.parse(start) / 1000);
        expect(trailFilterDateBoundary(day, true)).toBe(Date.parse(end) / 1000);
    });

    it('does not compile invalid calendar values', () => {
        for (const value of [undefined, '', 'invalid', '2026-02-30']) {
            expect(trailFilterDateBoundary(value)).toBeUndefined();
        }
    });

    it('retains valid sorting and falls back for invalid stored values', () => {
        expect(sanitizeTrailSort('distance', 'created')).toBe('distance');
        expect(sanitizeTrailSort('unknown', 'created')).toBe('created');
        expect(sanitizeTrailSortOrder('+', '-')).toBe('+');
        expect(sanitizeTrailSortOrder('raw', '-')).toBe('-');
    });

    it.each([400, 403, 503])('preserves the real SDK HTTP status %s', status => {
        const failure = new MeilisearchApiError(new Response(null, { status }));
        expect(getHTTPErrorStatus(failure)).toBe(status);
        expect(handleError(failure).status).toBe(status);
    });

    it('treats transport failures as server errors', () => {
        expect(getHTTPErrorStatus(new Error('network failed'))).toBe(500);
    });
});
