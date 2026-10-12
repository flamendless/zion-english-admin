import { expect, Page } from "@playwright/test";

export async function gotoPayments(page: Page) {
	await page.goto("payments");
	await expect(
		page.getByRole("heading", { name: /Payment History|My Payment History/ }),
	).toBeVisible();
}

export async function filterPayments(page: Page) {
	await page.locator("#paymentFilterBtn").click();
	await expect(page.locator("#paymentsTableBody")).toBeVisible({ timeout: 15_000 });
}

export async function confirmFirstPaymentReceived(page: Page) {
	await page
		.locator("#paymentsTableBody tr")
		.first()
		.getByRole("button", { name: "Mark as received" })
		.click();
	await expect(page.getByRole("heading", { name: "Confirm payment received" })).toBeVisible();
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/received") && r.request().method() === "POST",
		),
		page.locator("#paymentReceiptModal .payment-receipt-confirm").click(),
	]);
}
