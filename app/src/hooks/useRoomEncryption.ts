import { useEffect, useState } from "react";
import type { Room } from "@/lib/api/generated/openAPI.schemas";
import { useEncryptionKeys } from "./useEncryptionKeys";
import { useRoomPassword } from "./useRoomPassword";

/**
 * Hook that manages encryption key lifecycle for a room
 *
 * Automatically derives encryption keys when:
 * - Room is encrypted
 * - Password is available
 * - Key hasn't been derived yet
 *
 * This ensures keys are available after page refresh or room navigation
 */
export function useRoomEncryption(
	roomName: string | null,
	room: Room | undefined,
) {
	const { password } = useRoomPassword(roomName || undefined);
	const { deriveAndStoreKey, getKey, hasKey } = useEncryptionKeys();
	const [encryptionKey, setEncryptionKey] = useState<CryptoKey | undefined>(
		undefined,
	);

	useEffect(() => {
		// If no room or not encrypted, clear the key
		if (!roomName || !room?.is_encrypted) {
			setEncryptionKey(undefined);
			return;
		}

		// Check if key already exists
		const existingKey = getKey(roomName);
		if (existingKey) {
			setEncryptionKey(existingKey);
			return;
		}

		// Derive key if we have password and salt but no key yet
		if (password !== undefined && room.encryption_salt && !hasKey(roomName)) {
			deriveAndStoreKey(roomName, password, room.encryption_salt).then(() => {
				// After derivation, get and set the key to trigger re-render
				const newKey = getKey(roomName);
				setEncryptionKey(newKey);
			});
		}
	}, [
		roomName,
		room?.is_encrypted,
		room?.encryption_salt,
		password,
		deriveAndStoreKey,
		getKey,
		hasKey,
	]);

	return encryptionKey;
}
