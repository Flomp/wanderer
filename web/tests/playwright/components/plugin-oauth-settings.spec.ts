import { expect, test } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

const callbackPath = fileURLToPath(new URL('../../../src/routes/settings/plugins/oauth/callback/+page.svelte', import.meta.url));
const callbackSource = `<script>import Callback from ${JSON.stringify(callbackPath)};</script><Callback />`;

function source(status: string, unreadable = false, enabled = false) {
    return `<script>
        import Settings from '$lib/components/settings/plugins/plugin_instance_settings_modal.svelte';
        let modal;
        const auth = { type: 'oauth2', fields: ['clientId', 'clientSecret'], secretFields: ['clientSecret'] };
        const plugin = { id: 'oauth', name: 'OAuth', type: 'trails', status: 'available',
            auth, authContext: 'oauth', authContexts: { oauth: auth }, hostConfig: { merge: { available: false } } };
        const instance = { id: 'abcdefghijklmn1', plugin_id: 'oauth', user: 'tester', enabled: ${enabled},
            status: '${status}', auth_context: 'oauth', connection_generation: 'read-generation', updated: '2026-10-08 10:00:00.000Z', busy: true,
            auth: { clientId: 'client', clientSecret: '', accessToken: '', refreshToken: '', oauthContext: 'oauth' },
            last_error: ${unreadable ? "{ code: 'credentials_unreadable' }" : 'undefined'} };
        window.fixturePluginInstance = () => instance;
        function save(candidate) {
            window.modalCalls.push(JSON.parse(JSON.stringify(candidate)));
            return { ...instance };
        }
        </script>
        <button data-testid="open" onclick={() => modal.openModal()}>Settings</button>
        <Settings {plugin} {instance} onsave={save} bind:this={modal} />`;
}

function activationSource(kind: 'new' | 'repair' | 'active' | 'disabled' | 'disconnected') {
    const initial = kind === 'new' ? 'undefined' : JSON.stringify({
        id: 'abcdefghijklmn1', plugin_id: 'oauth', user: 'tester', enabled: kind === 'repair' || kind === 'active',
        status: kind === 'repair' || kind === 'active' ? 'configured' : 'disabled', auth_context: 'oauth',
        connection_generation: 'read-generation',
        updated: '2026-10-08 10:00:00.000Z',
        auth: { clientId: 'client', clientSecret: '', ...(kind === 'disconnected' ? {} : {
            accessToken: '', refreshToken: '', oauthContext: 'oauth',
        }) },
        ...(kind === 'repair' ? { last_error: { code: 'credentials_unreadable' } } : {}),
    });
    return `<script>
        import Settings from '$lib/components/settings/plugins/plugin_instance_settings_modal.svelte';
        let modal;
        const auth = { type: 'oauth2', fields: ['clientId', 'clientSecret'], secretFields: ['clientSecret'] };
        const plugin = { id: 'oauth', name: 'OAuth', type: 'trails', status: 'available',
            auth, authContext: 'oauth', authContexts: { oauth: auth }, hostConfig: { merge: { available: false } } };
        let instance = $state(${initial});
        window.fixturePluginInstance = () => instance;
        window.fixtureRevokeApplied = () => {
            instance = { ...instance, enabled: false, connection_generation: 'revoked-generation', updated: '2026-10-08 10:01:00.000Z',
                auth: { clientId: 'client', clientSecret: '' }, last_error: undefined };
        };
        function save(candidate) {
            window.modalCalls.push(JSON.parse(JSON.stringify(candidate)));
            return instance ?? { id: 'abcdefghijklmn1', plugin_id: 'oauth', user: 'tester', enabled: false,
                status: 'disabled', auth_context: 'oauth', connection_generation: 'created-generation',
                auth: { clientId: candidate.auth.clientId, clientSecret: '' } };
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
            test(`${unreadable ? 'repair clears unreadable grant then reconnects' : 'reconnect keeps saved grant'} preserves disabled activation choice`, async ({ page }) => {
                const bundle = await buildTrailModalFixture(dev, source('disabled', unreadable));
                await mountTrailModalFixture(page, bundle);
                const starts: unknown[] = [];
                const revokes: unknown[] = [];
                let reads = 0;
                await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                    if (reads++ === 0) {
                        await route.fulfill({ json: await page.evaluate(() => (window as any).fixturePluginInstance()) });
                        return;
                    }
                    await route.fulfill({ json: { id: 'abcdefghijklmn1', plugin_id: 'oauth', enabled: false,
                        auth_context: 'oauth', connection_generation: 'revoked-generation', status: 'disabled', auth: { clientId: 'client', clientSecret: '' } } });
                });
                await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                    revokes.push(route.request().postDataJSON());
                    await route.fulfill({ json: { ok: true, connection_generation: 'revoked-generation', updated: '2026-10-08 10:01:00.000Z' } });
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
                    expected_connection_generation: unreadable ? 'revoked-generation' : 'read-generation', activate_on_success: false });
                expect(revokes).toEqual(unreadable ? [{ instanceId: 'abcdefghijklmn1', expected_connection_generation: 'read-generation' }] : []);
                expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([{ id: 'abcdefghijklmn1', plugin_id: 'oauth' }]);
                await expect.poll(() => page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_flow-state'))).toBe('abcdefghijklmn1');
            });
        }

        for (const scenario of [
            { kind: 'new', initial: true, requested: true },
            { kind: 'repair', initial: true, requested: true },
            { kind: 'disabled', initial: false, requested: false },
            { kind: 'disconnected', initial: false, requested: false },
            { kind: 'new', initial: true, requested: false },
            { kind: 'disabled', initial: false, requested: true },
            { kind: 'active', initial: true, requested: false },
        ] as const) {
            test(`${scenario.kind} connection sends visible activation choice ${scenario.requested} through callback`, async ({ page }) => {
                const pageErrors: string[] = [];
                page.on('pageerror', error => pageErrors.push(error.message));
                const bundle = await buildTrailModalFixture(dev, activationSource(scenario.kind));
                const callbackBundle = await buildTrailModalFixture(dev, callbackSource);
                await mountTrailModalFixture(page, bundle);
                const starts: any[] = [];
                const callbacks: unknown[] = [];
                const revokes: unknown[] = [];
                // This fixture models the unchanged host's flow intent: false preserves
                // enabled, true activates, and repair revoke disables before callback.
                // The real modal/store/callback page run; host CAS is covered by Go tests.
                const stored = { id: 'abcdefghijklmn1', plugin_id: 'oauth', enabled: scenario.kind === 'repair' || scenario.kind === 'active',
                    auth_context: 'oauth', connection_generation: scenario.kind === 'new' ? 'created-generation' : 'read-generation',
                    auth: { clientId: 'client', clientSecret: '' } as Record<string, string> };
                await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                    revokes.push(route.request().postDataJSON());
                    stored.enabled = false;
                    stored.connection_generation = 'revoked-generation';
                    await page.evaluate(() => (window as any).fixtureRevokeApplied());
                    await route.fulfill({ json: { ok: true, connection_generation: stored.connection_generation, updated: '2026-10-08 10:01:00.000Z' } });
                });
                await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                    starts.push(route.request().postDataJSON());
                    await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'activation-flow', instanceId: stored.id } });
                });
                await page.route('**/api/v1/plugin-system/oauth/callback', async route => {
                    callbacks.push(route.request().postDataJSON());
                    stored.enabled = stored.enabled || starts[0].activate_on_success;
                    stored.auth = { ...stored.auth, accessToken: '', refreshToken: '', oauthContext: 'oauth' };
                    await route.fulfill({ json: { ok: true } });
                });
                const loaded: typeof stored[] = [];
                await page.route(url => url.pathname === '/api/v1/plugin-instance', async route => {
                    loaded.push(JSON.parse(JSON.stringify(stored)));
                    await route.fulfill({ json: { items: [stored] } });
                });
                await page.getByTestId('open').click();
                expect(pageErrors).toEqual([]);
                const dialog = page.locator('dialog[open]');
                const checkbox = dialog.locator('input[name="activate-after-connect"]');
                await expect(checkbox).toBeVisible();
                await expect(checkbox.locator('..')).toHaveText('plugin-oauth-activate-after-connect');
                if (scenario.initial) await expect(checkbox).toBeChecked();
                else await expect(checkbox).not.toBeChecked();
                if (scenario.kind === 'new') {
                    await dialog.locator('input[name="clientId"]').fill('client');
                    await dialog.locator('input[name="clientSecret"]').fill('client-secret');
                }
                await checkbox.setChecked(scenario.requested);
                await dialog.locator('button').filter({ hasText: /^save-and-(re)?connect$/ }).click();
                await expect.poll(() => starts.length).toBe(1);
                expect(starts[0].activate_on_success).toBe(scenario.requested);
                expect(starts[0].expected_connection_generation).toBe(scenario.kind === 'repair' ? 'revoked-generation' : stored.connection_generation);
                expect(revokes.length).toBe(scenario.kind === 'repair' ? 1 : 0);
                if (scenario.kind === 'repair') expect(stored.enabled).toBe(false);
                await expect.poll(() => page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_activation-flow'))).toBe(stored.id);
                await page.goto('http://trail-modals.test/?code=callback-code&state=activation-flow');
                await page.addScriptTag({ content: callbackBundle });
                await expect.poll(() => page.evaluate(() => (window as any).fixtureNavigation)).toBe('/settings/plugins');
                expect(callbacks).toEqual([{ instanceId: stored.id, code: 'callback-code', state: 'activation-flow' }]);
                expect(loaded.at(-1)?.enabled).toBe(scenario.kind === 'active' || scenario.requested);
                expect(loaded.at(-1)?.auth).toMatchObject({ accessToken: '', refreshToken: '', oauthContext: 'oauth' });
                expect(await page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_activation-flow'))).toBeNull();
                expect(pageErrors).toEqual([]);
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
                await expect(page.locator('dialog[open]')).toBeVisible();
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
            await expect(page.locator('dialog[open]')).toBeVisible();
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
                await route.fulfill({ json: { ok: true, connection_generation: 'own-revoke-generation', updated: '2026-10-08 10:01:00.000Z' } });
            });
            let reads = 0;
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                if (reads++ === 0) {
                    await route.fulfill({ json: await page.evaluate(() => (window as any).fixturePluginInstance()) });
                    return;
                }
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
