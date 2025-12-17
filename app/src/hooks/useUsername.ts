import { useEffect, useState } from "react";
import {
	derivePublicKey,
	generateKeyPair,
	type KeyPair,
	nsecToHex,
} from "@/lib/crypto";
import { storage } from "@/lib/storage";

export function useUsername() {
	const [username, setUsernameState] = useState<string | null>(null);
	const [keys, setKeysState] = useState<KeyPair | null>(null);
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		const stored = storage.getUsername();
		const storedKeys = storage.getKeys();
		setUsernameState(stored);
		setKeysState(storedKeys);
		setIsLoading(false);
	}, []);

	const setUsername = async (name: string) => {
		// Generate new cryptographic keys for this user
		const newKeys = await generateKeyPair();

		storage.setUsername(name);
		storage.setKeys(newKeys);

		setUsernameState(name);
		setKeysState(newKeys);

		console.log("🔐 Generated new keypair for user:", name);
		console.log("Public key:", newKeys.publicKey);
	};

	const importProfile = (name: string, nsec: string) => {
		try {
			// Convert nsec to hex and derive public key
			const privateKey = nsecToHex(nsec);
			const publicKey = derivePublicKey(privateKey);

			const importedKeys: KeyPair = { privateKey, publicKey };

			storage.setUsername(name);
			storage.setKeys(importedKeys);

			setUsernameState(name);
			setKeysState(importedKeys);

			console.log("🔑 Imported profile for user:", name);
			console.log("Public key:", publicKey);
		} catch (error) {
			console.error("Failed to import profile:", error);
			throw new Error("Invalid nsec key. Please check and try again.");
		}
	};

	const clearUsername = () => {
		storage.clearAll();
		setUsernameState(null);
		setKeysState(null);
	};

	return {
		username,
		keys,
		setUsername,
		importProfile,
		clearUsername,
		isLoading,
	};
}
