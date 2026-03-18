import { createFileRoute } from "@tanstack/react-router";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { hexToNpub, npubToHex } from "@/lib/core/crypto";
import { useContacts } from "@/lib/web/hooks/useContacts";

export const Route = createFileRoute("/settings/contacts")({
	component: ContactsSettings,
});

function generateColorFromNpub(npub: string): string {
	let hash = 0;
	for (let i = 0; i < npub.length; i++) {
		hash = npub.charCodeAt(i) + ((hash << 5) - hash);
		hash = hash & hash;
	}
	const hue = Math.abs(hash) % 360;
	return `hsl(${hue}, 70%, 55%)`;
}

function truncateNpub(npub: string): string {
	if (npub.length <= 20) return npub;
	return `${npub.slice(0, 10)}...${npub.slice(-6)}`;
}

function ContactsSettings() {
	const { contacts, addContact, removeContact } = useContacts();
	const [npubInput, setNpubInput] = useState("");
	const [displayNameInput, setDisplayNameInput] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [editingNpub, setEditingNpub] = useState<string | null>(null);
	const [editingName, setEditingName] = useState("");

	const handleAdd = () => {
		setError(null);
		try {
			// Validate npub by converting to hex (throws if invalid)
			npubToHex(npubInput.trim());
			addContact(npubInput.trim(), displayNameInput.trim() || npubInput.trim());
			setNpubInput("");
			setDisplayNameInput("");
		} catch {
			setError("Invalid npub. Please enter a valid Nostr public key.");
		}
	};

	return (
		<div className="space-y-4">
			<h2 className="text-2xl font-semibold mb-4">Contacts</h2>

			<div className="space-y-2">
				{contacts.length === 0 && (
					<p className="text-sm text-muted-foreground py-2">
						No contacts added yet.
					</p>
				)}
				{contacts.map((contact) => {
					const color = generateColorFromNpub(contact.npub);
					let fullNpub = contact.npub;
					// If stored as hex, convert for display
					try {
						if (!contact.npub.startsWith("npub")) {
							fullNpub = hexToNpub(contact.npub);
						}
					} catch {
						// keep as-is
					}
					const isEditing = editingNpub === contact.npub;
					return (
						<div
							key={contact.npub}
							className="flex items-center gap-3 p-3 rounded-lg border border-border"
						>
							<div
								className="h-4 w-4 rounded-full shrink-0"
								style={{ backgroundColor: color }}
							/>
							{isEditing ? (
								<div className="flex-1 flex items-center gap-2 min-w-0">
									<Input
										className="h-7 text-sm"
										value={editingName}
										onChange={(e) => setEditingName(e.target.value)}
										onKeyDown={(e) => {
											if (e.key === "Enter") {
												addContact(
													editingNpub!,
													editingName.trim() || editingNpub!,
												);
												setEditingNpub(null);
											} else if (e.key === "Escape") {
												setEditingNpub(null);
											}
										}}
										autoFocus
									/>
									<Button
										variant="outline"
										size="sm"
										className="h-7 px-2 shrink-0"
										onClick={() => {
											addContact(
												editingNpub!,
												editingName.trim() || editingNpub!,
											);
											setEditingNpub(null);
										}}
									>
										Save
									</Button>
									<Button
										variant="ghost"
										size="sm"
										className="h-7 px-2 shrink-0"
										onClick={() => setEditingNpub(null)}
									>
										Cancel
									</Button>
								</div>
							) : (
								<>
									<div className="flex-1 min-w-0">
										<span className="font-medium text-sm" style={{ color }}>
											{contact.displayName}
										</span>
										<p
											className="text-xs text-muted-foreground truncate"
											title={fullNpub}
										>
											{truncateNpub(fullNpub)}
										</p>
									</div>
									<Button
										variant="ghost"
										size="icon"
										className="h-8 w-8 shrink-0 text-muted-foreground hover:text-foreground"
										onClick={() => {
											setEditingNpub(contact.npub);
											setEditingName(contact.displayName);
										}}
									>
										<Pencil className="h-4 w-4" />
									</Button>
									<Button
										variant="ghost"
										size="icon"
										className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
										onClick={() => removeContact(contact.npub)}
									>
										<Trash2 className="h-4 w-4" />
									</Button>
								</>
							)}
						</div>
					);
				})}
			</div>

			<div className="space-y-2 pt-2 border-t border-border">
				<h3 className="text-sm font-medium">Add Contact</h3>
				<Input
					placeholder="npub1..."
					value={npubInput}
					onChange={(e) => setNpubInput(e.target.value)}
				/>
				<Input
					placeholder="Display name"
					value={displayNameInput}
					onChange={(e) => setDisplayNameInput(e.target.value)}
				/>
				{error && <p className="text-xs text-destructive">{error}</p>}
				<Button
					onClick={handleAdd}
					variant="outline"
					className="w-full"
					disabled={!npubInput.trim()}
				>
					<Plus className="h-4 w-4 mr-2" />
					Add Contact
				</Button>
			</div>
		</div>
	);
}
