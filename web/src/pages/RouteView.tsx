/**
 * Route View Page - Detailed Operational Route Display
 * Full-featured route analysis and management interface
 */

import React, { useState, useEffect } from 'react';
import { useRouter } from '../utils/router';
import { Icon } from '../components/ui/Icon';
import './RouteView.css';

interface Waypoint {
  id: string;
  name: string;
  position: { lat: number; lng: number; alt?: number };
  type: 'start' | 'waypoint' | 'checkpoint' | 'objective' | 'extraction';
  eta?: string;
  notes?: string;
  threats?: string[];
  terrain?: string;
  visibility?: 'good' | 'limited' | 'poor';
}

interface Route {
  id: string;
  name: string;
  type: 'primary' | 'alternate' | 'emergency';
  status: 'planned' | 'active' | 'completed';
  waypoints: Waypoint[];
  distance: number;
  estimatedTime: number;
  threatLevel: 'low' | 'medium' | 'high' | 'critical';
  assignedUnits: string[];
  terrain: string[];
  weather?: {
    condition: string;
    visibility: string;
    wind: string;
  };
  created: string;
  lastModified: string;
  creator: string;
}

const RouteView: React.FC = () => {
  const router = useRouter();
  const id = router.params.id;
  const [route, setRoute] = useState<Route | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'waypoints' | 'threats' | 'resources'>('overview');
  const [selectedWaypoint, setSelectedWaypoint] = useState<Waypoint | null>(null);

  useEffect(() => {
    // Mock loading route data
    const mockRoute: Route = {
      id: id || 'route-1',
      name: 'Alpha Approach',
      type: 'primary',
      status: 'active',
      waypoints: [
        { 
          id: 'wp-1', 
          name: 'Base - Start Point', 
          position: { lat: 38.8951, lng: -77.0364, alt: 120 }, 
          type: 'start',
          notes: 'Secure staging area with full equipment check',
          terrain: 'Urban',
          visibility: 'good'
        },
        { 
          id: 'wp-2', 
          name: 'Checkpoint Alpha', 
          position: { lat: 38.9051, lng: -77.0464, alt: 135 }, 
          type: 'checkpoint',
          eta: '00:15:00',
          notes: 'Radio check required. Alternative route available if compromised.',
          threats: ['Possible surveillance', 'High traffic area'],
          terrain: 'Urban/Residential',
          visibility: 'limited'
        },
        { 
          id: 'wp-3', 
          name: 'Rally Point Bravo', 
          position: { lat: 38.9101, lng: -77.0514, alt: 128 }, 
          type: 'waypoint',
          eta: '00:25:00',
          notes: 'Final equipment check before objective',
          terrain: 'Mixed',
          visibility: 'good'
        },
        { 
          id: 'wp-4', 
          name: 'Objective Sierra', 
          position: { lat: 38.9151, lng: -77.0564, alt: 142 }, 
          type: 'objective',
          eta: '00:35:00',
          notes: 'Primary objective. Multiple ingress/egress routes identified.',
          threats: ['Active security', 'Electronic surveillance'],
          terrain: 'Urban/Industrial',
          visibility: 'poor'
        },
        { 
          id: 'wp-5', 
          name: 'Extract Point', 
          position: { lat: 38.9201, lng: -77.0614, alt: 125 }, 
          type: 'extraction',
          eta: '00:45:00',
          notes: 'Primary extraction. Alt extract 500m west if needed.',
          terrain: 'Open',
          visibility: 'good'
        },
      ],
      distance: 12500,
      estimatedTime: 45,
      threatLevel: 'medium',
      assignedUnits: ['ALPHA-1', 'ALPHA-2', 'OVERWATCH-1'],
      terrain: ['Urban', 'Residential', 'Industrial'],
      weather: {
        condition: 'Clear',
        visibility: '10km',
        wind: 'NW 5mph'
      },
      created: '2024-01-15T14:00:00Z',
      lastModified: '2024-01-15T16:30:00Z',
      creator: 'OPS-COMMANDER'
    };
    setRoute(mockRoute);
    setSelectedWaypoint(mockRoute.waypoints[0]);
  }, [id]);

  if (!route) {
    return <div className="loading">Loading route...</div>;
  }

  const formatDistance = (meters: number) => {
    if (meters < 1000) return `${meters}m`;
    return `${(meters / 1000).toFixed(1)}km`;
  };

  const formatTime = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return hours > 0 ? `${hours}h ${mins}m` : `${mins}min`;
  };

  const getThreatColor = (level: string) => {
    switch (level) {
      case 'low': return '#4caf50';
      case 'medium': return '#ff9800';
      case 'high': return '#ff5722';
      case 'critical': return '#f44336';
      default: return '#888';
    }
  };

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
    <div className="route-view-container">
      {/* Header */}
      <header className="route-view-header">
        <div className="header-nav">
          <button className="back-btn" onClick={() => router.navigate('/routes')}>
            <Icon name="route" size={20} />
            Back to Routes
          </button>
          
          <div className="route-status">
            <span className={`status-indicator ${route.status}`} />
            {route.status.toUpperCase()}
          </div>
        </div>

        <div className="route-title-section">
          <h1 className="route-title">{route.name}</h1>
          <div className="route-meta">
            <span className="meta-item">
              <Icon name="user" size={14} />
              {route.creator}
            </span>
            <span className="meta-item">
              <Icon name="speed" size={14} />
              Modified {new Date(route.lastModified).toLocaleString()}
            </span>
          </div>
        </div>

        <div className="header-actions">
          <button className="btn-secondary">
            <Icon name="wrench" size={16} />
            Edit
          </button>
          <button className="btn-secondary">
            <Icon name="users" size={16} />
            Share
          </button>
          <button className="btn-primary" onClick={() => router.navigate('/routes/navigate')}>
            <Icon name="send" size={16} />
            Start Navigation
          </button>
        </div>
      </header>

      {/* Tabs */}
      <div className="view-tabs">
        <button 
          className={`tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button 
          className={`tab ${activeTab === 'waypoints' ? 'active' : ''}`}
          onClick={() => setActiveTab('waypoints')}
        >
          Waypoints
        </button>
        <button 
          className={`tab ${activeTab === 'threats' ? 'active' : ''}`}
          onClick={() => setActiveTab('threats')}
        >
          Threat Analysis
        </button>
        <button 
          className={`tab ${activeTab === 'resources' ? 'active' : ''}`}
          onClick={() => setActiveTab('resources')}
        >
          Resources
        </button>
      </div>

      {/* Content */}
      <div className="route-view-content">
        {activeTab === 'overview' && (
          <div className="overview-tab">
            {/* Key Metrics */}
            <div className="metrics-grid">
              <div className="metric-card">
                <Icon name="map" size={24} />
                <div className="metric-info">
                  <span className="metric-value">{formatDistance(route.distance)}</span>
                  <span className="metric-label">Total Distance</span>
                </div>
              </div>
              <div className="metric-card">
                <Icon name="speed" size={24} />
                <div className="metric-info">
                  <span className="metric-value">{formatTime(route.estimatedTime)}</span>
                  <span className="metric-label">Est. Duration</span>
                </div>
              </div>
              <div className="metric-card">
                <Icon name="pin" size={24} />
                <div className="metric-info">
                  <span className="metric-value">{route.waypoints.length}</span>
                  <span className="metric-label">Waypoints</span>
                </div>
              </div>
              <div className="metric-card threat">
                <Icon name="alert-triangle" size={24} />
                <div className="metric-info">
                  <span 
                    className="metric-value" 
                    style={{ color: getThreatColor(route.threatLevel) }}
                  >
                    {route.threatLevel.toUpperCase()}
                  </span>
                  <span className="metric-label">Threat Level</span>
                </div>
              </div>
            </div>

            {/* Map Placeholder */}
            <div className="map-section">
              <div className="map-placeholder">
                <Icon name="map" size={48} />
                <h3>Interactive Route Map</h3>
                <p>Live map view with waypoints and terrain</p>
              </div>
            </div>

            {/* Route Info Grid */}
            <div className="info-grid">
              <div className="info-card">
                <h3>
                  <Icon name="users" size={18} />
                  Assigned Units
                </h3>
                <div className="units-list">
                  {route.assignedUnits.map(unit => (
                    <div key={unit} className="unit-item">
                      <Icon name="user" size={14} />
                      {unit}
                    </div>
                  ))}
                </div>
              </div>

              <div className="info-card">
                <h3>
                  <Icon name="world" size={18} />
                  Terrain Types
                </h3>
                <div className="terrain-list">
                  {route.terrain.map(terrain => (
                    <span key={terrain} className="terrain-tag">{terrain}</span>
                  ))}
                </div>
              </div>

              <div className="info-card">
                <h3>
                  <Icon name="sparkle" size={18} />
                  Weather Conditions
                </h3>
                {route.weather && (
                  <div className="weather-info">
                    <div className="weather-item">
                      <span className="label">Condition:</span>
                      <span className="value">{route.weather.condition}</span>
                    </div>
                    <div className="weather-item">
                      <span className="label">Visibility:</span>
                      <span className="value">{route.weather.visibility}</span>
                    </div>
                    <div className="weather-item">
                      <span className="label">Wind:</span>
                      <span className="value">{route.weather.wind}</span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'waypoints' && (
          <div className="waypoints-tab">
            <div className="waypoints-grid">
              <div className="waypoints-list">
                {route.waypoints.map((waypoint, index) => (
                  <div 
                    key={waypoint.id}
                    className={`waypoint-card ${selectedWaypoint?.id === waypoint.id ? 'selected' : ''}`}
                    onClick={() => setSelectedWaypoint(waypoint)}
                  >
                    <div className="waypoint-header">
                      <div className="waypoint-number">{index + 1}</div>
                      <Icon name={getWaypointIcon(waypoint.type)} size={20} />
                      <div className="waypoint-info">
                        <h4>{waypoint.name}</h4>
                        <span className="waypoint-type">{waypoint.type}</span>
                      </div>
                      {waypoint.eta && (
                        <span className="waypoint-eta">ETA: {waypoint.eta}</span>
                      )}
                    </div>
                    {waypoint.threats && waypoint.threats.length > 0 && (
                      <div className="waypoint-threats">
                        <Icon name="warning" size={14} />
                        {waypoint.threats.length} threat{waypoint.threats.length > 1 ? 's' : ''}
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {selectedWaypoint && (
                <div className="waypoint-details">
                  <h3>Waypoint Details</h3>
                  
                  <div className="detail-section">
                    <h4>Location</h4>
                    <div className="coords">
                      <span>Lat: {selectedWaypoint.position.lat.toFixed(6)}</span>
                      <span>Lng: {selectedWaypoint.position.lng.toFixed(6)}</span>
                      {selectedWaypoint.position.alt && (
                        <span>Alt: {selectedWaypoint.position.alt}m</span>
                      )}
                    </div>
                  </div>

                  {selectedWaypoint.terrain && (
                    <div className="detail-section">
                      <h4>Terrain</h4>
                      <p>{selectedWaypoint.terrain}</p>
                    </div>
                  )}

                  {selectedWaypoint.visibility && (
                    <div className="detail-section">
                      <h4>Visibility</h4>
                      <span className={`visibility-badge ${selectedWaypoint.visibility}`}>
                        {selectedWaypoint.visibility}
                      </span>
                    </div>
                  )}

                  {selectedWaypoint.notes && (
                    <div className="detail-section">
                      <h4>Operational Notes</h4>
                      <p>{selectedWaypoint.notes}</p>
                    </div>
                  )}

                  {selectedWaypoint.threats && selectedWaypoint.threats.length > 0 && (
                    <div className="detail-section">
                      <h4>Identified Threats</h4>
                      <ul className="threats-list">
                        {selectedWaypoint.threats.map((threat, idx) => (
                          <li key={idx}>
                            <Icon name="alert-circle" size={14} />
                            {threat}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'threats' && (
          <div className="threats-tab">
            <div className="threat-summary">
              <div className="threat-level-card" style={{ borderColor: getThreatColor(route.threatLevel) }}>
                <h3>Overall Threat Assessment</h3>
                <div className="threat-level-display">
                  <Icon name="alert-triangle" size={32} />
                  <span className="level" style={{ color: getThreatColor(route.threatLevel) }}>
                    {route.threatLevel.toUpperCase()}
                  </span>
                </div>
              </div>
            </div>

            <div className="threats-breakdown">
              <h3>Threat Analysis by Waypoint</h3>
              {route.waypoints.filter(wp => wp.threats && wp.threats.length > 0).map(waypoint => (
                <div key={waypoint.id} className="threat-waypoint">
                  <h4>
                    <Icon name={getWaypointIcon(waypoint.type)} size={16} />
                    {waypoint.name}
                  </h4>
                  <ul>
                    {waypoint.threats?.map((threat, idx) => (
                      <li key={idx}>{threat}</li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>

            <div className="mitigation-section">
              <h3>Recommended Mitigations</h3>
              <ul>
                <li>Maintain radio silence between checkpoints</li>
                <li>Use alternate route if primary is compromised</li>
                <li>Ensure overwatch coverage at high-threat waypoints</li>
                <li>Conduct final equipment check at Rally Point Bravo</li>
              </ul>
            </div>
          </div>
        )}

        {activeTab === 'resources' && (
          <div className="resources-tab">
            <div className="resources-grid">
              <div className="resource-card">
                <h3>
                  <Icon name="users" size={18} />
                  Personnel
                </h3>
                <ul>
                  <li>Alpha Team: 6 operators</li>
                  <li>Overwatch: 2 snipers</li>
                  <li>QRF: 4 operators (on standby)</li>
                  <li>Command: 2 coordinators</li>
                </ul>
              </div>

              <div className="resource-card">
                <h3>
                  <Icon name="vehicle" size={18} />
                  Vehicles
                </h3>
                <ul>
                  <li>2x Tactical vehicles</li>
                  <li>1x Support vehicle</li>
                  <li>1x Medical evacuation (on standby)</li>
                </ul>
              </div>

              <div className="resource-card">
                <h3>
                  <Icon name="equipment" size={18} />
                  Equipment
                </h3>
                <ul>
                  <li>Standard tactical loadout</li>
                  <li>Night vision equipment</li>
                  <li>Breaching tools</li>
                  <li>Emergency medical supplies</li>
                </ul>
              </div>

              <div className="resource-card">
                <h3>
                  <Icon name="broadcast" size={18} />
                  Communications
                </h3>
                <ul>
                  <li>Primary: Encrypted radio (Ch. 7)</li>
                  <li>Secondary: Satellite phone</li>
                  <li>Emergency: Signal flares</li>
                  <li>Data: TAK server sync</li>
                </ul>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default RouteView;