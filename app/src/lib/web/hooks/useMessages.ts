import { useQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";
import { gETApiRoomsRoomMessages } from "@/lib/api/generated/chat/chat";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import type { MutatorOptions } from "@/lib/api/mutator";

const DEFAULT_LIMIT = 50;

export function useMessages(
	roomName: string | null,
	password?: string,
	serverUrl = "",
) {
	const [messages, setMessages] = useState<Message[]>([]);
	const [oldestTimestamp, setOldestTimestamp] = useState<string | null>(null);
	const [hasMore, setHasMore] = useState(true);
	const [isLoadingOlder, setIsLoadingOlder] = useState(false);
	const [loadOlderCursor, setLoadOlderCursor] = useState<string | null>(null);
	// Track room changes to reset state
	const roomRef = useRef<string | null>(null);

	// Reset accumulated state when room changes
	if (roomRef.current !== roomName) {
		roomRef.current = roomName;
		setMessages([]);
		setOldestTimestamp(null);
		setHasMore(true);
		setIsLoadingOlder(false);
		setLoadOlderCursor(null);
	}

	// Poll for latest messages (no before param = gets most recent 50)
	const { data: pollData, isLoading } = useQuery({
		queryKey: ["messages", serverUrl, roomName, password],
		queryFn: async (): Promise<Message[]> => {
			if (!roomName) throw new Error("No room");
			const opts: MutatorOptions = { baseUrl: serverUrl };
			const res = await gETApiRoomsRoomMessages(
				roomName,
				{ password, limit: DEFAULT_LIMIT },
				opts,
			);
			if (res.status !== 200) throw new Error(`${res.status}`);
			return res.data as Message[];
		},
		enabled: !!roomName,
		refetchInterval: 10000,
		staleTime: 1000,
	});

	// Merge polled data into accumulated messages
	useEffect(() => {
		if (!pollData) return;
		setMessages((prev) => {
			const existingIds = new Set(prev.map((m) => m.id));
			const newMsgs = pollData.filter((m) => !existingIds.has(m.id));
			if (newMsgs.length === 0) return prev;
			return [...prev, ...newMsgs];
		});
	}, [pollData]);

	// Update oldestTimestamp after messages update
	useEffect(() => {
		if (messages.length > 0 && oldestTimestamp === null) {
			setOldestTimestamp(messages[0].timestamp ?? null);
		}
	}, [messages, oldestTimestamp]);

	// Fetch older messages query (only runs when cursor is set)
	const { data: olderData } = useQuery({
		queryKey: [
			"messages-older",
			serverUrl,
			roomName,
			password,
			loadOlderCursor,
		],
		queryFn: async (): Promise<Message[]> => {
			if (!roomName || !loadOlderCursor) throw new Error("No cursor");
			const opts: MutatorOptions = { baseUrl: serverUrl };
			const res = await gETApiRoomsRoomMessages(
				roomName,
				{ password, limit: DEFAULT_LIMIT, before: loadOlderCursor },
				opts,
			);
			if (res.status !== 200) throw new Error(`${res.status}`);
			return res.data as Message[];
		},
		enabled: !!roomName && !!loadOlderCursor,
	});

	// Handle older messages result
	useEffect(() => {
		if (!olderData || !loadOlderCursor) return;
		setMessages((prev) => {
			const existingIds = new Set(prev.map((m) => m.id));
			const newMsgs = olderData.filter((m) => !existingIds.has(m.id));
			if (newMsgs.length === 0) return prev;
			return [...newMsgs, ...prev];
		});
		if (olderData.length < DEFAULT_LIMIT) {
			setHasMore(false);
		}
		if (olderData.length > 0) {
			setOldestTimestamp(olderData[0].timestamp ?? null);
		}
		setLoadOlderCursor(null);
		setIsLoadingOlder(false);
	}, [olderData, loadOlderCursor]);

	const loadOlder = useCallback(() => {
		if (!hasMore || isLoadingOlder || !oldestTimestamp) return;
		setIsLoadingOlder(true);
		setLoadOlderCursor(oldestTimestamp);
	}, [hasMore, isLoadingOlder, oldestTimestamp]);

	return { messages, isLoading, isLoadingOlder, hasMore, loadOlder };
}
