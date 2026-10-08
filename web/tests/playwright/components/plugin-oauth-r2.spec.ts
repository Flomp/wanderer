import { expect, test } from '@playwright/test';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

const initial = {
    id: 'abcdefghijklmn1', plugin_id: 'mixed', user: 'tester', enabled: true, status: 'configured',
    auth_context: 'oauthA', connection_generation: 'opening-generation', updated: '2026-10-08 10:00:00.000Z',
    auth: { clientId: 'client', clientSecret: '', accessToken: '', refreshToken: '', oauthContext: 'oauthA' },
};

function source(record: unknown = initial) {
    return `<script>
        import Settings from '$lib/components/settings/plugins/plugin_instance_settings_modal.svelte';
        import { plugin_instances_update } from '$lib/stores/plugin_instance_store';
        let modal;
        const oauth = { type: 'oauth2', fields: ['clientId', 'clientSecret'], secretFields: ['clientSecret'] };
        const plugin = { id: 'mixed', name: 'Mixed', type: 'trails', status: 'available', auth: oauth,
            authContexts: { oauthA: oauth, oauthB: oauth, key: { type: 'api_key', secretField: 'accessToken' } },
            hostConfig: { merge: { available: false } } };
        let instance = $state(${JSON.stringify(record)});
        window.fixturePluginInstance = () => instance;
        async function save(candidate) {
            window.modalCalls.push(JSON.parse(JSON.stringify(candidate)));
            instance = await plugin_instances_update({ ...candidate, id: instance.id });
            return instance;
        }
    </script>
    <button data-testid="open" onclick={() => modal.openModal()}>Settings</button>
    <Settings {plugin} {instance} onsave={save} bind:this={modal} />`;
}

for (const dev of [true, false]) {
    test.describe(`OAuth method and repair intent (${dev ? 'development' : 'production'})`, () => {
        for (const provenance of ['unattributed', 'attributed', 'blank-selection']) test(`OAuth-only manifest update handles ${provenance} material`, async ({ page }) => {
            const legacy = { ...initial, auth_context: provenance === 'blank-selection' ? '' : 'removed-api-context',
                auth: { clientId: 'client', accessToken: '', ...(provenance === 'attributed' ? { oauthContext: 'oauthA' } : {}) } };
            const legacySource = source(legacy).replace("authContexts: { oauthA: oauth, oauthB: oauth, key: { type: 'api_key', secretField: 'accessToken' } }", "authContexts: { oauthA: oauth }");
            const bundle = await buildTrailModalFixture(dev, legacySource);
            await mountTrailModalFixture(page, bundle);
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('oauthA');
            await expect(dialog).toContainText(provenance === 'unattributed' ? 'plugin-oauth-needs-connect-hint' : 'plugin-oauth-material-present-hint');
            await expect(dialog.locator('input[name="clientSecret"]')).toHaveAttribute('type', 'password');
            await expect(dialog.locator('input[name="clientSecret"]')).toHaveValue('');
        });
        test('blank sole OAuth selection repair starts with its own credential-save generation', async ({ page }) => {
            const legacy = { ...initial, auth_context: '' };
            const soleSource = source(legacy).replace("authContexts: { oauthA: oauth, oauthB: oauth, key: { type: 'api_key', secretField: 'accessToken' } }", "authContexts: { oauthA: oauth }");
            const bundle = await buildTrailModalFixture(dev, soleSource);
            await mountTrailModalFixture(page, bundle);
            let stored: any = structuredClone(legacy);
            const patches: any[] = [], starts: any[] = [];
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                if (route.request().method() === 'GET') return route.fulfill({ json: stored });
                const patch = route.request().postDataJSON();
                patches.push(patch);
                expect(patch).toEqual({ authContext: 'oauthA', auth: { clientSecret: 'repair-client-secret' },
                    expected_connection_generation: 'opening-generation' });
                stored = { ...stored, auth_context: 'oauthA', enabled: false, status: 'disabled',
                    auth: { clientId: 'client', clientSecret: '' }, connection_generation: 'own-credential-save-generation',
                    updated: '2026-10-08 10:01:00.000Z' };
                await route.fulfill({ json: stored });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                const start = route.request().postDataJSON();
                starts.push(start);
                if (start.expected_connection_generation !== stored.connection_generation) return route.fulfill({ status: 409,
                    json: { status: 409, message: 'Connection changed', data: { operation: { code: 'connection_conflict' } } } });
                stored.updated = '2026-10-08 10:02:00.000Z';
                await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'sole-repair-flow',
                    instanceId: stored.id, connection_generation: stored.connection_generation, updated: stored.updated } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('oauthA');
            await dialog.locator('input[name="clientSecret"]').fill('repair-client-secret');
            await dialog.locator('button').filter({ hasText: 'save-and-reconnect' }).click();
            await expect.poll(() => page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_sole-repair-flow'))).toBe(stored.id);
            expect(patches).toHaveLength(1);
            expect(starts).toHaveLength(1);
            expect(starts[0]).toMatchObject({ authContext: 'oauthA', expected_connection_generation: 'own-credential-save-generation', activate_on_success: true });
        });
        for (const grantSource of ['active-oauth', 'removed-provenance', 'removed-refresh']) test(`API-key preview hides ${grantSource} grant material and always uses a password input`, async ({ page }) => {
            const stored = grantSource === 'active-oauth' ? initial : { ...initial, auth_context: 'removed-oauth',
                auth: { clientId: 'client', accessToken: '', ...(grantSource === 'removed-provenance' ? { oauthContext: 'removed-oauth' } : { refreshToken: '' }) } };
            const bundle = await buildTrailModalFixture(dev, source(stored));
            await mountTrailModalFixture(page, bundle);
            const patches: any[] = [];
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                if (route.request().method() === 'GET') return route.fulfill({ json: stored });
                patches.push(route.request().postDataJSON());
                await route.fulfill({ json: { ...initial, enabled: false, auth_context: 'key',
                    connection_generation: 'key-generation', auth: { clientId: 'client', clientSecret: '', accessToken: '' } } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('key');
            const key = dialog.locator('input[name="accessToken"]');
            await expect(dialog).toContainText('plugin-auth-change-removes-oauth');
            await expect(key).toHaveValue('');
            await expect(key).toHaveAttribute('type', 'password');
            await expect(key).toHaveAttribute('placeholder', '');
            await expect(dialog.locator('input[name="clear-accessToken"]')).toHaveCount(0);
            await key.fill('new-key-secret');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('oauthB');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('key');
            await expect(key).toHaveAttribute('type', 'password');
            await expect(key).toHaveValue('new-key-secret');
            await dialog.locator('button[name="save"]').click();
            expect(patches).toEqual([{ auth: { accessToken: 'new-key-secret' }, authContext: 'key',
                expected_connection_generation: 'opening-generation' }]);
        });

        test('non-OAuth method switch preserves a shared saved API key', async ({ page }) => {
            const api = { ...initial, auth_context: 'keyA', auth: { accessToken: '' } };
            const apiSource = source(api).replace("authContexts: { oauthA: oauth, oauthB: oauth, key: { type: 'api_key', secretField: 'accessToken' } }",
                "authContexts: { keyA: { type: 'api_key', secretField: 'accessToken' }, keyB: { type: 'api_key', secretField: 'accessToken' } }");
            const bundle = await buildTrailModalFixture(dev, apiSource);
            await mountTrailModalFixture(page, bundle);
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('keyB');
            await expect(dialog).not.toContainText('plugin-auth-change-removes-oauth');
            await expect(dialog.locator('input[name="accessToken"]')).toHaveAttribute('placeholder', '(unchanged)');
            await expect(dialog.locator('input[name="accessToken"]')).toHaveAttribute('type', 'password');
            await expect(dialog.locator('input[name="clear-accessToken"]')).toHaveCount(1);
        });

        for (const inactiveInput of [false, true]) test(`another OAuth context is only a flow proposal with inactive input ${inactiveInput}`, async ({ page }) => {
            const mixedSource = source().replace('oauthB: oauth,', "oauthB: { ...oauth, fields: ['clientId', 'clientSecret', 'tenant'] },");
            const bundle = await buildTrailModalFixture(dev, mixedSource);
            await mountTrailModalFixture(page, bundle);
            const stored = structuredClone(initial);
            const patches: any[] = [], starts: any[] = [];
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                if (route.request().method() === 'GET') return route.fulfill({ json: stored });
                patches.push(route.request().postDataJSON());
                expect(patches.at(-1).authContext).toBeUndefined();
                if (inactiveInput) Object.assign(stored.auth, patches.at(-1).auth);
                await route.fulfill({ json: stored });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                starts.push(route.request().postDataJSON());
                expect(stored.auth_context).toBe('oauthA');
                expect(stored.auth.oauthContext).toBe('oauthA');
                await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'other-context-flow',
                    instanceId: stored.id, connection_generation: stored.connection_generation, updated: '2026-10-08 10:01:00.000Z' } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('oauthB');
            if (inactiveInput) await dialog.locator('input[name="tenant"]').fill('inactive-draft');
            await expect(dialog).toContainText('plugin-oauth-needs-connect-hint');
            await dialog.locator('button').filter({ hasText: 'save-and-connect' }).click();
            await expect.poll(() => starts.length).toBe(1);
            expect(patches).toEqual([{ expected_connection_generation: 'opening-generation',
                ...(inactiveInput ? { auth: { tenant: 'inactive-draft' } } : {}) }]);
            expect(starts[0]).toMatchObject({ authContext: 'oauthB', expected_connection_generation: 'opening-generation', activate_on_success: true });
            expect(stored).toEqual({ ...initial, auth: { ...initial.auth, ...(inactiveInput ? { tenant: 'inactive-draft' } : {}) } });
            await dialog.locator('button').filter({ hasText: 'cancel' }).click();
            await page.getByTestId('open').click();
            await expect(dialog.locator('select[name="pluginAuthContext"]')).toHaveValue('oauthA');
            await expect(dialog).toContainText('plugin-oauth-material-present-hint');
            expect(stored.enabled).toBe(true);
        });

        test('shared active client input must be explicitly saved before starting another context', async ({ page }) => {
            const bundle = await buildTrailModalFixture(dev, source());
            await mountTrailModalFixture(page, bundle);
            let starts = 0;
            await page.route('**/api/v1/plugin-system/oauth/start', async route => { starts++; await route.fulfill({ json: {} }); });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('select[name="pluginAuthContext"]').selectOption('oauthB');
            await dialog.locator('input[name="clientId"]').fill('another-client');
            await dialog.locator('button').filter({ hasText: 'save-and-connect' }).click();
            expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([]);
            expect(starts).toBe(0);
            await expect(dialog).toHaveCount(1);
        });

        for (const external of ['none', 'disable', 'revoke']) {
            test(`repair choice survives own credential auto-disable and respects external ${external}`, async ({ page }) => {
                const bundle = await buildTrailModalFixture(dev, source());
                await mountTrailModalFixture(page, bundle);
                let stored: any = structuredClone(initial);
                await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                    if (route.request().method() === 'GET') return route.fulfill({ json: stored });
                    expect(route.request().postDataJSON()).toMatchObject({ auth: { clientSecret: 'replacement-client-secret' },
                        expected_connection_generation: 'opening-generation' });
                    stored = { ...stored, enabled: false, status: 'disabled', connection_generation: 'own-repair-generation',
                        updated: '2026-10-08 10:01:00.000Z', auth: { clientId: 'client', clientSecret: '' } };
                    await route.fulfill({ json: stored });
                });
                await page.getByTestId('open').click();
                const dialog = page.locator('dialog[open]');
                await expect(dialog.locator('input[name="activate-after-connect"]')).toBeChecked();
                await dialog.locator('input[name="clientSecret"]').fill('replacement-client-secret');
                await dialog.locator('button[name="save"]').click();
                await expect(dialog).toHaveCount(0);
                if (external !== 'none') stored = { ...stored, updated: '2026-10-08 10:02:00.000Z',
                    connection_generation: external === 'revoke' ? 'external-revoke-generation' : stored.connection_generation };
                await page.getByTestId('open').click();
                if (external === 'none') await expect(dialog.locator('input[name="activate-after-connect"]')).toBeChecked();
                else await expect(dialog.locator('input[name="activate-after-connect"]')).not.toBeChecked();
            });
        }

        test('opening reads a fresh deliberate disable before choosing the checkbox default', async ({ page }) => {
            const bundle = await buildTrailModalFixture(dev, source());
            await mountTrailModalFixture(page, bundle);
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => {
                await route.fulfill({ json: { ...initial, enabled: false, status: 'disabled', updated: '2026-10-08 10:03:00.000Z' } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await expect(dialog.locator('input[name="activate-after-connect"]')).not.toBeChecked();
            await expect(dialog).not.toContainText('plugin-error-connection-conflict');
        });

        test('revoke repair choice survives a failed start and is cleared by an explicit disconnect', async ({ page }) => {
            const unreadable = { ...initial, auth: { ...initial.auth }, last_error: { code: 'credentials_unreadable' } };
            const bundle = await buildTrailModalFixture(dev, source(unreadable));
            await mountTrailModalFixture(page, bundle);
            let stored: any = structuredClone(unreadable);
            let revokes = 0;
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => route.fulfill({ json: stored }));
            await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                revokes++;
                stored = { ...stored, enabled: false, status: 'disabled', connection_generation: 'own-revoke-generation',
                    updated: '2026-10-08 10:01:00.000Z', auth: { clientId: 'client', clientSecret: '' }, last_error: undefined };
                await route.fulfill({ json: { ok: true, connection_generation: stored.connection_generation, updated: stored.updated } });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                expect(route.request().postDataJSON().activate_on_success).toBe(true);
                await route.fulfill({ status: 400, json: { status: 400, message: 'Unavailable', data: {} } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('button').filter({ hasText: 'save-and-reconnect' }).click();
            await expect.poll(() => revokes).toBe(1);
            await expect(dialog.locator('button').filter({ hasText: 'cancel' })).toBeEnabled();
            await dialog.locator('button').filter({ hasText: 'cancel' }).click();
            await page.getByTestId('open').click();
            await expect(dialog.locator('input[name="activate-after-connect"]')).toBeChecked();
            // A fresh saved grant makes Disconnect available, but its explicit clear
            // must not retain the earlier automatic-repair activation request.
            stored = { ...stored, auth: { ...stored.auth, accessToken: '', oauthContext: 'oauthA' } };
            await dialog.locator('button').filter({ hasText: 'cancel' }).click();
            await page.getByTestId('open').click();
            await dialog.locator('button').filter({ hasText: 'plugin-disconnect' }).click();
            await expect(dialog).toHaveCount(0);
            await page.getByTestId('open').click();
            await expect(dialog.locator('input[name="activate-after-connect"]')).not.toBeChecked();
        });

        test('repair choice survives its own pending flow when authorization is abandoned', async ({ page }) => {
            const unreadable = { ...initial, auth: { ...initial.auth }, last_error: { code: 'credentials_unreadable' } };
            const bundle = await buildTrailModalFixture(dev, source(unreadable));
            await mountTrailModalFixture(page, bundle);
            let stored: any = structuredClone(unreadable);
            await page.route('**/api/v1/plugin-instance/abcdefghijklmn1', async route => route.fulfill({ json: stored }));
            await page.route('**/api/v1/plugin-system/oauth/revoke', async route => {
                stored = { ...stored, enabled: false, status: 'disabled', connection_generation: 'own-revoke-generation',
                    updated: '2026-10-08 10:01:00.000Z', auth: { clientId: 'client', clientSecret: '' }, last_error: undefined };
                await route.fulfill({ json: { ok: true, connection_generation: stored.connection_generation, updated: stored.updated } });
            });
            await page.route('**/api/v1/plugin-system/oauth/start', async route => {
                expect(route.request().postDataJSON()).toMatchObject({ expected_connection_generation: 'own-revoke-generation', activate_on_success: true });
                stored.updated = '2026-10-08 10:02:00.000Z';
                await route.fulfill({ json: { url: 'http://trail-modals.test/#authorization', state: 'abandoned-flow', instanceId: stored.id,
                    connection_generation: stored.connection_generation, updated: stored.updated } });
            });
            await page.getByTestId('open').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('button').filter({ hasText: 'save-and-reconnect' }).click();
            await expect.poll(() => page.evaluate(() => sessionStorage.getItem('wanderer_plugin_oauth_abandoned-flow'))).toBe(stored.id);
            await dialog.locator('button').filter({ hasText: 'cancel' }).click();
            await page.getByTestId('open').click();
            await expect(dialog.locator('input[name="activate-after-connect"]')).toBeChecked();
            stored.updated = '2026-10-08 10:03:00.000Z'; // Explicit false->false from another client.
            await dialog.locator('button').filter({ hasText: 'cancel' }).click();
            await page.getByTestId('open').click();
            await expect(dialog.locator('input[name="activate-after-connect"]')).not.toBeChecked();
        });
    });
}
