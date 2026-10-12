import { defineConfig, devices } from "@playwright/test";
import path from "path";
import { loadRepoEnv } from "./helpers/load-env";

const repoRoot = path.resolve(__dirname, "..");
loadRepoEnv(repoRoot);
const e2ePort = process.env.E2E_PORT || "18080";
const dbPath = path.join(repoRoot, "e2e", ".data", "e2e.db");
const adminPath = "/zion-english-admin";
const baseURL = `http://localhost:${e2ePort}${adminPath}/`;

export default defineConfig({
	testDir: "./specs",
	globalSetup: require.resolve("./global-setup"),
	fullyParallel: false,
	workers: 1,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	reporter: [["list"]],
	use: {
		baseURL,
		trace: "on-first-retry",
		actionTimeout: 20_000,
		...devices["Desktop Chrome"],
	},
	timeout: 90_000,
	webServer: {
		// Seed before web so Playwright globalSetup (which may run after health) never resets DB while the server holds it open.
		command: `go run . e2e-seed && ZION_DB_PATH="${dbPath}" PORT="${e2ePort}" go run . web -p ${e2ePort}`,
		cwd: repoRoot,
		url: `${baseURL}health`,
		reuseExistingServer: process.env.E2E_REUSE_SERVER === "1",
		timeout: 120_000,
		env: {
			...process.env,
			ZION_DB_PATH: dbPath,
			PORT: e2ePort,
		},
	},
});
