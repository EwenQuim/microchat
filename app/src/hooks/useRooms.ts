import { useQueries } from "@tanstack/react-query";
import { gETApiRooms } from "@/lib/api/generated/chat/chat";
import type { Room } from "@/lib/api/generated/openAPI.schemas";
import type { MutatorOptions } from "@/lib/api/mutator";
import { getStoredPasswords } from "./useSearchRooms";
import { useServers } from "./useServers";

export type RoomWithServer = Room & {
	serverUrl: string;
	serverQuickname: string;
	visited?: boolean;
};

function tagRooms(
	rooms: Room[],
	server: { url: string; quickname: string },
	visitedRooms: string,
): RoomWithServer[] {
	const result: RoomWithServer[] = rooms.map((r) => ({
		...r,
		serverUrl: server.url,
		serverQuickname: server.quickname,
	}));

	for (const visitedRoom of visitedRooms.split(",")) {
		if (!visitedRoom) continue;
		for (const room of result) {
			if (room.name === visitedRoom && room.serverUrl === server.url) {
				room.visited = true;
				break;
			}
		}
		if (!result.find((r) => r.name === visitedRoom)) {
			result.push({
				name: visitedRoom,
				visited: true,
				serverUrl: server.url,
				serverQuickname: server.quickname,
			});
		}
	}

	return result;
}

export function useRooms() {
	const { servers } = useServers();
	const visitedRooms = Object.keys(getStoredPasswords())?.join(",");

	const results = useQueries({
		queries: servers.map((server) => ({
			queryKey: ["rooms", server.url, visitedRooms],
			queryFn: async (): Promise<RoomWithServer[]> => {
				const params = { visited: visitedRooms || undefined };
				const opts: MutatorOptions = {
					baseUrl: server.isLocal ? "" : server.url,
				};
				const res = await gETApiRooms(params, opts);
				if (res.status !== 200) throw new Error(`${res.status}`);
				return tagRooms(res.data as Room[], server, visitedRooms);
			},
			refetchInterval: 30000,
			staleTime: 3000,
		})),
	});

	const isLoading = results.some((r) => r.isLoading);
	const error = results.find((r) => r.error)?.error ?? null;
	const data: RoomWithServer[] = results
		.flatMap((r) => r.data ?? [])
		.sort((a, b) => (b.visited ? 1 : 0) - (a.visited ? 1 : 0));

	return { data, isLoading, error };
}
