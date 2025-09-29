import L from 'leaflet';

export interface Point {
  lat: number;
  lng: number;
}

// Earth's radius in meters
const EARTH_RADIUS = 6371000;

/**
 * Calculate distance between two points using Haversine formula
 */
export function calculateDistance(p1: Point, p2: Point): number {
  const lat1Rad = (p1.lat * Math.PI) / 180;
  const lat2Rad = (p2.lat * Math.PI) / 180;
  const deltaLatRad = ((p2.lat - p1.lat) * Math.PI) / 180;
  const deltaLngRad = ((p2.lng - p1.lng) * Math.PI) / 180;

  const a = 
    Math.sin(deltaLatRad / 2) * Math.sin(deltaLatRad / 2) +
    Math.cos(lat1Rad) * Math.cos(lat2Rad) *
    Math.sin(deltaLngRad / 2) * Math.sin(deltaLngRad / 2);
  
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  
  return EARTH_RADIUS * c;
}

/**
 * Calculate bearing between two points
 */
export function calculateBearing(p1: Point, p2: Point): number {
  const lat1Rad = (p1.lat * Math.PI) / 180;
  const lat2Rad = (p2.lat * Math.PI) / 180;
  const deltaLngRad = ((p2.lng - p1.lng) * Math.PI) / 180;

  const y = Math.sin(deltaLngRad) * Math.cos(lat2Rad);
  const x = Math.cos(lat1Rad) * Math.sin(lat2Rad) - 
           Math.sin(lat1Rad) * Math.cos(lat2Rad) * Math.cos(deltaLngRad);

  const bearingRad = Math.atan2(y, x);
  const bearingDeg = (bearingRad * 180) / Math.PI;
  
  return (bearingDeg + 360) % 360; // Normalize to 0-360 degrees
}

/**
 * Calculate area of a polygon using the spherical excess formula
 */
export function calculatePolygonArea(points: Point[]): number {
  if (points.length < 3) return 0;

  // Close the polygon if not already closed
  const closedPoints = [...points];
  if (points[0].lat !== points[points.length - 1].lat || 
      points[0].lng !== points[points.length - 1].lng) {
    closedPoints.push(points[0]);
  }

  // Convert to radians
  const pointsRad = closedPoints.map(p => ({
    lat: (p.lat * Math.PI) / 180,
    lng: (p.lng * Math.PI) / 180
  }));

  // Calculate area using spherical excess
  let area = 0;
  const n = pointsRad.length - 1;
  
  for (let i = 0; i < n; i++) {
    const p1 = pointsRad[i];
    const p2 = pointsRad[i + 1];
    area += (p2.lng - p1.lng) * Math.sin((p1.lat + p2.lat) / 2);
  }
  
  return Math.abs(area * EARTH_RADIUS * EARTH_RADIUS / 2);
}

/**
 * Calculate area of a circle
 */
export function calculateCircleArea(radius: number): number {
  return Math.PI * radius * radius;
}

/**
 * Calculate area of a rectangle defined by two opposite corners
 */
export function calculateRectangleArea(corner1: Point, corner2: Point): number {
  const width = calculateDistance(corner1, { lat: corner1.lat, lng: corner2.lng });
  const height = calculateDistance(corner1, { lat: corner2.lat, lng: corner1.lng });
  return width * height;
}

/**
 * Format distance for display
 */
export function formatDistance(meters: number): string {
  if (meters < 1000) {
    return `${Math.round(meters)} m`;
  } else if (meters < 10000) {
    return `${(meters / 1000).toFixed(1)} km`;
  } else {
    return `${Math.round(meters / 1000)} km`;
  }
}

/**
 * Format area for display
 */
export function formatArea(squareMeters: number): string {
  if (squareMeters < 10000) {
    return `${Math.round(squareMeters)} m²`;
  } else if (squareMeters < 1000000) {
    return `${(squareMeters / 10000).toFixed(1)} ha`;
  } else {
    return `${(squareMeters / 1000000).toFixed(1)} km²`;
  }
}

/**
 * Format bearing for display
 */
export function formatBearing(degrees: number): string {
  const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE',
                     'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
  const index = Math.round(degrees / 22.5) % 16;
  return `${Math.round(degrees)}° ${directions[index]}`;
}

/**
 * Format coordinates for display
 */
export function formatCoordinates(point: Point, format: 'dd' | 'dms' | 'mgrs' = 'dd'): string {
  switch (format) {
    case 'dms': {
      const latDeg = Math.floor(Math.abs(point.lat));
      const latMin = Math.floor((Math.abs(point.lat) - latDeg) * 60);
      const latSec = ((Math.abs(point.lat) - latDeg) * 60 - latMin) * 60;
      const latDir = point.lat >= 0 ? 'N' : 'S';
      
      const lngDeg = Math.floor(Math.abs(point.lng));
      const lngMin = Math.floor((Math.abs(point.lng) - lngDeg) * 60);
      const lngSec = ((Math.abs(point.lng) - lngDeg) * 60 - lngMin) * 60;
      const lngDir = point.lng >= 0 ? 'E' : 'W';
      
      return `${latDeg}°${latMin}'${latSec.toFixed(2)}"${latDir} ${lngDeg}°${lngMin}'${lngSec.toFixed(2)}"${lngDir}`;
    }
    case 'mgrs':
      // Simplified MGRS format - would need full MGRS conversion library for accuracy
      return `${point.lat.toFixed(6)}, ${point.lng.toFixed(6)} (approx MGRS)`;
    case 'dd':
    default:
      return `${point.lat.toFixed(6)}, ${point.lng.toFixed(6)}`;
  }
}

/**
 * Convert Leaflet LatLng to Point
 */
export function latLngToPoint(latLng: L.LatLng): Point {
  return { lat: latLng.lat, lng: latLng.lng };
}

/**
 * Convert Point to Leaflet LatLng
 */
export function pointToLatLng(point: Point): L.LatLng {
  return L.latLng(point.lat, point.lng);
}

/**
 * Create measurement tooltip content
 */
export function createMeasurementTooltip(
  type: 'distance' | 'area' | 'bearing',
  value: number,
  additionalInfo?: string
): string {
  let content = '';
  
  switch (type) {
    case 'distance':
      content = `Distance: ${formatDistance(value)}`;
      break;
    case 'area':
      content = `Area: ${formatArea(value)}`;
      break;
    case 'bearing':
      content = `Bearing: ${formatBearing(value)}`;
      break;
  }
  
  if (additionalInfo) {
    content += `<br/>${additionalInfo}`;
  }
  
  return content;
}

/**
 * Check if point is inside circle
 */
export function isPointInCircle(point: Point, center: Point, radius: number): boolean {
  const distance = calculateDistance(point, center);
  return distance <= radius;
}

/**
 * Check if point is inside polygon using ray casting algorithm
 */
export function isPointInPolygon(point: Point, polygon: Point[]): boolean {
  let inside = false;
  const x = point.lng;
  const y = point.lat;
  
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
}

/**
 * Get center point of polygon
 */
export function getPolygonCenter(points: Point[]): Point {
  let lat = 0;
  let lng = 0;
  
  for (const point of points) {
    lat += point.lat;
    lng += point.lng;
  }
  
  return {
    lat: lat / points.length,
    lng: lng / points.length
  };
}

/**
 * Convert degrees to radians
 */
export function toRadians(degrees: number): number {
  return (degrees * Math.PI) / 180;
}

/**
 * Convert radians to degrees
 */
export function toDegrees(radians: number): number {
  return (radians * 180) / Math.PI;
}

/**
 * Calculate bounding box for points
 */
export function calculateBounds(points: Point[]): L.LatLngBounds {
  if (points.length === 0) {
    throw new Error('Cannot calculate bounds for empty points array');
  }

  let minLat = points[0].lat;
  let maxLat = points[0].lat;
  let minLng = points[0].lng;
  let maxLng = points[0].lng;

  for (const point of points) {
    minLat = Math.min(minLat, point.lat);
    maxLat = Math.max(maxLat, point.lat);
    minLng = Math.min(minLng, point.lng);
    maxLng = Math.max(maxLng, point.lng);
  }

  return L.latLngBounds([minLat, minLng], [maxLat, maxLng]);
}
