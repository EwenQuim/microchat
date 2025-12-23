import { useQueryClient } from "@tanstack/react-query";
import { usePOSTApiRoomsRoomMessages } from "@/lib/api/generated/chat/chat";

export function useSendMessage() {
	const queryClient = useQueryClient();

	return usePOSTApiRoomsRoomMessages({
		mutation: {
			onSettled: () => queryClient.invalidateQueries(),
		},
	});
}
