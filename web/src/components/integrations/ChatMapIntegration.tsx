import React, { useEffect, useState, useCallback } from 'react';
import { locationSharingService, type SharedLocation } from '../../services/locationSharingService';
import type { ChatMessage } from '../../types/chat';
import type { UseTacticalOverlaysReturn } from '../../hooks/useTacticalOverlays';

interface ChatMapIntegrationProps {
  overlayManager: UseTacticalOverlaysReturn;
  map: L.Map | null;
  onLocationShared?: (location: SharedLocation, messageId: string) => void;
  onMapLocationClicked?: (location: SharedLocation) => void;
}

export const ChatMapIntegration: React.FC<ChatMapIntegrationProps> = ({
  overlayManager,
  map,
  onLocationShared,
  onMapLocationClicked,
}) => {
  const [sharedLocationLayer, setSharedLocationLayer] = useState<string>('shared_locations');
  
  // Initialize shared locations layer
  useEffect(() => {
    if (overlayManager) {
      // Create a dedicated layer for shared locations
      const layer = overlayManager.createLayer(
        'Shared Locations',
        'Locations shared through chat messages'
      );
      setSharedLocationLayer(layer.id);
    }
  }, [overlayManager]);
  
  // Handle location shared from chat
  const handleLocationShared = useCallback((location: SharedLocation, messageId: string) => {
    // Create a temporary message object for the overlay
    const tempMessage: ChatMessage = {
      id: messageId,
      roomId: '',
      senderId: '',
      senderCallsign: 'Chat User',
      messageText: `Location shared at ${locationSharingService.formatLocationForDisplay(location)}`,
      messageType: 'position',
      priority: 'normal',
      classification: 'UNCLASSIFIED',
      locationLat: location.lat,
      locationLng: location.lng,
      locationAlt: location.alt,
      isDeleted: false,
      createdAt: location.timestamp,
      updatedAt: location.timestamp,
      requiresAck: false,
      isAcknowledged: false,
    };
    
    // Create tactical overlay from the location
    const overlay = locationSharingService.createOverlayFromLocation(location, tempMessage);
    
    // Add to the shared locations layer
    overlayManager.addOverlay(overlay, sharedLocationLayer);
    
    // Pan map to the location
    if (map) {
      map.setView([location.lat, location.lng], Math.max(map.getZoom(), 15));
    }
    
    // Call external callback
    onLocationShared?.(location, messageId);
  }, [overlayManager, map, sharedLocationLayer, onLocationShared]);
  
  // Handle location clicked on map
  const handleMapLocationClicked = useCallback((location: SharedLocation) => {
    // Call external callback to potentially share this location in chat
    onMapLocationClicked?.(location);
  }, [onMapLocationClicked]);
  
  // Handle showing location on map (from chat message)
  const handleShowLocationOnMap = useCallback((location: SharedLocation, message: ChatMessage) => {
    if (!map) return;
    
    // Pan to location
    map.setView([location.lat, location.lng], Math.max(map.getZoom(), 15));
    
    // Check if overlay already exists
    const existingOverlay = overlayManager.overlays.find(
      overlay => overlay.metadata.messageId === message.id
    );
    
    if (!existingOverlay) {
      // Create overlay if it doesn't exist
      const overlay = locationSharingService.createOverlayFromLocation(location, message);
      overlayManager.addOverlay(overlay, sharedLocationLayer);
    } else {
      // Select existing overlay
      overlayManager.selectOverlay(existingOverlay.id);
    }
  }, [map, overlayManager, sharedLocationLayer]);
  
  // Handle going to location (from chat message)
  const handleGoToLocation = useCallback((location: SharedLocation, message: ChatMessage) => {
    if (!map) return;
    
    // Pan to location with higher zoom
    map.setView([location.lat, location.lng], 18);
    
    // Create temporary highlight overlay
    const highlightOverlay = {
      ...locationSharingService.createOverlayFromLocation(location, message),
      id: `temp_highlight_${Date.now()}`,
      name: 'Location Highlight',
      style: {
        color: '#ff6b6b',
        weight: 3,
        opacity: 1,
        fillColor: '#ff6b6b',
        fillOpacity: 0.3,
        className: 'location-highlight-marker',
      },
    };
    
    overlayManager.addOverlay(highlightOverlay, sharedLocationLayer);
    
    // Remove highlight after 5 seconds
    setTimeout(() => {
      overlayManager.deleteOverlay(highlightOverlay.id);
    }, 5000);
  }, [map, overlayManager, sharedLocationLayer]);
  
  // Handle right-click on map to share location
  const handleMapRightClick = useCallback((e: L.LeafletMouseEvent) => {
    const location = locationSharingService.createLocationFromMapClick(
      e.latlng.lat,
      e.latlng.lng
    );
    
    handleMapLocationClicked(location);
  }, [handleMapLocationClicked]);
  
  // Set up location sharing service callbacks
  useEffect(() => {
    locationSharingService.setLocationSharedCallback(handleLocationShared);
    locationSharingService.setLocationClickedCallback(handleMapLocationClicked);
    
    return () => {
      // Cleanup callbacks
      locationSharingService.setLocationSharedCallback(() => {});
      locationSharingService.setLocationClickedCallback(() => {});
    };
  }, [handleLocationShared, handleMapLocationClicked]);
  
  // Set up map event handlers
  useEffect(() => {
    if (!map) return;
    
    // Add context menu handler for location sharing
    map.on('contextmenu', handleMapRightClick);
    
    return () => {
      map.off('contextmenu', handleMapRightClick);
    };
  }, [map, handleMapRightClick]);
  
  // Auto-process chat messages for location sharing
  const processLocationMessages = useCallback((messages: ChatMessage[]) => {
    messages.forEach(message => {
      if (message.locationLat && message.locationLng) {
        // Check if we already have an overlay for this message
        const existingOverlay = overlayManager.overlays.find(
          overlay => overlay.metadata.messageId === message.id
        );
        
        if (!existingOverlay) {
          const location = locationSharingService.extractLocationFromMessage(message);
          if (location) {
            const overlay = locationSharingService.createOverlayFromLocation(location, message);
            overlayManager.addOverlay(overlay, sharedLocationLayer);
          }
        }
      }
    });
  }, [overlayManager, sharedLocationLayer]);
  
  return {
    // Exposed functions for chat components to use
    handleShowLocationOnMap,
    handleGoToLocation,
    processLocationMessages,
    sharedLocationLayer,
  };
};

// Hook version for easier integration
export function useChatMapIntegration(
  overlayManager: UseTacticalOverlaysReturn,
  map: L.Map | null
) {
  const [integration] = useState(() => 
    React.createElement(ChatMapIntegration, { overlayManager, map })
  );
  
  const handleShowLocationOnMap = useCallback((location: SharedLocation, message: ChatMessage) => {
    if (!map) return;
    
    // Pan to location
    map.setView([location.lat, location.lng], Math.max(map.getZoom(), 15));
    
    // Check if overlay already exists
    const existingOverlay = overlayManager.overlays.find(
      overlay => overlay.metadata.messageId === message.id
    );
    
    if (!existingOverlay) {
      // Create overlay if it doesn't exist
      const overlay = locationSharingService.createOverlayFromLocation(location, message);
      overlayManager.addOverlay(overlay);
    } else {
      // Select existing overlay
      overlayManager.selectOverlay(existingOverlay.id);
    }
  }, [map, overlayManager]);
  
  const handleGoToLocation = useCallback((location: SharedLocation, message: ChatMessage) => {
    if (!map) return;
    
    // Pan to location with higher zoom
    map.setView([location.lat, location.lng], 18);
    
    // Create temporary highlight overlay
    const highlightOverlay = {
      ...locationSharingService.createOverlayFromLocation(location, message),
      id: `temp_highlight_${Date.now()}`,
      name: 'Location Highlight',
      style: {
        color: '#ff6b6b',
        weight: 3,
        opacity: 1,
        fillColor: '#ff6b6b',
        fillOpacity: 0.3,
        className: 'location-highlight-marker',
      },
    };
    
    overlayManager.addOverlay(highlightOverlay);
    
    // Remove highlight after 5 seconds
    setTimeout(() => {
      overlayManager.deleteOverlay(highlightOverlay.id);
    }, 5000);
  }, [map, overlayManager]);
  
  const handleLocationSharedFromChat = useCallback((message: ChatMessage) => {
    locationSharingService.handleLocationShared(message);
  }, []);
  
  const handleMapRightClick = useCallback((lat: number, lng: number) => {
    const location = locationSharingService.createLocationFromMapClick(lat, lng);
    locationSharingService.handleLocationClicked(lat, lng);
    return location;
  }, []);
  
  // Set up map event handlers
  useEffect(() => {
    if (!map) return;
    
    const handleContextMenu = (e: L.LeafletMouseEvent) => {
      handleMapRightClick(e.latlng.lat, e.latlng.lng);
    };
    
    map.on('contextmenu', handleContextMenu);
    
    return () => {
      map.off('contextmenu', handleContextMenu);
    };
  }, [map, handleMapRightClick]);
  
  return {
    handleShowLocationOnMap,
    handleGoToLocation,
    handleLocationSharedFromChat,
    handleMapRightClick,
  };
}

export default ChatMapIntegration;
