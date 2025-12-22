import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { ArrowLeft, Eye, EyeOff, QrCode, Trash2 } from "lucide-react";
import QRCode from "qrcode";
import { useEffect, useState } from "react";
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
import { useUsername } from "@/hooks/useUsername";
import { hexToNpub, hexToNsec } from "@/lib/crypto";

export const Route = createFileRoute("/settings")({
	component: Settings,
	validateSearch: (search: Record<string, unknown>) => {
		return {
			import: (search.import as string) || undefined,
		};
	},
});

function Settings() {
	const [activeTab, setActiveTab] = useState<
		"user" | "import" | "export" | "options"
	>("user");
	const [showPrivateKey, setShowPrivateKey] = useState(false);
	const [importNsec, setImportNsec] = useState("");
	const [importUsername, setImportUsername] = useState("");
	const [importError, setImportError] = useState("");
	const [showQR, setShowQR] = useState(false);
	const [qrCodeDataUrl, setQrCodeDataUrl] = useState<string | null>(null);
	const [showDeleteDialog, setShowDeleteDialog] = useState(false);
	const { username, keys, importProfile, clearUsername } = useUsername();
	const navigate = useNavigate();
	const searchParams = Route.useSearch();

	// Handle URL import parameter
	useEffect(() => {
		if (searchParams.import) {
			setActiveTab("import");
			setImportNsec(searchParams.import);
		}
	}, [searchParams.import]);

	const handleShowQR = async () => {
		if (!showQR && keys?.privateKey) {
			const nsec = hexToNsec(keys.privateKey);
			const importUrl = `${window.location.origin}/settings?import=${nsec}`;
			try {
				const url = await QRCode.toDataURL(importUrl, {
					width: 300,
					margin: 2,
				});
				setQrCodeDataUrl(url);
				setShowQR(true);
			} catch (error) {
				console.error(error);
			}
		} else {
			setShowQR(false);
		}
	};

	const handleImport = () => {
		setImportError("");
		if (!importUsername.trim()) {
			setImportError("Please enter a username");
			return;
		}
		if (!importNsec.trim()) {
			setImportError("Please enter an nsec key");
			return;
		}

		try {
			importProfile(importUsername.trim(), importNsec.trim());
			navigate({ to: "/", search: {} });
		} catch (error) {
			setImportError(
				error instanceof Error ? error.message : "Failed to import profile",
			);
		}
	};

	const handleDeleteProfile = () => {
		clearUsername();
		setShowDeleteDialog(false);
		navigate({ to: "/", search: {} });
	};

	const handleGoBack = () => {
		navigate({ to: "/", search: {} });
	};

	const TabButton = ({
		value,
		label,
	}: {
		value: typeof activeTab;
		label: string;
	}) => (
		<button
			type="button"
			onClick={() => setActiveTab(value)}
			className={`px-4 md:px-6 py-3 font-medium transition-colors whitespace-nowrap ${
				activeTab === value
					? "border-b-2 border-cyan-500 text-cyan-500"
					: "text-gray-400 hover:text-white"
			}`}
		>
			{label}
		</button>
	);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-4 md:p-8">
			<div className="max-w-4xl mx-auto">
				<div className="flex items-center gap-4 mb-8">
					<Button
						onClick={handleGoBack}
						variant="ghost"
						size="icon"
						className="text-gray-400 hover:text-white h-12 w-12 md:h-10 md:w-10"
					>
						<ArrowLeft className="h-6 w-6 md:h-5 md:w-5" />
					</Button>
					<h1 className="text-xl md:text-3xl font-bold">Settings</h1>
				</div>

				{/* Tabs */}
				<div className="flex gap-2 mb-6 border-b border-gray-700 overflow-x-auto -mx-4 md:-mx-8 px-4 md:px-8">
					<TabButton value="user" label="User" />
					<TabButton value="import" label="Import" />
					<TabButton value="export" label="Export" />
					{/* <TabButton value="options" label="Options" /> */}
				</div>

				{/* Tab Content */}
				<div className="bg-gray-800 rounded-lg p-4 md:p-6">
					{activeTab === "user" && (
						<div className="space-y-6">
							<h2 className="text-2xl font-semibold mb-4">User Information</h2>

							{/* Username */}
							<div>
								<div className="block text-sm font-medium text-gray-400 mb-2">
									Username
								</div>
								<div className="bg-gray-700 px-4 py-3 rounded-lg text-white">
									{username || "Not set"}
								</div>
							</div>

							{/* Public Key */}
							<div>
								<div className="block text-sm font-medium text-gray-400 mb-2">
									Public Key
								</div>
								<div className="bg-gray-700 px-4 py-3 rounded-lg text-white font-mono text-sm break-all">
									{keys?.publicKey
										? hexToNpub(keys.publicKey)
										: "Not available"}
								</div>
							</div>

							{/* Secret/Private Key */}
							<div>
								<div className="block text-sm font-medium text-gray-400 mb-2">
									Secret Key
								</div>
								<div className="relative">
									<div className="bg-gray-700 px-4 py-3 rounded-lg text-white font-mono text-sm break-all pr-12">
										{showPrivateKey
											? keys?.privateKey
												? hexToNsec(keys.privateKey)
												: "Not available"
											: "•".repeat(63)}
									</div>
									<button
										type="button"
										onClick={() => setShowPrivateKey(!showPrivateKey)}
										className="absolute right-3 top-1/2 -translate-y-1/2 p-2 hover:bg-gray-600 rounded-lg transition-colors"
										aria-label={
											showPrivateKey ? "Hide secret key" : "Show secret key"
										}
									>
										{showPrivateKey ? <EyeOff size={20} /> : <Eye size={20} />}
									</button>
								</div>
								<p className="text-xs text-gray-500 mt-2">
									Keep your secret key safe. Never share it with anyone.
								</p>
							</div>

							{/* Danger Zone */}
							<div className="pt-6 mt-6 border-t border-red-500/20">
								<h3 className="text-lg font-semibold text-red-500 mb-4">
									Danger Zone
								</h3>
								<div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
									<div className="flex items-start justify-between gap-4">
										<div>
											<h4 className="font-medium text-red-400 mb-1">
												Delete Identity
											</h4>
											<p className="text-sm text-gray-400">
												Remove your identity from this device. Make sure to back
												up your secret key first (use Export tab). This only
												deletes the local data.
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
						</div>
					)}

					{activeTab === "import" && (
						<div className="space-y-6">
							<h2 className="text-2xl font-semibold mb-4">Import Profile</h2>
							<p className="text-sm text-gray-400">
								Import an existing profile using your secret key (nsec). You'll
								still need to choose a display name.
							</p>

							<div>
								<label
									htmlFor="import-username"
									className="block text-sm font-medium text-gray-400 mb-2"
								>
									Username
								</label>
								<Input
									id="import-username"
									value={importUsername}
									onChange={(e) => setImportUsername(e.target.value)}
									placeholder="Choose your display name"
									className="bg-gray-700 border-gray-600 text-white"
								/>
							</div>

							<div>
								<label
									htmlFor="import-nsec"
									className="block text-sm font-medium text-gray-400 mb-2"
								>
									Secret Key (nsec)
								</label>
								<Input
									id="import-nsec"
									value={importNsec}
									onChange={(e) => setImportNsec(e.target.value)}
									placeholder="nsec1..."
									className="bg-gray-700 border-gray-600 text-white font-mono"
								/>
							</div>

							{importError && (
								<div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg text-sm">
									{importError}
								</div>
							)}

							<Button onClick={handleImport} className="w-full">
								Import Profile
							</Button>
						</div>
					)}

					{activeTab === "export" && (
						<div className="space-y-6">
							<h2 className="text-2xl font-semibold mb-4">Export Profile</h2>
							<p className="text-sm text-gray-400 mb-4">
								Share your profile by displaying a QR code. Anyone scanning it
								can import your identity on their device.
							</p>

							{keys?.privateKey ? (
								<>
									<Button
										onClick={handleShowQR}
										className="w-full"
										variant={showQR ? "outline" : "default"}
									>
										<QrCode className="mr-2" size={20} />
										{showQR ? "Hide QR Code" : "Show QR Code"}
									</Button>

									{showQR && qrCodeDataUrl && (
										<div className="flex flex-col items-center space-y-4">
											<div className="bg-white p-4 rounded-lg">
												<img
													src={qrCodeDataUrl}
													alt="Profile QR Code"
													className="w-full max-w-xs"
												/>
											</div>
											<p className="text-xs text-gray-500 text-center">
												This QR code contains your secret key. Only share with
												devices you own.
											</p>
										</div>
									)}
								</>
							) : (
								<div className="bg-yellow-500/10 border border-yellow-500 text-yellow-500 px-4 py-3 rounded-lg text-sm">
									No keys available to export. Please create or import a profile
									first.
								</div>
							)}
						</div>
					)}

					{activeTab === "options" && (
						<div>
							<h2 className="text-2xl font-semibold mb-4">Options</h2>
							<div className="text-gray-400 text-center py-8">
								<p className="text-xl">WIP</p>
								<p className="text-sm mt-2">
									Additional options will be available here soon.
								</p>
							</div>
						</div>
					)}
				</div>

				{/* Delete Confirmation Dialog */}
				<Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
					<DialogContent className="bg-gray-800 border-gray-700 text-white">
						<DialogHeader>
							<DialogTitle className="text-red-500">
								Delete Identity
							</DialogTitle>
							<DialogDescription className="text-gray-400">
								Are you sure you want to delete your identity from this device?
								This will remove your username and keys. Make sure you've backed
								up your secret key if you want to restore this identity later.
							</DialogDescription>
						</DialogHeader>
						<DialogFooter>
							<Button
								onClick={() => setShowDeleteDialog(false)}
								variant="outline"
								className="border-gray-600 text-white hover:bg-gray-700"
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
		</div>
	);
}
