/**
 * GoTAK API Client Service
 * Provides HTTP client interface to the GoTAK backend server
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

// API Configuration
const API_BASE_URL = window.GOTAK_CONFIG?.apiUrl || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// Type definitions matching backend structures
export interface Entity {
  id: string;
  callsign: string;
  type: string;
  entityType: string; // CoT entity type (e.g. 'a-f-G', 'a-h-G')
  affiliation: 'friendly' | 'hostile' | 'neutral' | 'unknown';
  lat: number;
  lng: number;
  position?: {
    lat: number;
    lng: number;
    altitude?: number;
    speed?: number;
    course?: number;
    timestamp: string;
  };
  altitude?: number;
  speed?: number;
  course?: number;
  lastUpdate: string; // Make this required as the code expects it
  classification: string;
  status: string;
}

export interface PositionPoint {
  lat: number;
  lng: number;
  altitude?: number;
  speed?: number;
  course?: number;
  timestamp: string;
}

export interface EntityHistory {
  entityId: string;
  callsign: string;
  positions: PositionPoint[];
  totalCount: number;
}

export interface EntitiesResponse {
  entities: Entity[];
  count: number;
}

export interface ChatRoom {
  id: string;
  name: string;
  description?: string;
  type: 'public' | 'private' | 'tactical';
  classification: 'UNCLASSIFIED' | 'RESTRICTED' | 'CONFIDENTIAL' | 'SECRET' | 'TOP_SECRET';
  createdAt: string;
  updatedAt: string;
  participantCount: number;
}

export interface ChatMessage {
  id: string;
  roomId: string;
  senderId: string;
  senderCallsign: string;
  messageText: string;
  messageType: 'text' | 'image' | 'location' | 'alert' | 'system';
  priority: 'routine' | 'priority' | 'immediate' | 'flash';
  classification: 'UNCLASSIFIED' | 'RESTRICTED' | 'CONFIDENTIAL' | 'SECRET' | 'TOP_SECRET';
  timestamp: string;
  requiresAck: boolean;
  acknowledged: boolean;
  locationLat?: number;
  locationLng?: number;
  replyToId?: string;
}

export interface PositionStatistics {
  totalEntities: number;
  friendlyCount: number;
  hostileCount: number;
  neutralCount: number;
  unknownCount: number;
  lastUpdate: string;
}

// HTTP Client Class
export class GoTAKAPIClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${API_PREFIX}${endpoint}`;
    
    const defaultOptions: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    };

    const finalOptions = { ...defaultOptions, ...options };

    try {
      const response = await fetch(url, finalOptions);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error(`API request failed for ${endpoint}:`, error);
      throw error;
    }
  }

  // ===== ENTITY METHODS =====
  
  async getAllEntities(): Promise<EntitiesResponse> {
    return this.request<EntitiesResponse>('/entities');
  }

  // Alias for compatibility with hooks that expect getEntities method
  async getEntities(): Promise<Entity[]> {
    try {
      const response = await this.getAllEntities();
      return response.entities || [];
    } catch (error) {
      console.error('Failed to get entities:', error);
      return [];
    }
  }

  async createEntity(entity: Entity): Promise<Entity> {
    return this.request<Entity>('/entities', {
      method: 'POST',
      body: JSON.stringify(entity),
    });
  }

  async updateEntity(id: string, entity: Entity): Promise<Entity> {
    return this.request<Entity>(`/entities/${id}`, {
      method: 'PUT',
      body: JSON.stringify(entity),
    });
  }

  async deleteEntity(id: string): Promise<void> {
    return this.request<void>(`/entities/${id}`, { method: 'DELETE' });
  }

  async getEntity(id: string): Promise<Entity> {
    return this.request<Entity>(`/entities/${id}`);
  }

  async getEntityHistory(id: string, timeRange?: string, limit?: number): Promise<EntityHistory> {
    const params = new URLSearchParams();
    if (timeRange) params.append('timeRange', timeRange);
    if (limit) params.append('limit', limit.toString());
    
    const query = params.toString() ? `?${params.toString()}` : '';
    return this.request<EntityHistory>(`/entities/${id}/history${query}`);
  }

  // ===== POSITION METHODS =====

  async getAllPositions(): Promise<Entity[]> {
    return this.request<Entity[]>('/positions');
  }

  async getActivePositions(): Promise<Entity[]> {
    return this.request<Entity[]>('/positions/active');
  }

  async getFriendlyPositions(): Promise<Entity[]> {
    return this.request<Entity[]>('/positions/friendly');
  }

  async getHostilePositions(): Promise<Entity[]> {
    return this.request<Entity[]>('/positions/hostile');
  }

  async getPositionsInBounds(north: number, south: number, east: number, west: number): Promise<Entity[]> {
    const params = new URLSearchParams({
      north: north.toString(),
      south: south.toString(),
      east: east.toString(),
      west: west.toString(),
    });
    return this.request<Entity[]>(`/positions/bounds?${params.toString()}`);
  }

  async getPositionStatistics(): Promise<PositionStatistics> {
    return this.request<PositionStatistics>('/positions/statistics');
  }

  async getPosition(entityId: string): Promise<Entity> {
    return this.request<Entity>(`/positions/${entityId}`);
  }

  async deletePosition(entityId: string): Promise<void> {
    return this.request<void>(`/positions/${entityId}`, { method: 'DELETE' });
  }

  async getPositionTrail(entityId: string, timeRange?: string, limit?: number): Promise<PositionPoint[]> {
    const params = new URLSearchParams();
    if (timeRange) params.append('timeRange', timeRange);
    if (limit) params.append('limit', limit.toString());
    
    const query = params.toString() ? `?${params.toString()}` : '';
    return this.request<PositionPoint[]>(`/positions/${entityId}/trail${query}`);
  }

  // ===== CHAT METHODS =====

  async createChatRoom(room: Partial<ChatRoom>): Promise<ChatRoom> {
    return this.request<ChatRoom>('/chat/rooms', {
      method: 'POST',
      body: JSON.stringify(room),
    });
  }

  async getChatRooms(): Promise<ChatRoom[]> {
    return this.request<ChatRoom[]>('/chat/rooms');
  }

  async getChatRoom(roomId: string): Promise<ChatRoom> {
    return this.request<ChatRoom>(`/chat/rooms/${roomId}`);
  }

  async sendMessage(roomId: string, message: Partial<ChatMessage>): Promise<ChatMessage> {
    return this.request<ChatMessage>(`/chat/rooms/${roomId}/messages`, {
      method: 'POST',
      body: JSON.stringify(message),
    });
  }

  async getMessages(roomId: string, limit?: number, before?: string): Promise<ChatMessage[]> {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    if (before) params.append('before', before);
    
    const query = params.toString() ? `?${params.toString()}` : '';
    return this.request<ChatMessage[]>(`/chat/rooms/${roomId}/messages${query}`);
  }

  async acknowledgeMessage(messageId: string): Promise<void> {
    return this.request<void>(`/chat/messages/${messageId}/acknowledge`, {
      method: 'POST',
    });
  }

  async addReaction(messageId: string, reaction: string): Promise<void> {
    return this.request<void>(`/chat/messages/${messageId}/reactions`, {
      method: 'POST',
      body: JSON.stringify({ reaction }),
    });
  }

  async getChatStatistics(): Promise<any> {
    return this.request<any>('/chat/statistics');
  }

  // ===== HEALTH CHECK =====

  async getHealth(): Promise<{ status: string; service: string; timestamp: string }> {
    const url = `${this.baseUrl}/health`;
    const response = await fetch(url);
    
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    
    return response.json();
  }
}

// Export singleton instance
export const apiClient = new GoTAKAPIClient();

// Export default instance for convenience
export default apiClient;
