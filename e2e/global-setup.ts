import fs from "fs";
import path from "path";
import { loadRepoEnv } from "./helpers/load-env";

export default async function globalSetup() {
	const repoRoot = path.resolve(__dirname, "..");
	loadRepoEnv(repoRoot);
	const dataDir = path.join(repoRoot, "e2e", ".data");
	fs.mkdirSync(dataDir, { recursive: true });
}
