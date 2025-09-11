import { useState, useCallback, useRef } from 'react';
import L from 'leaflet';
import type { 
  TacticalOverlay, 
  OverlayLayer, 
  OverlayManager, 
  DrawingToolType,
  TacticalStyle,
  TacticalGeometry,
  ThreatCircle,
  RangeRing,
  TacticalRoute,
  TacticalArea
} from '../types/tactical';
import { DEFAULT_STYLES } from '../types/tactical';

export interface UseTacticalOverlaysOptions {
  map?: L.Map | null;
  onOverlayCreated?: (overlay: TacticalOverlay) => void;
  onOverlayUpdated?: (overlay: TacticalOverlay) => void;
  onOverlayDeleted?: (overlayId: string) => void;
}

export interface UseTacticalOverlaysReturn {
  overlayManager: OverlayManager;
  overlays: TacticalOverlay[];
  activeLayer: OverlayLayer | undefined;
  selectedOverlay: TacticalOverlay | undefined;
  
  // Layer management
  createLayer: (name: string, description?: string) => OverlayLayer;
  deleteLayer: (layerId: string) => void;
  setActiveLayer: (layerId: string) => void;
  toggleLayerVisibility: (layerId: string) => void;
  setLayerOpacity: (layerId: string, opacity: number) => void;
  
  // Overlay management
  addOverlay: (overlay: TacticalOverlay, layerId?: string) => void;
  updateOverlay: (overlayId: string, updates: Partial<TacticalOverlay>) => void;
  deleteOverlay: (overlayId: string) => void;
  selectOverlay: (overlayId: string) => void;
  deselectOverlay: () => void;
  
  // Drawing mode
  setDrawingMode: (enabled: boolean) => void;
  setActiveDrawingTool: (toolType: DrawingToolType | undefined) => void;
  
  // Utility functions
  getOverlayById: (overlayId: string) => TacticalOverlay | undefined;
  getLayerById: (layerId: string) => OverlayLayer | undefined;
  clearAllOverlays: () => void;
  exportOverlays: () => string;
  importOverlays: (data: string) => void;
}

export function useTacticalOverlays(options: UseTacticalOverlaysOptions = {}): UseTacticalOverlaysReturn {
  const { map, onOverlayCreated, onOverlayUpdated, onOverlayDeleted } = options;
  
  // State
  const [overlayManager, setOverlayManager] = useState<OverlayManager>(() => ({
    layers: [
      {
        id: 'default',
        name: 'Default Layer',
        description: 'Default tactical overlay layer',
        overlays: [],
        visible: true,
        locked: false,
        opacity: 1.0,
        zIndex: 1000,
      }
    ],
    activeLayerId: 'default',
    drawingMode: false,
  }));
  
  // Refs for Leaflet layers
  const leafletLayersRef = useRef<Map<string, L.Layer>>(new Map());
  
  // Computed values
  const overlays = overlayManager.layers.flatMap(layer => layer.overlays);
  const activeLayer = overlayManager.layers.find(layer => layer.id === overlayManager.activeLayerId);
  const selectedOverlay = overlays.find(overlay => overlay.id === overlayManager.selectedOverlayId);
  
  // Layer management functions
  const createLayer = useCallback((name: string, description?: string): OverlayLayer => {
    const newLayer: OverlayLayer = {
      id: `layer_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name,
      description,
      overlays: [],
      visible: true,
      locked: false,
      opacity: 1.0,
      zIndex: overlayManager.layers.length + 1000,
    };
    
    setOverlayManager(prev => ({
      ...prev,
      layers: [...prev.layers, newLayer],
      activeLayerId: newLayer.id,
    }));
    
    return newLayer;
  }, [overlayManager.layers.length]);
  
  const deleteLayer = useCallback((layerId: string) => {
    if (layerId === 'default') return; // Don't delete default layer
    
    setOverlayManager(prev => {
      const layerToDelete = prev.layers.find(layer => layer.id === layerId);
      if (layerToDelete) {
        // Remove Leaflet layers
        layerToDelete.overlays.forEach(overlay => {
          const leafletLayer = leafletLayersRef.current.get(overlay.id);
          if (leafletLayer && map) {
            map.removeLayer(leafletLayer);
            leafletLayersRef.current.delete(overlay.id);
          }
        });
      }
      
      const newLayers = prev.layers.filter(layer => layer.id !== layerId);
      const newActiveLayerId = prev.activeLayerId === layerId 
        ? newLayers[0]?.id || 'default' 
        : prev.activeLayerId;
      
      return {
        ...prev,
        layers: newLayers,
        activeLayerId: newActiveLayerId,
        selectedOverlayId: undefined,
      };
    });
  }, [map]);
  
  const setActiveLayer = useCallback((layerId: string) => {
    setOverlayManager(prev => ({
      ...prev,
      activeLayerId: layerId,
    }));
  }, []);
  
  const toggleLayerVisibility = useCallback((layerId: string) => {
    setOverlayManager(prev => {
      const newLayers = prev.layers.map(layer => {
        if (layer.id === layerId) {
          const newLayer = { ...layer, visible: !layer.visible };
          
          // Update Leaflet layer visibility
          newLayer.overlays.forEach(overlay => {
            const leafletLayer = leafletLayersRef.current.get(overlay.id);
            if (leafletLayer && map) {
              if (newLayer.visible && overlay.visible) {
                map.addLayer(leafletLayer);
              } else {
                map.removeLayer(leafletLayer);
              }
            }
          });
          
          return newLayer;
        }
        return layer;
      });
      
      return { ...prev, layers: newLayers };
    });
  }, [map]);
  
  const setLayerOpacity = useCallback((layerId: string, opacity: number) => {
    setOverlayManager(prev => {
      const newLayers = prev.layers.map(layer => {
        if (layer.id === layerId) {
          const newLayer = { ...layer, opacity };
          
          // Update Leaflet layer opacity
          newLayer.overlays.forEach(overlay => {
            const leafletLayer = leafletLayersRef.current.get(overlay.id);
            if (leafletLayer && map) {
              if ('setStyle' in leafletLayer) {
                (leafletLayer as any).setStyle({ opacity });
              }
            }
          });
          
          return newLayer;
        }
        return layer;
      });
      
      return { ...prev, layers: newLayers };
    });
  }, [map]);
  
  // Create Leaflet layer from tactical overlay
  const createLeafletLayer = useCallback((overlay: TacticalOverlay): L.Layer | null => {
    if (!map) return null;
    
    const { geometry, style } = overlay;
    const leafletStyle = {
      color: style.color,
      weight: style.weight,
      opacity: style.opacity,
      fillColor: style.fillColor,
      fillOpacity: style.fillOpacity,
      dashArray: style.dashArray,
      lineCap: style.lineCap,
      lineJoin: style.lineJoin,
      className: style.className,
    };
    
    let layer: L.Layer | null = null;
    
    switch (geometry.type) {
      case 'Point': {
        const coords = geometry.coordinates as [number, number];
        layer = L.marker([coords[0], coords[1]]);
        break;
      }
      case 'LineString': {
        const coords = geometry.coordinates as [number, number][];
        layer = L.polyline(coords, leafletStyle);
        break;
      }
      case 'Polygon': {
        const coords = geometry.coordinates as [number, number][];
        layer = L.polygon(coords, leafletStyle);
        break;
      }
      case 'Circle': {
        const coords = geometry.coordinates as [number, number];
        const radius = geometry.radius || 1000;
        layer = L.circle([coords[0], coords[1]], { 
          ...leafletStyle,
          radius 
        });
        break;
      }
    }
    
    if (layer) {
      // Add popup with overlay information
      const popupContent = `
        <div class="tactical-overlay-popup">
          <h4>${overlay.name}</h4>
          ${overlay.description ? `<p>${overlay.description}</p>` : ''}
          <div class="overlay-meta">
            <div><strong>Type:</strong> ${overlay.type}</div>
            <div><strong>Priority:</strong> ${overlay.metadata.priority}</div>
            ${overlay.metadata.classification ? `<div><strong>Classification:</strong> ${overlay.metadata.classification}</div>` : ''}
          </div>
        </div>
      `;
      
      layer.bindPopup(popupContent);
      
      // Add click handler
      layer.on('click', () => {
        selectOverlay(overlay.id);
      });
    }
    
    return layer;
  }, [map]);
  
  // Overlay management functions
  const addOverlay = useCallback((overlay: TacticalOverlay, layerId?: string) => {
    const targetLayerId = layerId || overlayManager.activeLayerId || 'default';
    
    // Create Leaflet layer
    const leafletLayer = createLeafletLayer(overlay);
    if (leafletLayer) {
      leafletLayersRef.current.set(overlay.id, leafletLayer);
      
      // Add to map if layer is visible
      const targetLayer = overlayManager.layers.find(layer => layer.id === targetLayerId);
      if (targetLayer?.visible && overlay.visible && map) {
        map.addLayer(leafletLayer);
      }
    }
    
    setOverlayManager(prev => {
      const newLayers = prev.layers.map(layer => {
        if (layer.id === targetLayerId) {
          return {
            ...layer,
            overlays: [...layer.overlays, overlay],
          };
        }
        return layer;
      });
      
      return { ...prev, layers: newLayers };
    });
    
    onOverlayCreated?.(overlay);
  }, [overlayManager.activeLayerId, overlayManager.layers, createLeafletLayer, map, onOverlayCreated]);
  
  const updateOverlay = useCallback((overlayId: string, updates: Partial<TacticalOverlay>) => {
    setOverlayManager(prev => {
      const newLayers = prev.layers.map(layer => ({
        ...layer,
        overlays: layer.overlays.map(overlay => {
          if (overlay.id === overlayId) {
            const updatedOverlay = { ...overlay, ...updates };
            
            // Update Leaflet layer
            const leafletLayer = leafletLayersRef.current.get(overlayId);
            if (leafletLayer) {
              // Remove old layer
              if (map) map.removeLayer(leafletLayer);
              leafletLayersRef.current.delete(overlayId);
              
              // Create new layer with updates
              const newLeafletLayer = createLeafletLayer(updatedOverlay);
              if (newLeafletLayer) {
                leafletLayersRef.current.set(overlayId, newLeafletLayer);
                if (map && layer.visible && updatedOverlay.visible) {
                  map.addLayer(newLeafletLayer);
                }
              }
            }
            
            onOverlayUpdated?.(updatedOverlay);
            return updatedOverlay;
          }
          return overlay;
        }),
      }));
      
      return { ...prev, layers: newLayers };
    });
  }, [map, createLeafletLayer, onOverlayUpdated]);
  
  const deleteOverlay = useCallback((overlayId: string) => {
    // Remove Leaflet layer
    const leafletLayer = leafletLayersRef.current.get(overlayId);
    if (leafletLayer && map) {
      map.removeLayer(leafletLayer);
      leafletLayersRef.current.delete(overlayId);
    }
    
    setOverlayManager(prev => {
      const newLayers = prev.layers.map(layer => ({
        ...layer,
        overlays: layer.overlays.filter(overlay => overlay.id !== overlayId),
      }));
      
      return {
        ...prev,
        layers: newLayers,
        selectedOverlayId: prev.selectedOverlayId === overlayId ? undefined : prev.selectedOverlayId,
      };
    });
    
    onOverlayDeleted?.(overlayId);
  }, [map, onOverlayDeleted]);
  
  const selectOverlay = useCallback((overlayId: string) => {
    setOverlayManager(prev => ({
      ...prev,
      selectedOverlayId: overlayId,
    }));
  }, []);
  
  const deselectOverlay = useCallback(() => {
    setOverlayManager(prev => ({
      ...prev,
      selectedOverlayId: undefined,
    }));
  }, []);
  
  // Drawing mode functions
  const setDrawingMode = useCallback((enabled: boolean) => {
    setOverlayManager(prev => ({
      ...prev,
      drawingMode: enabled,
      activeDrawingTool: enabled ? prev.activeDrawingTool : undefined,
    }));
  }, []);
  
  const setActiveDrawingTool = useCallback((toolType: DrawingToolType | undefined) => {
    setOverlayManager(prev => ({
      ...prev,
      activeDrawingTool: toolType,
      drawingMode: toolType !== undefined,
    }));
  }, []);
  
  // Utility functions
  const getOverlayById = useCallback((overlayId: string): TacticalOverlay | undefined => {
    return overlays.find(overlay => overlay.id === overlayId);
  }, [overlays]);
  
  const getLayerById = useCallback((layerId: string): OverlayLayer | undefined => {
    return overlayManager.layers.find(layer => layer.id === layerId);
  }, [overlayManager.layers]);
  
  const clearAllOverlays = useCallback(() => {
    // Remove all Leaflet layers
    leafletLayersRef.current.forEach(leafletLayer => {
      if (map) map.removeLayer(leafletLayer);
    });
    leafletLayersRef.current.clear();
    
    setOverlayManager(prev => ({
      ...prev,
      layers: prev.layers.map(layer => ({
        ...layer,
        overlays: [],
      })),
      selectedOverlayId: undefined,
    }));
  }, [map]);
  
  const exportOverlays = useCallback((): string => {
    return JSON.stringify(overlayManager, null, 2);
  }, [overlayManager]);
  
  const importOverlays = useCallback((data: string) => {
    try {
      const imported = JSON.parse(data) as OverlayManager;
      
      // Clear existing overlays
      clearAllOverlays();
      
      // Set imported data
      setOverlayManager(imported);
      
      // Recreate Leaflet layers
      imported.layers.forEach(layer => {
        layer.overlays.forEach(overlay => {
          const leafletLayer = createLeafletLayer(overlay);
          if (leafletLayer) {
            leafletLayersRef.current.set(overlay.id, leafletLayer);
            if (map && layer.visible && overlay.visible) {
              map.addLayer(leafletLayer);
            }
          }
        });
      });
    } catch (error) {
      console.error('Failed to import overlays:', error);
    }
  }, [clearAllOverlays, createLeafletLayer, map]);
  
  return {
    overlayManager,
    overlays,
    activeLayer,
    selectedOverlay,
    
    // Layer management
    createLayer,
    deleteLayer,
    setActiveLayer,
    toggleLayerVisibility,
    setLayerOpacity,
    
    // Overlay management
    addOverlay,
    updateOverlay,
    deleteOverlay,
    selectOverlay,
    deselectOverlay,
    
    // Drawing mode
    setDrawingMode,
    setActiveDrawingTool,
    
    // Utility functions
    getOverlayById,
    getLayerById,
    clearAllOverlays,
    exportOverlays,
    importOverlays,
  };
}
