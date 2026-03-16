import { createFileRoute } from "@tanstack/react-router";
import { QrCode } from "lucide-react";
import QRCode from "qrcode";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { hexToNsec } from "@/lib/core/crypto";
import { useUsername } from "@/lib/web/hooks/useUsername";

export const Route = createFileRoute("/settings/export")({
	component: ExportSettings,
});

function ExportSettings() {
	const [showQR, setShowQR] = useState(false);
	const [qrCodeDataUrl, setQrCodeDataUrl] = useState<string | null>(null);
	const { keys } = useUsername();

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

	return (
		<div className="space-y-6">
			<h2 className="text-2xl font-semibold mb-4">Export Profile</h2>
			<p className="text-sm text-muted-foreground mb-4">
				Share your profile by displaying a QR code. Anyone scanning it can
				import your identity on their device.
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
							<p className="text-xs text-muted-foreground text-center">
								This QR code contains your secret key. Only share with devices
								you own.
							</p>
						</div>
					)}
				</>
			) : (
				<div className="bg-yellow-500/10 border border-yellow-500 text-yellow-500 px-4 py-3 rounded-lg text-sm">
					No keys available to export. Please create or import a profile first.
				</div>
			)}
		</div>
	);
}
