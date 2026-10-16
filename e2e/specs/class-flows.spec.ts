import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { addDaysPHT, todayPHT } from "../helpers/dates";
import { loginAsTeacher } from "../helpers/login";
import {
	clickRepeatCalendarDate,
	fillConductWizard,
	fillScheduleWizard,
	openConductWizard,
	openScheduleWizard,
} from "../helpers/wizard";
import { classTableRowWithStatus, filterClassesTable } from "../helpers/classes";
import { clickScheduleAction, gotoSchedule, scheduleRowForStudent } from "../helpers/schedule";

const creds = loadCredentials();

/** HTMX modals and global handlers use several success strings. */
const successBanner = /successfully|Saved successfully|Operation completed successfully/i;

test.describe.configure({ mode: "serial" });

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("record past class conducted", async ({ page }) => {
	const date = addDaysPHT(todayPHT(), -1);
	await openConductWizard(page);
	await fillConductWizard(page, { studentName: creds.studentName, date, status: "conducted" });
	await page.goto("classes");
	await filterClassesTable(page, { studentName: creds.studentName, testSeedFilter: "all" });
	await expect(classTableRowWithStatus(page, creds.studentName, "Conducted").first()).toBeVisible();
});

test("record past class cancelled", async ({ page }) => {
	const date = addDaysPHT(todayPHT(), -2);
	await openConductWizard(page);
	await fillConductWizard(page, {
		studentName: creds.studentName,
		date,
		status: "cancelled",
		reason: "Student unavailable E2E",
	});
	await page.goto("classes");
	await filterClassesTable(page, { studentName: creds.studentName, testSeedFilter: "all" });
	await expect(classTableRowWithStatus(page, creds.studentName, "Cancelled").first()).toBeVisible();
});

test("record past class rescheduled", async ({ page }) => {
	const date = addDaysPHT(todayPHT(), -3);
	await openConductWizard(page);
	await fillConductWizard(page, {
		studentName: creds.studentName,
		date,
		status: "rescheduled",
		reason: "E2E reschedule reason",
	});
	await page.goto("classes");
	await filterClassesTable(page, { studentName: creds.studentName, testSeedFilter: "all" });
	await expect(classTableRowWithStatus(page, creds.studentName, "Rescheduled").first()).toBeVisible();
});

test("schedule upcoming class", async ({ page }) => {
	const scheduledDate = addDaysPHT(todayPHT(), 7);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate });
	await scheduleRowForStudent(page, creds.studentName, scheduledDate);
});

test("schedule repeating classes", async ({ page }) => {
	const d1 = addDaysPHT(todayPHT(), 10);
	const d2 = addDaysPHT(todayPHT(), 11);
	await page.goto("schedule/repeat");
	await page.locator("#student").selectOption({ label: creds.studentName });
	await clickRepeatCalendarDate(page, d1);
	await clickRepeatCalendarDate(page, d2);
	await page.locator("#start_time").fill("11:00");
	await page.locator("#end_time").fill("12:00");
	await page.locator("#currency").selectOption("PHP");
	await page.locator("#rate").fill("500");
	await page.getByRole("button", { name: "Review Dates" }).click();
	await expect(page.getByRole("heading", { name: "Review selected dates" })).toBeVisible();
	await page.locator("#repeatConfirmBtn").click();
	await expect(page.locator("#successBanner")).toContainText(/created/i, { timeout: 20_000 });
	await gotoSchedule(page);
	await expect(page.locator("#calendarGrid")).toBeVisible();
});

test("repeating series list", async ({ page }) => {
	await page.goto("schedule/series");
	await expect(page.getByRole("heading", { name: "Repeating Scheduled Classes" })).toBeVisible();
	await expect(page.locator("body")).toContainText(creds.studentName);
});

test("conduct scheduled class", async ({ page }) => {
	const scheduledDate = addDaysPHT(todayPHT(), 14);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate });
	await clickScheduleAction(page, creds.studentName, scheduledDate, "Conduct");
	await expect(page.getByRole("heading", { name: "Conduct Class" })).toBeVisible();
	await page.locator("#conductClassSubmitBtn").click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("cancel scheduled class", async ({ page }) => {
	const scheduledDate = addDaysPHT(todayPHT(), 15);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate });
	await clickScheduleAction(page, creds.studentName, scheduledDate, "Cancel");
	await page.locator("#cancelClassReason").fill("E2E cancel reason");
	await page.locator("#cancelClassSubmitBtn").click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("edit scheduled class", async ({ page }) => {
	const scheduledDate = addDaysPHT(todayPHT(), 16);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate });
	await clickScheduleAction(page, creds.studentName, scheduledDate, "Edit");
	await expect(page.getByRole("heading", { name: "Edit Scheduled Class" })).toBeVisible();
	await page.locator("#editScheduledClassForm #rate").fill("650");
	await page.locator("#editScheduledClassSubmitBtn").click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("delete scheduled class", async ({ page }) => {
	const scheduledDate = addDaysPHT(todayPHT(), 17);
	await openScheduleWizard(page);
	await fillScheduleWizard(page, { studentName: creds.studentName, scheduledDate });
	await clickScheduleAction(page, creds.studentName, scheduledDate, "Delete");
	await page.locator("#deleteClassReason").fill("E2E delete schedule");
	await page.locator("#deleteClassForm").getByRole("button", { name: "Delete" }).click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("edit class record", async ({ page }) => {
	await page.goto("classes");
	await filterClassesTable(page, { studentName: creds.studentName, testSeedFilter: "all" });
	const conductedRow = classTableRowWithStatus(page, creds.studentName, "Conducted");
	const editHref = await conductedRow.first().getByRole("link", { name: "Edit" }).getAttribute("href");
	expect(editHref).toBeTruthy();
	await page.goto(editHref!);
	await expect(page).toHaveTitle(/Edit Class/i);
	await page.locator("#notes").fill("E2E updated notes");
	await page.getByRole("button", { name: "Save Changes" }).click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});

test("delete class record", async ({ page }) => {
	await page.goto("classes");
	await filterClassesTable(page, { studentName: creds.studentName, testSeedFilter: "all" });
	const conductedRow = classTableRowWithStatus(page, creds.studentName, "Conducted");
	await conductedRow.first().getByRole("button", { name: "Delete" }).click();
	await page.locator("#deleteClassReason").fill("E2E delete record");
	await page.locator("#deleteClassForm").getByRole("button", { name: "Delete" }).click();
	await expect(page.locator("#successBanner")).toContainText(successBanner, { timeout: 15_000 });
});
