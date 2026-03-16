export interface Server {
	url: string; // domain or http://... (no https://, no trailing slash)
	quickname: string;
	description?: string;
	color?: string;
	addedAt: number;
	isLocal?: boolean;
}

/** Strip https:// before storage; keep http://; trim and remove trailing slash. */
export function normalizeServerUrl(url: string): string {
	return url
		.trim()
		.replace(/\/$/, "")
		.replace(/^https:\/\//, "");
}

/** Add https:// at usage time if no protocol is present. */
export function getServerUrl(url: string): string {
	if (!url || /^https?:\/\//.test(url)) return url;
	return `https://${url}`;
}

export const SERVERS_KEY = "microchat_servers";

export function getServers(): Server[] {
	try {
		const raw = localStorage.getItem(SERVERS_KEY);
		if (!raw) return [];
		return JSON.parse(raw) as Server[];
	} catch {
		return [];
	}
}

export function setServers(servers: Server[]): void {
	localStorage.setItem(SERVERS_KEY, JSON.stringify(servers));
}

export function upsertServer(server: Server): void {
	const servers = getServers();
	const idx = servers.findIndex((s) => s.url === server.url);
	if (idx >= 0) {
		servers[idx] = { ...servers[idx], ...server };
	} else {
		servers.push(server);
	}
	setServers(servers);
}

export function removeServer(url: string): void {
	setServers(getServers().filter((s) => s.url !== url));
}
