import React, { useState, useEffect, useCallback } from 'react';
import { mappingService, Route, RouteOptions } from '../../services/mappingService';
import { formatDistance, formatCoordinates } from '../../utils/mappingUtils';
import './RouteManagementPanel.css';

export interface RouteManagementPanelProps {
  className?: string;
  onRouteSelect?: (route: Route | null) => void;
  onRouteEdit?: (route: Route) => void;
  onRouteDelete?: (routeId: string) => void;
  onRouteCreate?: () => void;
  selectedRouteId?: string;
  isVisible?: boolean;
  readOnly?: boolean;
}

interface RouteListState {
  routes: Route[];
  loading: boolean;
  error: string | null;
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

export function RouteManagementPanel({
  className = '',
  onRouteSelect,
  onRouteEdit,
  onRouteDelete,
  onRouteCreate,
  selectedRouteId,
  isVisible = true,
  readOnly = false
}: RouteManagementPanelProps) {
  const [routeState, setRouteState] = useState<RouteListState>({
    routes: [],
    loading: false,
    error: null,
    totalCount: 0,
    currentPage: 1,
    pageSize: 10
  });

  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'distance'>('created');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [showRouteDetails, setShowRouteDetails] = useState<string | null>(null);
  const [editingRoute, setEditingRoute] = useState<Route | null>(null);

  // Load routes from backend
  const loadRoutes = useCallback(async () => {
    setRouteState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const params = {
        page: routeState.currentPage,
        limit: routeState.pageSize,
        search: searchQuery || undefined,
        sortBy: sortBy,
        sortOrder: sortOrder
      };
      
      const response = await mappingService.listRoutes(params);
      
      setRouteState(prev => ({
        ...prev,
        routes: response.routes,
        totalCount: response.total,
        loading: false
      }));
    } catch (error) {
      console.error('Failed to load routes:', error);
      setRouteState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load routes',
        loading: false
      }));
    }
  }, [routeState.currentPage, routeState.pageSize, searchQuery, sortBy, sortOrder]);

  // Load routes on mount and when dependencies change
  useEffect(() => {
    if (isVisible) {
      loadRoutes();
    }
  }, [isVisible, loadRoutes]);

  // Handle route selection
  const handleRouteSelect = useCallback((route: Route | null) => {
    onRouteSelect?.(route);
    setShowRouteDetails(route?.id || null);
  }, [onRouteSelect]);

  // Handle route deletion
  const handleDeleteRoute = useCallback(async (routeId: string) => {
    if (!confirm('Are you sure you want to delete this route?')) {
      return;
    }

    try {
      await mappingService.deleteRoute(routeId);
      onRouteDelete?.(routeId);
      await loadRoutes(); // Refresh list
      
      // Clear selection if deleted route was selected
      if (selectedRouteId === routeId) {
        handleRouteSelect(null);
      }
    } catch (error) {
      console.error('Failed to delete route:', error);
      alert('Failed to delete route. Please try again.');
    }
  }, [onRouteDelete, selectedRouteId, loadRoutes, handleRouteSelect]);

  // Handle route editing
  const handleEditRoute = useCallback((route: Route) => {
    setEditingRoute(route);
    onRouteEdit?.(route);
  }, [onRouteEdit]);

  // Handle route update
  const handleUpdateRoute = useCallback(async (updatedRoute: Partial<Route>) => {
    if (!editingRoute) return;

    try {
      const updated = await mappingService.updateRoute(editingRoute.id, updatedRoute);
      setEditingRoute(null);
      await loadRoutes(); // Refresh list
      
      // Update selection if this route is currently selected
      if (selectedRouteId === editingRoute.id) {
        handleRouteSelect(updated);
      }
    } catch (error) {
      console.error('Failed to update route:', error);
      alert('Failed to update route. Please try again.');
    }
  }, [editingRoute, selectedRouteId, loadRoutes, handleRouteSelect]);

  // Handle pagination
  const handlePageChange = useCallback((page: number) => {
    setRouteState(prev => ({ ...prev, currentPage: page }));
  }, []);

  // Calculate route statistics
  const getRouteStats = useCallback((route: Route) => {
    return {
      waypointCount: route.waypoints.length,
      distance: route.distance || 0,
      created: new Date(route.created).toLocaleDateString(),
      modified: new Date(route.modified).toLocaleDateString()
    };
  }, []);

  // Format route for display
  const formatRouteForDisplay = useCallback((route: Route) => {
    const stats = getRouteStats(route);
    return {
      ...route,
      displayName: route.name || `Route ${route.id.slice(-6)}`,
      formattedDistance: formatDistance(stats.distance),
      formattedCreated: stats.created,
      formattedModified: stats.modified,
      waypointCount: stats.waypointCount
    };
  }, [getRouteStats]);

  // Filter and sort routes
  const filteredAndSortedRoutes = React.useMemo(() => {
    let filtered = routeState.routes;

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(route => 
        route.name?.toLowerCase().includes(query) ||
        route.description?.toLowerCase().includes(query) ||
        route.id.toLowerCase().includes(query)
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      let aValue, bValue;
      
      switch (sortBy) {
        case 'name':
          aValue = a.name || a.id;
          bValue = b.name || b.id;
          break;
        case 'distance':
          aValue = a.distance || 0;
          bValue = b.distance || 0;
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

    return filtered.map(formatRouteForDisplay);
  }, [routeState.routes, searchQuery, sortBy, sortOrder, formatRouteForDisplay]);

  if (!isVisible) {
    return null;
  }

  return (
    <div className={`route-management-panel ${className}`}>
      <div className="route-panel-header">
        <h3>Route Management</h3>
        <div className="route-panel-actions">
          {!readOnly && (
            <button 
              className="btn-primary create-route-btn"
              onClick={onRouteCreate}
              title="Create new route"
            >
              ➕ New Route
            </button>
          )}
          <button 
            className="btn-secondary refresh-btn"
            onClick={loadRoutes}
            disabled={routeState.loading}
            title="Refresh routes"
          >
            🔄 Refresh
          </button>
        </div>
      </div>

      <div className="route-panel-controls">
        <div className="search-controls">
          <input
            type="text"
            placeholder="Search routes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="sort-controls">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'name' | 'created' | 'distance')}
            className="sort-select"
          >
            <option value="created">Sort by Created</option>
            <option value="name">Sort by Name</option>
            <option value="distance">Sort by Distance</option>
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

      <div className="route-panel-content">
        {routeState.error && (
          <div className="error-message">
            <span className="error-icon">⚠️</span>
            {routeState.error}
          </div>
        )}

        {routeState.loading ? (
          <div className="loading-message">
            <span className="loading-spinner">⏳</span>
            Loading routes...
          </div>
        ) : (
          <>
            <div className="routes-list">
              {filteredAndSortedRoutes.length === 0 ? (
                <div className="empty-state">
                  <span className="empty-icon">🗺️</span>
                  <p>No routes found</p>
                  {!readOnly && (
                    <button 
                      className="btn-primary"
                      onClick={onRouteCreate}
                    >
                      Create your first route
                    </button>
                  )}
                </div>
              ) : (
                filteredAndSortedRoutes.map((route) => (
                  <div
                    key={route.id}
                    className={`route-item ${selectedRouteId === route.id ? 'selected' : ''}`}
                    onClick={() => handleRouteSelect(route)}
                  >
                    <div className="route-item-header">
                      <div className="route-name">
                        {route.displayName}
                      </div>
                      <div className="route-actions">
                        {!readOnly && (
                          <>
                            <button
                              className="btn-icon edit-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleEditRoute(route);
                              }}
                              title="Edit route"
                            >
                              ✏️
                            </button>
                            <button
                              className="btn-icon delete-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteRoute(route.id);
                              }}
                              title="Delete route"
                            >
                              🗑️
                            </button>
                          </>
                        )}
                        <button
                          className="btn-icon details-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            setShowRouteDetails(
                              showRouteDetails === route.id ? null : route.id
                            );
                          }}
                          title="Show/hide details"
                        >
                          {showRouteDetails === route.id ? '🔼' : '🔽'}
                        </button>
                      </div>
                    </div>

                    <div className="route-item-summary">
                      <span className="route-stat">
                        📍 {route.waypointCount} waypoints
                      </span>
                      <span className="route-stat">
                        📏 {route.formattedDistance}
                      </span>
                      <span className="route-stat">
                        📅 {route.formattedCreated}
                      </span>
                    </div>

                    {route.description && (
                      <div className="route-description">
                        {route.description}
                      </div>
                    )}

                    {showRouteDetails === route.id && (
                      <div className="route-details">
                        <div className="route-detail-row">
                          <strong>ID:</strong> <code>{route.id}</code>
                        </div>
                        <div className="route-detail-row">
                          <strong>Created:</strong> {new Date(route.created).toLocaleString()}
                        </div>
                        <div className="route-detail-row">
                          <strong>Modified:</strong> {new Date(route.modified).toLocaleString()}
                        </div>
                        <div className="route-detail-row">
                          <strong>Waypoints:</strong>
                          <div className="waypoints-list">
                            {route.waypoints.map((waypoint, index) => (
                              <div key={index} className="waypoint-item">
                                <span className="waypoint-number">{index + 1}.</span>
                                <span className="waypoint-coords">
                                  {formatCoordinates(waypoint.lat, waypoint.lng, { precision: 4 })}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>

            {routeState.totalCount > routeState.pageSize && (
              <div className="pagination-controls">
                <button
                  className="btn-secondary"
                  onClick={() => handlePageChange(routeState.currentPage - 1)}
                  disabled={routeState.currentPage <= 1}
                >
                  ← Previous
                </button>
                
                <span className="pagination-info">
                  Page {routeState.currentPage} of {Math.ceil(routeState.totalCount / routeState.pageSize)}
                </span>
                
                <button
                  className="btn-secondary"
                  onClick={() => handlePageChange(routeState.currentPage + 1)}
                  disabled={routeState.currentPage >= Math.ceil(routeState.totalCount / routeState.pageSize)}
                >
                  Next →
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Route Edit Modal */}
      {editingRoute && (
        <RouteEditModal
          route={editingRoute}
          onSave={handleUpdateRoute}
          onCancel={() => setEditingRoute(null)}
        />
      )}
    </div>
  );
}

// Route Edit Modal Component
interface RouteEditModalProps {
  route: Route;
  onSave: (updatedRoute: Partial<Route>) => void;
  onCancel: () => void;
}

function RouteEditModal({ route, onSave, onCancel }: RouteEditModalProps) {
  const [name, setName] = useState(route.name || '');
  const [description, setDescription] = useState(route.description || '');

  const handleSave = () => {
    onSave({
      name: name.trim() || undefined,
      description: description.trim() || undefined
    });
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content route-edit-modal">
        <div className="modal-header">
          <h4>Edit Route</h4>
          <button className="btn-icon close-btn" onClick={onCancel}>
            ✖️
          </button>
        </div>
        
        <div className="modal-body">
          <div className="form-group">
            <label htmlFor="route-name">Route Name</label>
            <input
              id="route-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter route name..."
              maxLength={100}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="route-description">Description</label>
            <textarea
              id="route-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter route description..."
              rows={4}
              maxLength={500}
            />
          </div>

          <div className="route-info">
            <div className="info-item">
              <strong>Waypoints:</strong> {route.waypoints.length}
            </div>
            <div className="info-item">
              <strong>Distance:</strong> {formatDistance(route.distance || 0)}
            </div>
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
