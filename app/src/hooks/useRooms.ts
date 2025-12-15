import { useGETApiRooms } from "@/lib/api/generated/api/api";
import type { Room } from "@/lib/api/generated/openAPI.schemas";

export function useRooms() {
	return useGETApiRooms<Room[]>({
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
	});
}
