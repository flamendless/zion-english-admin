/** Matches internal/utils.TeacherCompactDisplayName for list cells. */
export function teacherCompactDisplayName(fullName: string): string {
	const trimmed = fullName.trim();
	if (trimmed === "" || trimmed === "-") {
		return trimmed;
	}
	const parts = trimmed.split(/\s+/);
	if (parts.length === 1) {
		return parts[0];
	}
	const last = parts[parts.length - 1];
	const initial = last.charAt(0).toUpperCase();
	return `${parts[0]} ${initial}`;
}
