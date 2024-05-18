// place files you want to import through the `$lib` alias in this folder.
const isDev = import.meta.env.MODE === 'development';

export type transaction = {
	date: number;
	name: string;
	amount: number;
};
export const server = isDev ? 'http://localhost:8080' : 'https://card-transactions-backend.fly.dev';

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

export const getTransactionsTotal = async (timestamp: number, token: string) => {
	const res = await fetch(`${server}/transactions?from=${timestamp}&t=${token}`);

	const transactions = await res.json();
	let total = 0;
	transactions.forEach((transaction: transaction) => {
		total += Number(transaction.amount);
	});
	return Math.round(total * 100) / 100;
};
