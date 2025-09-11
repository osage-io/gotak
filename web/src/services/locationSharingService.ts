/**
 * Service for handling location sharing between chat and map components
 */

import type { ChatMessage, SendMessageRequest } from '../types/chat';
import type { TacticalOverlay } from '../types/tactical';
import { generateId, getCurrentTimestamp, formatCoordinates } from '../utils/common';

export interface LocationShareOptions {
  includeAltitude?: boolean;
  requiresAck?: boolean;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
  customMessage?: string;
}

export interface SharedLocation {
  lat: number;
  lng: number;
  alt?: number;
  accuracy?: number;
  timestamp: string;
  source: 'gps' | 'map_click' | 'manual';
}

export interface LocationMessageData {
  location: SharedLocation;
  message: string;
  context?: string;
}

export class LocationSharingService {
  private static instance: LocationSharingService;
  
  // Callbacks for integration with other services
  private onLocationShared?: (location: SharedLocation, messageId: string) => void;
  private onLocationClicked?: (location: SharedLocation) => void;
  
  public static getInstance(): LocationSharingService {
    if (!LocationSharingService.instance) {
      LocationSharingService.instance = new LocationSharingService();
    }
    return LocationSharingService.instance;
  }
  
  /**
   * Set callback for when a location is shared in chat
   */
  public setLocationSharedCallback(callback: (location: SharedLocation, messageId: string) => void) {
    this.onLocationShared = callback;
  }
  
  /**
   * Set callback for when a location is clicked on the map
   */
  public setLocationClickedCallback(callback: (location: SharedLocation) => void) {
    this.onLocationClicked = callback;
  }
  
  /**
   * Get current user location using geolocation API
   */
  public async getCurrentLocation(options?: PositionOptions): Promise<SharedLocation> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('Geolocation is not supported by this browser'));
        return;
      }
      
      const defaultOptions: PositionOptions = {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 60000,
        ...options,
      };
      
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const location: SharedLocation = {
            lat: position.coords.latitude,
            lng: position.coords.longitude,
            alt: position.coords.altitude || undefined,
            accuracy: position.coords.accuracy,
            timestamp: getCurrentTimestamp(),
            source: 'gps',
          };
          resolve(location);
        },
        (error) => {
          reject(new Error(`Geolocation error: ${error.message}`));
        },
        defaultOptions
      );
    });
  }
  
  /**
   * Create a location from map click coordinates
   */
  public createLocationFromMapClick(lat: number, lng: number, alt?: number): SharedLocation {
    return {
      lat,
      lng,
      alt,
      timestamp: getCurrentTimestamp(),
      source: 'map_click',
    };
  }
  
  /**
   * Create a location from manual input
   */
  public createLocationFromInput(lat: number, lng: number, alt?: number): SharedLocation {
    return {
      lat,
      lng,
      alt,
      timestamp: getCurrentTimestamp(),
      source: 'manual',
    };
  }
  
  /**
   * Format location for display
   */
  public formatLocationForDisplay(location: SharedLocation): string {
    const coords = formatCoordinates(location.lat, location.lng, 6);
    const altitude = location.alt ? ` @ ${Math.round(location.alt)}m` : '';
    const accuracy = location.accuracy ? ` (±${Math.round(location.accuracy)}m)` : '';
    return `${coords}${altitude}${accuracy}`;
  }
  
  /**
   * Create a chat message with location data
   */
  public createLocationMessage(
    roomId: string,
    location: SharedLocation,
    options: LocationShareOptions = {}
  ): SendMessageRequest {
    const {
      includeAltitude = true,
      requiresAck = false,
      priority = 'normal',
      customMessage,
    } = options;
    
    const locationText = this.formatLocationForDisplay(location);
    const sourceText = this.getLocationSourceText(location.source);
    
    const messageText = customMessage || 
      `📍 Sharing location: ${locationText} ${sourceText}`;
    
    return {
      roomId,
      messageText,
      messageType: 'position',
      priority,
      classification: 'UNCLASSIFIED',
      locationLat: location.lat,
      locationLng: location.lng,
      locationAlt: includeAltitude ? location.alt : undefined,
      requiresAck,
      keywords: ['location', 'position'],
    };
  }
  
  /**
   * Extract location data from a chat message
   */
  public extractLocationFromMessage(message: ChatMessage): SharedLocation | null {
    if (!message.locationLat || !message.locationLng) {
      return null;
    }
    
    return {
      lat: message.locationLat,
      lng: message.locationLng,
      alt: message.locationAlt,
      timestamp: message.createdAt,
      source: message.messageType === 'position' ? 'gps' : 'manual',
    };
  }
  
  /**
   * Create a tactical overlay from a shared location
   */
  public createOverlayFromLocation(
    location: SharedLocation,
    message: ChatMessage
  ): TacticalOverlay {
    const senderName = message.senderCallsign || message.senderUsername || 'Unknown';
    
    return {
      id: generateId(),
      name: `${senderName}'s Location`,
      description: `Location shared by ${senderName}: ${message.messageText}`,
      type: 'marker',
      geometry: {
        type: 'Point',
        coordinates: [location.lat, location.lng],
      },
      style: {
        color: '#3b82f6',
        weight: 2,
        opacity: 1,
        fillColor: '#3b82f6',
        fillOpacity: 0.3,
        className: 'shared-location-marker',
      },
      visible: true,
      metadata: {
        createdAt: location.timestamp,
        updatedAt: getCurrentTimestamp(),
        priority: message.priority as any,
        source: 'chat_location',
        messageId: message.id,
        senderId: message.senderId,
        senderName,
        classification: message.classification,
      },
    };
  }
  
  /**
   * Handle location shared from chat
   */
  public handleLocationShared(message: ChatMessage) {
    const location = this.extractLocationFromMessage(message);
    if (location && this.onLocationShared) {
      this.onLocationShared(location, message.id);
    }
  }
  
  /**
   * Handle location clicked on map
   */
  public handleLocationClicked(lat: number, lng: number, alt?: number) {
    const location = this.createLocationFromMapClick(lat, lng, alt);
    if (this.onLocationClicked) {
      this.onLocationClicked(location);
    }
  }
  
  /**
   * Parse location from text (for manual input)
   */
  public parseLocationFromText(text: string): SharedLocation | null {
    // Try to parse various coordinate formats
    const patterns = [
      // Decimal degrees: "40.7128, -74.0060"
      /(-?\d+\.?\d*),?\s*(-?\d+\.?\d*)/,
      // With altitude: "40.7128, -74.0060, 100m"
      /(-?\d+\.?\d*),?\s*(-?\d+\.?\d*),?\s*(\d+\.?\d*)m?/,
      // MGRS or other formats could be added here
    ];
    
    for (const pattern of patterns) {
      const match = text.match(pattern);
      if (match) {
        const lat = parseFloat(match[1]);
        const lng = parseFloat(match[2]);
        const alt = match[3] ? parseFloat(match[3]) : undefined;
        
        // Basic validation
        if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
          return this.createLocationFromInput(lat, lng, alt);
        }
      }
    }
    
    return null;
  }
  
  /**
   * Get human-readable text for location source
   */
  private getLocationSourceText(source: SharedLocation['source']): string {
    switch (source) {
      case 'gps':
        return '(GPS)';
      case 'map_click':
        return '(Map)';
      case 'manual':
        return '(Manual)';
      default:
        return '';
    }
  }
  
  /**
   * Validate location coordinates
   */
  public validateLocation(lat: number, lng: number): boolean {
    return (
      typeof lat === 'number' && 
      typeof lng === 'number' &&
      lat >= -90 && 
      lat <= 90 && 
      lng >= -180 && 
      lng <= 180 &&
      !isNaN(lat) && 
      !isNaN(lng)
    );
  }
  
  /**
   * Calculate distance between two locations
   */
  public calculateDistance(loc1: SharedLocation, loc2: SharedLocation): number {
    const R = 6371000; // Earth's radius in meters
    const dLat = this.toRadians(loc2.lat - loc1.lat);
    const dLon = this.toRadians(loc2.lng - loc1.lng);
    
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
              Math.cos(this.toRadians(loc1.lat)) * Math.cos(this.toRadians(loc2.lat)) *
              Math.sin(dLon / 2) * Math.sin(dLon / 2);
    
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }
  
  private toRadians(degrees: number): number {
    return degrees * (Math.PI / 180);
  }
}

// Export singleton instance
export const locationSharingService = LocationSharingService.getInstance();
