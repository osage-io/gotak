import React, { useState } from 'react';
import { TacticalMap } from '../maps/TacticalMap';
import { Route, Geofence } from '../../services/mappingService';
import { LatLng } from '../../utils/coordinates';

export function TacticalMapExample() {
  const [mapConfig, setMapConfig] = useState({
    enableRouting: true,
    enableGeofencing: true,
    enableMeasurement: true,
    showRoutes: true,
    showGeofences: true,
    showEntities: true,
    showCoordinates: true,
    coordinateFormat: 'dd' as const,
  });

  const [recentActivity, setRecentActivity] = useState<string[]>([]);

  const addActivity = (activity: string) => {
    setRecentActivity(prev => [
      `${new Date().toLocaleTimeString()}: ${activity}`,
      ...prev.slice(0, 9) // Keep only last 10 activities
    ]);
  };

  const handleRouteCreated = (route: Route) => {
    console.log('Route created:', route);
    addActivity(`Route created: ${route.name || 'Unnamed'} with ${route.waypoints.length} waypoints`);
  };

  const handleGeofenceCreated = (geofence: Geofence) => {
    console.log('Geofence created:', geofence);
    addActivity(`Geofence created: ${geofence.name || 'Unnamed'} (${geofence.type})`);
  };

  const handleMeasurement = (type: string, value: number, points: LatLng[]) => {
    console.log('Measurement taken:', { type, value, points });
    let description = '';
    
    switch (type) {
      case 'distance':
        description = `${(value / 1000).toFixed(2)} km`;
        break;
      case 'area':
        description = value < 1000000 
          ? `${Math.round(value)} m²`
          : `${(value / 1000000).toFixed(2)} km²`;
        break;
      case 'bearing':
        description = `${Math.round(value)}°`;
        break;
    }
    
    addActivity(`${type.charAt(0).toUpperCase() + type.slice(1)} measured: ${description}`);
  };

  const handleMapClick = (latLng: LatLng) => {
    console.log('Map clicked at:', latLng);
  };

  const handleMapMove = (center: LatLng, zoom: number) => {
    console.log('Map moved to:', center, 'zoom:', zoom);
  };

  return (
    <div style={{ display: 'flex', height: '100vh', fontFamily: 'system-ui' }}>
      {/* Configuration Panel */}
      <div style={{ 
        width: '300px', 
        background: '#f5f5f5', 
        padding: '20px',
        borderRight: '1px solid #ddd',
        overflow: 'auto'
      }}>
        <h2 style={{ marginTop: 0, color: '#333' }}>TacticalMap Demo</h2>
        
        <div style={{ marginBottom: '20px' }}>
          <h3 style={{ color: '#666', fontSize: '14px', marginBottom: '10px' }}>Features</h3>
          
          <label style={{ display: 'block', marginBottom: '8px', fontSize: '13px' }}>
            <input
              type="checkbox"
              checked={mapConfig.enableRouting}
              onChange={e => setMapConfig(prev => ({ ...prev, enableRouting: e.target.checked }))}
              style={{ marginRight: '6px' }}
            />
            Route Planning
          </label>
          
          <label style={{ display: 'block', marginBottom: '8px', fontSize: '13px' }}>
            <input
              type="checkbox"
              checked={mapConfig.enableGeofencing}
              onChange={e => setMapConfig(prev => ({ ...prev, enableGeofencing: e.target.checked }))}
              style={{ marginRight: '6px' }}
            />
            Geofencing
          </label>
          
          <label style={{ display: 'block', marginBottom: '8px', fontSize: '13px' }}>
            <input
              type="checkbox"
              checked={mapConfig.enableMeasurement}
              onChange={e => setMapConfig(prev => ({ ...prev, enableMeasurement: e.target.checked }))}
              style={{ marginRight: '6px' }}
            />
            Measurement Tools
          </label>
          
          <label style={{ display: 'block', marginBottom: '8px', fontSize: '13px' }}>
            <input
              type="checkbox"
              checked={mapConfig.showEntities}
              onChange={e => setMapConfig(prev => ({ ...prev, showEntities: e.target.checked }))}
              style={{ marginRight: '6px' }}
            />
            Entity Tracking
          </label>
          
          <label style={{ display: 'block', marginBottom: '8px', fontSize: '13px' }}>
            <input
              type="checkbox"
              checked={mapConfig.showCoordinates}
              onChange={e => setMapConfig(prev => ({ ...prev, showCoordinates: e.target.checked }))}
              style={{ marginRight: '6px' }}
            />
            Show Coordinates
          </label>
        </div>

        <div style={{ marginBottom: '20px' }}>
          <h3 style={{ color: '#666', fontSize: '14px', marginBottom: '10px' }}>Coordinate Format</h3>
          <select
            value={mapConfig.coordinateFormat}
            onChange={e => setMapConfig(prev => ({ ...prev, coordinateFormat: e.target.value as any }))}
            style={{ width: '100%', padding: '4px', fontSize: '13px' }}
          >
            <option value="dd">Decimal Degrees</option>
            <option value="dms">Degrees Minutes Seconds</option>
            <option value="mgrs">MGRS</option>
          </select>
        </div>

        {recentActivity.length > 0 && (
          <div>
            <h3 style={{ color: '#666', fontSize: '14px', marginBottom: '10px' }}>Recent Activity</h3>
            <div style={{ 
              background: '#fff',
              border: '1px solid #ddd',
              borderRadius: '4px',
              padding: '8px',
              maxHeight: '200px',
              overflow: 'auto',
              fontSize: '12px'
            }}>
              {recentActivity.map((activity, index) => (
                <div key={index} style={{ 
                  padding: '2px 0',
                  borderBottom: index < recentActivity.length - 1 ? '1px solid #eee' : 'none'
                }}>
                  {activity}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Map Area */}
      <div style={{ flex: 1 }}>
        <TacticalMap
          initialCenter={{ lat: 39.8283, lng: -98.5795 }} // Center of US
          initialZoom={6}
          height="100%"
          width="100%"
          
          // Basic features
          showCoordinates={mapConfig.showCoordinates}
          coordinateFormat={mapConfig.coordinateFormat}
          showScale={true}
          
          // Entity tracking
          showEntities={mapConfig.showEntities}
          showTrails={false}
          autoCenter={false}
          
          // Advanced mapping features
          enableRouting={mapConfig.enableRouting}
          enableGeofencing={mapConfig.enableGeofencing}
          enableMeasurement={mapConfig.enableMeasurement}
          showRoutes={mapConfig.showRoutes}
          showGeofences={mapConfig.showGeofences}
          
          // Event handlers
          onMapClick={handleMapClick}
          onMapMove={handleMapMove}
          onRouteCreated={handleRouteCreated}
          onGeofenceCreated={handleGeofenceCreated}
          onMeasurement={handleMeasurement}
        />
      </div>
    </div>
  );
}

export default TacticalMapExample;
