import { createFileRoute } from "@tanstack/react-router";
import { ChatLayout } from "@/components/chat/ChatLayout";
import { getServers } from "@/lib/servers";

export const Route = createFileRoute("/chat/$roomName")({
	component: ChatRoomPage,
});

function ChatRoomPage() {
	const { roomName: roomId } = Route.useParams();
	const tildeIdx = roomId.indexOf("~");
	const serverHost = tildeIdx === -1 ? null : roomId.slice(0, tildeIdx);
	const roomName = tildeIdx === -1 ? roomId : roomId.slice(tildeIdx + 1);
	const servers = getServers();
	const server = serverHost
		? servers.find((s) => new URL(s.url).host === serverHost)
		: null;
	const serverUrl = server?.url ?? (serverHost ? `https://${serverHost}` : "");
	return <ChatLayout roomName={roomName} serverUrl={serverUrl} />;
}
