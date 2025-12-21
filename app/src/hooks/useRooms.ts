import { useGETApiRooms } from "@/lib/api/generated/api/api";
import type { Room } from "@/lib/api/generated/openAPI.schemas";
import { getStoredPasswords } from "./useSearchRooms";

export function useRooms() {
	const visitedRooms = Object.keys(getStoredPasswords())?.join(",");
	return useGETApiRooms<Room[]>(
		{
			visited: visitedRooms || undefined,
		},
		{
			query: {
				refetchInterval: 30000,
				staleTime: 3000,
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
