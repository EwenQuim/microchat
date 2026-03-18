export type ServerStatus = "active" | "hidden" | "advertise";

export interface Server {
	url: string; // domain or http://... (no https://, no trailing slash)
	quickname: string;
	description?: string;
	color?: string;
	addedAt: number;
	isLocal?: boolean;
	status?: ServerStatus; // omitted = "active"
	suggestedBy?: string; // URL of server that suggested this one
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
