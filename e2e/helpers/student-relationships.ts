import { expect, Page } from "@playwright/test";

export async function gotoStudentRelationships(page: Page) {
	await page.goto("student-relationships");
	await expect(page.getByRole("heading", { name: "Students Relationship Diagram" })).toBeVisible();
}

export async function waitForRelationshipGraphReady(page: Page) {
	await expect(page.locator("#student-relationship-graph")).not.toContainText("Loading diagram...", {
		timeout: 20_000,
	});
}

/** Test-seed students are omitted from diagram search; focus via prev/next over loaded graph nodes. */
export async function focusStudentOnDiagram(page: Page, studentName: string) {
	await waitForRelationshipGraphReady(page);
	const next = page.locator("#diagramStudentNext");
	const focus = page.locator("#diagramStudentFocusValue");
	for (let i = 0; i < 40; i++) {
		const text = (await focus.textContent()) ?? "";
		if (text.includes(studentName)) {
			return;
		}
		await next.click();
	}
	throw new Error(`Could not focus diagram on student ${studentName}`);
}

export async function expectDiagramLoaded(page: Page) {
	await expect(page.locator("#student-relationship-graph")).not.toContainText("Loading diagram...", {
		timeout: 20_000,
	});
	await expect(page.locator("#student-relationships-empty.is-visible")).toHaveCount(0);
}
