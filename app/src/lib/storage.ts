import type { KeyPair } from "./crypto";

const USERNAME_KEY = "microchat_username";
const PRIVATE_KEY = "microchat_private_key";
const PUBLIC_KEY = "microchat_public_key";

export const storage = {
	getUsername: (): string | null => {
		return localStorage.getItem(USERNAME_KEY);
	},

	setUsername: (username: string): void => {
		localStorage.setItem(USERNAME_KEY, username);
	},

	clearUsername: (): void => {
		localStorage.removeItem(USERNAME_KEY);
	},

	/**
	 * Store cryptographic keys
	 * WARNING: localStorage is NOT fully secure against XSS attacks
	 * For production use, consider using more secure storage mechanisms
	 */
	setKeys: (keys: KeyPair): void => {
		localStorage.setItem(PRIVATE_KEY, keys.privateKey);
		localStorage.setItem(PUBLIC_KEY, keys.publicKey);
	},

	/**
	 * Retrieve stored keys
	 * Returns null if keys are not found
	 */
	getKeys: (): KeyPair | null => {
		const privateKey = localStorage.getItem(PRIVATE_KEY);
		const publicKey = localStorage.getItem(PUBLIC_KEY);

		if (!privateKey || !publicKey) {
			return null;
		}

		return { privateKey, publicKey };
	},

	/**
	 * Clear stored keys
	 */
	clearKeys: (): void => {
		localStorage.removeItem(PRIVATE_KEY);
		localStorage.removeItem(PUBLIC_KEY);
	},

	/**
	 * Clear all stored data
	 */
	clearAll: (): void => {
		storage.clearUsername();
		storage.clearKeys();
	},
};
