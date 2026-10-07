const pht = "Asia/Manila";

export function todayPHT(): string {
	return new Intl.DateTimeFormat("en-CA", { timeZone: pht }).format(new Date());
}

/** Add calendar days to an ISO date (YYYY-MM-DD) in PHT. */
export function addDaysPHT(isoDate: string, days: number): string {
	const [y, m, d] = isoDate.split("-").map((v) => parseInt(v, 10));
	const utc = Date.UTC(y, m - 1, d + days);
	return new Intl.DateTimeFormat("en-CA", { timeZone: pht }).format(new Date(utc));
}

export function dayOfMonth(isoDate: string): number {
	return parseInt(isoDate.split("-")[2], 10);
}
