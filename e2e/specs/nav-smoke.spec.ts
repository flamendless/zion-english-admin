import { test } from "@playwright/test";
import { loginAsTeacher } from "../helpers/login";
import { loadCredentials } from "../helpers/credentials";
import { expectPageLoads } from "../helpers/nav";

const creds = loadCredentials();

const teacherRoutes: { path: string; heading: RegExp | string }[] = [
	{ path: "dashboard", heading: "Quick actions" },
	{ path: "my-students", heading: "My Students" },
	{ path: "classes", heading: "My Classes" },
	{ path: "schedule", heading: "Class Schedule" },
	{ path: "schedule/repeat", heading: "Schedule Repeating Classes" },
	{ path: "schedule/series", heading: "Repeating Scheduled Classes" },
	{ path: "profile", heading: "My Profile" },
	{ path: "learning-materials", heading: "Learning Library" },
	{ path: "guides", heading: "Guides" },
	{ path: "analytics", heading: "My Analytics" },
	{ path: "my-earnings", heading: "My Earnings" },
];

test("teacher nav pages load", async ({ page }) => {
	await loginAsTeacher(page, creds);
	for (const route of teacherRoutes) {
		await expectPageLoads(page, route.path, route.heading);
	}
});
