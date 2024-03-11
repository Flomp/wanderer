import { test, expect } from '@playwright/test';

test('logs the user out', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: 'avatar' }).click();
    await page.locator(".menu .menu-item").filter({ hasText: "Logout" }).click();

    const cookies = await page.context().cookies();
    const pbAuthCookie = cookies.find(cookie => cookie.name === 'pb_auth');
    expect(pbAuthCookie).toBeFalsy();
});
