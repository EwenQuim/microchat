import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { getGETApiRoomsQueryKey } from "@/lib/api/generated/chat/chat";
import { getStoredPasswords } from "./useSearchRooms";

/**
 * Hook that tracks visited rooms and updates the query cache
 *
 * When a room is visited, it stores the room password (or empty string for public rooms)
 * and invalidates the rooms query to include the visited room in the list
 */
export function useTrackVisitedRoom(
	roomName: string | null,
	password: string | undefined,
	setPassword: (room: string, password: string) => void,
) {
	const queryClient = useQueryClient();

	useEffect(() => {
		if (!roomName) {
			return;
		}

		// Successfully accessed room - save to visited list
		// Store with current password (empty string for public rooms)
		setPassword(roomName, password || "");

		const visitedRooms = Object.keys(getStoredPasswords())?.join(",");
		queryClient.invalidateQueries({
			queryKey: getGETApiRoomsQueryKey({
				visited: visitedRooms || undefined,
			}),
		});
	}, [roomName, password, setPassword, queryClient]);
}
