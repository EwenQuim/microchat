import { createFileRoute } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useGETApiUsersPublicKey } from "@/lib/api/generated/user/user";
import { hexToNpub } from "@/lib/core/crypto";
import { useContacts } from "@/lib/web/hooks/useContacts";

export const Route = createFileRoute("/user/$pubkey")({
	component: UserProfilePage,
	validateSearch: (search: Record<string, unknown>) => ({
		displayName: (search.displayName as string) || "",
	}),
});

function UserProfilePage() {
	const { pubkey } = Route.useParams();
	const { displayName } = Route.useSearch();
	const { contacts, addContact, removeContact } = useContacts();

	const { data: response, isLoading, error } = useGETApiUsersPublicKey(pubkey);

	const [displayNameInput, setDisplayNameInput] = useState(displayName);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [added, setAdded] = useState(false);

	if (isLoading) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center">Loading user profile...</div>
			</div>
		);
	}

	if (error || !response || response.status !== 200) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center text-red-500">
					Error loading user profile
				</div>
			</div>
		);
	}

	const user = response.data;

	if (!user.created_at || !user.public_key) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center text-red-500">Invalid user data</div>
			</div>
		);
	}

	const createdDate = formatDistanceToNow(new Date(user.created_at), {
		addSuffix: true,
	});

	const npub = hexToNpub(user.public_key);
	const isContact = contacts.some((c) => c.npub === npub);

	return (
		<div className="container mx-auto p-8">
			<div className="max-w-2xl mx-auto">
				<h1 className="text-3xl font-bold mb-6">User Profile</h1>

				<div className="bg-muted rounded-lg p-6 space-y-4">
					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Public Key (npub)
						</div>
						<p className="font-mono text-sm break-all mt-1">{npub}</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Verified by {window.location.hostname} admins
						</div>
						<p className="mt-1">
							{user.verified ? (
								<span className="text-green-500">✓ Verified</span>
							) : (
								<span className="text-muted-foreground">Not verified</span>
							)}
						</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Member since
						</div>
						<p className="mt-1">{createdDate}</p>
					</div>
				</div>

				<div className="mt-6 space-y-3">
					{isContact ? (
						<Button variant="destructive" onClick={() => removeContact(npub)}>
							Remove Contact
						</Button>
					) : added ? (
						<p className="text-sm text-green-500">✓ Contact added</p>
					) : (
						<Button variant="outline" onClick={() => setDialogOpen(true)}>
							Add Contact
						</Button>
					)}
				</div>
			</div>

			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Add Contact</DialogTitle>
					</DialogHeader>
					<Input
						placeholder="Display name"
						value={displayNameInput}
						onChange={(e) => setDisplayNameInput(e.target.value)}
						autoFocus
					/>
					<DialogFooter>
						<DialogClose asChild>
							<Button variant="ghost">Cancel</Button>
						</DialogClose>
						<Button
							onClick={() => {
								addContact(npub, displayNameInput.trim() || npub);
								setDialogOpen(false);
								setAdded(true);
							}}
						>
							Confirm
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
