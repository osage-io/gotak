import { useEffect, useCallback, useRef } from 'react';
import L from 'leaflet';
import 'leaflet-draw';
import type { 
  DrawingToolType, 
  TacticalOverlay, 
  TacticalGeometry, 
  UseTacticalOverlaysReturn 
} from './useTacticalOverlays';
import { DEFAULT_STYLES } from '../types/tactical';
import { generateId, getCurrentTimestamp } from '../utils/common';

// Extend Leaflet Draw options
declare module 'leaflet' {
  namespace Control {
    interface DrawConstructorOptions {
      position?: ControlPosition;
      draw?: any;
      edit?: any;
    }
  }
}

export interface UseDrawingHandlerOptions {
  map: L.Map | null;
  overlayManager: UseTacticalOverlaysReturn;
  activeDrawingTool?: DrawingToolType;
  onDrawComplete?: (overlay: TacticalOverlay) => void;
}

export interface UseDrawingHandlerReturn {
  isDrawing: boolean;
  startDrawing: (toolType: DrawingToolType) => void;
  stopDrawing: () => void;
  cancelDrawing: () => void;
}

export function useDrawingHandler(options: UseDrawingHandlerOptions): UseDrawingHandlerReturn {
  const { map, overlayManager, activeDrawingTool, onDrawComplete } = options;
  
  // Refs for Leaflet Draw
  const drawControlRef = useRef<L.Control.Draw | null>(null);
  const currentDrawingLayerRef = useRef<L.Layer | null>(null);
  const isDrawingRef = useRef(false);
  
  // Convert Leaflet layer to tactical geometry
  const layerToGeometry = useCallback((layer: L.Layer): TacticalGeometry | null => {
    if (layer instanceof L.Marker) {
      const latLng = layer.getLatLng();
      return {
        type: 'Point',
        coordinates: [latLng.lat, latLng.lng],
      };
    }
    
    if (layer instanceof L.Polyline && !(layer instanceof L.Polygon)) {
      const latLngs = layer.getLatLngs() as L.LatLng[];
      return {
        type: 'LineString',
        coordinates: latLngs.map(ll => [ll.lat, ll.lng]),
      };
    }
    
    if (layer instanceof L.Polygon && !(layer instanceof L.Rectangle)) {
      const latLngs = layer.getLatLngs()[0] as L.LatLng[];
      return {
        type: 'Polygon',
        coordinates: latLngs.map(ll => [ll.lat, ll.lng]),
      };
    }
    
    if (layer instanceof L.Rectangle) {
      const bounds = layer.getBounds();
      const ne = bounds.getNorthEast();
      const sw = bounds.getSouthWest();
      const nw = bounds.getNorthWest();
      const se = bounds.getSouthEast();
      
      return {
        type: 'Polygon',
        coordinates: [
          [ne.lat, ne.lng],
          [nw.lat, nw.lng],
          [sw.lat, sw.lng],
          [se.lat, se.lng],
          [ne.lat, ne.lng], // Close the polygon
        ],
      };
    }
    
    if (layer instanceof L.Circle) {
      const center = layer.getLatLng();
      const radius = layer.getRadius();
      return {
        type: 'Circle',
        coordinates: [center.lat, center.lng],
        radius,
      };
    }
    
    return null;
  }, []);
  
  // Get default overlay name based on tool type
  const getDefaultOverlayName = useCallback((toolType: DrawingToolType): string => {
    const typeNames: Record<DrawingToolType, string> = {
      marker: 'Marker',
      line: 'Line',
      polygon: 'Area',
      rectangle: 'Rectangle',
      circle: 'Circle',
      route: 'Route',
      boundary: 'Boundary',
      threat_circle: 'Threat Circle',
      range_ring: 'Range Ring',
      symbol: 'Symbol',
    };
    
    const baseName = typeNames[toolType] || 'Overlay';
    const existingCount = overlayManager.overlays.filter(o => 
      o.name.startsWith(baseName)
    ).length;
    
    return `${baseName} ${existingCount + 1}`;
  }, [overlayManager.overlays]);
  
  // Create tactical overlay from drawn layer
  const createTacticalOverlay = useCallback((
    layer: L.Layer, 
    toolType: DrawingToolType
  ): TacticalOverlay | null => {
    const geometry = layerToGeometry(layer);
    if (!geometry) return null;
    
    // Get style based on tool type
    let style = { ...DEFAULT_STYLES.default };
    let overlayType = toolType;
    
    switch (toolType) {
      case 'threat_circle':
        style = { ...DEFAULT_STYLES.threat };
        overlayType = 'threat_circle';
        break;
      case 'range_ring':
        style = { ...DEFAULT_STYLES.friendly };
        overlayType = 'range_ring';
        break;
      case 'route':
        style = { ...DEFAULT_STYLES.route };
        overlayType = 'tactical_route';
        break;
      case 'boundary':
        style = { ...DEFAULT_STYLES.boundary };
        overlayType = 'tactical_area';
        break;
      case 'symbol':
        overlayType = 'mil_symbol';
        break;
      default:
        // For basic shapes, use geometry type
        overlayType = geometry.type.toLowerCase() as any;
        break;
    }
    
    const overlay: TacticalOverlay = {
      id: generateId(),
      name: getDefaultOverlayName(toolType),
      type: overlayType,
      geometry,
      style,
      visible: true,
      metadata: {
        createdAt: getCurrentTimestamp(),
        updatedAt: getCurrentTimestamp(),
        priority: 'medium',
        source: 'user_drawn',
      },
    };
    
    return overlay;
  }, [layerToGeometry, getDefaultOverlayName]);
  
  // Initialize Leaflet Draw
  useEffect(() => {
    if (!map) return;
    
    // Create draw control with all tools disabled initially
    const drawControl = new L.Control.Draw({
      position: 'topleft',
      draw: {
        polyline: false,
        polygon: false,
        rectangle: false,
        circle: false,
        marker: false,
        circlemarker: false,
      },
      edit: false,
    });
    
    // Hide the draw control (we'll control it programmatically)
    const controlElement = drawControl.onAdd(map);
    if (controlElement) {
      controlElement.style.display = 'none';
    }
    
    drawControlRef.current = drawControl;
    
    // Set up event handlers
    const onDrawCreated = (e: any) => {
      const layer = e.layer;
      const toolType = activeDrawingTool;
      
      if (!toolType) return;
      
      // Create tactical overlay
      const overlay = createTacticalOverlay(layer, toolType);
      if (overlay) {
        // Add to overlay manager
        overlayManager.addOverlay(overlay);
        onDrawComplete?.(overlay);
      }
      
      // Clean up
      isDrawingRef.current = false;
      currentDrawingLayerRef.current = null;
      
      // Disable drawing mode
      overlayManager.setActiveDrawingTool(undefined);
    };
    
    const onDrawStart = (e: any) => {
      isDrawingRef.current = true;
      currentDrawingLayerRef.current = e.layer;
    };
    
    const onDrawStop = () => {
      isDrawingRef.current = false;
      currentDrawingLayerRef.current = null;
    };
    
    // Add event listeners
    map.on(L.Draw.Event.CREATED, onDrawCreated);
    map.on(L.Draw.Event.DRAWSTART, onDrawStart);
    map.on(L.Draw.Event.DRAWSTOP, onDrawStop);
    
    return () => {
      // Clean up
      map.off(L.Draw.Event.CREATED, onDrawCreated);
      map.off(L.Draw.Event.DRAWSTART, onDrawStart);
      map.off(L.Draw.Event.DRAWSTOP, onDrawStop);
      
      if (drawControlRef.current) {
        map.removeControl(drawControlRef.current);
        drawControlRef.current = null;
      }
    };
  }, [map, activeDrawingTool, createTacticalOverlay, overlayManager, onDrawComplete]);
  
  // Start drawing with specified tool
  const startDrawing = useCallback((toolType: DrawingToolType) => {
    if (!map || !drawControlRef.current) return;
    
    // Stop any current drawing
    if (isDrawingRef.current) {
      stopDrawing();
    }
    
    // Configure draw options based on tool type
    const drawOptions: any = {
      shapeOptions: {
        color: '#2563eb',
        weight: 2,
        opacity: 0.8,
        fillOpacity: 0.1,
      },
    };
    
    // Customize options based on tool type
    switch (toolType) {
      case 'threat_circle':
      case 'range_ring':
        drawOptions.shapeOptions.color = toolType === 'threat_circle' ? '#dc2626' : '#059669';
        break;
      case 'route':
        drawOptions.shapeOptions.color = '#7c3aed';
        drawOptions.shapeOptions.fillOpacity = 0;
        break;
      case 'boundary':
        drawOptions.shapeOptions.color = '#f59e0b';
        drawOptions.shapeOptions.fillOpacity = 0.2;
        drawOptions.shapeOptions.dashArray = '5, 5';
        break;
    }
    
    // Start the appropriate drawing handler
    let drawHandler: L.Draw.Feature | null = null;
    
    switch (toolType) {
      case 'marker':
      case 'symbol':
        drawHandler = new L.Draw.Marker(map, {
          icon: L.divIcon({
            className: 'tactical-marker',
            html: '📍',
            iconSize: [20, 20],
          }),
        });
        break;
      case 'line':
      case 'route':
        drawHandler = new L.Draw.Polyline(map, drawOptions);
        break;
      case 'polygon':
      case 'boundary':
        drawHandler = new L.Draw.Polygon(map, drawOptions);
        break;
      case 'rectangle':
        drawHandler = new L.Draw.Rectangle(map, drawOptions);
        break;
      case 'circle':
      case 'threat_circle':
      case 'range_ring':
        drawHandler = new L.Draw.Circle(map, drawOptions);
        break;
    }
    
    if (drawHandler) {
      drawHandler.enable();
      overlayManager.setActiveDrawingTool(toolType);
    }
  }, [map, overlayManager]);
  
  // Stop current drawing
  const stopDrawing = useCallback(() => {
    if (!map || !isDrawingRef.current) return;
    
    // Disable all drawing handlers
    map.fire(L.Draw.Event.DRAWSTOP);
    
    // Reset state
    isDrawingRef.current = false;
    currentDrawingLayerRef.current = null;
    overlayManager.setActiveDrawingTool(undefined);
  }, [map, overlayManager]);
  
  // Cancel current drawing
  const cancelDrawing = useCallback(() => {
    if (!map || !isDrawingRef.current) return;
    
    // Fire cancel event
    map.fire(L.Draw.Event.DRAWSTOP);
    
    // Remove any temporary drawing layer
    if (currentDrawingLayerRef.current) {
      map.removeLayer(currentDrawingLayerRef.current);
    }
    
    // Reset state
    isDrawingRef.current = false;
    currentDrawingLayerRef.current = null;
    overlayManager.setActiveDrawingTool(undefined);
  }, [map, overlayManager]);
  
  // Handle active tool changes
  useEffect(() => {
    if (activeDrawingTool) {
      startDrawing(activeDrawingTool);
    } else if (isDrawingRef.current) {
      stopDrawing();
    }
  }, [activeDrawingTool, startDrawing, stopDrawing]);
  
  return {
    isDrawing: isDrawingRef.current,
    startDrawing,
    stopDrawing,
    cancelDrawing,
  };
}

export default useDrawingHandler;
