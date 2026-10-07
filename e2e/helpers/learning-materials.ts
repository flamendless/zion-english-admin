import { expect, Page } from "@playwright/test";

export type CreateLearningMaterialOptions = {
	title: string;
	url: string;
	description?: string;
};

export async function createLearningMaterial(page: Page, opts: CreateLearningMaterialOptions) {
	await page.goto("learning-materials");
	await page.locator("#openCreateMaterialBtn").click();
	await expect(page.locator("#lmMaterialForm")).toBeVisible();
	await page.locator("#lmTitle").fill(opts.title);
	await page.locator("#lmDescription").fill(opts.description ?? "E2E learning material description");
	await page.locator("#lmURL").fill(opts.url);
	await page.locator("#lmTagInput").fill("e2e");
	await page.locator("#lmTagInput").blur();
	await expect(page.locator("#lmMaterialForm input[name='tags']")).toHaveCount(1);
	await Promise.all([
		page.waitForResponse(
			(r) =>
				r.url().includes("/learning-materials/create") &&
				r.request().method() === "POST" &&
				r.status() >= 200 &&
				r.status() < 400,
		),
		page.locator("#lmMaterialForm .lm-submit-btn").click(),
	]);
	await expect(page.locator(".lm-grid .lm-card-title")).toContainText(opts.title, {
		timeout: 15_000,
	});
}
