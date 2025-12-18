import { useQuery } from "@tanstack/react-query";
import type { Room } from "@/lib/api/generated/openAPI.schemas";

export function useSearchRooms(query: string) {
	return useQuery({
		queryKey: ["rooms", "search", query],
		queryFn: async () => {
			if (!query.trim()) {
				return [];
			}
			const url = new URL("/api/rooms/search", window.location.origin);
			url.searchParams.set("q", query);

			const response = await fetch(url.toString());
			if (!response.ok) {
				throw new Error("Failed to search rooms");
			}
			const data: Room[] = await response.json();
			return data || [];
		},
		enabled: query.trim().length > 0,
		staleTime: 30_000,
	});
}
