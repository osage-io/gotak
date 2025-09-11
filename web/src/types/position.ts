export interface EntityPosition {
  entityId: string;
  uid: string;
  type: string;
  callsign: string;
  group: string;
  
  // Position data
  lat: number;
  lng: number;
  altitude?: number;
  speed?: number;
  course?: number;
  
  // Metadata
  lastUpdate: string; // ISO timestamp
  staleTime: string; // ISO timestamp
  isStale: boolean;
  
  // Tactical information
  isFriendly: boolean;
  isHostile: boolean;
  
  // Position accuracy
  circularError?: number;
  linearError?: number;
}

export interface PositionHistory {
  lat: number;
  lng: number;
  altitude?: number;
  timestamp: string; // ISO timestamp
}

export interface PositionUpdate {
  entityId: string;
  position: {
    lat: number;
    lng: number;
    altitude?: number;
    speed?: number;
    course?: number;
    timestamp: string;
  };
}

export interface PositionStatistics {
  total_entities: number;
  active_entities: number;
  stale_entities: number;
  friendly_entities: number;
  hostile_entities: number;
  trails_tracked: number;
}

export interface TacticalWSMessage {
  type: string;
  payload: any;
  timestamp: string;
  roomId?: string;
}

// WebSocket message types for position updates
export const WS_MESSAGE_TYPES = {
  POSITION_UPDATE: 'position_update',
  ENTITY_REMOVED: 'entity_removed',
  HEARTBEAT: 'heartbeat',
  ERROR: 'error',
} as const;

export type WSMessageType = typeof WS_MESSAGE_TYPES[keyof typeof WS_MESSAGE_TYPES];

// Entity marker configuration
export interface EntityMarkerConfig {
  color: string;
  iconUrl?: string;
  size: [number, number];
  anchor: [number, number];
  popupAnchor: [number, number];
}

export const ENTITY_MARKER_CONFIGS: Record<string, EntityMarkerConfig> = {
  friendly: {
    color: '#00ff00',
    size: [32, 32],
    anchor: [16, 32],
    popupAnchor: [0, -32],
  },
  hostile: {
    color: '#ff0000', 
    size: [32, 32],
    anchor: [16, 32],
    popupAnchor: [0, -32],
  },
  neutral: {
    color: '#ffff00',
    size: [32, 32], 
    anchor: [16, 32],
    popupAnchor: [0, -32],
  },
  unknown: {
    color: '#888888',
    size: [32, 32],
    anchor: [16, 32],
    popupAnchor: [0, -32],
  },
};

// API endpoints
export const POSITION_API_ENDPOINTS = {
  ALL_POSITIONS: '/api/v1/positions',
  ACTIVE_POSITIONS: '/api/v1/positions/active',
  FRIENDLY_POSITIONS: '/api/v1/positions/friendly',
  HOSTILE_POSITIONS: '/api/v1/positions/hostile',
  POSITIONS_IN_BOUNDS: '/api/v1/positions/bounds',
  POSITION_BY_ID: '/api/v1/positions/:entityId',
  POSITION_TRAIL: '/api/v1/positions/:entityId/trail',
  POSITION_STATISTICS: '/api/v1/positions/statistics',
} as const;
