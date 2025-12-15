import { formatDistanceToNow } from "date-fns";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { cn } from "@/lib/utils";

interface MessageItemProps {
	message: Message;
	isOwn: boolean;
}

export function MessageItem({ message, isOwn }: MessageItemProps) {
	const getInitials = (name: string) => {
		return name
			.split(" ")
			.map((n) => n[0])
			.join("")
			.toUpperCase()
			.slice(0, 2);
	};

	const formattedTime = message.timestamp
		? formatDistanceToNow(new Date(message.timestamp), {
				addSuffix: true,
			})
		: "";

	return (
		<div className={cn("flex gap-3", isOwn && "flex-row-reverse")}>
			<Avatar className="h-8 w-8">
				<AvatarFallback>
					{getInitials(message.user || "Anonymous")}
				</AvatarFallback>
			</Avatar>

			<div className={cn("flex flex-col", isOwn && "items-end")}>
				<div className="flex items-center gap-2 mb-1">
					<span className="text-sm font-semibold">
						{message.user || "Anonymous"}
					</span>
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
					<p className="text-sm whitespace-pre-wrap break-words">
						{message.content}
					</p>
				</div>
			</div>
		</div>
	);
}
