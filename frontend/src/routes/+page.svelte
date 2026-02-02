<script lang="ts">
	import { pb } from '$lib/pocketbase';
	import { goto } from '$app/navigation';

	let email = '';
	let password = '';
	let loading = false;
	let error = '';

	async function handleLogin() {
		loading = true;
		error = '';

		try {
			await pb.collection('users').authWithPassword(email, password);
			goto('/dashboard');
		} catch (err: any) {
			error = err.message || 'Login failed. Please check your credentials.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Login - AI Journal</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="max-w-md w-full space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				Welcome to AI Journal
			</h2>
			<p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
				Your private thoughts, AI-powered insights
			</p>
		</div>

		<form class="mt-8 space-y-6" on:submit|preventDefault={handleLogin}>
			<div class="rounded-md shadow-sm space-y-4">
				<div>
					<label for="email" class="sr-only">Email address</label>
					<input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={email}
						placeholder="Email address"
						class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
					/>
				</div>
				<div>
					<label for="password" class="sr-only">Password</label>
					<input
						id="password"
						name="password"
						type="password"
						autocomplete="current-password"
						required
						bind:value={password}
						placeholder="Password"
						class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
					/>
				</div>
			</div>

			{#if error}
				<div class="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
					<p class="text-sm text-red-800 dark:text-red-400">{error}</p>
				</div>
			{/if}

			<div>
				<button
					type="submit"
					disabled={loading}
					class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{loading ? 'Signing in...' : 'Sign in'}
				</button>
			</div>

			<div class="text-center">
				<p class="text-sm text-gray-600 dark:text-gray-400">
					Don't have an account?
					<a href="/register" class="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400">Sign up</a>
				</p>
			</div>
		</form>
	</div>
</div>
