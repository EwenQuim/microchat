import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Send } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MessageInputProps {
  onSend: (content: string) => void;
  disabled?: boolean;
  className?: string;
}

export function MessageInput({ onSend, disabled, className }: MessageInputProps) {
  const [content, setContent] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim() && !disabled) {
      onSend(content.trim());
      setContent('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className={cn('p-4', className)}>
      <div className="flex gap-2">
        <Input
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Type a message..."
          disabled={disabled}
          className="flex-1"
        />
        <Button type="submit" disabled={!content.trim() || disabled}>
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </form>
  );
}
