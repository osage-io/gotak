/**
 * Enhanced Tactical Map Page
 * Comprehensive tactical mapping interface with entity visualization and real-time updates
 */

import React, { useState, useRef, useCallback, useEffect } from 'react';
import EntityMap, { EntityMapRef, MAP_LAYERS, AreaMapOverlay } from '../components/Map/EntityMap';
import DrawingTools, { DrawingTool, DrawingAnnotation } from '../components/Map/DrawingTools';
import { useFilteredEntities, useSelectedEntity, useBulkEntityOperations } from '../hooks/useEntityTracker';
import { EntityFilterType } from '../stores/entityTracker';
import { wsService, ConnectionState } from '../services/websocketService';
import { Entity } from '../services/apiClient';
import { formatCoordinates, CoordinateDisplayOptions } from '../utils/coordinates';

// Map control types
type MapLayer = keyof typeof MAP_LAYERS;
type MapTool = 'select' | 'measure' | 'draw' | 'route';

interface MapViewport {
  center: { lat: number; lng: number };
  zoom: number;
}

const TacticalMap: React.FC = () => {
  // Map state
  const [mapLayer, setMapLayer] = useState<MapLayer>('osm');
  const [activeTool, setActiveTool] = useState<MapTool>('select');
  const [drawingTool, setDrawingTool] = useState<DrawingTool>('select');
  const [showLabels, setShowLabels] = useState(true);
  const [showTrails, setShowTrails] = useState(false);
  const [enableClustering, setEnableClustering] = useState(true);
  const [showAreaMaps, setShowAreaMaps] = useState(false);
  const [coordinateFormat, setCoordinateFormat] = useState<CoordinateDisplayOptions['format']>('dd');
  const [viewport, setViewport] = useState<MapViewport>({
    center: { lat: 38.8951, lng: -77.0364 }, // Washington, DC
    zoom: 10
  });
  
  // Area maps and annotations state
  const [areaMaps, setAreaMaps] = useState<AreaMapOverlay[]>([]);
  const [annotations, setAnnotations] = useState<DrawingAnnotation[]>([]);
  
  // UI state
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [fullscreen, setFullscreen] = useState(false);
  const [measurementMode, setMeasurementMode] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // Connection state
  const [connectionStatus, setConnectionStatus] = useState<ConnectionState>(
    wsService.connectionState
  );

  // References
  const mapRef = useRef<EntityMapRef>(null);

  // Entity hooks
  const { entities, counts, filter, setFilter } = useFilteredEntities();
  const { selectedEntity, selectEntity, clearSelection } = useSelectedEntity();
  const { loadInitialEntities, refreshFromServer, isLoading } = useBulkEntityOperations();

  // Initialize data and WebSocket listeners
  useEffect(() => {
    const initialize = async () => {
      try {
        if (!wsService.isConnected) {
          await wsService.connect();
        }
        await loadInitialEntities();
      } catch (error) {
        console.error('Failed to initialize tactical map:', error);
      }
    };

    initialize();

    // Monitor connection status
    const updateConnectionStatus = () => {
      setConnectionStatus(wsService.connectionState);
    };

    const unsubscribeConnection = wsService.onConnection(updateConnectionStatus);
    const unsubscribeDisconnection = wsService.onDisconnection(updateConnectionStatus);
    const unsubscribeError = wsService.onError(updateConnectionStatus);

    return () => {
      unsubscribeConnection();
      unsubscribeDisconnection();
      unsubscribeError();
    };
  }, [loadInitialEntities]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.altKey) {
        switch (event.code) {
          case 'KeyF': setFullscreen(prev => !prev); break;
          case 'KeyS': setSidebarOpen(prev => !prev); break;
          case 'KeyL': setShowLabels(prev => !prev); break;
          case 'KeyT': setShowTrails(prev => !prev); break;
          case 'KeyC': setEnableClustering(prev => !prev); break;
          case 'KeyR': refreshFromServer(); break;
          case 'KeyA': setShowAreaMaps(prev => !prev); break;
          case 'KeyD': setDrawingTool(prev => prev === 'select' ? 'polyline' : 'select'); break;
          case 'Escape': clearSelection(); setActiveTool('select'); setDrawingTool('select'); break;
        }
        event.preventDefault();
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, [refreshFromServer, clearSelection]);

  // Handle entity selection
  const handleEntityClick = useCallback((entity: Entity) => {
    selectEntity(entity.id);
  }, [selectEntity]);
  
  // Handle drawing events
  const handleDrawingCreated = useCallback((annotation: DrawingAnnotation) => {
    setAnnotations(prev => [...prev, annotation]);
    console.log('Drawing created:', annotation);
  }, []);
  
  const handleDrawingEdited = useCallback((annotation: DrawingAnnotation) => {
    setAnnotations(prev => prev.map(a => a.id === annotation.id ? annotation : a));
    console.log('Drawing edited:', annotation);
  }, []);
  
  const handleDrawingDeleted = useCallback((annotationId: string) => {
    setAnnotations(prev => prev.filter(a => a.id !== annotationId));
    console.log('Drawing deleted:', annotationId);
  }, []);
  
  // Load sample area maps
  const loadSampleAreaMaps = useCallback(() => {
    const sampleAreaMaps: AreaMapOverlay[] = [
      {
        id: 'sample-geojson',
        name: 'Sample Area',
        type: 'geojson',
        visible: true,
        opacity: 0.7,
        data: {
          type: 'FeatureCollection',
          features: [
            {
              type: 'Feature',
              properties: {
                name: 'Sample Polygon',
                description: 'This is a sample polygon area'
              },
              geometry: {
                type: 'Polygon',
                coordinates: [[
                  [-77.05, 38.87],
                  [-77.04, 38.87],
                  [-77.04, 38.88],
                  [-77.05, 38.88],
                  [-77.05, 38.87]
                ]]
              }
            }
          ]
        },
        style: {
          color: '#ff0000',
          weight: 2,
          opacity: 0.8,
          fillOpacity: 0.3
        }
      }
    ];
    setAreaMaps(sampleAreaMaps);
  }, []);

  // Map control handlers
  const handleZoomIn = () => {
    const newZoom = Math.min(viewport.zoom + 1, 18);
    setViewport(prev => ({ ...prev, zoom: newZoom }));
    mapRef.current?.setZoom(newZoom);
  };

  const handleZoomOut = () => {
    const newZoom = Math.max(viewport.zoom - 1, 3);
    setViewport(prev => ({ ...prev, zoom: newZoom }));
    mapRef.current?.setZoom(newZoom);
  };

  const handleFitAll = () => {
    mapRef.current?.fitBounds(entities);
  };

  const handleResetView = () => {
    const defaultViewport = { lat: 38.8951, lng: -77.0364 };
    setViewport({ center: defaultViewport, zoom: 10 });
    mapRef.current?.panTo(defaultViewport.lat, defaultViewport.lng);
    mapRef.current?.setZoom(10);
  };

  // Filter entities by search query
  const filteredEntities = entities.filter(entity => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      entity.callsign?.toLowerCase().includes(query) ||
      entity.uid.toLowerCase().includes(query) ||
      entity.entityType.toLowerCase().includes(query)
    );
  });

  // Get filter button styles
  const getFilterButtonClass = (filterType: EntityFilterType) => {
    return `filter-btn ${filter === filterType ? 'active' : ''}`;
  };

  return (
    <div className={`tactical-map-container ${fullscreen ? 'fullscreen' : ''}`}>
      {/* Map Header */}
      <header className="map-header">
        <div className="header-left">
          <h1 className="map-title font-display font-bold text-xl text-primary">
            🗺️ Map
          </h1>
          <div className="connection-indicator">
            <div 
              className="status-dot"
              style={{ 
                backgroundColor: connectionStatus === ConnectionState.CONNECTED 
                  ? 'var(--color-success)' 
                  : 'var(--color-error)' 
              }}
            />
            <span className="status-text font-mono text-xs">
              {connectionStatus.toUpperCase()}
            </span>
          </div>
        </div>

        <div className="header-controls">
          <div className="search-box">
            <input
              type="text"
              placeholder="Search entities..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
            />
            <span className="search-icon">🔍</span>
          </div>
          
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="btn-ghost"
            title="Toggle Sidebar"
          >
            {sidebarOpen ? '◀' : '▶'}
          </button>
          
          <button
            onClick={() => setFullscreen(!fullscreen)}
            className="btn-ghost"
            title="Toggle Fullscreen"
          >
            {fullscreen ? '📱' : '🖥️'}
          </button>
        </div>
      </header>

      <div className="map-content">
        {/* Control Sidebar */}
        {sidebarOpen && (
          <aside className="map-sidebar">
            {/* Entity Filters */}
            <div className="control-section">
              <h3 className="section-title">Entity Filters</h3>
              <div className="filter-buttons">
                <button
                  className={getFilterButtonClass('all')}
                  onClick={() => setFilter('all')}
                >
                  All ({counts.total})
                </button>
                <button
                  className={getFilterButtonClass('friendly')}
                  onClick={() => setFilter('friendly')}
                >
                  <span className="filter-indicator friendly"></span>
                  Friendly ({counts.friendly})
                </button>
                <button
                  className={getFilterButtonClass('hostile')}
                  onClick={() => setFilter('hostile')}
                >
                  <span className="filter-indicator hostile"></span>
                  Hostile ({counts.hostile})
                </button>
                <button
                  className={getFilterButtonClass('unknown')}
                  onClick={() => setFilter('unknown')}
                >
                  <span className="filter-indicator unknown"></span>
                  Unknown ({counts.unknown})
                </button>
              </div>
            </div>

            {/* Map Layers */}
            <div className="control-section">
              <h3 className="section-title">Map Layers</h3>
              <div className="layer-buttons">
                {Object.entries(MAP_LAYERS).map(([key, config]) => (
                  <button
                    key={key}
                    className={`layer-btn ${mapLayer === key ? 'active' : ''}`}
                    onClick={() => setMapLayer(key as MapLayer)}
                  >
                    {config.name}
                  </button>
                ))}
              </div>
            </div>
            
            {/* Drawing Tools */}
            <div className="control-section">
              <h3 className="section-title">Drawing Tools</h3>
              <div className="tool-buttons">
                {(['select', 'marker', 'polyline', 'polygon', 'rectangle', 'circle', 'measure'] as DrawingTool[]).map(tool => (
                  <button
                    key={tool}
                    className={`tool-btn ${drawingTool === tool ? 'active' : ''}`}
                    onClick={() => setDrawingTool(tool)}
                  >
                    {tool === 'select' && '🔍'}
                    {tool === 'marker' && '📍'}
                    {tool === 'polyline' && '📏'}
                    {tool === 'polygon' && '⬟'}
                    {tool === 'rectangle' && '▭'}
                    {tool === 'circle' && '⭕'}
                    {tool === 'measure' && '📐'}
                    {tool.charAt(0).toUpperCase() + tool.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Map Tools */}
            <div className="control-section">
              <h3 className="section-title">Tools</h3>
              <div className="tool-buttons">
                {(['select', 'measure', 'draw', 'route'] as MapTool[]).map(tool => (
                  <button
                    key={tool}
                    className={`tool-btn ${activeTool === tool ? 'active' : ''}`}
                    onClick={() => setActiveTool(tool)}
                  >
                    {tool === 'select' && '🔍'}
                    {tool === 'measure' && '📏'}
                    {tool === 'draw' && '✏️'}
                    {tool === 'route' && '🛤️'}
                    {tool.charAt(0).toUpperCase() + tool.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Display Options */}
            <div className="control-section">
              <h3 className="section-title">Display</h3>
              <div className="display-options">
                <label className="option-item">
                  <input
                    type="checkbox"
                    checked={showLabels}
                    onChange={(e) => setShowLabels(e.target.checked)}
                  />
                  <span>Show Labels</span>
                </label>
                <label className="option-item">
                  <input
                    type="checkbox"
                    checked={showTrails}
                    onChange={(e) => setShowTrails(e.target.checked)}
                  />
                  <span>Show Trails</span>
                </label>
                <label className="option-item">
                  <input
                    type="checkbox"
                    checked={enableClustering}
                    onChange={(e) => setEnableClustering(e.target.checked)}
                  />
                  <span>Cluster Entities</span>
                </label>
                <label className="option-item">
                  <input
                    type="checkbox"
                    checked={showAreaMaps}
                    onChange={(e) => setShowAreaMaps(e.target.checked)}
                  />
                  <span>Area Maps</span>
                </label>
              </div>
              
              <div style={{ marginTop: 'var(--space-3)' }}>
                <label style={{ display: 'block', fontWeight: 'var(--font-weight-semibold)', marginBottom: 'var(--space-1)', fontSize: 'var(--font-size-xs)' }}>Coordinates:</label>
                <select 
                  value={coordinateFormat}
                  onChange={(e) => setCoordinateFormat(e.target.value as any)}
                  style={{
                    width: '100%',
                    padding: 'var(--space-1)',
                    fontSize: 'var(--font-size-xs)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--radius-sm)',
                    backgroundColor: 'var(--color-bg-primary)'
                  }}
                >
                  <option value="dd">Decimal Degrees</option>
                  <option value="dms">Degrees, Minutes, Seconds</option>
                  <option value="utm">UTM</option>
                  <option value="mgrs">MGRS</option>
                </select>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="control-section">
              <h3 className="section-title">Actions</h3>
              <div className="action-buttons">
                <button onClick={handleFitAll} className="action-btn">
                  📐 Fit All
                </button>
                <button onClick={handleResetView} className="action-btn">
                  🏠 Reset View
                </button>
                <button 
                  onClick={refreshFromServer} 
                  className="action-btn"
                  disabled={isLoading}
                >
                  {isLoading ? '⟳' : '🔄'} Refresh
                </button>
                <button onClick={clearSelection} className="action-btn">
                  ❌ Clear Selection
                </button>
                <button onClick={loadSampleAreaMaps} className="action-btn">
                  🗺️ Load Areas
                </button>
                <button 
                  onClick={() => {
                    const geojson = {
                      type: 'FeatureCollection',
                      features: annotations.map(a => ({
                        type: 'Feature',
                        properties: a.properties,
                        geometry: null // Would need proper geometry extraction
                      }))
                    };
                    console.log('Export annotations:', geojson);
                  }}
                  className="action-btn"
                >
                  📤 Export
                </button>
              </div>
            </div>

            {/* Entity List */}
            {filteredEntities.length > 0 && (
              <div className="control-section">
                <h3 className="section-title">
                  Entities ({filteredEntities.length})
                </h3>
                <div className="entity-list">
                  {filteredEntities.slice(0, 10).map((entity) => (
                    <div
                      key={entity.id}
                      className={`entity-item ${selectedEntity?.id === entity.id ? 'selected' : ''}`}
                      onClick={() => handleEntityClick(entity)}
                    >
                      <div className="entity-header">
                        <span className="entity-callsign">
                          {entity.callsign || entity.uid.slice(0, 8)}
                        </span>
                        <span className={`entity-type-badge ${
                          entity.entityType.startsWith('a-f') ? 'friendly' :
                          entity.entityType.startsWith('a-h') ? 'hostile' : 'unknown'
                        }`}>
                          {entity.entityType.startsWith('a-f') ? 'F' :
                           entity.entityType.startsWith('a-h') ? 'H' : 'U'}
                        </span>
                      </div>
                      <div className="entity-details">
                        <div className="entity-position font-mono text-xs">
                          {entity.position ? 
                            `${entity.position.lat.toFixed(4)}, ${entity.position.lng.toFixed(4)}` :
                            'No position'
                          }
                        </div>
                        <div className="entity-updated text-xs text-muted">
                          {new Date(entity.lastUpdate).toLocaleTimeString()}
                        </div>
                      </div>
                    </div>
                  ))}
                  {filteredEntities.length > 10 && (
                    <div className="entity-more">
                      +{filteredEntities.length - 10} more entities
                    </div>
                  )}
                </div>
              </div>
            )}
          </aside>
        )}

        {/* Main Map Area */}
        <main className="map-main">
          {/* Map Controls Overlay */}
          <div className="map-controls-overlay">
            <div className="zoom-controls">
              <button onClick={handleZoomIn} className="zoom-btn">+</button>
              <div className="zoom-level font-mono text-xs">
                {viewport.zoom}
              </div>
              <button onClick={handleZoomOut} className="zoom-btn">-</button>
            </div>

            <div className="layer-info">
              <div className="current-layer font-mono text-xs">
                {mapLayer.toUpperCase()}
              </div>
            </div>
          </div>

          {/* Entity Map */}
          <EntityMap
            ref={mapRef}
            onEntityClick={handleEntityClick}
            showEntityLabels={showLabels}
            showEntityTrails={showTrails}
            enableClustering={enableClustering}
            currentLayer={mapLayer}
            showAreaMaps={showAreaMaps}
            areaMaps={areaMaps}
            style={{ width: '100%', height: '100%' }}
          >
            <DrawingTools
              activeTool={drawingTool}
              onDrawingCreated={handleDrawingCreated}
              onDrawingEdited={handleDrawingEdited}
              onDrawingDeleted={handleDrawingDeleted}
              enabled={drawingTool !== 'select'}
            />
          </EntityMap>

          {/* Status Overlay */}
          <div className="status-overlay">
            <div className="status-item">
              <span className="status-label">Zoom:</span>
              <span className="status-value">{viewport.zoom}</span>
            </div>
            <div className="status-item">
              <span className="status-label">Center:</span>
              <span className="status-value font-mono">
                {formatCoordinates(viewport.center.lat, viewport.center.lng, { format: coordinateFormat })}
              </span>
            </div>
            <div className="status-item">
              <span className="status-label">Entities:</span>
              <span className="status-value">{filteredEntities.length}</span>
            </div>
          </div>
        </main>
      </div>

      {/* Tactical Map Styles */}
      <style jsx>{`
        .tactical-map-container {
          height: 100vh;
          display: flex;
          flex-direction: column;
          background-color: var(--color-bg-primary);
        }

        .tactical-map-container.fullscreen {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          z-index: var(--z-modal);
        }

        .map-header {
          height: 60px;
          background-color: var(--color-bg-secondary);
          border-bottom: 1px solid var(--color-border);
          padding: 0 var(--space-6);
          display: flex;
          justify-content: space-between;
          align-items: center;
          box-shadow: var(--shadow-sm);
        }

        .header-left {
          display: flex;
          align-items: center;
          gap: var(--space-4);
        }

        .map-title {
          margin: 0;
        }

        .connection-indicator {
          display: flex;
          align-items: center;
          gap: var(--space-1);
          padding: var(--space-1) var(--space-2);
          background-color: var(--color-surface);
          border-radius: var(--radius-sm);
        }

        .status-dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          animation: pulse 2s infinite;
        }

        .header-controls {
          display: flex;
          align-items: center;
          gap: var(--space-3);
        }

        .search-box {
          position: relative;
        }

        .search-input {
          padding: var(--space-2) var(--space-8) var(--space-2) var(--space-3);
          background-color: var(--color-bg-primary);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          color: var(--color-text-primary);
          font-size: var(--font-size-sm);
          width: 200px;
        }

        .search-input:focus {
          outline: none;
          border-color: var(--color-border-accent);
        }

        .search-icon {
          position: absolute;
          right: var(--space-2);
          top: 50%;
          transform: translateY(-50%);
          color: var(--color-text-muted);
        }

        .map-content {
          flex: 1;
          display: flex;
          overflow: hidden;
        }

        .map-sidebar {
          width: 280px;
          background-color: var(--color-bg-secondary);
          border-right: 1px solid var(--color-border);
          padding: var(--space-4);
          overflow-y: auto;
          display: flex;
          flex-direction: column;
          gap: var(--space-6);
        }

        .control-section {
          background-color: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-lg);
          padding: var(--space-4);
        }

        .section-title {
          margin: 0 0 var(--space-3) 0;
          color: var(--color-text-primary);
          font-size: var(--font-size-sm);
          font-weight: var(--font-weight-semibold);
          text-transform: uppercase;
          letter-spacing: var(--tracking-wide);
        }

        .filter-buttons,
        .layer-buttons,
        .tool-buttons {
          display: flex;
          flex-direction: column;
          gap: var(--space-2);
        }

        .filter-btn,
        .layer-btn,
        .tool-btn,
        .action-btn {
          padding: var(--space-2) var(--space-3);
          background-color: var(--color-bg-tertiary);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          color: var(--color-text-secondary);
          font-size: var(--font-size-sm);
          cursor: pointer;
          transition: all var(--transition-fast);
          text-align: left;
        }

        .filter-btn:hover,
        .layer-btn:hover,
        .tool-btn:hover,
        .action-btn:hover {
          background-color: var(--color-surface-hover);
          border-color: var(--color-border-light);
          color: var(--color-text-primary);
        }

        .filter-btn.active,
        .layer-btn.active,
        .tool-btn.active {
          background-color: var(--color-surface-active);
          border-color: var(--color-border-accent);
          color: var(--color-text-accent);
        }

        .filter-btn {
          display: flex;
          align-items: center;
          gap: var(--space-2);
        }

        .filter-indicator {
          width: 8px;
          height: 8px;
          border-radius: 50%;
        }

        .filter-indicator.friendly {
          background-color: var(--color-friendly);
        }

        .filter-indicator.hostile {
          background-color: var(--color-hostile);
        }

        .filter-indicator.unknown {
          background-color: var(--color-unknown);
        }

        .tool-btn {
          display: flex;
          align-items: center;
          gap: var(--space-2);
        }

        .display-options {
          display: flex;
          flex-direction: column;
          gap: var(--space-2);
        }

        .option-item {
          display: flex;
          align-items: center;
          gap: var(--space-2);
          cursor: pointer;
          font-size: var(--font-size-sm);
          color: var(--color-text-secondary);
        }

        .option-item input[type="checkbox"] {
          accent-color: var(--color-friendly);
        }

        .action-buttons {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: var(--space-2);
        }

        .action-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .entity-list {
          max-height: 300px;
          overflow-y: auto;
          display: flex;
          flex-direction: column;
          gap: var(--space-2);
        }

        .entity-item {
          padding: var(--space-3);
          background-color: var(--color-bg-tertiary);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          cursor: pointer;
          transition: all var(--transition-fast);
        }

        .entity-item:hover {
          background-color: var(--color-surface-hover);
          border-color: var(--color-border-light);
        }

        .entity-item.selected {
          background-color: var(--color-surface-active);
          border-color: var(--color-border-accent);
        }

        .entity-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: var(--space-1);
        }

        .entity-callsign {
          font-weight: var(--font-weight-semibold);
          color: var(--color-text-primary);
          font-size: var(--font-size-sm);
        }

        .entity-type-badge {
          width: 20px;
          height: 20px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: var(--font-size-xs);
          font-weight: var(--font-weight-bold);
          color: white;
        }

        .entity-type-badge.friendly {
          background-color: var(--color-friendly);
        }

        .entity-type-badge.hostile {
          background-color: var(--color-hostile);
        }

        .entity-type-badge.unknown {
          background-color: var(--color-unknown);
        }

        .entity-details {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .entity-more {
          text-align: center;
          padding: var(--space-2);
          color: var(--color-text-muted);
          font-size: var(--font-size-xs);
        }

        .map-main {
          flex: 1;
          position: relative;
          overflow: hidden;
        }

        .map-controls-overlay {
          position: absolute;
          top: var(--space-4);
          right: var(--space-4);
          z-index: var(--z-sticky);
          display: flex;
          flex-direction: column;
          gap: var(--space-3);
        }

        .zoom-controls {
          background-color: rgba(0, 0, 0, 0.8);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          padding: var(--space-1);
          display: flex;
          flex-direction: column;
          align-items: center;
        }

        .zoom-btn {
          width: 32px;
          height: 32px;
          background-color: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-sm);
          color: var(--color-text-primary);
          font-size: var(--font-size-lg);
          font-weight: var(--font-weight-bold);
          cursor: pointer;
          transition: all var(--transition-fast);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .zoom-btn:hover {
          background-color: var(--color-surface-hover);
        }

        .zoom-level {
          padding: var(--space-1);
          color: var(--color-text-secondary);
          text-align: center;
          min-width: 24px;
        }

        .layer-info {
          background-color: rgba(0, 0, 0, 0.8);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          padding: var(--space-2) var(--space-3);
        }

        .current-layer {
          color: var(--color-text-secondary);
          text-align: center;
        }

        .status-overlay {
          position: absolute;
          bottom: var(--space-4);
          left: var(--space-4);
          background-color: rgba(0, 0, 0, 0.8);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          padding: var(--space-3);
          display: flex;
          gap: var(--space-4);
          font-size: var(--font-size-xs);
        }

        .status-item {
          display: flex;
          gap: var(--space-1);
        }

        .status-label {
          color: var(--color-text-muted);
        }

        .status-value {
          color: var(--color-text-secondary);
        }

        /* Responsive Design */
        @media (max-width: 768px) {
          .map-sidebar {
            width: 100%;
            position: absolute;
            left: -100%;
            top: 0;
            height: 100%;
            z-index: var(--z-overlay);
            transition: left var(--transition-base);
            box-shadow: var(--shadow-lg);
          }

          .map-sidebar.open {
            left: 0;
          }

          .search-input {
            width: 150px;
          }

          .status-overlay {
            flex-direction: column;
            gap: var(--space-1);
          }

          .action-buttons {
            grid-template-columns: 1fr;
          }
        }

        @media (max-width: 480px) {
          .map-header {
            padding: 0 var(--space-4);
          }

          .header-controls {
            gap: var(--space-2);
          }

          .search-input {
            width: 120px;
          }
        }
      `}</style>
    </div>
  );
};

export default TacticalMap;
