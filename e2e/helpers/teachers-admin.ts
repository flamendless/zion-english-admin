import { expect, Page } from "@playwright/test";

export async function expectTeachersListReady(page: Page) {
	await expect(page.getByRole("heading", { name: "Teachers" })).toBeVisible({ timeout: 15_000 });
	await expect(page.locator("#q")).toBeVisible();
	await expect(page.locator("body")).not.toContainText(/Failed to (count|fetch) teachers/i);
}

async function gotoTeachersHub(page: Page) {
	const response = await page.goto("teachers");
	expect(response?.status() ?? 0).toBeLessThan(400);
	await expectTeachersListReady(page);
}

async function filterTeacherRow(page: Page, email: string) {
	await expect(page.locator("#q")).toBeVisible({ timeout: 15_000 });
	await page.locator("#q").fill(email);
	await Promise.all([
		page.waitForURL((url) => {
			const u = new URL(url);
			return u.pathname.includes("/teachers") && u.searchParams.get("q") === email;
		}),
		page.locator("#listFilterSubmit").click(),
	]);
	const row = page.locator("#teachersTable tbody tr", { hasText: email });
	await expect(row).toBeVisible({ timeout: 15_000 });
	return row;
}

async function teacherRow(page: Page, email: string) {
	await gotoTeachersHub(page);
	return filterTeacherRow(page, email);
}

export async function approvePendingTeacher(page: Page, email: string) {
	let row = await teacherRow(page, email);
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/teachers/approve") && r.request().method() === "POST",
		),
		row.getByRole("button", { name: "Approve", exact: true }).click(),
	]);
	await page.waitForURL("**/teachers**");
	await expect(page.getByRole("heading", { name: "Teachers" })).toBeVisible({ timeout: 15_000 });
	row = await filterTeacherRow(page, email);
	await expect(row.getByRole("button", { name: "Un-approve" })).toBeVisible({ timeout: 15_000 });
}

export async function deletePendingTeacher(page: Page, email: string) {
	let row = await teacherRow(page, email);
	page.once("dialog", (dialog) => dialog.accept());
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/teachers/delete") && r.request().method() === "POST",
		),
		row.getByRole("button", { name: "Delete", exact: true }).click(),
	]);
	await page.waitForURL("**/teachers**");
	await expect(page.getByRole("heading", { name: "Teachers" })).toBeVisible({ timeout: 15_000 });
	await page.locator("#q").fill(email);
	await page.locator("#listFilterSubmit").click();
	await expect(page.locator("#teachersTable tbody tr", { hasText: email })).toHaveCount(0);
}

export async function unapproveTeacher(page: Page, email: string) {
	let row = await teacherRow(page, email);
	await expect(row).toBeVisible({ timeout: 15_000 });
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/teachers/unapprove") && r.request().method() === "POST",
		),
		row.getByRole("button", { name: "Un-approve", exact: true }).click(),
	]);
	await page.waitForURL("**/teachers**");
	await expect(page.getByRole("heading", { name: "Teachers" })).toBeVisible({ timeout: 15_000 });
	row = await filterTeacherRow(page, email);
	await expect(row.getByRole("button", { name: "Approve", exact: true })).toBeVisible({
		timeout: 15_000,
	});
}
