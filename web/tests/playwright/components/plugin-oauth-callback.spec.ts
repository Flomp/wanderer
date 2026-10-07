import { expect, test } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { buildTrailModalFixture } from './trail-modal-fixture';

const callbackPath = fileURLToPath(new URL('../../../src/routes/settings/plugins/oauth/callback/+page.svelte', import.meta.url));
const source = `<script>import Callback from ${JSON.stringify(callbackPath)};</script><Callback />`;

for (const dev of [true, false]) {
    test.describe(`plugin OAuth callback (${dev ? 'development' : 'production'})`, () => {
        let bundle: string;
        test.beforeAll(async () => { bundle = await buildTrailModalFixture(dev, source); });

        for (const scenario of [
            { name: 'local save failure', status: 400,
                response: { message: 'Failed to save plugin connection.', detail: { status: 400, data: {} } },
                text: 'error-completing-oauth' },
            { name: 'typed provider failure', status: 400,
                response: { message: 'Plugin operation failed.', detail: { data: { operation: { code: 'rate_limited', message: 'private-provider-sentinel' } } } },
                text: 'plugin-error-rate-limited' },
            { name: 'successful connection', status: 200, response: { ok: true }, text: '' },
        ]) {
            test(scenario.name, async ({ page }) => {
                const pageErrors: string[] = [];
                const requests: unknown[] = [];
                page.on('pageerror', error => pageErrors.push(error.message));
                await page.route('http://plugin-oauth.test/**', async route => {
                    if (new URL(route.request().url()).pathname === '/api/v1/plugin-system/oauth/callback') {
                        requests.push(route.request().postDataJSON());
                        await route.fulfill({ status: scenario.status, json: scenario.response });
                        return;
                    }
                    await route.fulfill({ contentType: 'text/html', body: '<!doctype html><html><body></body></html>' });
                });
                await page.goto('http://plugin-oauth.test/?code=test-code&state=test-state');
                await page.evaluate(() => sessionStorage.setItem('wanderer_plugin_oauth_test-state', 'test-instance'));
                await page.addScriptTag({ content: bundle });
                if (scenario.status === 200) {
                    await expect.poll(() => page.evaluate(() => (window as any).fixtureNavigation)).toBe('/settings/plugins');
                    await expect(page.locator('p')).toHaveCount(0);
                } else {
                    await expect(page.locator('p')).toHaveText(scenario.text);
                    await expect(page.getByRole('link')).toHaveAttribute('href', '/settings/plugins');
                    expect(await page.evaluate(() => (window as any).fixtureNavigation)).toBeUndefined();
                }
                expect(requests).toEqual([{ instanceId: 'test-instance', code: 'test-code', state: 'test-state' }]);
                expect(await page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_test-state'))).toBeNull();
                await expect(page.locator('body')).not.toContainText('private-provider-sentinel');
                expect(pageErrors).toEqual([]);
            });
        }
    });
}
