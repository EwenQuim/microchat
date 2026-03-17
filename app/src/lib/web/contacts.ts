export interface Contact {
	npub: string;
	displayName: string;
}

export const CONTACTS_KEY = "microchat_contacts";

export function getContacts(): Contact[] {
	try {
		const raw = localStorage.getItem(CONTACTS_KEY);
		if (!raw) return [];
		return JSON.parse(raw) as Contact[];
	} catch {
		return [];
	}
}

export function setContacts(contacts: Contact[]): void {
	localStorage.setItem(CONTACTS_KEY, JSON.stringify(contacts));
}

export function addContact(contact: Contact): void {
	const contacts = getContacts();
	const idx = contacts.findIndex((c) => c.npub === contact.npub);
	if (idx >= 0) {
		contacts[idx] = contact;
	} else {
		contacts.push(contact);
	}
	setContacts(contacts);
}

export function removeContact(npub: string): void {
	setContacts(getContacts().filter((c) => c.npub !== npub));
}
