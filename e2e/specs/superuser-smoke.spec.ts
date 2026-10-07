import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser } from "../helpers/login";

const creds = loadCredentials();

test.describe.configure({ mode: "serial" });

test.beforeEach(async ({ page }) => {
	await loginAsSuperuser(page, creds);
});

test("students hub loads", async ({ page }) => {
	await page.goto("students");
	await expect(page.getByRole("heading", { name: "Students" })).toBeVisible();
});

test("process page loads", async ({ page }) => {
	await page.goto("process");
	await expect(page.locator("#teacherSearch")).toBeVisible({ timeout: 15_000 });
});

test("reports filter loads table", async ({ page }) => {
	await page.goto("reports");
	await expect(page.getByRole("heading", { name: "Reports" })).toBeVisible();
	await expect(page.locator("#reportsTable")).toBeVisible({ timeout: 15_000 });
	await page.getByRole("button", { name: "Filter" }).click();
	await expect(page.locator("#reportsTable")).toBeVisible({ timeout: 20_000 });
});
