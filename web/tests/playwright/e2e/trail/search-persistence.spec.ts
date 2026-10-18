import { expect, test, type BrowserContext, type Page } from '@playwright/test';

const resultsPath = '/api/v1/search/trails';
const clusterPath = `${resultsPath}/cluster`;
const input = (page: Page) => page.locator('.search-heading > input');
const categoryChip = (page: Page) => page.locator('[data-filter-key^="category:"]');

// Exercise the real router and its snapshot lifecycle, with anonymous access
// and empty search responses so no test trails or external geocoder are needed.
test.use({ storageState: { cookies: [], origins: [] }, viewport: { width: 1440, height: 1000 } });

test.beforeEach(async ({ context }) => {
    await context.addInitScript(() => {
        localStorage.setItem('searchLayout', JSON.stringify({ position: 'top', expansion: 'overlay' }));
    });
    await context.route(`**${resultsPath}`, route => {
        const { options } = route.request().postDataJSON();
        return route.fulfill({ json: {
            hits: [], totalHits: 0, totalPages: 0, page: options.page,
            hitsPerPage: options.hitsPerPage,
        } });
    });
    await context.route(`**${clusterPath}`, route => route.fulfill({ json: {
        type: 'FeatureCollection', features: [], totalHits: 0, hasOutsideBounds: true,
    } }));
    await context.route('**/api/v1/geocoding/search?*', route => route.fulfill({ json: {
        type: 'FeatureCollection', features: [],
    } }));
    await context.route('**/api/v1/search/lists', route => route.fulfill({ json: { hits: [] } }));
});

function nextResults(page: Page, q?: string) {
    return page.waitForResponse(response => response.request().method() === 'POST'
        && new URL(response.url()).pathname === resultsPath
        && (q === undefined || response.request().postDataJSON().q === q));
}

async function openTrails(page: Page, url = '/trails?view=list') {
    const response = nextResults(page);
    await page.goto(url);
    const loaded = await response;
    expect(loaded.status()).toBe(200);
    return loaded.request().postDataJSON();
}

async function seedOldSearch(context: BrowserContext) {
    await context.addInitScript(() => {
        localStorage.setItem('trailListFilter', JSON.stringify({
            q: 'OLD_SHARED_SEARCH', category: ['oldcategory0001'], tags: ['old-tag'],
            near: { lat: 10, lon: 20, radius: 1000 },
        }));
        localStorage.setItem('trailSearchBounds', JSON.stringify({
            northEast: { lat: 12, lng: 22 }, southWest: { lat: 8, lng: 18 },
        }));
    });
}

test('preserves independent tab filters on reload and back navigation', async ({ page, context }) => {
    const other = await context.newPage();
    await openTrails(page);
    await openTrails(other);
    const searches = [];
    for (const [index, tab] of [page, other].entries()) {
        await input(tab).click();
        await tab.locator('.activity-grid button').nth(index).click();
        await expect(categoryChip(tab)).toHaveCount(1);
        const key = await categoryChip(tab).getAttribute('data-filter-key');
        const q = `search in tab ${index + 1}`;
        const applied = nextResults(tab, q);
        await input(tab).fill(q);
        await applied;
        await input(tab).press('Escape');
        searches.push({ tab, q, key });
    }
    for (const { tab, q, key } of searches) {
        const restored = nextResults(tab, q);
        await tab.reload();
        await restored;
        await expect(input(tab)).toHaveValue(q);
        await expect(categoryChip(tab)).toHaveAttribute('data-filter-key', key!);
    }
    const first = searches[0];
    await page.getByRole('link', { name: 'Home', exact: true }).click();
    await expect(page).toHaveURL('/');
    const restored = nextResults(page, first.q);
    await page.goBack();
    await restored;
    await expect(input(page)).toHaveValue(first.q);
    await expect(categoryChip(page)).toHaveAttribute('data-filter-key', first.key!);
});

test('starts a fresh entry without stale shared criteria or bounds and keeps view preferences', async ({ page, context }) => {
    await seedOldSearch(context);
    await context.addInitScript(() => {
        localStorage.setItem('trailSearchSort', 'name');
        localStorage.setItem('trailSearchSortOrder', '-');
        localStorage.setItem('trailSearchView', 'table');
    });
    const body = await openTrails(page, '/trails');
    expect(body.q).toBe('');
    expect(body.options.filter).not.toContain('oldcategory0001');
    expect(body.options.filter).not.toContain('old-tag');
    expect(body.options.filter).not.toContain('_geoRadius');
    expect(body.options.filter).not.toContain('min_lon');
    expect(body.options.sort).toEqual(['name:desc']);
    await expect(input(page)).toHaveValue('');
    await expect(categoryChip(page)).toHaveCount(0);
    await expect(page).toHaveURL(/view=table/);
});

for (const query of ['lat=46&lon=7', 'tl_lat=49&tl_lon=11&br_lat=45&br_lon=5']) {
    test(`opens a fresh map bookmark with URL coordinates and no old filters: ${query}`, async ({ page, context }) => {
        await seedOldSearch(context);
        const cluster = page.waitForRequest(request => request.method() === 'POST'
            && new URL(request.url()).pathname === clusterPath);
        await page.goto(`/map?${query}`);
        const body = (await cluster).postDataJSON();
        expect(body.q).toBe('');
        expect(body.filterText).not.toContain('oldcategory0001');
        expect(body.filterText).not.toContain('_geoRadius');
        const center = {
            lat: (body.northEast.lat + body.southWest.lat) / 2,
            lon: (body.northEast.lng + body.southWest.lng) / 2,
        };
        expect(center.lat).toBeCloseTo(query.startsWith('lat=') ? 46 : 47, query.startsWith('lat=') ? 2 : 0);
        expect(center.lon).toBeCloseTo(query.startsWith('lat=') ? 7 : 8, 2);
        if (query.startsWith('lat=')) expect(body.zoom).toBe(14);
        else {
            expect(body.northEast.lat).toBeGreaterThanOrEqual(49 - 1e-7);
            expect(body.northEast.lng).toBeGreaterThanOrEqual(11 - 1e-7);
            expect(body.southWest.lat).toBeLessThanOrEqual(45 + 1e-7);
            expect(body.southWest.lng).toBeLessThanOrEqual(5 + 1e-7);
        }
        await expect(page).toHaveURL(/\/trails\?.*view=map/);
    });
}
