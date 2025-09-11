import { useEffect, useRef, useState, useCallback } from 'react';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import 'leaflet-draw/dist/leaflet.draw.css';
import './TacticalMap.css';
import { formatCoordinates, type LatLng, type CoordinateDisplayOptions } from '../../utils/coordinates';
import type { EntityPosition, PositionUpdate } from '../../types/position';
import { ENTITY_MARKER_CONFIGS } from '../../types/position';
import { useWebSocket } from '../../hooks/useWebSocket';
import { positionService } from '../../services/positionService';
import { mappingService, type Route, type Geofence, type Point, type RouteOptions } from '../../services/mappingService';
import {
  calculateDistance,
  calculateBearing,
  calculatePolygonArea,
  calculateCircleArea,
  calculateRectangleArea,
  formatDistance,
  formatArea,
  formatBearing,
  latLngToPoint,
  pointToLatLng,
  createMeasurementTooltip
} from '../../utils/mappingUtils';

// Fix for default markers in Leaflet
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';

delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
});

export interface TacticalMapProps {
  initialCenter?: LatLng;
  initialZoom?: number;
  height?: string;
  width?: string;
  className?: string;
  onMapClick?: (latLng: LatLng) => void;
  onMapMove?: (center: LatLng, zoom: number) => void;
  showCoordinates?: boolean;
  coordinateFormat?: CoordinateDisplayOptions['format'];
  showScale?: boolean;
  enableDrawing?: boolean;
  readOnly?: boolean;
  // Entity tracking
  showEntities?: boolean;
  showTrails?: boolean;
  showFriendlyOnly?: boolean;
  showHostileOnly?: boolean;
  autoCenter?: boolean;
  onEntityClick?: (entity: EntityPosition) => void;
  // Advanced mapping features
  enableRouting?: boolean;
  enableGeofencing?: boolean;
  enableMeasurement?: boolean;
  showRoutes?: boolean;
  showGeofences?: boolean;
  onRouteCreated?: (route: Route) => void;
  onGeofenceCreated?: (geofence: Geofence) => void;
  onMeasurement?: (type: string, value: number, points: LatLng[]) => void;
}

export type DrawingMode = 'none' | 'route' | 'geofence-circle' | 'geofence-polygon' | 'geofence-rectangle' | 'measure-distance' | 'measure-area' | 'measure-bearing';

export type MapMode = 'view' | 'drawing' | 'measuring';

export interface MapInteraction {
  mode: MapMode;
  drawingMode: DrawingMode;
  isActive: boolean;
  tempName?: string;
  tempDescription?: string;
}

export interface MapLayer {
  id: string;
  name: string;
  url: string;
  attribution: string;
  maxZoom?: number;
  active: boolean;
}

const DEFAULT_LAYERS: MapLayer[] = [
  {
    id: 'osm',
    name: 'OpenStreetMap',
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '© OpenStreetMap contributors',
    maxZoom: 19,
    active: true,
  },
  {
    id: 'satellite',
    name: 'Satellite',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: '© Esri, Maxar, Earthstar Geographics',
    maxZoom: 18,
    active: false,
  },
  {
    id: 'topo',
    name: 'Topographic',
    url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
    attribution: '© OpenTopoMap contributors',
    maxZoom: 17,
    active: false,
  },
];

export function TacticalMap({
  initialCenter = { lat: 39.8283, lng: -98.5795 }, // Geographic center of US
  initialZoom = 4,
  height = '100%',
  width = '100%',
  className = '',
  onMapClick,
  onMapMove,
  showCoordinates = true,
  coordinateFormat = 'dd',
  showScale = true,
  enableDrawing = false,
  readOnly = false,
  // Entity tracking props
  showEntities = true,
  showTrails: _showTrails = false,
  showFriendlyOnly = false,
  showHostileOnly = false,
  autoCenter = false,
  onEntityClick,
  // Advanced mapping features
  enableRouting = false,
  enableGeofencing = false,
  enableMeasurement = false,
  showRoutes = true,
  showGeofences = true,
  onRouteCreated,
  onGeofenceCreated,
  onMeasurement,
}: TacticalMapProps) {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);
  const [currentCenter, setCurrentCenter] = useState<LatLng>(initialCenter);
  const [currentZoom, setCurrentZoom] = useState<number>(initialZoom);
  const [availableLayers] = useState<MapLayer[]>(DEFAULT_LAYERS);
  const [activeLayer, setActiveLayer] = useState<string>('osm');
  const [mousePosition, setMousePosition] = useState<LatLng | null>(null);
  
  // Entity tracking state
  const [entities, setEntities] = useState<Map<string, EntityPosition>>(new Map());
  const [entityMarkers, setEntityMarkers] = useState<Map<string, L.Marker>>(new Map());
  const [_entityTrails, _setEntityTrails] = useState<Map<string, L.Polyline>>(new Map()); // TODO: Implement trail lines
  const [isLoadingEntities, setIsLoadingEntities] = useState(false);
  
  // Advanced mapping state
  const [mapInteraction, setMapInteraction] = useState<MapInteraction>({
    mode: 'view',
    drawingMode: 'none',
    isActive: false
  });
  
  // Route management
  const [routes, setRoutes] = useState<Map<string, Route>>(new Map());
  const [routePolylines, setRoutePolylines] = useState<Map<string, L.Polyline>>(new Map());
  const [routeWaypoints, setRouteWaypoints] = useState<LatLng[]>([]);
  const [isCreatingRoute, setIsCreatingRoute] = useState(false);
  
  // Geofence management
  const [geofences, setGeofences] = useState<Map<string, Geofence>>(new Map());
  const [geofenceShapes, setGeofenceShapes] = useState<Map<string, L.Layer>>(new Map());
  
  // Drawing and measurement state
  const [drawingPoints, setDrawingPoints] = useState<LatLng[]>([]);
  const [measurementLayer, setMeasurementLayer] = useState<L.LayerGroup | null>(null);
  const [tempMarkers, setTempMarkers] = useState<L.Marker[]>([]);
  const [isDrawing, setIsDrawing] = useState(false);
  
  // Handle position updates from WebSocket
  const handlePositionUpdate = useCallback((update: PositionUpdate) => {
    const { entityId, position } = update;
    
    setEntities(prev => {
      const newEntities = new Map(prev);
      const existing = newEntities.get(entityId);
      
      // Update or create entity
      const updatedEntity: EntityPosition = {
        ...existing,
        entityId,
        lat: position.lat,
        lng: position.lng,
        altitude: position.altitude,
        speed: position.speed,
        course: position.course,
        lastUpdate: position.timestamp,
        isStale: false,
        // Preserve other fields from existing entity if available
        uid: existing?.uid || entityId,
        type: existing?.type || 'unknown',
        callsign: existing?.callsign || entityId,
        group: existing?.group || '',
        staleTime: existing?.staleTime || new Date(Date.now() + 5 * 60 * 1000).toISOString(),
        isFriendly: existing?.isFriendly || false,
        isHostile: existing?.isHostile || false,
      };
      
      newEntities.set(entityId, updatedEntity);
      return newEntities;
    });
  }, []);
  
  const handleEntityRemoved = useCallback((entityId: string) => {
    setEntities(prev => {
      const newEntities = new Map(prev);
      newEntities.delete(entityId);
      return newEntities;
    });
  }, []);
  
  // WebSocket connection
  const { connected: wsConnected, error: wsError } = useWebSocket(undefined, {
    onPositionUpdate: handlePositionUpdate,
    onEntityRemoved: handleEntityRemoved,
  });
  
  // Create marker for entity
  const createEntityMarker = useCallback((entity: EntityPosition): L.Marker => {
    const markerType = entity.isFriendly ? 'friendly' : 
                     entity.isHostile ? 'hostile' : 
                     'unknown';
    const config = ENTITY_MARKER_CONFIGS[markerType];
    
    // Create custom icon
    const icon = L.divIcon({
      className: `entity-marker ${markerType} ${entity.isStale ? 'stale' : ''}`,
      html: `
        <div class="marker-content" style="color: ${config.color}">
          <div class="marker-icon">●</div>
          <div class="marker-label">${entity.callsign}</div>
        </div>
      `,
      iconSize: config.size,
      iconAnchor: config.anchor,
      popupAnchor: config.popupAnchor,
    });
    
    const marker = L.marker([entity.lat, entity.lng], { icon });
    
    // Create popup content
    const popupContent = `
      <div class="entity-popup">
        <h3>${entity.callsign}</h3>
        <div class="entity-info">
          <div><strong>Type:</strong> ${entity.isFriendly ? 'Friendly' : entity.isHostile ? 'Hostile' : 'Unknown'}</div>
          <div><strong>Group:</strong> ${entity.group || 'N/A'}</div>
          <div><strong>Position:</strong> ${formatCoordinates(entity.lat, entity.lng, { format: coordinateFormat })}</div>
          ${entity.altitude ? `<div><strong>Altitude:</strong> ${entity.altitude.toFixed(0)}m</div>` : ''}
          ${entity.speed ? `<div><strong>Speed:</strong> ${entity.speed.toFixed(1)} m/s</div>` : ''}
          ${entity.course ? `<div><strong>Course:</strong> ${entity.course.toFixed(0)}°</div>` : ''}
          <div><strong>Last Update:</strong> ${new Date(entity.lastUpdate).toLocaleTimeString()}</div>
          <div class="status ${entity.isStale ? 'stale' : 'active'}">
            ${entity.isStale ? 'STALE' : 'ACTIVE'}
          </div>
        </div>
      </div>
    `;
    
    marker.bindPopup(popupContent);
    
    // Handle click events
    marker.on('click', () => {
      onEntityClick?.(entity);
    });
    
    return marker;
  }, [coordinateFormat, onEntityClick]);
  
  // Update entity markers on map
  const updateEntityMarkers = useCallback(() => {
    if (!mapInstanceRef.current || !showEntities) return;
    
    const map = mapInstanceRef.current;
    const currentMarkers = new Map(entityMarkers);
    const newMarkers = new Map<string, L.Marker>();
    
    // Filter entities based on display options
    const entitiesToShow = Array.from(entities.values()).filter(entity => {
      if (showFriendlyOnly && !entity.isFriendly) return false;
      if (showHostileOnly && !entity.isHostile) return false;
      return true;
    });
    
    // Update or create markers for entities
    entitiesToShow.forEach(entity => {
      const existingMarker = currentMarkers.get(entity.entityId);
      
      if (existingMarker) {
        // Update existing marker position and popup
        existingMarker.setLatLng([entity.lat, entity.lng]);
        
        // Update popup content
        const popupContent = `
          <div class="entity-popup">
            <h3>${entity.callsign}</h3>
            <div class="entity-info">
              <div><strong>Type:</strong> ${entity.isFriendly ? 'Friendly' : entity.isHostile ? 'Hostile' : 'Unknown'}</div>
              <div><strong>Group:</strong> ${entity.group || 'N/A'}</div>
              <div><strong>Position:</strong> ${formatCoordinates(entity.lat, entity.lng, { format: coordinateFormat })}</div>
              ${entity.altitude ? `<div><strong>Altitude:</strong> ${entity.altitude.toFixed(0)}m</div>` : ''}
              ${entity.speed ? `<div><strong>Speed:</strong> ${entity.speed.toFixed(1)} m/s</div>` : ''}
              ${entity.course ? `<div><strong>Course:</strong> ${entity.course.toFixed(0)}°</div>` : ''}
              <div><strong>Last Update:</strong> ${new Date(entity.lastUpdate).toLocaleTimeString()}</div>
              <div class="status ${entity.isStale ? 'stale' : 'active'}">
                ${entity.isStale ? 'STALE' : 'ACTIVE'}
              </div>
            </div>
          </div>
        `;
        
        existingMarker.setPopupContent(popupContent);
        newMarkers.set(entity.entityId, existingMarker);
        currentMarkers.delete(entity.entityId);
      } else {
        // Create new marker
        const marker = createEntityMarker(entity);
        marker.addTo(map);
        newMarkers.set(entity.entityId, marker);
      }
    });
    
    // Remove markers for entities that are no longer shown
    currentMarkers.forEach((marker) => {
      map.removeLayer(marker);
    });
    
    setEntityMarkers(newMarkers);
  }, [entities, entityMarkers, showEntities, showFriendlyOnly, showHostileOnly, createEntityMarker, coordinateFormat]);
  
  // Load initial entities
  const loadEntities = useCallback(async () => {
    if (!showEntities) return;
    
    setIsLoadingEntities(true);
    try {
      let positions: EntityPosition[];
      
      if (showFriendlyOnly) {
        positions = await positionService.getFriendlyPositions();
      } else if (showHostileOnly) {
        positions = await positionService.getHostilePositions();
      } else {
        positions = await positionService.getActivePositions();
      }
      
      const entitiesMap = new Map<string, EntityPosition>();
      positions.forEach(entity => {
        entitiesMap.set(entity.entityId, entity);
      });
      
      setEntities(entitiesMap);
    } catch (error) {
      console.error('Failed to load entities:', error);
    } finally {
      setIsLoadingEntities(false);
    }
  }, [showEntities, showFriendlyOnly, showHostileOnly]);
  
  // Load entities on mount and when filters change
  useEffect(() => {
    loadEntities();
  }, [loadEntities]);
  
  // Update markers when entities change
  useEffect(() => {
    updateEntityMarkers();
  }, [updateEntityMarkers]);
  
  // Auto-center map on entities if enabled
  useEffect(() => {
    if (!mapInstanceRef.current || !autoCenter || entities.size === 0) return;
    
    const entityArray = Array.from(entities.values());
    if (entityArray.length === 0) return;
    
    // Calculate bounds for all entities
    const bounds = L.latLngBounds(
      entityArray.map(entity => [entity.lat, entity.lng])
    );
    
    // Fit map to bounds with some padding
    mapInstanceRef.current.fitBounds(bounds, {
      padding: [20, 20],
      maxZoom: 15,
    });
  }, [entities, autoCenter]);

  // Initialize map
  useEffect(() => {
    if (!mapRef.current || mapInstanceRef.current) return;

    // Create map instance
    const map = L.map(mapRef.current, {
      center: [initialCenter.lat, initialCenter.lng],
      zoom: initialZoom,
      zoomControl: false, // We'll add custom controls
      attributionControl: true,
    });

    // Add zoom control to top-right
    L.control.zoom({ position: 'topright' }).addTo(map);

    // Add scale control if enabled
    if (showScale) {
      L.control.scale({
        position: 'bottomleft',
        metric: true,
        imperial: true,
      }).addTo(map);
    }

    // Add initial tile layer
    const initialLayer = availableLayers.find(layer => layer.id === activeLayer);
    if (initialLayer) {
      L.tileLayer(initialLayer.url, {
        attribution: initialLayer.attribution,
        maxZoom: initialLayer.maxZoom,
      }).addTo(map);
    }

    // Map event handlers
    map.on('click', (e: L.LeafletMouseEvent) => {
      const latLng = { lat: e.latlng.lat, lng: e.latlng.lng };
      
      // Handle different interaction modes
      if (mapInteraction.isActive) {
        if (mapInteraction.drawingMode === 'route') {
          // Add waypoint for route creation
          setRouteWaypoints(prev => [...prev, latLng]);
          
          // Add temporary marker
          const marker = L.marker([latLng.lat, latLng.lng], {
            icon: L.divIcon({
              className: 'route-waypoint-marker',
              html: `<div style="background: #3b82f6; color: white; border-radius: 50%; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; font-size: 12px;">${routeWaypoints.length + 1}</div>`,
              iconSize: [20, 20],
              iconAnchor: [10, 10]
            })
          }).addTo(map);
          
          setTempMarkers(prev => [...prev, marker]);
          
        } else if (mapInteraction.drawingMode.startsWith('geofence-')) {
          // Add point for geofence creation
          setDrawingPoints(prev => [...prev, latLng]);
          
          // Add temporary marker
          const marker = L.circleMarker([latLng.lat, latLng.lng], {
            color: '#ef4444',
            fillColor: '#ef4444',
            fillOpacity: 0.8,
            radius: 8
          }).addTo(map);
          
          setTempMarkers(prev => [...prev, marker]);
          
        } else if (mapInteraction.drawingMode.startsWith('measure-')) {
          // Add measurement point
          addMeasurementPoint(latLng);
        }
      } else {
        // Normal map click
        onMapClick?.(latLng);
      }
    });

    map.on('moveend', () => {
      const center = map.getCenter();
      const zoom = map.getZoom();
      const newCenter = { lat: center.lat, lng: center.lng };
      setCurrentCenter(newCenter);
      setCurrentZoom(zoom);
      onMapMove?.(newCenter, zoom);
    });

    map.on('mousemove', (e: L.LeafletMouseEvent) => {
      setMousePosition({ lat: e.latlng.lat, lng: e.latlng.lng });
    });

    map.on('mouseout', () => {
      setMousePosition(null);
    });

    mapInstanceRef.current = map;

    // Cleanup function
    return () => {
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }
    };
  }, [initialCenter, initialZoom, activeLayer, availableLayers, onMapClick, onMapMove, showScale]);

  // Handle layer switching
  const switchLayer = (layerId: string) => {
    if (mapInstanceRef.current && layerId !== activeLayer) {
      // Remove all tile layers
      mapInstanceRef.current.eachLayer((layer) => {
        if (layer instanceof L.TileLayer) {
          mapInstanceRef.current!.removeLayer(layer);
        }
      });

      // Add new layer
      const newLayer = availableLayers.find(layer => layer.id === layerId);
      if (newLayer) {
        L.tileLayer(newLayer.url, {
          attribution: newLayer.attribution,
          maxZoom: newLayer.maxZoom,
        }).addTo(mapInstanceRef.current);
        setActiveLayer(layerId);
      }
    }
  };

  // Advanced mapping functions
  
  // Route planning functions
  const startRouteCreation = useCallback(() => {
    setMapInteraction({
      mode: 'drawing',
      drawingMode: 'route',
      isActive: true
    });
    setIsCreatingRoute(true);
    setRouteWaypoints([]);
  }, []);

  const finishRouteCreation = useCallback(async (name: string, options: RouteOptions) => {
    if (routeWaypoints.length < 2) {
      console.warn('Route requires at least 2 waypoints');
      return;
    }

    try {
      const waypoints: Point[] = routeWaypoints.map(wp => ({ lat: wp.lat, lng: wp.lng }));
      const route = await mappingService.createRoute({
        name,
        waypoints,
        options
      });
      
      setRoutes(prev => new Map(prev).set(route.id, route));
      onRouteCreated?.(route);
      
      // Add route visualization to map
      addRouteToMap(route);
      
    } catch (error) {
      console.error('Failed to create route:', error);
    } finally {
      cancelDrawing();
    }
  }, [routeWaypoints, onRouteCreated]);

  const addRouteToMap = useCallback((route: Route) => {
    if (!mapInstanceRef.current) return;
    
    const coordinates: L.LatLngTuple[] = route.geometry.coordinates.map(
      coord => [coord[1], coord[0]] as L.LatLngTuple
    );
    
    const polyline = L.polyline(coordinates, {
      color: '#3b82f6',
      weight: 4,
      opacity: 0.8
    });
    
    polyline.bindPopup(`
      <div class="route-popup">
        <h3>${route.name}</h3>
        <div class="route-info">
          <div><strong>Distance:</strong> ${formatDistance(route.distance)}</div>
          <div><strong>Type:</strong> ${route.route_type}</div>
          <div><strong>Vehicle:</strong> ${route.vehicle}</div>
        </div>
      </div>
    `);
    
    polyline.addTo(mapInstanceRef.current);
    setRoutePolylines(prev => new Map(prev).set(route.id, polyline));
  }, []);

  // Geofence functions
  const startGeofenceCreation = useCallback((type: 'circle' | 'polygon' | 'rectangle') => {
    setMapInteraction({
      mode: 'drawing',
      drawingMode: `geofence-${type}` as DrawingMode,
      isActive: true
    });
    setDrawingPoints([]);
    setIsDrawing(true);
  }, []);

  const finishGeofenceCreation = useCallback(async (name: string, alertOnEnter: boolean, alertOnExit: boolean) => {
    if (drawingPoints.length === 0) return;
    
    let geometry: any;
    const type = mapInteraction.drawingMode.replace('geofence-', '') as 'circle' | 'polygon' | 'rectangle';
    
    if (type === 'circle' && drawingPoints.length >= 2) {
      const center = drawingPoints[0];
      const edge = drawingPoints[1];
      const radius = calculateDistance(latLngToPoint(center), latLngToPoint(edge));
      
      geometry = {
        center: { lat: center.lat, lng: center.lng },
        radius
      };
    } else if (type === 'polygon' && drawingPoints.length >= 3) {
      geometry = {
        type: 'Polygon',
        coordinates: [drawingPoints.map(p => [p.lng, p.lat])]
      };
    } else if (type === 'rectangle' && drawingPoints.length >= 2) {
      const bounds = L.latLngBounds(drawingPoints);
      geometry = {
        bounds: {
          north: bounds.getNorth(),
          south: bounds.getSouth(),
          east: bounds.getEast(),
          west: bounds.getWest()
        }
      };
    }
    
    if (!geometry) {
      console.warn('Invalid geometry for geofence');
      return;
    }
    
    try {
      const geofence = await mappingService.createGeofence({
        name,
        type,
        geometry,
        alert_on_enter: alertOnEnter,
        alert_on_exit: alertOnExit,
        enabled: true
      });
      
      setGeofences(prev => new Map(prev).set(geofence.id, geofence));
      onGeofenceCreated?.(geofence);
      
      // Add geofence visualization to map
      addGeofenceToMap(geofence);
      
    } catch (error) {
      console.error('Failed to create geofence:', error);
    } finally {
      cancelDrawing();
    }
  }, [drawingPoints, mapInteraction.drawingMode, onGeofenceCreated]);

  const addGeofenceToMap = useCallback((geofence: Geofence) => {
    if (!mapInstanceRef.current) return;
    
    let shape: L.Layer;
    
    if (geofence.type === 'circle') {
      const { center, radius } = geofence.geometry;
      shape = L.circle([center.lat, center.lng], {
        radius,
        color: '#ef4444',
        fillColor: '#ef4444',
        weight: 2,
        opacity: 0.8,
        fillOpacity: 0.1
      });
    } else if (geofence.type === 'polygon') {
      const coordinates = geofence.geometry.coordinates[0].map(
        (coord: number[]) => [coord[1], coord[0]] as L.LatLngTuple
      );
      shape = L.polygon(coordinates, {
        color: '#ef4444',
        fillColor: '#ef4444',
        weight: 2,
        opacity: 0.8,
        fillOpacity: 0.1
      });
    } else if (geofence.type === 'rectangle') {
      const { bounds } = geofence.geometry;
      shape = L.rectangle(
        [[bounds.south, bounds.west], [bounds.north, bounds.east]],
        {
          color: '#ef4444',
          fillColor: '#ef4444',
          weight: 2,
          opacity: 0.8,
          fillOpacity: 0.1
        }
      );
    } else {
      return;
    }
    
    shape.bindPopup(`
      <div class="geofence-popup">
        <h3>${geofence.name}</h3>
        <div class="geofence-info">
          <div><strong>Type:</strong> ${geofence.type}</div>
          <div><strong>Alerts:</strong> ${geofence.alert_on_enter ? 'Enter ' : ''}${geofence.alert_on_exit ? 'Exit' : ''}</div>
          <div><strong>Status:</strong> ${geofence.enabled ? 'Enabled' : 'Disabled'}</div>
        </div>
      </div>
    `);
    
    shape.addTo(mapInstanceRef.current);
    setGeofenceShapes(prev => new Map(prev).set(geofence.id, shape));
  }, []);

  // Measurement functions
  const startMeasurement = useCallback((type: 'distance' | 'area' | 'bearing') => {
    setMapInteraction({
      mode: 'measuring',
      drawingMode: `measure-${type}` as DrawingMode,
      isActive: true
    });
    setDrawingPoints([]);
    setIsDrawing(true);
    
    // Create measurement layer
    if (mapInstanceRef.current && !measurementLayer) {
      const layer = L.layerGroup().addTo(mapInstanceRef.current);
      setMeasurementLayer(layer);
    }
  }, [measurementLayer]);

  const addMeasurementPoint = useCallback((point: LatLng) => {
    if (!mapInteraction.isActive || !mapInstanceRef.current || !measurementLayer) return;
    
    const newPoints = [...drawingPoints, point];
    setDrawingPoints(newPoints);
    
    // Add marker for the point
    const marker = L.circleMarker([point.lat, point.lng], {
      color: '#3b82f6',
      fillColor: '#3b82f6',
      fillOpacity: 0.8,
      radius: 6
    }).addTo(measurementLayer);
    
    setTempMarkers(prev => [...prev, marker]);
    
    // Handle different measurement types
    const measureType = mapInteraction.drawingMode.replace('measure-', '');
    
    if (measureType === 'distance' && newPoints.length >= 2) {
      // Draw line and show distance
      const line = L.polyline(newPoints.map(p => [p.lat, p.lng]), {
        color: '#3b82f6',
        weight: 2,
        dashArray: '5, 10'
      }).addTo(measurementLayer);
      
      let totalDistance = 0;
      for (let i = 1; i < newPoints.length; i++) {
        totalDistance += calculateDistance(
          latLngToPoint(newPoints[i - 1]),
          latLngToPoint(newPoints[i])
        );
      }
      
      const tooltip = createMeasurementTooltip('distance', totalDistance);
      line.bindTooltip(tooltip, { permanent: true, className: 'measurement-tooltip' });
      
      onMeasurement?.('distance', totalDistance, newPoints);
    }
    
    if (measureType === 'area' && newPoints.length >= 3) {
      // Draw polygon and show area
      const polygon = L.polygon(newPoints.map(p => [p.lat, p.lng]), {
        color: '#22c55e',
        fillColor: '#22c55e',
        weight: 2,
        fillOpacity: 0.2
      }).addTo(measurementLayer);
      
      const area = calculatePolygonArea(newPoints.map(latLngToPoint));
      const tooltip = createMeasurementTooltip('area', area);
      polygon.bindTooltip(tooltip, { permanent: true, className: 'measurement-tooltip' });
      
      onMeasurement?.('area', area, newPoints);
    }
    
    if (measureType === 'bearing' && newPoints.length === 2) {
      // Draw line and show bearing
      const line = L.polyline(newPoints.map(p => [p.lat, p.lng]), {
        color: '#f59e0b',
        weight: 3,
        dashArray: '10, 5'
      }).addTo(measurementLayer);
      
      const bearing = calculateBearing(
        latLngToPoint(newPoints[0]),
        latLngToPoint(newPoints[1])
      );
      const distance = calculateDistance(
        latLngToPoint(newPoints[0]),
        latLngToPoint(newPoints[1])
      );
      
      const tooltip = createMeasurementTooltip('bearing', bearing, `Distance: ${formatDistance(distance)}`);
      line.bindTooltip(tooltip, { permanent: true, className: 'measurement-tooltip' });
      
      onMeasurement?.('bearing', bearing, newPoints);
      
      // Auto-finish bearing measurement after 2 points
      setTimeout(() => cancelDrawing(), 100);
    }
  }, [drawingPoints, mapInteraction, measurementLayer, onMeasurement]);

  // Common drawing functions
  const cancelDrawing = useCallback(() => {
    setMapInteraction({
      mode: 'view',
      drawingMode: 'none',
      isActive: false
    });
    setIsCreatingRoute(false);
    setIsDrawing(false);
    setDrawingPoints([]);
    setRouteWaypoints([]);
    
    // Clear temporary markers
    tempMarkers.forEach(marker => {
      if (mapInstanceRef.current?.hasLayer(marker)) {
        mapInstanceRef.current.removeLayer(marker);
      }
    });
    setTempMarkers([]);
  }, [tempMarkers]);

  const clearMeasurements = useCallback(() => {
    if (measurementLayer && mapInstanceRef.current) {
      measurementLayer.clearLayers();
    }
  }, [measurementLayer]);

  // Load routes and geofences on mount
  useEffect(() => {
    const loadMappingData = async () => {
      try {
        if (enableRouting && showRoutes) {
          const routesData = await mappingService.listRoutes({ limit: 100 });
          const routesMap = new Map<string, Route>();
          routesData.routes.forEach(route => {
            routesMap.set(route.id, route);
            if (mapInstanceRef.current) {
              addRouteToMap(route);
            }
          });
          setRoutes(routesMap);
        }
        
        if (enableGeofencing && showGeofences) {
          const geofencesData = await mappingService.listGeofences({ limit: 100 });
          const geofencesMap = new Map<string, Geofence>();
          geofencesData.geofences.forEach(geofence => {
            geofencesMap.set(geofence.id, geofence);
            if (mapInstanceRef.current) {
              addGeofenceToMap(geofence);
            }
          });
          setGeofences(geofencesMap);
        }
      } catch (error) {
        console.error('Failed to load mapping data:', error);
      }
    };
    
    if (mapInstanceRef.current) {
      loadMappingData();
    }
  }, [enableRouting, enableGeofencing, showRoutes, showGeofences, addRouteToMap, addGeofenceToMap]);

  // Public API - expose functions for external control
  const mapApi = {
    // Route functions
    startRouteCreation,
    finishRouteCreation,
    cancelDrawing,
    
    // Geofence functions
    startGeofenceCreation,
    finishGeofenceCreation,
    
    // Measurement functions
    startMeasurement,
    clearMeasurements,
    
    // Data access
    getRoutes: () => routes,
    getGeofences: () => geofences,
    getMapInstance: () => mapInstanceRef.current,
    
    // State
    mapInteraction,
    isDrawing,
    drawingPoints,
    routeWaypoints
  };
  
  // Expose API through a ref callback if parent needs it
  useEffect(() => {
    if (typeof onMapMove === 'function') {
      // We can extend this pattern to pass the API to parent
      (onMapMove as any).mapApi = mapApi;
    }
  }, [mapApi, onMapMove]);

  // Get map instance for external use (for future features)
  // const getMapInstance = () => mapInstanceRef.current;

  return (
    <div className={`tactical-map-container ${className}`} style={{ height, width }}>
      {/* Layer Controls */}
      <div className="map-layer-controls">
        {availableLayers.map((layer) => (
          <button
            key={layer.id}
            className={`layer-btn ${layer.id === activeLayer ? 'active' : ''}`}
            onClick={() => switchLayer(layer.id)}
            title={layer.name}
            disabled={readOnly}
          >
            {layer.name}
          </button>
        ))}
      </div>

      {/* Status Display */}
      {showEntities && (
        <div className="map-status-display">
          <div className={`connection-status ${wsConnected ? 'connected' : 'disconnected'}`}>
            {wsConnected ? '🟢 Connected' : '🔴 Disconnected'}
          </div>
          <div className="entity-count">
            Entities: {entities.size}
          </div>
          {isLoadingEntities && (
            <div className="loading-status">
              Loading entities...
            </div>
          )}
          {wsError && (
            <div className="error-status">
              ⚠️ {wsError}
            </div>
          )}
        </div>
      )}

      {/* Advanced Mapping Controls */}
      {(enableRouting || enableGeofencing || enableMeasurement) && !readOnly && (
        <div className="map-advanced-controls">
          {enableRouting && (
            <div className="routing-controls">
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'route' ? 'active' : ''}`}
                onClick={mapInteraction.drawingMode === 'route' ? cancelDrawing : startRouteCreation}
                title={mapInteraction.drawingMode === 'route' ? 'Cancel Route Creation' : 'Create Route'}
              >
                🛣️ Route
              </button>
              {routeWaypoints.length > 0 && (
                <div className="waypoint-count">
                  {routeWaypoints.length} waypoint{routeWaypoints.length !== 1 ? 's' : ''}
                </div>
              )}
            </div>
          )}
          
          {enableGeofencing && (
            <div className="geofencing-controls">
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'geofence-circle' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'geofence-circle' ? cancelDrawing() : startGeofenceCreation('circle')}
                title="Create Circular Geofence"
              >
                ⭕ Circle
              </button>
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'geofence-polygon' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'geofence-polygon' ? cancelDrawing() : startGeofenceCreation('polygon')}
                title="Create Polygon Geofence"
              >
                🔷 Polygon
              </button>
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'geofence-rectangle' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'geofence-rectangle' ? cancelDrawing() : startGeofenceCreation('rectangle')}
                title="Create Rectangle Geofence"
              >
                ◼️ Rectangle
              </button>
            </div>
          )}
          
          {enableMeasurement && (
            <div className="measurement-controls">
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'measure-distance' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'measure-distance' ? cancelDrawing() : startMeasurement('distance')}
                title="Measure Distance"
              >
                📏 Distance
              </button>
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'measure-area' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'measure-area' ? cancelDrawing() : startMeasurement('area')}
                title="Measure Area"
              >
                📐 Area
              </button>
              <button
                className={`control-btn ${mapInteraction.drawingMode === 'measure-bearing' ? 'active' : ''}`}
                onClick={() => mapInteraction.drawingMode === 'measure-bearing' ? cancelDrawing() : startMeasurement('bearing')}
                title="Measure Bearing"
              >
                🧭 Bearing
              </button>
              <button
                className="control-btn clear-btn"
                onClick={clearMeasurements}
                title="Clear Measurements"
              >
                🗑️ Clear
              </button>
            </div>
          )}
        </div>
      )}
      
      {/* Route Creation Completion */}
      {mapInteraction.drawingMode === 'route' && routeWaypoints.length >= 2 && (
        <div className="map-completion-dialog route-completion">
          <div className="completion-content">
            <h3>Create Route</h3>
            <div className="completion-details">
              <div>Waypoints: {routeWaypoints.length}</div>
              {routeWaypoints.length > 0 && (
                <div>Distance: {formatDistance(
                  routeWaypoints.reduce((total, point, index) => {
                    if (index === 0) return 0;
                    const prev = routeWaypoints[index - 1];
                    return total + calculateDistance(latLngToPoint(prev), latLngToPoint(point));
                  }, 0)
                )}</div>
              )}
            </div>
            <div className="completion-inputs">
              <input
                type="text"
                placeholder="Route name (optional)"
                maxLength={100}
                onChange={(e) => {
                  setMapInteraction(prev => ({
                    ...prev,
                    tempName: e.target.value
                  }));
                }}
              />
              <textarea
                placeholder="Route description (optional)"
                maxLength={500}
                rows={3}
                onChange={(e) => {
                  setMapInteraction(prev => ({
                    ...prev,
                    tempDescription: e.target.value
                  }));
                }}
              />
            </div>
            <div className="completion-actions">
              <button className="btn-primary" onClick={finishRouteCreation}>
                Create Route
              </button>
              <button className="btn-secondary" onClick={cancelDrawing}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Geofence Creation Completion */}
      {mapInteraction.drawingMode?.startsWith('geofence-') && (
        (mapInteraction.drawingMode === 'geofence-circle' && drawingPoints.length >= 2) ||
        (mapInteraction.drawingMode === 'geofence-polygon' && drawingPoints.length >= 3) ||
        (mapInteraction.drawingMode === 'geofence-rectangle' && drawingPoints.length >= 2)
      ) && (
        <div className="map-completion-dialog geofence-completion">
          <div className="completion-content">
            <h3>Create Geofence</h3>
            <div className="completion-details">
              <div>Type: {mapInteraction.drawingMode.replace('geofence-', '').charAt(0).toUpperCase() + mapInteraction.drawingMode.replace('geofence-', '').slice(1)}</div>
              <div>Points: {drawingPoints.length}</div>
              {mapInteraction.drawingMode === 'geofence-circle' && drawingPoints.length >= 2 && (
                <div>Radius: {formatDistance(calculateDistance(latLngToPoint(drawingPoints[0]), latLngToPoint(drawingPoints[1])))}</div>
              )}
              {mapInteraction.drawingMode === 'geofence-polygon' && drawingPoints.length >= 3 && (
                <div>Area: {formatArea(calculatePolygonArea(drawingPoints.map(latLngToPoint)))}</div>
              )}
              {mapInteraction.drawingMode === 'geofence-rectangle' && drawingPoints.length >= 2 && (
                <div>Area: {formatArea(calculateRectangleArea(latLngToPoint(drawingPoints[0]), latLngToPoint(drawingPoints[1])))}</div>
              )}
            </div>
            <div className="completion-inputs">
              <input
                type="text"
                placeholder="Geofence name (optional)"
                maxLength={100}
                onChange={(e) => {
                  setMapInteraction(prev => ({
                    ...prev,
                    tempName: e.target.value
                  }));
                }}
              />
              <textarea
                placeholder="Geofence description (optional)"
                maxLength={500}
                rows={3}
                onChange={(e) => {
                  setMapInteraction(prev => ({
                    ...prev,
                    tempDescription: e.target.value
                  }));
                }}
              />
            </div>
            <div className="completion-actions">
              <button className="btn-primary" onClick={finishGeofenceCreation}>
                Create Geofence
              </button>
              <button className="btn-secondary" onClick={cancelDrawing}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Drawing Status */}
      {mapInteraction.isActive && (
        <div className="map-drawing-status">
          <div className="drawing-instructions">
            {mapInteraction.drawingMode === 'route' && (
              <span>Click on the map to add waypoints. {routeWaypoints.length >= 2 ? 'Ready to create route.' : `Need ${2 - routeWaypoints.length} more waypoint${2 - routeWaypoints.length > 1 ? 's' : ''}.`}</span>
            )}
            {mapInteraction.drawingMode === 'geofence-circle' && (
              <span>Click center, then click edge to define radius. {drawingPoints.length >= 2 ? 'Ready to create geofence.' : `Need ${2 - drawingPoints.length} more point${2 - drawingPoints.length > 1 ? 's' : ''}.`}</span>
            )}
            {mapInteraction.drawingMode === 'geofence-polygon' && (
              <span>Click to add polygon points. {drawingPoints.length >= 3 ? 'Ready to create geofence.' : `Need ${3 - drawingPoints.length} more point${3 - drawingPoints.length > 1 ? 's' : ''}.`}</span>
            )}
            {mapInteraction.drawingMode === 'geofence-rectangle' && (
              <span>Click two opposite corners. {drawingPoints.length >= 2 ? 'Ready to create geofence.' : `Need ${2 - drawingPoints.length} more point${2 - drawingPoints.length > 1 ? 's' : ''}.`}</span>
            )}
            {mapInteraction.drawingMode.startsWith('measure-') && (
              <span>Click on the map to add measurement points.</span>
            )}
          </div>
          <button className="cancel-drawing-btn" onClick={cancelDrawing}>
            Cancel
          </button>
        </div>
      )}

      {/* Coordinate Display */}
      {showCoordinates && (
        <div className="map-coordinate-display">
          <div className="current-center">
            Center: {formatCoordinates(currentCenter.lat, currentCenter.lng, { format: coordinateFormat })}
          </div>
          {mousePosition && (
            <div className="mouse-position">
              Cursor: {formatCoordinates(mousePosition.lat, mousePosition.lng, { format: coordinateFormat })}
            </div>
          )}
          <div className="zoom-level">
            Zoom: {currentZoom}
          </div>
          {/* Mapping stats */}
          {(showRoutes || showGeofences) && (
            <div className="mapping-stats">
              {showRoutes && routes.size > 0 && <div>Routes: {routes.size}</div>}
              {showGeofences && geofences.size > 0 && <div>Geofences: {geofences.size}</div>}
            </div>
          )}
        </div>
      )}

      {/* Map Container */}
      <div
        ref={mapRef}
        className="tactical-map"
        style={{ height: '100%', width: '100%' }}
      />
    </div>
  );
}
