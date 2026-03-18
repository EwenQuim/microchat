import { createFileRoute } from "@tanstack/react-router";
import { Check, Eye, EyeOff, Plus, QrCode, Trash2 } from "lucide-react";
import QRCode from "qrcode";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { hexToNpub, hexToNsec } from "@/lib/core/crypto";
import { useIdentities } from "@/lib/web/hooks/useIdentities";

export const Route = createFileRoute("/settings/identities")({
	component: IdentitiesSettings,
	validateSearch: (search: Record<string, unknown>) => {
		return {
			import: (search.import as string) || undefined,
		};
	},
});

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

	const [nsecRevealIndex, setNsecRevealIndex] = useState<number | null>(null);
	const [qrExpandIndex, setQrExpandIndex] = useState<number | null>(null);
	const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);

	const searchParams = Route.useSearch();

	useEffect(() => {
		if (searchParams.import) {
			setNsecInput(searchParams.import);
			setAddMode("import");
		}
	}, [searchParams.import]);

	const handleShowQR = async (index: number) => {
		if (qrExpandIndex === index) {
			setQrExpandIndex(null);
			setQrDataUrl(null);
		} else {
			const identity = identities[index];
			if (!identity?.privateKey) return;
			const nsec = hexToNsec(identity.privateKey);
			const importUrl = `${window.location.origin}/settings/identities?import=${nsec}`;
			try {
				const url = await QRCode.toDataURL(importUrl, {
					width: 300,
					margin: 2,
				});
				setQrDataUrl(url);
				setQrExpandIndex(index);
				setNsecRevealIndex(null);
			} catch (err) {
				console.error(err);
			}
		}
	};

	const handleToggleNsec = (index: number) => {
		if (nsecRevealIndex === index) {
			setNsecRevealIndex(null);
		} else {
			setNsecRevealIndex(index);
			setQrExpandIndex(null);
			setQrDataUrl(null);
		}
	};

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
		<div className="space-y-6">
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
					let nsec = "";
					try {
						if (identity.privateKey) nsec = hexToNsec(identity.privateKey);
					} catch {
						// keep as-is
					}
					const nsecShown = nsecRevealIndex === index;
					const qrShown = qrExpandIndex === index;
					return (
						<div key={index} className="space-y-0">
							<div
								className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
									isActive
										? "border-cyan-500 bg-cyan-500/5"
										: "border-border hover:border-muted-foreground"
								} ${nsecShown || qrShown ? "rounded-b-none border-b-0" : ""}`}
								onClick={() => switchIdentity(index)}
							>
								<div className="flex-1 min-w-0">
									<div className="flex items-center gap-2">
										<span className="font-medium text-sm">{identity.name}</span>
										{isActive && (
											<Check className="h-4 w-4 text-cyan-500 shrink-0" />
										)}
									</div>
									<p className="text-xs text-muted-foreground font-mono break-all">
										<span className="text-muted-foreground/60">
											Public Key:{" "}
										</span>
										{npub}
									</p>
								</div>
								<Button
									variant="ghost"
									size="icon"
									className={`h-9 w-9 shrink-0 ${nsecShown ? "text-foreground" : "text-muted-foreground"}`}
									onClick={(e) => {
										e.stopPropagation();
										handleToggleNsec(index);
									}}
									aria-label={nsecShown ? "Hide secret key" : "Show secret key"}
								>
									{nsecShown ? (
										<EyeOff className="h-5 w-5" />
									) : (
										<Eye className="h-5 w-5" />
									)}
								</Button>
								{identity.privateKey && (
									<Button
										variant="ghost"
										size="icon"
										className={`h-8 w-8 shrink-0 ${qrShown ? "text-foreground" : "text-muted-foreground"}`}
										onClick={(e) => {
											e.stopPropagation();
											handleShowQR(index);
										}}
										aria-label={qrShown ? "Hide QR code" : "Show QR code"}
									>
										<QrCode className="h-4 w-4" />
									</Button>
								)}
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
							{nsecShown && (
								<div
									className={`px-3 py-2 border rounded-b-lg font-mono text-xs break-all bg-muted ${isActive ? "border-cyan-500" : "border-border"}`}
								>
									<span className="text-muted-foreground/60">Secret Key: </span>
									{nsec || "Not available"}
								</div>
							)}
							{qrShown && qrDataUrl && (
								<div
									className={`px-3 py-3 border rounded-b-lg bg-muted flex flex-col items-center gap-2 ${isActive ? "border-cyan-500" : "border-border"}`}
								>
									<div className="bg-white p-3 rounded-lg">
										<img
											src={qrDataUrl}
											alt="Identity QR Code"
											className="w-72 h-72"
										/>
									</div>
									<p className="text-xs text-muted-foreground text-center">
										This QR code contains your secret key. Only share with
										devices you own.
									</p>
								</div>
							)}
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
