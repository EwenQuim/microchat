import { Link } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { hexToNpub } from "@/lib/crypto";
import { cn } from "@/lib/utils";

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
}

export function MessageItem({ message, isOwn }: MessageItemProps) {
	const formatPubkey = (pubkey: string): string => {
		if (pubkey === "anonymous" || !pubkey) return "anonymous";
		try {
			const npub = hexToNpub(pubkey);
			// Display last 6 characters of npub
			return npub.slice(-6);
		} catch {
			return pubkey.slice(-6);
		}
	};

	const pubkey = message.pubkey || "anonymous";
	const displayPubkey = formatPubkey(pubkey);
	const fullNpub =
		pubkey !== "anonymous" && pubkey ? hexToNpub(pubkey) : pubkey;
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
						className="text-sm font-semibold hover:underline cursor-pointer"
						style={{ color: userColor }}
					>
						{message.user || "Anonymous"}
					</Link>
					<Link
						to="/user/$pubkey"
						params={{ pubkey }}
						className="text-xs text-muted-foreground font-mono hover:underline cursor-pointer"
						title={fullNpub}
					>
						@{displayPubkey}
					</Link>
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
