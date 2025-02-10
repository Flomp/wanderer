import { test as teardown } from '@playwright/test';

teardown('delete user', async ({ page }) => {
    await page.goto('/settings/account', { waitUntil: 'networkidle' });
    await page.locator("#delete-account").click();
    await page.locator("#confirm").click();

    await page.waitForURL('/');
});