import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

interface RoomPasswordDialogProps {
	open: boolean;
	roomName: string;
	onSubmit: (password: string) => void;
	onCancel: () => void;
	error?: string;
}

export function RoomPasswordDialog({
	open,
	roomName,
	onSubmit,
	onCancel,
	error,
}: RoomPasswordDialogProps) {
	const [password, setPassword] = useState("");

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (password) {
			onSubmit(password);
		}
	};

	const handleOpenChange = (newOpen: boolean) => {
		if (!newOpen) {
			onCancel();
		}
	};

	return (
		<Dialog open={open} onOpenChange={handleOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Room Password Required</DialogTitle>
					<DialogDescription>
						This room is password-protected. Enter the password to continue.
					</DialogDescription>
				</DialogHeader>
				<form onSubmit={handleSubmit}>
					<div className="space-y-4">
						<div>
							<p className="text-sm font-medium mb-2">Room: #{roomName}</p>
							<Input
								type="password"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								placeholder="Enter password"
								autoFocus
							/>
						</div>
						{error && <p className="text-sm text-destructive">{error}</p>}
					</div>
					<DialogFooter className="mt-4">
						<Button type="button" variant="outline" onClick={onCancel}>
							Cancel
						</Button>
						<Button type="submit" disabled={!password}>
							Enter Room
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
