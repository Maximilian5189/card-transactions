<script lang="ts">
	import {
		getTransactionsTotal,
		getStartOfWeekTimestamp,
		getTransactions,
		postTransaction,
		deleteTransaction,
		type transaction
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

		await calculate();

		name = '';
		amount = 0;
	}

	async function deleteTransactionHandler() {
		await deleteTransaction(id, token);
		await calculate();
		id = '';
	}

	async function calculate() {
		let t = getStartOfWeekTimestamp();
		data = await getTransactions(t, token);
		totalSpentCurrent = 0;
		for (let transaction of data) {
			totalSpentCurrent += transaction.amount;
		}
		totalSpentCurrent = Math.round(totalSpentCurrent * 100) / 100;

		totalsPastWeeks = [];
		totalSaved = 0;
		for (let i = 0; i < pastWeeksToDisplay; i++) {
			const offset = -1 - i;
			const t = getStartOfWeekTimestamp(offset);
			const totalPastWeek = await getTransactionsTotal(t, token);
			totalsPastWeeks.push(totalPastWeek);

			totalSaved -= totalPastWeek - 1000;
		}
		totalSaved = Math.round(totalSaved * 100) / 100;

		totalsPastWeeks = totalsPastWeeks;
	}

	let totalsPastWeeks: any[] = [];
	let pastWeeksToDisplay = 2;
	let data: transaction[] = [];
	let name = '';
	let amount = 0;
	let id = '';
	let date = '';
	let totalSaved = 0;
	let totalSpentCurrent = 0;
	onMount(async () => {
		token = $page.url.searchParams.get('t') || '';

		await calculate();
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
<br /><br />
<input type="number" bind:value={pastWeeksToDisplay} placeholder="past weeks to display" />
<button on:click={calculate}>set past weeks</button>

<br /><br />

<ul>
	{#each totalsPastWeeks as total, index}
		<li>
			<div>t{-index - 1}: {total}</div>
		</li>
	{/each}
</ul>

<div>total saved: {totalSaved}</div>
<br />
<div>total spent this week: {totalSpentCurrent}</div>
<div>budget: {1000 - totalSpentCurrent}</div>
<Grid {columns} {data} sort search pagination={{ enabled: true, limit: 100 }} />

{JSON.stringify(data)}

<style global>
	@import 'https://cdn.jsdelivr.net/npm/gridjs/dist/theme/mermaid.min.css';
</style>
