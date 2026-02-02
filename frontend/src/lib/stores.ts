import { writable, derived } from 'svelte/store';
import { pb, isAuthenticated, getCurrentUser } from '$lib/pocketbase';
import type { AuthRecord } from 'pocketbase';

// Auth store
export const authStore = writable({
	isValid: isAuthenticated(),
	user: getCurrentUser() as AuthRecord | null
});

// Update auth store when PocketBase auth changes
pb.authStore.onChange(() => {
	authStore.set({
		isValid: pb.authStore.isValid,
		user: pb.authStore.model as AuthRecord | null
	});
});

// Encryption key store (stored in memory only, never persisted)
export const encryptionKeyStore = writable<string | null>(null);

// User stats store
export const userStatsStore = writable({
	totalEntries: 0,
	totalWords: 0,
	currentStreak: 0,
	longestStreak: 0,
	lastEntryDate: null as string | null,
	preferredAnalysisFrequency: 'weekly'
});

// Journal entries store
export const entriesStore = writable<any[]>([]);

// Loading state store
export const loadingStore = writable(false);

// Error store
export const errorStore = writable<string | null>(null);

// Theme store (dark/light mode)
export const themeStore = writable<'light' | 'dark'>('light');
