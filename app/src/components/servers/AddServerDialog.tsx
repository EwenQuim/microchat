import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { gETApiServerInfo } from "@/lib/api/generated/default/default";
import type { MutatorOptions } from "@/lib/api/mutator";
import { normalizeServerUrl, type Server } from "@/lib/servers";
import { cn } from "@/lib/utils";

const PRESET_COLORS = [
	"#6366f1",
	"#ec4899",
	"#f59e0b",
	"#10b981",
	"#3b82f6",
	"#ef4444",
	"#8b5cf6",
	"#14b8a6",
];

interface AddServerDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onAdd: (server: Server) => void;
}

export function AddServerDialog({
	open,
	onOpenChange,
	onAdd,
}: AddServerDialogProps) {
	const [url, setUrl] = useState("");
	const [quickname, setQuickname] = useState("");
	const [description, setDescription] = useState("");
	const [color, setColor] = useState(PRESET_COLORS[0]);
	const [step, setStep] = useState<"url" | "details">("url");
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const handleUrlSubmit = async () => {
		const stored = normalizeServerUrl(url);
		if (!stored) return;
		setIsLoading(true);
		setError(null);
		try {
			const opts: MutatorOptions = { baseUrl: stored };
			const res = await gETApiServerInfo(opts);
			if (res.status !== 200) throw new Error(`Server returned ${res.status}`);
			const info = res.data;
			setQuickname(info.suggested_quickname ?? stored);
			setDescription(info.description ?? "");
			setUrl(stored);
			setStep("details");
		} catch {
			setError("Could not connect to server. Check the URL and try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleSave = () => {
		onAdd({
			url,
			quickname: quickname.trim() || url,
			description: description || undefined,
			color,
			addedAt: Date.now(),
		});
		handleClose();
	};

	const handleClose = () => {
		setUrl("");
		setQuickname("");
		setDescription("");
		setColor(PRESET_COLORS[0]);
		setStep("url");
		setError(null);
		onOpenChange(false);
	};

	return (
		<Dialog open={open} onOpenChange={handleClose}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Add Server</DialogTitle>
				</DialogHeader>

				{step === "url" && (
					<div className="space-y-4">
						<div className="space-y-2">
							<label className="text-sm font-medium" htmlFor="server-url">
								Server URL
							</label>
							<Input
								id="server-url"
								value={url}
								onChange={(e) => setUrl(e.target.value)}
								placeholder="https://chat.example.com"
								onKeyDown={(e) => e.key === "Enter" && handleUrlSubmit()}
							/>
							{error && <p className="text-sm text-destructive">{error}</p>}
						</div>
						<Button
							onClick={handleUrlSubmit}
							disabled={isLoading || !url.trim()}
							className="w-full"
						>
							{isLoading ? "Connecting..." : "Connect"}
						</Button>
					</div>
				)}

				{step === "details" && (
					<div className="space-y-4">
						{description && (
							<p className="text-sm text-muted-foreground">{description}</p>
						)}
						<div className="space-y-2">
							<label className="text-sm font-medium" htmlFor="server-quickname">
								Display name
							</label>
							<Input
								id="server-quickname"
								value={quickname}
								onChange={(e) => setQuickname(e.target.value)}
								placeholder="My Server"
							/>
						</div>
						<div className="space-y-2">
							<p className="text-sm font-medium">Color</p>
							<div className="flex gap-2 flex-wrap">
								{PRESET_COLORS.map((c) => (
									<button
										key={c}
										type="button"
										onClick={() => setColor(c)}
										className={cn(
											"h-7 w-7 rounded-full border-2 transition-transform hover:scale-110",
											color === c
												? "border-white ring-2 ring-offset-1"
												: "border-transparent",
										)}
										style={{ backgroundColor: c }}
									/>
								))}
							</div>
						</div>
						<div className="flex gap-2">
							<Button
								variant="outline"
								onClick={() => setStep("url")}
								className="flex-1"
							>
								Back
							</Button>
							<Button onClick={handleSave} className="flex-1">
								Add Server
							</Button>
						</div>
					</div>
				)}
			</DialogContent>
		</Dialog>
	);
}
