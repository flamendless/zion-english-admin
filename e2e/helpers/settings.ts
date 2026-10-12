import { expect, Page } from "@playwright/test";

export async function revealSettingsSecret(page: Page, superuserPassword: string) {
	const secretRow = page.locator('[data-field-key="secret"]');
	await secretRow.click();
	await expect(page.locator("#settingsRevealModal")).toBeVisible();
	await page.locator("#settingsRevealPassword").fill(superuserPassword);
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/settings/reveal") && r.request().method() === "POST" && r.ok(),
		),
		page.locator("#settingsRevealForm").getByRole("button", { name: "View value" }).click(),
	]);
}

export async function expectRevealedSecretValue(page: Page, expected: string) {
	await expect(page.locator("#settings-value-secret")).toContainText(expected);
}
