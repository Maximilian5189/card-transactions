// place files you want to import through the `$lib` alias in this folder.

export type transaction = {
	date: number;
	name: string;
	amount: number;
};

export function getStartOfWeekTimestamp(offset = 0) {
	const today = new Date();

	// day of the week is like so: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
	const dayOfWeek = today.getDay();
	const diffToMonday = (dayOfWeek === 0 ? 6 : dayOfWeek - 1) * 24 * 60 * 60 * 1000;

	const startOfWeek = new Date(today.getTime() - diffToMonday);
	startOfWeek.setDate(startOfWeek.getDate() + offset * 7);
	startOfWeek.setHours(0, 0, 0, 0);

	return startOfWeek.getTime();
}
