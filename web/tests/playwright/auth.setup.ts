import { test as setup } from '@playwright/test';

const authFile = 'playwright/.auth/user.json';

setup('create user', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'networkidle' });
    await page.locator('input[name="username"]').fill('Test');
    await page.locator('input[name="password"]').fill('password');
    const loginPromise = page.waitForResponse('**/api/v1/auth/login')
    await page.getByRole('button', { name: 'Login' }).click();
    const response = await loginPromise;

    let responseJson;
    try {
        responseJson = await response.json();

    } catch (e) {
        if (e instanceof Error && e.message.includes("No resource with given identifier found")) {
            console.log("Already logged in!")
        } else {
            throw e
        }
    }

    if (responseJson?.message === "Failed to authenticate.") {
        await page.goto('/register', { waitUntil: 'networkidle' });
        await page.locator('input[name="username"]').fill('Test');
        await page.locator('input[name="email"]').fill('test@test.de');
        await page.locator('input[name="password"]').fill('password');
        await page.locator('#submit').click();
    }



    // Wait until the page receives the cookies.
    //
    // Sometimes login flow sets cookies in the process of several redirects.
    // Wait for the final URL to ensure that the cookies are actually set.
    await page.waitForURL('/');

    // End of authentication steps.

    await page.context().storageState({ path: authFile });
});