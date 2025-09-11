export interface Entity {
  id: string;
  callsign: string;
  type: EntityType;
  affiliation: Affiliation;
  lat: number;
  lng: number;
  altitude?: number;
  speed?: number;
  course?: number;
  lastUpdate?: Date;
  classification: Classification;
  status: EntityStatus;
}

export type EntityType =
  | 'ground-infantry'
  | 'ground-vehicle-wheeled'
  | 'ground-vehicle-tracked'
  | 'air-rotorcraft'
  | 'air-fixed-wing'
  | 'naval-surface'
  | 'unknown';

export type Affiliation = 'friendly' | 'hostile' | 'neutral' | 'unknown';

export type Classification = 'UNCLASSIFIED' | 'RESTRICTED' | 'CONFIDENTIAL' | 'SECRET' | 'TOP_SECRET' | 'unclassified';

export type EntityStatus = 'active' | 'inactive' | 'destroyed' | 'missing';

export interface EntityHistory {
  uid: string;
  positions: PositionPoint[];
  totalCount: number;
}

export interface PositionPoint {
  lat: number;
  lon: number;
  altitude?: number;
  timestamp: string;
  speed?: number;
  course?: number;
}

export interface TacticalOverlay {
  id: string;
  name: string;
  type: OverlayType;
  geometry: GeoJSON.Geometry;
  properties: OverlayProperties;
  groupId: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export type OverlayType = 'point' | 'line' | 'polygon' | 'text' | 'symbol';

export interface OverlayProperties {
  icon?: string;
  color?: string;
  description?: string;
  strokeWidth?: number;
  fillColor?: string;
  fillOpacity?: number;
  text?: string;
  fontSize?: number;
}

export interface WebSocketMessage {
  type: MessageType;
  data: any;
  timestamp: string;
}

export type MessageType = 
  | 'position_update'
  | 'overlay_update'
  | 'entity_removed'
  | 'heartbeat'
  | 'error';

export interface PositionUpdateMessage {
  type: 'position_update';
  data: Entity;
  timestamp: string;
}

export interface OverlayUpdateMessage {
  type: 'overlay_update';
  data: {
    id: string;
    action: 'create' | 'update' | 'delete';
    overlay?: TacticalOverlay;
  };
  timestamp: string;
}
