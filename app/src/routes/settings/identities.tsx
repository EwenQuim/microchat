import { createFileRoute } from "@tanstack/react-router";
import { Check, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { hexToNpub } from "@/lib/core/crypto";
import { useIdentities } from "@/lib/web/hooks/useIdentities";

export const Route = createFileRoute("/settings/identities")({
	component: IdentitiesSettings,
});

function truncateNpub(npub: string): string {
	if (npub.length <= 20) return npub;
	return `${npub.slice(0, 10)}...${npub.slice(-6)}`;
}

type AddMode = null | "generate" | "import";

function IdentitiesSettings() {
	const {
		identities,
		activeIndex,
		addIdentity,
		switchIdentity,
		removeIdentity,
	} = useIdentities();

	const [addMode, setAddMode] = useState<AddMode>(null);
	const [nameInput, setNameInput] = useState("");
	const [nsecInput, setNsecInput] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [isAdding, setIsAdding] = useState(false);

	const handleAdd = async () => {
		setError(null);
		if (!nameInput.trim()) {
			setError("Name is required.");
			return;
		}
		setIsAdding(true);
		try {
			await addIdentity(
				nameInput.trim(),
				addMode === "import" ? nsecInput.trim() : undefined,
			);
			setAddMode(null);
			setNameInput("");
			setNsecInput("");
		} catch (e) {
			setError(e instanceof Error ? e.message : "Failed to add identity.");
		} finally {
			setIsAdding(false);
		}
	};

	const handleCancel = () => {
		setAddMode(null);
		setNameInput("");
		setNsecInput("");
		setError(null);
	};

	return (
		<div className="space-y-4">
			<h2 className="text-2xl font-semibold mb-4">Identities</h2>

			<div className="space-y-2">
				{identities.length === 0 && (
					<p className="text-sm text-muted-foreground py-2">
						No identities found. Create one below.
					</p>
				)}
				{identities.map((identity, index) => {
					const isActive = index === activeIndex;
					let npub = identity.publicKey;
					try {
						npub = hexToNpub(identity.publicKey);
					} catch {
						// keep as-is
					}
					return (
						<div
							key={index}
							className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
								isActive
									? "border-cyan-500 bg-cyan-500/5"
									: "border-border hover:border-muted-foreground"
							}`}
							onClick={() => switchIdentity(index)}
						>
							<div className="flex-1 min-w-0">
								<div className="flex items-center gap-2">
									<span className="font-medium text-sm">{identity.name}</span>
									{isActive && (
										<Check className="h-4 w-4 text-cyan-500 shrink-0" />
									)}
								</div>
								<p
									className="text-xs text-muted-foreground truncate"
									title={npub}
								>
									{truncateNpub(npub)}
								</p>
							</div>
							<Button
								variant="ghost"
								size="icon"
								className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
								disabled={identities.length <= 1}
								onClick={(e) => {
									e.stopPropagation();
									removeIdentity(index);
								}}
							>
								<Trash2 className="h-4 w-4" />
							</Button>
						</div>
					);
				})}
			</div>

			{addMode === null ? (
				<div className="flex gap-2 pt-2 border-t border-border">
					<Button
						onClick={() => setAddMode("generate")}
						variant="outline"
						className="flex-1"
					>
						<Plus className="h-4 w-4 mr-2" />
						Generate
					</Button>
					<Button
						onClick={() => setAddMode("import")}
						variant="outline"
						className="flex-1"
					>
						<Plus className="h-4 w-4 mr-2" />
						Import nsec
					</Button>
				</div>
			) : (
				<div className="space-y-2 pt-2 border-t border-border">
					<h3 className="text-sm font-medium">
						{addMode === "generate"
							? "Generate New Identity"
							: "Import Identity"}
					</h3>
					<Input
						placeholder="Name"
						value={nameInput}
						onChange={(e) => setNameInput(e.target.value)}
					/>
					{addMode === "import" && (
						<Input
							placeholder="nsec1..."
							value={nsecInput}
							onChange={(e) => setNsecInput(e.target.value)}
						/>
					)}
					{error && <p className="text-xs text-destructive">{error}</p>}
					<div className="flex gap-2">
						<Button
							onClick={handleAdd}
							variant="outline"
							className="flex-1"
							disabled={isAdding || !nameInput.trim()}
						>
							{isAdding ? "Adding..." : "Add"}
						</Button>
						<Button onClick={handleCancel} variant="ghost" className="flex-1">
							Cancel
						</Button>
					</div>
				</div>
			)}
		</div>
	);
}
