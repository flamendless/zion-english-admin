import { expect, Page } from "@playwright/test";

export async function gotoStudentRelationships(page: Page) {
	await page.goto("student-relationships");
	await expect(page.getByRole("heading", { name: "Students Relationship Diagram" })).toBeVisible();
}

export async function focusStudentOnDiagram(page: Page, studentName: string) {
	await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes("/student-relationships/partials/student-search") && r.ok(),
			{ timeout: 20_000 },
		),
		page.locator("#diagramStudentSearch").fill(studentName),
	]);
	await expect(page.locator(".entity-search-item", { hasText: studentName })).toBeVisible({
		timeout: 15_000,
	});
	await page.locator(".entity-search-item", { hasText: studentName }).click();
	await expect(page.locator("#diagramStudentFocusValue")).toContainText(studentName);
}

export async function expectDiagramLoaded(page: Page) {
	await expect(page.locator("#student-relationship-graph")).not.toContainText("Loading diagram...", {
		timeout: 20_000,
	});
	await expect(page.locator("#student-relationships-empty.is-visible")).toHaveCount(0);
}
