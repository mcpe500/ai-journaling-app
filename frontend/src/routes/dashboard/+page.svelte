<script lang="ts">
	import { onMount } from 'svelte';
	import { pb } from '$lib/pocketbase';
	import { userStatsStore } from '$lib/stores';

	let stats = {
		totalEntries: 0,
		totalWords: 0,
		currentStreak: 0,
		longestStreak: 0,
		lastEntryDate: null as string | null,
		preferredAnalysisFrequency: 'weekly'
	};

	let loading = true;

	onMount(async () => {
		try {
			const user = pb.authStore.model;
			if (user) {
				stats = {
					totalEntries: user.totalEntries || 0,
					totalWords: user.totalWords || 0,
					currentStreak: user.currentStreak || 0,
					longestStreak: user.longestStreak || 0,
					lastEntryDate: user.lastEntryDate || null,
					preferredAnalysisFrequency: user.preferredAnalysisFrequency || 'weekly'
				};
				userStatsStore.set(stats);
			}
		} catch (error) {
			console.error('Failed to load user stats:', error);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Dashboard - AI Journal</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
		<p class="mt-2 text-gray-600 dark:text-gray-400">Welcome back! Here's your journaling overview.</p>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
			<p class="mt-4 text-gray-600 dark:text-gray-400">Loading your stats...</p>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<span class="text-3xl">📝</span>
					</div>
					<div class="ml-4">
						<p class="text-sm font-medium text-gray-600 dark:text-gray-400">Total Entries</p>
						<p class="text-2xl font-semibold text-gray-900 dark:text-white">{stats.totalEntries}</p>
					</div>
				</div>
			</div>

			<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<span class="text-3xl">📖</span>
					</div>
					<div class="ml-4">
						<p class="text-sm font-medium text-gray-600 dark:text-gray-400">Total Words</p>
						<p class="text-2xl font-semibold text-gray-900 dark:text-white">{stats.totalWords.toLocaleString()}</p>
					</div>
				</div>
			</div>

			<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<span class="text-3xl">🔥</span>
					</div>
					<div class="ml-4">
						<p class="text-sm font-medium text-gray-600 dark:text-gray-400">Current Streak</p>
						<p class="text-2xl font-semibold text-gray-900 dark:text-white">{stats.currentStreak} days</p>
					</div>
				</div>
			</div>

			<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<span class="text-3xl">🏆</span>
					</div>
					<div class="ml-4">
						<p class="text-sm font-medium text-gray-600 dark:text-gray-400">Longest Streak</p>
						<p class="text-2xl font-semibold text-gray-900 dark:text-white">{stats.longestStreak} days</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Quick Actions</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<a
					href="/journal/new"
					class="flex items-center justify-center px-4 py-3 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
				>
					<span class="mr-2">✍️</span>
					New Entry
				</a>
				<a
					href="/calendar"
					class="flex items-center justify-center px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600"
				>
					<span class="mr-2">📅</span>
					View Calendar
				</a>
				<a
					href="/heatmap"
					class="flex items-center justify-center px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600"
				>
					<span class="mr-2">🟩</span>
					Heatmap
				</a>
			</div>
		</div>

		<!-- Last Entry Date -->
		{#if stats.lastEntryDate}
			<div class="mt-6 bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
				<p class="text-sm text-blue-800 dark:text-blue-400">
					Last entry: <strong>{new Date(stats.lastEntryDate).toLocaleDateString()}</strong>
				</p>
			</div>
		{/if}
	{/if}
</div>
