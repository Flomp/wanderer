import { expect, test } from '@playwright/test';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

const fixtureSource = `
    <script>
        import Settings from '$lib/components/settings/plugins/plugin_instance_settings_modal.svelte';
        let settings;
        const context = { type: 'session', fields: ['email', 'password'], secretFields: ['password'] };
        const plugin = {
            id: 'configured-session', type: 'trails', name: 'Configured session', status: 'available',
            auth: context, authContext: 'session', authContexts: { session: context },
            configSchema: [{ key: 'selector', type: 'text', label: 'Selector', default: 'default' }],
            hostConfig: { connectors: { service: { baseURL: 'https://installed.example', allowPrivate: true,
                tls: { mode: 'customCA', caBundle: 'installed-ca' } } }, merge: { available: false } },
        };
        const instance = {
            id: 'abcdefghijklmn1', plugin_id: plugin.id, user: 'tester', enabled: true, status: 'configured',
            auth_context: 'session', connection_generation: 'revision',
            auth: { email: 'old@example.org', password: '' },
            config: { host: { connectors: { service: { baseURL: 'https://saved.example' } } },
                plugin: { selector: 'saved' } },
        };
        window.fixturePluginInstance = () => instance;
        function save(candidate) { window.modalCalls.push(JSON.parse(JSON.stringify(candidate))); }
    </script>
    <button data-testid="open-settings" onclick={() => settings.openModal()}>Settings</button>
    <Settings {plugin} {instance} onsave={save} bind:this={settings} />
`;

const liveFixtureSource = fixtureSource.replace('const instance = {', 'let instance = $state({')
    .replace('        };\n        window.fixturePluginInstance', `        });
        window.advanceConnection = () => { instance = { ...instance, connection_generation: 'new-generation',
            auth: { email: 'new-connection@example.org', password: '' } }; };
        window.fixturePluginInstance`);

for (const dev of [true, false]) {
    test(`saves the exact draft validated before a pending provider call (${dev ? 'development' : 'production'})`, async ({ page }) => {
        const errors: string[] = [];
        page.on('pageerror', error => errors.push(error.message));
        const bundle = await buildTrailModalFixture(dev, fixtureSource);
        await mountTrailModalFixture(page, bundle);
        let validation: Record<string, any> | undefined;
        let complete!: () => void;
        const pending = new Promise<void>(resolve => { complete = resolve; });
        await page.route('**/api/v1/plugin-system/auth/validate', async route => {
            validation = route.request().postDataJSON();
            await pending;
            await route.fulfill({ json: { ok: true, authContext: 'session' } });
        });
        await page.getByTestId('open-settings').click();
        const dialog = page.locator('dialog[open]');
        await dialog.locator('input[name="email"]').fill('new@example.org');
        await dialog.locator('input[name="password"]').fill('replacement-secret');
        await dialog.locator('input[name="selector"]').fill('validated-selector');
        await dialog.locator('button[name="save-validate"]').click();
        await expect.poll(() => validation).toBeDefined();
        expect(validation!.authContext).toBe('session');
        expect(validation!.expected_connection_generation).toBe('revision');
        // Unchanged connector defaults and stored overrides stay on the host.
        expect(validation!.config.host).toBeUndefined();
        expect(validation!.config.plugin.selector).toBe('validated-selector');
        // Inputs remain editable. Their later values must not replace the validated snapshot.
        await dialog.locator('input[name="selector"]').fill('unvalidated-selector');
        await dialog.locator('input[name="password"]').fill('unvalidated-password');
        complete();
        await expect.poll(() => page.evaluate(() => (window as any).modalCalls.length)).toBe(1);
        const saved = await page.evaluate(() => (window as any).modalCalls[0]);
        expect(saved.auth).toEqual(validation!.auth);
        expect(saved.config).toEqual(validation!.config);
        await expect(page.locator('dialog[open]')).toHaveCount(0);
        expect(errors).toEqual([]);
    });
}

for (const dev of [true, false]) {
    for (const action of ['save', 'save-validate']) {
        test(`open draft keeps its original generation for ${action} after a parent refresh (${dev ? 'development' : 'production'})`, async ({ page }) => {
            const bundle = await buildTrailModalFixture(dev, liveFixtureSource);
            await mountTrailModalFixture(page, bundle);
            const validations: any[] = [];
            await page.route('**/api/v1/plugin-system/auth/validate', async route => {
                validations.push(route.request().postDataJSON());
                await route.fulfill({ status: 409, json: { status: 409, message: 'safe',
                    data: { operation: { code: 'plugin_connection_conflict', message: 'safe' } } } });
            });
            await page.getByTestId('open-settings').click();
            const dialog = page.locator('dialog[open]');
            await dialog.locator('input[name="password"]').fill('replacement-password');
            await page.evaluate(() => (window as any).advanceConnection());
            await expect(dialog).toContainText('plugin-error-connection-conflict');
            await dialog.locator(`button[name="${action}"]`).click();
            if (action === 'save') {
                const saved = await page.evaluate(() => (window as any).modalCalls[0]);
                expect(saved.auth).toEqual({ password: 'replacement-password' });
                expect(saved.expected_connection_generation).toBe('revision');
            } else {
                await expect.poll(() => validations.length).toBe(1);
                expect(validations[0].auth).toEqual({ password: 'replacement-password' });
                expect(validations[0].expected_connection_generation).toBe('revision');
                expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([]);
                await dialog.locator('button').filter({ hasText: 'plugin-reload-settings' }).click();
                await expect(dialog.locator('input[name="email"]')).toHaveValue('new-connection@example.org');
                await dialog.locator('input[name="password"]').fill('conscious-retry');
                await dialog.locator('button[name="save-validate"]').click();
                await expect.poll(() => validations.length).toBe(2);
                expect(validations[1].expected_connection_generation).toBe('new-generation');
                expect(validations[1].auth).toEqual({ password: 'conscious-retry' });
            }
        });
    }
}

for (const dev of [true, false]) {
    for (const scenario of ['no-op', 'credential repair', 'explicit clear']) {
        test(`plain Save permits ${scenario} with invalid stored config (${dev ? 'development' : 'production'})`, async ({ page }) => {
            const invalidSource = fixtureSource
                .replace("configSchema: [{ key: 'selector'", "configSchema: [{ key: 'baseURL', type: 'url', required: true, hidden: true }, { key: 'selector'")
                .replace("status: 'configured',", "status: 'disabled',")
                .replace("enabled: true", "enabled: false");
            const bundle = await buildTrailModalFixture(dev, invalidSource);
            await mountTrailModalFixture(page, bundle);
            let validations = 0;
            await page.route('**/api/v1/plugin-system/auth/validate', async route => {
                validations++;
                await route.fulfill({ json: { ok: true } });
            });
            await page.getByTestId('open-settings').click();
            const dialog = page.locator('dialog[open]');
            if (scenario === 'credential repair') await dialog.locator('input[name="password"]').fill('replacement');
            if (scenario === 'explicit clear') await dialog.locator('input[name="clear-password"]').check();
            await dialog.locator('button[name="save"]').click();
            await expect.poll(() => page.evaluate(() => (window as any).modalCalls.length)).toBe(1);
            const saved = await page.evaluate(() => (window as any).modalCalls[0]);
            expect(saved.config).toBeUndefined();
            expect(saved.enabled).toBeUndefined();
            expect(saved.status).toBeUndefined();
            expect(saved.state).toBeUndefined();
            expect(validations).toBe(0);
            if (scenario === 'credential repair') expect(saved.auth).toEqual({ password: 'replacement' });
            if (scenario === 'explicit clear') expect(saved.auth).toEqual({ password: null });
            if (scenario === 'no-op') expect(saved.auth).toBeUndefined();
        });
    }

    test(`deliberate Save and Validate checks invalid config before provider I/O (${dev ? 'development' : 'production'})`, async ({ page }) => {
        const invalidSource = fixtureSource.replace("configSchema: [{ key: 'selector'", "configSchema: [{ key: 'baseURL', type: 'url', required: true, hidden: true }, { key: 'selector'");
        const bundle = await buildTrailModalFixture(dev, invalidSource);
        await mountTrailModalFixture(page, bundle);
        let validations = 0;
        await page.route('**/api/v1/plugin-system/auth/validate', async route => {
            validations++;
            await route.fulfill({ json: { ok: true } });
        });
        await page.getByTestId('open-settings').click();
        await page.locator('dialog[open] button[name="save-validate"]').click();
        await expect(page.locator('dialog[open]')).toHaveCount(1);
        expect(await page.evaluate(() => (window as any).modalCalls.length)).toBe(0);
        expect(validations).toBe(0);
    });
}

for (const dev of [true, false]) {
    test(`context switch keeps saved inactive nonsecret credentials (${dev ? 'development' : 'production'})`, async ({ page }) => {
        const contextSource = fixtureSource
            .replace("authContexts: { session: context }", "authContexts: { oauth: { type: 'oauth2', fields: ['clientId'], secretFields: ['clientSecret'] }, session: context }")
            .replace("auth_context: 'session'", "auth_context: 'oauth'")
            .replace("auth: { email: 'old@example.org', password: '' }", "auth: { clientId: 'client', clientSecret: '', email: 'old@example.org', password: '', oauthContext: 'oauth', accessToken: '' }");
        const bundle = await buildTrailModalFixture(dev, contextSource);
        await mountTrailModalFixture(page, bundle);
        await page.getByTestId('open-settings').click();
        const dialog = page.locator('dialog[open]');
        await dialog.locator('select[name="pluginAuthContext"]').selectOption('session');
        await expect(dialog.locator('input[name="email"]')).toHaveValue('old@example.org');
        await dialog.locator('button[name="save"]').click();
        const saved = await page.evaluate(() => (window as any).modalCalls[0]);
        expect(saved.auth).toBeUndefined();
        expect(saved.authContext).toBe('session');
        expect(saved.expected_connection_generation).toBe('revision');
    });

    for (const scenario of ['secret clear', 'empty nonsecret']) {
        test(`Save and Validate forwards intentional ${scenario} instead of validating stored credentials (${dev ? 'development' : 'production'})`, async ({ page }) => {
            const bundle = await buildTrailModalFixture(dev, fixtureSource);
            await mountTrailModalFixture(page, bundle);
            let validation: Record<string, any> | undefined;
            await page.route('**/api/v1/plugin-system/auth/validate', async route => {
                validation = route.request().postDataJSON();
                await route.fulfill({ status: 400, json: { status: 400, message: 'Credentials missing', data: { operation: { code: 'auth_failed', message: 'safe' } } } });
            });
            await page.getByTestId('open-settings').click();
            const dialog = page.locator('dialog[open]');
            if (scenario === 'secret clear') await dialog.locator('input[name="clear-password"]').check();
            else await dialog.locator('input[name="email"]').fill('');
            await dialog.locator('button[name="save-validate"]').click();
            await expect.poll(() => validation).toBeDefined();
            expect(validation!.auth).toEqual(scenario === 'secret clear' ? { password: null } : { email: '' });
            expect(await page.evaluate(() => (window as any).modalCalls)).toEqual([]);
            await expect(dialog).toHaveCount(1);
        });
    }
}
