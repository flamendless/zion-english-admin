import { test } from "@playwright/test";
import { loadCredentials } from "../helpers/credentials";
import { loginAsSuperuser } from "../helpers/login";
import {
	expectDiagramLoaded,
	focusStudentOnDiagram,
	gotoStudentRelationships,
} from "../helpers/student-relationships";

const creds = loadCredentials();

test.beforeEach(async ({ page }) => {
	await loginAsSuperuser(page, creds);
});

test("student relationship diagram focuses seeded student", async ({ page }) => {
	await gotoStudentRelationships(page);
	if (!creds.siblingStudentName) {
		throw new Error("siblingStudentName required in e2e credentials");
	}
	await focusStudentOnDiagram(page, creds.siblingStudentName);
	await expectDiagramLoaded(page);
});
