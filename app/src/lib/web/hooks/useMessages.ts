import { useQuery } from "@tanstack/react-query";
import { gETApiRoomsRoomMessages } from "@/lib/api/generated/chat/chat";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import type { MutatorOptions } from "@/lib/api/mutator";

export function useMessages(
	roomName: string | null,
	password?: string,
	serverUrl = "",
) {
	return useQuery({
		queryKey: ["messages", serverUrl, roomName, password],
		queryFn: async (): Promise<Message[]> => {
			if (!roomName) throw new Error("No room");
			const opts: MutatorOptions = { baseUrl: serverUrl };
			const res = await gETApiRoomsRoomMessages(roomName, { password }, opts);
			if (res.status !== 200) throw new Error(`${res.status}`);
			return res.data as Message[];
		},
		enabled: !!roomName,
		refetchInterval: 10000,
		staleTime: 1000,
	});
}
