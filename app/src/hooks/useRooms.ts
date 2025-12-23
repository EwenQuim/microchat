import { useGETApiRooms } from "@/lib/api/generated/chat/chat";
import type { Room } from "@/lib/api/generated/openAPI.schemas";
import { getStoredPasswords } from "./useSearchRooms";

export function useRooms() {
	const visitedRooms = Object.keys(getStoredPasswords())?.join(",");

	const query = useGETApiRooms<Room[]>(
		{
			visited: visitedRooms || undefined,
		},
		{
			query: {
				refetchInterval: 30000,
				staleTime: 3000,
				select: (response) => {
					if (response.status === 200) {
						const onlineRooms = response.data;

						for (const visitedRoom of visitedRooms.split(",")) {
							for (const onlineRoom of onlineRooms) {
								if (onlineRoom.name === visitedRoom) {
									onlineRoom.visited = true;
									break;
								}
							}
							if (!onlineRooms.find((room) => room.name === visitedRoom)) {
								onlineRooms.push({
									name: visitedRoom,
									visited: true,
								});
							}
						}

						// Sort by visited first
						onlineRooms.sort((a, b) => (b.visited ? 1 : a.visited ? -1 : 0));

						return onlineRooms;
					}
					return [];
				},
			},
		},
	);

	return query;
}
