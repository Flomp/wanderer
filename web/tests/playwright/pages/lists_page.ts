import { type Locator, type Page } from '@playwright/test';

export class ListsPage {
  readonly page: Page;
  readonly createListButton: Locator;

  readonly listItems: Locator;
  readonly listItemsImage: Locator;

  readonly listModal: Locator;
  readonly listModalAvatar: Locator;
  readonly listModalName: Locator;
  readonly listModalDescription: Locator;
  readonly listModalSaveButton: Locator;


  readonly confirmModal: Locator;
  readonly confirmModalConfirmButton: Locator;


  constructor(page: Page) {
    this.page = page;
    this.createListButton = page.locator("#create-list-button");
    this.listModal = page.locator("#list-modal");

    this.listModalAvatar = this.listModal.locator('input[name="avatar"]')
    this.listModalName = this.listModal.locator('input[name="name"]')
    this.listModalDescription = this.listModal.locator('textarea[name="description"]')
    this.listModalSaveButton = this.listModal.getByText('Save');

    this.listItems = page.locator('.list-list-item');
    this.listItemsImage = page.locator('.list-list-item img');


    this.confirmModal = page.locator("#confirm-modal");
    this.confirmModalConfirmButton = this.confirmModal.locator("button").filter({ hasText: "Delete" });
  }

  async goto() {
    await this.page.goto('/lists', { waitUntil: 'networkidle' });
  }

  async create(name: string = "Test List") {
    await this.createListButton.click();
    await this.listModalName.fill(name);
    await this.listModalAvatar.setInputFiles([
      "./tests/playwright/fixtures/avatar.webp"
    ]);
    await Promise.all([
      this.page.waitForResponse(resp => resp.url().includes('/api/v1/list') && resp.status() === 200),
      this.listModalSaveButton.click()
    ]);
  }

  async update(name: string = "Updated List", description = "New Description") {
    await this.listItems.first().locator(".dropdown button").click();
    await this.listItems.first().locator(".menu .menu-item").filter({ hasText: "Edit" }).click();
    await this.listModalName.fill(name);
    await this.listModalDescription.fill(description);

    await Promise.all([
      this.page.waitForResponse(resp => resp.url().includes('/api/v1/list') && resp.status() === 200),
      this.listModalSaveButton.click()
    ]);
  }

  async delete() {
    await this.listItems.first().locator(".dropdown button").click();
    await this.listItems.first().locator(".menu .menu-item").filter({ hasText: "Delete" }).click();
    await Promise.all([
      this.page.waitForResponse(resp => resp.url().includes('/api/v1/list') && resp.status() === 200),
      this.confirmModalConfirmButton.click()
    ]);
  }

  async removeAll() {
    while ((await this.listItems.count()) > 0) {
      await this.delete();
    }
  }


}