import { Link, useNavigate } from "@tanstack/react-router";
import { MessageSquare, Plus } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useRooms } from "@/hooks/useRooms";
import { cn } from "@/lib/utils";
import { CreateRoomDialog } from "./CreateRoomDialog";

interface RoomsSidebarProps {
	selectedRoom: string | null;
	className?: string;
}

export function RoomsSidebar({ selectedRoom, className }: RoomsSidebarProps) {
	const { data: rooms, isLoading, error } = useRooms();
	const [showCreateDialog, setShowCreateDialog] = useState(false);
	const navigate = useNavigate();

	return (
		<aside className={cn("border-r bg-muted/10", className)}>
			<div className="flex flex-col h-full">
				<div className="p-4 border-b">
					<h2 className="font-semibold text-lg mb-2">Rooms</h2>
					<Button
						onClick={() => setShowCreateDialog(true)}
						className="w-full"
						size="sm"
					>
						<Plus className="mr-2 h-4 w-4" />
						Create Room
					</Button>
				</div>

				<ScrollArea className="flex-1">
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

					<div className="p-2 space-y-1">
						{rooms?.map((room) => {
							const roomName = room.name || "Unnamed Room";
							return (
								<Link
									key={roomName}
									to="/chat/$roomName"
									params={{ roomName }}
									className={cn(
										"w-full flex items-center justify-between p-3 rounded-lg transition-colors",
										"hover:bg-accent",
										selectedRoom === roomName && "bg-accent",
									)}
								>
									<div className="flex items-center gap-2">
										<MessageSquare className="h-4 w-4" />
										<span className="font-medium">{roomName}</span>
									</div>
									{room.message_count && room.message_count > 0 && (
										<Badge variant="secondary" className="text-xs">
											{room.message_count}
										</Badge>
									)}
								</Link>
							);
						})}
					</div>
				</ScrollArea>
			</div>

			<CreateRoomDialog
				open={showCreateDialog}
				onOpenChange={setShowCreateDialog}
				onRoomCreated={(roomName) =>
					navigate({ to: "/chat/$roomName", params: { roomName } })
				}
			/>
		</aside>
	);
}
