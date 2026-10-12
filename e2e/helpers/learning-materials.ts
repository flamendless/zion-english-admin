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
	await expect(page.locator(".lm-card-title", { hasText: opts.title })).toBeVisible({
		timeout: 15_000,
	});
}

export async function editLearningMaterialTitle(page: Page, currentTitle: string, newTitle: string) {
	const card = page.locator(".lm-card", { hasText: currentTitle });
	await card.getByRole("button", { name: "Edit" }).click();
	const editForm = page.locator("#lmEditModalHost #lmMaterialForm");
	await expect(editForm).toBeVisible();
	await editForm.locator("#lmTitle").fill(newTitle);
	await Promise.all([
		page.waitForResponse(
			(r) =>
				r.url().includes("/learning-materials/") &&
				r.request().method() === "POST" &&
				r.status() >= 200 &&
				r.status() < 400,
		),
		editForm.locator(".lm-submit-btn").click(),
	]);
	await expect(page.locator(".lm-card-title", { hasText: newTitle })).toBeVisible({
		timeout: 15_000,
	});
}

export async function deleteLearningMaterialByTitle(page: Page, title: string) {
	const card = page.locator(".lm-card", { hasText: title });
	await card.getByRole("button", { name: "Edit" }).click();
	const editForm = page.locator("#lmEditModalHost #lmMaterialForm");
	await expect(editForm).toBeVisible();
	await editForm.locator("#lmStatus").selectOption("deleted");
	await Promise.all([
		page.waitForResponse(
			(r) =>
				r.url().includes("/learning-materials/") &&
				r.request().method() === "POST" &&
				r.status() >= 200 &&
				r.status() < 400,
		),
		editForm.locator(".lm-submit-btn").click(),
	]);
	await expect(page.locator(".lm-grid .lm-card-title", { hasText: title })).toHaveCount(0, {
		timeout: 15_000,
	});
}
