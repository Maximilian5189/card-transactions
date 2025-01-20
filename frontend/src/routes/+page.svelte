<script lang="ts">
	import {
		getTransactionsTotal,
		getStartOfWeekTimestamp,
		getTransactions,
		postTransaction,
		deleteTransaction,
		type transaction,
		getWeekNumber,
		fetchAndPrintHTML
	} from '$lib';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	// Table state
	let searchQuery = '';
	let sortColumn = 'date';
	let sortDirection: 'asc' | 'desc' = 'desc';
	let showDeleteModal = false;
	let transactionToDelete: { id: string; name: string } | null = null;

	function isValidTransactionId(id: string | undefined): boolean {
		return !isNaN(Number(id)) && id?.trim() !== '';
	}

	function formatDate(timestamp: number) {
		return new Date(timestamp * 1000).toLocaleDateString('en-us', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
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
	let bigSnowPricing = '';
	let isLoadingPricing = false;
	const currBudget = 1000;

	const currentWeek = getWeekNumber(new Date());
	pastWeeksToDisplay = currentWeek - 1;

	$: filteredData = data
		.filter((item) =>
			Object.values(item).some((val) =>
				String(val).toLowerCase().includes(searchQuery.toLowerCase())
			)
		)
		.sort((a, b) => {
			const aVal = sortColumn === 'date' ? a[sortColumn] : String(a[sortColumn]).toLowerCase();
			const bVal = sortColumn === 'date' ? b[sortColumn] : String(b[sortColumn]).toLowerCase();

			if (sortDirection === 'asc') {
				return aVal > bVal ? 1 : -1;
			}
			return aVal < bVal ? 1 : -1;
		});

	function handleSort(column: string) {
		if (sortColumn === column) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortColumn = column;
			sortDirection = 'asc';
		}
	}

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

	async function deleteTransactionHandler(transactionId: string) {
		await deleteTransaction(transactionId, token);
		await calculate();
		closeDeleteModal();
	}

	function openDeleteModal(transaction: { id: string; name: string }) {
		transactionToDelete = transaction;
		showDeleteModal = true;
	}

	function closeDeleteModal() {
		showDeleteModal = false;
		transactionToDelete = null;
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
		for (let i = 0; i < pastWeeksToDisplay - weeksOffset; i++) {
			const offset = i + 1 + weeksOffset;
			const t = getStartOfWeekTimestamp(offset);
			const totalPastWeek = await getTransactionsTotal(t, token);

			let budget = 1000;
			// timestamp for the week since I have new budget
			if (t >= 19975 * 24 * 60 * 60 * 1000) {
				budget = currBudget;
			}
			// comment in to find current day for a new if condition
			// console.log(t / 24 / 60 / 60 / 1000);

			if (i < 10) {
				totalsPastWeeksLocal.push([totalPastWeek, budget]);
			}

			totalSavedLocal -= totalPastWeek - budget;
		}
		totalSaved = Math.round(totalSavedLocal * 100) / 100;

		totalsPastWeeks = totalsPastWeeksLocal;
	}

	async function decreaseWeekOffset() {
		weeksOffset -= 1;
		if (weeksOffset === 0) {
			pastWeeksToDisplay = currentWeek - 1;
		}
		await calculate();
	}

	async function increaseWeekOffset() {
		weeksOffset += 1;
		if (pastWeeksToDisplay === 0) {
			if (new Date().getFullYear() > 2025) {
				pastWeeksToDisplay = 52;
			} else {
				pastWeeksToDisplay = 52 - 19 + 1;
			}
		}
		await calculate();
	}

	onMount(async () => {
		token = $page.url.searchParams.get('t') || '';

		await calculate();

		try {
			isLoadingPricing = true;
			bigSnowPricing = await fetchAndPrintHTML(
				'https://bigsnowad.snowcloud.shop/shop/page/1E7B1BEE-0982-4F86-0F80-FC2A96F03E19',
				token
			);
		} catch (error) {
			bigSnowPricing = 'le error';
		} finally {
			isLoadingPricing = false;
		}
	});
</script>

<h1>Welcome to Jochen</h1>

<button class="nav-btn" on:click={() => goto(`/old?${$page.url.searchParams.toString()}`)}
	>Switch to Old View</button
>

<div class="form-group">
	<input type="text" class="input-field" bind:value={name} placeholder="name" />
	<input type="number" class="input-field" bind:value={amount} placeholder="amount" />
	<input type="date" class="input-field" bind:value={date} />
	<button class="primary-btn" on:click={createNewTransactionHandler}>Create new transaction</button>
</div>

<div class="form-group">
	<input
		type="number"
		class="input-field"
		bind:value={pastWeeksToDisplay}
		placeholder="past weeks to display"
	/>
	<button class="primary-btn" on:click={calculate}>Set past weeks</button>
</div>

<div class="form-group">
	<button class="secondary-btn" on:click={increaseWeekOffset}>Previous week</button>
	<button class="secondary-btn" on:click={decreaseWeekOffset}>Next week</button>
</div>

<div>week offset: {weeksOffset}</div>

<br />

<ul>
	{#each totalsPastWeeks as total, index}
		<li>
			<div>t{-index - 1}: {total[0]}, budget: {total[1]}</div>
		</li>
	{/each}
</ul>

<div>
	total saved: <span
		style={`color: ${totalSaved < 0 ? 'var(--error-color)' : 'var(--success-color)'}`}
		>{totalSaved}</span
	>
</div>
<br />
<div>total spent this week: {totalSpentCurrent}</div>
<div>budget: {Math.round((currBudget - totalSpentCurrent) * 100) / 100}</div>

<div class="html-content">
	<div>big snow:</div>
	{#if isLoadingPricing}
		<p>Loading...</p>
	{:else}
		{bigSnowPricing}
	{/if}
</div>

<div class="search-container">
	<input type="text" class="input-field" bind:value={searchQuery} placeholder="Search..." />
</div>

<div class="table-container">
	<table>
		<thead>
			<tr>
				<th on:click={() => handleSort('name')}>
					Name
					{#if sortColumn === 'name'}
						<span class="sort-indicator">{sortDirection === 'asc' ? '↑' : '↓'}</span>
					{/if}
				</th>
				<th on:click={() => handleSort('amount')}>
					Amount
					{#if sortColumn === 'amount'}
						<span class="sort-indicator">{sortDirection === 'asc' ? '↑' : '↓'}</span>
					{/if}
				</th>
				<th on:click={() => handleSort('date')}>
					Date
					{#if sortColumn === 'date'}
						<span class="sort-indicator">{sortDirection === 'asc' ? '↑' : '↓'}</span>
					{/if}
				</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each filteredData as row}
				<tr>
					<td>{row.name}</td>
					<td>{row.amount}</td>
					<td>{formatDate(row.date)}</td>
					<td>
						{#if isValidTransactionId(row.messageID)}
							<button
								class="delete-btn"
								on:click={() => openDeleteModal({ id: row.id, name: row.name })}
								title="Delete transaction"
							>
								Delete
							</button>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

{#if showDeleteModal && transactionToDelete}
	<div class="modal-backdrop" on:click={closeDeleteModal}>
		<div class="modal" on:click|stopPropagation>
			<h2>Confirm Delete</h2>
			<p>Are you sure you want to delete the transaction "{transactionToDelete.name}"?</p>
			<div class="modal-actions">
				<button class="cancel-btn" on:click={closeDeleteModal}>Cancel</button>
				<button
					class="delete-btn"
					on:click={() => deleteTransactionHandler(transactionToDelete.id)}
				>
					Delete
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:root {
		--primary-color: #3b82f6;
		--primary-hover: #2563eb;
		--secondary-color: #6b7280;
		--secondary-hover: #4b5563;
		--error-color: #ef4444;
		--error-hover: #dc2626;
		--success-color: #22c55e;
		--border-color: #e2e8f0;
		--background-color: white;
		--text-color: #374151;
	}

	@media (prefers-color-scheme: dark) {
		:root {
			--primary-color: #2563eb;
			--primary-hover: #1d4ed8;
			--secondary-color: #4b5563;
			--secondary-hover: #374151;
			--error-color: #dc2626;
			--error-hover: #b91c1c;
			--success-color: #16a34a;
			--border-color: #374151;
			--background-color: #1a1a1a;
			--text-color: #e5e7eb;
		}
	}

	.form-group {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
		flex-wrap: wrap;
		align-items: center;
	}

	.input-field {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 0.25rem;
		font-size: 0.875rem;
		background-color: var(--background-color);
		color: var(--text-color);
		min-width: 150px;
	}

	.input-field:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 1px var(--primary-color);
	}

	.primary-btn,
	.nav-btn {
		background-color: var(--primary-color);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
		transition: background-color 0.2s;
	}

	.primary-btn:hover,
	.nav-btn:hover {
		background-color: var(--primary-hover);
	}

	.secondary-btn {
		background-color: var(--secondary-color);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
		transition: background-color 0.2s;
	}

	.secondary-btn:hover {
		background-color: var(--secondary-hover);
	}

	.delete-btn {
		background-color: var(--error-color);
		color: white;
		border: none;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
		transition: background-color 0.2s;
	}

	.delete-btn:hover {
		background-color: var(--error-hover);
	}

	.table-container {
		margin: 1rem 0;
		border: 1px solid var(--border-color);
		border-radius: 0.5rem;
		overflow: hidden;
	}

	@media (prefers-color-scheme: dark) {
		:global(body) {
			background-color: #1a1a1a;
			color: #e5e7eb;
		}

		.table-container {
			border-color: #374151;
		}

		input[type='text'],
		input[type='number'],
		input[type='date'] {
			background-color: #2d2d2d;
			border: 1px solid #4b5563;
			color: #e5e7eb;
		}

		input::placeholder {
			color: #9ca3af;
		}

		.search-input {
			background-color: #2d2d2d;
			border-color: #4b5563;
			color: #e5e7eb;
		}
	}

	.search-container {
		margin-bottom: 1rem;
	}

	.search-input {
		padding: 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 0.25rem;
		width: 100%;
		max-width: 300px;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		background-color: var(--background-color);
	}

	th {
		background-color: var(--secondary-color);
		color: white;
		font-weight: 600;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	th:hover {
		background-color: var(--secondary-hover);
	}

	@media (prefers-color-scheme: dark) {
		table {
			background-color: #2d2d2d;
		}

		td {
			border-color: #374151;
		}

		tr:hover {
			background-color: #374151;
		}
	}

	th,
	td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid var(--border-color);
	}

	.sort-indicator {
		display: inline-block;
		margin-left: 0.5rem;
	}

	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-color: rgba(0, 0, 0, 0.5);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
	}

	.modal {
		background-color: var(--background-color);
		color: var(--text-color);
		padding: 1.5rem;
		border-radius: 0.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
		max-width: 400px;
		width: 90%;
	}

	.modal h2 {
		margin: 0 0 1rem 0;
		font-size: 1.25rem;
		color: var(--text-color);
	}

	.modal p {
		margin: 0 0 1.5rem 0;
		color: var(--text-color);
		opacity: 0.8;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
	}

	@media (prefers-color-scheme: dark) {
		.modal {
			box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
		}

		.modal-backdrop {
			background-color: rgba(0, 0, 0, 0.7);
		}
	}

	.cancel-btn {
		background-color: var(--secondary-color);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
		transition: background-color 0.2s;
	}

	.cancel-btn:hover {
		background-color: var(--secondary-hover);
	}

	.html-content {
		margin-top: 20px;
		padding: 15px;
		background-color: #f3f4f6;
		border-radius: 4px;
		overflow-x: auto;
		border: 1px solid var(--border-color);
	}

	.html-content pre {
		white-space: pre-wrap;
		word-wrap: break-word;
		font-family: monospace;
		margin: 0;
		color: var(--text-color);
	}

	@media (prefers-color-scheme: dark) {
		.html-content {
			background-color: #2d2d2d;
			border-color: #374151;
		}

		.html-content pre {
			color: #e5e7eb;
		}
	}
</style>
