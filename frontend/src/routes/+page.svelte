<script lang="ts">
	import type { transaction } from '$lib';
	import { useSelector, PluginPosition } from 'gridjs';
	import Grid from 'gridjs-svelte';
	import 'gridjs/dist/theme/mermaid.css';

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
				console.log(row[1]);
				total += Number(row[1]);
			}

			return `Total: ${Math.round(total * 100) / 100}`;
		},
		position: PluginPosition.Footer
	};

	let total = 0;
	const getTransactions = (transactions: transaction[]) => {
		transactions.forEach((transaction: transaction) => {
			total += Number(transaction.amount);
		});
		transactions.sort((a, b) => b.date - a.date);
		total = Math.round(total * 100) / 100;
		return transactions.map((transaction) => {
			return [transaction.name, transaction.amount, transaction.date];
		});
	};
</script>

<h1>Welcome to Jochen</h1>

<div>total: {total}</div>

<Grid
	{columns}
	sort
	search
	pagination={{ enabled: true, limit: 100 }}
	server={{
		url: 'http://localhost:8080/transactions',
		then: getTransactions
	}}
	plugins={[sumPlugin]}
/>

<style global>
	@import 'https://cdn.jsdelivr.net/npm/gridjs/dist/theme/mermaid.min.css';
</style>
