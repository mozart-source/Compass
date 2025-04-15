import { useEffect, useRef, useState } from 'react';
import { getApiUrls } from '@/config';

interface Notification {
  id: string;
  title: string;
  message?: string; // Optional for backward compatibility
  content?: string; // From server format
  type: 'info' | 'success' | 'warning' | 'error' | string;
  status?: string;
  createdAt: string;
  read: boolean;
}

export const useWebSocket = (token: string | null) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const ws = useRef<WebSocket | null>(null);
  const reconnectTimeout = useRef<NodeJS.Timeout | null>(null);

  const connect = () => {
    if (!token) return;

    const { WS_GO_URL } = getApiUrls();
    // Construct WebSocket URL for notifications
    const wsUrl = `${WS_GO_URL.replace(/^http/, 'ws')}/api/notifications/ws?token=${token}`;
    
    console.log('Connecting to WebSocket:', wsUrl);
    
    try {
      ws.current = new WebSocket(wsUrl);

      ws.current.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        // Clear any existing reconnect timeout
        if (reconnectTimeout.current) {
          clearTimeout(reconnectTimeout.current);
          reconnectTimeout.current = null;
        }
      };

      ws.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('WebSocket message received:', data);
          
          if (data.type === 'notification') {
            const notification: Notification = {
              id: data.id,
              title: data.title,
              message: data.message || data.content, // Handle both formats
              content: data.content,
              type: data.type || 'info',
              status: data.status,
              createdAt: data.createdAt || new Date().toISOString(),
              read: false
            };
            
            setNotifications(prev => [notification, ...prev]);
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      ws.current.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        // Only attempt to reconnect if we have a token and no existing timeout
        if (token && !reconnectTimeout.current) {
          console.log('WebSocket disconnected, attempting to reconnect...');
          reconnectTimeout.current = setTimeout(() => {
            reconnectTimeout.current = null;
            connect();
          }, 3000);
        }
      };

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setIsConnected(false);
      };

    } catch (error) {
      console.error('Error creating WebSocket:', error);
      setIsConnected(false);
    }
  };

  useEffect(() => {
    if (token) {
      connect();
    }

    return () => {
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
        reconnectTimeout.current = null;
      }
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [token]);

  const markAsRead = (id: string) => {
    setNotifications(prev => 
      prev.map(notif => 
        notif.id === id ? { ...notif, read: true } : notif
      )
    );
    
    // Optionally send to server
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        type: 'mark_read',
        id: id
      }));
    }
  };

  const clearNotifications = () => {
    setNotifications([]);
  };

  return {
    notifications,
    isConnected,
    markAsRead,
    clearNotifications
  };
}; 