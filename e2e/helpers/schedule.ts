import { expect, Page } from "@playwright/test";
import { loadCredentials } from "./credentials";

export async function gotoSchedule(page: Page) {
	await page.goto("schedule");
	await expect(page.getByRole("heading", { name: "Class Schedule" })).toBeVisible();
	await page.locator("#calendarGrid button.calendar-day").first().waitFor({ timeout: 15_000 });
}

export async function openScheduleDay(page: Page, isoDate: string) {
	await gotoSchedule(page);
	const dayBtn = page.locator(`#calendarGrid button.calendar-day[data-date="${isoDate}"]`);
	await expect(dayBtn).toBeVisible({ timeout: 15_000 });
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/schedule/partials/day-timeline") && r.ok(),
		),
		dayBtn.click(),
	]);
	await expect(page.locator("#scheduleModal")).toBeVisible();
}

function viewClassButton(page: Page, studentName: string) {
	return page.getByRole("button", {
		name: new RegExp(`View class for ${studentName}`),
	});
}

export async function scheduleRowForStudent(page: Page, studentName: string, isoDate: string) {
	await openScheduleDay(page, isoDate);
	await expect(viewClassButton(page, studentName)).toBeVisible({ timeout: 20_000 });
	return viewClassButton(page, studentName);
}

export async function openScheduledClassView(page: Page, studentName: string, isoDate: string) {
	await openScheduleDay(page, isoDate);
	await viewClassButton(page, studentName).click();
	await expect(page.locator("#classViewModalHost .modal-dialog")).toBeVisible({ timeout: 15_000 });
	return page.locator("#classViewModalHost");
}

export async function clickScheduleAction(
	page: Page,
	studentName: string,
	isoDate: string,
	action: string,
) {
	const host = await openScheduledClassView(page, studentName, isoDate);
	await host.getByRole("button", { name: action }).click();
}

export async function openOverdueScheduledClass(page: Page) {
	const creds = loadCredentials();
	const isoDate = creds.overdueScheduledDate;
	if (!isoDate) {
		throw new Error("overdueScheduledDate missing from e2e credentials");
	}
	return openScheduledClassView(page, creds.studentName, isoDate);
}

export async function rescheduleViaEditModal(
	page: Page,
	studentName: string,
	oldDate: string,
	newDate: string,
	newRate?: string,
) {
	await clickScheduleAction(page, studentName, oldDate, "Edit");
	await expect(page.getByRole("heading", { name: "Edit Scheduled Class" })).toBeVisible();
	const dateInput = page.locator("#editScheduledClassForm #date");
	await dateInput.fill(newDate);
	if (newRate) {
		await page.locator("#editScheduledClassForm #rate").fill(newRate);
	}
	await page.locator("#editScheduledClassSubmitBtn").click();
}
