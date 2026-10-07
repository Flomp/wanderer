import { expect, test } from '@playwright/test';
import { buildTrailModalFixture, mountTrailModalFixture } from './trail-modal-fixture';

const source = `<script>
    import Send from '$lib/components/trail/trail_send_modal.svelte';
    let modal;
    const trail = { id: 'trail', name: 'Trail' };
    </script>
    <button data-testid="open-send" onclick={() => modal.openModal()}>Send</button>
    <Send {trail} bind:this={modal} />`;

for (const dev of [true, false]) {
    test.describe(`plugin send conflicts (${dev ? 'development' : 'production'})`, () => {
        let bundle: string;
        test.beforeAll(async () => { bundle = await buildTrailModalFixture(dev, source, {
            '\0fixture:toast': 'export function show_toast(toast) { window.modalCalls.push(toast); }',
        }); });

        for (const [code, text] of [
            ['plugin_instance_busy', 'plugin-error-busy'],
            ['plugin_connection_conflict', 'plugin-error-connection-conflict'],
            ['plugin_instance_disabled', 'plugin-error-instance-disabled'],
        ]) {
            test(code, async ({ page }) => {
                await mountTrailModalFixture(page, bundle);
                let sends = 0;
                let reads = 0;
                await page.route('**/api/v1/plugin-instance?perPage=-1', async route => {
                    reads++;
                    await route.fulfill({ json: { items: [{ plugin_id: 'device', enabled: sends === 0 || code === 'plugin_instance_busy' }] } });
                });
                await page.route('**/api/v1/plugin-system/trail-send', async route => {
                    sends++;
                    await route.fulfill({ status: 409, json: { status: 409, message: 'safe',
                        data: { operation: { code, message: 'provider-private-sentinel' } } } });
                });
                await page.getByTestId('open-send').click();
                const dialog = page.locator('dialog[open]');
                await dialog.locator('button[aria-label="Test Device"]').click();
                await expect.poll(() => page.evaluate(() => (window as any).modalCalls[0]?.text)).toBe(text);
                expect(sends).toBe(1);
                expect(reads).toBeGreaterThan(1);
                await expect(dialog).not.toContainText('provider-private-sentinel');
                if (code !== 'plugin_instance_busy') await expect(dialog).toContainText('no-send-plugins');
            });
        }

        test('refreshes temporary busy after the lock is released', async ({ page }) => {
            await page.clock.install();
            await mountTrailModalFixture(page, bundle);
            let busy = true;
            await page.route('**/api/v1/plugin-instance?perPage=-1', async route => {
                await route.fulfill({ json: { items: [{ plugin_id: 'device', enabled: true, busy }] } });
            });
            await page.getByTestId('open-send').click();
            const dialog = page.locator('dialog[open]');
            await expect(dialog).toContainText('plugin-operation-running');
            busy = false;
            await page.clock.fastForward(3000);
            await expect(dialog).not.toContainText('plugin-operation-running');
            await expect(dialog.locator('button[aria-label="Test Device"]')).toBeEnabled();
        });
    });
}
