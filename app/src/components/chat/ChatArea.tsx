import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ArrowLeft, Share2, ShieldCheck } from "lucide-react";
import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useDecryptedMessages } from "@/hooks/useDecryptedMessages";
import { useEncryptionKeys } from "@/hooks/useEncryptionKeys";
import { useMessages } from "@/hooks/useMessages";
import { useRoomPassword } from "@/hooks/useRoomPassword";
import { useRooms } from "@/hooks/useRooms";
import { getStoredPasswords } from "@/hooks/useSearchRooms";
import { useSendMessage } from "@/hooks/useSendMessage";
import { getGETApiRoomsQueryKey } from "@/lib/api/generated/chat/chat";
import { type KeyPair, signMessage } from "@/lib/crypto";
import { encryptMessage } from "@/lib/crypto/e2e";
import { cn } from "@/lib/utils";
import { MessageInput } from "./MessageInput";
import { MessageList } from "./MessageList";
import { RoomPasswordDialog } from "./RoomPasswordDialog";

interface ChatAreaProps {
	roomName: string | null;
	username: string;
	currentPubKey: string;
	keys: KeyPair | null;
	className?: string;
}
export function ChatArea({
	roomName,
	username,
	currentPubKey,
	keys,
	className,
}: ChatAreaProps) {
	const navigate = useNavigate();
	const { password, setPassword } = useRoomPassword(roomName || undefined);
	const { deriveAndStoreKey, getKey } = useEncryptionKeys();
	const [showPasswordDialog, setShowPasswordDialog] = useState(false);
	const [passwordError, setPasswordError] = useState<string | undefined>();
	const queryClient = useQueryClient();

	const { data: rooms } = useRooms();
	const { data: messages, isLoading } = useMessages(roomName, password);
	const sendMessageMutation = useSendMessage();

	// Get current room data
	const currentRoom = rooms?.find((r) => r.name === roomName);

	// Get encryption key for decrypting messages
	const encryptionKey = roomName ? getKey(roomName) : undefined;

	// Decrypt messages if room is encrypted
	const decryptedMessages = useDecryptedMessages(messages, encryptionKey);

	// Check if error indicates password required
	useEffect(() => {
		if (roomName) {
			// Successfully accessed room - save to visited list
			// Store with current password (empty string for public rooms)
			setPassword(roomName, password || "");
			const visitedRooms = Object.keys(getStoredPasswords())?.join(",");
			queryClient.invalidateQueries({
				queryKey: getGETApiRoomsQueryKey({
					visited: visitedRooms || undefined,
				}),
			});
		}
	}, [roomName, password, setPassword, queryClient]);

	const handlePasswordSubmit = async (newPassword: string) => {
		if (roomName) {
			setPassword(roomName, newPassword);

			// Derive encryption key if room is encrypted
			if (currentRoom?.is_encrypted && currentRoom?.encryption_salt) {
				await deriveAndStoreKey(
					roomName,
					newPassword,
					currentRoom.encryption_salt,
				);
			}

			setShowPasswordDialog(false);
			setPasswordError(undefined);
		}
	};

	const handlePasswordCancel = () => {
		setShowPasswordDialog(false);
		navigate({ to: "/" });
	};

	const handleSendMessage = async (content: string) => {
		if (!roomName || !keys) return;

		let finalContent = content;
		let isEncrypted = false;
		let nonce: string | undefined;

		// Encrypt message if room is encrypted
		if (currentRoom?.is_encrypted) {
			const encryptionKey = getKey(roomName);
			if (!encryptionKey) {
				console.error("No encryption key available for encrypted room");
				return;
			}

			const encrypted = await encryptMessage(content, encryptionKey);
			finalContent = encrypted.ciphertext;
			nonce = encrypted.nonce;
			isEncrypted = true;
		}

		// Sign the message using Nostr-style cryptography (sign encrypted content if encrypted)
		const { signature, timestamp } = await signMessage({
			privateKey: keys.privateKey,
			publicKey: keys.publicKey,
			content: finalContent,
			room: roomName,
		});

		sendMessageMutation.mutate({
			room: roomName,
			data: {
				content: finalContent,
				user: username,
				signature,
				pubkey: keys.publicKey,
				room_password: password,
				timestamp,
				is_encrypted: isEncrypted,
				nonce,
			},
		});
	};

	const handleBack = () => {
		navigate({ to: "/" });
	};

	const handleShare = async () => {
		const shareUrl = `${window.location.href}`;

		if (navigator.share) {
			try {
				await navigator.share({
					title: `Join ${roomName} on MicroChat`,
					text: `Join the conversation in ${window.location.host} #${roomName}`,
					url: shareUrl,
				});
			} catch (_e) {
				// User cancelled or share failed
				console.log("Share cancelled");
			}
		} else {
			// Fallback: copy to clipboard
			await navigator.clipboard.writeText(shareUrl);
			// Could show a toast notification here
		}
	};

	if (!roomName) {
		return (
			<div className={cn("flex items-center justify-center", className)}>
				<div className="text-center">
					<h2 className="text-2xl font-semibold mb-2">Welcome to MicroChat</h2>
					<p className="text-muted-foreground">
						Select a room from the sidebar to start chatting
					</p>
				</div>
			</div>
		);
	}

	return (
		<>
			<div className={cn("flex flex-col h-screen", className)}>
				<div className="p-4 border-b flex items-center gap-3 bg-background sticky top-0 z-10">
					<Button
						type="button"
						variant="ghost"
						size="icon"
						onClick={handleBack}
						className="md:hidden"
						aria-label="Back to rooms"
					>
						<ArrowLeft className="h-5 w-5" />
					</Button>
					<h2 className="font-semibold text-lg">{roomName}</h2>
					{currentRoom?.is_encrypted && (
						<Badge
							variant="outline"
							className="text-xs text-green-600 border-green-600 flex items-center gap-1"
						>
							<ShieldCheck className="h-3 w-3" />
							Encrypted
						</Badge>
					)}
					<div className="flex-1" />
					<Button
						type="button"
						variant="ghost"
						size="icon"
						onClick={handleShare}
						aria-label="Share room"
					>
						<Share2 className="h-5 w-5" />
					</Button>
				</div>

				<MessageList
					messages={decryptedMessages}
					isLoading={isLoading}
					currentPubKey={currentPubKey}
					className="flex-1 min-h-0"
					onRetryPassword={() => setShowPasswordDialog(true)}
				/>

				<MessageInput
					onSend={handleSendMessage}
					disabled={sendMessageMutation.isPending}
					className="border-t bg-background sticky bottom-0 z-10"
				/>
			</div>

			{/* Password dialog */}
			<RoomPasswordDialog
				open={showPasswordDialog}
				roomName={roomName || ""}
				onSubmit={handlePasswordSubmit}
				onCancel={handlePasswordCancel}
				error={passwordError}
			/>
		</>
	);
}
