import { RegisterTeacherData } from "./register-teacher";

const driveUrl = "https://docs.google.com/spreadsheets/d/e2e-test-sheet-id/edit";

function registration(overrides: Partial<RegisterTeacherData> & Pick<RegisterTeacherData, "email" | "mobileNumber">): RegisterTeacherData {
	return {
		firstName: "Auth",
		lastName: "E2E",
		password: "E2eTeacherPass1!",
		birthdate: "1995-06-15",
		address: "E2E Auth Registration Address",
		ratePerClass: "500",
		currency: "PHP",
		driveUrl,
		...overrides,
	};
}

export const pendingTeacherRegistration = registration({
	email: "e2e-auth-register@zion-english.test",
	mobileNumber: "+6391700000099",
});

export const pendingTeacherForDelete = registration({
	email: "e2e-auth-delete@zion-english.test",
	mobileNumber: "+6391700000100",
	firstName: "Delete",
	lastName: "E2E",
});

export const pendingTeacherForUnapprove = registration({
	email: "e2e-auth-unapprove@zion-english.test",
	mobileNumber: "+6391700000101",
	firstName: "Unapprove",
	lastName: "E2E",
});

export const superuserCreatedTeacher = registration({
	email: "e2e-auth-superuser@zion-english.test",
	mobileNumber: "+6391700000102",
	firstName: "Super",
	lastName: "Created",
});

export const passwordResetNewPassword = "E2eAuthResetPass2!";
