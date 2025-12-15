import { useEffect, useRef } from 'react';
import { MessageItem } from './MessageItem';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { Message } from '@/lib/api/generated/openAPI.schemas';
import { cn } from '@/lib/utils';

interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
  currentUsername: string;
  className?: string;
}

export function MessageList({ messages, isLoading, currentUsername, className }: MessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (shouldAutoScroll.current && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    const isAtBottom = target.scrollHeight - target.scrollTop <= target.clientHeight + 50;
    shouldAutoScroll.current = isAtBottom;
  };

  return (
    <ScrollArea className={cn('p-4', className)} onScroll={handleScroll}>
      <div ref={scrollRef}>
        {isLoading && messages.length === 0 && (
          <div className="text-center text-muted-foreground">Loading messages...</div>
        )}

        {!isLoading && messages.length === 0 && (
          <div className="text-center text-muted-foreground">
            No messages yet. Be the first to say something!
          </div>
        )}

        <div className="space-y-4">
          {messages.map((message) => (
            <MessageItem
              key={message.id}
              message={message}
              isOwn={message.user === currentUsername}
            />
          ))}
        </div>
      </div>
    </ScrollArea>
  );
}
