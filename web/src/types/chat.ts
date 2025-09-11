export type ChatRoomType = 'group' | 'private' | 'tactical' | 'emergency';

export type Classification = 'UNCLASSIFIED' | 'RESTRICTED' | 'CONFIDENTIAL' | 'SECRET' | 'TOP_SECRET';

export type MessageType = 'text' | 'system' | 'position' | 'emergency' | 'tactical_report';

export type MessagePriority = 'low' | 'normal' | 'high' | 'urgent' | 'emergency';

export type ParticipantRole = 'admin' | 'moderator' | 'member' | 'observer';

export type ReactionType = 'roger' | 'wilco' | 'negative' | 'like' | 'important' | 'question';

export interface ChatRoom {
  id: string;
  name: string;
  description?: string;
  type: ChatRoomType;
  classification: Classification;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
  isActive: boolean;
  missionId?: string;
  geographicBounds?: Record<string, any>;
  settings: Record<string, any>;
  participants?: ChatRoomParticipant[];
  participantCount: number;
  lastMessage?: ChatMessage;
  unreadCount: number;
}

export interface ChatRoomParticipant {
  id: string;
  roomId: string;
  userId?: string;
  callsign?: string;
  role: ParticipantRole;
  permissions: Record<string, any>;
  joinedAt: string;
  lastSeen: string;
  isActive: boolean;
  username?: string;
  displayName?: string;
}

export interface ChatMessage {
  id: string;
  roomId: string;
  senderId?: string;
  senderCallsign?: string;
  messageText: string;
  messageType: MessageType;
  priority: MessagePriority;
  cotEventUid?: string;
  cotEventType?: string;
  cotRawXml?: string;
  locationLat?: number;
  locationLng?: number;
  locationAlt?: number;
  classification: Classification;
  keywords?: string[];
  replyToId?: string;
  threadId?: string;
  isDeleted: boolean;
  deletedAt?: string;
  deletedBy?: string;
  createdAt: string;
  updatedAt: string;
  requiresAck: boolean;
  senderUsername?: string;
  senderDisplayName?: string;
  replyTo?: ChatMessage;
  reactions?: MessageReaction[];
  acknowledgments?: MessageAcknowledgment[];
  isAcknowledged: boolean;
}

export interface MessageReaction {
  id: string;
  messageId: string;
  userId: string;
  reactionType: ReactionType;
  createdAt: string;
  username?: string;
  callsign?: string;
}

export interface MessageAcknowledgment {
  id: string;
  messageId: string;
  userId: string;
  acknowledgedAt: string;
  username?: string;
  callsign?: string;
}

export interface CreateChatRoomRequest {
  name: string;
  description?: string;
  type: ChatRoomType;
  classification: Classification;
  missionId?: string;
  geographicBounds?: Record<string, any>;
  settings?: Record<string, any>;
}

export interface SendMessageRequest {
  roomId: string;
  messageText: string;
  messageType?: MessageType;
  priority?: MessagePriority;
  classification?: Classification;
  keywords?: string[];
  locationLat?: number;
  locationLng?: number;
  locationAlt?: number;
  replyToId?: string;
  requiresAck?: boolean;
}

export interface GetMessagesRequest {
  roomId: string;
  limit?: number;
  offset?: number;
  beforeId?: string;
  afterId?: string;
  messageType?: MessageType;
  priority?: MessagePriority;
  classification?: Classification;
  keywords?: string[];
  startTime?: string;
  endTime?: string;
}

export interface ChatStatistics {
  totalRooms: number;
  activeRooms: number;
  totalMessages: number;
  messagesToday: number;
  activeUsers: number;
  emergencyMessages: number;
  unacknowledgedCount: number;
}

// WebSocket message types
export interface TacticalWSMessage {
  type: string;
  payload: any;
  timestamp: string;
  roomId?: string;
}

export interface ChatMessagePayload {
  message: ChatMessage;
  action: 'new' | 'update' | 'delete';
}

export interface ChatRoomPayload {
  room: ChatRoom;
  action: 'created' | 'updated' | 'joined' | 'left';
}

export interface MessageReactionPayload {
  messageId: string;
  reaction: MessageReaction;
  action: 'added' | 'removed';
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

// Component props interfaces
export interface ChatPanelProps {
  isVisible: boolean;
  onToggle: () => void;
}

export interface MessageListProps {
  roomId: string;
  messages: ChatMessage[];
  currentUserId: string;
  onReaction: (messageId: string, reactionType: ReactionType) => void;
  onAcknowledge: (messageId: string) => void;
  onReply: (message: ChatMessage) => void;
}

export interface MessageInputProps {
  roomId: string;
  onSendMessage: (message: SendMessageRequest) => void;
  replyTo?: ChatMessage;
  onCancelReply?: () => void;
  disabled?: boolean;
}

export interface RoomsListProps {
  rooms: ChatRoom[];
  selectedRoomId?: string;
  onSelectRoom: (roomId: string) => void;
  onCreateRoom: () => void;
  onLeaveRoom: (roomId: string) => void;
}

export interface UsersListProps {
  participants: ChatRoomParticipant[];
  currentUserId: string;
  onRemoveUser?: (userId: string) => void;
  onChangeRole?: (userId: string, role: ParticipantRole) => void;
}
