import React, { useState } from 'react';
import { RouteManagementPanel } from './RouteManagementPanel';
import { GeofenceManagementPanel } from './GeofenceManagementPanel';
import { MeasurementToolsPanel } from './MeasurementToolsPanel';
import { OfflineMapManager } from './OfflineMapManager';
import './MapToolsPanel.css';

interface MapToolsPanelProps {
  isOpen: boolean;
  onClose: () => void;
  onMapInteraction?: (type: 'route' | 'geofence' | 'measurement', data: any) => void;
}

type ActiveTool = 'routes' | 'geofences' | 'measurements' | 'offline' | null;

export const MapToolsPanel: React.FC<MapToolsPanelProps> = ({
  isOpen,
  onClose,
  onMapInteraction
}) => {
  const [activeTool, setActiveTool] = useState<ActiveTool>('routes');

  if (!isOpen) return null;

  const renderActiveToolContent = () => {
    switch (activeTool) {
      case 'routes':
        return (
          <RouteManagementPanel
            onCreateRoute={(routeData) => onMapInteraction?.('route', { action: 'create', data: routeData })}
            onEditRoute={(routeId, routeData) => onMapInteraction?.('route', { action: 'edit', id: routeId, data: routeData })}
            onDeleteRoute={(routeId) => onMapInteraction?.('route', { action: 'delete', id: routeId })}
          />
        );
      case 'geofences':
        return (
          <GeofenceManagementPanel
            onCreateGeofence={(geofenceData) => onMapInteraction?.('geofence', { action: 'create', data: geofenceData })}
            onEditGeofence={(geofenceId, geofenceData) => onMapInteraction?.('geofence', { action: 'edit', id: geofenceId, data: geofenceData })}
            onDeleteGeofence={(geofenceId) => onMapInteraction?.('geofence', { action: 'delete', id: geofenceId })}
          />
        );
      case 'measurements':
        return (
          <MeasurementToolsPanel
            onStartMeasurement={(type) => onMapInteraction?.('measurement', { action: 'start', type })}
            onClearMeasurements={() => onMapInteraction?.('measurement', { action: 'clear' })}
          />
        );
      case 'offline':
        return <OfflineMapManager />;
      default:
        return <div className="tool-placeholder">Select a tool from the toolbar</div>;
    }
  };

  return (
    <div className="map-tools-panel">
      <div className="tools-header">
        <h2>Map Tools</h2>
        <button className="close-button" onClick={onClose} aria-label="Close map tools">
          ×
        </button>
      </div>

      <div className="tools-toolbar">
        <button
          className={`tool-button ${activeTool === 'routes' ? 'active' : ''}`}
          onClick={() => setActiveTool('routes')}
          title="Route Management"
        >
          <span className="tool-icon">🛤️</span>
          <span className="tool-label">Routes</span>
        </button>

        <button
          className={`tool-button ${activeTool === 'geofences' ? 'active' : ''}`}
          onClick={() => setActiveTool('geofences')}
          title="Geofence Management"
        >
          <span className="tool-icon">⭕</span>
          <span className="tool-label">Geofences</span>
        </button>

        <button
          className={`tool-button ${activeTool === 'measurements' ? 'active' : ''}`}
          onClick={() => setActiveTool('measurements')}
          title="Measurement Tools"
        >
          <span className="tool-icon">📏</span>
          <span className="tool-label">Measure</span>
        </button>

        <button
          className={`tool-button ${activeTool === 'offline' ? 'active' : ''}`}
          onClick={() => setActiveTool('offline')}
          title="Offline Maps"
        >
          <span className="tool-icon">📱</span>
          <span className="tool-label">Offline</span>
        </button>
      </div>

      <div className="tools-content">
        {renderActiveToolContent()}
      </div>
    </div>
  );
};
