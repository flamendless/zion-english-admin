import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { addDaysPHT, todayPHT } from "../helpers/dates";
import { loginAsSuperuser } from "../helpers/login";
import { expectTeachersListReady } from "../helpers/teachers-admin";

const creds = loadCredentials();
const announcementTitle = `E2E Announcement ${Date.now()}`;

test.describe.configure({ mode: "serial" });

test.beforeEach(async ({ page }) => {
	await loginAsSuperuser(page, creds);
});

test("superuser opens student edit form", async ({ page }) => {
	if (!creds.studentId) {
		throw new Error("studentId required in e2e credentials");
	}
	await page.goto(`students/${creds.studentId}/edit`);
	await expect(page.locator("#editStudentForm")).toBeVisible({ timeout: 15_000 });
	await expect(page.locator("#name")).toHaveValue(creds.studentName);
	const contact = `e2e-contact-${Date.now()}@test.local`;
	await page.locator("#contact").fill(contact);
	await expect(page.locator("#contact")).toHaveValue(contact);
});

test("superuser views seeded teacher row", async ({ page }) => {
	await page.goto("teachers");
	await expectTeachersListReady(page);
	const row = page.locator("#teachersTable tbody tr", { hasText: creds.email });
	await expect(row).toBeVisible({ timeout: 15_000 });
	await row.getByRole("button", { name: "E2E Teacher" }).click();
	await expect(page.locator("#teacherViewModalHost .modal-dialog")).toBeVisible({
		timeout: 20_000,
	});
});

test("superuser creates announcement", async ({ page }) => {
	await page.goto("announcements/register");
	await page.locator("#title").fill(announcementTitle);
	await page.locator("#description").fill("E2E announcement body for smoke test.");
	await page.locator("#start_date").fill(todayPHT());
	await page.locator("#end_date").fill(addDaysPHT(todayPHT(), 7));
	await page.locator('input[name="status"][value="published"]').check();
	await page.locator(".announcement-submit-btn").click();
	await page.waitForURL(/\/announcements\/?(\?.*)?$/, { timeout: 20_000 });
	await expect(page).not.toHaveURL(/\/announcements\/register/);
	await expect(page.locator(".announcement-list-title")).toContainText(announcementTitle, {
		timeout: 15_000,
	});
});
