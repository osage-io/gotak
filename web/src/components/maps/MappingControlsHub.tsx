import React, { useState, useCallback, useRef } from 'react';
import { TacticalMap, TacticalMapProps } from './TacticalMap';
import { RouteManagementPanel } from './RouteManagementPanel';
import { GeofenceManagementPanel } from './GeofenceManagementPanel';
import { MeasurementToolsPanel, MeasurementResult } from './MeasurementToolsPanel';
import { Route, Geofence } from '../../services/mappingService';
import { LatLng } from '../../utils/coordinates';
import './MappingControlsHub.css';

export type PanelType = 'routes' | 'geofences' | 'measurements' | null;

export interface MappingControlsHubProps {
  className?: string;
  // Map props
  mapProps?: Partial<TacticalMapProps>;
  // Control settings
  showRoutePanel?: boolean;
  showGeofencePanel?: boolean;
  showMeasurementPanel?: boolean;
  defaultActivePanel?: PanelType;
  panelPosition?: 'left' | 'right';
  readOnly?: boolean;
  // Event handlers
  onRouteCreated?: (route: Route) => void;
  onGeofenceCreated?: (geofence: Geofence) => void;
  onMeasurementTaken?: (measurement: MeasurementResult) => void;
  onPanelChange?: (panel: PanelType) => void;
}

export function MappingControlsHub({
  className = '',
  mapProps = {},
  showRoutePanel = true,
  showGeofencePanel = true,
  showMeasurementPanel = true,
  defaultActivePanel = null,
  panelPosition = 'left',
  readOnly = false,
  onRouteCreated,
  onGeofenceCreated,
  onMeasurementTaken,
  onPanelChange
}: MappingControlsHubProps) {
  // State for active panel
  const [activePanel, setActivePanel] = useState<PanelType>(defaultActivePanel);
  
  // State for selected items
  const [selectedRoute, setSelectedRoute] = useState<Route | null>(null);
  const [selectedGeofence, setSelectedGeofence] = useState<Geofence | null>(null);
  const [selectedMeasurement, setSelectedMeasurement] = useState<MeasurementResult | null>(null);
  
  // State for measurements
  const [measurements, setMeasurements] = useState<MeasurementResult[]>([]);
  
  // State for current measurement mode
  const [currentMeasurementMode, setCurrentMeasurementMode] = useState<'distance' | 'area' | 'bearing' | null>(null);
  
  // Ref for the tactical map to access its API
  const mapRef = useRef<any>(null);

  // Handle panel change
  const handlePanelChange = useCallback((panel: PanelType) => {
    setActivePanel(activePanel === panel ? null : panel);
    onPanelChange?.(activePanel === panel ? null : panel);
  }, [activePanel, onPanelChange]);

  // Handle route selection
  const handleRouteSelect = useCallback((route: Route | null) => {
    setSelectedRoute(route);
    // Clear other selections
    setSelectedGeofence(null);
    setSelectedMeasurement(null);
    // TODO: Highlight route on map
  }, []);

  // Handle geofence selection
  const handleGeofenceSelect = useCallback((geofence: Geofence | null) => {
    setSelectedGeofence(geofence);
    // Clear other selections
    setSelectedRoute(null);
    setSelectedMeasurement(null);
    // TODO: Highlight geofence on map
  }, []);

  // Handle measurement selection
  const handleMeasurementSelect = useCallback((measurement: MeasurementResult | null) => {
    setSelectedMeasurement(measurement);
    // Clear other selections
    setSelectedRoute(null);
    setSelectedGeofence(null);
    // TODO: Highlight measurement on map
  }, []);

  // Handle route creation
  const handleRouteCreate = useCallback(() => {
    // Start route creation mode on map
    setCurrentMeasurementMode(null);
    // TODO: Activate route creation mode on map
    console.log('Starting route creation...');
  }, []);

  // Handle geofence creation
  const handleGeofenceCreate = useCallback(() => {
    // Start geofence creation mode on map
    setCurrentMeasurementMode(null);
    // TODO: Activate geofence creation mode on map
    console.log('Starting geofence creation...');
  }, []);

  // Handle measurement start
  const handleMeasurementStart = useCallback((type: 'distance' | 'area' | 'bearing') => {
    setCurrentMeasurementMode(type);
    // TODO: Activate measurement mode on map
    console.log(`Starting ${type} measurement...`);
  }, []);

  // Handle measurement completion
  const handleMeasurement = useCallback((type: string, value: number, points: LatLng[]) => {
    const measurement: MeasurementResult = {
      id: `measurement-${Date.now()}`,
      type: type as 'distance' | 'area' | 'bearing',
      value,
      points,
      timestamp: new Date(),
      additionalInfo: type === 'bearing' && points.length === 2 
        ? `Distance: ${Math.round(value)} m` 
        : undefined
    };

    setMeasurements(prev => [measurement, ...prev]);
    onMeasurementTaken?.(measurement);
    
    // Clear measurement mode after completion (except for distance/area which can be continuous)
    if (type === 'bearing') {
      setCurrentMeasurementMode(null);
    }
  }, [onMeasurementTaken]);

  // Handle measurement deletion
  const handleMeasurementDelete = useCallback((measurementId: string) => {
    setMeasurements(prev => prev.filter(m => m.id !== measurementId));
    if (selectedMeasurement?.id === measurementId) {
      setSelectedMeasurement(null);
    }
  }, [selectedMeasurement]);

  // Handle measurement rename
  const handleMeasurementRename = useCallback((measurementId: string, name: string) => {
    setMeasurements(prev => 
      prev.map(m => m.id === measurementId ? { ...m, name } : m)
    );
  }, []);

  // Handle clear all measurements
  const handleClearAllMeasurements = useCallback(() => {
    if (measurements.length === 0) return;
    
    if (confirm(`Are you sure you want to delete all ${measurements.length} measurements?`)) {
      setMeasurements([]);
      setSelectedMeasurement(null);
      setCurrentMeasurementMode(null);
      // TODO: Clear measurements from map
    }
  }, [measurements.length]);

  // Calculate panel visibility
  const availablePanels = [
    showRoutePanel && { id: 'routes', label: 'Routes', icon: '🛣️' },
    showGeofencePanel && { id: 'geofences', label: 'Geofences', icon: '📍' },
    showMeasurementPanel && { id: 'measurements', label: 'Measurements', icon: '📏' }
  ].filter(Boolean);

  return (
    <div className={`mapping-controls-hub ${className}`}>
      {/* Control Panel Tabs */}
      {availablePanels.length > 0 && (
        <div className={`panel-tabs ${panelPosition}`}>
          {availablePanels.map((panel: any) => (
            <button
              key={panel.id}
              className={`panel-tab ${activePanel === panel.id ? 'active' : ''}`}
              onClick={() => handlePanelChange(panel.id)}
              title={panel.label}
              disabled={readOnly && panel.id !== 'measurements'}
            >
              <span className="tab-icon">{panel.icon}</span>
              <span className="tab-label">{panel.label}</span>
            </button>
          ))}
        </div>
      )}

      {/* Main Content Area */}
      <div className="hub-content">
        {/* Side Panel */}
        {activePanel && (
          <div className={`side-panel ${panelPosition}`}>
            {/* Route Management Panel */}
            {activePanel === 'routes' && showRoutePanel && (
              <RouteManagementPanel
                onRouteSelect={handleRouteSelect}
                onRouteCreate={handleRouteCreate}
                onRouteEdit={(route) => console.log('Edit route:', route)}
                onRouteDelete={(routeId) => console.log('Delete route:', routeId)}
                selectedRouteId={selectedRoute?.id}
                readOnly={readOnly}
                isVisible={true}
              />
            )}

            {/* Geofence Management Panel */}
            {activePanel === 'geofences' && showGeofencePanel && (
              <GeofenceManagementPanel
                onGeofenceSelect={handleGeofenceSelect}
                onGeofenceCreate={handleGeofenceCreate}
                onGeofenceEdit={(geofence) => console.log('Edit geofence:', geofence)}
                onGeofenceDelete={(geofenceId) => console.log('Delete geofence:', geofenceId)}
                onGeofenceToggle={(geofenceId, active) => console.log('Toggle geofence:', geofenceId, active)}
                selectedGeofenceId={selectedGeofence?.id}
                readOnly={readOnly}
                isVisible={true}
              />
            )}

            {/* Measurement Tools Panel */}
            {activePanel === 'measurements' && showMeasurementPanel && (
              <MeasurementToolsPanel
                measurements={measurements}
                onMeasurementSelect={handleMeasurementSelect}
                onMeasurementDelete={handleMeasurementDelete}
                onMeasurementRename={handleMeasurementRename}
                onStartMeasurement={handleMeasurementStart}
                onClearAllMeasurements={handleClearAllMeasurements}
                selectedMeasurementId={selectedMeasurement?.id}
                currentMeasurementMode={currentMeasurementMode}
                readOnly={readOnly}
                isVisible={true}
              />
            )}
          </div>
        )}

        {/* Map Area */}
        <div className={`map-area ${activePanel ? 'with-panel' : 'full-width'}`}>
          <TacticalMap
            ref={mapRef}
            {...mapProps}
            enableRouting={showRoutePanel && !readOnly}
            enableGeofencing={showGeofencePanel && !readOnly}
            enableMeasurement={showMeasurementPanel && !readOnly}
            onRouteCreated={onRouteCreated}
            onGeofenceCreated={onGeofenceCreated}
            onMeasurement={handleMeasurement}
            readOnly={readOnly}
          />
        </div>
      </div>

      {/* Status Bar */}
      <div className="hub-status-bar">
        <div className="status-left">
          {currentMeasurementMode && (
            <div className="active-measurement-indicator">
              <span className="indicator-icon">
                {currentMeasurementMode === 'distance' && '📏'}
                {currentMeasurementMode === 'area' && '📐'}
                {currentMeasurementMode === 'bearing' && '🧭'}
              </span>
              <span className="indicator-text">
                {currentMeasurementMode.charAt(0).toUpperCase() + currentMeasurementMode.slice(1)} Mode Active
              </span>
            </div>
          )}
        </div>
        
        <div className="status-right">
          <div className="status-item">
            <span className="status-icon">🛣️</span>
            <span className="status-count">{selectedRoute ? '1' : '0'} route</span>
          </div>
          <div className="status-item">
            <span className="status-icon">📍</span>
            <span className="status-count">{selectedGeofence ? '1' : '0'} geofence</span>
          </div>
          <div className="status-item">
            <span className="status-icon">📊</span>
            <span className="status-count">{measurements.length} measurements</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default MappingControlsHub;
