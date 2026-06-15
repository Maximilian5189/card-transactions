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

	function fmt(n: number) {
		return (n ?? 0).toLocaleString('en-US', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
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
	let patagoniaNanoPuffPricing = '';
	let isLoadingPricing = false;
	const currBudget = 1100;

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

	// Display-only derived values
	$: remainingBudget = Math.round((currBudget - totalSpentCurrent) * 100) / 100;
	$: budgetUsedPct = currBudget > 0 ? Math.max(0, Math.round((totalSpentCurrent / currBudget) * 100)) : 0;
	$: budgetFillPct = Math.min(100, budgetUsedPct);
	$: viewLabel =
		weeksOffset === 0 ? 'This week' : `${weeksOffset} week${weeksOffset === 1 ? '' : 's'} ago`;

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
			d = new Date(date + 'T00:00:00').valueOf();
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

	async function reverseTransactionHandler(transactionToReverse: transaction) {
		const reversedTransaction = {
			name: `Reversal: ${transactionToReverse.name}`,
			amount: -transactionToReverse.amount,
			date: transactionToReverse.date * 1000
		};
		await postTransaction(reversedTransaction, token);
		await calculate();
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
			// 20458 is days, this number should be incremented +7 if budget is changed
			if (t >= 20458 * 24 * 60 * 60 * 1000) {
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

			const [patagoniaNanoPuffResult, bigSnowResult] = await Promise.all([
				fetchAndPrintHTML('patagonia-nano-puff', token)
				// fetchAndPrintHTML('bigsnow', token)
			]);

			patagoniaNanoPuffPricing = patagoniaNanoPuffResult;
			// bigSnowPricing = bigSnowResult;
		} catch (error) {
			console.log(error);
		} finally {
			isLoadingPricing = false;
		}
	});
</script>

<div class="page">
	<header class="app-header">
		<div class="brand">
			<span class="brand-mark">J</span>
			<div>
				<h1>Jochen</h1>
				<p class="subtitle">Weekly spending &amp; budget</p>
			</div>
		</div>

		<div class="week-nav">
			<button class="nav-btn" on:click={increaseWeekOffset} title="Go to previous week">
				‹ Prev
			</button>
			<div class="week-pill">
				<span class="week-pill-label">{viewLabel}</span>
				<span class="week-pill-sub">offset {weeksOffset}</span>
			</div>
			<button
				class="nav-btn"
				on:click={decreaseWeekOffset}
				disabled={weeksOffset === 0}
				title="Go to next week"
			>
				Next ›
			</button>
		</div>
	</header>

	<!-- Summary stats -->
	<section class="stats-grid">
		<div class="stat-card">
			<div class="stat-label">Spent · {viewLabel}</div>
			<div class="stat-value">{fmt(totalSpentCurrent)}</div>
			<div class="progress-track">
				<div
					class="progress-fill"
					class:over={totalSpentCurrent > currBudget}
					style={`width:${budgetFillPct}%`}
				></div>
			</div>
			<div class="stat-foot">
				{budgetUsedPct}% of {fmt(currBudget)} budget
				{#if totalSpentCurrent > currBudget}
					· <span class="over-text">{fmt(totalSpentCurrent - currBudget)} over</span>
				{/if}
			</div>
		</div>

		<div class="stat-card">
			<div class="stat-label">Remaining budget</div>
			<div class="stat-value" class:neg={remainingBudget < 0} class:pos={remainingBudget >= 0}>
				{fmt(remainingBudget)}
			</div>
			<div class="stat-foot">
				{remainingBudget >= 0 ? 'left to spend this week' : 'over budget this week'}
			</div>
		</div>

		<div class="stat-card">
			<div class="stat-label">Total saved</div>
			<div class="stat-value" class:neg={totalSaved < 0} class:pos={totalSaved >= 0}>
				{fmt(totalSaved)}
			</div>
			<div class="stat-foot">cumulative vs. budget</div>
		</div>
	</section>

	<div class="columns">
		<div class="col-main">
			<!-- Add transaction -->
			<section class="card">
				<h2 class="card-title">Add transaction</h2>
				<div class="form-grid">
					<label class="field">
						<span>Name</span>
						<input type="text" class="input-field" bind:value={name} placeholder="e.g. Groceries" />
					</label>
					<label class="field">
						<span>Amount</span>
						<input type="number" class="input-field" bind:value={amount} placeholder="0.00" />
					</label>
					<label class="field">
						<span>Date</span>
						<input type="date" class="input-field" bind:value={date} />
					</label>
					<button class="primary-btn add-btn" on:click={createNewTransactionHandler}>
						+ Add
					</button>
				</div>
			</section>

			<!-- Transactions table -->
			<section class="card">
				<div class="card-head">
					<h2 class="card-title">Transactions</h2>
					<input
						type="text"
						class="input-field search"
						bind:value={searchQuery}
						placeholder="Search transactions…"
					/>
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
								<th class="num" on:click={() => handleSort('amount')}>
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
								<th class="actions-col">Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each filteredData as row}
								<tr>
									<td>{row.name}</td>
									<td class="num" class:negative-amount={row.amount < 0}>{fmt(row.amount)}</td>
									<td>{formatDate(row.date)}</td>
									<td class="actions-col">
										{#if isValidTransactionId(row.messageID)}
											<button
												class="delete-btn"
												on:click={() => openDeleteModal({ id: row.id, name: row.name })}
												title="Delete transaction"
											>
												Delete
											</button>
										{:else}
											<button
												class="secondary-btn"
												on:click={() => reverseTransactionHandler(row)}
												title="Reverse transaction"
											>
												Reverse
											</button>
										{/if}
									</td>
								</tr>
							{:else}
								<tr>
									<td colspan="4" class="empty-state">No transactions for this week.</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		</div>

		<aside class="col-side">
			<!-- Past weeks -->
			<section class="card">
				<div class="card-head">
					<h2 class="card-title">Past weeks</h2>
					<div class="weeks-control">
						<input
							type="number"
							class="input-field weeks-input"
							bind:value={pastWeeksToDisplay}
							placeholder="weeks"
						/>
						<button class="secondary-btn" on:click={calculate}>Set</button>
					</div>
				</div>

				<ul class="weeks-list">
					{#each totalsPastWeeks as total, index}
						{@const spent = total[0]}
						{@const budget = total[1]}
						{@const saved = Math.round((budget - spent) * 100) / 100}
						{@const pct = Math.min(100, Math.max(0, (spent / budget) * 100))}
						<li class="week-row">
							<div class="week-row-top">
								<span class="week-tag">t−{index + 1}</span>
								<span class="week-amount">{fmt(spent)} <small>/ {fmt(budget)}</small></span>
							</div>
							<div class="progress-track sm">
								<div class="progress-fill" class:over={spent > budget} style={`width:${pct}%`}></div>
							</div>
							<div class="week-saved" class:neg={saved < 0} class:pos={saved >= 0}>
								{saved >= 0 ? '+' : ''}{fmt(saved)}
							</div>
						</li>
					{:else}
						<li class="empty-state">No past weeks to show.</li>
					{/each}
				</ul>
			</section>

			<!-- Price watch -->
			<section class="card">
				<h2 class="card-title">Price watch</h2>
				<div class="pricing-item">
					<div class="pricing-name">Patagonia Nano Puff</div>
					{#if isLoadingPricing}
						<div class="pricing-loading">Loading…</div>
					{:else}
						<div class="pricing-value">{patagoniaNanoPuffPricing}</div>
					{/if}
				</div>
			</section>
		</aside>
	</div>
</div>

{#if showDeleteModal && transactionToDelete}
	<div class="modal-backdrop" on:click={closeDeleteModal}>
		<div class="modal" on:click|stopPropagation>
			<h2>Confirm Delete</h2>
			<p>Are you sure you want to delete the transaction "{transactionToDelete.name}"?</p>
			<div class="modal-actions">
				<button class="cancel-btn" on:click={closeDeleteModal}>Cancel</button>
				<button class="delete-btn" on:click={() => deleteTransactionHandler(transactionToDelete.id)}>
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
		--success-color: #16a34a;
		--border-color: #e5e7eb;
		--background-color: #ffffff;
		--surface-color: #ffffff;
		--page-bg: #f4f6fb;
		--text-color: #1f2937;
		--text-muted: #6b7280;
		--shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 4px 12px rgba(16, 24, 40, 0.06);
		--radius: 14px;
	}

	@media (prefers-color-scheme: dark) {
		:root {
			--primary-color: #3b82f6;
			--primary-hover: #60a5fa;
			--secondary-color: #4b5563;
			--secondary-hover: #6b7280;
			--error-color: #ef4444;
			--error-hover: #f87171;
			--success-color: #22c55e;
			--border-color: #2b3344;
			--background-color: #161a23;
			--surface-color: #1b2030;
			--page-bg: #0e1117;
			--text-color: #e5e7eb;
			--text-muted: #9aa4b2;
			--shadow: 0 1px 2px rgba(0, 0, 0, 0.3), 0 8px 24px rgba(0, 0, 0, 0.35);
		}
	}

	:global(body) {
		margin: 0;
		background-color: var(--page-bg);
		color: var(--text-color);
		font-family:
			'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
		-webkit-font-smoothing: antialiased;
	}

	.page {
		max-width: 1080px;
		margin: 0 auto;
		padding: 1.5rem 1.25rem 4rem;
	}

	/* Header */
	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		flex-wrap: wrap;
		margin-bottom: 1.5rem;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.85rem;
	}

	.brand-mark {
		display: grid;
		place-items: center;
		width: 44px;
		height: 44px;
		border-radius: 12px;
		background: linear-gradient(135deg, var(--primary-color), #8b5cf6);
		color: #fff;
		font-weight: 700;
		font-size: 1.25rem;
		box-shadow: var(--shadow);
	}

	.brand h1 {
		margin: 0;
		font-size: 1.4rem;
		font-weight: 700;
		letter-spacing: -0.01em;
	}

	.subtitle {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.week-nav {
		display: flex;
		align-items: stretch;
		gap: 0.5rem;
	}

	.week-pill {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-width: 120px;
		padding: 0.25rem 0.9rem;
		background: var(--surface-color);
		border: 1px solid var(--border-color);
		border-radius: 10px;
		box-shadow: var(--shadow);
	}

	.week-pill-label {
		font-weight: 600;
		font-size: 0.85rem;
	}

	.week-pill-sub {
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	/* Stats */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		margin-bottom: 1.25rem;
	}

	.stat-card {
		background: var(--surface-color);
		border: 1px solid var(--border-color);
		border-radius: var(--radius);
		padding: 1.1rem 1.2rem;
		box-shadow: var(--shadow);
	}

	.stat-label {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--text-muted);
		font-weight: 600;
	}

	.stat-value {
		font-size: 1.85rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		margin: 0.3rem 0 0.5rem;
	}

	.stat-value.pos {
		color: var(--success-color);
	}

	.stat-value.neg {
		color: var(--error-color);
	}

	.stat-foot {
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	.progress-track {
		height: 8px;
		background: var(--border-color);
		border-radius: 999px;
		overflow: hidden;
		margin: 0.35rem 0 0.5rem;
	}

	.progress-track.sm {
		height: 6px;
		margin: 0.3rem 0;
	}

	.progress-fill {
		height: 100%;
		background: var(--primary-color);
		border-radius: 999px;
		transition: width 0.35s ease;
	}

	.progress-fill.over {
		background: var(--error-color);
	}

	.over-text {
		color: var(--error-color);
		font-weight: 600;
	}

	/* Layout columns */
	.columns {
		display: grid;
		grid-template-columns: 1.6fr 1fr;
		gap: 1rem;
		align-items: start;
	}

	.col-main,
	.col-side {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	/* Cards */
	.card {
		background: var(--surface-color);
		border: 1px solid var(--border-color);
		border-radius: var(--radius);
		padding: 1.2rem;
		box-shadow: var(--shadow);
	}

	.card-title {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
	}

	.card-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		margin-bottom: 1rem;
		flex-wrap: wrap;
	}

	/* Forms */
	.form-grid {
		display: grid;
		grid-template-columns: 1.4fr 1fr 1fr auto;
		gap: 0.75rem;
		align-items: end;
		margin-top: 1rem;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		font-size: 0.75rem;
		color: var(--text-muted);
		font-weight: 600;
	}

	.input-field {
		padding: 0.55rem 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 9px;
		font-size: 0.9rem;
		background-color: var(--background-color);
		color: var(--text-color);
		width: 100%;
		box-sizing: border-box;
		transition:
			border-color 0.15s,
			box-shadow 0.15s;
	}

	.input-field:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
	}

	.search {
		max-width: 260px;
	}

	.weeks-control {
		display: flex;
		gap: 0.5rem;
	}

	.weeks-input {
		width: 80px;
	}

	/* Buttons */
	.primary-btn {
		background-color: var(--primary-color);
		color: white;
		border: none;
		padding: 0.55rem 1.1rem;
		border-radius: 9px;
		cursor: pointer;
		font-size: 0.9rem;
		font-weight: 600;
		transition:
			background-color 0.15s,
			transform 0.05s;
	}

	.primary-btn:hover {
		background-color: var(--primary-hover);
	}

	.primary-btn:active {
		transform: translateY(1px);
	}

	.add-btn {
		height: fit-content;
	}

	.secondary-btn {
		background-color: transparent;
		color: var(--text-color);
		border: 1px solid var(--border-color);
		padding: 0.5rem 0.9rem;
		border-radius: 9px;
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 600;
		transition: background-color 0.15s;
	}

	.secondary-btn:hover {
		background-color: var(--page-bg);
		border-color: var(--secondary-hover);
	}

	.nav-btn {
		background-color: var(--surface-color);
		color: var(--text-color);
		border: 1px solid var(--border-color);
		padding: 0 0.85rem;
		border-radius: 10px;
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 600;
		box-shadow: var(--shadow);
		transition: background-color 0.15s;
	}

	.nav-btn:hover:not(:disabled) {
		background-color: var(--page-bg);
	}

	.nav-btn:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}

	.delete-btn {
		background-color: var(--error-color);
		color: white;
		border: none;
		padding: 0.35rem 0.8rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 0.82rem;
		font-weight: 600;
		transition: background-color 0.15s;
	}

	.delete-btn:hover {
		background-color: var(--error-hover);
	}

	/* Table */
	.table-container {
		border: 1px solid var(--border-color);
		border-radius: 10px;
		overflow: hidden;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		background-color: var(--surface-color);
		font-size: 0.9rem;
	}

	th {
		background-color: var(--page-bg);
		color: var(--text-muted);
		font-weight: 600;
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.03em;
		cursor: pointer;
		user-select: none;
		transition: color 0.15s;
	}

	th:hover {
		color: var(--text-color);
	}

	th,
	td {
		padding: 0.7rem 0.9rem;
		text-align: left;
		border-bottom: 1px solid var(--border-color);
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	tbody tr:hover {
		background-color: var(--page-bg);
	}

	.num {
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.negative-amount {
		color: var(--success-color);
	}

	.actions-col {
		text-align: right;
		width: 1%;
		white-space: nowrap;
	}

	.sort-indicator {
		display: inline-block;
		margin-left: 0.25rem;
	}

	.empty-state {
		text-align: center;
		color: var(--text-muted);
		padding: 1.5rem;
		font-size: 0.9rem;
	}

	/* Past weeks list */
	.weeks-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}

	.week-row {
		display: grid;
		grid-template-columns: 1fr;
		gap: 0.15rem;
	}

	.week-row-top {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
	}

	.week-tag {
		font-size: 0.72rem;
		font-weight: 700;
		color: var(--text-muted);
		background: var(--page-bg);
		padding: 0.1rem 0.45rem;
		border-radius: 6px;
	}

	.week-amount {
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		font-size: 0.9rem;
	}

	.week-amount small {
		color: var(--text-muted);
		font-weight: 400;
	}

	.week-saved {
		font-size: 0.78rem;
		font-weight: 600;
		text-align: right;
	}

	.week-saved.pos {
		color: var(--success-color);
	}

	.week-saved.neg {
		color: var(--error-color);
	}

	/* Pricing */
	.pricing-item {
		margin-top: 0.85rem;
		padding: 0.9rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 10px;
		background: var(--page-bg);
	}

	.pricing-name {
		font-weight: 600;
		font-size: 0.9rem;
		margin-bottom: 0.35rem;
	}

	.pricing-value {
		font-size: 0.82rem;
		color: var(--text-muted);
		word-break: break-word;
	}

	.pricing-loading {
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	/* Modal */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background-color: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(2px);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
	}

	.modal {
		background-color: var(--surface-color);
		color: var(--text-color);
		padding: 1.5rem;
		border-radius: var(--radius);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
		max-width: 400px;
		width: 90%;
	}

	.modal h2 {
		margin: 0 0 0.75rem 0;
		font-size: 1.2rem;
	}

	.modal p {
		margin: 0 0 1.5rem 0;
		color: var(--text-muted);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
	}

	.cancel-btn {
		background-color: transparent;
		color: var(--text-color);
		border: 1px solid var(--border-color);
		padding: 0.5rem 1rem;
		border-radius: 9px;
		cursor: pointer;
		font-size: 0.9rem;
		font-weight: 600;
		transition: background-color 0.15s;
	}

	.cancel-btn:hover {
		background-color: var(--page-bg);
	}

	/* Responsive */
	@media (max-width: 880px) {
		.columns {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 720px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.form-grid {
			grid-template-columns: 1fr 1fr;
		}

		.add-btn {
			grid-column: 1 / -1;
		}

		.app-header {
			flex-direction: column;
			align-items: flex-start;
		}

		.search {
			max-width: 100%;
			flex: 1;
		}
	}
</style>
