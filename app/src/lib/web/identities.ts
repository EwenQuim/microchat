export interface Identity {
	name: string;
	privateKey: string;
	publicKey: string;
}

export const IDENTITIES_KEY = "microchat_identities";
export const ACTIVE_IDENTITY_KEY = "microchat_active_identity";

export function getIdentities(): Identity[] {
	try {
		const raw = localStorage.getItem(IDENTITIES_KEY);
		if (!raw) return [];
		return JSON.parse(raw) as Identity[];
	} catch {
		return [];
	}
}

export function setIdentities(identities: Identity[]): void {
	localStorage.setItem(IDENTITIES_KEY, JSON.stringify(identities));
}

export function getActiveIdentityIndex(): number {
	try {
		const raw = localStorage.getItem(ACTIVE_IDENTITY_KEY);
		if (raw === null) return 0;
		return parseInt(raw, 10);
	} catch {
		return 0;
	}
}

export function setActiveIdentityIndex(index: number): void {
	localStorage.setItem(ACTIVE_IDENTITY_KEY, String(index));
}

/**
 * Migrate legacy single-identity storage into the identities list.
 * If identities list is empty but microchat_username / microchat_private_key exist,
 * create a single Identity entry and set it as active.
 */
export function migrateLegacyIdentity(): void {
	const existing = getIdentities();
	if (existing.length > 0) return;

	const name = localStorage.getItem("microchat_username");
	const privateKey = localStorage.getItem("microchat_private_key");
	const publicKey = localStorage.getItem("microchat_public_key");

	if (name && privateKey && publicKey) {
		setIdentities([{ name, privateKey, publicKey }]);
		setActiveIdentityIndex(0);
	}
}
