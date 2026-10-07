import { expect, Page } from "@playwright/test";
import { expectTeachersListReady } from "./teachers-admin";

export type RegisterTeacherData = {
	firstName: string;
	lastName: string;
	email: string;
	password: string;
	mobileNumber: string;
	birthdate: string;
	address: string;
	ratePerClass: string;
	currency: string;
	driveUrl: string;
};

export async function fillRegisterTeacherForm(page: Page, data: RegisterTeacherData) {
	await page.locator("#firstName").fill(data.firstName);
	await page.locator("#lastName").fill(data.lastName);
	await page.locator("#birthdate").fill(data.birthdate);
	await page.locator("#address").fill(data.address);
	await page.locator("#driveUrl").fill(data.driveUrl);
	await page.locator("#mobileNumber").fill(data.mobileNumber);
	await page.locator("#email").fill(data.email);
	await page.locator("#password").fill(data.password);
	await page.locator("#retypePassword").fill(data.password);
	await page.locator("#currency").selectOption(data.currency);
	await page.locator("#ratePerClass").fill(data.ratePerClass);
}

export async function registerTeacher(page: Page, data: RegisterTeacherData) {
	await page.goto("teachers/register");
	await fillRegisterTeacherForm(page, data);
	await page.locator("#submitBtn").click();
	await page.waitForURL("**/auth/login**", { timeout: 30_000 });
	await expect(page.locator("#successBanner")).toContainText(/admin approval/i, { timeout: 15_000 });
}

export async function registerTeacherExpectError(page: Page, data: RegisterTeacherData, error: RegExp) {
	await page.goto("teachers/register");
	await fillRegisterTeacherForm(page, data);
	await page.locator("#submitBtn").click();
	await expect(page.locator("#logOutput")).toContainText(error, { timeout: 15_000 });
}

export async function registerTeacherAsSuperuser(page: Page, data: RegisterTeacherData) {
	await page.goto("teachers/register");
	await fillRegisterTeacherForm(page, data);
	await page.locator("#submitBtn").click();
	await page.waitForURL("**/teachers**", { timeout: 30_000 });
	await expectTeachersListReady(page);
	await page.locator("#q").fill(data.email);
	await page.locator("#listFilterSubmit").click();
	const row = page.locator("#teachersTable tbody tr", { hasText: data.email });
	await expect(row).toContainText("Approved", { timeout: 15_000 });
}
