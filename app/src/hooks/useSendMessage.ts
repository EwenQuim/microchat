import { useQueryClient } from "@tanstack/react-query";
import {
	getGETApiRoomsRoomMessagesQueryKey,
	usePOSTApiRoomsRoomMessages,
} from "@/lib/api/generated/api/api";

export function useSendMessage(roomName: string) {
	const queryClient = useQueryClient();

	return usePOSTApiRoomsRoomMessages({
		mutation: {
			onSuccess: () => {
				// Invalidate messages query to refetch immediately
				queryClient.invalidateQueries({
					queryKey: getGETApiRoomsRoomMessagesQueryKey(roomName),
				});
			},
		},
	});
}
