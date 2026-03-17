import { useCallback, useEffect, useState } from "react";
import { derivePublicKey, generateKeyPair, nsecToHex } from "@/lib/core/crypto";
import type { Identity } from "@/lib/web/identities";
import {
	ACTIVE_IDENTITY_KEY,
	getActiveIdentityIndex,
	getIdentities,
	IDENTITIES_KEY,
	migrateLegacyIdentity,
	setActiveIdentityIndex,
	setIdentities,
} from "@/lib/web/identities";

export function useIdentities() {
	const [identities, setIdentitiesState] = useState<Identity[]>(() => {
		migrateLegacyIdentity();
		return getIdentities();
	});
	const [activeIndex, setActiveIndexState] = useState<number>(
		getActiveIdentityIndex,
	);

	useEffect(() => {
		const handler = (e: StorageEvent) => {
			if (e.key === IDENTITIES_KEY) {
				setIdentitiesState(getIdentities());
			}
			if (e.key === ACTIVE_IDENTITY_KEY) {
				setActiveIndexState(getActiveIdentityIndex());
			}
		};
		window.addEventListener("storage", handler);
		return () => window.removeEventListener("storage", handler);
	}, []);

	const applyIdentity = useCallback((identity: Identity) => {
		localStorage.setItem("microchat_username", identity.name);
		localStorage.setItem("microchat_private_key", identity.privateKey);
		localStorage.setItem("microchat_public_key", identity.publicKey);
	}, []);

	const addIdentity = useCallback(
		async (name: string, nsec?: string) => {
			let privateKey: string;
			let publicKey: string;

			if (nsec) {
				privateKey = nsecToHex(nsec);
				publicKey = derivePublicKey(privateKey);
			} else {
				const kp = await generateKeyPair();
				privateKey = kp.privateKey;
				publicKey = kp.publicKey;
			}

			const newIdentity: Identity = { name, privateKey, publicKey };
			const updated = [...getIdentities(), newIdentity];
			setIdentities(updated);
			setIdentitiesState(updated);

			// If this is the first identity, also apply it to active storage
			if (updated.length === 1) {
				setActiveIdentityIndex(0);
				setActiveIndexState(0);
				applyIdentity(newIdentity);
			}
		},
		[applyIdentity],
	);

	const switchIdentity = useCallback(
		(index: number) => {
			const current = getIdentities();
			if (index < 0 || index >= current.length) return;
			setActiveIdentityIndex(index);
			setActiveIndexState(index);
			applyIdentity(current[index]);
		},
		[applyIdentity],
	);

	const removeIdentity = useCallback(
		(index: number) => {
			const current = getIdentities();
			if (current.length <= 1) return; // Prevent removing last identity
			const updated = current.filter((_, i) => i !== index);
			setIdentities(updated);
			setIdentitiesState(updated);

			let newActive = getActiveIdentityIndex();
			if (index === newActive) {
				newActive = 0;
				setActiveIdentityIndex(0);
				setActiveIndexState(0);
				applyIdentity(updated[0]);
			} else if (index < newActive) {
				newActive = newActive - 1;
				setActiveIdentityIndex(newActive);
				setActiveIndexState(newActive);
			}
		},
		[applyIdentity],
	);

	return {
		identities,
		activeIndex,
		addIdentity,
		switchIdentity,
		removeIdentity,
	};
}
