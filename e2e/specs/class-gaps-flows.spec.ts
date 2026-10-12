import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { addDaysPHT, todayPHT } from "../helpers/dates";
import { loginAsTeacher } from "../helpers/login";
import { createLearningMaterial } from "../helpers/learning-materials";
import { expectCalendarMonthChanges } from "../helpers/schedule";
import {
	fillConductWizard,
	openConductWizard,
} from "../helpers/wizard";

const creds = loadCredentials();

test.describe.configure({ mode: "serial" });

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("record trial class conducted", async ({ page }) => {
	const date = addDaysPHT(todayPHT(), -4);
	await openConductWizard(page);
	await fillConductWizard(page, {
		studentName: creds.studentName,
		date,
		trialClass: true,
	});
	await page.goto("classes");
	await expect(page.locator("#classesTableBody")).toContainText(creds.studentName, {
		timeout: 15_000,
	});
});

test("conduct class with learning material attached", async ({ page }) => {
	const lmTitle = `E2E Class LM ${Date.now()}`;
	await createLearningMaterial(page, {
		title: lmTitle,
		url: "https://example.com/e2e-class-lm",
	});
	const date = addDaysPHT(todayPHT(), -5);
	await openConductWizard(page);
	await fillConductWizard(page, {
		studentName: creds.studentName,
		date,
		learningMaterialTitle: lmTitle,
	});
});

test("persistent overdue classes panel on classes page", async ({ page }) => {
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/classes/overdue/persistent") && r.ok(),
		),
		page.goto("classes"),
	]);
	await expect(page.getByRole("heading", { name: "My Classes" })).toBeVisible();
	await expect(page.locator("#persistentOverdueClassesHost")).toContainText("Overdue classes", {
		timeout: 20_000,
	});
	await expect(page.locator("#persistentOverdueClassesHost")).toContainText(creds.studentName, {
		timeout: 20_000,
	});
});

test("schedule calendar month navigation", async ({ page }) => {
	await expectCalendarMonthChanges(page);
});
