import { createFileRoute } from '@tanstack/react-router'
import { ChatLayout } from '@/components/chat/ChatLayout'

export const Route = createFileRoute('/chat/$roomName')({
  component: ChatRoomPage,
})

function ChatRoomPage() {
  const { roomName } = Route.useParams()
  return <ChatLayout roomName={roomName} />
}
