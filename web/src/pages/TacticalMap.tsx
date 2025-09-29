/**
 * Enhanced Full-Width Tactical Map
 * Modern, full-screen tactical mapping interface with top toolbar controls
 */

import React, { useState, useRef, useCallback, useEffect } from 'react';
import EntityMap, { EntityMapRef, MAP_LAYERS } from '../components/Map/EntityMap';
import DrawingTools, { DrawingTool, DrawingAnnotation } from '../components/Map/DrawingTools';
import { useFilteredEntities, useSelectedEntity, useBulkEntityOperations } from '../hooks/useEntityTracker';
import { EntityFilterType } from '../stores/entityTracker';
import { wsService, ConnectionState } from '../services/websocketService';
import { Entity } from '../services/apiClient';
import { formatCoordinates, CoordinateDisplayOptions } from '../utils/coordinates';
import './TacticalMap.css';

type MapLayer = keyof typeof MAP_LAYERS;

const TacticalMap: React.FC = () => {
  // Map state - Default to dark tactical mode for operations
  const [mapLayer, setMapLayer] = useState<MapLayer>('darkTactical');
  const [drawingTool, setDrawingTool] = useState<DrawingTool>('select');
  const [showLabels, setShowLabels] = useState(true);
  const [showTrails, setShowTrails] = useState(false);
  const [enableClustering, setEnableClustering] = useState(true);
  const [coordinateFormat, setCoordinateFormat] = useState<CoordinateDisplayOptions['format']>('dd');
  const [searchQuery, setSearchQuery] = useState('');
  
  // Dropdown states for toolbar menus
  const [layerMenuOpen, setLayerMenuOpen] = useState(false);
  const [drawingMenuOpen, setDrawingMenuOpen] = useState(false);
  const [optionsMenuOpen, setOptionsMenuOpen] = useState(false);
  
  // Annotations state
  const [annotations, setAnnotations] = useState<DrawingAnnotation[]>([]);
  
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

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('.dropdown-menu')) {
        setLayerMenuOpen(false);
        setDrawingMenuOpen(false);
        setOptionsMenuOpen(false);
      }
    };

    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // ESC to clear selection
      if (event.key === 'Escape') {
        clearSelection();
        setDrawingTool('select');
        setLayerMenuOpen(false);
        setDrawingMenuOpen(false);
        setOptionsMenuOpen(false);
      }
      
      // L for labels
      if (event.key === 'l' && !event.ctrlKey && !event.metaKey) {
        setShowLabels(prev => !prev);
      }
      
      // T for trails
      if (event.key === 't' && !event.ctrlKey && !event.metaKey) {
        setShowTrails(prev => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, [clearSelection]);

  // Handle entity selection
  const handleEntityClick = useCallback((entity: Entity) => {
    selectEntity(entity.id);
  }, [selectEntity]);
  
  // Handle drawing events
  const handleDrawingCreated = useCallback((annotation: DrawingAnnotation) => {
    setAnnotations(prev => [...prev, annotation]);
  }, []);
  
  const handleDrawingEdited = useCallback((annotation: DrawingAnnotation) => {
    setAnnotations(prev => prev.map(a => a.id === annotation.id ? annotation : a));
  }, []);
  
  const handleDrawingDeleted = useCallback((annotationId: string) => {
    setAnnotations(prev => prev.filter(a => a.id !== annotationId));
  }, []);
  
  // Map control handlers
  const handleFitAll = () => {
    mapRef.current?.fitBounds(entities);
  };

  const handleResetView = () => {
    const defaultViewport = { lat: 38.8951, lng: -77.0364 };
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

  const drawingToolIcons: Record<DrawingTool, string> = {
    select: '↖',
    marker: '📍',
    polyline: '〰',
    polygon: '⬟',
    rectangle: '▭',
    circle: '○',
    measure: '📏'
  };

  return (
    <div className="tactical-map-page">
      {/* Enhanced Toolbar Header */}
      <header className="map-toolbar">
        {/* Left Section - Title and Entity Filters */}
        <div className="toolbar-section toolbar-left">
          <h1 className="map-title">TACTICAL MAP</h1>
          
          <div className="divider" />
          
          {/* Entity Filters */}
          <div className="filter-group">
            <button
              className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
              onClick={() => setFilter('all')}
              title={`All Entities (${counts.total})`}
            >
              All <span className="count">{counts.total}</span>
            </button>
            <button
              className={`filter-btn friendly ${filter === 'friendly' ? 'active' : ''}`}
              onClick={() => setFilter('friendly')}
              title={`Friendly (${counts.friendly})`}
            >
              <span className="indicator"></span>
              {counts.friendly}
            </button>
            <button
              className={`filter-btn hostile ${filter === 'hostile' ? 'active' : ''}`}
              onClick={() => setFilter('hostile')}
              title={`Hostile (${counts.hostile})`}
            >
              <span className="indicator"></span>
              {counts.hostile}
            </button>
            <button
              className={`filter-btn unknown ${filter === 'unknown' ? 'active' : ''}`}
              onClick={() => setFilter('unknown')}
              title={`Unknown (${counts.unknown})`}
            >
              <span className="indicator"></span>
              {counts.unknown}
            </button>
          </div>
        </div>

        {/* Center Section - Map Controls and Tools */}
        <div className="toolbar-section toolbar-center">
          {/* Map Layer Dropdown */}
          <div className="dropdown-menu">
            <button
              className={`toolbar-btn ${layerMenuOpen ? 'active' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                setLayerMenuOpen(!layerMenuOpen);
                setDrawingMenuOpen(false);
                setOptionsMenuOpen(false);
              }}
              title="Select Map Layer"
            >
              <span className="btn-icon">🗺️</span>
              <span className="btn-label">Layers</span>
              <span className="dropdown-arrow">▾</span>
            </button>
            {layerMenuOpen && (
              <div className="dropdown-content">
                <div className="dropdown-section-label">Dark Mode (Tactical)</div>
                {Object.entries(MAP_LAYERS)
                  .filter(([key]) => key.startsWith('dark'))
                  .map(([key, config]) => (
                    <button
                      key={key}
                      className={`dropdown-item ${mapLayer === key ? 'active' : ''}`}
                      onClick={() => {
                        setMapLayer(key as MapLayer);
                        setLayerMenuOpen(false);
                      }}
                    >
                      <span className="item-icon">🌙</span>
                      {config.name}
                    </button>
                  ))}
                <div className="dropdown-divider" />
                <div className="dropdown-section-label">Light Mode</div>
                {Object.entries(MAP_LAYERS)
                  .filter(([key]) => !key.startsWith('dark'))
                  .map(([key, config]) => (
                    <button
                      key={key}
                      className={`dropdown-item ${mapLayer === key ? 'active' : ''}`}
                      onClick={() => {
                        setMapLayer(key as MapLayer);
                        setLayerMenuOpen(false);
                      }}
                    >
                      <span className="item-icon">☀️</span>
                      {config.name}
                    </button>
                  ))}
              </div>
            )}
          </div>

          {/* Drawing Tools Dropdown */}
          <div className="dropdown-menu">
            <button
              className={`toolbar-btn ${drawingMenuOpen ? 'active' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                setDrawingMenuOpen(!drawingMenuOpen);
                setLayerMenuOpen(false);
                setOptionsMenuOpen(false);
              }}
              title="Drawing & Annotation Tools"
            >
              <span className="btn-icon">{drawingToolIcons[drawingTool]}</span>
              <span className="btn-label">Draw</span>
              <span className="dropdown-arrow">▾</span>
            </button>
            {drawingMenuOpen && (
              <div className="dropdown-content">
                {(Object.keys(drawingToolIcons) as DrawingTool[]).map(tool => (
                  <button
                    key={tool}
                    className={`dropdown-item ${drawingTool === tool ? 'active' : ''}`}
                    onClick={() => {
                      setDrawingTool(tool);
                      setDrawingMenuOpen(false);
                    }}
                  >
                    <span className="item-icon">{drawingToolIcons[tool]}</span>
                    <span className="item-label">{tool.charAt(0).toUpperCase() + tool.slice(1)}</span>
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="divider" />

          {/* Quick Toggle Buttons */}
          <button
            className={`toolbar-btn icon-btn ${mapLayer.startsWith('dark') ? 'active' : ''}`}
            onClick={() => {
              // Toggle between dark and light modes
              if (mapLayer.startsWith('dark')) {
                setMapLayer('osm');
              } else {
                setMapLayer('darkTactical');
              }
            }}
            title={mapLayer.startsWith('dark') ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
          >
            <span className="btn-icon">{mapLayer.startsWith('dark') ? '🌙' : '☀️'}</span>
          </button>
          <button
            className={`toolbar-btn icon-btn ${showLabels ? 'active' : ''}`}
            onClick={() => setShowLabels(!showLabels)}
            title="Toggle Entity Labels • Press L"
          >
            <span className="btn-icon">🏷️</span>
          </button>
          <button
            className={`toolbar-btn icon-btn ${showTrails ? 'active' : ''}`}
            onClick={() => setShowTrails(!showTrails)}
            title="Toggle Movement Trails • Press T"
          >
            <span className="btn-icon">〰️</span>
          </button>
          <button
            className={`toolbar-btn icon-btn ${enableClustering ? 'active' : ''}`}
            onClick={() => setEnableClustering(!enableClustering)}
            title="Group Nearby Entities"
          >
            <span className="btn-icon">⚡</span>
          </button>

          <div className="divider" />

          {/* View Controls */}
          <button
            className="toolbar-btn icon-btn"
            onClick={handleFitAll}
            title="Fit All Entities in View"
          >
            <span className="btn-icon">⊑</span>
          </button>
          <button
            className="toolbar-btn icon-btn"
            onClick={handleResetView}
            title="Reset to Default View"
          >
            <span className="btn-icon">⌂</span>
          </button>

          {/* Options Dropdown */}
          <div className="dropdown-menu">
            <button
              className={`toolbar-btn icon-btn ${optionsMenuOpen ? 'active' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                setOptionsMenuOpen(!optionsMenuOpen);
                setLayerMenuOpen(false);
                setDrawingMenuOpen(false);
              }}
              title="Map Settings & Options"
            >
              <span className="btn-icon">⚙️</span>
            </button>
            {optionsMenuOpen && (
              <div className="dropdown-content dropdown-right">
                <div className="dropdown-section">
                  <label className="dropdown-label">Coordinate Format</label>
                  <select 
                    className="dropdown-select"
                    value={coordinateFormat}
                    onChange={(e) => setCoordinateFormat(e.target.value as any)}
                  >
                    <option value="dd">Decimal Degrees</option>
                    <option value="dms">DMS</option>
                    <option value="utm">UTM</option>
                    <option value="mgrs">MGRS</option>
                  </select>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Right Section - Search and Status */}
        <div className="toolbar-section toolbar-right">
          {/* Search */}
          <div className="search-container" title="Search by callsign, ID, or type">
            <input
              type="text"
              placeholder="Search entities..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
              title="Search entities by callsign, ID, or type"
            />
            <span className="search-icon">🔍</span>
          </div>

          {/* Refresh Button */}
          <button
            className={`toolbar-btn icon-btn ${isLoading ? 'loading' : ''}`}
            onClick={refreshFromServer}
            disabled={isLoading}
            title="Refresh Entity Data"
          >
            <span className="btn-icon">↻</span>
          </button>

          {/* Connection Status */}
          <div className="connection-status">
            <div 
              className="status-indicator"
              data-status={connectionStatus}
              title={connectionStatus}
            />
          </div>
        </div>
      </header>

      {/* Selected Entity Info Bar */}
      {selectedEntity && (
        <div className="entity-info-bar">
          <div className="entity-info">
            <span className="info-label">Callsign:</span>
            <span className="info-value">{selectedEntity.callsign || 'Unknown'}</span>
          </div>
          <div className="entity-info">
            <span className="info-label">Type:</span>
            <span className="info-value">{selectedEntity.entityType}</span>
          </div>
          <div className="entity-info">
            <span className="info-label">Position:</span>
            <span className="info-value mono">
              {selectedEntity.position ? 
                formatCoordinates(
                  selectedEntity.position.lat, 
                  selectedEntity.position.lng, 
                  { format: coordinateFormat }
                ) : 'No position'
              }
            </span>
          </div>
          <button 
            className="clear-btn"
            onClick={clearSelection}
            title="Clear Selection (ESC)"
          >
            ✕
          </button>
        </div>
      )}

      {/* Full-Width Map */}
      <div className="map-container">
        <EntityMap
          ref={mapRef}
          onEntityClick={handleEntityClick}
          showEntityLabels={showLabels}
          showEntityTrails={showTrails}
          enableClustering={enableClustering}
          currentLayer={mapLayer}
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
      </div>
    </div>
  );
};

export default TacticalMap;
