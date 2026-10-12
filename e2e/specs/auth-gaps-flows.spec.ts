import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsAdmin } from "../helpers/login";

const creds = loadCredentials();

test("admin teacher reaches students hub", async ({ page }) => {
	await loginAsAdmin(page, creds);
	await page.goto("students");
	await expect(page.getByRole("heading", { name: "Students" })).toBeVisible({ timeout: 15_000 });
});
