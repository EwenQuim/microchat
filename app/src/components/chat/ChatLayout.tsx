import { useState, useEffect } from 'react';
import { RoomsSidebar } from './RoomsSidebar';
import { ChatArea } from './ChatArea';
import { useUsername } from '@/hooks/useUsername';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

interface ChatLayoutProps {
  roomName: string | null;
}

export function ChatLayout({ roomName }: ChatLayoutProps) {
  const { username, setUsername, isLoading } = useUsername();
  const [showUsernameDialog, setShowUsernameDialog] = useState(false);
  const [tempUsername, setTempUsername] = useState('');

  useEffect(() => {
    if (!isLoading && !username) {
      setShowUsernameDialog(true);
    }
  }, [isLoading, username]);

  const handleSaveUsername = () => {
    if (tempUsername.trim()) {
      setUsername(tempUsername.trim());
      setShowUsernameDialog(false);
    }
  };

  if (isLoading) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>;
  }

  return (
    <>
      <div className="flex h-[calc(100vh-64px)]">
        <RoomsSidebar
          selectedRoom={roomName}
          className="w-64 flex-shrink-0"
        />
        <ChatArea
          roomName={roomName}
          username={username!}
          className="flex-1"
        />
      </div>

      <Dialog open={showUsernameDialog} onOpenChange={setShowUsernameDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Welcome to MicroChat</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Please enter your name to start chatting
            </p>
            <Input
              value={tempUsername}
              onChange={(e) => setTempUsername(e.target.value)}
              placeholder="Your name"
              onKeyDown={(e) => e.key === 'Enter' && handleSaveUsername()}
            />
            <Button onClick={handleSaveUsername} className="w-full">
              Start Chatting
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
