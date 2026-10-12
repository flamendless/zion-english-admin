import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser } from "../helpers/login";
import { expectPageLoads } from "../helpers/nav";

const creds = loadCredentials();

const superuserRoutes: { path: string; heading: RegExp | string; hubTitle?: boolean }[] = [
	{ path: "teachers", heading: "Teachers" },
	{ path: "reports/history", heading: "Reports History" },
	{ path: "teachers-earnings", heading: "Teachers Earnings" },
	{ path: "payments", heading: "Payment History" },
	{ path: "student-relationships", heading: "Students Relationship Diagram" },
	{ path: "affiliate-ads/reports", heading: "Ads/Affiliates reports" },
	{ path: "feature-flags", heading: "Feature Flags" },
	{ path: "settings", heading: "Settings" },
	{ path: "meta", heading: "Meta", hubTitle: true },
	{ path: "affiliates", heading: "Affiliates" },
	{ path: "ads", heading: "Ads" },
	{ path: "plans", heading: "Plans" },
	{ path: "announcements", heading: "Announcements" },
	{ path: "changelogs", heading: "Changelogs" },
	{ path: "logs", heading: "Logs" },
	{ path: "upload-logs", heading: "Upload Logs" },
];

test.beforeEach(async ({ page }) => {
	await loginAsSuperuser(page, creds);
});

test("superuser nav pages load", async ({ page }) => {
	for (const route of superuserRoutes) {
		await expectPageLoads(page, route.path, route.heading, {
			hubTitle: route.hubTitle,
		});
	}
});

test("reports history filter loads table", async ({ page }) => {
	await page.goto("reports/history");
	await expect(page.getByRole("heading", { name: "Reports History" })).toBeVisible();
	await page.locator("#reportHistoryFilterBtn").click();
	await expect(page.locator("#reportsHistoryTable")).toBeVisible({ timeout: 20_000 });
});
