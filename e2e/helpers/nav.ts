import { expect, Page } from "@playwright/test";

export type ExpectPageLoadsOptions = {
	/** Match primary hub title only (avoids duplicate heading matches on complex pages). */
	hubTitle?: boolean;
};

export async function expectPageLoads(
	page: Page,
	path: string,
	heading: RegExp | string,
	opts?: ExpectPageLoadsOptions,
) {
	await page.goto(path);
	const headingLocator =
		opts?.hubTitle
			? page.locator("h2.hub-title").filter({
					hasText: typeof heading === "string" ? heading : undefined,
				})
			: typeof heading === "string" && heading === "My Profile"
				? page.getByRole("heading", { name: heading, level: 1 })
				: page.getByRole("heading", { name: heading });
	await expect(headingLocator.first()).toBeVisible({ timeout: 15_000 });
	await expect(page.locator("body")).not.toContainText("404 page not found");
}
