import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useUsername } from "@/lib/web/hooks/useUsername";

export const Route = createFileRoute("/settings/import")({
	component: ImportSettings,
	validateSearch: (search: Record<string, unknown>) => {
		return {
			import: (search.import as string) || undefined,
		};
	},
});

function ImportSettings() {
	const [importNsec, setImportNsec] = useState("");
	const [importUsername, setImportUsername] = useState("");
	const [importError, setImportError] = useState("");
	const { importProfile } = useUsername();
	const navigate = useNavigate();
	const searchParams = Route.useSearch();

	useEffect(() => {
		if (searchParams.import) {
			setImportNsec(searchParams.import);
		}
	}, [searchParams.import]);

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

	return (
		<div className="space-y-6">
			<h2 className="text-2xl font-semibold mb-4">Import Profile</h2>
			<p className="text-sm text-muted-foreground">
				Import an existing profile using your secret key (nsec). You'll still
				need to choose a display name.
			</p>

			<div>
				<label
					htmlFor="import-username"
					className="block text-sm font-medium text-muted-foreground mb-2"
				>
					Username
				</label>
				<Input
					id="import-username"
					value={importUsername}
					onChange={(e) => setImportUsername(e.target.value)}
					placeholder="Choose your display name"
					className=""
				/>
			</div>

			<div>
				<label
					htmlFor="import-nsec"
					className="block text-sm font-medium text-muted-foreground mb-2"
				>
					Secret Key (nsec)
				</label>
				<Input
					id="import-nsec"
					value={importNsec}
					onChange={(e) => setImportNsec(e.target.value)}
					placeholder="nsec1..."
					className="font-mono"
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
	);
}
