import { test } from "@playwright/test";
import { loginAsTeacher } from "../helpers/login";
import { loadCredentials } from "../helpers/credentials";
import { expectPageLoads } from "../helpers/nav";

const creds = loadCredentials();

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("guides index and getting started load", async ({ page }) => {
	await expectPageLoads(page, "guides", "Guides");
	await expectPageLoads(page, "guides/getting-started", "Getting Started");
});
