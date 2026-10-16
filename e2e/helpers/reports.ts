import { expect, Page } from "@playwright/test";

export async function gotoReports(page: Page) {
	await page.goto("reports");
	await expect(page.getByRole("heading", { name: "Reports" })).toBeVisible();
	await expect(page.locator("#reportsTableBody")).not.toContainText("Loading reports...", {
		timeout: 20_000,
	});
}

export async function filterReportsByTeacherQuery(page: Page, query: string) {
	await gotoReports(page);
	await page.locator("#reportSearch").fill(query);
	const rowsResponse = await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/reports/partials/rows") && r.ok(),
			{ timeout: 20_000 },
		),
		page.locator("#filterBtn").click(),
	]);
	expect(rowsResponse[0].ok()).toBeTruthy();
}

export async function openSendPaymentForTeacherRow(page: Page, teacherNameMatch: string) {
	const row = page.locator("#reportsTableBody tr", { hasText: teacherNameMatch });
	await expect(row).toBeVisible({ timeout: 15_000 });
	await row.getByRole("button", { name: "Send payment" }).click();
	await expect(page.getByRole("heading", { name: "Send payment" })).toBeVisible({
		timeout: 15_000,
	});
}

export async function submitReportPaymentModal(page: Page, opts: {
	referenceNumber: string;
	amount: string;
}) {
	const modal = page.locator("#reportPaymentModal");
	await modal.locator("#referenceNumber").fill(opts.referenceNumber);
	await modal.locator("#amount").fill(opts.amount);
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/payment") && r.request().method() === "POST" && r.ok(),
		),
		modal.getByRole("button", { name: "Send payment" }).click(),
	]);
}
