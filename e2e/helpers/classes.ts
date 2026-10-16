import { expect, Locator, Page } from "@playwright/test";

export function classTableRowWithStatus(
	page: Page,
	studentName: string,
	statusLabel: string,
): Locator {
	return page.locator("#classesTable tbody tr", { hasText: studentName }).filter({
		has: page.locator(`[aria-label="${statusLabel}"]`),
	});
}

export async function filterClassesTable(
	page: Page,
	opts: { studentName?: string; testSeedFilter?: "all" | "only" | "" } = {},
) {
	if (opts.testSeedFilter !== undefined) {
		await page.locator("#testSeedFilter").selectOption(opts.testSeedFilter);
	}
	if (opts.studentName) {
		await page.locator("#studentSearch").fill(opts.studentName);
	}
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/classes/partials/rows") && r.ok(),
			{ timeout: 20_000 },
		),
		page.locator("#filterBtn").click(),
	]);
	await expect(page.locator("#classesTableBody")).not.toContainText("Loading classes...", {
		timeout: 20_000,
	});
}
