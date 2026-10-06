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
            auth: { email: 'old@example.org', password: '', oauthContext: 'session' },
            config: { host: { connectors: { service: { baseURL: 'https://saved.example' } } },
                plugin: { selector: 'saved' } },
        };
        function save(candidate) { window.modalCalls.push(JSON.parse(JSON.stringify(candidate))); }
    </script>
    <button data-testid="open-settings" onclick={() => settings.openModal()}>Settings</button>
    <Settings {plugin} {instance} onsave={save} bind:this={settings} />
`;

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
        await dialog.locator('button[name="save"]').click();
        await expect.poll(() => validation).toBeDefined();
        expect(validation!.authContext).toBe('session');
        expect(validation!.config.host.connectors.service).toEqual({
            baseURL: 'https://saved.example', allowPrivate: true,
            tls: { mode: 'customCA', caBundle: 'installed-ca' },
        });
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
