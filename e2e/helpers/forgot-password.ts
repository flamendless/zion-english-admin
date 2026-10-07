import { expect, Page } from "@playwright/test";

export async function requestPasswordReset(page: Page, email: string) {
	await page.goto("auth/forgot-password");
	await page.locator("#email").fill(email);
	await page.getByRole("button", { name: "Send reset password" }).click();
	await page.waitForURL("**/auth/forgot-password/reset**", { timeout: 30_000 });
}

export async function completePasswordReset(page: Page, newPassword: string) {
	await page.locator("#password").fill(newPassword);
	await page.locator("#confirmPassword").fill(newPassword);
	await page.locator("#confirmPassword").dispatchEvent("blur");
	await Promise.all([
		page.waitForURL("**/auth/login**", { timeout: 30_000 }),
		page.locator("#resetBtn").click(),
	]);
	await expect(page.locator("#successBanner")).toContainText(/reset successfully/i, { timeout: 15_000 });
}
