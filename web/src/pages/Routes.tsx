/**
 * Routes Page - Advanced Mission Planning & Navigation
 * Comprehensive route planning tool for tactical operations
 */

import React, { useState, useRef, useCallback, useEffect } from 'react';
import { wsService } from '../services/websocketService';
import './Routes.css';

// Types
interface Waypoint {
  id: string;
  name: string;
  position: { lat: number; lng: number; alt?: number };
  type: 'start' | 'waypoint' | 'checkpoint' | 'rally' | 'objective' | 'extraction';
  eta?: string;
  notes?: string;
  threats?: string[];
  completed?: boolean;
}

interface Route {
  id: string;
  name: string;
  type: 'primary' | 'alternate' | 'emergency' | 'patrol';
  status: 'planned' | 'active' | 'completed' | 'aborted';
  waypoints: Waypoint[];
  distance: number; // in meters
  estimatedTime: number; // in minutes
  terrainType: string[];
  threatLevel: 'low' | 'medium' | 'high' | 'critical';
  assignedUnits: string[];
  createdAt: string;
  updatedAt: string;
}

interface MissionPhase {
  id: string;
  name: string;
  routes: string[]; // Route IDs
  startTime: string;
  endTime?: string;
  status: 'pending' | 'active' | 'completed';
}

const Routes: React.FC = () => {
  // State
  const [routes, setRoutes] = useState<Route[]>([]);
  const [selectedRoute, setSelectedRoute] = useState<Route | null>(null);
  const [activeView, setActiveView] = useState<'planning' | 'navigation' | 'analysis'>('planning');
  const [showNewRouteDialog, setShowNewRouteDialog] = useState(false);
  const [missionPhases, setMissionPhases] = useState<MissionPhase[]>([]);
  const [filterType, setFilterType] = useState<'all' | 'primary' | 'alternate' | 'emergency'>('all');
  const [isDrawing, setIsDrawing] = useState(false);
  const [showImportDialog, setShowImportDialog] = useState(false);
  
  // New route form
  const [newRoute, setNewRoute] = useState({
    name: '',
    type: 'primary' as Route['type'],
    threatLevel: 'low' as Route['threatLevel'],
  });

  // Mock data initialization
  useEffect(() => {
    // Initialize with sample routes
    const sampleRoutes: Route[] = [
      {
        id: 'route-1',
        name: 'Alpha Approach',
        type: 'primary',
        status: 'planned',
        waypoints: [
          { id: 'wp-1', name: 'Base', position: { lat: 38.8951, lng: -77.0364 }, type: 'start' },
          { id: 'wp-2', name: 'Checkpoint Alpha', position: { lat: 38.9051, lng: -77.0464 }, type: 'checkpoint' },
          { id: 'wp-3', name: 'Objective', position: { lat: 38.9151, lng: -77.0564 }, type: 'objective' },
        ],
        distance: 12500,
        estimatedTime: 45,
        terrainType: ['urban', 'residential'],
        threatLevel: 'medium',
        assignedUnits: ['ALPHA-1', 'ALPHA-2'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 'route-2',
        name: 'Bravo Alternate',
        type: 'alternate',
        status: 'planned',
        waypoints: [
          { id: 'wp-4', name: 'Base', position: { lat: 38.8951, lng: -77.0364 }, type: 'start' },
          { id: 'wp-5', name: 'Rally Point', position: { lat: 38.8851, lng: -77.0264 }, type: 'rally' },
          { id: 'wp-6', name: 'Objective', position: { lat: 38.9151, lng: -77.0564 }, type: 'objective' },
        ],
        distance: 15200,
        estimatedTime: 55,
        terrainType: ['forest', 'hills'],
        threatLevel: 'low',
        assignedUnits: ['BRAVO-1'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 'route-3',
        name: 'Emergency Extract',
        type: 'emergency',
        status: 'planned',
        waypoints: [
          { id: 'wp-7', name: 'Current Position', position: { lat: 38.9151, lng: -77.0564 }, type: 'start' },
          { id: 'wp-8', name: 'Extract Point', position: { lat: 38.9251, lng: -77.0664 }, type: 'extraction' },
        ],
        distance: 3200,
        estimatedTime: 15,
        terrainType: ['open'],
        threatLevel: 'critical',
        assignedUnits: ['QRF-1'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];

    setRoutes(sampleRoutes);
    setSelectedRoute(sampleRoutes[0]);

    // Sample mission phases
    setMissionPhases([
      {
        id: 'phase-1',
        name: 'Infiltration',
        routes: ['route-1', 'route-2'],
        startTime: '2024-01-15T02:00:00Z',
        status: 'pending',
      },
      {
        id: 'phase-2',
        name: 'Objective',
        routes: ['route-1'],
        startTime: '2024-01-15T03:00:00Z',
        status: 'pending',
      },
      {
        id: 'phase-3',
        name: 'Extraction',
        routes: ['route-3'],
        startTime: '2024-01-15T04:00:00Z',
        status: 'pending',
      },
    ]);
  }, []);

  // Create new route
  const handleCreateRoute = () => {
    const route: Route = {
      id: `route-${Date.now()}`,
      name: newRoute.name,
      type: newRoute.type,
      status: 'planned',
      waypoints: [],
      distance: 0,
      estimatedTime: 0,
      terrainType: [],
      threatLevel: newRoute.threatLevel,
      assignedUnits: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    setRoutes([...routes, route]);
    setSelectedRoute(route);
    setShowNewRouteDialog(false);
    setNewRoute({ name: '', type: 'primary', threatLevel: 'low' });
    setIsDrawing(true);
  };

  // Filter routes
  const filteredRoutes = routes.filter(route => 
    filterType === 'all' || route.type === filterType
  );

  // Calculate route statistics
  const routeStats = {
    total: routes.length,
    active: routes.filter(r => r.status === 'active').length,
    totalDistance: routes.reduce((sum, r) => sum + r.distance, 0),
    avgTime: routes.length > 0 
      ? Math.round(routes.reduce((sum, r) => sum + r.estimatedTime, 0) / routes.length)
      : 0,
  };

  // Format distance
  const formatDistance = (meters: number) => {
    if (meters < 1000) return `${meters}m`;
    return `${(meters / 1000).toFixed(1)}km`;
  };

  // Format time
  const formatTime = (minutes: number) => {
    if (minutes < 60) return `${minutes}min`;
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}h ${mins}m`;
  };

  // Get threat level color
  const getThreatColor = (level: Route['threatLevel']) => {
    switch (level) {
      case 'low': return '#4caf50';
      case 'medium': return '#ff9800';
      case 'high': return '#ff5722';
      case 'critical': return '#f44336';
    }
  };

  // Get route type icon
  const getRouteIcon = (type: Route['type']) => {
    switch (type) {
      case 'primary': return '🎯';
      case 'alternate': return '🔄';
      case 'emergency': return '🚨';
      case 'patrol': return '🔍';
    }
  };

  // Get waypoint icon
  const getWaypointIcon = (type: Waypoint['type']) => {
    switch (type) {
      case 'start': return '🚩';
      case 'waypoint': return '📍';
      case 'checkpoint': return '✅';
      case 'rally': return '🤝';
      case 'objective': return '🎯';
      case 'extraction': return '🚁';
    }
  };

  return (
    <div className="routes-container">
      {/* Header Toolbar */}
      <header className="routes-header">
        <div className="header-left">
          <h1 className="page-title">MISSION ROUTES</h1>
          
          {/* View Tabs */}
          <div className="view-tabs">
            <button 
              className={`tab-btn ${activeView === 'planning' ? 'active' : ''}`}
              onClick={() => setActiveView('planning')}
            >
              Planning
            </button>
            <button 
              className={`tab-btn ${activeView === 'navigation' ? 'active' : ''}`}
              onClick={() => setActiveView('navigation')}
            >
              Navigation
            </button>
            <button 
              className={`tab-btn ${activeView === 'analysis' ? 'active' : ''}`}
              onClick={() => setActiveView('analysis')}
            >
              Analysis
            </button>
          </div>
        </div>

        <div className="header-right">
          {/* Quick Stats */}
          <div className="quick-stats">
            <div className="stat">
              <span className="stat-value">{routeStats.total}</span>
              <span className="stat-label">Routes</span>
            </div>
            <div className="stat">
              <span className="stat-value">{routeStats.active}</span>
              <span className="stat-label">Active</span>
            </div>
            <div className="stat">
              <span className="stat-value">{formatDistance(routeStats.totalDistance)}</span>
              <span className="stat-label">Total</span>
            </div>
          </div>

          {/* Action Buttons */}
          <button 
            className="action-btn primary"
            onClick={() => setShowNewRouteDialog(true)}
          >
            + New Route
          </button>
          <button 
            className="action-btn"
            onClick={() => setShowImportDialog(true)}
          >
            Import
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div className="routes-content">
        {/* Sidebar - Routes List */}
        <aside className="routes-sidebar">
          {/* Filter Tabs */}
          <div className="filter-tabs">
            <button 
              className={`filter-btn ${filterType === 'all' ? 'active' : ''}`}
              onClick={() => setFilterType('all')}
            >
              All
            </button>
            <button 
              className={`filter-btn ${filterType === 'primary' ? 'active' : ''}`}
              onClick={() => setFilterType('primary')}
            >
              Primary
            </button>
            <button 
              className={`filter-btn ${filterType === 'alternate' ? 'active' : ''}`}
              onClick={() => setFilterType('alternate')}
            >
              Alternate
            </button>
            <button 
              className={`filter-btn ${filterType === 'emergency' ? 'active' : ''}`}
              onClick={() => setFilterType('emergency')}
            >
              Emergency
            </button>
          </div>

          {/* Routes List */}
          <div className="routes-list">
            {filteredRoutes.map(route => (
              <div 
                key={route.id}
                className={`route-card ${selectedRoute?.id === route.id ? 'selected' : ''}`}
                onClick={() => setSelectedRoute(route)}
              >
                <div className="route-header">
                  <span className="route-icon">{getRouteIcon(route.type)}</span>
                  <div className="route-info">
                    <h3 className="route-name">{route.name}</h3>
                    <span className="route-type">{route.type}</span>
                  </div>
                  <div 
                    className="threat-indicator"
                    style={{ backgroundColor: getThreatColor(route.threatLevel) }}
                    title={`Threat: ${route.threatLevel}`}
                  />
                </div>
                <div className="route-stats">
                  <span className="stat">
                    <span className="icon">📏</span>
                    {formatDistance(route.distance)}
                  </span>
                  <span className="stat">
                    <span className="icon">⏱️</span>
                    {formatTime(route.estimatedTime)}
                  </span>
                  <span className="stat">
                    <span className="icon">📍</span>
                    {route.waypoints.length}
                  </span>
                </div>
                {route.status === 'active' && (
                  <div className="active-badge">ACTIVE</div>
                )}
              </div>
            ))}
          </div>

          {/* Mission Phases */}
          <div className="mission-phases">
            <h3 className="section-title">Mission Phases</h3>
            {missionPhases.map(phase => (
              <div key={phase.id} className={`phase-item ${phase.status}`}>
                <span className="phase-name">{phase.name}</span>
                <span className="phase-time">
                  {new Date(phase.startTime).toLocaleTimeString([], { 
                    hour: '2-digit', 
                    minute: '2-digit' 
                  })}
                </span>
              </div>
            ))}
          </div>
        </aside>

        {/* Main Panel */}
        <main className="routes-main">
          {activeView === 'planning' && selectedRoute && (
            <div className="planning-view">
              {/* Route Details Header */}
              <div className="route-details-header">
                <div className="details-left">
                  <h2 className="route-title">
                    {getRouteIcon(selectedRoute.type)} {selectedRoute.name}
                  </h2>
                  <div className="route-meta">
                    <span className="meta-item">
                      Status: <strong>{selectedRoute.status}</strong>
                    </span>
                    <span className="meta-item">
                      Threat: <strong style={{ color: getThreatColor(selectedRoute.threatLevel) }}>
                        {selectedRoute.threatLevel}
                      </strong>
                    </span>
                    <span className="meta-item">
                      Terrain: <strong>{selectedRoute.terrainType.join(', ') || 'Unknown'}</strong>
                    </span>
                  </div>
                </div>
                <div className="details-actions">
                  {isDrawing ? (
                    <>
                      <button className="btn-secondary" onClick={() => setIsDrawing(false)}>
                        Cancel Drawing
                      </button>
                      <button className="btn-primary">
                        Finish Route
                      </button>
                    </>
                  ) : (
                    <>
                      <button className="btn-secondary">
                        Edit Route
                      </button>
                      <button className="btn-primary">
                        Activate Route
                      </button>
                    </>
                  )}
                </div>
              </div>

              {/* Waypoints List */}
              <div className="waypoints-section">
                <div className="section-header">
                  <h3>Waypoints</h3>
                  {isDrawing && (
                    <button className="add-waypoint-btn">
                      + Add Waypoint
                    </button>
                  )}
                </div>
                <div className="waypoints-list">
                  {selectedRoute.waypoints.map((waypoint, index) => (
                    <div key={waypoint.id} className="waypoint-item">
                      <div className="waypoint-number">{index + 1}</div>
                      <div className="waypoint-icon">{getWaypointIcon(waypoint.type)}</div>
                      <div className="waypoint-info">
                        <div className="waypoint-name">{waypoint.name}</div>
                        <div className="waypoint-coords">
                          {waypoint.position.lat.toFixed(4)}, {waypoint.position.lng.toFixed(4)}
                          {waypoint.position.alt && ` • ${waypoint.position.alt}m`}
                        </div>
                      </div>
                      {index < selectedRoute.waypoints.length - 1 && (
                        <div className="waypoint-distance">
                          {/* Distance to next waypoint */}
                          ~2.3km
                        </div>
                      )}
                      {waypoint.threats && waypoint.threats.length > 0 && (
                        <div className="threat-badge" title={waypoint.threats.join(', ')}>
                          ⚠️
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              {/* Assigned Units */}
              <div className="units-section">
                <div className="section-header">
                  <h3>Assigned Units</h3>
                  <button className="add-unit-btn">+ Assign Unit</button>
                </div>
                <div className="units-grid">
                  {selectedRoute.assignedUnits.map(unit => (
                    <div key={unit} className="unit-badge">
                      <span className="unit-icon">👤</span>
                      <span className="unit-name">{unit}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Route Analysis */}
              <div className="analysis-section">
                <h3>Route Analysis</h3>
                <div className="analysis-grid">
                  <div className="analysis-item">
                    <span className="label">Total Distance</span>
                    <span className="value">{formatDistance(selectedRoute.distance)}</span>
                  </div>
                  <div className="analysis-item">
                    <span className="label">Estimated Time</span>
                    <span className="value">{formatTime(selectedRoute.estimatedTime)}</span>
                  </div>
                  <div className="analysis-item">
                    <span className="label">Avg Speed</span>
                    <span className="value">
                      {((selectedRoute.distance / 1000) / (selectedRoute.estimatedTime / 60)).toFixed(1)} km/h
                    </span>
                  </div>
                  <div className="analysis-item">
                    <span className="label">Checkpoints</span>
                    <span className="value">
                      {selectedRoute.waypoints.filter(w => w.type === 'checkpoint').length}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeView === 'navigation' && (
            <div className="navigation-view">
              <div className="nav-header">
                <h2>Live Navigation</h2>
                <div className="nav-controls">
                  <button className="nav-btn">Start Navigation</button>
                  <button className="nav-btn">Share Route</button>
                  <button className="nav-btn">Export GPX</button>
                </div>
              </div>
              
              {/* Navigation Map Placeholder */}
              <div className="nav-map-container">
                <div className="map-placeholder">
                  <p>Interactive Map View</p>
                  <p className="hint">Map integration with route overlay</p>
                </div>
              </div>

              {/* Navigation Instructions */}
              <div className="nav-instructions">
                <h3>Turn-by-Turn Instructions</h3>
                <div className="instructions-list">
                  <div className="instruction-item">
                    <span className="instruction-icon">➡️</span>
                    <span className="instruction-text">Head north on Main St</span>
                    <span className="instruction-distance">500m</span>
                  </div>
                  <div className="instruction-item">
                    <span className="instruction-icon">↩️</span>
                    <span className="instruction-text">Turn left onto Oak Ave</span>
                    <span className="instruction-distance">1.2km</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeView === 'analysis' && (
            <div className="analysis-view">
              <h2>Route Analysis & Optimization</h2>
              
              {/* Analysis Tools */}
              <div className="analysis-tools">
                <div className="tool-card">
                  <h3>🎯 Route Comparison</h3>
                  <p>Compare multiple routes side-by-side</p>
                  <button className="tool-btn">Compare Routes</button>
                </div>
                <div className="tool-card">
                  <h3>⚡ Optimization</h3>
                  <p>Find the optimal path based on criteria</p>
                  <button className="tool-btn">Optimize</button>
                </div>
                <div className="tool-card">
                  <h3>⚠️ Threat Assessment</h3>
                  <p>Analyze threats along the route</p>
                  <button className="tool-btn">Assess Threats</button>
                </div>
                <div className="tool-card">
                  <h3>🌤️ Weather Impact</h3>
                  <p>Check weather conditions along route</p>
                  <button className="tool-btn">Check Weather</button>
                </div>
              </div>

              {/* Route Metrics */}
              <div className="metrics-section">
                <h3>Performance Metrics</h3>
                <div className="metrics-grid">
                  <div className="metric-card">
                    <span className="metric-label">Success Rate</span>
                    <span className="metric-value">92%</span>
                  </div>
                  <div className="metric-card">
                    <span className="metric-label">Avg Completion</span>
                    <span className="metric-value">43 min</span>
                  </div>
                  <div className="metric-card">
                    <span className="metric-label">Risk Score</span>
                    <span className="metric-value">3.2/10</span>
                  </div>
                  <div className="metric-card">
                    <span className="metric-label">Fuel Efficiency</span>
                    <span className="metric-value">87%</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>

      {/* New Route Dialog */}
      {showNewRouteDialog && (
        <div className="modal-overlay" onClick={() => setShowNewRouteDialog(false)}>
          <div className="modal-dialog" onClick={e => e.stopPropagation()}>
            <h3>Create New Route</h3>
            <div className="form-group">
              <label>Route Name</label>
              <input
                type="text"
                value={newRoute.name}
                onChange={(e) => setNewRoute({ ...newRoute, name: e.target.value })}
                placeholder="Enter route name..."
                autoFocus
              />
            </div>
            <div className="form-group">
              <label>Route Type</label>
              <select 
                value={newRoute.type}
                onChange={(e) => setNewRoute({ ...newRoute, type: e.target.value as Route['type'] })}
              >
                <option value="primary">Primary</option>
                <option value="alternate">Alternate</option>
                <option value="emergency">Emergency</option>
                <option value="patrol">Patrol</option>
              </select>
            </div>
            <div className="form-group">
              <label>Threat Level</label>
              <select 
                value={newRoute.threatLevel}
                onChange={(e) => setNewRoute({ ...newRoute, threatLevel: e.target.value as Route['threatLevel'] })}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <div className="modal-actions">
              <button 
                className="btn-cancel"
                onClick={() => setShowNewRouteDialog(false)}
              >
                Cancel
              </button>
              <button 
                className="btn-create"
                onClick={handleCreateRoute}
                disabled={!newRoute.name}
              >
                Create & Draw
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Import Dialog */}
      {showImportDialog && (
        <div className="modal-overlay" onClick={() => setShowImportDialog(false)}>
          <div className="modal-dialog" onClick={e => e.stopPropagation()}>
            <h3>Import Route</h3>
            <div className="import-options">
              <button className="import-option">
                <span className="option-icon">📁</span>
                <span className="option-text">GPX File</span>
              </button>
              <button className="import-option">
                <span className="option-icon">📋</span>
                <span className="option-text">KML/KMZ</span>
              </button>
              <button className="import-option">
                <span className="option-icon">🗺️</span>
                <span className="option-text">From Map</span>
              </button>
              <button className="import-option">
                <span className="option-icon">📡</span>
                <span className="option-text">From TAK Server</span>
              </button>
            </div>
            <button 
              className="btn-cancel"
              onClick={() => setShowImportDialog(false)}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Routes;
