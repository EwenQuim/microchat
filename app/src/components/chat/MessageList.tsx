import { useRef } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { cn } from "@/lib/utils";
import { MessageItem } from "./MessageItem";

interface MessageListProps {
	messages: Message[];
	isLoading: boolean;
	currentPubKey: string;
	className?: string;
}

export function MessageList({
	messages,
	isLoading,
	currentPubKey,
	className,
}: MessageListProps) {
	const scrollRef = useRef<HTMLDivElement>(null);
	const shouldAutoScroll = useRef(true);

	const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
		const target = e.currentTarget;
		const isAtBottom =
			target.scrollHeight - target.scrollTop <= target.clientHeight + 50;
		shouldAutoScroll.current = isAtBottom;
	};

	return (
		<ScrollArea className={cn("px-4", className)} onScroll={handleScroll}>
			<div ref={scrollRef} className="pb-4">
				{isLoading && messages.length === 0 && (
					<div className="text-center text-muted-foreground my-8">
						Loading messages...
					</div>
				)}

				{!isLoading && messages.length === 0 && (
					<div className="text-center text-muted-foreground my-8">
						No messages yet. Be the first to say something!
					</div>
				)}

				<div className="space-y-4 my-4">
					{messages.map((message) => (
						<MessageItem
							key={message.id}
							message={message}
							isOwn={message.pubkey === currentPubKey}
						/>
					))}
				</div>
			</div>
		</ScrollArea>
	);
}
