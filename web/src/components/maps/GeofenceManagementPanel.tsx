import React, { useState, useEffect, useCallback } from 'react';
import { mappingService, Geofence } from '../../services/mappingService';
import { formatDistance, formatArea, formatCoordinates } from '../../utils/mappingUtils';
import './GeofenceManagementPanel.css';

export interface GeofenceManagementPanelProps {
  className?: string;
  onGeofenceSelect?: (geofence: Geofence | null) => void;
  onGeofenceEdit?: (geofence: Geofence) => void;
  onGeofenceDelete?: (geofenceId: string) => void;
  onGeofenceCreate?: () => void;
  onGeofenceToggle?: (geofenceId: string, active: boolean) => void;
  selectedGeofenceId?: string;
  isVisible?: boolean;
  readOnly?: boolean;
}

interface GeofenceListState {
  geofences: Geofence[];
  loading: boolean;
  error: string | null;
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

export function GeofenceManagementPanel({
  className = '',
  onGeofenceSelect,
  onGeofenceEdit,
  onGeofenceDelete,
  onGeofenceCreate,
  onGeofenceToggle,
  selectedGeofenceId,
  isVisible = true,
  readOnly = false
}: GeofenceManagementPanelProps) {
  const [geofenceState, setGeofenceState] = useState<GeofenceListState>({
    geofences: [],
    loading: false,
    error: null,
    totalCount: 0,
    currentPage: 1,
    pageSize: 10
  });

  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<'all' | 'circle' | 'polygon' | 'rectangle'>('all');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all');
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'type'>('created');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [showGeofenceDetails, setShowGeofenceDetails] = useState<string | null>(null);
  const [editingGeofence, setEditingGeofence] = useState<Geofence | null>(null);

  // Load geofences from backend
  const loadGeofences = useCallback(async () => {
    setGeofenceState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const params = {
        page: geofenceState.currentPage,
        limit: geofenceState.pageSize,
        search: searchQuery || undefined,
        sortBy: sortBy,
        sortOrder: sortOrder
      };
      
      const response = await mappingService.listGeofences(params);
      
      setGeofenceState(prev => ({
        ...prev,
        geofences: response.geofences,
        totalCount: response.total,
        loading: false
      }));
    } catch (error) {
      console.error('Failed to load geofences:', error);
      setGeofenceState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load geofences',
        loading: false
      }));
    }
  }, [geofenceState.currentPage, geofenceState.pageSize, searchQuery, sortBy, sortOrder]);

  // Load geofences on mount and when dependencies change
  useEffect(() => {
    if (isVisible) {
      loadGeofences();
    }
  }, [isVisible, loadGeofences]);

  // Handle geofence selection
  const handleGeofenceSelect = useCallback((geofence: Geofence | null) => {
    onGeofenceSelect?.(geofence);
    setShowGeofenceDetails(geofence?.id || null);
  }, [onGeofenceSelect]);

  // Handle geofence deletion
  const handleDeleteGeofence = useCallback(async (geofenceId: string) => {
    if (!confirm('Are you sure you want to delete this geofence?')) {
      return;
    }

    try {
      await mappingService.deleteGeofence(geofenceId);
      onGeofenceDelete?.(geofenceId);
      await loadGeofences(); // Refresh list
      
      // Clear selection if deleted geofence was selected
      if (selectedGeofenceId === geofenceId) {
        handleGeofenceSelect(null);
      }
    } catch (error) {
      console.error('Failed to delete geofence:', error);
      alert('Failed to delete geofence. Please try again.');
    }
  }, [onGeofenceDelete, selectedGeofenceId, loadGeofences, handleGeofenceSelect]);

  // Handle geofence toggle
  const handleToggleGeofence = useCallback(async (geofenceId: string, active: boolean) => {
    try {
      await mappingService.updateGeofence(geofenceId, { active });
      onGeofenceToggle?.(geofenceId, active);
      await loadGeofences(); // Refresh list
    } catch (error) {
      console.error('Failed to toggle geofence:', error);
      alert('Failed to update geofence. Please try again.');
    }
  }, [onGeofenceToggle, loadGeofences]);

  // Handle geofence editing
  const handleEditGeofence = useCallback((geofence: Geofence) => {
    setEditingGeofence(geofence);
    onGeofenceEdit?.(geofence);
  }, [onGeofenceEdit]);

  // Handle geofence update
  const handleUpdateGeofence = useCallback(async (updatedGeofence: Partial<Geofence>) => {
    if (!editingGeofence) return;

    try {
      const updated = await mappingService.updateGeofence(editingGeofence.id, updatedGeofence);
      setEditingGeofence(null);
      await loadGeofences(); // Refresh list
      
      // Update selection if this geofence is currently selected
      if (selectedGeofenceId === editingGeofence.id) {
        handleGeofenceSelect(updated);
      }
    } catch (error) {
      console.error('Failed to update geofence:', error);
      alert('Failed to update geofence. Please try again.');
    }
  }, [editingGeofence, selectedGeofenceId, loadGeofences, handleGeofenceSelect]);

  // Handle pagination
  const handlePageChange = useCallback((page: number) => {
    setGeofenceState(prev => ({ ...prev, currentPage: page }));
  }, []);

  // Calculate geofence statistics
  const getGeofenceStats = useCallback((geofence: Geofence) => {
    let area = 0;
    let radius = 0;
    
    if (geofence.type === 'circle' && geofence.geometry.center && geofence.geometry.radius) {
      radius = geofence.geometry.radius;
      area = Math.PI * radius * radius;
    } else if (geofence.type === 'polygon' && geofence.geometry.points) {
      // Calculate polygon area (simplified)
      area = geofence.geometry.points.length * 1000; // Placeholder calculation
    } else if (geofence.type === 'rectangle' && geofence.geometry.bounds) {
      // Calculate rectangle area (simplified)
      area = 10000; // Placeholder calculation
    }
    
    return {
      area,
      radius,
      pointCount: geofence.geometry.points?.length || 0,
      created: new Date(geofence.created).toLocaleDateString(),
      modified: new Date(geofence.modified).toLocaleDateString()
    };
  }, []);

  // Format geofence for display
  const formatGeofenceForDisplay = useCallback((geofence: Geofence) => {
    const stats = getGeofenceStats(geofence);
    return {
      ...geofence,
      displayName: geofence.name || `${geofence.type.charAt(0).toUpperCase() + geofence.type.slice(1)} ${geofence.id.slice(-6)}`,
      formattedType: geofence.type.charAt(0).toUpperCase() + geofence.type.slice(1),
      formattedArea: formatArea(stats.area),
      formattedRadius: stats.radius > 0 ? formatDistance(stats.radius) : null,
      formattedCreated: stats.created,
      formattedModified: stats.modified,
      pointCount: stats.pointCount
    };
  }, [getGeofenceStats]);

  // Filter and sort geofences
  const filteredAndSortedGeofences = React.useMemo(() => {
    let filtered = geofenceState.geofences;

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(geofence => 
        geofence.name?.toLowerCase().includes(query) ||
        geofence.description?.toLowerCase().includes(query) ||
        geofence.id.toLowerCase().includes(query) ||
        geofence.type.toLowerCase().includes(query)
      );
    }

    // Apply type filter
    if (filterType !== 'all') {
      filtered = filtered.filter(geofence => geofence.type === filterType);
    }

    // Apply status filter
    if (filterStatus !== 'all') {
      const isActive = filterStatus === 'active';
      filtered = filtered.filter(geofence => geofence.active === isActive);
    }

    // Apply sorting
    filtered.sort((a, b) => {
      let aValue, bValue;
      
      switch (sortBy) {
        case 'name':
          aValue = a.name || a.id;
          bValue = b.name || b.id;
          break;
        case 'type':
          aValue = a.type;
          bValue = b.type;
          break;
        case 'created':
        default:
          aValue = new Date(a.created).getTime();
          bValue = new Date(b.created).getTime();
          break;
      }
      
      if (sortOrder === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

    return filtered.map(formatGeofenceForDisplay);
  }, [geofenceState.geofences, searchQuery, filterType, filterStatus, sortBy, sortOrder, formatGeofenceForDisplay]);

  // Get geofence type icon
  const getGeofenceTypeIcon = useCallback((type: string) => {
    switch (type) {
      case 'circle': return '⭕';
      case 'polygon': return '🔷';
      case 'rectangle': return '◼️';
      default: return '📍';
    }
  }, []);

  if (!isVisible) {
    return null;
  }

  return (
    <div className={`geofence-management-panel ${className}`}>
      <div className="geofence-panel-header">
        <h3>Geofence Management</h3>
        <div className="geofence-panel-actions">
          {!readOnly && (
            <button 
              className="btn-primary create-geofence-btn"
              onClick={onGeofenceCreate}
              title="Create new geofence"
            >
              ➕ New Geofence
            </button>
          )}
          <button 
            className="btn-secondary refresh-btn"
            onClick={loadGeofences}
            disabled={geofenceState.loading}
            title="Refresh geofences"
          >
            🔄 Refresh
          </button>
        </div>
      </div>

      <div className="geofence-panel-controls">
        <div className="search-controls">
          <input
            type="text"
            placeholder="Search geofences..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-controls">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value as any)}
            className="filter-select"
          >
            <option value="all">All Types</option>
            <option value="circle">Circle</option>
            <option value="polygon">Polygon</option>
            <option value="rectangle">Rectangle</option>
          </select>
          
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value as any)}
            className="filter-select"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>
        
        <div className="sort-controls">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'name' | 'created' | 'type')}
            className="sort-select"
          >
            <option value="created">Sort by Created</option>
            <option value="name">Sort by Name</option>
            <option value="type">Sort by Type</option>
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

      <div className="geofence-panel-content">
        {geofenceState.error && (
          <div className="error-message">
            <span className="error-icon">⚠️</span>
            {geofenceState.error}
          </div>
        )}

        {geofenceState.loading ? (
          <div className="loading-message">
            <span className="loading-spinner">⏳</span>
            Loading geofences...
          </div>
        ) : (
          <>
            <div className="geofences-list">
              {filteredAndSortedGeofences.length === 0 ? (
                <div className="empty-state">
                  <span className="empty-icon">📍</span>
                  <p>No geofences found</p>
                  {!readOnly && (
                    <button 
                      className="btn-primary"
                      onClick={onGeofenceCreate}
                    >
                      Create your first geofence
                    </button>
                  )}
                </div>
              ) : (
                filteredAndSortedGeofences.map((geofence) => (
                  <div
                    key={geofence.id}
                    className={`geofence-item ${selectedGeofenceId === geofence.id ? 'selected' : ''} ${geofence.active ? 'active' : 'inactive'}`}
                    onClick={() => handleGeofenceSelect(geofence)}
                  >
                    <div className="geofence-item-header">
                      <div className="geofence-name">
                        <span className="geofence-type-icon">
                          {getGeofenceTypeIcon(geofence.type)}
                        </span>
                        {geofence.displayName}
                      </div>
                      <div className="geofence-actions">
                        {!readOnly && (
                          <>
                            <button
                              className={`btn-icon toggle-btn ${geofence.active ? 'active' : 'inactive'}`}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleToggleGeofence(geofence.id, !geofence.active);
                              }}
                              title={geofence.active ? 'Deactivate geofence' : 'Activate geofence'}
                            >
                              {geofence.active ? '🟢' : '🔴'}
                            </button>
                            <button
                              className="btn-icon edit-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleEditGeofence(geofence);
                              }}
                              title="Edit geofence"
                            >
                              ✏️
                            </button>
                            <button
                              className="btn-icon delete-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteGeofence(geofence.id);
                              }}
                              title="Delete geofence"
                            >
                              🗑️
                            </button>
                          </>
                        )}
                        <button
                          className="btn-icon details-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            setShowGeofenceDetails(
                              showGeofenceDetails === geofence.id ? null : geofence.id
                            );
                          }}
                          title="Show/hide details"
                        >
                          {showGeofenceDetails === geofence.id ? '🔼' : '🔽'}
                        </button>
                      </div>
                    </div>

                    <div className="geofence-item-summary">
                      <span className="geofence-stat">
                        📐 {geofence.formattedType}
                      </span>
                      <span className="geofence-stat">
                        📏 {geofence.formattedArea}
                      </span>
                      {geofence.formattedRadius && (
                        <span className="geofence-stat">
                          ⚪ {geofence.formattedRadius}
                        </span>
                      )}
                      <span className="geofence-stat">
                        📅 {geofence.formattedCreated}
                      </span>
                      <span className={`geofence-status ${geofence.active ? 'active' : 'inactive'}`}>
                        {geofence.active ? '🟢 Active' : '🔴 Inactive'}
                      </span>
                    </div>

                    {geofence.description && (
                      <div className="geofence-description">
                        {geofence.description}
                      </div>
                    )}

                    {showGeofenceDetails === geofence.id && (
                      <div className="geofence-details">
                        <div className="geofence-detail-row">
                          <strong>ID:</strong> <code>{geofence.id}</code>
                        </div>
                        <div className="geofence-detail-row">
                          <strong>Type:</strong> {geofence.formattedType}
                        </div>
                        <div className="geofence-detail-row">
                          <strong>Created:</strong> {new Date(geofence.created).toLocaleString()}
                        </div>
                        <div className="geofence-detail-row">
                          <strong>Modified:</strong> {new Date(geofence.modified).toLocaleString()}
                        </div>
                        {geofence.type === 'circle' && geofence.geometry.center && (
                          <div className="geofence-detail-row">
                            <strong>Center:</strong> {formatCoordinates(geofence.geometry.center.lat, geofence.geometry.center.lng, { precision: 4 })}
                          </div>
                        )}
                        {geofence.geometry.points && geofence.geometry.points.length > 0 && (
                          <div className="geofence-detail-row">
                            <strong>Points:</strong>
                            <div className="geometry-points-list">
                              {geofence.geometry.points.map((point, index) => (
                                <div key={index} className="geometry-point-item">
                                  <span className="point-number">{index + 1}.</span>
                                  <span className="point-coords">
                                    {formatCoordinates(point.lat, point.lng, { precision: 4 })}
                                  </span>
                                </div>
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>

            {geofenceState.totalCount > geofenceState.pageSize && (
              <div className="pagination-controls">
                <button
                  className="btn-secondary"
                  onClick={() => handlePageChange(geofenceState.currentPage - 1)}
                  disabled={geofenceState.currentPage <= 1}
                >
                  ← Previous
                </button>
                
                <span className="pagination-info">
                  Page {geofenceState.currentPage} of {Math.ceil(geofenceState.totalCount / geofenceState.pageSize)}
                </span>
                
                <button
                  className="btn-secondary"
                  onClick={() => handlePageChange(geofenceState.currentPage + 1)}
                  disabled={geofenceState.currentPage >= Math.ceil(geofenceState.totalCount / geofenceState.pageSize)}
                >
                  Next →
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Geofence Edit Modal */}
      {editingGeofence && (
        <GeofenceEditModal
          geofence={editingGeofence}
          onSave={handleUpdateGeofence}
          onCancel={() => setEditingGeofence(null)}
        />
      )}
    </div>
  );
}

// Geofence Edit Modal Component
interface GeofenceEditModalProps {
  geofence: Geofence;
  onSave: (updatedGeofence: Partial<Geofence>) => void;
  onCancel: () => void;
}

function GeofenceEditModal({ geofence, onSave, onCancel }: GeofenceEditModalProps) {
  const [name, setName] = useState(geofence.name || '');
  const [description, setDescription] = useState(geofence.description || '');
  const [active, setActive] = useState(geofence.active);

  const handleSave = () => {
    onSave({
      name: name.trim() || undefined,
      description: description.trim() || undefined,
      active
    });
  };

  const getGeofenceTypeIcon = (type: string) => {
    switch (type) {
      case 'circle': return '⭕';
      case 'polygon': return '🔷';
      case 'rectangle': return '◼️';
      default: return '📍';
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content geofence-edit-modal">
        <div className="modal-header">
          <h4>
            <span className="modal-type-icon">{getGeofenceTypeIcon(geofence.type)}</span>
            Edit {geofence.type.charAt(0).toUpperCase() + geofence.type.slice(1)} Geofence
          </h4>
          <button className="btn-icon close-btn" onClick={onCancel}>
            ✖️
          </button>
        </div>
        
        <div className="modal-body">
          <div className="form-group">
            <label htmlFor="geofence-name">Geofence Name</label>
            <input
              id="geofence-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter geofence name..."
              maxLength={100}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="geofence-description">Description</label>
            <textarea
              id="geofence-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter geofence description..."
              rows={4}
              maxLength={500}
            />
          </div>

          <div className="form-group">
            <label>
              <input
                type="checkbox"
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
                style={{ marginRight: '8px' }}
              />
              Active (monitoring enabled)
            </label>
          </div>

          <div className="geofence-info">
            <div className="info-item">
              <strong>Type:</strong> {geofence.type.charAt(0).toUpperCase() + geofence.type.slice(1)}
            </div>
            <div className="info-item">
              <strong>Created:</strong> {new Date(geofence.created).toLocaleString()}
            </div>
            {geofence.type === 'circle' && geofence.geometry.radius && (
              <div className="info-item">
                <strong>Radius:</strong> {formatDistance(geofence.geometry.radius)}
              </div>
            )}
            {geofence.geometry.points && (
              <div className="info-item">
                <strong>Points:</strong> {geofence.geometry.points.length}
              </div>
            )}
          </div>
        </div>
        
        <div className="modal-footer">
          <button className="btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button className="btn-primary" onClick={handleSave}>
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}
