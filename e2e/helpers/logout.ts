import { expect, Page } from "@playwright/test";

export async function logout(page: Page) {
	page.once("dialog", (dialog) => dialog.accept());
	await page.getByRole("button", { name: "Logout" }).click();
	await page.waitForURL("**/auth/login**");
	await expect(page.getByRole("heading", { name: "Login" })).toBeVisible();
}
