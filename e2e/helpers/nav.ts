import { expect, Page } from "@playwright/test";

export async function expectPageLoads(
	page: Page,
	path: string,
	heading: RegExp | string,
) {
	await page.goto(path);
	const headingLocator =
		typeof heading === "string" && heading === "My Profile"
			? page.getByRole("heading", { name: heading, level: 1 })
			: page.getByRole("heading", { name: heading });
	await expect(headingLocator).toBeVisible({ timeout: 15_000 });
	await expect(page.locator("body")).not.toContainText("404 page not found");
}
