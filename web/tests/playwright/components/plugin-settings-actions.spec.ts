import { expect, test } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

const pagePath = fileURLToPath(new URL('../../../src/routes/settings/plugins/+page.svelte', import.meta.url));
const source = `<script>
    import Plugins from ${JSON.stringify(pagePath)};
    let ready = $state(false);
    const auth = { type: 'session', fields: ['email', 'password'], secretFields: ['password'] };
    const plugin = { id: 'session', name: 'Session', type: 'trails', status: 'available', auth,
        authContext: 'session', authContexts: { session: auth }, capabilities: ['list_routes.v1'],
        configSchema: [{ key: 'baseURL', type: 'url', hidden: true, required: true }] };
    const instance = { id: 'abcdefghijklmn1', user: 'owner', plugin_id: 'session', enabled: true, status: 'needs_reauth',
        auth_context: 'session', connection_generation: 'generation', auth: { email: 'stored@example.org', password: '' },
        config: { plugin: { baseURL: '' }, host: { categoryMappingUpdatedAt: '2020-01-01' } }, state: { cursor: 'preserve' } };
    const data = { pluginInstances: [instance], pluginProviders: [plugin], categories: [], subcategories: [] };
    </script>
    <button data-testid="mount-plugins" onclick={() => ready = true}>Plugins</button>
    {#if ready}<Plugins {data} />{/if}`;

for (const dev of [true, false]) {
    test.describe(`plugin settings actions (${dev ? 'development' : 'production'})`, () => {
        let bundle: string;
        test.beforeAll(async () => { bundle = await buildTrailModalFixture(dev, source); });

        test('disable remains available with invalid config and an authentication diagnosis', async ({ page }) => {
            await mountTrailModalFixture(page, bundle);
            let patch: unknown;
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                patch = route.request().postDataJSON();
                await route.fulfill({ json: { id: 'abcdefghijklmn1', plugin_id: 'session', enabled: false,
                    status: 'disabled', auth_context: 'session', connection_generation: 'generation' } });
            });
            await page.getByTestId('mount-plugins').click();
            const toggle = page.locator('.plugin-card-toggle input[type="checkbox"]');
            await expect(toggle).toBeEnabled();
            await toggle.uncheck();
            await expect.poll(() => patch).toEqual({ enabled: false });
        });

        test('dismiss sends only its explicit host metadata patch', async ({ page }) => {
            await mountTrailModalFixture(page, bundle);
            let patch: any;
            await page.route('**/api/v1/plugin-system/category-remap/preview', async route => {
                await route.fulfill({ json: { count: 3, backfilledSinceMapping: 1 } });
            });
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                patch = route.request().postDataJSON();
                await route.fulfill({ json: { id: 'abcdefghijklmn1', plugin_id: 'session', enabled: true,
                    status: 'needs_reauth', auth_context: 'session', connection_generation: 'generation' } });
            });
            await page.getByTestId('mount-plugins').click();
            await page.locator('dialog[open] button').filter({ hasText: 'plugin-category-remap-ignore' }).click();
            await expect.poll(() => patch).toBeDefined();
            expect(patch).toEqual({ config: { host: { categoryRemapDismissedAt: expect.any(String) } } });
        });
    });
}
