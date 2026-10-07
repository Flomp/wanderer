import { expect, test } from '@playwright/test';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

for (const dev of [true, false]) {
    test.describe(`plugin card diagnosis (${dev ? 'development' : 'production'})`, () => {
        for (const lastSyncAt of ['', '2026-10-07T12:00:00Z']) {
            test(`shows translated connection diagnosis ${lastSyncAt ? 'with' : 'without'} a sync timestamp`, async ({ page }) => {
                const source = `<script>
                    import Card from '$lib/components/settings/plugins/plugin_card.svelte';
                    import { translatePluginError } from '$lib/util/plugin_error_i18n';
                    const error = translatePluginError('credentials_unreadable', 'private-provider-sentinel');
                </script>
                <Card title="OAuth" active={false} onclick={() => {}} oninfo={() => {}} ontoggle={() => {}}
                    lastSyncAt=${JSON.stringify(lastSyncAt)} {error} />`;
                const bundle = await buildTrailModalFixture(dev, source);
                await mountTrailModalFixture(page, bundle);
                const diagnosis = page.locator('[title="plugin-error-credentials-unreadable"]');
                await expect(diagnosis).toHaveCount(1);
                if (!lastSyncAt) await expect(diagnosis).toHaveText('plugin-error-credentials-unreadable');
                await expect(page.locator('body')).not.toContainText('private-provider-sentinel');
                await expect(page.locator('body')).not.toContainText('plugin-setup-error');
            });
        }
    });
}
