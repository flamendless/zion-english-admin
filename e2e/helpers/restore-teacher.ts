import { execSync } from "child_process";
import path from "path";
import { loadRepoEnv } from "./load-env";

export function restoreE2ETeacherPassword() {
	const repoRoot = path.resolve(__dirname, "..", "..");
	loadRepoEnv(repoRoot);
	const env = { ...process.env };
	delete env.ZION_DB_PATH;
	try {
		execSync("go run . e2e-restore-teacher-password", {
			cwd: repoRoot,
			env,
			stdio: "pipe",
		});
	} catch {
		// Teardown only; next globalSetup reseeds if this fails.
	}
}
