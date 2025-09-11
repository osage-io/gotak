import { useState, useEffect, useRef, useCallback } from 'react';
import type { TacticalWSMessage, PositionUpdate } from '../types/position';
import { WS_MESSAGE_TYPES } from '../types/position';

// Environment variables
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true';
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws/tactical';
const DEBUG = import.meta.env.VITE_DEBUG === 'true';

export interface WebSocketMessage {
  type: string;
  data?: any;
  timestamp?: string;
}

export interface WebSocketState {
  connected: boolean;
  connecting: boolean;
  error: string | null;
  lastMessage: TacticalWSMessage | null;
}

export interface WebSocketOptions {
  onPositionUpdate?: (update: PositionUpdate) => void;
  onEntityRemoved?: (entityId: string) => void;
  onMessage?: (message: TacticalWSMessage) => void;
}

export interface WebSocketHook extends WebSocketState {
  send: (message: any) => void;
  reconnect: () => void;
}

export const useWebSocket = (url?: string, options: WebSocketOptions = {}): WebSocketHook => {
  const { onPositionUpdate, onEntityRemoved, onMessage } = options;
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastMessage, setLastMessage] = useState<TacticalWSMessage | null>(null);
  
  const ws = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<number | null>(null);
  const reconnectAttempts = useRef(0);
  
  // Use environment variable or provided URL
  const wsUrl = url || WS_URL;
  
  const connect = useCallback(() => {
    if (connecting || connected) return;
    
    if (USE_MOCK) {
      // Use mock WebSocket service
      if (DEBUG) console.log('Using mock WebSocket service');
      setConnected(true);
      setConnecting(false);
      setError(null);
      return;
    }
    
    setConnecting(true);
    setError(null);
    
    try {
      ws.current = new WebSocket(wsUrl);
      
      ws.current.onopen = () => {
        if (DEBUG) console.log('WebSocket connected to', wsUrl);
        setConnected(true);
        setConnecting(false);
        setError(null);
        reconnectAttempts.current = 0;
      };
      
      ws.current.onmessage = (event) => {
        try {
          const message: TacticalWSMessage = JSON.parse(event.data);
          setLastMessage(message);
          
          // Handle different message types
          switch (message.type) {
            case WS_MESSAGE_TYPES.POSITION_UPDATE:
              onPositionUpdate?.(message.payload as PositionUpdate);
              break;
            case WS_MESSAGE_TYPES.ENTITY_REMOVED:
              onEntityRemoved?.(message.payload.entityId);
              break;
            case WS_MESSAGE_TYPES.HEARTBEAT:
              // Respond to heartbeat to keep connection alive
              if (ws.current?.readyState === WebSocket.OPEN) {
                ws.current.send(JSON.stringify({
                  type: 'heartbeat_response',
                  timestamp: new Date().toISOString()
                }));
              }
              break;
            case WS_MESSAGE_TYPES.ERROR:
              console.error('WebSocket error message:', message.payload);
              setError(message.payload.error || 'Unknown error');
              break;
            default:
              // Handle other message types
              if (DEBUG) console.log('Unhandled WebSocket message type:', message.type);
              break;
          }
          
          // Call general message handler
          onMessage?.(message);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', event.data, err);
        }
      };
      
      ws.current.onerror = (event) => {
        console.error('WebSocket error:', event);
        setError('Connection error occurred');
        setConnecting(false);
      };
      
      ws.current.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        setConnected(false);
        setConnecting(false);
        
        // Auto-reconnect with exponential backoff
        if (event.code !== 1000) { // Not a normal closure
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
          console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current + 1})`);
          
          reconnectTimer.current = setTimeout(() => {
            reconnectAttempts.current += 1;
            connect();
          }, delay);
        }
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to connect to server');
      setConnecting(false);
    }
  }, [wsUrl, connecting, connected]);
  
  const send = useCallback((message: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected, cannot send message:', message);
    }
  }, []);
  
  const reconnect = useCallback(() => {
    if (ws.current) {
      ws.current.close();
    }
    reconnectAttempts.current = 0;
    connect();
  }, [connect]);
  
  // Connect on mount
  useEffect(() => {
    connect();
    
    return () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
      if (ws.current) {
        ws.current.close(1000, 'Component unmounting');
      }
    };
  }, [connect]);
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
      if (ws.current) {
        ws.current.close(1000, 'Component unmounting');
      }
    };
  }, []);
  
  return {
    connected,
    connecting,
    error,
    lastMessage,
    send,
    reconnect
  };
};
