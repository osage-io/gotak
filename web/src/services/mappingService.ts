import { api } from './api';

// Route types
export interface Point {
  lat: number;
  lng: number;
}

export interface Waypoint {
  id: string;
  route_id: string;
  sequence: number;
  lat: number;
  lng: number;
  name: string;
  description?: string;
  eta?: string;
  created_at: string;
}

export interface Route {
  id: string;
  name: string;
  description: string;
  created_by: string;
  group_id: string;
  waypoints: Waypoint[];
  geometry: {
    type: 'LineString';
    coordinates: number[][];
  };
  distance: number;
  duration: number;
  route_type: 'fastest' | 'shortest' | 'tactical' | 'offroad';
  vehicle: 'car' | 'truck' | 'bicycle' | 'foot' | 'motorcycle';
  optimize: boolean;
  created_at: string;
  updated_at: string;
}

export interface RouteOptions {
  route_type: 'fastest' | 'shortest' | 'tactical' | 'offroad';
  vehicle: 'car' | 'truck' | 'bicycle' | 'foot' | 'motorcycle';
  optimize: boolean;
  avoid_tolls?: boolean;
  avoid_highways?: boolean;
}

export interface CreateRouteRequest {
  name: string;
  description?: string;
  waypoints: Point[];
  options: RouteOptions;
}

// Geofence types
export interface Geofence {
  id: string;
  name: string;
  description: string;
  type: 'circle' | 'polygon' | 'rectangle';
  geometry: any; // GeoJSON-like geometry
  enabled: boolean;
  alert_on_enter: boolean;
  alert_on_exit: boolean;
  created_by: string;
  group_id: string;
  created_at: string;
  updated_at: string;
}

export interface GeofenceViolation {
  id: string;
  geofence_id: string;
  entity_id: string;
  violation_type: 'enter' | 'exit';
  position: Point;
  timestamp: string;
  acknowledged: boolean;
  acknowledged_by?: string;
  acknowledged_at?: string;
}

export interface CreateGeofenceRequest {
  name: string;
  description?: string;
  type: 'circle' | 'polygon' | 'rectangle';
  geometry: any;
  alert_on_enter: boolean;
  alert_on_exit: boolean;
  enabled: boolean;
}

// Offline area types
export interface BoundingBox {
  north: number;
  south: number;
  east: number;
  west: number;
}

export interface OfflineArea {
  id: string;
  name: string;
  bounds: BoundingBox;
  min_zoom: number;
  max_zoom: number;
  layers: string[];
  status: 'pending' | 'downloading' | 'complete' | 'error';
  progress: number;
  size_mb: number;
  created_at: string;
  updated_at: string;
}

export interface CreateOfflineAreaRequest {
  name: string;
  bounds: BoundingBox;
  min_zoom: number;
  max_zoom: number;
  layers: string[];
}

export interface DownloadProgress {
  area_id: string;
  total_tiles: number;
  completed_tiles: number;
  failed_tiles: number;
  start_time: string;
  estimated_time?: number;
}

export class MappingService {
  private baseUrl = '/api/v1/mapping';

  // Route management
  async createRoute(request: CreateRouteRequest): Promise<Route> {
    const response = await api.post(`${this.baseUrl}/routes`, request);
    return response.data;
  }

  async getRoute(routeId: string): Promise<Route> {
    const response = await api.get(`${this.baseUrl}/routes/${routeId}`);
    return response.data;
  }

  async listRoutes(params?: { limit?: number; offset?: number }): Promise<{ routes: Route[] }> {
    const response = await api.get(`${this.baseUrl}/routes`, { params });
    return response.data;
  }

  async updateRoute(routeId: string, updates: Partial<Route>): Promise<Route> {
    const response = await api.put(`${this.baseUrl}/routes/${routeId}`, updates);
    return response.data;
  }

  async deleteRoute(routeId: string): Promise<void> {
    await api.delete(`${this.baseUrl}/routes/${routeId}`);
  }

  async recalculateRoute(routeId: string, options: RouteOptions): Promise<Route> {
    const response = await api.post(`${this.baseUrl}/routes/${routeId}/recalculate`, options);
    return response.data;
  }

  // Geofence management
  async createGeofence(request: CreateGeofenceRequest): Promise<Geofence> {
    const response = await api.post(`${this.baseUrl}/geofences`, request);
    return response.data;
  }

  async getGeofence(geofenceId: string): Promise<Geofence> {
    const response = await api.get(`${this.baseUrl}/geofences/${geofenceId}`);
    return response.data;
  }

  async listGeofences(params?: { limit?: number; offset?: number }): Promise<{ geofences: Geofence[] }> {
    const response = await api.get(`${this.baseUrl}/geofences`, { params });
    return response.data;
  }

  async updateGeofence(geofenceId: string, updates: Partial<Geofence>): Promise<Geofence> {
    const response = await api.put(`${this.baseUrl}/geofences/${geofenceId}`, updates);
    return response.data;
  }

  async deleteGeofence(geofenceId: string): Promise<void> {
    await api.delete(`${this.baseUrl}/geofences/${geofenceId}`);
  }

  async getViolations(params?: {
    geofence_id?: string;
    entity_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ violations: GeofenceViolation[] }> {
    const response = await api.get(`${this.baseUrl}/geofences/violations`, { params });
    return response.data;
  }

  async acknowledgeViolation(violationId: string): Promise<void> {
    await api.put(`${this.baseUrl}/geofences/violations/${violationId}/acknowledge`);
  }

  // Offline map management
  async createOfflineArea(request: CreateOfflineAreaRequest): Promise<OfflineArea> {
    const response = await api.post(`${this.baseUrl}/offline/areas`, request);
    return response.data;
  }

  async getOfflineArea(areaId: string): Promise<OfflineArea> {
    const response = await api.get(`${this.baseUrl}/offline/areas/${areaId}`);
    return response.data;
  }

  async listOfflineAreas(params?: { limit?: number; offset?: number }): Promise<{ areas: OfflineArea[] }> {
    const response = await api.get(`${this.baseUrl}/offline/areas`, { params });
    return response.data;
  }

  async deleteOfflineArea(areaId: string): Promise<void> {
    await api.delete(`${this.baseUrl}/offline/areas/${areaId}`);
  }

  async getDownloadProgress(areaId: string): Promise<DownloadProgress> {
    const response = await api.get(`${this.baseUrl}/offline/areas/${areaId}/progress`);
    return response.data;
  }

  async getCachedTile(layer: string, z: number, x: number, y: number): Promise<Blob> {
    const response = await api.get(
      `${this.baseUrl}/offline/tiles/${layer}/${z}/${x}/${y}`,
      { responseType: 'blob' }
    );
    return response.data;
  }
}

export const mappingService = new MappingService();
