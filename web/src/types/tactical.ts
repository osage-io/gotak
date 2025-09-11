export interface TacticalOverlay {
  id: string;
  type: TacticalOverlayType;
  name: string;
  description?: string;
  geometry: TacticalGeometry;
  style: TacticalStyle;
  metadata: TacticalMetadata;
  visible: boolean;
  editable: boolean;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

export type TacticalOverlayType = 
  | 'symbol'
  | 'route'
  | 'area'
  | 'boundary'
  | 'threat_circle'
  | 'range_ring'
  | 'line'
  | 'polygon'
  | 'circle'
  | 'marker';

export interface TacticalGeometry {
  type: 'Point' | 'LineString' | 'Polygon' | 'Circle';
  coordinates: number[] | number[][] | number[][][]; // GeoJSON-like format
  radius?: number; // For circles
  properties?: Record<string, any>;
}

export interface TacticalStyle {
  color: string;
  fillColor?: string;
  weight: number;
  opacity: number;
  fillOpacity?: number;
  dashArray?: string;
  lineCap?: 'butt' | 'round' | 'square';
  lineJoin?: 'miter' | 'round' | 'bevel';
  className?: string;
}

export interface TacticalMetadata {
  classification?: SecurityClassification;
  priority: TacticalPriority;
  unit?: string;
  command?: string;
  effectiveTime?: string;
  expirationTime?: string;
  source?: string;
  tags?: string[];
  attributes?: Record<string, any>;
}

export type SecurityClassification = 
  | 'UNCLASSIFIED'
  | 'CONFIDENTIAL'
  | 'SECRET'
  | 'TOP_SECRET';

export type TacticalPriority = 
  | 'LOW'
  | 'MEDIUM' 
  | 'HIGH'
  | 'CRITICAL';

// MIL-STD-2525 Symbol Types (basic implementation)
export interface MilSymbol {
  id: string;
  sidc: string; // Symbol Identification Code
  name: string;
  description?: string;
  category: MilSymbolCategory;
  affiliation: MilAffiliation;
  symbolSet: MilSymbolSet;
  iconUrl?: string;
  size: [number, number];
}

export type MilSymbolCategory = 
  | 'LAND_UNIT'
  | 'AIR_UNIT'
  | 'SEA_SURFACE_UNIT'
  | 'LAND_INSTALLATION'
  | 'CONTROL_MEASURE'
  | 'FIRE_SUPPORT'
  | 'SPECIAL_OPERATION_FORCES';

export type MilAffiliation = 
  | 'FRIENDLY'
  | 'HOSTILE'
  | 'NEUTRAL' 
  | 'UNKNOWN';

export type MilSymbolSet = 
  | 'LAND_UNIT'
  | 'LAND_CIVILIAN_UNIT'
  | 'LAND_EQUIPMENT'
  | 'LAND_INSTALLATION'
  | 'AIR'
  | 'SEA_SURFACE'
  | 'SEA_SUBSURFACE'
  | 'SPACE'
  | 'CYBERSPACE'
  | 'CONTROL_MEASURES';

// Drawing Tool Types
export interface DrawingTool {
  id: string;
  name: string;
  icon: string;
  type: DrawingToolType;
  enabled: boolean;
  options?: DrawingToolOptions;
}

export type DrawingToolType = 
  | 'marker'
  | 'line'
  | 'polygon' 
  | 'rectangle'
  | 'circle'
  | 'route'
  | 'boundary'
  | 'threat_circle'
  | 'range_ring'
  | 'symbol';

export interface DrawingToolOptions {
  style?: Partial<TacticalStyle>;
  snapToGrid?: boolean;
  allowIntersection?: boolean;
  showArea?: boolean;
  showLength?: boolean;
  metric?: boolean;
  feet?: boolean;
  precision?: number;
}

// Threat Circle Types
export interface ThreatCircle {
  id: string;
  center: [number, number]; // [lat, lng]
  radius: number; // in meters
  threatType: ThreatType;
  threatLevel: ThreatLevel;
  name: string;
  description?: string;
  style: TacticalStyle;
  visible: boolean;
}

export type ThreatType = 
  | 'AIR_DEFENSE'
  | 'ARTILLERY'
  | 'MISSILE'
  | 'SMALL_ARMS'
  | 'EXPLOSIVE'
  | 'CHEMICAL'
  | 'BIOLOGICAL'
  | 'NUCLEAR'
  | 'CYBER'
  | 'ELECTRONIC_WARFARE';

export type ThreatLevel = 
  | 'LOW'
  | 'MODERATE'
  | 'HIGH'
  | 'EXTREME';

// Range Ring Types
export interface RangeRing {
  id: string;
  center: [number, number]; // [lat, lng]
  rings: RangeRingData[];
  name: string;
  description?: string;
  visible: boolean;
  style: TacticalStyle;
}

export interface RangeRingData {
  radius: number; // in meters
  label?: string;
  style?: Partial<TacticalStyle>;
}

// Route Types
export interface TacticalRoute {
  id: string;
  name: string;
  description?: string;
  waypoints: Waypoint[];
  style: TacticalStyle;
  routeType: RouteType;
  classification?: SecurityClassification;
  priority: TacticalPriority;
  visible: boolean;
  editable: boolean;
}

export interface Waypoint {
  id: string;
  position: [number, number]; // [lat, lng]
  name?: string;
  description?: string;
  elevation?: number;
  speed?: number;
  time?: string;
  waypointType: WaypointType;
}

export type WaypointType = 
  | 'START'
  | 'END'
  | 'CHECKPOINT' 
  | 'OBJECTIVE'
  | 'RALLY_POINT'
  | 'SUPPLY_POINT'
  | 'CASUALTY_COLLECTION_POINT'
  | 'LANDING_ZONE'
  | 'PICKUP_ZONE';

export type RouteType = 
  | 'PRIMARY'
  | 'ALTERNATE'
  | 'CONTINGENCY'
  | 'SUPPLY'
  | 'WITHDRAWAL'
  | 'INFILTRATION'
  | 'EXFILTRATION';

// Area/Boundary Types
export interface TacticalArea {
  id: string;
  name: string;
  description?: string;
  boundary: [number, number][]; // Array of [lat, lng] coordinates
  areaType: AreaType;
  style: TacticalStyle;
  classification?: SecurityClassification;
  priority: TacticalPriority;
  visible: boolean;
  editable: boolean;
}

export type AreaType = 
  | 'AREA_OF_OPERATIONS'
  | 'AREA_OF_INTEREST'
  | 'OBJECTIVE_AREA'
  | 'ASSEMBLY_AREA'
  | 'BATTLE_POSITION'
  | 'ENGAGEMENT_AREA'
  | 'KILL_ZONE'
  | 'NO_FLY_ZONE'
  | 'RESTRICTED_AREA'
  | 'DANGER_AREA'
  | 'CONTAMINATED_AREA';

// Overlay Management
export interface OverlayLayer {
  id: string;
  name: string;
  description?: string;
  overlays: TacticalOverlay[];
  visible: boolean;
  locked: boolean;
  opacity: number;
  zIndex: number;
}

export interface OverlayManager {
  layers: OverlayLayer[];
  activeLayerId?: string;
  selectedOverlayId?: string;
  drawingMode: boolean;
  activeDrawingTool?: DrawingToolType;
}

// Default styles for different overlay types
export const DEFAULT_STYLES: Record<TacticalOverlayType, TacticalStyle> = {
  symbol: {
    color: '#ffffff',
    weight: 2,
    opacity: 1,
    className: 'tactical-symbol',
  },
  route: {
    color: '#3b82f6',
    weight: 4,
    opacity: 0.8,
    lineCap: 'round',
    lineJoin: 'round',
  },
  area: {
    color: '#22c55e',
    fillColor: '#22c55e',
    weight: 2,
    opacity: 0.8,
    fillOpacity: 0.2,
    dashArray: '10,5',
  },
  boundary: {
    color: '#f59e0b',
    weight: 3,
    opacity: 0.9,
    dashArray: '15,5',
    lineCap: 'butt',
  },
  threat_circle: {
    color: '#ef4444',
    fillColor: '#ef4444',
    weight: 2,
    opacity: 0.8,
    fillOpacity: 0.1,
    dashArray: '5,5',
  },
  range_ring: {
    color: '#6366f1',
    weight: 2,
    opacity: 0.7,
    dashArray: '10,10',
    fillOpacity: 0,
  },
  line: {
    color: '#8b5cf6',
    weight: 3,
    opacity: 0.8,
    lineCap: 'round',
  },
  polygon: {
    color: '#ec4899',
    fillColor: '#ec4899',
    weight: 2,
    opacity: 0.8,
    fillOpacity: 0.15,
  },
  circle: {
    color: '#06b6d4',
    fillColor: '#06b6d4',
    weight: 2,
    opacity: 0.8,
    fillOpacity: 0.1,
  },
  marker: {
    color: '#f97316',
    weight: 2,
    opacity: 1,
  },
};

// Color schemes for different classifications and priorities
export const CLASSIFICATION_COLORS = {
  UNCLASSIFIED: '#22c55e',
  CONFIDENTIAL: '#3b82f6',
  SECRET: '#f59e0b',
  TOP_SECRET: '#ef4444',
} as const;

export const PRIORITY_COLORS = {
  LOW: '#6b7280',
  MEDIUM: '#f59e0b',
  HIGH: '#f97316',
  CRITICAL: '#ef4444',
} as const;

export const AFFILIATION_COLORS = {
  FRIENDLY: '#22c55e',
  HOSTILE: '#ef4444',
  NEUTRAL: '#f59e0b',
  UNKNOWN: '#6b7280',
} as const;
