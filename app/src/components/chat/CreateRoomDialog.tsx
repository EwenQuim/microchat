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
import {
	getGETApiRoomsQueryKey,
	usePOSTApiRooms,
} from "@/lib/api/generated/api/api";

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
	const queryClient = useQueryClient();

	const createRoomMutation = usePOSTApiRooms({
		mutation: {
			onSuccess: (response) => {
				queryClient.invalidateQueries({ queryKey: getGETApiRoomsQueryKey() });
				if (response.status === 200 && response.data.name) {
					onRoomCreated(response.data.name);
				}
				setRoomName("");
				onOpenChange(false);
			},
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (roomName.trim()) {
			createRoomMutation.mutate({ data: { name: roomName.trim() } });
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
						<Input
							value={roomName}
							onChange={(e) => setRoomName(e.target.value)}
							placeholder="Room name"
							maxLength={50}
						/>
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
