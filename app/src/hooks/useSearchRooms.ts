import { useGETApiRoomsSearch } from "@/lib/api/generated/chat/chat";
import type { Room } from "@/types/room";

export function getStoredPasswords(): Record<string, string> {
	try {
		const stored = sessionStorage.getItem("microchat_room_passwords");
		return stored ? JSON.parse(stored) : {};
	} catch {
		return {};
	}
}

export function useSearchRooms(query: string) {
	const visitedRooms = Object.keys(getStoredPasswords()).join(",");

	return useGETApiRoomsSearch<Room[]>(
		{
			q: query,
			visited: visitedRooms || undefined,
		},
		{
			query: {
				enabled: query.trim().length > 0,
				staleTime: 30_000,
				select: (response) => {
					if (response.status === 200) {
						return response.data as Room[];
					}
					return [];
				},
			},
		},
	);
}
