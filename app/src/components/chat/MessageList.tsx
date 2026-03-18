import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import type { SigStatus } from "@/lib/web/hooks/useSigVerification";
import { cn } from "@/lib/web/utils";
import { MessageItem } from "./MessageItem";

interface MessageListProps {
	messages: Message[];
	isLoading: boolean;
	currentPubKey: string;
	className?: string;
	onRetryPassword?: () => void;
	sigStatuses?: Record<string, SigStatus> | null;
	hasMore?: boolean;
	isLoadingOlder?: boolean;
	onLoadOlder?: () => void;
}

export function MessageList({
	messages,
	isLoading,
	currentPubKey,
	className,
	onRetryPassword,
	sigStatuses,
	hasMore,
	isLoadingOlder,
	onLoadOlder,
}: MessageListProps) {
	const viewportRef = useRef<HTMLDivElement>(null);
	const shouldAutoScroll = useRef(true);
	const prevScrollHeightRef = useRef<number | null>(null);
	const prevMessagesLengthRef = useRef(0);

	const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
		const target = e.currentTarget;
		const isAtBottom =
			target.scrollHeight - target.scrollTop <= target.clientHeight + 50;
		shouldAutoScroll.current = isAtBottom;
	};

	// Auto-scroll to bottom when new messages arrive at bottom
	useEffect(() => {
		const viewport = viewportRef.current;
		if (!viewport) return;

		const messagesLengthIncreased =
			messages.length > prevMessagesLengthRef.current;
		const messagesAppended =
			messagesLengthIncreased &&
			prevScrollHeightRef.current !== null &&
			viewport.scrollHeight > prevScrollHeightRef.current;

		if (messages.length > 0 && prevMessagesLengthRef.current === 0) {
			// Initial load: scroll to bottom
			viewport.scrollTop = viewport.scrollHeight;
		} else if (messagesAppended && shouldAutoScroll.current) {
			// New messages appended at bottom and user is near bottom
			viewport.scrollTop = viewport.scrollHeight;
		} else if (
			messagesLengthIncreased &&
			prevScrollHeightRef.current !== null &&
			!shouldAutoScroll.current
		) {
			// Older messages prepended: preserve scroll position
			const heightDiff = viewport.scrollHeight - prevScrollHeightRef.current;
			viewport.scrollTop += heightDiff;
		}

		prevMessagesLengthRef.current = messages.length;
		prevScrollHeightRef.current = viewport.scrollHeight;
	}, [messages]);

	return (
		<ScrollArea
			className={cn("px-4", className)}
			viewportRef={viewportRef}
			onScroll={handleScroll}
		>
			<div className="pb-4">
				{hasMore && (
					<div className="text-center my-2">
						<Button
							variant="outline"
							size="sm"
							onClick={onLoadOlder}
							disabled={isLoadingOlder}
						>
							{isLoadingOlder ? "Loading..." : "Load older messages"}
						</Button>
					</div>
				)}

				{isLoading && messages.length === 0 && (
					<div className="text-center text-muted-foreground my-8">
						Loading messages...
					</div>
				)}

				{!isLoading && messages.length === 0 && (
					<div className="text-center text-muted-foreground my-8">
						<p className="mb-4">
							No messages yet, or incorrect password. Be the first to say
							something!
						</p>
						{onRetryPassword && (
							<Button
								variant="outline"
								onClick={onRetryPassword}
								className="mt-2"
							>
								Enter another password
							</Button>
						)}
					</div>
				)}

				<div className="space-y-4 my-4">
					{messages.map((message) => (
						<MessageItem
							key={message.id}
							message={message}
							isOwn={message.pubkey === currentPubKey}
							sigStatus={
								sigStatuses && message.id ? sigStatuses[message.id] : undefined
							}
						/>
					))}
				</div>
			</div>
		</ScrollArea>
	);
}
