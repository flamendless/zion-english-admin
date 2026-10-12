import path from "path";
import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser } from "../helpers/login";
import { loadRepoEnv } from "../helpers/load-env";
import { expectRevealedSecretValue, revealSettingsSecret } from "../helpers/settings";

const creds = loadCredentials();
loadRepoEnv(path.resolve(__dirname, "..", ".."));

test("reveal SECRET on settings page", async ({ page }) => {
	const secret = process.env.SECRET;
	if (!secret) {
		throw new Error("SECRET must be set in repo .env for settings reveal e2e");
	}
	await loginAsSuperuser(page, creds);
	await page.goto("settings");
	await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
	await revealSettingsSecret(page, creds.superuserPassword);
	await expectRevealedSecretValue(page, secret);
});
