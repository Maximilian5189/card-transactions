<script lang="ts">
	import {
		getTransactionsTotal,
		getStartOfWeekTimestamp,
		getTransactions,
		postTransaction,
		deleteTransaction
	} from '$lib';
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
		position: PluginPosition.Header
	};

	async function createNewTransactionHandler() {
		let d;
		if (date) {
			d = new Date(date).valueOf();
		} else {
			d = new Date().valueOf();
		}

		const t = { name, amount, date: d };
		await postTransaction(t, token);

		data = await getTransactions(0, token);

		name = '';
		amount = 0;
	}

	async function deleteTransactionHandler() {
		await deleteTransaction(id, token);
		data = await getTransactions(0, token);
		id = '';
	}

	let totalLastWeek = 0;
	let weekMinusTwo = 0;
	let data: any[] = [];
	let name = '';
	let amount = 0;
	let id = '';
	let date = '';
	onMount(async () => {
		token = $page.url.searchParams.get('t') || '';

		data = await getTransactions(0, token);

		let t = getStartOfWeekTimestamp(-1);
		totalLastWeek = await getTransactionsTotal(t, token);

		t = getStartOfWeekTimestamp(-2);
		weekMinusTwo = await getTransactionsTotal(t, token);
	});
</script>

<h1>Welcome to Jochen</h1>

<input type="text" bind:value={name} placeholder="name" />
<input type="number" bind:value={amount} placeholder="amount" />
<input type="date" bind:value={date} />

<button on:click={createNewTransactionHandler}>Create new transaction</button>
<br /><br />
<input type="text" bind:value={id} placeholder="id" />
<button on:click={deleteTransactionHandler}>Delete transaction</button>

<div>total last week: {totalLastWeek}</div>
<div>week before: {weekMinusTwo}</div>

<Grid
	{columns}
	{data}
	sort
	search
	pagination={{ enabled: true, limit: 100 }}
	plugins={[sumPlugin]}
/>

{JSON.stringify(data)}

<style global>
	@import 'https://cdn.jsdelivr.net/npm/gridjs/dist/theme/mermaid.min.css';
</style>
