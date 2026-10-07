import { expect, Page } from "@playwright/test";

export type RegisterStudentOptions = {
	name: string;
	rate: string;
	currency?: string;
};

export async function registerStudent(page: Page, opts: RegisterStudentOptions) {
	await page.goto("students/register");
	await page.locator("#name").fill(opts.name);
	await page.locator("#ratePerClass").fill(opts.rate);
	await page.locator("#currency").selectOption({ value: opts.currency ?? "PHP" });
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/students/register") && r.request().method() === "POST" && r.ok(),
		),
		page.locator("#registerForm").getByRole("button", { name: "Register Student" }).click(),
	]);
	await page.goto("my-students");
	await page.locator("#q").fill(opts.name);
	await page.locator("#listFilterSubmit").click();
	await expect(page.locator("table")).toContainText(opts.name, { timeout: 15_000 });
}
