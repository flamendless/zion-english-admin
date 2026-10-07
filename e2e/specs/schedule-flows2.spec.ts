import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { addDaysPHT, todayPHT } from "../helpers/dates";
import { loginAsTeacher } from "../helpers/login";
import { fillScheduleWizard, openScheduleWizard } from "../helpers/wizard";
import {
	openOverdueScheduledClass,
	rescheduleViaEditModal,
	scheduleRowForStudent,
} from "../helpers/schedule";

const creds = loadCredentials();
const successBanner = /successfully|Saved successfully|Operation completed successfully/i;

test.describe.configure({ mode: "serial" });

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("conduct overdue scheduled class", async ({ page }) => {
	await openOverdueScheduledClass(page);
	await page.locator("#classViewModalHost").getByRole("button", { name: "Conduct" }).click();
	await expect(page.getByRole("heading", { name: "Conduct Class" })).toBeVisible();
	await page.locator("#conductClassSubmitBtn").click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("move upcoming class via edit modal", async ({ page }) => {
	const oldDate = addDaysPHT(todayPHT(), 8);
	const newDate = addDaysPHT(todayPHT(), 9);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate: oldDate });
	await rescheduleViaEditModal(page, creds.studentName, oldDate, newDate);
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
	await scheduleRowForStudent(page, creds.studentName, newDate);
});

test("delete repeating series from series list", async ({ page }) => {
	await page.goto("schedule/series");
	const row = page.locator("table tbody tr", { hasText: creds.studentName }).first();
	await expect(row).toBeVisible({ timeout: 15_000 });
	await row.getByRole("button", { name: "Delete" }).click();
	await page.locator("#deleteClassReason").fill("E2E delete series");
	await page.locator("#deleteClassForm").getByRole("button", { name: "Delete" }).click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});
