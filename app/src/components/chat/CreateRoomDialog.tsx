import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useRoomPassword } from "@/hooks/useRoomPassword";
import { usePOSTApiRooms } from "@/lib/api/generated/api/api";

interface CreateRoomDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onRoomCreated: (roomName: string) => void;
}

export function CreateRoomDialog({
	open,
	onOpenChange,
	onRoomCreated,
}: CreateRoomDialogProps) {
	const [roomName, setRoomName] = useState("");
	const [password, setPassword] = useState("");
	const queryClient = useQueryClient();
	const { setPassword: storePassword } = useRoomPassword();

	const createRoomMutation = usePOSTApiRooms({
		mutation: {
			onSuccess: (response) => {
				if (response.status === 200 && response.data.name) {
					// Store password if room was created with one
					if (password) {
						storePassword(response.data.name, password);
					}
					onRoomCreated(response.data.name);
				}
				setRoomName("");
				setPassword("");
				onOpenChange(false);
			},
			onSettled: () => queryClient.invalidateQueries(),
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (roomName.trim()) {
			createRoomMutation.mutate({
				data: {
					name: roomName.trim(),
					password: password || undefined,
				},
			});
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Create New Room</DialogTitle>
				</DialogHeader>
				<form onSubmit={handleSubmit}>
					<div className="space-y-4">
						<div>
							<Input
								value={roomName}
								onChange={(e) => setRoomName(e.target.value)}
								placeholder="Room name"
								maxLength={50}
							/>
						</div>
						<div>
							<label htmlFor="room-password" className="text-sm font-medium">
								Password (optional)
							</label>
							<Input
								id="room-password"
								type="password"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								placeholder="Leave empty for public room"
								maxLength={72}
								className="mt-1"
							/>
							<p className="text-xs text-muted-foreground mt-1">
								Set a password to make this room private
							</p>
						</div>
						{createRoomMutation.isError && (
							<p className="text-sm text-destructive">
								Failed to create room. It may already exist.
							</p>
						)}
					</div>
					<DialogFooter className="mt-4">
						<Button
							type="button"
							variant="outline"
							onClick={() => onOpenChange(false)}
						>
							Cancel
						</Button>
						<Button
							type="submit"
							disabled={!roomName.trim() || createRoomMutation.isPending}
						>
							{createRoomMutation.isPending ? "Creating..." : "Create"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
