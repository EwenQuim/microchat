import { createFileRoute } from "@tanstack/react-router";
import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import { useUsername } from "@/hooks/useUsername";
import { hexToNpub, hexToNsec } from "@/lib/crypto";

export const Route = createFileRoute("/settings")({
	component: Settings,
});

function Settings() {
	const [activeTab, setActiveTab] = useState<"user" | "options">("user");
	const [showPrivateKey, setShowPrivateKey] = useState(false);
	const { username, keys } = useUsername();

	return (
		<div className="min-h-screen bg-gray-900 text-white p-8">
			<div className="max-w-4xl mx-auto">
				<h1 className="text-3xl font-bold mb-8">Settings</h1>

				{/* Tabs */}
				<div className="flex gap-2 mb-6 border-b border-gray-700">
					<button
						type="button"
						onClick={() => setActiveTab("user")}
						className={`px-6 py-3 font-medium transition-colors ${
							activeTab === "user"
								? "border-b-2 border-cyan-500 text-cyan-500"
								: "text-gray-400 hover:text-white"
						}`}
					>
						User
					</button>
					<button
						type="button"
						onClick={() => setActiveTab("options")}
						className={`px-6 py-3 font-medium transition-colors ${
							activeTab === "options"
								? "border-b-2 border-cyan-500 text-cyan-500"
								: "text-gray-400 hover:text-white"
						}`}
					>
						Options
					</button>
				</div>

				{/* Tab Content */}
				<div className="bg-gray-800 rounded-lg p-6">
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
									{keys?.publicKey ? hexToNpub(keys.publicKey) : "Not available"}
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
			</div>
		</div>
	);
}
