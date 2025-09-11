import proj4 from 'proj4';
import * as mgrs from 'mgrs';

// Define common coordinate reference systems
proj4.defs([
  ['EPSG:4326', '+proj=longlat +datum=WGS84 +no_defs'], // WGS84 Geographic
  ['EPSG:3857', '+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs'], // Web Mercator
  ['EPSG:32633', '+proj=utm +zone=33 +datum=WGS84 +units=m +no_defs'], // UTM Zone 33N (Europe)
  ['EPSG:32635', '+proj=utm +zone=35 +datum=WGS84 +units=m +no_defs'], // UTM Zone 35N (Middle East)
  ['EPSG:32638', '+proj=utm +zone=38 +datum=WGS84 +units=m +no_defs'], // UTM Zone 38N (Asia)
]);

export interface LatLng {
  lat: number;
  lng: number;
}

export interface UTMCoordinate {
  easting: number;
  northing: number;
  zone: number;
  hemisphere: 'N' | 'S';
}

export interface MGRSCoordinate {
  mgrs: string;
}

export interface CartesianCoordinate {
  x: number;
  y: number;
}

/**
 * Convert decimal degrees to degrees, minutes, seconds
 */
export function ddToDms(decimal: number, isLongitude = false): string {
  const direction = isLongitude 
    ? (decimal >= 0 ? 'E' : 'W')
    : (decimal >= 0 ? 'N' : 'S');
  
  const absDecimal = Math.abs(decimal);
  const degrees = Math.floor(absDecimal);
  const minutesFloat = (absDecimal - degrees) * 60;
  const minutes = Math.floor(minutesFloat);
  const seconds = (minutesFloat - minutes) * 60;
  
  return `${degrees}° ${minutes}' ${seconds.toFixed(2)}" ${direction}`;
}

/**
 * Convert degrees, minutes, seconds to decimal degrees
 */
export function dmsToDd(degrees: number, minutes: number, seconds: number, direction: 'N' | 'S' | 'E' | 'W'): number {
  let decimal = Math.abs(degrees) + minutes/60 + seconds/3600;
  if (direction === 'S' || direction === 'W') {
    decimal = -decimal;
  }
  return decimal;
}

/**
 * Convert latitude/longitude to UTM coordinates
 */
export function latLngToUtm(lat: number, lng: number): UTMCoordinate {
  // Determine UTM zone
  const zone = Math.floor((lng + 180) / 6) + 1;
  const hemisphere = lat >= 0 ? 'N' : 'S';
  
  // Create UTM projection string
  const utmProj = `+proj=utm +zone=${zone} +datum=WGS84 +units=m +no_defs`;
  
  // Convert coordinates
  const [easting, northing] = proj4('EPSG:4326', utmProj, [lng, lat]);
  
  return {
    easting: Math.round(easting),
    northing: Math.round(northing),
    zone,
    hemisphere
  };
}

/**
 * Convert UTM coordinates to latitude/longitude
 */
export function utmToLatLng(utm: UTMCoordinate): LatLng {
  const utmProj = `+proj=utm +zone=${utm.zone} +${utm.hemisphere === 'S' ? '+south' : ''} +datum=WGS84 +units=m +no_defs`;
  const [lng, lat] = proj4(utmProj, 'EPSG:4326', [utm.easting, utm.northing]);
  
  return { lat, lng };
}

/**
 * Convert latitude/longitude to MGRS coordinates
 */
export function latLngToMgrs(lat: number, lng: number, precision = 5): string {
  try {
    return mgrs.forward([lng, lat], precision);
  } catch (error) {
    console.error('Error converting to MGRS:', error);
    return '';
  }
}

/**
 * Convert MGRS coordinates to latitude/longitude
 */
export function mgrsToLatLng(mgrsString: string): LatLng | null {
  try {
    const [lng, lat] = mgrs.inverse(mgrsString);
    return { lat, lng };
  } catch (error) {
    console.error('Error converting from MGRS:', error);
    return null;
  }
}

/**
 * Calculate distance between two points using Haversine formula
 * Returns distance in meters
 */
export function calculateDistance(point1: LatLng, point2: LatLng): number {
  const R = 6371000; // Earth's radius in meters
  const lat1Rad = (point1.lat * Math.PI) / 180;
  const lat2Rad = (point2.lat * Math.PI) / 180;
  const deltaLat = ((point2.lat - point1.lat) * Math.PI) / 180;
  const deltaLng = ((point2.lng - point1.lng) * Math.PI) / 180;

  const a = Math.sin(deltaLat / 2) * Math.sin(deltaLat / 2) +
    Math.cos(lat1Rad) * Math.cos(lat2Rad) *
    Math.sin(deltaLng / 2) * Math.sin(deltaLng / 2);
  
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

/**
 * Calculate bearing between two points
 * Returns bearing in degrees (0-360)
 */
export function calculateBearing(point1: LatLng, point2: LatLng): number {
  const lat1Rad = (point1.lat * Math.PI) / 180;
  const lat2Rad = (point2.lat * Math.PI) / 180;
  const deltaLng = ((point2.lng - point1.lng) * Math.PI) / 180;

  const x = Math.sin(deltaLng) * Math.cos(lat2Rad);
  const y = Math.cos(lat1Rad) * Math.sin(lat2Rad) - 
    Math.sin(lat1Rad) * Math.cos(lat2Rad) * Math.cos(deltaLng);

  const bearingRad = Math.atan2(x, y);
  const bearingDeg = (bearingRad * 180) / Math.PI;
  
  return (bearingDeg + 360) % 360;
}

/**
 * Format coordinates for display based on preference
 */
export interface CoordinateDisplayOptions {
  format: 'dd' | 'dms' | 'utm' | 'mgrs';
  precision?: number;
}

export function formatCoordinates(lat: number, lng: number, options: CoordinateDisplayOptions = { format: 'dd' }): string {
  switch (options.format) {
    case 'dms':
      return `${ddToDms(lat, false)} ${ddToDms(lng, true)}`;
    
    case 'utm': {
      const utm = latLngToUtm(lat, lng);
      return `${utm.zone}${utm.hemisphere} ${utm.easting}E ${utm.northing}N`;
    }
    
    case 'mgrs':
      return latLngToMgrs(lat, lng, options.precision || 5);
    
    case 'dd':
    default: {
      const precision = options.precision || 6;
      return `${lat.toFixed(precision)}, ${lng.toFixed(precision)}`;
    }
  }
}

/**
 * Parse coordinate string in various formats
 */
export function parseCoordinates(coordString: string): LatLng | null {
  const trimmed = coordString.trim();
  
  // Try MGRS format first
  if (trimmed.match(/^\d{1,2}[A-Z]{3}\d+$/)) {
    return mgrsToLatLng(trimmed);
  }
  
  // Try decimal degrees format: "lat, lng"
  const ddMatch = trimmed.match(/^(-?\d+\.?\d*)\s*,\s*(-?\d+\.?\d*)$/);
  if (ddMatch) {
    const lat = parseFloat(ddMatch[1]);
    const lng = parseFloat(ddMatch[2]);
    if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
      return { lat, lng };
    }
  }
  
  // Try DMS format
  const dmsMatch = trimmed.match(/^(\d+)°\s*(\d+)'\s*(\d+\.?\d*)"?\s*([NS])\s+(\d+)°\s*(\d+)'\s*(\d+\.?\d*)"?\s*([EW])$/);
  if (dmsMatch) {
    const latDeg = parseInt(dmsMatch[1]);
    const latMin = parseInt(dmsMatch[2]);
    const latSec = parseFloat(dmsMatch[3]);
    const latDir = dmsMatch[4] as 'N' | 'S';
    const lngDeg = parseInt(dmsMatch[5]);
    const lngMin = parseInt(dmsMatch[6]);
    const lngSec = parseFloat(dmsMatch[7]);
    const lngDir = dmsMatch[8] as 'E' | 'W';
    
    return {
      lat: dmsToDd(latDeg, latMin, latSec, latDir),
      lng: dmsToDd(lngDeg, lngMin, lngSec, lngDir)
    };
  }
  
  return null;
}

/**
 * Check if coordinates are within valid ranges
 */
export function isValidLatLng(lat: number, lng: number): boolean {
  return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
}

/**
 * Normalize longitude to -180 to 180 range
 */
export function normalizeLongitude(lng: number): number {
  while (lng > 180) lng -= 360;
  while (lng < -180) lng += 360;
  return lng;
}

/**
 * Create a bounding box around a point with given radius (in meters)
 */
export function createBoundingBox(center: LatLng, radiusMeters: number): { 
  north: number; 
  south: number; 
  east: number; 
  west: number 
} {
  const lat = center.lat;
  const lng = center.lng;
  
  // Earth's radius in meters
  const R = 6371000;
  
  // Angular distance in radians
  const angular = radiusMeters / R;
  
  const north = lat + (angular * 180 / Math.PI);
  const south = lat - (angular * 180 / Math.PI);
  
  // Account for latitude when calculating longitude offset
  const latRad = lat * Math.PI / 180;
  const lngOffset = (angular * 180 / Math.PI) / Math.cos(latRad);
  
  const east = normalizeLongitude(lng + lngOffset);
  const west = normalizeLongitude(lng - lngOffset);
  
  return { north, south, east, west };
}
