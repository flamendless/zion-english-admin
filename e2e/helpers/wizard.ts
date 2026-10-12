import { expect, Locator, Page } from "@playwright/test";

export async function openRecordClassesChoice(page: Page) {
	await page.goto("classes");
	await page.getByRole("button", { name: "New Class" }).click();
	await expect(page.getByRole("heading", { name: "Record classes" })).toBeVisible();
}

export async function openConductWizard(page: Page) {
	await openRecordClassesChoice(page);
	await page.getByRole("button", { name: "Record a past class" }).click();
	await expect(page.getByRole("heading", { name: "Record a past class" })).toBeVisible();
	const studentSelect = page.locator("#conductClassWizardModal #conductWizard_student");
	await expect(studentSelect.locator("option[value]:not([value=''])")).not.toHaveCount(0, {
		timeout: 15_000,
	});
}

export async function openScheduleWizard(page: Page) {
	await openRecordClassesChoice(page);
	await page.getByRole("button", { name: "Schedule an upcoming class" }).click();
	await expect(page.getByRole("heading", { name: "Schedule an upcoming class" })).toBeVisible();
	const studentSelect = page.locator("#scheduleClassWizardModal #scheduleWizard_student");
	await expect(studentSelect.locator("option[value]:not([value=''])")).not.toHaveCount(0, {
		timeout: 15_000,
	});
}

async function wizardNext(modal: Locator, nextId: string) {
	const next = modal.locator(`#${nextId}`);
	await expect(next).toBeEnabled({ timeout: 10_000 });
	await next.click();
}

async function selectStudentInWizard(modal: Locator, studentName: string, selectId: string) {
	const select = modal.locator(`#${selectId}`);
	await expect(select.locator("option", { hasText: studentName })).toHaveCount(1, { timeout: 15_000 });
	await select.selectOption({ label: studentName });
	await select.dispatchEvent("change");
}

export type FillConductWizardOptions = {
	studentName: string;
	date: string;
	status?: "conducted" | "cancelled" | "rescheduled";
	reason?: string;
	trialClass?: boolean;
	learningMaterialTitle?: string;
};

export async function fillConductWizard(page: Page, opts: FillConductWizardOptions) {
	const modal = page.locator("#conductClassWizardModal");
	const status = opts.status ?? "conducted";

	await selectStudentInWizard(modal, opts.studentName, "conductWizard_student");
	await wizardNext(modal, "conductWizardNext");

	if (status !== "conducted") {
		const statusLabel =
			status === "cancelled" ? "Cancelled" : status === "rescheduled" ? "Rescheduled" : status;
		await modal.locator("label.status-radio-card").filter({ hasText: statusLabel }).click();
		await modal.locator("#conductWizard_reason").fill(opts.reason ?? "E2E test reason");
	}
	await wizardNext(modal, "conductWizardNext");

	await modal.locator("#conductWizard_date").fill(opts.date);
	await modal.locator("#conductWizard_start_time").fill("09:00");
	await modal.locator("#conductWizard_end_time").fill("10:00");
	await wizardNext(modal, "conductWizardNext");

	await modal.locator("#conductWizard_currency").selectOption("PHP");
	await modal.locator("#conductWizard_rate").fill("500");
	await wizardNext(modal, "conductWizardNext"); // materials
	if (opts.learningMaterialTitle) {
		await modal.locator("#conductWizard_lmSearch").fill(opts.learningMaterialTitle);
		await expect(modal.locator(".class-lm-search-item", { hasText: opts.learningMaterialTitle })).toBeVisible({
			timeout: 15_000,
		});
		await modal.locator(".class-lm-search-item", { hasText: opts.learningMaterialTitle }).click();
		await expect(modal.locator(".class-lm-chip")).toContainText(opts.learningMaterialTitle);
	}
	await wizardNext(modal, "conductWizardNext"); // trial
	if (opts.trialClass) {
		await modal.locator("label.trial-yesno-card").filter({ hasText: "Yes" }).click();
	}
	await wizardNext(modal, "conductWizardNext"); // notes
	await wizardNext(modal, "conductWizardNext"); // summary
	await expect(modal.locator("#conductWizardSave")).toBeVisible();
	await modal.locator("#conductWizardSave").click();
	await expect(page.locator("#successBanner")).toContainText(/class recorded/i, { timeout: 15_000 });
}

export async function fillScheduleWizard(page: Page, opts: { studentName: string; scheduledDate: string }) {
	const modal = page.locator("#scheduleClassWizardModal");

	await selectStudentInWizard(modal, opts.studentName, "scheduleWizard_student");
	await wizardNext(modal, "scheduleWizardNext");

	await modal.locator("#scheduleWizard_scheduled_date").fill(opts.scheduledDate);
	await modal.locator("#scheduleWizard_start_time").fill("10:00");
	await modal.locator("#scheduleWizard_end_time").fill("11:00");
	await wizardNext(modal, "scheduleWizardNext"); // materials
	await wizardNext(modal, "scheduleWizardNext"); // trial
	await wizardNext(modal, "scheduleWizardNext"); // summary
	await expect(modal.locator("#scheduleWizardSave")).toBeVisible();
	await modal.locator("#scheduleWizardSave").click();
	await expect(page.locator("#successBanner")).toContainText(/class scheduled/i, { timeout: 15_000 });
}

export async function clickRepeatCalendarDate(page: Page, isoDate: string) {
	const day = parseInt(isoDate.split("-")[2], 10);
	const grid = page.locator("[data-cal-grid]");
	const btn = grid.locator("button.schedule-repeat-day:not(.is-disabled)").filter({ hasText: String(day) });
	const count = await btn.count();
	if (count === 0) {
		await page.locator("[data-cal-next]").click();
		await grid.locator("button.schedule-repeat-day:not(.is-disabled)").filter({ hasText: String(day) }).click();
	} else {
		await btn.first().click();
	}
}
