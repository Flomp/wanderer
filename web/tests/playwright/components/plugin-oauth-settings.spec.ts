import { expect, test } from '@playwright/test';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

function source(status: string, unreadable = false) {
    return `<script>
        import Settings from '$lib/components/settings/plugins/plugin_instance_settings_modal.svelte';
        let modal;
        const auth = { type: 'oauth2', fields: ['clientId', 'clientSecret'], secretFields: ['clientSecret'] };
        const plugin = { id: 'oauth', name: 'OAuth', type: 'trails', status: 'available',
            auth, authContext: 'oauth', authContexts: { oauth: auth }, hostConfig: { merge: { available: false } } };
        const instance = { id: 'abcdefghijklmn1', plugin_id: 'oauth', user: 'tester', enabled: false,
            status: '${status}', auth_context: 'oauth', connection_generation: 'read-generation', busy: true,
            auth: { clientId: 'client', clientSecret: '', accessToken: '', refreshToken: '', oauthContext: 'oauth' },
            last_error: ${unreadable ? "{ code: 'credentials_unreadable' }" : 'undefined'} };
        function save(candidate) {
            window.modalCalls.push(JSON.parse(JSON.stringify(candidate)));
            return { ...instance };
        }
        </script>
        <button data-testid="open" onclick={() => modal.openModal()}>Settings</button>
        <Settings {plugin} {instance} onsave={save} bind:this={modal} />`;
}

for (const dev of [true, false]) {
    test.describe(`OAuth settings (${dev ? 'development' : 'production'})`, () => {
        for (const status of ['disabled', 'rate_limited', 'unavailable', 'configured']) {
            test(`plain Save retains grant indication and activation at ${status}`, async ({ page }) => {
                const bundle = await buildTrailModalFixture(dev, source(status));
                await mountTrailModalFixture(page, bundle);
                await page.getByTestId('open').click();
                const dialog = page.locator('dialog[open]');
                await expect(dialog).toContainText('plugin-oauth-material-present-hint');
                await dialog.locator('button[name="save"]').click();
                expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([{ id: 'abcdefghijklmn1', plugin_id: 'oauth' }]);
            });
        }

        for (const unreadable of [false, true]) {
            test(`${unreadable ? 'repair clears unreadable grant then reconnects' : 'reconnect keeps saved grant'} with an activation request`, async ({ page }) => {
                const bundle = await buildTrailModalFixture(dev, source('disabled', unreadable));
                await mountTrailModalFixture(page, bundle);
                const starts: unknown[] = [];
                const revokes: unknown[] = [];
                await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                    await route.fulfill({ json: { id: 'abcdefghijklmn1', plugin_id: 'oauth', enabled: false,
                        auth_context: 'oauth', connection_generation: 'revoked-generation', status: 'disabled', auth: { clientId: 'client', clientSecret: '' } } });
                });
                await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                    revokes.push(route.request().postDataJSON());
                    await route.fulfill({ json: { ok: true, connection_generation: 'revoked-generation' } });
                });
                await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                    starts.push(route.request().postDataJSON());
                    await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'flow-state', instanceId: 'abcdefghijklmn1' } });
                });
                await page.getByTestId('open').click();
                await page.locator('dialog[open] button').filter({ hasText: 'save-and-reconnect' }).click();
                await expect.poll(() => starts.length).toBe(1);
                expect(starts[0]).toEqual({ pluginId: 'oauth', instanceId: 'abcdefghijklmn1', authContext: 'oauth',
                    redirectUri: 'http://trail-modals.test/settings/plugins/oauth/callback',
                    expected_connection_generation: unreadable ? 'revoked-generation' : 'read-generation', activate_on_success: true });
                expect(revokes).toEqual(unreadable ? [{ instanceId: 'abcdefghijklmn1', expected_connection_generation: 'read-generation' }] : []);
                expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([{ id: 'abcdefghijklmn1', plugin_id: 'oauth' }]);
                await expect.poll(() => page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_flow-state'))).toBe('abcdefghijklmn1');
            });
        }

        for (const action of ['start', 'revoke']) {
            test(`${action} keeps the opening generation after a refreshed connection`, async ({ page }) => {
                const liveSource = source('disabled').replace('const instance = {', 'let instance = $state({')
                    .replace('last_error: undefined };', `last_error: undefined });
                        window.advanceConnection = () => { instance = { ...instance, connection_generation: 'new-generation',
                            auth: { ...instance.auth, clientId: 'another-client' } }; };`);
                const bundle = await buildTrailModalFixture(dev, liveSource);
                await mountTrailModalFixture(page, bundle);
                const requests: any[] = [];
                await page.route(`**/api/v1/plugin-system/oauth/${action}`, async route => {
                    requests.push(route.request().postDataJSON());
                    await route.fulfill({ status: 409, json: { status: 409, message: 'safe',
                        data: { operation: { code: 'plugin_connection_conflict', message: 'safe' } } } });
                });
                await page.getByTestId('open').click();
                await page.evaluate(() => (window as any).advanceConnection());
                const dialog = page.locator('dialog[open]');
                await expect(dialog).toContainText('plugin-error-connection-conflict');
                await dialog.locator('button').filter({ hasText: action === 'start' ? 'save-and-reconnect' : 'plugin-disconnect' }).click();
                await expect.poll(() => requests.length).toBe(1);
                expect(requests[0].expected_connection_generation).toBe('read-generation');
                if (action === 'start') {
                    expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([{ id: 'abcdefghijklmn1', plugin_id: 'oauth' }]);
                }
                await expect(dialog).toHaveCount(1);
            });
        }

        test('stale unreadable-grant draft cannot revoke a replacement connection', async ({ page }) => {
            const liveSource = source('disabled', true).replace('const instance = {', 'let instance = $state({')
                .replace("last_error: { code: 'credentials_unreadable' } };", `last_error: { code: 'credentials_unreadable' } });
                    window.advanceConnection = () => { instance = { ...instance, connection_generation: 'healthy-new-generation', last_error: undefined }; };`);
            const bundle = await buildTrailModalFixture(dev, liveSource);
            await mountTrailModalFixture(page, bundle);
            const revokes: any[] = [];
            let starts = 0;
            await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                revokes.push(route.request().postDataJSON());
                await route.fulfill({ status: 409, json: { status: 409, message: 'safe',
                    data: { operation: { code: 'plugin_connection_conflict', message: 'safe' } } } });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                starts++;
                await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'forbidden-flow', instanceId: 'abcdefghijklmn1' } });
            });
            await page.getByTestId('open').click();
            await page.evaluate(() => (window as any).advanceConnection());
            const dialog = page.locator('dialog[open]');
            await expect(dialog).toContainText('plugin-error-connection-conflict');
            await dialog.locator('button').filter({ hasText: 'save-and-reconnect' }).click();
            await expect.poll(() => revokes.length).toBe(1);
            await expect(dialog.locator('button').filter({ hasText: 'save-and-reconnect' })).toBeEnabled();
            expect(revokes[0]).toEqual({ instanceId: 'abcdefghijklmn1', expected_connection_generation: 'read-generation' });
            expect(starts).toBe(0);
            await expect(dialog).toHaveCount(1);
        });

        test('repair start pins its revoke commit even when the next read sees another connection', async ({ page }) => {
            const bundle = await buildTrailModalFixture(dev, source('disabled', true));
            await mountTrailModalFixture(page, bundle);
            const starts: any[] = [];
            await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                expect(route.request().postDataJSON().expected_connection_generation).toBe('read-generation');
                await route.fulfill({ json: { ok: true, connection_generation: 'own-revoke-generation' } });
            });
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                await route.fulfill({ json: { id: 'abcdefghijklmn1', plugin_id: 'oauth', enabled: true,
                    auth_context: 'oauth', connection_generation: 'healthy-new-generation', status: 'configured',
                    auth: { clientId: 'other-client', clientSecret: '', accessToken: '', oauthContext: 'oauth' } } });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                starts.push(route.request().postDataJSON());
                await route.fulfill({ status: 409, json: { status: 409, message: 'safe',
                    data: { operation: { code: 'plugin_connection_conflict', message: 'safe' } } } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('button').filter({ hasText: 'save-and-reconnect' }).click();
            await expect.poll(() => starts.length).toBe(1);
            expect(starts[0].expected_connection_generation).toBe('own-revoke-generation');
            expect(await page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_forbidden-flow'))).toBeNull();
            await expect(dialog).toHaveCount(1);
        });
    });
}
