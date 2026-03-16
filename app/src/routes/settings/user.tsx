import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Eye, EyeOff, Trash2 } from "lucide-react";
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
import { useUsername } from "@/hooks/useUsername";
import { hexToNpub, hexToNsec } from "@/lib/crypto";

export const Route = createFileRoute("/settings/user")({
	component: UserSettings,
});

function UserSettings() {
	const [showPrivateKey, setShowPrivateKey] = useState(false);
	const [showDeleteDialog, setShowDeleteDialog] = useState(false);
	const { username, keys, clearUsername } = useUsername();
	const navigate = useNavigate();

	const handleDeleteProfile = () => {
		clearUsername();
		setShowDeleteDialog(false);
		navigate({ to: "/", search: {} });
	};

	return (
		<div className="space-y-6">
			<h2 className="text-2xl font-semibold mb-4">User Information</h2>

			<div>
				<div className="block text-sm font-medium text-muted-foreground mb-2">
					Username
				</div>
				<div className="bg-muted px-4 py-3 rounded-lg text-foreground">
					{username || "Not set"}
				</div>
			</div>

			<div>
				<div className="block text-sm font-medium text-muted-foreground mb-2">
					Public Key
				</div>
				<div className="bg-muted px-4 py-3 rounded-lg text-foreground font-mono text-sm break-all">
					{keys?.publicKey ? hexToNpub(keys.publicKey) : "Not available"}
				</div>
			</div>

			<div>
				<div className="block text-sm font-medium text-muted-foreground mb-2">
					Secret Key
				</div>
				<div className="relative">
					<div className="bg-muted px-4 py-3 rounded-lg text-foreground font-mono text-sm break-all pr-12">
						{showPrivateKey
							? keys?.privateKey
								? hexToNsec(keys.privateKey)
								: "Not available"
							: "•".repeat(63)}
					</div>
					<button
						type="button"
						onClick={() => setShowPrivateKey(!showPrivateKey)}
						className="absolute right-3 top-1/2 -translate-y-1/2 p-2 hover:bg-accent rounded-lg transition-colors"
						aria-label={showPrivateKey ? "Hide secret key" : "Show secret key"}
					>
						{showPrivateKey ? <EyeOff size={20} /> : <Eye size={20} />}
					</button>
				</div>
				<p className="text-xs text-muted-foreground mt-2">
					Keep your secret key safe. Never share it with anyone.
				</p>
			</div>

			<div className="pt-6 mt-6 border-t border-red-500/20">
				<h3 className="text-lg font-semibold text-red-500 mb-4">Danger Zone</h3>
				<div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
					<div className="flex items-start justify-between gap-4">
						<div>
							<h4 className="font-medium text-red-400 mb-1">Delete Identity</h4>
							<p className="text-sm text-muted-foreground">
								Remove your identity from this device. Make sure to back up your
								secret key first (use Export tab). This only deletes the local
								data.
							</p>
						</div>
						<Button
							onClick={() => setShowDeleteDialog(true)}
							variant="destructive"
							className="shrink-0"
						>
							<Trash2 size={16} className="mr-2" />
							Delete
						</Button>
					</div>
				</div>
			</div>

			<Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle className="text-red-500">Delete Identity</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete your identity from this device?
							This will remove your username and keys. Make sure you've backed
							up your secret key if you want to restore this identity later.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button
							onClick={() => setShowDeleteDialog(false)}
							variant="outline"
							className=""
						>
							Cancel
						</Button>
						<Button onClick={handleDeleteProfile} variant="destructive">
							Delete Identity
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
