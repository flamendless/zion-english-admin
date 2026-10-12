import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser } from "../helpers/login";

const creds = loadCredentials();

test("save feature flags without changes", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await page.goto("feature-flags");
	await expect(page.getByRole("heading", { name: "Feature Flags" })).toBeVisible();
	await page.locator("#featureFlagsForm").getByRole("button", { name: "Save" }).click();
	await expect(page.locator("#successBanner")).toContainText("Feature flags saved successfully", {
		timeout: 15_000,
	});
});
