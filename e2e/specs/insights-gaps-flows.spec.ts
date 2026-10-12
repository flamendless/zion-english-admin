import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsTeacher } from "../helpers/login";

const creds = loadCredentials();

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("my earnings shows cutoff rows after conducted classes", async ({ page }) => {
	await page.goto("my-earnings");
	await expect(page.getByRole("heading", { name: "My Earnings" })).toBeVisible();
	const table = page.locator(".earnings-cutoff-table tbody");
	await expect(table.locator("tr")).not.toHaveCount(0, { timeout: 15_000 });
	await expect(table).not.toContainText("No class records yet");
});

test("my analytics charts load", async ({ page }) => {
	await page.goto("analytics");
	await expect(page.getByRole("heading", { name: "My Analytics" })).toBeVisible();
	await expect(page.locator(".analytics-charts, .analytics-chart-card").first()).toBeVisible({
		timeout: 20_000,
	});
});

test("pro teacher downloads students export", async ({ page }) => {
	await page.goto("exports");
	await expect(page.getByRole("heading", { name: "My data export" })).toBeVisible();
	await expect(page.getByText("Teacher Pro required")).toHaveCount(0);
	const response = await page.request.get("exports/students");
	expect(response.ok()).toBeTruthy();
});
