import { Link, useNavigate } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import { Globe, Lock, Plus, Search, User } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useKeyboardShortcut } from "@/hooks/useKeyboardShortcut";
import { useRooms } from "@/hooks/useRooms";
import { useUsername } from "@/hooks/useUsername";
import { cn } from "@/lib/utils";
import { CreateRoomDialog } from "./CreateRoomDialog";
import { SearchCommand } from "./SearchCommand";

interface RoomsSidebarProps {
	selectedRoom: string | null;
	className?: string;
}

export function RoomsSidebar({ selectedRoom, className }: RoomsSidebarProps) {
	const { data: rooms, isLoading, error } = useRooms();
	const { username } = useUsername();
	const [showCreateDialog, setShowCreateDialog] = useState(false);
	const [showSearch, setShowSearch] = useState(false);
	const navigate = useNavigate();

	useKeyboardShortcut((e) => {
		if (e.key === "n" && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			setShowCreateDialog(true);
		}
	});

	return (
		<aside className={cn("border-r bg-muted/10 relative", className)}>
			<div className="flex flex-col h-full w-full">
				<div className="p-4 shrink-0 space-y-3">
					<div className="flex items-center justify-between">
						<h2 className="font-semibold text-lg">Murmure</h2>
						{username && (
							<Link
								to="/settings"
								search={{ import: undefined }}
								className={cn(
									"flex items-center gap-2 px-2 py-1 rounded-lg transition-colors",
									"hover:bg-accent",
								)}
							>
								<User className="h-4 w-4" />
								<span className="text-sm font-medium truncate max-w-24">
									{username}
								</span>
							</Link>
						)}
					</div>
					<button
						type="button"
						onClick={() => setShowSearch(true)}
						className="w-full flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground bg-muted/50 rounded-md hover:bg-muted transition-colors"
					>
						<Search className="h-4 w-4" />
						<span>Search rooms...</span>
						<kbd className="ml-auto pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
							<span className="text-xs">⌘</span>K
						</kbd>
					</button>
				</div>

				<ScrollArea className="flex-1 min-h-0">
					{isLoading && (
						<div className="p-4 text-sm text-muted-foreground">
							Loading rooms...
						</div>
					)}

					{!!error && (
						<div className="p-4 text-sm text-destructive">
							Failed to load rooms
						</div>
					)}

					{!isLoading && !error && rooms?.length === 0 && (
						<div className="p-4 text-sm text-muted-foreground">
							No rooms yet. Create one to start chatting!
						</div>
					)}

					<div className="p-2 space-y-1 pb-16">
						{rooms?.map((room) => {
							const roomName = room.name || "Unnamed Room";
							return (
								<Link
									key={roomName}
									to="/chat/$roomName"
									params={{ roomName }}
									className={cn(
										"w-full flex flex-col p-3 rounded-lg transition-colors",
										"hover:bg-accent",
										selectedRoom === roomName && "bg-accent",
									)}
								>
									<div className="flex items-center justify-between w-full min-w-0">
										<div className="flex items-center gap-2 min-w-0">
											{room.has_password && (
												<Lock className="h-3 w-3 text-muted-foreground shrink-0" />
											)}
											{!room.has_password && !room.visited && (
												<Globe className="h-3 w-3 text-muted-foreground shrink-0" />
											)}
											<span className="font-medium truncate">{roomName}</span>
										</div>
										{room.last_message_timestamp && (
											<span className="text-xs text-muted-foreground shrink-0">
												{formatDistanceToNow(
													new Date(room.last_message_timestamp),
													{
														addSuffix: true,
													},
												)}
											</span>
										)}
									</div>
									{room.last_message_content && (
										<div className="text-xs text-muted-foreground mt-1 truncate min-w-0 max-w-48">
											{room.last_message_user && (
												<span className="font-medium">
													{room.last_message_user}:{" "}
												</span>
											)}
											{room.last_message_content}
										</div>
									)}
								</Link>
							);
						})}
					</div>
				</ScrollArea>

				{/* Floating Create Button */}
				<Button
					onClick={() => setShowCreateDialog(true)}
					size="lg"
					className="absolute bottom-4 right-4 rounded-full shadow-lg h-12 w-12 p-0 z-10 cursor-pointer active:scale-90"
				>
					<Plus className="h-8 w-8" />
				</Button>
			</div>

			<CreateRoomDialog
				open={showCreateDialog}
				onOpenChange={setShowCreateDialog}
				onRoomCreated={(roomName) =>
					navigate({ to: "/chat/$roomName", params: { roomName } })
				}
			/>

			<SearchCommand open={showSearch} onOpenChange={setShowSearch} />
		</aside>
	);
}
