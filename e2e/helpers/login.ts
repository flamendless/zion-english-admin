import { expect, Page } from "@playwright/test";
import { E2ECredentials } from "./credentials";

export type LoginOptions = {
	expectError?: string | RegExp;
};

export async function login(page: Page, email: string, password: string, opts?: LoginOptions) {
	await page.goto("auth/login");
	await page.locator("#email").fill(email);
	await page.locator("#password").fill(password);
	const loginResponse = await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/auth/login") && r.request().method() === "POST",
		),
		page.getByRole("button", { name: "Login" }).click(),
	]);

	if (opts?.expectError) {
		const loginError = page.locator("#loginError");
		await expect(loginError).toContainText(opts.expectError, { timeout: 15_000 });
		return;
	}

	expect(loginResponse[0].status()).toBeGreaterThanOrEqual(300);
	expect(loginResponse[0].status()).toBeLessThan(400);
	await page.waitForURL("**/dashboard**");
	await expect(page.getByRole("heading", { name: "Quick actions" })).toBeVisible({ timeout: 15_000 });
}

export async function loginAsTeacher(page: Page, creds: E2ECredentials) {
	await login(page, creds.email, creds.password);
}

export async function loginAsSuperuser(page: Page, creds: E2ECredentials) {
	await page.goto("auth/login");
	await page.locator("#email").fill(creds.superuserEmail);
	await page.locator("#password").fill(creds.superuserPassword);
	await page.getByRole("button", { name: "Login" }).click();
	await page.waitForURL("**/dashboard**");
	await expect(page.getByRole("heading", { name: /Welcome, superuser/i })).toBeVisible({
		timeout: 15_000,
	});
}
