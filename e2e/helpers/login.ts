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

	const status = loginResponse[0].status();
	if (status < 300 || status >= 400) {
		const loginError = page.locator("#loginError");
		const errText = (await loginError.textContent())?.trim() ?? "";
		throw new Error(
			errText
				? `Login failed (HTTP ${status}): ${errText}`
				: `Login failed: expected redirect, got HTTP ${status}`,
		);
	}
	await page.waitForURL("**/dashboard**", { waitUntil: "domcontentloaded" });
	await expect(page.getByRole("heading", { name: "Quick actions" })).toBeVisible({ timeout: 15_000 });
}

export async function loginAsTeacher(page: Page, creds: E2ECredentials) {
	await login(page, creds.email, creds.password);
}

export async function loginAsSuperuser(page: Page, creds: E2ECredentials) {
	await login(page, creds.superuserEmail, creds.superuserPassword);
	await expect(page.getByRole("heading", { name: /Welcome, superuser/i })).toBeVisible({
		timeout: 15_000,
	});
}

export async function loginAsAdmin(page: Page, creds: E2ECredentials) {
	if (!creds.adminEmail || !creds.adminPassword) {
		throw new Error("adminEmail and adminPassword required in e2e credentials");
	}
	await login(page, creds.adminEmail, creds.adminPassword);
}
