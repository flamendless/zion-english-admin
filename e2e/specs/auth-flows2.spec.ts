import { test } from "@playwright/test";
import {
	passwordResetNewPassword,
	pendingTeacherForDelete,
	pendingTeacherForUnapprove,
	pendingTeacherRegistration,
	superuserCreatedTeacher,
} from "../helpers/auth-fixtures";
import { loadCredentials } from "../helpers/credentials";
import { completePasswordReset, requestPasswordReset } from "../helpers/forgot-password";
import { login, loginAsSuperuser } from "../helpers/login";
import { logout } from "../helpers/logout";
import {
	registerTeacher,
	registerTeacherAsSuperuser,
	registerTeacherExpectError,
} from "../helpers/register-teacher";
import {
	approvePendingTeacher,
	deletePendingTeacher,
	unapproveTeacher,
} from "../helpers/teachers-admin";

const creds = loadCredentials();

test.describe.configure({ mode: "serial" });

test("invalid login", async ({ page }) => {
	await login(page, creds.email, "WrongPassword1!", {
		expectError: /Invalid email or password/i,
	});
});

test("duplicate registration", async ({ page }) => {
	await registerTeacherExpectError(page, pendingTeacherRegistration, /already exists|awaiting approval/i);
});

test("superuser registers approved teacher", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await registerTeacherAsSuperuser(page, superuserCreatedTeacher);
});

test("delete pending teacher", async ({ page }) => {
	await registerTeacher(page, pendingTeacherForDelete);
	await loginAsSuperuser(page, creds);
	await deletePendingTeacher(page, pendingTeacherForDelete.email);
});

test("unapprove teacher", async ({ page }) => {
	await registerTeacher(page, pendingTeacherForUnapprove);
	await loginAsSuperuser(page, creds);
	await approvePendingTeacher(page, pendingTeacherForUnapprove.email);
	await unapproveTeacher(page, pendingTeacherForUnapprove.email);
	await logout(page);
	await login(page, pendingTeacherForUnapprove.email, pendingTeacherForUnapprove.password, {
		expectError: /pending approval/i,
	});
	await loginAsSuperuser(page, creds);
	await approvePendingTeacher(page, pendingTeacherForUnapprove.email);
});

test("forgot password and reset", async ({ page }) => {
	await requestPasswordReset(page, pendingTeacherRegistration.email);
	await completePasswordReset(page, passwordResetNewPassword);
	await login(page, pendingTeacherRegistration.email, passwordResetNewPassword);
});

test("login fails with old password after reset", async ({ page }) => {
	await login(page, pendingTeacherRegistration.email, pendingTeacherRegistration.password, {
		expectError: /Invalid email or password/i,
	});
});
