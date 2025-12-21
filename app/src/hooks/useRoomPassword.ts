import { useCallback, useEffect, useState } from "react";

const STORAGE_KEY = "microchat_room_passwords";

interface RoomPasswords {
	[roomName: string]: string;
}

function getStoredPasswords(): RoomPasswords {
	try {
		const stored = sessionStorage.getItem(STORAGE_KEY);
		return stored ? JSON.parse(stored) : {};
	} catch {
		return {};
	}
}

function savePasswords(passwords: RoomPasswords): void {
	try {
		sessionStorage.setItem(STORAGE_KEY, JSON.stringify(passwords));
	} catch (error) {
		console.error("Failed to save room passwords:", error);
	}
}

export function useRoomPassword(roomName?: string) {
	const [passwords, setPasswords] = useState<RoomPasswords>(getStoredPasswords);

	// Listen for storage changes (in case of multiple tabs)
	useEffect(() => {
		const handleStorageChange = () => {
			setPasswords(getStoredPasswords());
		};

		window.addEventListener("storage", handleStorageChange);
		return () => window.removeEventListener("storage", handleStorageChange);
	}, []);

	const setPassword = useCallback((room: string, password: string) => {
		setPasswords((prev) => {
			const updated = { ...prev, [room]: password };
			savePasswords(updated);
			return updated;
		});
	}, []);

	const getPassword = useCallback(
		(room: string) => {
			return passwords[room];
		},
		[passwords],
	);

	const clearPassword = useCallback((room: string) => {
		setPasswords((prev) => {
			const updated = { ...prev };
			delete updated[room];
			savePasswords(updated);
			return updated;
		});
	}, []);

	const clearAllPasswords = useCallback(() => {
		setPasswords({});
		sessionStorage.removeItem(STORAGE_KEY);
	}, []);

	// If roomName is provided, return the password for that specific room
	const currentPassword = roomName ? passwords[roomName] : undefined;

	return {
		password: currentPassword,
		setPassword,
		getPassword,
		clearPassword,
		clearAllPasswords,
	};
}
