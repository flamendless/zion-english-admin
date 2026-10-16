import { expect, Page } from "@playwright/test";
import { loadCredentials } from "./credentials";

const calendarMonthNames = [
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
];

function parseCalMonthLabel(label: string): { year: number; month: number } {
	const parts = label.trim().split(/\s+/);
	const year = parseInt(parts[parts.length - 1], 10);
	const month = calendarMonthNames.indexOf(parts[0]) + 1;
	return { year, month };
}

/** Schedule calendar only renders the visible month; jump prev/next until isoDate is on screen. */
async function ensureScheduleCalendarShowsDate(page: Page, isoDate: string) {
	const [targetYear, targetMonth] = isoDate.split("-").map((v) => parseInt(v, 10));
	const dayBtn = page.locator(`#calendarGrid button.calendar-day[data-date="${isoDate}"]`);
	const monthLabel = page.locator("#calMonthLabel");
	for (let i = 0; i < 24; i++) {
		if (await dayBtn.isVisible()) {
			return;
		}
		const text = (await monthLabel.textContent())?.trim() ?? "";
		const { year, month } = parseCalMonthLabel(text);
		const nav = targetYear > year || (targetYear === year && targetMonth > month)
			? page.locator("#calNext")
			: page.locator("#calPrev");
		await Promise.all([
			page.waitForResponse((r) => r.url().includes("/api/scheduled-classes") && r.ok()),
			nav.click(),
		]);
	}
	await expect(dayBtn).toBeVisible({ timeout: 15_000 });
}

export async function gotoSchedule(page: Page) {
	await page.goto("schedule");
	await expect(page.getByRole("heading", { name: "Class Schedule" })).toBeVisible();
	await page.locator("#calendarGrid button.calendar-day").first().waitFor({ timeout: 15_000 });
}

export async function expectCalendarMonthChanges(page: Page) {
	await gotoSchedule(page);
	const label = page.locator("#calMonthLabel");
	const before = (await label.textContent())?.trim() ?? "";
	await page.locator("#calNext").click();
	await expect(label).not.toHaveText(before, { timeout: 15_000 });
	const afterNext = (await label.textContent())?.trim() ?? "";
	await page.locator("#calPrev").click();
	await expect(label).not.toHaveText(afterNext, { timeout: 15_000 });
}

export async function openScheduleDay(page: Page, isoDate: string) {
	await gotoSchedule(page);
	await ensureScheduleCalendarShowsDate(page, isoDate);
	const dayBtn = page.locator(`#calendarGrid button.calendar-day[data-date="${isoDate}"]`);
	await dayBtn.click();
	await expect(page.locator("#scheduleModal")).toBeVisible({ timeout: 20_000 });
	await page
		.waitForResponse((r) => r.url().includes("/schedule/partials/day-timeline") && r.ok(), {
			timeout: 20_000,
		})
		.catch(() => {});
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
	const studentName = creds.siblingStudentName ?? creds.studentName;
	return openScheduledClassView(page, studentName, isoDate);
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
