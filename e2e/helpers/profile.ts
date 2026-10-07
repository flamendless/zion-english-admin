import { expect, Page } from "@playwright/test";

export async function gotoProfile(page: Page) {
	await page.goto("profile");
	await expect(page.getByRole("heading", { name: "My Profile", level: 1 })).toBeVisible();
}

export async function updateProfileMiddleName(page: Page, value: string) {
	await gotoProfile(page);
	await page.getByRole("tab", { name: "Security" }).click();
	await page.locator("#middleName").fill(value);
	await page.getByRole("button", { name: "Save name" }).click();
	await expect(page.locator("#successBanner")).toContainText(/success/i, { timeout: 15_000 });
	await expect(page.locator(".profile-identity h2")).toContainText(value);
}
