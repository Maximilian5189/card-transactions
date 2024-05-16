<script lang="ts">
	import { getTransactionsTotal, type transaction, getStartOfWeekTimestamp } from '$lib';
	import { useSelector, PluginPosition } from 'gridjs';
	import Grid from 'gridjs-svelte';
	import 'gridjs/dist/theme/mermaid.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let token: string;

	const columns = [
		{ name: 'name', sort: false },
		{ name: 'amount', sort: false },
		{
			name: 'date',
			formatter: (cell: number) => {
				return new Date(cell * 1000).toLocaleDateString('en-us', {
					year: 'numeric',
					month: 'short',
					day: 'numeric'
				});
			},
			sort: true
		}
	];

	const sumPlugin = {
		id: 'salaryplugin',
		component: function TotalSalaryPlugin() {
			const data = useSelector((state) => state.data);

			if (!data) return;

			let total = 0;
			for (const row of data.toArray()) {
				total += Number(row[1]);
			}

			return `Total: ${Math.round(total * 100) / 100}`;
		},
		position: PluginPosition.Footer
	};

	const getTransactions = (transactions: transaction[]) => {
		return transactions.map((transaction) => {
			return [transaction.name, transaction.amount, transaction.date];
		});
	};

	let totalLastWeek = 0;
	onMount(async () => {
		token = $page.url.searchParams.get('t') || '';

		const t = getStartOfWeekTimestamp(-1);
		totalLastWeek = await getTransactionsTotal(t, token);
	});
</script>

<h1>Welcome to Jochen</h1>
<Grid
	{columns}
	sort
	search
	pagination={{ enabled: true, limit: 100 }}
	server={{
		url: `http://localhost:8080/transactions?t=${token}`,
		then: getTransactions
	}}
	plugins={[sumPlugin]}
/>

<div>total last week: {totalLastWeek}</div>

<style global>
	@import 'https://cdn.jsdelivr.net/npm/gridjs/dist/theme/mermaid.min.css';
</style>
