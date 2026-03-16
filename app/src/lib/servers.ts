export interface Server {
	url: string; // e.g. "https://chat.mysite.com/basepath" (no trailing slash)
	quickname: string;
	description?: string;
	color?: string;
	addedAt: number;
	isLocal?: boolean;
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
