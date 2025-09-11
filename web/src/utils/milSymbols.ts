import type { EntityType, Affiliation } from '../types/entity';
import type { MilSymbol as TacticalMilSymbol, MilAffiliation, MilSymbolCategory } from '../types/tactical';

export interface MilSymbol {
  html: string;
  size: [number, number];
  anchor: [number, number];
}

// Color schemes for different affiliations following NATO standards
export const AFFILIATION_COLORS = {
  friendly: '#00FF00',    // Green
  hostile: '#FF0000',     // Red
  neutral: '#FFFF00',     // Yellow
  unknown: '#FF00FF'      // Magenta
};

// Simple tactical symbols using Unicode and CSS
export const getMilSymbol = (type: EntityType, affiliation: Affiliation): MilSymbol => {
  const color = AFFILIATION_COLORS[affiliation] || AFFILIATION_COLORS.unknown;
  const baseStyle = `
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: ${color};
    border: 2px solid #000;
    border-radius: 3px;
    color: #000;
    font-weight: bold;
    font-size: 12px;
    font-family: 'Courier New', monospace;
  `;
  
  // Determine symbol based on entity type
  let symbol: string;
  let shape = 'square'; // Default shape
  
  switch (type) {
    case 'ground-infantry':
      symbol = '👤';
      break;
    case 'ground-vehicle-wheeled':
    case 'ground-vehicle-tracked':
      symbol = '🚗';
      break;
    case 'air-rotorcraft':
    case 'air-fixed-wing':
      symbol = '✈️';
      shape = 'diamond';
      break;
    case 'naval-surface':
      symbol = '🚢';
      break;
    case 'unknown':
    default:
      symbol = '?';
      break;
  }
  
  // Apply shape styling
  const shapeStyle = shape === 'diamond' 
    ? 'transform: rotate(45deg); border-radius: 0;'
    : '';
  
  const html = `
    <div style="${baseStyle} ${shapeStyle}">
      <span style="${shape === 'diamond' ? 'transform: rotate(-45deg);' : ''}">${symbol}</span>
    </div>
  `;
  
  return {
    html,
    size: [32, 32],
    anchor: [16, 16]
  };
};

// Create enhanced symbols with callsign labels
export const getMilSymbolWithLabel = (
  type: EntityType, 
  affiliation: Affiliation, 
  callsign: string
): MilSymbol => {
  // Color is determined by affiliation
  const baseSymbol = getMilSymbol(type, affiliation);
  
  const html = `
    <div style="display: flex; flex-direction: column; align-items: center;">
      ${baseSymbol.html}
      <div style="
        background-color: rgba(255, 255, 255, 0.9);
        border: 1px solid #000;
        border-radius: 2px;
        padding: 1px 3px;
        font-size: 10px;
        font-weight: bold;
        margin-top: 2px;
        white-space: nowrap;
      ">
        ${callsign}
      </div>
    </div>
  `;
  
  return {
    html,
    size: [50, 50],
    anchor: [25, 32]
  };
};

// Speed and course indicator
export const getSpeedCourseIndicator = (speed: number, course: number): string => {
  const speedKnots = Math.round(speed * 1.944); // m/s to knots
  const courseRounded = Math.round(course);
  
  return `
    <div style="
      position: absolute;
      top: -20px;
      left: 50%;
      transform: translateX(-50%);
      background-color: rgba(0, 0, 0, 0.7);
      color: white;
      padding: 2px 4px;
      border-radius: 3px;
      font-size: 9px;
      white-space: nowrap;
    ">
      ${speedKnots}kts ${courseRounded}°
    </div>
  `;
};

// Trail line for movement history
export const createTrailStyle = (age: number, maxAge: number) => {
  const opacity = Math.max(0.1, 1 - (age / maxAge));
  const width = Math.max(1, 3 * opacity);
  
  return {
    color: '#4A90E2',
    weight: width,
    opacity: opacity,
    dashArray: '2, 4'
  };
};

// Classification colors for security levels
export const CLASSIFICATION_COLORS = {
  UNCLASSIFIED: '#00FF00',
  RESTRICTED: '#FFFF00', 
  CONFIDENTIAL: '#0000FF',
  SECRET: '#FF8C00',
  TOP_SECRET: '#FF0000'
};

// Get border color based on classification
export const getClassificationBorder = (classification: string): string => {
  const classLevel = (classification || 'unclassified').toUpperCase().replace(/\s/g, '_');
  return CLASSIFICATION_COLORS[classLevel as keyof typeof CLASSIFICATION_COLORS] || '#808080';
};

// Tactical overlay symbols
export const TACTICAL_SYMBOLS = {
  checkpoint: '⛳',
  objective: '🎯', 
  waypoint: '📍',
  boundary: '━━━',
  route: '➡️',
  danger: '⚠️',
  safe: '✅',
  unknown: '❓'
};

export const getTacticalOverlayIcon = (overlayType: string): string => {
  return TACTICAL_SYMBOLS[overlayType as keyof typeof TACTICAL_SYMBOLS] || TACTICAL_SYMBOLS.unknown;
};

// Enhanced MIL-STD-2525 Symbol Library for Tactical Overlays
export const ENHANCED_MIL_SYMBOLS: Record<string, TacticalMilSymbol> = {
  // Land Units - Infantry
  'FRIENDLY_INFANTRY': {
    id: 'friendly_infantry',
    sidc: '10031000000000000000',
    name: 'Infantry',
    description: 'Friendly Infantry Unit',
    category: 'LAND_UNIT',
    affiliation: 'FRIENDLY',
    symbolSet: 'LAND_UNIT',
    size: [32, 32],
  },
  'HOSTILE_INFANTRY': {
    id: 'hostile_infantry',
    sidc: '10031000000000000000',
    name: 'Infantry',
    description: 'Hostile Infantry Unit',
    category: 'LAND_UNIT',
    affiliation: 'HOSTILE',
    symbolSet: 'LAND_UNIT',
    size: [32, 32],
  },
  
  // Land Units - Armor
  'FRIENDLY_ARMOR': {
    id: 'friendly_armor',
    sidc: '10031200000000000000',
    name: 'Armor',
    description: 'Friendly Armored Unit',
    category: 'LAND_UNIT',
    affiliation: 'FRIENDLY',
    symbolSet: 'LAND_UNIT',
    size: [32, 32],
  },
  'HOSTILE_ARMOR': {
    id: 'hostile_armor',
    sidc: '10031200000000000000',
    name: 'Armor',
    description: 'Hostile Armored Unit',
    category: 'LAND_UNIT',
    affiliation: 'HOSTILE',
    symbolSet: 'LAND_UNIT',
    size: [32, 32],
  },
  
  // Control Measures
  'CHECKPOINT': {
    id: 'checkpoint',
    sidc: '25010000000000000000',
    name: 'Checkpoint',
    description: 'Checkpoint',
    category: 'CONTROL_MEASURE',
    affiliation: 'FRIENDLY',
    symbolSet: 'CONTROL_MEASURES',
    size: [24, 24],
  },
  'OBJECTIVE': {
    id: 'objective',
    sidc: '25020000000000000000',
    name: 'Objective',
    description: 'Objective Area',
    category: 'CONTROL_MEASURE',
    affiliation: 'FRIENDLY',
    symbolSet: 'CONTROL_MEASURES',
    size: [32, 32],
  },
  'LANDING_ZONE': {
    id: 'landing_zone',
    sidc: '25030000000000000000',
    name: 'Landing Zone',
    description: 'Landing Zone',
    category: 'CONTROL_MEASURE',
    affiliation: 'FRIENDLY',
    symbolSet: 'CONTROL_MEASURES',
    size: [32, 32],
  },
};

// Create SVG symbol for tactical overlays
export function createTacticalSymbolSVG(symbol: TacticalMilSymbol, size: number = 32): string {
  const { affiliation, category } = symbol;
  
  // Base shape depends on affiliation
  const getBaseShape = () => {
    switch (affiliation) {
      case 'FRIENDLY':
        return `<rect x="2" y="8" width="${size-4}" height="${size-16}" fill="none" stroke="#22c55e" stroke-width="2" rx="2"/>`;
      case 'HOSTILE':
        return `<polygon points="2,8 ${size-2},8 ${size/2},${size-8}" fill="none" stroke="#ef4444" stroke-width="2"/>`;
      case 'NEUTRAL':
        return `<rect x="2" y="8" width="${size-4}" height="${size-16}" fill="none" stroke="#f59e0b" stroke-width="2"/>`;
      case 'UNKNOWN':
        return `<polygon points="2,${size/2} ${size/2},8 ${size-2},${size/2} ${size/2},${size-8}" fill="none" stroke="#6b7280" stroke-width="2"/>`;
      default:
        return `<circle cx="${size/2}" cy="${size/2}" r="${size/2-4}" fill="none" stroke="#6b7280" stroke-width="2"/>`;
    }
  };
  
  // Icon inside shape depends on category
  const getIcon = () => {
    const color = affiliation === 'FRIENDLY' ? '#22c55e' : 
                  affiliation === 'HOSTILE' ? '#ef4444' :
                  affiliation === 'NEUTRAL' ? '#f59e0b' : '#6b7280';
    
    switch (symbol.id) {
      case 'friendly_infantry':
      case 'hostile_infantry':
        return `<text x="${size/2}" y="${size/2+4}" text-anchor="middle" font-family="Arial" font-size="12" font-weight="bold" fill="${color}">I</text>`;
      case 'friendly_armor':
      case 'hostile_armor':
        return `<circle cx="${size/2}" cy="${size/2}" r="6" fill="${color}"/>`;
      case 'checkpoint':
        return `<text x="${size/2}" y="${size/2+3}" text-anchor="middle" font-family="Arial" font-size="8" font-weight="bold" fill="${color}">CP</text>`;
      case 'objective':
        return `<text x="${size/2}" y="${size/2+4}" text-anchor="middle" font-family="Arial" font-size="10" font-weight="bold" fill="${color}">OBJ</text>`;
      case 'landing_zone':
        return `<text x="${size/2}" y="${size/2+3}" text-anchor="middle" font-family="Arial" font-size="9" font-weight="bold" fill="${color}">LZ</text>`;
      default:
        return `<circle cx="${size/2}" cy="${size/2}" r="3" fill="${color}"/>`;
    }
  };
  
  return `
    <svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">
      ${getBaseShape()}
      ${getIcon()}
    </svg>
  `;
}

// Convert SVG to data URL for use as marker icon
export function tacticalSymbolToDataURL(symbol: TacticalMilSymbol, size: number = 32): string {
  const svg = createTacticalSymbolSVG(symbol, size);
  return 'data:image/svg+xml;base64,' + btoa(svg);
}

// Get tactical symbol by ID
export function getTacticalSymbol(id: string): TacticalMilSymbol | undefined {
  return ENHANCED_MIL_SYMBOLS[id];
}

// Get tactical symbols by category
export function getTacticalSymbolsByCategory(category: MilSymbolCategory): TacticalMilSymbol[] {
  return Object.values(ENHANCED_MIL_SYMBOLS).filter(symbol => symbol.category === category);
}

// Get tactical symbols by affiliation
export function getTacticalSymbolsByAffiliation(affiliation: MilAffiliation): TacticalMilSymbol[] {
  return Object.values(ENHANCED_MIL_SYMBOLS).filter(symbol => symbol.affiliation === affiliation);
}

// Search tactical symbols
export function searchTacticalSymbols(query: string): TacticalMilSymbol[] {
  const searchTerm = query.toLowerCase();
  return Object.values(ENHANCED_MIL_SYMBOLS).filter(symbol => 
    symbol.name.toLowerCase().includes(searchTerm) ||
    symbol.description?.toLowerCase().includes(searchTerm) ||
    symbol.id.toLowerCase().includes(searchTerm)
  );
}
