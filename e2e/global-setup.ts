import { execSync } from "child_process";
import fs from "fs";
import path from "path";
import { loadRepoEnv } from "./helpers/load-env";

export default async function globalSetup() {
	const repoRoot = path.resolve(__dirname, "..");
	loadRepoEnv(repoRoot);
	const dataDir = path.join(repoRoot, "e2e", ".data");
	fs.mkdirSync(dataDir, { recursive: true });
	const dbPath = path.join(dataDir, "e2e.db");

	execSync("go run . e2e-seed", {
		cwd: repoRoot,
		env: { ...process.env, ZION_DB_PATH: dbPath },
		stdio: "inherit",
	});
}
