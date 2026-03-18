import { Link } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { hexToNpub } from "@/lib/core/crypto";
import { useContacts } from "@/lib/web/hooks/useContacts";
import type { SigStatus } from "@/lib/web/hooks/useSigVerification";
import { cn } from "@/lib/web/utils";

const generateColorFromPubkey = (pubkey: string): string => {
	let hash = 0;
	for (let i = 0; i < pubkey.length; i++) {
		hash = pubkey.charCodeAt(i) + ((hash << 5) - hash);
		hash = hash & hash;
	}
	const hue = Math.abs(hash) % 360;
	return `hsl(${hue}, 70%, 55%)`;
};

interface MessageItemProps {
	message: Message;
	isOwn: boolean;
	sigStatus?: SigStatus;
}

export function MessageItem({ message, isOwn, sigStatus }: MessageItemProps) {
	const { contacts } = useContacts();

	const pubkey = message.pubkey || "anonymous";
	const fullNpub =
		pubkey !== "anonymous" && pubkey ? hexToNpub(pubkey) : pubkey;

	const contact = contacts.find((c) => c.npub === fullNpub);
	const isContact = !!contact;
	const displayName = isContact
		? contact.displayName
		: message.user || "Anonymous";

	const formatPubkey = (pk: string): string => {
		if (pk === "anonymous" || !pk) return "anonymous";
		try {
			const npub = hexToNpub(pk);
			return npub.slice(-6);
		} catch {
			return pk.slice(-6);
		}
	};

	const displayPubkey = formatPubkey(pubkey);
	const userColor = generateColorFromPubkey(pubkey);

	const formattedTime = message.timestamp
		? formatDistanceToNow(new Date(message.timestamp), {
				addSuffix: true,
			})
		: "";

	return (
		<div className={cn("flex gap-3", isOwn && "flex-row-reverse")}>
			<div className={cn("flex flex-col", isOwn && "items-end")}>
				<div className="flex items-center gap-2 mb-1">
					<Link
						to="/user/$pubkey"
						params={{ pubkey }}
						search={{ displayName: message.user || "" }}
						className="text-sm font-semibold hover:underline cursor-pointer"
						style={{ color: userColor }}
					>
						{displayName}
					</Link>
					{isContact ? (
						<span className="text-xs text-muted-foreground" title={fullNpub}>
							✓
						</span>
					) : (
						<Link
							to="/user/$pubkey"
							params={{ pubkey }}
							search={{ displayName: message.user || "" }}
							className="text-xs text-muted-foreground font-mono hover:underline cursor-pointer"
							title={fullNpub}
						>
							@{displayPubkey}
						</Link>
					)}
					{sigStatus === "invalid" && (
						<span
							className="text-xs text-amber-500"
							title="Signature verification failed — message may have been tampered"
						>
							⚠
						</span>
					)}
					{sigStatus === "unknown" && (
						<span
							className="text-xs text-amber-500"
							title="No signature — cannot verify authenticity"
						>
							⚠
						</span>
					)}
					{formattedTime && (
						<span className="text-xs text-muted-foreground">
							{formattedTime}
						</span>
					)}
				</div>
				<div
					className={cn(
						"rounded-lg px-4 py-2 max-w-md",
						isOwn ? "bg-primary text-primary-foreground" : "bg-muted",
					)}
				>
					<p className="text-sm whitespace-pre-wrap wrap-break-word">
						{message.content}
					</p>
				</div>
			</div>
		</div>
	);
}
