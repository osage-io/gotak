/**
 * Entities Page - Redesigned
 * Full-width tactical entity management interface
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { wsService } from '../services/websocketService';

interface Entity {
  id: string;
  callsign: string;
  type: 'friendly' | 'hostile' | 'neutral' | 'unknown' | 'drone' | 'sensor' | 'camera' | 'vehicle' | 'equipment';
  subType?: 'uav' | 'ugv' | 'thermal' | 'motion' | 'acoustic' | 'radiation' | 'cbrn' | 'weather' | 'surveillance' | 'tactical';
  status: 'active' | 'inactive' | 'stale' | 'lost' | 'standby' | 'maintenance';
  team: string;
  role: string;
  position: {
    lat: number;
    lng: number;
    alt?: number;
    speed?: number;
    heading?: number;
  };
  lastUpdate: string;
  battery?: number;
  signal?: number;
  temperature?: number;
  fuelLevel?: number;
  operatingTime?: number;
  sensorData?: {
    type: string;
    value: any;
    unit: string;
    timestamp: string;
  }[];
  equipment?: string[];
  capabilities?: string[];
  notes?: string;
}

type EntityView = 'grid' | 'list' | 'tactical';
type EntityFilter = 'all' | 'personnel' | 'drones' | 'sensors' | 'vehicles' | 'friendly' | 'hostile' | 'neutral' | 'unknown';

const Entities: React.FC = () => {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedEntity, setSelectedEntity] = useState<string | null>(null);
  const [view, setView] = useState<EntityView>('list'); // Changed default to 'list'
  const [filter, setFilter] = useState<EntityFilter>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showOffline, setShowOffline] = useState(true);

  // Entity filter categories
  const categories = [
    { id: 'all', label: 'All Entities' },
    { id: 'personnel', label: 'Personnel' },
    { id: 'drones', label: 'Drones' },
    { id: 'sensors', label: 'Sensors' },
    { id: 'vehicles', label: 'Vehicles' },
    { id: 'cameras', label: 'Cameras' },
    { id: 'friendly', label: 'Friendly' },
    { id: 'hostile', label: 'Hostile' },
    { id: 'neutral', label: 'Neutral' },
    { id: 'unknown', label: 'Unknown' },
  ] as const;

  // Initialize with mock data
  useEffect(() => {
    const mockEntities: Entity[] = [
      {
        id: '1',
        callsign: 'ALPHA-1',
        type: 'friendly',
        status: 'active',
        team: 'Blue Force',
        role: 'Squad Leader',
        position: {
          lat: 38.8951,
          lng: -77.0364,
          alt: 125,
          speed: 0,
          heading: 45
        },
        lastUpdate: new Date().toISOString(),
        battery: 85,
        signal: 92,
        equipment: ['Radio', 'NVG', 'GPS', 'Medical Kit'],
        capabilities: ['Command', 'Medical', 'Navigation'],
      },
      {
        id: '2',
        callsign: 'BRAVO-2',
        type: 'friendly',
        status: 'active',
        team: 'Blue Force',
        role: 'Rifleman',
        position: {
          lat: 38.8977,
          lng: -77.0365,
          alt: 130,
          speed: 5.2,
          heading: 90
        },
        lastUpdate: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
        battery: 67,
        signal: 88,
        equipment: ['Radio', 'GPS'],
        capabilities: ['Reconnaissance'],
      },
      {
        id: '3',
        callsign: 'CHARLIE-3',
        type: 'friendly',
        status: 'stale',
        team: 'Blue Force',
        role: 'Medic',
        position: {
          lat: 38.8923,
          lng: -77.0389,
          alt: 110,
          speed: 0,
          heading: 180
        },
        lastUpdate: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        battery: 45,
        signal: 65,
        equipment: ['Radio', 'Medical Supplies', 'GPS'],
        capabilities: ['Medical', 'CASEVAC'],
      },
      {
        id: '4',
        callsign: 'HOSTILE-X1',
        type: 'hostile',
        status: 'active',
        team: 'Red Force',
        role: 'Unknown',
        position: {
          lat: 38.9012,
          lng: -77.0423,
          alt: 150,
          speed: 8.3,
          heading: 270
        },
        lastUpdate: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        equipment: ['Unknown'],
        capabilities: ['Armed'],
      },
      {
        id: '5',
        callsign: 'NEUTRAL-N1',
        type: 'neutral',
        status: 'active',
        team: 'Civilian',
        role: 'Observer',
        position: {
          lat: 38.8889,
          lng: -77.0298,
          alt: 100,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 95,
        signal: 100,
        equipment: ['Radio', 'Camera'],
        capabilities: ['Observation'],
      },
      {
        id: '6',
        callsign: 'DELTA-4',
        type: 'friendly',
        status: 'lost',
        team: 'Blue Force',
        role: 'Scout',
        position: {
          lat: 38.9072,
          lng: -77.0369,
          alt: 140,
          speed: 0,
          heading: 315
        },
        lastUpdate: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
        battery: 12,
        signal: 0,
        equipment: ['Radio', 'NVG', 'GPS'],
        capabilities: ['Reconnaissance', 'Sniper'],
        notes: 'Last contact 45 minutes ago. Possible equipment failure.',
      },
      // Drone entities
      {
        id: '7',
        callsign: 'EAGLE-EYE-1',
        type: 'drone',
        subType: 'uav',
        status: 'active',
        team: 'Air Assets',
        role: 'ISR Platform',
        position: {
          lat: 38.9021,
          lng: -77.0367,
          alt: 450,
          speed: 25.5,
          heading: 135
        },
        lastUpdate: new Date().toISOString(),
        battery: 78,
        signal: 95,
        operatingTime: 2.5,
        equipment: ['EO/IR Camera', '4K Video', 'Thermal Imaging'],
        capabilities: ['Surveillance', 'Reconnaissance', 'Target Acquisition'],
      },
      {
        id: '8',
        callsign: 'RAVEN-2',
        type: 'drone',
        subType: 'uav',
        status: 'standby',
        team: 'Air Assets',
        role: 'Tactical UAV',
        position: {
          lat: 38.8901,
          lng: -77.0321,
          alt: 0,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
        battery: 100,
        signal: 100,
        operatingTime: 0,
        equipment: ['HD Camera', 'Night Vision'],
        capabilities: ['Quick Deploy', 'Low Altitude Recon'],
      },
      // Sensor entities
      {
        id: '9',
        callsign: 'THERMAL-CAM-01',
        type: 'sensor',
        subType: 'thermal',
        status: 'active',
        team: 'Perimeter Security',
        role: 'Thermal Imaging',
        position: {
          lat: 38.8955,
          lng: -77.0401,
          alt: 15,
          speed: 0,
          heading: 270
        },
        lastUpdate: new Date().toISOString(),
        temperature: 22.5,
        signal: 100,
        sensorData: [
          { type: 'heat_signature', value: 3, unit: 'targets', timestamp: new Date().toISOString() },
          { type: 'temperature_range', value: '18-37', unit: '°C', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Heat Detection', '360° Coverage', 'Auto-Tracking'],
      },
      {
        id: '10',
        callsign: 'MOTION-SENSOR-05',
        type: 'sensor',
        subType: 'motion',
        status: 'active',
        team: 'Perimeter Security',
        role: 'Motion Detection',
        position: {
          lat: 38.8912,
          lng: -77.0378,
          alt: 2,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 92,
        signal: 87,
        sensorData: [
          { type: 'motion_events', value: 2, unit: 'detections/hr', timestamp: new Date().toISOString() },
          { type: 'sensitivity', value: 85, unit: '%', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Motion Detection', 'Vibration Sensing', 'Alert Triggering'],
      },
      // Camera entity
      {
        id: '11',
        callsign: 'OVERWATCH-CAM-3',
        type: 'camera',
        subType: 'surveillance',
        status: 'active',
        team: 'Surveillance',
        role: 'Fixed Camera',
        position: {
          lat: 38.8945,
          lng: -77.0355,
          alt: 25,
          speed: 0,
          heading: 45
        },
        lastUpdate: new Date().toISOString(),
        signal: 100,
        equipment: ['4K Resolution', 'PTZ Control', 'IR Illumination'],
        capabilities: ['24/7 Recording', 'Motion Tracking', 'Face Detection'],
      },
      // Vehicle entity
      {
        id: '12',
        callsign: 'VICTOR-1',
        type: 'vehicle',
        subType: 'tactical',
        status: 'active',
        team: 'Mobile Command',
        role: 'Command Vehicle',
        position: {
          lat: 38.8967,
          lng: -77.0342,
          alt: 105,
          speed: 15.2,
          heading: 180
        },
        lastUpdate: new Date().toISOString(),
        fuelLevel: 68,
        signal: 90,
        temperature: 85,
        equipment: ['Communications Hub', 'Satellite Uplink', 'Command Console'],
        capabilities: ['Mobile Command', 'Communications Relay', 'Power Generation'],
      },
      // CBRN Sensor
      {
        id: '13',
        callsign: 'CBRN-DETECT-1',
        type: 'sensor',
        subType: 'cbrn',
        status: 'active',
        team: 'HAZMAT',
        role: 'CBRN Detection',
        position: {
          lat: 38.8934,
          lng: -77.0398,
          alt: 5,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 88,
        signal: 92,
        sensorData: [
          { type: 'radiation_level', value: 0.12, unit: 'μSv/h', timestamp: new Date().toISOString() },
          { type: 'chemical_agents', value: 'None', unit: 'detection', timestamp: new Date().toISOString() },
          { type: 'biological_threat', value: 'Clear', unit: 'status', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Chemical Detection', 'Radiation Monitoring', 'Biological Agent Detection'],
        notes: 'Continuous monitoring active. All readings within safe parameters.',
      },
    ];

    setEntities(mockEntities);
  }, []);

  // Handle entity selection
  const handleSelectEntity = useCallback((entityId: string) => {
    setSelectedEntity(entityId === selectedEntity ? null : entityId);
  }, [selectedEntity]);

  // Handle entity deletion
  const handleDeleteEntity = useCallback((entityId: string) => {
    if (confirm('Are you sure you want to remove this entity?')) {
      setEntities(prev => prev.filter(e => e.id !== entityId));
      if (selectedEntity === entityId) {
        setSelectedEntity(null);
      }
    }
  }, [selectedEntity]);

  // Filter and search entities
  const filteredEntities = useMemo(() => {
    let filtered = entities;

    // Apply type filter
    if (filter !== 'all') {
      switch (filter) {
        case 'personnel':
          filtered = filtered.filter(e => ['friendly', 'hostile', 'neutral', 'unknown'].includes(e.type));
          break;
        case 'drones':
          filtered = filtered.filter(e => e.type === 'drone');
          break;
        case 'sensors':
          filtered = filtered.filter(e => e.type === 'sensor');
          break;
        case 'vehicles':
          filtered = filtered.filter(e => e.type === 'vehicle');
          break;
        default:
          filtered = filtered.filter(e => e.type === filter);
      }
    }

    // Apply offline filter
    if (!showOffline) {
      filtered = filtered.filter(e => e.status === 'active');
    }

    // Apply search
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(entity =>
        entity.callsign.toLowerCase().includes(query) ||
        entity.team.toLowerCase().includes(query) ||
        entity.role.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [entities, filter, showOffline, searchQuery]);

  // Get counts by type
  const counts = useMemo(() => ({
    all: entities.length,
    personnel: entities.filter(e => ['friendly', 'hostile', 'neutral', 'unknown'].includes(e.type)).length,
    drones: entities.filter(e => e.type === 'drone').length,
    sensors: entities.filter(e => e.type === 'sensor').length,
    vehicles: entities.filter(e => e.type === 'vehicle').length,
    cameras: entities.filter(e => e.type === 'camera').length,
    friendly: entities.filter(e => e.type === 'friendly').length,
    hostile: entities.filter(e => e.type === 'hostile').length,
    neutral: entities.filter(e => e.type === 'neutral').length,
    unknown: entities.filter(e => e.type === 'unknown').length,
    active: entities.filter(e => e.status === 'active').length,
    offline: entities.filter(e => !['active', 'standby'].includes(e.status)).length,
  }), [entities]);

  const selectedEntityData = entities.find(e => e.id === selectedEntity);

  // Get status color
  const getStatusColor = (status: Entity['status']) => {
    switch (status) {
      case 'active': return '#2ed573';
      case 'inactive': return '#ffa502';
      case 'stale': return '#ff6348';
      case 'lost': return '#d32f2f';
      case 'standby': return '#3498db';
      case 'maintenance': return '#9b59b6';
      default: return '#57606f';
    }
  };

  // Get type icon
  const getTypeIcon = (type: Entity['type'], subType?: Entity['subType']) => {
    switch (type) {
      case 'friendly': return '🟢';
      case 'hostile': return '🔴';
      case 'neutral': return '⚪';
      case 'unknown': return '❓';
      case 'drone': return subType === 'ugv' ? '🤖' : '🚁';
      case 'sensor': 
        switch (subType) {
          case 'thermal': return '🌡️';
          case 'motion': return '📡';
          case 'acoustic': return '🎙️';
          case 'radiation': return '☢️';
          case 'cbrn': return '⚠️';
          case 'weather': return '🌤️';
          default: return '📊';
        }
      case 'camera': return '📹';
      case 'vehicle': return '🚗';
      case 'equipment': return '📦';
      default: return '⚫';
    }
  };

  // Format time ago
  const getTimeAgo = (timestamp: string) => {
    const seconds = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  };

  return (
    <div className="entities-fullpage">
      {/* Header */}
      <header className="entities-header">
        <div className="header-title">
          <h1>Entity Management</h1>
          <div className="entity-stats">
            <span className="stat active">{counts.active} Active</span>
            <span className="stat offline">{counts.offline} Offline</span>
            <span className="stat total">{counts.all} Total</span>
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
          </div>

          <div className="filter-group">
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as EntityFilter)}
              className="filter-select"
            >
              {categories.map(category => {
                let count = 0;
                switch (category.id) {
                  case 'all': count = counts.all; break;
                  case 'personnel': count = counts.personnel; break;
                  case 'drones': count = counts.drones; break;
                  case 'sensors': count = counts.sensors; break;
                  case 'vehicles': count = counts.vehicles; break;
                  case 'cameras': count = counts.cameras; break;
                  case 'friendly': count = counts.friendly; break;
                  case 'hostile': count = counts.hostile; break;
                  case 'neutral': count = counts.neutral; break;
                  case 'unknown': count = counts.unknown; break;
                  default: count = 0;
                }
                return (
                  <option key={category.id} value={category.id}>
                    {category.label} ({count})
                  </option>
                );
              })}
            </select>

            <label className="offline-toggle">
              <input
                type="checkbox"
                checked={showOffline}
                onChange={(e) => setShowOffline(e.target.checked)}
              />
              <span>Show Offline</span>
            </label>
          </div>

          <div className="view-toggles">
            <button
              className={`view-btn ${view === 'grid' ? 'active' : ''}`}
              onClick={() => setView('grid')}
              title="Grid View"
            >
              ⊞
            </button>
            <button
              className={`view-btn ${view === 'list' ? 'active' : ''}`}
              onClick={() => setView('list')}
              title="List View"
            >
              ☰
            </button>
            <button
              className={`view-btn ${view === 'tactical' ? 'active' : ''}`}
              onClick={() => setView('tactical')}
              title="Tactical View"
            >
              ⊕
            </button>
          </div>

          <button className="btn-primary">
            + Add Entity
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div className="entities-content">
        {/* Entities Display */}
        <div className={`entities-display view-${view}`}>
          {filteredEntities.length === 0 ? (
            <div className="no-entities">
              <span className="no-entities-icon">📡</span>
              <h3>No entities found</h3>
              <p>{searchQuery ? `No results for "${searchQuery}"` : 'No entities match the current filter'}</p>
            </div>
          ) : view === 'grid' ? (
            <div className="entities-grid">
              {filteredEntities.map(entity => (
                <div
                  key={entity.id}
                  className={`entity-card ${selectedEntity === entity.id ? 'selected' : ''} ${entity.type}`}
                  onClick={() => handleSelectEntity(entity.id)}
                >
                  <div className="card-header">
                    <span className="entity-icon">{getTypeIcon(entity.type, entity.subType)}</span>
                    <h3 className="entity-callsign">{entity.callsign}</h3>
                    <span 
                      className="status-indicator"
                      style={{ backgroundColor: getStatusColor(entity.status) }}
                      title={entity.status}
                    />
                  </div>

                  <div className="card-content">
                    <div className="entity-info">
                      <span className="info-label">Team:</span>
                      <span className="info-value">{entity.team}</span>
                    </div>
                    <div className="entity-info">
                      <span className="info-label">Role:</span>
                      <span className="info-value">{entity.role}</span>
                    </div>
                    <div className="entity-info">
                      <span className="info-label">Position:</span>
                      <span className="info-value">
                        {entity.position.lat.toFixed(4)}, {entity.position.lng.toFixed(4)}
                      </span>
                    </div>
                    <div className="entity-info">
                      <span className="info-label">Last Update:</span>
                      <span className="info-value">{getTimeAgo(entity.lastUpdate)}</span>
                    </div>
                  </div>

                  <div className="card-footer">
                    {entity.battery !== undefined && (
                      <div className="metric">
                        <span className="metric-icon">🔋</span>
                        <span className="metric-value">{entity.battery}%</span>
                      </div>
                    )}
                    {entity.signal !== undefined && (
                      <div className="metric">
                        <span className="metric-icon">📶</span>
                        <span className="metric-value">{entity.signal}%</span>
                      </div>
                    )}
                    {entity.position.speed !== undefined && entity.position.speed > 0 && (
                      <div className="metric">
                        <span className="metric-icon">➡️</span>
                        <span className="metric-value">{entity.position.speed.toFixed(1)} m/s</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : view === 'list' ? (
            <div className="entities-list">
              <table className="entities-table">
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Callsign</th>
                    <th>Team</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Position</th>
                    <th>Battery</th>
                    <th>Signal</th>
                    <th>Last Update</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredEntities.map(entity => (
                    <tr
                      key={entity.id}
                      className={`entity-row ${selectedEntity === entity.id ? 'selected' : ''}`}
                      onClick={() => handleSelectEntity(entity.id)}
                    >
                      <td>{getTypeIcon(entity.type, entity.subType)}</td>
                      <td className="callsign">{entity.callsign}</td>
                      <td>{entity.team}</td>
                      <td>{entity.role}</td>
                      <td>
                        <span 
                          className="status-badge"
                          style={{ backgroundColor: getStatusColor(entity.status) }}
                        >
                          {entity.status}
                        </span>
                      </td>
                      <td className="position">
                        {entity.position.lat.toFixed(4)}, {entity.position.lng.toFixed(4)}
                      </td>
                      <td>{entity.battery ? `${entity.battery}%` : '-'}</td>
                      <td>{entity.signal ? `${entity.signal}%` : '-'}</td>
                      <td>{getTimeAgo(entity.lastUpdate)}</td>
                      <td>
                        <button
                          className="action-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteEntity(entity.id);
                          }}
                        >
                          ✕
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="entities-tactical">
              {filteredEntities.map(entity => (
                <div
                  key={entity.id}
                  className={`tactical-entity ${selectedEntity === entity.id ? 'selected' : ''} ${entity.type}`}
                  onClick={() => handleSelectEntity(entity.id)}
                >
                  <div className="tactical-icon">{getTypeIcon(entity.type, entity.subType)}</div>
                  <div className="tactical-info">
                    <h4>{entity.callsign}</h4>
                    <span className="tactical-team">{entity.team}</span>
                    <span className="tactical-role">{entity.role}</span>
                  </div>
                  <div className="tactical-status">
                    <span 
                      className="status-light"
                      style={{ backgroundColor: getStatusColor(entity.status) }}
                    />
                    <span className="update-time">{getTimeAgo(entity.lastUpdate)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Detail Panel */}
        {selectedEntityData && (
          <div className="entity-details-panel">
            <div className="details-header">
              <h2>
                {getTypeIcon(selectedEntityData.type, selectedEntityData.subType)} {selectedEntityData.callsign}
              </h2>
              <button 
                className="close-btn"
                onClick={() => setSelectedEntity(null)}
              >
                ✕
              </button>
            </div>

            <div className="details-content">
              <div className="detail-section">
                <h3>Basic Information</h3>
                <div className="detail-grid">
                  <div className="detail-item">
                    <span className="label">Callsign</span>
                    <span className="value">{selectedEntityData.callsign}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Type</span>
                    <span className="value">{selectedEntityData.type}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Status</span>
                    <span 
                      className="value status"
                      style={{ color: getStatusColor(selectedEntityData.status) }}
                    >
                      {selectedEntityData.status}
                    </span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Team</span>
                    <span className="value">{selectedEntityData.team}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Role</span>
                    <span className="value">{selectedEntityData.role}</span>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <h3>Position & Movement</h3>
                <div className="detail-grid">
                  <div className="detail-item">
                    <span className="label">Latitude</span>
                    <span className="value">{selectedEntityData.position.lat.toFixed(6)}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Longitude</span>
                    <span className="value">{selectedEntityData.position.lng.toFixed(6)}</span>
                  </div>
                  {selectedEntityData.position.alt !== undefined && (
                    <div className="detail-item">
                      <span className="label">Altitude</span>
                      <span className="value">{selectedEntityData.position.alt}m</span>
                    </div>
                  )}
                  {selectedEntityData.position.speed !== undefined && (
                    <div className="detail-item">
                      <span className="label">Speed</span>
                      <span className="value">{selectedEntityData.position.speed.toFixed(1)} m/s</span>
                    </div>
                  )}
                  {selectedEntityData.position.heading !== undefined && (
                    <div className="detail-item">
                      <span className="label">Heading</span>
                      <span className="value">{selectedEntityData.position.heading}°</span>
                    </div>
                  )}
                </div>
              </div>

              {(selectedEntityData.equipment || selectedEntityData.capabilities) && (
                <div className="detail-section">
                  <h3>Equipment & Capabilities</h3>
                  {selectedEntityData.equipment && (
                    <div className="detail-tags">
                      <span className="label">Equipment:</span>
                      <div className="tags">
                        {selectedEntityData.equipment.map(item => (
                          <span key={item} className="tag">{item}</span>
                        ))}
                      </div>
                    </div>
                  )}
                  {selectedEntityData.capabilities && (
                    <div className="detail-tags">
                      <span className="label">Capabilities:</span>
                      <div className="tags">
                        {selectedEntityData.capabilities.map(cap => (
                          <span key={cap} className="tag capability">{cap}</span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              <div className="detail-section">
                <h3>System Status</h3>
                <div className="detail-grid">
                  {selectedEntityData.battery !== undefined && (
                    <div className="detail-item">
                      <span className="label">Battery</span>
                      <span className="value">{selectedEntityData.battery}%</span>
                    </div>
                  )}
                  {selectedEntityData.signal !== undefined && (
                    <div className="detail-item">
                      <span className="label">Signal</span>
                      <span className="value">{selectedEntityData.signal}%</span>
                    </div>
                  )}
                  {selectedEntityData.temperature !== undefined && (
                    <div className="detail-item">
                      <span className="label">Temperature</span>
                      <span className="value">{selectedEntityData.temperature}°C</span>
                    </div>
                  )}
                  {selectedEntityData.fuelLevel !== undefined && (
                    <div className="detail-item">
                      <span className="label">Fuel Level</span>
                      <span className="value">{selectedEntityData.fuelLevel}%</span>
                    </div>
                  )}
                  {selectedEntityData.operatingTime !== undefined && (
                    <div className="detail-item">
                      <span className="label">Operating Time</span>
                      <span className="value">{selectedEntityData.operatingTime} hours</span>
                    </div>
                  )}
                  <div className="detail-item">
                    <span className="label">Last Update</span>
                    <span className="value">{new Date(selectedEntityData.lastUpdate).toLocaleString()}</span>
                  </div>
                </div>
              </div>

              {selectedEntityData.sensorData && selectedEntityData.sensorData.length > 0 && (
                <div className="detail-section">
                  <h3>Sensor Readings</h3>
                  <div className="sensor-data-grid">
                    {selectedEntityData.sensorData.map((data, index) => (
                      <div key={index} className="sensor-reading">
                        <span className="sensor-type">{data.type.replace(/_/g, ' ').toUpperCase()}</span>
                        <span className="sensor-value">
                          {data.value} {data.unit}
                        </span>
                        <span className="sensor-time">
                          {new Date(data.timestamp).toLocaleTimeString()}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {selectedEntityData.notes && (
                <div className="detail-section">
                  <h3>Notes</h3>
                  <p className="notes">{selectedEntityData.notes}</p>
                </div>
              )}

              <div className="detail-actions">
                <button className="btn-primary">Track Entity</button>
                <button className="btn-secondary">Send Message</button>
                <button 
                  className="btn-danger"
                  onClick={() => handleDeleteEntity(selectedEntityData.id)}
                >
                  Remove Entity
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Styles */}
      <style jsx>{`
        .entities-fullpage {
          height: 100vh;
          width: 100vw;
          display: flex;
          flex-direction: column;
          background: var(--color-bg-primary);
          overflow: hidden;
        }

        /* Header */
        .entities-header {
          padding: 20px 24px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border-bottom: 1px solid rgba(0, 212, 170, 0.15);
          backdrop-filter: blur(10px);
          display: flex;
          justify-content: space-between;
          align-items: center;
          flex-wrap: wrap;
          gap: 20px;
        }

        .header-title h1 {
          margin: 0;
          color: var(--color-accent);
          text-shadow: 0 0 8px rgba(0, 212, 170, 0.3);
          font-size: 1.5rem;
        }

        .entity-stats {
          display: flex;
          gap: 16px;
          margin-top: 8px;
        }

        .stat {
          font-size: 0.85rem;
          color: var(--color-text-secondary);
        }

        .stat.active {
          color: var(--color-success);
          font-weight: 600;
        }

        .stat.offline {
          color: var(--color-warning);
        }

        .header-controls {
          display: flex;
          align-items: center;
          gap: 16px;
        }

        .search-input {
          padding: 8px 16px;
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          color: var(--color-text-primary);
          width: 250px;
        }

        .filter-group {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .filter-select {
          padding: 8px 12px;
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          color: var(--color-text-primary);
        }

        .offline-toggle {
          display: flex;
          align-items: center;
          gap: 8px;
          cursor: pointer;
          color: var(--color-text-secondary);
        }

        .view-toggles {
          display: flex;
          gap: 4px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 6px;
          padding: 2px;
        }

        .view-btn {
          width: 32px;
          height: 32px;
          border: none;
          background: transparent;
          color: var(--color-text-secondary);
          cursor: pointer;
          border-radius: 4px;
          transition: all 0.2s ease;
        }

        .view-btn:hover {
          background: rgba(0, 212, 170, 0.1);
        }

        .view-btn.active {
          background: rgba(0, 212, 170, 0.2);
          color: var(--color-accent);
        }

        .btn-primary {
          padding: 8px 20px;
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.2) 0%, 
            rgba(0, 212, 170, 0.1) 100%);
          border: 1px solid rgba(0, 212, 170, 0.3);
          border-radius: 6px;
          color: var(--color-accent);
          font-weight: 600;
          cursor: pointer;
        }

        /* Content */
        .entities-content {
          flex: 1;
          display: flex;
          overflow: hidden;
        }

        .entities-display {
          flex: 1;
          overflow-y: auto;
          padding: 24px;
        }

        /* Grid View */
        .entities-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
          gap: 16px;
        }

        .entity-card {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.3) 0%, 
            rgba(15, 20, 25, 0.5) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
          padding: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .entity-card:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 212, 170, 0.1);
        }

        .entity-card.selected {
          border-color: rgba(0, 212, 170, 0.3);
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.05) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
        }

        .card-header {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 12px;
        }

        .entity-icon {
          font-size: 1.2rem;
        }

        .entity-callsign {
          flex: 1;
          margin: 0;
          color: var(--color-text-primary);
          font-size: 1.1rem;
        }

        .status-indicator {
          width: 10px;
          height: 10px;
          border-radius: 50%;
        }

        .card-content {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .entity-info {
          display: flex;
          justify-content: space-between;
          font-size: 0.85rem;
        }

        .info-label {
          color: var(--color-text-muted);
        }

        .info-value {
          color: var(--color-text-secondary);
        }

        .card-footer {
          display: flex;
          gap: 12px;
          margin-top: 12px;
          padding-top: 12px;
          border-top: 1px solid rgba(255, 255, 255, 0.05);
        }

        .metric {
          display: flex;
          align-items: center;
          gap: 4px;
          font-size: 0.8rem;
          color: var(--color-text-secondary);
        }

        /* List View */
        .entities-table {
          width: 100%;
          border-collapse: collapse;
        }

        .entities-table th {
          text-align: left;
          padding: 12px;
          border-bottom: 1px solid rgba(0, 212, 170, 0.2);
          color: var(--color-accent);
          font-weight: 600;
        }

        .entity-row {
          border-bottom: 1px solid rgba(255, 255, 255, 0.05);
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .entity-row:hover {
          background: rgba(0, 212, 170, 0.05);
        }

        .entity-row.selected {
          background: rgba(0, 212, 170, 0.1);
        }

        .entity-row td {
          padding: 12px;
          color: var(--color-text-secondary);
        }

        .callsign {
          color: var(--color-text-primary);
          font-weight: 600;
        }

        .status-badge {
          padding: 2px 8px;
          border-radius: 12px;
          font-size: 0.75rem;
          text-transform: uppercase;
          color: white;
        }

        .position {
          font-family: monospace;
          font-size: 0.85rem;
        }

        .action-btn {
          width: 24px;
          height: 24px;
          border-radius: 4px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(0, 0, 0, 0.2);
          color: var(--color-text-secondary);
          cursor: pointer;
        }

        .action-btn:hover {
          background: var(--color-error);
          color: white;
        }

        /* Tactical View */
        .entities-tactical {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
          gap: 12px;
        }

        .tactical-entity {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 16px;
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
          cursor: pointer;
        }

        .tactical-entity:hover {
          background: rgba(0, 212, 170, 0.05);
        }

        .tactical-entity.selected {
          border-color: rgba(0, 212, 170, 0.3);
          background: rgba(0, 212, 170, 0.1);
        }

        .tactical-icon {
          font-size: 2rem;
        }

        .tactical-info {
          flex: 1;
        }

        .tactical-info h4 {
          margin: 0 0 4px 0;
          color: var(--color-text-primary);
        }

        .tactical-team,
        .tactical-role {
          display: block;
          font-size: 0.85rem;
          color: var(--color-text-secondary);
        }

        .tactical-status {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
        }

        .status-light {
          width: 12px;
          height: 12px;
          border-radius: 50%;
        }

        .update-time {
          font-size: 0.75rem;
          color: var(--color-text-muted);
        }

        /* No Entities */
        .no-entities {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          color: var(--color-text-muted);
        }

        .no-entities-icon {
          font-size: 4rem;
          margin-bottom: 16px;
        }

        .no-entities h3 {
          margin: 0 0 8px 0;
          color: var(--color-text-primary);
        }

        /* Details Panel */
        .entity-details-panel {
          width: 400px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border-left: 1px solid rgba(0, 212, 170, 0.2);
          display: flex;
          flex-direction: column;
        }

        .details-header {
          padding: 20px;
          border-bottom: 1px solid rgba(0, 212, 170, 0.1);
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .details-header h2 {
          margin: 0;
          color: var(--color-accent);
          font-size: 1.2rem;
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .close-btn {
          width: 32px;
          height: 32px;
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(0, 0, 0, 0.2);
          color: var(--color-text-secondary);
          cursor: pointer;
        }

        .details-content {
          flex: 1;
          padding: 20px;
          overflow-y: auto;
        }

        .detail-section {
          margin-bottom: 24px;
        }

        .detail-section h3 {
          margin: 0 0 12px 0;
          color: var(--color-text-primary);
          font-size: 0.9rem;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .detail-grid {
          display: grid;
          gap: 12px;
        }

        .detail-item {
          display: flex;
          justify-content: space-between;
          padding: 8px 12px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 4px;
        }

        .detail-item .label {
          color: var(--color-text-muted);
          font-size: 0.85rem;
        }

        .detail-item .value {
          color: var(--color-text-primary);
          font-size: 0.85rem;
        }

        .detail-item .value.status {
          font-weight: 600;
          text-transform: uppercase;
        }

        .detail-tags {
          margin-bottom: 12px;
        }

        .detail-tags .label {
          display: block;
          margin-bottom: 8px;
          color: var(--color-text-muted);
          font-size: 0.85rem;
        }

        .tags {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        .tag {
          padding: 4px 8px;
          background: rgba(0, 212, 170, 0.1);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 4px;
          font-size: 0.8rem;
          color: var(--color-text-secondary);
        }

        .tag.capability {
          background: rgba(33, 150, 243, 0.1);
          border-color: rgba(33, 150, 243, 0.2);
        }

        .notes {
          padding: 12px;
          background: rgba(255, 152, 0, 0.05);
          border-left: 3px solid var(--color-warning);
          border-radius: 4px;
          color: var(--color-text-secondary);
          font-size: 0.85rem;
          line-height: 1.5;
        }

        /* Sensor Data Styles */
        .sensor-data-grid {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .sensor-reading {
          display: grid;
          grid-template-columns: 2fr 1fr auto;
          gap: 12px;
          padding: 12px;
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.05) 0%, 
            rgba(0, 0, 0, 0.3) 100%);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          align-items: center;
        }

        .sensor-type {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--color-accent);
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .sensor-value {
          font-size: 1rem;
          font-weight: 700;
          color: var(--color-text-primary);
          text-align: right;
          font-family: monospace;
        }

        .sensor-time {
          font-size: 0.7rem;
          color: var(--color-text-muted);
          opacity: 0.7;
        }

        .detail-actions {
          margin-top: 24px;
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .btn-secondary,
        .btn-danger {
          padding: 10px;
          border-radius: 6px;
          border: 1px solid;
          font-weight: 600;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .btn-secondary {
          background: rgba(0, 0, 0, 0.2);
          border-color: rgba(255, 255, 255, 0.1);
          color: var(--color-text-secondary);
        }

        .btn-danger {
          background: rgba(211, 47, 47, 0.1);
          border-color: rgba(211, 47, 47, 0.2);
          color: var(--color-error);
        }

        /* Responsive */
        @media (max-width: 1200px) {
          .entity-details-panel {
            position: absolute;
            right: 0;
            top: 0;
            bottom: 0;
            box-shadow: -4px 0 24px rgba(0, 0, 0, 0.6);
            z-index: 10;
          }
        }

        @media (max-width: 768px) {
          .entities-header {
            flex-direction: column;
            align-items: stretch;
          }

          .header-controls {
            flex-direction: column;
            gap: 12px;
          }

          .search-input {
            width: 100%;
          }

          .entities-grid {
            grid-template-columns: 1fr;
          }

          .entity-details-panel {
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
};

export default Entities;
