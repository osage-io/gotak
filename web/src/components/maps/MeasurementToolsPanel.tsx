import React, { useState, useCallback, useEffect } from 'react';
import { formatDistance, formatArea, formatBearing, formatCoordinates } from '../../utils/mappingUtils';
import { LatLng } from '../../utils/coordinates';
import './MeasurementToolsPanel.css';

export interface MeasurementResult {
  id: string;
  type: 'distance' | 'area' | 'bearing';
  value: number;
  points: LatLng[];
  timestamp: Date;
  name?: string;
  additionalInfo?: string;
}

export interface MeasurementToolsPanelProps {
  className?: string;
  measurements: MeasurementResult[];
  onMeasurementSelect?: (measurement: MeasurementResult | null) => void;
  onMeasurementDelete?: (measurementId: string) => void;
  onMeasurementRename?: (measurementId: string, name: string) => void;
  onStartMeasurement?: (type: 'distance' | 'area' | 'bearing') => void;
  onClearAllMeasurements?: () => void;
  selectedMeasurementId?: string;
  isVisible?: boolean;
  readOnly?: boolean;
  currentMeasurementMode?: 'distance' | 'area' | 'bearing' | null;
}

export function MeasurementToolsPanel({
  className = '',
  measurements = [],
  onMeasurementSelect,
  onMeasurementDelete,
  onMeasurementRename,
  onStartMeasurement,
  onClearAllMeasurements,
  selectedMeasurementId,
  isVisible = true,
  readOnly = false,
  currentMeasurementMode
}: MeasurementToolsPanelProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<'all' | 'distance' | 'area' | 'bearing'>('all');
  const [sortBy, setSortBy] = useState<'timestamp' | 'type' | 'value'>('timestamp');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [showMeasurementDetails, setShowMeasurementDetails] = useState<string | null>(null);
  const [renamingMeasurement, setRenamingMeasurement] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState('');

  // Handle measurement selection
  const handleMeasurementSelect = useCallback((measurement: MeasurementResult | null) => {
    onMeasurementSelect?.(measurement);
  }, [onMeasurementSelect]);

  // Handle measurement deletion
  const handleDeleteMeasurement = useCallback((measurementId: string) => {
    if (!confirm('Are you sure you want to delete this measurement?')) {
      return;
    }
    onMeasurementDelete?.(measurementId);
  }, [onMeasurementDelete]);

  // Handle measurement rename
  const handleStartRename = useCallback((measurement: MeasurementResult) => {
    setRenamingMeasurement(measurement.id);
    setRenameValue(measurement.name || '');
  }, []);

  const handleSaveRename = useCallback(() => {
    if (renamingMeasurement && renameValue.trim()) {
      onMeasurementRename?.(renamingMeasurement, renameValue.trim());
    }
    setRenamingMeasurement(null);
    setRenameValue('');
  }, [renamingMeasurement, renameValue, onMeasurementRename]);

  const handleCancelRename = useCallback(() => {
    setRenamingMeasurement(null);
    setRenameValue('');
  }, []);

  // Calculate statistics
  const measurementStats = React.useMemo(() => {
    const total = measurements.length;
    const distanceCount = measurements.filter(m => m.type === 'distance').length;
    const areaCount = measurements.filter(m => m.type === 'area').length;
    const bearingCount = measurements.filter(m => m.type === 'bearing').length;

    return {
      total,
      distanceCount,
      areaCount,
      bearingCount
    };
  }, [measurements]);

  // Format measurement for display
  const formatMeasurementForDisplay = useCallback((measurement: MeasurementResult) => {
    let formattedValue = '';
    let unitSymbol = '';

    switch (measurement.type) {
      case 'distance':
        formattedValue = formatDistance(measurement.value);
        unitSymbol = '📏';
        break;
      case 'area':
        formattedValue = formatArea(measurement.value);
        unitSymbol = '📐';
        break;
      case 'bearing':
        formattedValue = formatBearing(measurement.value);
        unitSymbol = '🧭';
        break;
    }

    return {
      ...measurement,
      displayName: measurement.name || `${measurement.type.charAt(0).toUpperCase() + measurement.type.slice(1)} ${measurement.id.slice(-6)}`,
      formattedValue,
      formattedType: measurement.type.charAt(0).toUpperCase() + measurement.type.slice(1),
      unitSymbol,
      formattedTimestamp: new Date(measurement.timestamp).toLocaleString()
    };
  }, []);

  // Filter and sort measurements
  const filteredAndSortedMeasurements = React.useMemo(() => {
    let filtered = measurements;

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(measurement => 
        measurement.name?.toLowerCase().includes(query) ||
        measurement.type.toLowerCase().includes(query) ||
        measurement.id.toLowerCase().includes(query)
      );
    }

    // Apply type filter
    if (filterType !== 'all') {
      filtered = filtered.filter(measurement => measurement.type === filterType);
    }

    // Apply sorting
    filtered.sort((a, b) => {
      let aValue, bValue;
      
      switch (sortBy) {
        case 'type':
          aValue = a.type;
          bValue = b.type;
          break;
        case 'value':
          aValue = a.value;
          bValue = b.value;
          break;
        case 'timestamp':
        default:
          aValue = new Date(a.timestamp).getTime();
          bValue = new Date(b.timestamp).getTime();
          break;
      }
      
      if (sortOrder === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

    return filtered.map(formatMeasurementForDisplay);
  }, [measurements, searchQuery, filterType, sortBy, sortOrder, formatMeasurementForDisplay]);

  // Get measurement type icon
  const getMeasurementTypeIcon = useCallback((type: string) => {
    switch (type) {
      case 'distance': return '📏';
      case 'area': return '📐';
      case 'bearing': return '🧭';
      default: return '📊';
    }
  }, []);

  // Get measurement mode status
  const getMeasurementModeStatus = useCallback(() => {
    if (!currentMeasurementMode) return null;
    
    const modeNames = {
      distance: 'Distance Measurement',
      area: 'Area Measurement',  
      bearing: 'Bearing Measurement'
    };

    return {
      mode: currentMeasurementMode,
      name: modeNames[currentMeasurementMode],
      icon: getMeasurementTypeIcon(currentMeasurementMode)
    };
  }, [currentMeasurementMode, getMeasurementTypeIcon]);

  const measurementModeStatus = getMeasurementModeStatus();

  if (!isVisible) {
    return null;
  }

  return (
    <div className={`measurement-tools-panel ${className}`}>
      <div className="measurement-panel-header">
        <h3>Measurement Tools</h3>
        <div className="measurement-panel-actions">
          {!readOnly && (
            <button 
              className="btn-secondary clear-all-btn"
              onClick={onClearAllMeasurements}
              disabled={measurements.length === 0}
              title="Clear all measurements"
            >
              🗑️ Clear All
            </button>
          )}
        </div>
      </div>

      {/* Current Measurement Mode Status */}
      {measurementModeStatus && (
        <div className="current-measurement-mode">
          <span className="mode-icon">{measurementModeStatus.icon}</span>
          <span className="mode-text">Active: {measurementModeStatus.name}</span>
          <span className="mode-indicator">●</span>
        </div>
      )}

      {/* Measurement Tools */}
      {!readOnly && (
        <div className="measurement-tools">
          <div className="tool-buttons">
            <button
              className={`tool-btn distance-btn ${currentMeasurementMode === 'distance' ? 'active' : ''}`}
              onClick={() => onStartMeasurement?.('distance')}
              title="Measure Distance"
            >
              📏 Distance
            </button>
            <button
              className={`tool-btn area-btn ${currentMeasurementMode === 'area' ? 'active' : ''}`}
              onClick={() => onStartMeasurement?.('area')}
              title="Measure Area"
            >
              📐 Area
            </button>
            <button
              className={`tool-btn bearing-btn ${currentMeasurementMode === 'bearing' ? 'active' : ''}`}
              onClick={() => onStartMeasurement?.('bearing')}
              title="Measure Bearing"
            >
              🧭 Bearing
            </button>
          </div>
        </div>
      )}

      {/* Statistics */}
      <div className="measurement-stats">
        <div className="stat-item">
          <span className="stat-icon">📊</span>
          <span className="stat-value">{measurementStats.total}</span>
          <span className="stat-label">Total</span>
        </div>
        <div className="stat-item">
          <span className="stat-icon">📏</span>
          <span className="stat-value">{measurementStats.distanceCount}</span>
          <span className="stat-label">Distance</span>
        </div>
        <div className="stat-item">
          <span className="stat-icon">📐</span>
          <span className="stat-value">{measurementStats.areaCount}</span>
          <span className="stat-label">Area</span>
        </div>
        <div className="stat-item">
          <span className="stat-icon">🧭</span>
          <span className="stat-value">{measurementStats.bearingCount}</span>
          <span className="stat-label">Bearing</span>
        </div>
      </div>

      {/* Controls */}
      <div className="measurement-panel-controls">
        <div className="search-controls">
          <input
            type="text"
            placeholder="Search measurements..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-sort-controls">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value as any)}
            className="filter-select"
          >
            <option value="all">All Types</option>
            <option value="distance">Distance</option>
            <option value="area">Area</option>
            <option value="bearing">Bearing</option>
          </select>
          
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'timestamp' | 'type' | 'value')}
            className="sort-select"
          >
            <option value="timestamp">Sort by Time</option>
            <option value="type">Sort by Type</option>
            <option value="value">Sort by Value</option>
          </select>
          
          <button
            className="sort-order-btn"
            onClick={() => setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc')}
            title={`Sort ${sortOrder === 'asc' ? 'descending' : 'ascending'}`}
          >
            {sortOrder === 'asc' ? '⬆️' : '⬇️'}
          </button>
        </div>
      </div>

      {/* Measurements List */}
      <div className="measurement-panel-content">
        {filteredAndSortedMeasurements.length === 0 ? (
          <div className="empty-state">
            <span className="empty-icon">📊</span>
            <p>
              {measurements.length === 0 
                ? 'No measurements taken' 
                : 'No measurements match your search'
              }
            </p>
            {!readOnly && measurements.length === 0 && (
              <div className="empty-actions">
                <button 
                  className="btn-primary"
                  onClick={() => onStartMeasurement?.('distance')}
                >
                  Start Measuring
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="measurements-list">
            {filteredAndSortedMeasurements.map((measurement) => (
              <div
                key={measurement.id}
                className={`measurement-item ${selectedMeasurementId === measurement.id ? 'selected' : ''} ${measurement.type}`}
                onClick={() => handleMeasurementSelect(measurement)}
              >
                <div className="measurement-item-header">
                  <div className="measurement-name">
                    <span className="measurement-type-icon">
                      {measurement.unitSymbol}
                    </span>
                    {renamingMeasurement === measurement.id ? (
                      <div className="rename-input-group">
                        <input
                          type="text"
                          value={renameValue}
                          onChange={(e) => setRenameValue(e.target.value)}
                          className="rename-input"
                          placeholder="Enter name..."
                          maxLength={50}
                          autoFocus
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              handleSaveRename();
                            } else if (e.key === 'Escape') {
                              handleCancelRename();
                            }
                          }}
                        />
                        <button 
                          className="btn-icon save-rename-btn"
                          onClick={handleSaveRename}
                          title="Save name"
                        >
                          ✅
                        </button>
                        <button 
                          className="btn-icon cancel-rename-btn"
                          onClick={handleCancelRename}
                          title="Cancel rename"
                        >
                          ❌
                        </button>
                      </div>
                    ) : (
                      measurement.displayName
                    )}
                  </div>
                  <div className="measurement-actions">
                    {!readOnly && renamingMeasurement !== measurement.id && (
                      <>
                        <button
                          className="btn-icon rename-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleStartRename(measurement);
                          }}
                          title="Rename measurement"
                        >
                          ✏️
                        </button>
                        <button
                          className="btn-icon delete-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteMeasurement(measurement.id);
                          }}
                          title="Delete measurement"
                        >
                          🗑️
                        </button>
                      </>
                    )}
                    <button
                      className="btn-icon details-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        setShowMeasurementDetails(
                          showMeasurementDetails === measurement.id ? null : measurement.id
                        );
                      }}
                      title="Show/hide details"
                    >
                      {showMeasurementDetails === measurement.id ? '🔼' : '🔽'}
                    </button>
                  </div>
                </div>

                <div className="measurement-item-summary">
                  <span className="measurement-value">
                    {measurement.formattedValue}
                  </span>
                  <span className="measurement-type">
                    {measurement.formattedType}
                  </span>
                  <span className="measurement-time">
                    {new Date(measurement.timestamp).toLocaleTimeString()}
                  </span>
                </div>

                {measurement.additionalInfo && (
                  <div className="measurement-additional-info">
                    {measurement.additionalInfo}
                  </div>
                )}

                {showMeasurementDetails === measurement.id && (
                  <div className="measurement-details">
                    <div className="measurement-detail-row">
                      <strong>ID:</strong> <code>{measurement.id}</code>
                    </div>
                    <div className="measurement-detail-row">
                      <strong>Type:</strong> {measurement.formattedType}
                    </div>
                    <div className="measurement-detail-row">
                      <strong>Value:</strong> {measurement.formattedValue}
                    </div>
                    <div className="measurement-detail-row">
                      <strong>Timestamp:</strong> {measurement.formattedTimestamp}
                    </div>
                    <div className="measurement-detail-row">
                      <strong>Points:</strong>
                      <div className="measurement-points-list">
                        {measurement.points.map((point, index) => (
                          <div key={index} className="measurement-point-item">
                            <span className="point-number">{index + 1}.</span>
                            <span className="point-coords">
                              {formatCoordinates(point.lat, point.lng, { precision: 4 })}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
