import { useGETApiRoomsRoomMessages } from "@/lib/api/generated/chat/chat";
import type { Message } from "@/lib/api/generated/openAPI.schemas";

export function useMessages(roomName: string | null, password?: string) {
	return useGETApiRoomsRoomMessages<Message[]>(
		roomName ?? "",
		{
			password,
		},
		{
			query: {
				enabled: !!roomName,
				refetchInterval: 10000,
				staleTime: 1000,
				select: (response) => {
					if (response.status === 200) {
						return response.data as Message[];
					}
					return [];
				},
			},
		},
	);
}
