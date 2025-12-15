import { useState, useEffect } from 'react';
import { storage } from '@/lib/storage';

export function useUsername() {
  const [username, setUsernameState] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const stored = storage.getUsername();
    setUsernameState(stored);
    setIsLoading(false);
  }, []);

  const setUsername = (name: string) => {
    storage.setUsername(name);
    setUsernameState(name);
  };

  const clearUsername = () => {
    storage.clearUsername();
    setUsernameState(null);
  };

  return { username, setUsername, clearUsername, isLoading };
}
