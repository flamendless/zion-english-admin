import fs from "fs";
import path from "path";

export type E2ECredentials = {
	email: string;
	password: string;
	studentName: string;
	studentRate: number;
	currency: string;
	superuserEmail: string;
	superuserPassword: string;
	overdueScheduledDate?: string;
};

export function loadCredentials(): E2ECredentials {
	const filePath = path.resolve(__dirname, "..", ".auth-credentials.json");
	const raw = fs.readFileSync(filePath, "utf8");
	return JSON.parse(raw) as E2ECredentials;
}
