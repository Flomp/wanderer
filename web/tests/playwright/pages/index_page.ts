import { expect, type Locator, type Page } from '@playwright/test';

export class IndexPage {
  readonly page: Page;
  readonly error: Locator;

  constructor(page: Page) {
    this.page = page;
    this.error = page.getByText("Internal Error");
  }

  async goto() {
    await this.page.goto('/');
  }

  async search() {
    await this.page.locator('input[name="q"]').fill('Munich');
    await this.page.waitForResponse('**/api/v1/search/multi');
    await this.page.locator('.menu-item').first().click();
    await this.page.waitForURL('/map?lat=48.13743&lon=11.57549');

  }


  async hasNoError() {
    await expect(this.error).toHaveCount(0);
  }
}