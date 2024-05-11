<script>
	const getTransactions = async () => {
		const options = {
			method: 'GET',
			// headers: {
			// 	'Content-Type': 'application/json'
			// },
			mode: 'cors'
		};
		let res = await fetch('http://localhost:8080/transactions', options);
		console.log(await res.json());
	};
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>
<button on:click={getTransactions}>klick</button>
