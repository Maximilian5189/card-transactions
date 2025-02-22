// place files you want to import through the `$lib` alias in this folder.
const isDev = import.meta.env.MODE === 'development';

export type transaction = {
	date: number;
	name: string;
	amount: number;
	messageID?: string;
};
export const server = isDev ? 'http://localhost:8080' : 'https://card-transactions-backend.fly.dev';
export const nodeServer = isDev ? 'http://localhost:3000' : 'https://backend-node.fly.dev';

export function getStartOfWeekTimestamp(offset = 0) {
	const today = new Date();

	// day of the week is like so: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
	const dayOfWeek = today.getDay();
	const diffToMonday = (dayOfWeek === 0 ? 6 : dayOfWeek - 1) * 24 * 60 * 60 * 1000;

	const startOfWeek = new Date(today.getTime() - diffToMonday);
	startOfWeek.setDate(startOfWeek.getDate() - offset * 7);
	startOfWeek.setHours(0, 0, 0, 0);

	return startOfWeek.getTime();
}

export const getTransactions = async (timestamp: number, token: string) => {
	let url = `${server}/transactions?t=${token}`;
	if (timestamp) {
		url += `&from=${timestamp}`;
	}
	const res = await fetch(url);

	return await res.json();
};

export const getTransactionsTotal = async (timestamp: number, token: string) => {
	const transactions = await getTransactions(timestamp, token);

	let total = 0;
	transactions.forEach((transaction: transaction) => {
		total += transaction.amount;
	});
	return Math.round(total * 100) / 100;
};

export const postTransaction = async (transaction: transaction, token: string) => {
	const randomNumber = Math.round(Math.random() * 1000000000000);
	transaction.messageID = randomNumber.toString();
	const res = await fetch(`${server}/transaction?t=${token}`, {
		method: 'POST',
		body: JSON.stringify(transaction)
	});
	console.log(res.status);
};

export const deleteTransaction = async (id: string, token: string) => {
	console.log(id);
	const res = await fetch(`${server}/transaction?t=${token}&id=${id}`, {
		method: 'DELETE',
		mode: 'cors'
	});
	console.log(res.status);
};

export const getWeekNumber = (d: Date) => {
	d = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
	// Set to nearest Thursday: current date + 4 - current day number
	// Make Sunday's day number 7
	d.setUTCDate(d.getUTCDate() + 4 - (d.getUTCDay() || 7));
	// Get first day of year
	const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
	// Calculate full weeks to nearest Thursday
	const weekNo = Math.ceil(((d.valueOf() - yearStart.valueOf()) / 86400000 + 1) / 7);

	return weekNo;
};

export const fetchAndPrintHTML = async (
	url: string,
	selector: string,
	token: string
): Promise<string> => {
	const response = await fetch(
		`${nodeServer}/fetch-website?t=${token}&url=${encodeURIComponent(url)}&selector=${encodeURIComponent(selector)}`
	);

	const data = await response.json();
	return JSON.stringify(data);
};
