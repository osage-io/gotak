import { LatLng } from 'leaflet';

export type CoordinateFormat = 'decimal' | 'dms' | 'mgrs';

/**
 * Calculate distance between two coordinates using Haversine formula
 * @param coord1 First coordinate point
 * @param coord2 Second coordinate point
 * @returns Distance in meters
 */
export const calculateDistance = (coord1: LatLng, coord2: LatLng): number => {
  const R = 6371000; // Earth's radius in meters
  const φ1 = coord1.lat * Math.PI / 180;
  const φ2 = coord2.lat * Math.PI / 180;
  const Δφ = (coord2.lat - coord1.lat) * Math.PI / 180;
  const Δλ = (coord2.lng - coord1.lng) * Math.PI / 180;

  const a = Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
            Math.cos(φ1) * Math.cos(φ2) *
            Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

  return R * c;
};

/**
 * Calculate bearing between two coordinates
 * @param coord1 Starting coordinate point
 * @param coord2 Ending coordinate point
 * @returns Bearing in degrees (0-360)
 */
export const calculateBearing = (coord1: LatLng, coord2: LatLng): number => {
  const φ1 = coord1.lat * Math.PI / 180;
  const φ2 = coord2.lat * Math.PI / 180;
  const Δλ = (coord2.lng - coord1.lng) * Math.PI / 180;

  const y = Math.sin(Δλ) * Math.cos(φ2);
  const x = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(Δλ);

  const θ = Math.atan2(y, x);
  return ((θ * 180 / Math.PI) + 360) % 360;
};

/**
 * Format distance for display
 * @param meters Distance in meters
 * @returns Formatted distance string
 */
export const formatDistance = (meters: number): string => {
  if (meters < 1000) {
    return `${Math.round(meters)} m`;
  } else if (meters < 10000) {
    return `${(meters / 1000).toFixed(2)} km`;
  } else {
    return `${(meters / 1000).toFixed(1)} km`;
  }
};

/**
 * Format area for display
 * @param squareMeters Area in square meters
 * @returns Formatted area string
 */
export const formatArea = (squareMeters: number): string => {
  if (squareMeters < 10000) {
    return `${Math.round(squareMeters)} m²`;
  } else if (squareMeters < 1000000) {
    return `${(squareMeters / 10000).toFixed(2)} ha`;
  } else {
    return `${(squareMeters / 1000000).toFixed(2)} km²`;
  }
};

/**
 * Format bearing for display
 * @param degrees Bearing in degrees
 * @returns Formatted bearing string
 */
export const formatBearing = (degrees: number): string => {
  const cardinal = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
  const index = Math.round(degrees / 22.5) % 16;
  return `${Math.round(degrees)}° ${cardinal[index]}`;
};

/**
 * Convert decimal degrees to DMS format
 * @param decimal Decimal degree value
 * @param type Coordinate type ('lat' or 'lng')
 * @returns DMS formatted string
 */
const decimalToDMS = (decimal: number, type: 'lat' | 'lng'): string => {
  const abs = Math.abs(decimal);
  const degrees = Math.floor(abs);
  const minutes = Math.floor((abs - degrees) * 60);
  const seconds = ((abs - degrees) * 60 - minutes) * 60;
  
  const direction = type === 'lat' 
    ? (decimal >= 0 ? 'N' : 'S')
    : (decimal >= 0 ? 'E' : 'W');
  
  return `${degrees}°${minutes}'${seconds.toFixed(2)}"${direction}`;
};

/**
 * Format coordinates for display
 * @param lat Latitude
 * @param lng Longitude
 * @param options Formatting options
 * @returns Formatted coordinate string
 */
export const formatCoordinates = (
  lat: number,
  lng: number,
  options: {
    format?: CoordinateFormat;
    precision?: number;
  } = {}
): string => {
  const { format = 'decimal', precision = 6 } = options;
  
  switch (format) {
    case 'decimal':
      return `${lat.toFixed(precision)}, ${lng.toFixed(precision)}`;
    
    case 'dms':
      const latDMS = decimalToDMS(lat, 'lat');
      const lngDMS = decimalToDMS(lng, 'lng');
      return `${latDMS}, ${lngDMS}`;
    
    case 'mgrs':
      // MGRS conversion would require additional library
      // For now, fallback to decimal
      return `${lat.toFixed(precision)}, ${lng.toFixed(precision)}`;
    
    default:
      return `${lat.toFixed(precision)}, ${lng.toFixed(precision)}`;
  }
};

/**
 * Calculate the area of a polygon using the shoelace formula
 * @param points Array of coordinate points
 * @returns Area in square meters
 */
export const calculatePolygonArea = (points: LatLng[]): number => {
  if (points.length < 3) return 0;
  
  let area = 0;
  const n = points.length;
  
  for (let i = 0; i < n; i++) {
    const j = (i + 1) % n;
    const xi = points[i].lng * Math.PI / 180;
    const yi = points[i].lat * Math.PI / 180;
    const xj = points[j].lng * Math.PI / 180;
    const yj = points[j].lat * Math.PI / 180;
    
    area += xi * yj - xj * yi;
  }
  
  area = Math.abs(area) / 2;
  
  // Convert from square radians to square meters
  const R = 6378137; // Earth's radius in meters
  return area * R * R;
};

/**
 * Calculate the area of a rectangle defined by two opposite corners
 * @param corner1 First corner coordinate
 * @param corner2 Opposite corner coordinate
 * @returns Area in square meters
 */
export const calculateRectangleArea = (corner1: LatLng, corner2: LatLng): number => {
  const width = calculateDistance(corner1, { lat: corner1.lat, lng: corner2.lng });
  const height = calculateDistance(corner1, { lat: corner2.lat, lng: corner1.lng });
  return width * height;
};

/**
 * Check if a point is within a circular geofence
 * @param point Point to check
 * @param center Center of the circle
 * @param radius Radius in meters
 * @returns True if point is within the geofence
 */
export const isPointInCircle = (point: LatLng, center: LatLng, radius: number): boolean => {
  const distance = calculateDistance(point, center);
  return distance <= radius;
};

/**
 * Check if a point is within a polygon geofence using ray casting
 * @param point Point to check
 * @param polygon Array of polygon vertices
 * @returns True if point is within the polygon
 */
export const isPointInPolygon = (point: LatLng, polygon: LatLng[]): boolean => {
  const x = point.lng;
  const y = point.lat;
  let inside = false;
  
  for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
    const xi = polygon[i].lng;
    const yi = polygon[i].lat;
    const xj = polygon[j].lng;
    const yj = polygon[j].lat;
    
    if (((yi > y) !== (yj > y)) && (x < (xj - xi) * (y - yi) / (yj - yi) + xi)) {
      inside = !inside;
    }
  }
  
  return inside;
};

/**
 * Check if a point is within a rectangular geofence
 * @param point Point to check
 * @param corner1 First corner of rectangle
 * @param corner2 Opposite corner of rectangle
 * @returns True if point is within the rectangle
 */
export const isPointInRectangle = (point: LatLng, corner1: LatLng, corner2: LatLng): boolean => {
  const minLat = Math.min(corner1.lat, corner2.lat);
  const maxLat = Math.max(corner1.lat, corner2.lat);
  const minLng = Math.min(corner1.lng, corner2.lng);
  const maxLng = Math.max(corner1.lng, corner2.lng);
  
  return point.lat >= minLat && point.lat <= maxLat &&
         point.lng >= minLng && point.lng <= maxLng;
};
