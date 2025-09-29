/**
 * Routes Page - Simplified Mission Planning & Navigation
 * Clean and modern route planning interface
 */

import React, { useState, useEffect } from 'react';
import { useRouter } from '../utils/router';
import { Icon } from '../components/ui/Icon';
import './Routes-new.css';

// Types
interface Waypoint {
  id: string;
  name: string;
  position: { lat: number; lng: number; alt?: number };
  type: 'start' | 'waypoint' | 'checkpoint' | 'objective' | 'extraction';
  eta?: string;
  notes?: string;
}

interface Route {
  id: string;
  name: string;
  type: 'primary' | 'alternate' | 'emergency';
  status: 'planned' | 'active' | 'completed';
  waypoints: Waypoint[];
  distance: number; // in meters
  estimatedTime: number; // in minutes
  threatLevel: 'low' | 'medium' | 'high' | 'critical';
  assignedUnits: string[];
}

const Routes: React.FC = () => {
  const router = useRouter();
  // State
  const [routes, setRoutes] = useState<Route[]>([]);
  const [selectedRoute, setSelectedRoute] = useState<Route | null>(null);
  const [activeView, setActiveView] = useState<'list' | 'map'>('list');
  const [showNewRouteDialog, setShowNewRouteDialog] = useState(false);
  
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
        status: 'active',
        waypoints: [
          { id: 'wp-1', name: 'Base', position: { lat: 38.8951, lng: -77.0364 }, type: 'start' },
          { id: 'wp-2', name: 'Checkpoint Alpha', position: { lat: 38.9051, lng: -77.0464 }, type: 'checkpoint' },
          { id: 'wp-3', name: 'Objective', position: { lat: 38.9151, lng: -77.0564 }, type: 'objective' },
        ],
        distance: 12500,
        estimatedTime: 45,
        threatLevel: 'medium',
        assignedUnits: ['ALPHA-1', 'ALPHA-2'],
      },
      {
        id: 'route-2',
        name: 'Bravo Alternate',
        type: 'alternate',
        status: 'planned',
        waypoints: [
          { id: 'wp-4', name: 'Base', position: { lat: 38.8951, lng: -77.0364 }, type: 'start' },
          { id: 'wp-5', name: 'Rally Point', position: { lat: 38.8851, lng: -77.0264 }, type: 'waypoint' },
          { id: 'wp-6', name: 'Objective', position: { lat: 38.9151, lng: -77.0564 }, type: 'objective' },
        ],
        distance: 15200,
        estimatedTime: 55,
        threatLevel: 'low',
        assignedUnits: ['BRAVO-1'],
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
        threatLevel: 'critical',
        assignedUnits: ['QRF-1'],
      },
    ];

    setRoutes(sampleRoutes);
    setSelectedRoute(sampleRoutes[0]);
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
      threatLevel: newRoute.threatLevel,
      assignedUnits: [],
    };

    setRoutes([...routes, route]);
    setSelectedRoute(route);
    setShowNewRouteDialog(false);
    setNewRoute({ name: '', type: 'primary', threatLevel: 'low' });
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
  const getRouteIcon = (type: Route['type']): any => {
    switch (type) {
      case 'primary': return 'route';
      case 'alternate': return 'sync';
      case 'emergency': return 'alert-triangle';
    }
  };

  // Get waypoint icon
  const getWaypointIcon = (type: Waypoint['type']): any => {
    switch (type) {
      case 'start': return 'target';
      case 'waypoint': return 'pin';
      case 'checkpoint': return 'check-circle';
      case 'objective': return 'target';
      case 'extraction': return 'rocket';
    }
  };

  return (
    <div className="routes-new-container">
      {/* Header */}
      <header className="routes-header">
        <div className="header-content">
          <h1 className="page-title">
            <Icon name="route" size={24} />
            Mission Routes
          </h1>
          
          <div className="header-actions">
            {/* View Toggle */}
            <div className="view-toggle">
              <button 
                className={`toggle-btn ${activeView === 'list' ? 'active' : ''}`}
                onClick={() => setActiveView('list')}
              >
                <Icon name="list" size={18} />
              </button>
              <button 
                className={`toggle-btn ${activeView === 'map' ? 'active' : ''}`}
                onClick={() => setActiveView('map')}
              >
                <Icon name="map" size={18} />
              </button>
            </div>

            {/* Add Route Button */}
            <button 
              className="btn-primary"
              onClick={() => router.navigate('/routes/builder')}
            >
              <Icon name="sparkle" size={18} />
              New Route
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="routes-content">
        {activeView === 'list' ? (
          <div className="routes-grid">
            {routes.map(route => (
              <div 
                key={route.id}
                className={`route-card ${selectedRoute?.id === route.id ? 'selected' : ''} ${route.status}`}
                onClick={() => setSelectedRoute(route)}
              >
                {/* Route Header */}
                <div className="route-header">
                  <Icon name={getRouteIcon(route.type)} size={24} />
                  <div className="route-info">
                    <h3 className="route-name">{route.name}</h3>
                    <span className="route-type">{route.type}</span>
                  </div>
                  {route.status === 'active' && (
                    <span className="status-badge active">
                      <Icon name="broadcast" size={14} />
                      Active
                    </span>
                  )}
                </div>

                {/* Route Stats */}
                <div className="route-stats">
                  <div className="stat">
                    <Icon name="map" size={16} />
                    <span>{formatDistance(route.distance)}</span>
                  </div>
                  <div className="stat">
                    <Icon name="speed" size={16} />
                    <span>{formatTime(route.estimatedTime)}</span>
                  </div>
                  <div className="stat">
                    <Icon name="pin" size={16} />
                    <span>{route.waypoints.length} points</span>
                  </div>
                </div>

                {/* Threat Level */}
                <div className="threat-level">
                  <span className="threat-label">Threat Level:</span>
                  <div className="threat-indicator" style={{ backgroundColor: getThreatColor(route.threatLevel) }}>
                    {route.threatLevel}
                  </div>
                </div>

                {/* Waypoints Preview */}
                <div className="waypoints-preview">
                  {route.waypoints.map((wp, idx) => (
                    <div key={wp.id} className="waypoint-item">
                      <Icon name={getWaypointIcon(wp.type)} size={14} />
                      <span className="waypoint-name">{wp.name}</span>
                      {idx < route.waypoints.length - 1 && (
                        <Icon name="link" size={12} className="waypoint-connector" />
                      )}
                    </div>
                  ))}
                </div>

                {/* Assigned Units */}
                {route.assignedUnits.length > 0 && (
                  <div className="assigned-units">
                    <Icon name="users" size={14} />
                    <span>{route.assignedUnits.join(', ')}</span>
                  </div>
                )}

                {/* Actions */}
                <div className="route-actions">
                  <button className="action-btn" onClick={(e) => {
                    e.stopPropagation();
                    router.navigate(`/routes/view`);
                  }}>
                    <Icon name="eye" size={16} />
                    View
                  </button>
                  <button className="action-btn" onClick={(e) => {
                    e.stopPropagation();
                    router.navigate(`/routes/navigate`);
                  }}>
                    <Icon name="send" size={16} />
                    Navigate
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="map-view">
            <div className="map-placeholder">
              <Icon name="map" size={48} />
              <h3>Interactive Map View</h3>
              <p>Map integration will display all routes and waypoints</p>
            </div>
          </div>
        )}
      </div>

      {/* New Route Dialog */}
      {showNewRouteDialog && (
        <div className="modal-overlay" onClick={() => setShowNewRouteDialog(false)}>
          <div className="modal-dialog" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3>
                <Icon name="route" size={20} />
                Create New Route
              </h3>
              <button className="close-btn" onClick={() => setShowNewRouteDialog(false)}>
                <Icon name="x" size={20} />
              </button>
            </div>
            
            <div className="modal-body">
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
              
              <div className="form-row">
                <div className="form-group">
                  <label>Route Type</label>
                  <select 
                    value={newRoute.type}
                    onChange={(e) => setNewRoute({ ...newRoute, type: e.target.value as Route['type'] })}
                  >
                    <option value="primary">Primary</option>
                    <option value="alternate">Alternate</option>
                    <option value="emergency">Emergency</option>
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
              </div>
            </div>
            
            <div className="modal-footer">
              <button 
                className="btn-secondary"
                onClick={() => setShowNewRouteDialog(false)}
              >
                Cancel
              </button>
              <button 
                className="btn-primary"
                onClick={handleCreateRoute}
                disabled={!newRoute.name}
              >
                <Icon name="check" size={16} />
                Create Route
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Routes;