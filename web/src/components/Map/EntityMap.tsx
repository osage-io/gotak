/**
 * EntityMap Component
 * Enhanced Leaflet-based tactical map with real-time entity tracking and area map support
 */

import React, { useRef, useEffect, useState, useCallback, forwardRef, useImperativeHandle } from 'react';
import { MapContainer, TileLayer, Marker, Popup, useMap, useMapEvents } from 'react-leaflet';
import MarkerClusterGroup from 'react-leaflet-markercluster';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import 'leaflet.markercluster/dist/MarkerCluster.css';
import 'leaflet.markercluster/dist/MarkerCluster.Default.css';
import { Entity } from '../../services/apiClient';
import { useFilteredEntities, useSelectedEntity } from '../../hooks/useEntityTracker';
import { wsService, ConnectionState } from '../../services/websocketService';

// Map configuration
const MAP_CONFIG = {
  defaultCenter: [38.8951, -77.0364] as [number, number], // Washington, DC
  defaultZoom: 10,
  maxZoom: 18,
  minZoom: 3,
};

// Available map layers
export interface MapLayerConfig {
  id: string;
  name: string;
  url: string;
  attribution: string;
  maxZoom?: number;
}

export const MAP_LAYERS: { [key: string]: MapLayerConfig } = {
  // Dark mode layers for tactical operations (listed first for priority)
  darkTactical: {
    id: 'darkTactical',
    name: 'Dark Tactical',
    url: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',  // CartoDB: keyless (Stadia 401s without an API key/domain allowlist)
    attribution: '&copy; <a href="https://carto.com/attributions">CARTO</a> &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors',
    maxZoom: 20,
  },
  darkSatellite: {
    id: 'darkSatellite',
    name: 'Night Satellite',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Base/MapServer/tile/{z}/{y}/{x}',
    attribution: '&copy; <a href="https://www.esri.com/">Esri</a> &mdash; Esri, DeLorme, NAVTEQ',
    maxZoom: 16,
  },
  darkStreets: {
    id: 'darkStreets',
    name: 'Dark Streets',
    url: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    maxZoom: 19,
  },
  darkMatter: {
    id: 'darkMatter',
    name: 'Dark Matter',
    url: 'https://{s}.basemaps.cartocdn.com/dark_nolabels/{z}/{x}/{y}{r}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    maxZoom: 19,
  },
  // Light mode layers (moved down)
  osm: {
    id: 'osm',
    name: 'OpenStreetMap',
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    maxZoom: 19,
  },
  satellite: {
    id: 'satellite',
    name: 'Satellite',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: '&copy; <a href="https://www.esri.com/">Esri</a> &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community',
    maxZoom: 19,
  },
  terrain: {
    id: 'terrain',
    name: 'Terrain',
    url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
    attribution: 'Map data: &copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors, <a href="http://viewfinderpanoramas.org">SRTM</a> | Map style: &copy; <a href="https://opentopomap.org">OpenTopoMap</a> (<a href="https://creativecommons.org/licenses/by-sa/3.0/">CC-BY-SA</a>)',
    maxZoom: 17,
  },
  hybrid: {
    id: 'hybrid',
    name: 'Hybrid',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: '&copy; <a href="https://www.esri.com/">Esri</a>',
    maxZoom: 19,
  },
};

// Entity colors based on affiliation
const ENTITY_COLORS = {
  friendly: '#0066cc',   // Blue
  hostile: '#cc0000',    // Red
  unknown: '#ffaa00',    // Orange
  neutral: '#00cc66',    // Green
  default: '#666666',    // Gray
};

// Fix Leaflet default marker issue
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

// Props interface
interface EntityMapProps {
  className?: string;
  style?: React.CSSProperties;
  onEntityClick?: (entity: Entity) => void;
  showEntityLabels?: boolean;
  showEntityTrails?: boolean;
  enableClustering?: boolean;
  currentLayer?: string;
  showAreaMaps?: boolean;
  areaMaps?: AreaMapOverlay[];
  children?: React.ReactNode;
}

// Area map overlay interface
export interface AreaMapOverlay {
  id: string;
  name: string;
  type: 'geojson' | 'kml' | 'wms' | 'image';
  data?: any;
  url?: string;
  bounds?: [[number, number], [number, number]];
  opacity?: number;
  style?: any;
  visible?: boolean;
}

// Map reference interface for external control
export interface EntityMapRef {
  panTo: (lat: number, lng: number) => void;
  setZoom: (zoom: number) => void;
  getBounds: () => { north: number; south: number; east: number; west: number } | null;
  fitBounds: (entities: Entity[]) => void;
  getMap: () => L.Map | null;
  addAreaMap: (areaMap: AreaMapOverlay) => void;
  removeAreaMap: (id: string) => void;
  toggleAreaMap: (id: string) => void;
}

// Create custom tactical icons for different entity types
const createTacticalIcon = (entity: Entity, isSelected: boolean = false): L.DivIcon => {
  let color = ENTITY_COLORS.default;
  let symbol = '●';
  
  // Determine color based on affiliation
  if (entity.entityType.startsWith('a-f')) {
    color = ENTITY_COLORS.friendly;
    symbol = '▲'; // NATO friendly symbol
  } else if (entity.entityType.startsWith('a-h')) {
    color = ENTITY_COLORS.hostile;
    symbol = '◆'; // NATO hostile symbol
  } else if (entity.entityType.startsWith('a-u')) {
    color = ENTITY_COLORS.unknown;
    symbol = '▼'; // NATO unknown symbol
  } else if (entity.entityType.startsWith('a-n')) {
    color = ENTITY_COLORS.neutral;
    symbol = '■'; // NATO neutral symbol
  }
  
  const size = isSelected ? 24 : 16;
  const borderWidth = isSelected ? 3 : 1;
  
  return L.divIcon({
    html: `
      <div style="
        width: ${size}px;
        height: ${size}px;
        background-color: ${color};
        border: ${borderWidth}px solid ${isSelected ? '#ffffff' : '#000000'};
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-weight: bold;
        font-size: ${size * 0.6}px;
        box-shadow: ${isSelected ? '0 0 10px rgba(255, 255, 255, 0.8)' : '0 1px 3px rgba(0,0,0,0.3)'};
        transition: all 0.3s ease;
      ">${symbol}</div>
      ${entity.callsign ? `
        <div style="
          position: absolute;
          top: ${size + 2}px;
          left: 50%;
          transform: translateX(-50%);
          background-color: rgba(0, 0, 0, 0.8);
          color: white;
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 10px;
          white-space: nowrap;
          pointer-events: none;
        ">${entity.callsign}</div>
      ` : ''}
    `,
    className: 'tactical-marker',
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  });
};

// Entity marker component using Leaflet Marker
const EntityMarker: React.FC<{
  entity: Entity;
  isSelected: boolean;
  showLabel: boolean;
  onClick: (entity: Entity) => void;
}> = ({ entity, isSelected, showLabel, onClick }) => {
  const [isUpdating, setIsUpdating] = useState(false);
  
  // Animation for real-time updates
  useEffect(() => {
    const lastUpdate = new Date(entity.lastUpdate).getTime();
    const now = Date.now();
    if (now - lastUpdate < 2000) { // Recent update within 2 seconds
      setIsUpdating(true);
      const timer = setTimeout(() => setIsUpdating(false), 1000);
      return () => clearTimeout(timer);
    }
  }, [entity.lastUpdate]);
  
  if (!entity.position) return null;
  
  const position: [number, number] = [entity.position.lat, entity.position.lng];
  const icon = createTacticalIcon(entity, isSelected);
  
  return (
    <Marker
      position={position}
      icon={icon}
      eventHandlers={{
        click: () => onClick(entity),
      }}
    >
      <Popup>
        <div style={{ minWidth: '200px' }}>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px', fontWeight: 'bold' }}>
            {entity.callsign || entity.uid}
          </h4>
          <div><strong>Type:</strong> {entity.entityType}</div>
          <div><strong>Position:</strong> {entity.position.lat.toFixed(6)}, {entity.position.lng.toFixed(6)}</div>
          <div><strong>Last Update:</strong> {new Date(entity.lastUpdate).toLocaleTimeString()}</div>
          {entity.speed && (
            <div><strong>Speed:</strong> {entity.speed.toFixed(1)} m/s</div>
          )}
          {entity.course && (
            <div><strong>Course:</strong> {entity.course.toFixed(0)}°</div>
          )}
        </div>
      </Popup>
    </Marker>
  );
};

// Map event handler component
const MapEventHandler: React.FC<{
  onMapReady?: (map: L.Map) => void;
  onViewChange?: (center: [number, number], zoom: number) => void;
}> = ({ onMapReady, onViewChange }) => {
  const map = useMap();
  
  useEffect(() => {
    if (map && onMapReady) {
      onMapReady(map);
    }
  }, [map, onMapReady]);
  
  useMapEvents({
    moveend: () => {
      if (onViewChange) {
        const center = map.getCenter();
        const zoom = map.getZoom();
        onViewChange([center.lat, center.lng], zoom);
      }
    },
    zoomend: () => {
      if (onViewChange) {
        const center = map.getCenter();
        const zoom = map.getZoom();
        onViewChange([center.lat, center.lng], zoom);
      }
    },
  });
  
  return null;
};

// Import the AreaMapOverlay component
import AreaMapOverlayComponent from './AreaMapOverlay';

// Area map overlays component
const AreaMapOverlays: React.FC<{
  areaMaps: AreaMapOverlay[];
}> = ({ areaMaps }) => {
  return (
    <>
      {areaMaps.map((areaMap) => (
        <AreaMapOverlayComponent
          key={areaMap.id}
          areaMap={areaMap}
        />
      ))}
    </>
  );
};

// Main EntityMap component
const EntityMap = forwardRef<EntityMapRef, EntityMapProps>((
  {
    className,
    style,
    onEntityClick,
    showEntityLabels = true,
    showEntityTrails = false,
    enableClustering = true,
    currentLayer = 'osm',
    showAreaMaps = false,
    areaMaps = [],
    children,
  },
  ref
) => {
  const [mapInstance, setMapInstance] = useState<L.Map | null>(null);
  const [mapCenter, setMapCenter] = useState<[number, number]>(MAP_CONFIG.defaultCenter);
  const [mapZoom, setMapZoom] = useState(MAP_CONFIG.defaultZoom);
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    wsService.connectionState
  );
  
  // Entity tracking hooks
  const { entities, counts } = useFilteredEntities();
  const { selectedEntity, selectEntity } = useSelectedEntity();
  
  // Monitor WebSocket connection state
  useEffect(() => {
    const handleConnectionChange = () => {
      setConnectionState(wsService.connectionState);
    };

    const unsubscribeConnection = wsService.onConnection(handleConnectionChange);
    const unsubscribeDisconnection = wsService.onDisconnection(handleConnectionChange);
    const unsubscribeError = wsService.onError(handleConnectionChange);

    return () => {
      unsubscribeConnection();
      unsubscribeDisconnection();
      unsubscribeError();
    };
  }, []);

  // Handle entity selection
  const handleEntityClick = useCallback((entity: Entity) => {
    selectEntity(entity.id);
    onEntityClick?.(entity);
  }, [selectEntity, onEntityClick]);

  // Pan to entity
  const panToEntity = useCallback((entity: Entity) => {
    if (entity.position && mapInstance) {
      mapInstance.panTo([entity.position.lat, entity.position.lng]);
    }
  }, [mapInstance]);

  // Fit bounds to show all entities
  const fitToEntities = useCallback((entitiesToFit: Entity[] = entities) => {
    if (entitiesToFit.length === 0 || !mapInstance) return;
    
    const positions = entitiesToFit
      .filter(e => e.position)
      .map(e => [e.position!.lat, e.position!.lng] as [number, number]);
    
    if (positions.length === 0) return;
    
    if (positions.length === 1) {
      mapInstance.setView(positions[0], 15);
    } else {
      const bounds = L.latLngBounds(positions);
      mapInstance.fitBounds(bounds, { padding: [20, 20] });
    }
  }, [entities, mapInstance]);

  // Handle map ready
  const handleMapReady = useCallback((map: L.Map) => {
    setMapInstance(map);
  }, []);

  // Handle view changes
  const handleViewChange = useCallback((center: [number, number], zoom: number) => {
    setMapCenter(center);
    setMapZoom(zoom);
  }, []);

  // Expose ref methods
  useImperativeHandle(ref, () => ({
    panTo: (lat: number, lng: number) => {
      if (mapInstance) {
        mapInstance.panTo([lat, lng]);
      }
    },
    setZoom: (zoom: number) => {
      if (mapInstance) {
        mapInstance.setZoom(zoom);
      }
    },
    getBounds: () => {
      if (mapInstance) {
        const bounds = mapInstance.getBounds();
        return {
          north: bounds.getNorth(),
          south: bounds.getSouth(),
          east: bounds.getEast(),
          west: bounds.getWest(),
        };
      }
      return null;
    },
    fitBounds: fitToEntities,
    getMap: () => mapInstance,
    addAreaMap: (areaMap: AreaMapOverlay) => {
      // TODO: Implement area map addition
      console.log('Adding area map:', areaMap);
    },
    removeAreaMap: (id: string) => {
      // TODO: Implement area map removal
      console.log('Removing area map:', id);
    },
    toggleAreaMap: (id: string) => {
      // TODO: Implement area map toggle
      console.log('Toggling area map:', id);
    },
  }), [fitToEntities, mapInstance]);

  // Auto-pan to selected entity
  useEffect(() => {
    if (selectedEntity) {
      panToEntity(selectedEntity);
    }
  }, [selectedEntity, panToEntity]);

  // Get current layer configuration
  const layerConfig = MAP_LAYERS[currentLayer] || MAP_LAYERS.osm;

  // Connection status style
  const getStatusStyle = () => {
    const baseStyle = {
      padding: '4px 8px',
      borderRadius: '4px',
      fontSize: '10px',
      fontWeight: 'bold' as const,
      textTransform: 'uppercase' as const,
    };

    switch (connectionState) {
      case ConnectionState.CONNECTED:
        return { ...baseStyle, backgroundColor: '#00cc66', color: 'white' };
      case ConnectionState.CONNECTING:
      case ConnectionState.RECONNECTING:
        return { ...baseStyle, backgroundColor: '#ffaa00', color: 'white' };
      case ConnectionState.ERROR:
      case ConnectionState.DISCONNECTED:
        return { ...baseStyle, backgroundColor: '#cc0000', color: 'white' };
      default:
        return { ...baseStyle, backgroundColor: '#666666', color: 'white' };
    }
  };

  // Render entity markers
  const renderEntityMarkers = () => {
    const entitiesWithPosition = entities.filter(entity => entity.position);
    
    if (enableClustering) {
      return (
        <MarkerClusterGroup>
          {entitiesWithPosition.map(entity => (
            <EntityMarker
              key={entity.id}
              entity={entity}
              isSelected={selectedEntity?.id === entity.id}
              showLabel={showEntityLabels}
              onClick={handleEntityClick}
            />
          ))}
        </MarkerClusterGroup>
      );
    }
    
    return entitiesWithPosition.map(entity => (
      <EntityMarker
        key={entity.id}
        entity={entity}
        isSelected={selectedEntity?.id === entity.id}
        showLabel={showEntityLabels}
        onClick={handleEntityClick}
      />
    ));
  };

  return (
    <div
      className={`entity-map ${className || ''}`}
      style={{
        width: '100%',
        height: '100%',
        position: 'relative',
        ...style,
      }}
    >
      <MapContainer
        center={mapCenter}
        zoom={mapZoom}
        minZoom={MAP_CONFIG.minZoom}
        maxZoom={MAP_CONFIG.maxZoom}
        style={{ height: '100%', width: '100%' }}
        zoomControl={false}
      >
        {/* Base tile layer */}
        <TileLayer
          key={layerConfig.id}
          url={layerConfig.url}
          attribution={layerConfig.attribution}
          maxZoom={layerConfig.maxZoom}
        />
        
        {/* Area map overlays */}
        {showAreaMaps && <AreaMapOverlays areaMaps={areaMaps} />}
        
        {/* Entity markers */}
        {renderEntityMarkers()}
        
        {/* Map event handlers */}
        <MapEventHandler
          onMapReady={handleMapReady}
          onViewChange={handleViewChange}
        />
        
        {/* Child components (like DrawingTools) */}
        {React.Children.map(children, child => child)}
      </MapContainer>
      
      {/* Connection Status Overlay */}
      <div style={{
        position: 'absolute',
        top: '10px',
        right: '10px',
        zIndex: 1000,
        ...getStatusStyle(),
      }}>
        {connectionState}
      </div>
      
      {/* Entity Count Display */}
      <div style={{
        position: 'absolute',
        top: '10px',
        left: '10px',
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        color: 'white',
        padding: '8px',
        borderRadius: '4px',
        fontSize: '12px',
        zIndex: 1000,
      }}>
        <div>Total: {counts.total}</div>
        <div style={{ color: ENTITY_COLORS.friendly }}>Friendly: {counts.friendly}</div>
        <div style={{ color: ENTITY_COLORS.hostile }}>Hostile: {counts.hostile}</div>
        <div style={{ color: ENTITY_COLORS.unknown }}>Unknown: {counts.unknown}</div>
      </div>
      
      {/* Map coordinates display */}
      <div style={{
        position: 'absolute',
        bottom: '10px',
        left: '10px',
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        color: 'white',
        padding: '4px 8px',
        borderRadius: '4px',
        fontSize: '10px',
        fontFamily: 'monospace',
        zIndex: 1000,
      }}>
        {mapCenter[0].toFixed(6)}, {mapCenter[1].toFixed(6)} | Z{mapZoom}
      </div>
      
      {/* CSS for tactical markers */}
      <style>{`
        .tactical-marker {
          transition: all 0.3s ease !important;
        }
        .tactical-marker:hover {
          transform: scale(1.1) !important;
        }
        .leaflet-cluster-anim .leaflet-marker-icon, .leaflet-cluster-anim .leaflet-marker-shadow {
          transition: transform 0.3s ease, opacity 0.3s ease;
        }
      `}</style>
    </div>
  );
});

EntityMap.displayName = 'EntityMap';

export default EntityMap;
