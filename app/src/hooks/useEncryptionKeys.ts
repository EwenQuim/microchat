import { useCallback, useRef } from "react";
import { deriveEncryptionKey } from "@/lib/crypto/e2e";

/**
 * Hook for managing E2E encryption keys in memory (never persisted)
 *
 * Encryption keys are derived from room passwords and stored only in memory
 * for the duration of the session. Keys are cleared when the user navigates
 * away or closes the tab.
 */
export function useEncryptionKeys() {
	// Store keys in memory only (not in state to prevent re-renders on key changes)
	const keysRef = useRef<Map<string, CryptoKey>>(new Map());

	/**
	 * Derives and stores an encryption key for a room
	 *
	 * @param roomName - The name of the room
	 * @param password - The room password
	 * @param salt - Hex-encoded salt from room metadata
	 */
	const deriveAndStoreKey = useCallback(
		async (roomName: string, password: string, salt: string) => {
			const key = await deriveEncryptionKey(password, salt);
			keysRef.current.set(roomName, key);
		},
		[],
	);

	/**
	 * Gets the encryption key for a room (if available)
	 *
	 * @param roomName - The name of the room
	 * @returns The encryption key, or undefined if not available
	 */
	const getKey = useCallback((roomName: string): CryptoKey | undefined => {
		return keysRef.current.get(roomName);
	}, []);

	/**
	 * Checks if an encryption key exists for a room
	 *
	 * @param roomName - The name of the room
	 * @returns True if a key exists for the room
	 */
	const hasKey = useCallback((roomName: string): boolean => {
		return keysRef.current.has(roomName);
	}, []);

	/**
	 * Clears the encryption key for a room
	 *
	 * @param roomName - The name of the room
	 */
	const clearKey = useCallback((roomName: string) => {
		keysRef.current.delete(roomName);
	}, []);

	/**
	 * Clears all encryption keys
	 */
	const clearAllKeys = useCallback(() => {
		keysRef.current.clear();
	}, []);

	return {
		deriveAndStoreKey,
		getKey,
		hasKey,
		clearKey,
		clearAllKeys,
	};
}
