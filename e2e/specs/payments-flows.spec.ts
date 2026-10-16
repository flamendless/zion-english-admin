import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { teacherCompactDisplayName } from "../helpers/display-names";
import { loginAsSuperuser, loginAsTeacher } from "../helpers/login";
import {
	filterReportsByTeacherQuery,
	openSendPaymentForTeacherRow,
	submitReportPaymentModal,
} from "../helpers/reports";
import { confirmFirstPaymentReceived, filterPayments, gotoPayments } from "../helpers/payments";

const creds = loadCredentials();
const paymentRef = `E2E-PAY-${Date.now()}`;

test.describe.configure({ mode: "serial" });

test("superuser sends payment from reports", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await filterReportsByTeacherQuery(page, "");
	await openSendPaymentForTeacherRow(page, teacherCompactDisplayName("E2E Teacher"));
	await submitReportPaymentModal(page, { referenceNumber: paymentRef, amount: "500" });
	await expect(page.locator("#successBanner, #errorBanner")).toBeVisible({ timeout: 10_000 });
});

test("teacher confirms payment received", async ({ page }) => {
	await loginAsTeacher(page, creds);
	await gotoPayments(page);
	await filterPayments(page);
	await expect(page.locator("#paymentsTableBody")).toContainText(paymentRef, { timeout: 15_000 });
	await confirmFirstPaymentReceived(page);
});

test("superuser sees received payment", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await gotoPayments(page);
	await page.locator("#paymentSearch").fill("E2E");
	await filterPayments(page);
	await expect(page.locator("#paymentsTableBody")).toContainText(paymentRef);
});
