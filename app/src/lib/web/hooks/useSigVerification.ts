import { useEffect, useState } from "react";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { verifySignature } from "@/lib/core/crypto";

export type SigStatus = "valid" | "invalid" | "unknown";

export function useSigVerification(
	messages: Message[],
	roomName: string,
): Record<string, SigStatus> | null {
	const [statuses, setStatuses] = useState<Record<string, SigStatus> | null>(
		null,
	);

	useEffect(() => {
		if (!messages.length) return;
		let cancelled = false;
		(async () => {
			const next: Record<string, SigStatus> = {};
			for (const msg of messages) {
				if (cancelled || !msg.id) continue;
				if (!msg.pubkey || !msg.signature || msg.signed_timestamp == null) {
					next[msg.id] = "unknown";
					continue;
				}
				try {
					const ok = await verifySignature({
						publicKey: msg.pubkey,
						signature: msg.signature,
						content: msg.content ?? "",
						room: roomName,
						timestamp: msg.signed_timestamp,
					});
					next[msg.id] = ok ? "valid" : "invalid";
				} catch {
					next[msg.id] = "invalid";
				}
			}
			if (!cancelled) setStatuses(next);
		})();
		return () => {
			cancelled = true;
		};
	}, [messages, roomName]);

	return statuses;
}
