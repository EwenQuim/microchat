import { createFileRoute } from '@tanstack/react-router'
import { ChatLayout } from '@/components/chat/ChatLayout'

export const Route = createFileRoute('/')({
  component: HomePage,
})

function HomePage() {
  return <ChatLayout roomName={null} />
}
