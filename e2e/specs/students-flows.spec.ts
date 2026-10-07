import { test, expect } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsTeacher, loginAsSuperuser } from "../helpers/login";
import { registerStudent } from "../helpers/students";

const creds = loadCredentials();
const registeredName = "E2E Registered Student";

test.describe.configure({ mode: "serial" });

test("teacher registers student", async ({ page }) => {
	await loginAsTeacher(page, creds);
	await registerStudent(page, { name: registeredName, rate: "450" });
});

test("registered student on my students", async ({ page }) => {
	await loginAsTeacher(page, creds);
	await page.goto("my-students");
	await page.locator("#q").fill(registeredName);
	await page.locator("#listFilterSubmit").click();
	await expect(page.locator("table")).toContainText(registeredName);
});

test("superuser sees registered student", async ({ page }) => {
	await loginAsSuperuser(page, creds);
	await page.goto("students");
	await expect(page.getByRole("heading", { name: "Students" })).toBeVisible();
	await page.locator("#q").fill(registeredName);
	await page.locator("#listFilterSubmit").click();
	await expect(page.locator("table")).toContainText(registeredName);
	await expect(page.locator("table")).toContainText("E2E");
});
