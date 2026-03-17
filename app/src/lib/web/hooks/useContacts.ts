import { useCallback, useEffect, useState } from "react";
import { npubToHex } from "@/lib/core/crypto";
import type { Contact } from "@/lib/web/contacts";
import {
	addContact as addContactToStorage,
	CONTACTS_KEY,
	getContacts,
	removeContact as removeContactFromStorage,
} from "@/lib/web/contacts";

export function useContacts() {
	const [contacts, setContactsState] = useState<Contact[]>(getContacts);

	useEffect(() => {
		const handler = (e: StorageEvent) => {
			if (e.key === CONTACTS_KEY) {
				setContactsState(getContacts());
			}
		};
		window.addEventListener("storage", handler);
		return () => window.removeEventListener("storage", handler);
	}, []);

	const addContact = useCallback((npub: string, displayName: string) => {
		// Validate npub format
		npubToHex(npub);
		addContactToStorage({ npub, displayName });
		setContactsState(getContacts());
	}, []);

	const removeContact = useCallback((npub: string) => {
		removeContactFromStorage(npub);
		setContactsState(getContacts());
	}, []);

	return { contacts, addContact, removeContact };
}
