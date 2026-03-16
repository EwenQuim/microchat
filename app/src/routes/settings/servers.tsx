import { createFileRoute } from "@tanstack/react-router";
import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { AddServerDialog } from "@/components/servers/AddServerDialog";
import { Button } from "@/components/ui/button";
import { useServers } from "@/hooks/useServers";
import type { Server } from "@/lib/servers";

export const Route = createFileRoute("/settings/servers")({
	component: ServersSettings,
});

function ServersSettings() {
	const [showAddServerDialog, setShowAddServerDialog] = useState(false);
	const { servers, addServer, removeServer } = useServers();

	return (
		<div className="space-y-4">
			<h2 className="text-2xl font-semibold mb-4">Servers</h2>
			<div className="space-y-2">
				{servers.length === 0 && (
					<p className="text-sm text-muted-foreground py-2">
						No servers configured.
					</p>
				)}
				{servers.map((server: Server) => (
					<div
						key={server.url}
						className="flex items-center gap-3 p-3 rounded-lg border border-border"
					>
						{server.color && (
							<div
								className="h-4 w-4 rounded-full shrink-0"
								style={{ backgroundColor: server.color }}
							/>
						)}
						<div className="flex-1 min-w-0">
							<div className="flex items-center gap-2">
								<span className="font-medium text-sm">{server.quickname}</span>
								{server.isLocal && (
									<span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
										local
									</span>
								)}
							</div>
							<p className="text-xs text-muted-foreground truncate">
								{server.url}
							</p>
							{server.description && (
								<p className="text-xs text-muted-foreground mt-0.5">
									{server.description}
								</p>
							)}
						</div>
						{!server.isLocal && (
							<Button
								variant="ghost"
								size="icon"
								className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
								onClick={() => removeServer(server.url)}
							>
								<Trash2 className="h-4 w-4" />
							</Button>
						)}
					</div>
				))}
			</div>
			<Button
				onClick={() => setShowAddServerDialog(true)}
				variant="outline"
				className="w-full"
			>
				<Plus className="h-4 w-4 mr-2" />
				Add Server
			</Button>

			<AddServerDialog
				open={showAddServerDialog}
				onOpenChange={setShowAddServerDialog}
				onAdd={(server) => addServer(server)}
			/>
		</div>
	);
}
