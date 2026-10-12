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
	await focusStudentOnDiagram(page, creds.studentName);
	await expectDiagramLoaded(page);
});
