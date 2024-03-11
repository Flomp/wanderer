import { test as base, expect } from '@playwright/test';
import { ListsPage } from '../../pages/lists_page';

const test = base.extend<{ listsPage: ListsPage }>({
    listsPage: async ({ page }, use) => {
        const listsPage = new ListsPage(page);
        await listsPage.goto();
        await listsPage.create();
        await use(listsPage);
        await listsPage.removeAll();
    },
});

test('shows a list card', async ({ listsPage }) => {
    const listItemCount = await listsPage.listItems.count();
    await listsPage.create();
    expect(listsPage.listItems).toHaveCount(listItemCount + 1);
    expect(listsPage.listItemsImage).toBeVisible();
});

test('update a list card', async ({ listsPage }) => {
    await listsPage.update();
    expect(listsPage.listItems.first()).toContainText("Updated List");
    expect(listsPage.listItems.first()).toContainText("New Description");
});

test('delete a list card', async ({ listsPage }) => {
    await listsPage.delete();
    expect(listsPage.listItems).toHaveCount(0);
});
