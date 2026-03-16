import { useCallback, useEffect, useState } from "react";
import { gETApiServerInfo } from "@/lib/api/generated/default/default";
import {
	getServers,
	normalizeServerUrl,
	removeServer as removeServerFromStorage,
	SERVERS_KEY,
	type Server,
	setServers,
	upsertServer,
} from "@/lib/servers";

export function useServers() {
	const [servers, setServersState] = useState<Server[]>(getServers);

	// Keep state in sync across hook instances via storage events
	useEffect(() => {
		const handler = (e: StorageEvent) => {
			if (e.key === SERVERS_KEY) {
				setServersState(getServers());
			}
		};
		window.addEventListener("storage", handler);
		return () => window.removeEventListener("storage", handler);
	}, []);

	// Auto-discover local server on first render using the generated API client
	useEffect(() => {
		const localUrl = normalizeServerUrl(
			window.location.origin +
				(import.meta.env.BASE_URL ?? "/").replace(/\/$/, ""),
		);

		gETApiServerInfo()
			.then((res) => {
				if (res.status !== 200) return;
				const info = res.data as {
					suggested_quickname?: string;
					description?: string;
				};
				upsertServer({
					url: localUrl,
					quickname: info.suggested_quickname ?? localUrl,
					description: info.description,
					isLocal: true,
					addedAt: Date.now(),
				});
				setServersState(getServers());
			})
			.catch(() => {
				// Ignore — server-info not available
			});
	}, []);

	const addServer = useCallback((server: Server) => {
		upsertServer(server);
		setServersState(getServers());
	}, []);

	const removeServer = useCallback((url: string) => {
		removeServerFromStorage(url);
		setServersState(getServers());
	}, []);

	const updateServers = useCallback((updated: Server[]) => {
		setServers(updated);
		setServersState(updated);
	}, []);

	return { servers, addServer, removeServer, updateServers };
}
