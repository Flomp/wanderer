import { expect, test } from '@playwright/test';
import { IndexPage } from '../../pages/index_page';

test('index page does not show error', async ({ page }) => {
	const indexPage = new IndexPage(page);
	await indexPage.goto()
	await indexPage.hasNoError()
});
