import { useEffect, useState } from "react";
import { generateKeyPair, type KeyPair } from "@/lib/crypto";
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

	const clearUsername = () => {
		storage.clearAll();
		setUsernameState(null);
		setKeysState(null);
	};

	return { username, keys, setUsername, clearUsername, isLoading };
}
