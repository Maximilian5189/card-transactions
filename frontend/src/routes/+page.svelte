<script lang="ts">
	import {
		getTransactionsTotal,
		getStartOfWeekTimestamp,
		getTransactions,
		postTransaction,
		deleteTransaction,
		type transaction,
		getWeekNumber
	} from '$lib';
	import Grid from 'gridjs-svelte';
	import 'gridjs/dist/theme/mermaid.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

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
		let t = getStartOfWeekTimestamp(weeksOffset);
		data = await getTransactions(t, token);
		let totalSpentCurrentLocal = 0;
		for (let transaction of data) {
			totalSpentCurrentLocal += transaction.amount;
		}
		totalSpentCurrent = Math.round(totalSpentCurrentLocal * 100) / 100;

		let totalsPastWeeksLocal = [];
		let totalSavedLocal = 0;
		for (let i = 0; i < pastWeeksToDisplay + weeksOffset; i++) {
			const offset = -1 - i + weeksOffset;
			const t = getStartOfWeekTimestamp(offset);
			const totalPastWeek = await getTransactionsTotal(t, token);
			totalsPastWeeksLocal.push(totalPastWeek);

			totalSavedLocal -= totalPastWeek - 1000;
		}
		totalSaved = Math.round(totalSavedLocal * 100) / 100;

		totalsPastWeeks = totalsPastWeeksLocal;
	}

	async function increaseWeekOffset() {
		if (!weeksOffset) return;
		weeksOffset += 1;
		await calculate();
	}

	async function decreaseWeekOffset() {
		weeksOffset -= 1;
		await calculate();
	}

	let token: string;
	let pastWeeksToDisplay = 0;
	let totalsPastWeeks: any[] = [];
	let data: transaction[] = [];
	let name = '';
	let amount = 0;
	let id = '';
	let date = '';
	let totalSaved = 0;
	let totalSpentCurrent = 0;
	let weeksOffset = 0;

	const currentWeek = getWeekNumber(new Date());
	pastWeeksToDisplay = currentWeek - 19;
	if (new Date().getFullYear() > 2024) {
		pastWeeksToDisplay = currentWeek;
	}

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

<button on:click={decreaseWeekOffset}>previous week</button>
<br />
<button on:click={increaseWeekOffset}>next week</button>

<br />
<div>week offset: {weeksOffset}</div>

<br />

<ul>
	{#each totalsPastWeeks as total, index}
		<li>
			<div>t{-index - 1}: {total}</div>
		</li>
	{/each}
</ul>

<div>
	total saved: <span style={`color: ${totalSaved < 0 ? 'red' : 'green'}`}>{totalSaved}</span>
</div>
<br />
<div>total spent this week: {totalSpentCurrent}</div>
<div>budget: {Math.round((1000 - totalSpentCurrent) * 100) / 100}</div>
<Grid {columns} {data} sort search pagination={{ enabled: true, limit: 100 }} />

{JSON.stringify(data)}

<style global>
	@import 'https://cdn.jsdelivr.net/npm/gridjs/dist/theme/mermaid.min.css';
</style>
