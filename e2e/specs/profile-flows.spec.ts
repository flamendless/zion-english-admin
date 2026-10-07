import { test, expect } from "@playwright/test";
import { loginAsTeacher } from "../helpers/login";
import { loadCredentials } from "../helpers/credentials";
import { gotoProfile, updateProfileMiddleName } from "../helpers/profile";

const creds = loadCredentials();

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("update middle name on profile", async ({ page }) => {
	await updateProfileMiddleName(page, "E2E");
	await gotoProfile(page);
	await expect(page.locator(".profile-identity h2")).toContainText("E2E E2E Teacher");
});
