import { useGETApiRoomsRoomMessages } from "@/lib/api/generated/api/api";
import type { Message } from "@/lib/api/generated/openAPI.schemas";

export function useMessages(roomName: string | null) {
	return useGETApiRoomsRoomMessages<Message[]>(roomName ?? "", {
		query: {
			enabled: !!roomName,
			refetchInterval: 3000, // Poll every 2 seconds
			staleTime: 1000,
			select: (response) => {
				if (response.status === 200) {
					return response.data as Message[];
				}
				return [];
			},
		},
	});
}
