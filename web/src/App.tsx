import { useState, useCallback } from 'react';
import { TacticalMap } from './components/maps/TacticalMap';
import { MapToolsPanel } from './components/maps/MapToolsPanel';
import type { LatLng } from './utils/coordinates';
import type { EntityPosition } from './types/position';
import './App.css';

function App() {
  const [selectedEntity, setSelectedEntity] = useState<EntityPosition | null>(null);
  const [mapCenter, setMapCenter] = useState<LatLng>({ lat: 38.9072, lng: -77.0369 }); // Washington DC
  const [mapZoom, setMapZoom] = useState<number>(12);
  const [showFriendlyOnly, setShowFriendlyOnly] = useState<boolean>(false);
  const [showHostileOnly, setShowHostileOnly] = useState<boolean>(false);
  const [autoCenter, setAutoCenter] = useState<boolean>(false);
  const [isMapToolsOpen, setIsMapToolsOpen] = useState<boolean>(false);

  const handleEntityClick = useCallback((entity: EntityPosition) => {
    console.log('Entity clicked:', entity);
    setSelectedEntity(entity);
  }, []);

  const handleMapClick = useCallback((latLng: LatLng) => {
    console.log('Map clicked at:', latLng);
    // Could be used for adding new entities, waypoints, etc.
  }, []);

  const handleMapMove = useCallback((center: LatLng, zoom: number) => {
    setMapCenter(center);
    setMapZoom(zoom);
  }, []);

  const handleMapToolsToggle = useCallback(() => {
    setIsMapToolsOpen(prev => !prev);
  }, []);

  const handleMapInteraction = useCallback((type: 'route' | 'geofence' | 'measurement', data: any) => {
    console.log(`Map interaction: ${type}`, data);
    // Here we would handle the actual map interactions with the specific tools
    // This would include creating/editing routes, geofences, and measurements
  }, []);

  return (
    <div className="app">
      <header className="app-header">
        <h1>GoTAK - Tactical Awareness Kit</h1>
        <div className="app-status">
          <span className="status-badge">OPERATIONAL</span>
          {selectedEntity && (
            <span className="selected-entity">
              Selected: {selectedEntity.callsign}
            </span>
          )}
        </div>
      </header>
      
      <main className="app-main">
        <div className="map-controls">
          <button
            className={`control-btn ${autoCenter ? 'active' : ''}`}
            onClick={() => setAutoCenter(!autoCenter)}
            title="Auto-center on entities"
          >
            🎯 Auto Center
          </button>
          <button
            className={`control-btn ${showFriendlyOnly ? 'active' : ''}`}
            onClick={() => {
              setShowFriendlyOnly(!showFriendlyOnly);
              if (!showFriendlyOnly) setShowHostileOnly(false);
            }}
            title="Show friendly entities only"
          >
            🟢 Friendly Only
          </button>
          <button
            className={`control-btn ${showHostileOnly ? 'active' : ''}`}
            onClick={() => {
              setShowHostileOnly(!showHostileOnly);
              if (!showHostileOnly) setShowFriendlyOnly(false);
            }}
            title="Show hostile entities only"
          >
            🔴 Hostile Only
          </button>
          <button
            className={`control-btn ${isMapToolsOpen ? 'active' : ''}`}
            onClick={handleMapToolsToggle}
            title="Map Tools"
          >
            🧰 Map Tools
          </button>
        </div>
        
        <TacticalMap
          className="main-map"
          initialCenter={mapCenter}
          initialZoom={mapZoom}
          height="100%"
          width="100%"
          showCoordinates={true}
          coordinateFormat="dd"
          showScale={true}
          showEntities={true}
          showTrails={false}
          showFriendlyOnly={showFriendlyOnly}
          showHostileOnly={showHostileOnly}
          autoCenter={autoCenter}
          onMapClick={handleMapClick}
          onMapMove={handleMapMove}
          onEntityClick={handleEntityClick}
        />
      </main>
      
      {/* Entity Details Panel (when entity selected) */}
      {selectedEntity && (
        <div className="entity-details-panel">
          <div className="panel-header">
            <h3>Entity Details</h3>
            <button
              className="close-button"
              onClick={() => setSelectedEntity(null)}
              aria-label="Close details panel"
            >
              ×
            </button>
          </div>
          <div className="panel-content">
            <div className="detail-row">
              <label>Callsign:</label>
              <span>{selectedEntity.callsign}</span>
            </div>
            <div className="detail-row">
              <label>Entity ID:</label>
              <span>{selectedEntity.entityId}</span>
            </div>
            <div className="detail-row">
              <label>Type:</label>
              <span className={`entity-type ${selectedEntity.isFriendly ? 'friendly' : selectedEntity.isHostile ? 'hostile' : 'unknown'}`}>
                {selectedEntity.isFriendly ? '🟢 Friendly' : selectedEntity.isHostile ? '🔴 Hostile' : '🟡 Unknown'}
              </span>
            </div>
            <div className="detail-row">
              <label>Group:</label>
              <span>{selectedEntity.group || 'N/A'}</span>
            </div>
            <div className="detail-row">
              <label>Position:</label>
              <span className="coordinates">
                {selectedEntity.lat.toFixed(6)}, {selectedEntity.lng.toFixed(6)}
              </span>
            </div>
            {selectedEntity.altitude && (
              <div className="detail-row">
                <label>Altitude:</label>
                <span>{selectedEntity.altitude.toFixed(0)}m</span>
              </div>
            )}
            {selectedEntity.speed && (
              <div className="detail-row">
                <label>Speed:</label>
                <span>{selectedEntity.speed.toFixed(1)} m/s</span>
              </div>
            )}
            {selectedEntity.course && (
              <div className="detail-row">
                <label>Course:</label>
                <span>{selectedEntity.course.toFixed(0)}°</span>
              </div>
            )}
            <div className="detail-row">
              <label>Last Update:</label>
              <span>{new Date(selectedEntity.lastUpdate).toLocaleString()}</span>
            </div>
            <div className="detail-row">
              <label>Status:</label>
              <span className={`status-badge ${selectedEntity.isStale ? 'stale' : 'active'}`}>
                {selectedEntity.isStale ? 'STALE' : 'ACTIVE'}
              </span>
            </div>
          </div>
        </div>
      )}
      
      {/* Map Tools Panel */}
      <MapToolsPanel 
        isOpen={isMapToolsOpen}
        onClose={handleMapToolsToggle}
        onMapInteraction={handleMapInteraction}
      />
    </div>
  );
}

export default App;
