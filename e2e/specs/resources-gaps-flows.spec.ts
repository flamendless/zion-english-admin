import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser, loginAsTeacher } from "../helpers/login";
import { logout } from "../helpers/logout";
import {
	createLearningMaterial,
	editLearningMaterialTitle,
} from "../helpers/learning-materials";
import { restoreE2ETeacherPassword } from "../helpers/restore-teacher";

const creds = loadCredentials();
const tempPassword = "E2eTempPass2!";

test.describe.configure({ mode: "serial" });

test("learning material edit and delete", async ({ page }) => {
	await loginAsTeacher(page, creds);
	const title = `E2E LM Edit ${Date.now()}`;
	const renamed = `${title} Updated`;
	await createLearningMaterial(page, {
		title,
		url: "https://example.com/e2e-lm-edit",
	});
	await editLearningMaterialTitle(page, title, renamed);
});

test("superuser creates training material", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await page.goto("training-materials");
	await expect(page.getByRole("heading", { name: "Training Materials" })).toBeVisible();
	await page.locator("#openCreateTrainingBtn").click();
	await expect(page.locator("#tmTitle")).toBeVisible();
	const title = `E2E Training ${Date.now()}`;
	await page.locator("#tmTitle").fill(title);
	await page.locator("#tmDescription").fill("E2E training material description.");
	await page.locator("#tmURL").fill("https://www.youtube.com/watch?v=dQw4w9WgXcQ");
	await page.locator("#tmTagInput").fill("e2e");
	await page.locator("#tmTagInput").press("Enter");
	await expect(page.locator("#tmTagChips .tm-tag-chip")).toHaveCount(1);
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/training-materials/create") && r.request().method() === "POST",
		),
		page.locator(".tm-submit-btn").click(),
	]);
	await expect(page.locator(".tm-grid")).toContainText(title, { timeout: 15_000 });
});

test("teacher opens training materials library", async ({ page }) => {
	await loginAsTeacher(page, creds);
	await page.goto("training-materials");
	await expect(page.getByRole("heading", { name: "Training Materials" })).toBeVisible();
});

test("teacher changes password and restores", async ({ page }) => {
	await loginAsTeacher(page, creds);
	await page.goto("profile");
	await page.getByRole("tab", { name: "Security" }).click();
	await page.locator("#currentPassword").fill(creds.password);
	await page.locator("#newPassword").fill(tempPassword);
	await page.locator("#confirmPassword").fill(tempPassword);
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/profile/password") && r.request().method() === "POST",
		),
		page.locator('form[action*="profile/password"]').getByRole("button", { name: "Update password" }).click(),
	]);
	await logout(page);
	await loginAsTeacher(page, { ...creds, password: tempPassword });
	await expect(page.getByRole("heading", { name: "Quick actions" })).toBeVisible();
});

test.afterAll(() => {
	restoreE2ETeacherPassword();
});
