import PocketBase from 'pocketbase';
import { browser } from '$app/environment';

// Get PocketBase URL from environment variable
const PB_URL = import.meta.env.VITE_POCKETBASE_URL || 'http://localhost:8090';

// Create PocketBase instance
export const pb = new PocketBase(PB_URL);

// Load auth data from localStorage if in browser
if (browser) {
	const authData = localStorage.getItem('pocketbase_auth');
	if (authData) {
		pb.authStore.loadFromString(authData);
	}

	// Save auth data to localStorage when it changes
	pb.authStore.onChange(() => {
		localStorage.setItem('pocketbase_auth', pb.authStore.exportToString());
	}, true);
}

// Helper function to check if user is authenticated
export function isAuthenticated(): boolean {
	return pb.authStore.isValid;
}

// Helper function to get current user
export function getCurrentUser() {
	return pb.authStore.model;
}

// Helper function to logout
export function logout() {
	pb.authStore.clear();
	if (browser) {
		window.location.href = '/';
	}
}
