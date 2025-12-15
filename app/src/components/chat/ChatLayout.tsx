import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useUsername } from "@/hooks/useUsername";
import { ChatArea } from "./ChatArea";
import { RoomsSidebar } from "./RoomsSidebar";

interface ChatLayoutProps {
	roomName: string | null;
}

export function ChatLayout({ roomName }: ChatLayoutProps) {
	const { username, setUsername, isLoading } = useUsername();
	const [showUsernameDialog, setShowUsernameDialog] = useState(false);
	const [tempUsername, setTempUsername] = useState("");

	useEffect(() => {
		if (!isLoading && !username) {
			setShowUsernameDialog(true);
		}
	}, [isLoading, username]);

	const handleSaveUsername = () => {
		if (tempUsername.trim()) {
			setUsername(tempUsername.trim());
			setShowUsernameDialog(false);
		}
	};

	if (isLoading) {
		return (
			<div className="flex items-center justify-center h-screen">
				Loading...
			</div>
		);
	}

	return (
		<>
			<div className="flex h-screen">
				{/* Sidebar - Hidden on mobile when a room is selected */}
				<RoomsSidebar
					selectedRoom={roomName}
					className={`w-64 shrink-0 ${roomName ? "hidden md:flex" : "flex"}`}
				/>
				{/* Chat Area - Hidden on mobile when no room is selected */}
				<ChatArea
					roomName={roomName}
					username={username ?? ""}
					className={`flex-1 ${!roomName ? "hidden md:flex" : "flex"}`}
				/>
			</div>

			<Dialog open={showUsernameDialog} onOpenChange={setShowUsernameDialog}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Welcome to MicroChat</DialogTitle>
					</DialogHeader>
					<div className="space-y-4">
						<p className="text-sm text-muted-foreground">
							Please enter your name to start chatting
						</p>
						<Input
							value={tempUsername}
							onChange={(e) => setTempUsername(e.target.value)}
							placeholder="Your name"
							onKeyDown={(e) => e.key === "Enter" && handleSaveUsername()}
						/>
						<Button onClick={handleSaveUsername} className="w-full">
							Start Chatting
						</Button>
					</div>
				</DialogContent>
			</Dialog>
		</>
	);
}
