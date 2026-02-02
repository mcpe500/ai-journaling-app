<script lang="ts">
	import { pb, hashKey } from '$lib/pocketbase';
	import { initializeEncryption } from '$lib/encryption';
	import { goto } from '$app/navigation';

	let name = '';
	let email = '';
	let password = '';
	let confirmPassword = '';
	let encryptionPassword = '';
	let loading = false;
	let error = '';

	async function handleRegister() {
		loading = true;
		error = '';

		// Validation
		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			loading = false;
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			loading = false;
			return;
		}

		if (!encryptionPassword) {
			error = 'Encryption password is required for securing your journal';
			loading = false;
			return;
		}

		try {
			// Initialize encryption
			const { salt, key, keyHash } = initializeEncryption(encryptionPassword);

			// Create user account
			await pb.collection('users').create({
				name,
				email,
				password,
				passwordConfirm: confirmPassword,
				encryption_key_hash: keyHash
			});

			// Auto-login after registration
			await pb.collection('users').authWithPassword(email, password);

			// Store encryption key in memory (NEVER in localStorage)
			// TODO: Use a secure in-memory store or session storage with encryption
			sessionStorage.setItem('journal_encryption_key', key);
			sessionStorage.setItem('journal_encryption_salt', salt);

			goto('/dashboard');
		} catch (err: any) {
			error = err.message || 'Registration failed. Please try again.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Register - AI Journal</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center px-4 py-12">
	<div class="max-w-md w-full space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				Create your account
			</h2>
			<p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
				Start your journaling journey with AI-powered insights
			</p>
		</div>

		<form class="mt-8 space-y-6" on:submit|preventDefault={handleRegister}>
			<div class="rounded-md shadow-sm space-y-4">
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Name</label>
					<input
						id="name"
						name="name"
						type="text"
						required
						bind:value={name}
						placeholder="Your name"
						class="mt-1 appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
					/>
				</div>

				<div>
					<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Email address</label>
					<input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={email}
						placeholder="Email address"
						class="mt-1 appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
					<input
						id="password"
						name="password"
						type="password"
						autocomplete="new-password"
						required
						bind:value={password}
						placeholder="Password (min 8 characters)"
						minlength="8"
						class="mt-1 appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
					/>
				</div>

				<div>
					<label for="confirm-password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Confirm Password</label>
					<input
						id="confirm-password"
						name="confirm-password"
						type="password"
						autocomplete="new-password"
						required
						bind:value={confirmPassword}
						placeholder="Confirm password"
						class="mt-1 appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
					/>
				</div>

				<div class="border-t border-gray-200 dark:border-gray-700 pt-4">
					<label for="encryption-password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						🔐 Encryption Password
					</label>
					<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
						This password encrypts your journal entries. <strong>Cannot be recovered if lost.</strong>
					</p>
					<input
						id="encryption-password"
						name="encryption-password"
						type="password"
						required
						bind:value={encryptionPassword}
						placeholder="Create a strong encryption password"
						class="mt-2 appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
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
					{loading ? 'Creating account...' : 'Create account'}
				</button>
			</div>

			<div class="text-center">
				<p class="text-sm text-gray-600 dark:text-gray-400">
					Already have an account?
					<a href="/" class="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400">Sign in</a>
				</p>
			</div>
		</form>
	</div>
</div>
