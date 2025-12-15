import { useNavigate } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useMessages } from "@/hooks/useMessages";
import { useSendMessage } from "@/hooks/useSendMessage";
import { type KeyPair, signMessage } from "@/lib/crypto";
import { cn } from "@/lib/utils";
import { MessageInput } from "./MessageInput";
import { MessageList } from "./MessageList";

interface ChatAreaProps {
	roomName: string | null;
	username: string;
	keys: KeyPair | null;
	className?: string;
}

export function ChatArea({
	roomName,
	username,
	keys,
	className,
}: ChatAreaProps) {
	const navigate = useNavigate();
	const { data: messages, isLoading } = useMessages(roomName);
	const sendMessageMutation = useSendMessage(roomName || "");

	const handleSendMessage = async (content: string) => {
		if (!roomName || !keys) return;

		// Sign the message using Nostr-style cryptography
		const { signature, timestamp } = await signMessage({
			privateKey: keys.privateKey,
			publicKey: keys.publicKey,
			content,
			room: roomName,
		});

		sendMessageMutation.mutate({
			room: roomName,
			data: {
				content,
				user: username,
				signature,
				pubkey: keys.publicKey,
				timestamp,
			},
		});
	};

	const handleBack = () => {
		navigate({ to: "/" });
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
		<div className={cn("flex flex-col", className)}>
			<div className="p-4 border-b flex items-center gap-3">
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
				<h2 className="font-semibold text-lg">#{roomName}</h2>
			</div>

			<MessageList
				messages={messages || []}
				isLoading={isLoading}
				currentUsername={username}
				className="flex-1"
			/>

			<MessageInput
				onSend={handleSendMessage}
				disabled={sendMessageMutation.isPending}
				className="border-t"
			/>
		</div>
	);
}
