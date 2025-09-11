import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from './useWebSocket';
import type {
  ChatRoom,
  ChatMessage,
  SendMessageRequest,
  CreateChatRoomRequest,
  ReactionType,
  TacticalWSMessage,
  ChatMessagePayload,
  ChatRoomPayload,
  MessageReactionPayload,
  UserTypingPayload,
  UserStatusPayload,
} from '../types/chat';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

interface UseChatReturn {
  // State
  rooms: ChatRoom[];
  messages: { [roomId: string]: ChatMessage[] };
  currentRoom: ChatRoom | null;
  typingUsers: { [roomId: string]: Set<string> };
  onlineUsers: Set<string>;
  loading: boolean;
  error: string | null;

  // Room operations
  createRoom: (room: CreateChatRoomRequest) => Promise<void>;
  joinRoom: (roomId: string) => Promise<void>;
  leaveRoom: (roomId: string) => Promise<void>;
  selectRoom: (roomId: string) => void;
  loadRooms: () => Promise<void>;

  // Message operations
  sendMessage: (message: SendMessageRequest) => Promise<void>;
  loadMessages: (roomId: string, limit?: number, offset?: number) => Promise<void>;
  acknowledgeMessage: (messageId: string) => Promise<void>;
  addReaction: (messageId: string, reactionType: ReactionType) => Promise<void>;

  // Typing indicators
  sendTypingIndicator: (roomId: string, typing: boolean) => void;

  // Connection state
  isConnected: boolean;
}

export function useChat(_currentUserId: string = 'anonymous'): UseChatReturn {
  const [rooms, setRooms] = useState<ChatRoom[]>([]);
  const [messages, setMessages] = useState<{ [roomId: string]: ChatMessage[] }>({});
  const [currentRoom, setCurrentRoom] = useState<ChatRoom | null>(null);
  const [typingUsers, setTypingUsers] = useState<{ [roomId: string]: Set<string> }>({});
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // WebSocket connection
  const { send: sendWsMessage, lastMessage, connected } = useWebSocket(`ws://localhost:8080/ws/tactical`);

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (!lastMessage) return;

    try {
      const message: TacticalWSMessage = lastMessage as any;
      
      switch (message.type) {
        case 'chat_message':
          handleChatMessage(message.payload as ChatMessagePayload);
          break;
        case 'chat_room_update':
          handleRoomUpdate(message.payload as ChatRoomPayload);
          break;
        case 'message_reaction':
          handleMessageReaction(message.payload as MessageReactionPayload);
          break;
        case 'user_typing':
          handleUserTyping(message.payload as UserTypingPayload);
          break;
        case 'user_online':
        case 'user_offline':
          handleUserStatus(message.payload as UserStatusPayload);
          break;
        default:
          // Handle other message types (position updates, etc.)
          break;
      }
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err);
    }
  }, [lastMessage]);

  // Handle chat message updates
  const handleChatMessage = useCallback((payload: ChatMessagePayload) => {
    const { message, action } = payload;
    
    setMessages(prev => {
      const roomMessages = prev[message.roomId] || [];
      
      switch (action) {
        case 'new':
          // Avoid duplicates
          if (roomMessages.find(m => m.id === message.id)) {
            return prev;
          }
          return {
            ...prev,
            [message.roomId]: [...roomMessages, message].sort((a, b) => 
              new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            )
          };
        case 'update':
          return {
            ...prev,
            [message.roomId]: roomMessages.map(m => 
              m.id === message.id ? message : m
            )
          };
        case 'delete':
          return {
            ...prev,
            [message.roomId]: roomMessages.filter(m => m.id !== message.id)
          };
        default:
          return prev;
      }
    });
  }, []);

  // Handle room updates
  const handleRoomUpdate = useCallback((payload: ChatRoomPayload) => {
    const { room, action } = payload;
    
    setRooms(prev => {
      switch (action) {
        case 'created':
          return [...prev, room];
        case 'updated':
          return prev.map(r => r.id === room.id ? room : r);
        default:
          return prev;
      }
    });
  }, []);

  // Handle message reactions
  const handleMessageReaction = useCallback((payload: MessageReactionPayload) => {
    const { messageId, reaction, action } = payload;
    
    setMessages(prev => {
      const newMessages = { ...prev };
      
      Object.keys(newMessages).forEach(roomId => {
        const roomMessages = newMessages[roomId];
        const messageIndex = roomMessages.findIndex(m => m.id === messageId);
        
        if (messageIndex !== -1) {
          const message = { ...roomMessages[messageIndex] };
          const reactions = message.reactions || [];
          
          if (action === 'added') {
            message.reactions = [...reactions, reaction];
          } else if (action === 'removed') {
            message.reactions = reactions.filter(r => 
              r.id !== reaction.id || r.userId !== reaction.userId
            );
          }
          
          newMessages[roomId] = [
            ...roomMessages.slice(0, messageIndex),
            message,
            ...roomMessages.slice(messageIndex + 1)
          ];
        }
      });
      
      return newMessages;
    });
  }, []);

  // Handle typing indicators
  const handleUserTyping = useCallback((payload: UserTypingPayload) => {
    const { roomId, userId, username, typing } = payload;
    
    setTypingUsers(prev => {
      const roomTyping = new Set(prev[roomId] || []);
      
      if (typing) {
        roomTyping.add(`${username}${userId}`);
      } else {
        roomTyping.delete(`${username}${userId}`);
      }
      
      return {
        ...prev,
        [roomId]: roomTyping
      };
    });
    
    // Clear typing indicator after 5 seconds
    if (typing) {
      setTimeout(() => {
        setTypingUsers(prev => {
          const roomTyping = new Set(prev[roomId] || []);
          roomTyping.delete(`${username}${userId}`);
          return {
            ...prev,
            [roomId]: roomTyping
          };
        });
      }, 5000);
    }
  }, []);

  // Handle user status changes
  const handleUserStatus = useCallback((payload: UserStatusPayload) => {
    const { userId, online } = payload;
    
    setOnlineUsers(prev => {
      const newSet = new Set(prev);
      if (online) {
        newSet.add(userId);
      } else {
        newSet.delete(userId);
      }
      return newSet;
    });
  }, []);

  // API helper function
  const apiCall = async (endpoint: string, options?: RequestInit) => {
    const response = await fetch(`${API_BASE_URL}/api/v1${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`API call failed: ${response.statusText}`);
    }

    return response.json();
  };

  // Room operations
  const createRoom = async (room: CreateChatRoomRequest) => {
    try {
      setLoading(true);
      const result = await apiCall('/chat/rooms', {
        method: 'POST',
        body: JSON.stringify(room),
      });
      
      // Room will be added via WebSocket update
      console.log('Room created:', result);
    } catch (err) {
      setError(`Failed to create room: ${err instanceof Error ? err.message : 'Unknown error'}`);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const joinRoom = async (roomId: string) => {
    if (sendWsMessage) {
      sendWsMessage(JSON.stringify({
        type: 'join_room',
        payload: { roomId }
      }));
    }
  };

  const leaveRoom = async (roomId: string) => {
    if (sendWsMessage) {
      sendWsMessage(JSON.stringify({
        type: 'leave_room',
        payload: { roomId }
      }));
    }
  };

  const selectRoom = (roomId: string) => {
    const room = rooms.find(r => r.id === roomId);
    if (room) {
      setCurrentRoom(room);
      joinRoom(roomId);
      
      // Load messages if we don't have them
      if (!messages[roomId]) {
        loadMessages(roomId);
      }
    }
  };

  const loadRooms = async () => {
    try {
      setLoading(true);
      const result = await apiCall('/chat/rooms');
      setRooms(result.rooms || []);
    } catch (err) {
      setError(`Failed to load rooms: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  // Message operations
  const sendMessage = async (message: SendMessageRequest) => {
    if (sendWsMessage) {
      sendWsMessage(JSON.stringify({
        type: 'send_message',
        payload: message
      }));
    }
  };

  const loadMessages = async (roomId: string, limit = 50, offset = 0) => {
    try {
      setLoading(true);
      const result = await apiCall(`/chat/rooms/${roomId}/messages?limit=${limit}&offset=${offset}`);
      
      setMessages(prev => ({
        ...prev,
        [roomId]: result.messages || []
      }));
    } catch (err) {
      setError(`Failed to load messages: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  const acknowledgeMessage = async (messageId: string) => {
    if (sendWsMessage) {
      sendWsMessage(JSON.stringify({
        type: 'acknowledge_message',
        payload: { messageId }
      }));
    }
  };

  const addReaction = async (messageId: string, reactionType: ReactionType) => {
    if (sendWsMessage) {
      sendWsMessage(JSON.stringify({
        type: 'react_to_message',
        payload: { messageId, reactionType }
      }));
    }
  };

  // Typing indicators
  const sendTypingIndicator = (roomId: string, typing: boolean) => {
    if (sendWsMessage) {
      const type = typing ? 'typing_start' : 'typing_stop';
      sendWsMessage(JSON.stringify({
        type,
        payload: { roomId }
      }));
    }
  };

  // Load rooms on mount
  useEffect(() => {
    loadRooms();
  }, []);

  return {
    // State
    rooms,
    messages,
    currentRoom,
    typingUsers,
    onlineUsers,
    loading,
    error,

    // Room operations
    createRoom,
    joinRoom,
    leaveRoom,
    selectRoom,
    loadRooms,

    // Message operations
    sendMessage,
    loadMessages,
    acknowledgeMessage,
    addReaction,

    // Typing indicators
    sendTypingIndicator,

    // Connection state
    isConnected: connected,
  };
}
