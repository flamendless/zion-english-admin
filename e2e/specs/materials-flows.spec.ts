import { test } from "@playwright/test";
import { loginAsTeacher } from "../helpers/login";
import { loadCredentials } from "../helpers/credentials";
import { createLearningMaterial } from "../helpers/learning-materials";

const creds = loadCredentials();

test.beforeEach(async ({ page }) => {
	await loginAsTeacher(page, creds);
});

test("create learning material", async ({ page }) => {
	const title = `E2E Material ${Date.now()}`;
	await createLearningMaterial(page, {
		title,
		url: "https://example.com/e2e-material",
	});
});
