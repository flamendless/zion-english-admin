import { test } from "@playwright/test";
import { pendingTeacherRegistration } from "../helpers/auth-fixtures";
import { loadCredentials } from "../helpers/credentials";
import { login, loginAsSuperuser } from "../helpers/login";
import { logout } from "../helpers/logout";
import { registerTeacher } from "../helpers/register-teacher";
import { approvePendingTeacher } from "../helpers/teachers-admin";

const creds = loadCredentials();
const pending = pendingTeacherRegistration;

test.describe.configure({ mode: "serial" });

test("superuser login", async ({ page }) => {
	await loginAsSuperuser(page, creds);
});

test("teacher registration", async ({ page }) => {
	await registerTeacher(page, pending);
});

test("teacher login pending", async ({ page }) => {
	await login(page, pending.email, pending.password, {
		expectError: /pending approval/i,
	});
});

test("superuser approves teacher", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await approvePendingTeacher(page, pending.email);
});

test("teacher login approved", async ({ page }) => {
	await login(page, pending.email, pending.password);
});

test("teacher logout", async ({ page }) => {
	await login(page, pending.email, pending.password);
	await logout(page);
});

test("superuser logout", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await logout(page);
});
