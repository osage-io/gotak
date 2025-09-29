/**
 * GoTAK WebSocket Service
 * Provides real-time communication with the GoTAK server via WebSocket
 */

// Type definition for runtime configuration
declare global {
  interface Window {
    GOTAK_CONFIG?: {
      apiUrl?: string;
      wsUrl?: string;
      serverUrl?: string;
      [key: string]: any;
    };
  }
}

import { Entity, PositionPoint, ChatMessage, ChatRoom } from './apiClient';

// WebSocket Configuration
const WS_BASE_URL = window.GOTAK_CONFIG?.wsUrl || 'ws://localhost:8087/ws';
const WS_ENDPOINT = '/tactical';

// Helper function to normalize WebSocket URL
const normalizeWSUrl = (baseUrl: string, endpoint: string): string => {
  // Remove trailing slash from base URL
  const normalizedBase = baseUrl.replace(/\/$/, '');
  // Ensure endpoint starts with slash
  const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  return `${normalizedBase}${normalizedEndpoint}`;
};

// WebSocket Message Types (matching backend)
export const WS_MESSAGE_TYPES = {
  // Position and entity messages
  POSITION_UPDATE: 'position_update',
  ENTITY_REMOVED: 'entity_removed',
  
  // Chat messages
  CHAT_MESSAGE: 'chat_message',
  CHAT_ROOM_UPDATE: 'chat_room_update',
  CHAT_ROOM_JOINED: 'chat_room_joined',
  CHAT_ROOM_LEFT: 'chat_room_left',
  MESSAGE_REACTION: 'message_reaction',
  MESSAGE_ACK: 'message_acknowledgment',
  USER_TYPING: 'user_typing',
  USER_ONLINE: 'user_online',
  USER_OFFLINE: 'user_offline',
  
  // System messages
  SYSTEM_ALERT: 'system_alert',
  HEARTBEAT: 'heartbeat',
  ERROR: 'error',
} as const;

export type WSMessageType = typeof WS_MESSAGE_TYPES[keyof typeof WS_MESSAGE_TYPES];

// WebSocket Message Structure
export interface WSMessage {
  type: WSMessageType;
  payload: any;
  timestamp: string;
  roomId?: string;
}

// Payload Types
export interface PositionUpdate {
  entityId: string;
  position: PositionPoint;
}

export interface ChatMessagePayload {
  message: ChatMessage;
  action: 'new' | 'update' | 'delete';
}

export interface ChatRoomPayload {
  room: ChatRoom;
  action: 'created' | 'updated' | 'joined' | 'left';
}

export interface UserTypingPayload {
  roomId: string;
  userId: string;
  username: string;
  callsign?: string;
  typing: boolean;
}

export interface UserStatusPayload {
  userId: string;
  username: string;
  callsign?: string;
  online: boolean;
  lastSeen: string;
}

export interface SystemAlert {
  id: string;
  type: 'info' | 'warning' | 'error' | 'critical';
  title: string;
  message: string;
  timestamp: string;
  requiresAck?: boolean;
}

// Event Handlers
export type WSEventHandler<T = any> = (payload: T) => void;
export type WSErrorHandler = (error: Event) => void;
export type WSConnectionHandler = () => void;

// WebSocket Connection States
export enum ConnectionState {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error'
}

// WebSocket Service Class
export class GoTAKWebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private reconnectTimer: number | null = null;
  private heartbeatTimer: number | null = null;
  private heartbeatInterval = 30000; // 30 seconds
  
  // Event handlers
  private eventHandlers: Map<WSMessageType, Set<WSEventHandler>> = new Map();
  private connectionHandlers: Set<WSConnectionHandler> = new Set();
  private disconnectionHandlers: Set<WSConnectionHandler> = new Set();
  private errorHandlers: Set<WSErrorHandler> = new Set();
  
  // State
  private _connectionState: ConnectionState = ConnectionState.DISCONNECTED;
  private lastHeartbeat: number = 0;
  
  constructor(baseUrl: string = WS_BASE_URL) {
    this.url = normalizeWSUrl(baseUrl, WS_ENDPOINT);
    
    // Initialize event handler maps
    Object.values(WS_MESSAGE_TYPES).forEach(type => {
      this.eventHandlers.set(type, new Set());
    });
  }

  // ===== CONNECTION MANAGEMENT =====
  
  get connectionState(): ConnectionState {
    return this._connectionState;
  }
  
  get isConnected(): boolean {
    return this._connectionState === ConnectionState.CONNECTED;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
        resolve();
        return;
      }

      this._connectionState = ConnectionState.CONNECTING;
      console.log('🔌 Connecting to GoTAK WebSocket:', this.url);

      try {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = (event) => {
          console.log('✅ WebSocket connected to GoTAK server');
          this._connectionState = ConnectionState.CONNECTED;
          this.reconnectAttempts = 0;
          this.reconnectDelay = 1000;
          
          this.startHeartbeat();
          this.connectionHandlers.forEach(handler => handler());
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

        this.ws.onerror = (error) => {
          console.error('❌ WebSocket error:', error);
          this._connectionState = ConnectionState.ERROR;
          this.errorHandlers.forEach(handler => handler(error));
          reject(new Error('WebSocket connection failed'));
        };

        this.ws.onclose = (event) => {
          console.log('🔌 WebSocket connection closed:', event.code, event.reason);
          this._connectionState = ConnectionState.DISCONNECTED;
          this.stopHeartbeat();
          this.disconnectionHandlers.forEach(handler => handler());
          
          // Attempt to reconnect unless explicitly closed
          if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

      } catch (error) {
        console.error('❌ Failed to create WebSocket connection:', error);
        this._connectionState = ConnectionState.ERROR;
        reject(error);
      }
    });
  }

  disconnect(): void {
    console.log('🔌 Disconnecting from GoTAK WebSocket');
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    this.stopHeartbeat();
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    this._connectionState = ConnectionState.DISCONNECTED;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    
    this._connectionState = ConnectionState.RECONNECTING;
    this.reconnectAttempts++;
    
    console.log(`🔄 Scheduling WebSocket reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${this.reconnectDelay}ms`);
    
    this.reconnectTimer = window.setTimeout(() => {
      this.connect().catch((error) => {
        console.error('❌ Reconnect attempt failed:', error);
        
        // Exponential backoff
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.scheduleReconnect();
        } else {
          console.error('❌ Max reconnect attempts reached. Giving up.');
          this._connectionState = ConnectionState.ERROR;
        }
      });
    }, this.reconnectDelay);
  }

  // ===== HEARTBEAT MANAGEMENT =====
  
  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.lastHeartbeat = Date.now();
    
    this.heartbeatTimer = window.setInterval(() => {
      if (this.isConnected) {
        this.send({
          type: WS_MESSAGE_TYPES.HEARTBEAT,
          payload: { timestamp: Date.now() },
          timestamp: new Date().toISOString(),
        });
      }
    }, this.heartbeatInterval);
  }
  
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  // ===== MESSAGE HANDLING =====
  
  private handleMessage(event: MessageEvent): void {
    try {
      const message: WSMessage = JSON.parse(event.data);
      
      // Update heartbeat timestamp for heartbeat messages
      if (message.type === WS_MESSAGE_TYPES.HEARTBEAT) {
        this.lastHeartbeat = Date.now();
      }
      
      // Dispatch to registered handlers
      const handlers = this.eventHandlers.get(message.type);
      if (handlers) {
        handlers.forEach(handler => {
          try {
            handler(message.payload);
          } catch (error) {
            console.error(`Error in WebSocket handler for ${message.type}:`, error);
          }
        });
      } else {
        console.warn('Unhandled WebSocket message type:', message.type);
      }
      
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error, event.data);
    }
  }

  send(message: Omit<WSMessage, 'timestamp'>): boolean {
    if (!this.isConnected || !this.ws) {
      console.warn('Cannot send WebSocket message - not connected');
      return false;
    }

    try {
      const fullMessage: WSMessage = {
        ...message,
        timestamp: new Date().toISOString(),
      };
      
      this.ws.send(JSON.stringify(fullMessage));
      return true;
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
      return false;
    }
  }

  // ===== EVENT HANDLERS =====
  
  on<T = any>(messageType: WSMessageType, handler: WSEventHandler<T>): () => void {
    const handlers = this.eventHandlers.get(messageType);
    if (handlers) {
      handlers.add(handler);
    }
    
    // Return unsubscribe function
    return () => {
      const handlers = this.eventHandlers.get(messageType);
      if (handlers) {
        handlers.delete(handler);
      }
    };
  }

  onConnection(handler: WSConnectionHandler): () => void {
    this.connectionHandlers.add(handler);
    return () => this.connectionHandlers.delete(handler);
  }

  onDisconnection(handler: WSConnectionHandler): () => void {
    this.disconnectionHandlers.add(handler);
    return () => this.disconnectionHandlers.delete(handler);
  }

  onError(handler: WSErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  // ===== CONVENIENCE METHODS =====
  
  onPositionUpdate(handler: WSEventHandler<PositionUpdate>): () => void {
    return this.on(WS_MESSAGE_TYPES.POSITION_UPDATE, handler);
  }

  onEntityRemoved(handler: WSEventHandler<{ entityId: string }>): () => void {
    return this.on(WS_MESSAGE_TYPES.ENTITY_REMOVED, handler);
  }

  onChatMessage(handler: WSEventHandler<ChatMessagePayload>): () => void {
    return this.on(WS_MESSAGE_TYPES.CHAT_MESSAGE, handler);
  }

  onChatRoomUpdate(handler: WSEventHandler<ChatRoomPayload>): () => void {
    return this.on(WS_MESSAGE_TYPES.CHAT_ROOM_UPDATE, handler);
  }

  onUserTyping(handler: WSEventHandler<UserTypingPayload>): () => void {
    return this.on(WS_MESSAGE_TYPES.USER_TYPING, handler);
  }

  onSystemAlert(handler: WSEventHandler<SystemAlert>): () => void {
    return this.on(WS_MESSAGE_TYPES.SYSTEM_ALERT, handler);
  }

  // ===== SEND CONVENIENCE METHODS =====
  
  sendChatMessage(roomId: string, messageText: string, options: {
    messageType?: ChatMessage['messageType'];
    priority?: ChatMessage['priority'];
    classification?: ChatMessage['classification'];
    requiresAck?: boolean;
    locationLat?: number;
    locationLng?: number;
    replyToId?: string;
  } = {}): boolean {
    return this.send({
      type: WS_MESSAGE_TYPES.CHAT_MESSAGE,
      payload: {
        roomId,
        messageText,
        ...options,
      },
      roomId,
    });
  }

  joinChatRoom(roomId: string): boolean {
    return this.send({
      type: WS_MESSAGE_TYPES.CHAT_ROOM_JOINED,
      payload: { roomId },
      roomId,
    });
  }

  leaveChatRoom(roomId: string): boolean {
    return this.send({
      type: WS_MESSAGE_TYPES.CHAT_ROOM_LEFT,
      payload: { roomId },
      roomId,
    });
  }

  setTyping(roomId: string, typing: boolean): boolean {
    return this.send({
      type: WS_MESSAGE_TYPES.USER_TYPING,
      payload: { roomId, typing },
      roomId,
    });
  }

  acknowledgeMessage(messageId: string): boolean {
    return this.send({
      type: WS_MESSAGE_TYPES.MESSAGE_ACK,
      payload: { messageId },
    });
  }
}

// Export singleton instance
export const wsService = new GoTAKWebSocketService();

// Export default instance for convenience
export default wsService;
