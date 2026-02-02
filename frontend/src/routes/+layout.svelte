<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { assets } from '$app/paths';
	import { authStore, themeStore } from '$lib/stores';
	import { pb } from '$lib/pocketbase';

	// Theme classes
	const darkThemeClass = 'dark';

	// Layout children snippet (Svelte 5 runes mode)
	let { children }: { children: Snippet } = $props();

	// Reactive state for authentication (Svelte 5 runes mode)
	let isAuthenticated = $state(false);

	onMount(() => {
		// Initialize auth store
		authStore.set({
			isValid: pb.authStore.isValid,
			user: pb.authStore.model
		});

		// Initialize theme from localStorage or system preference
		if (typeof localStorage !== 'undefined') {
			const savedTheme = localStorage.getItem('theme');
			const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

			if (savedTheme) {
				themeStore.set(savedTheme as 'light' | 'dark');
			} else if (prefersDark) {
				themeStore.set('dark');
			}
		}
	});

	// Sync auth state with store using $effect for Svelte 5
	$effect(() => {
		const unsubscribe = authStore.subscribe((auth) => {
			isAuthenticated = auth.isValid;
		});
		return () => {
			unsubscribe();
		};
	});

	// Apply theme to document using $effect for Svelte 5
	$effect(() => {
		if (typeof document !== 'undefined') {
			const html = document.documentElement;
			const currentTheme = $themeStore;
			if (currentTheme === 'dark') {
				html.classList.add(darkThemeClass);
			} else {
				html.classList.remove(darkThemeClass);
			}
		}
	});
</script>

<svelte:head>
	<title>AI Journal</title>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="icon" href="{assets}/favicon.svg" />
</svelte:head>

{#if isAuthenticated}
	<nav class="bg-white dark:bg-gray-900 shadow-md">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center">
					<a href="/dashboard" class="text-xl font-bold text-gray-900 dark:text-white">AI Journal</a>
				</div>
				<div class="flex items-center space-x-4">
					<a href="/dashboard" class="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white">Dashboard</a>
					<a href="/journal" class="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white">Journal</a>
					<a href="/calendar" class="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white">Calendar</a>
					<a href="/growth" class="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white">Growth</a>
					<button
						on:click={() => pb.authStore.clear()}
						class="text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
					>
						Logout
					</button>
				</div>
			</div>
		</div>
	</nav>
{/if}

<main class="min-h-screen bg-gray-50 dark:bg-gray-900">
	{@render children()}
</main>

<footer class="bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 py-4">
	<div class="max-w-7xl mx-auto px-4 text-center text-gray-600 dark:text-gray-400">
		<p>&copy; 2026 AI Journal. Your private thoughts, AI-powered insights.</p>
	</div>
</footer>

<style>
	:global(html) {
		background-color: #fff;
	}
	:global(html.dark) {
		background-color: #111827;
	}
	:global(body) {
		@apply bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100;
	}
</style>
