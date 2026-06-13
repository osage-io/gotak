import { authHeaders } from './authToken';
// API service for GoTAK Web UI

// Get configuration from runtime config
const getConfig = () => {
  return (window as any).GOTAK_CONFIG || {
    apiUrl: 'http://localhost:8082',
    wsUrl: 'ws://localhost:8087/ws',
    serverUrl: 'ws://localhost:8087'
  };
};

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export class ApiClient {
  private baseUrl: string;

  constructor() {
    const config = getConfig();
    this.baseUrl = config.apiUrl;
  }

  async request<T = any>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    // Don't add /api/v1 prefix here since endpoints already include it
    const url = `${this.baseUrl}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders(),
          ...options.headers,
        },
      });

      if (!response.ok) {
        return {
          success: false,
          error: `HTTP ${response.status}: ${response.statusText}`,
        };
      }

      const data = await response.json();
      return {
        success: true,
        data,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error occurred',
      };
    }
  }

  // GET request
  async get<T = any>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  // POST request
  async post<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // PUT request
  async put<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // DELETE request
  async delete<T = any>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

// Default API client instance
export const apiClient = new ApiClient();

// Health check endpoint
export const checkHealth = async (): Promise<boolean> => {
  try {
    const response = await apiClient.get('/health');
    return response.success;
  } catch {
    return false;
  }
};

// Export common API functions
export const api = {
  // Entity/Position APIs
  getEntities: () => apiClient.get('/api/v1/entities'),
  getEntity: (id: string) => apiClient.get(`/api/v1/entities/${id}`),
  
  // Route APIs
  getRoutes: () => apiClient.get('/api/v1/routes'),
  createRoute: (route: any) => apiClient.post('/api/v1/routes', route),
  updateRoute: (id: string, route: any) => apiClient.put(`/api/v1/routes/${id}`, route),
  deleteRoute: (id: string) => apiClient.delete(`/api/v1/routes/${id}`),
  
  // Geofence APIs
  getGeofences: () => apiClient.get('/api/v1/geofences'),
  createGeofence: (geofence: any) => apiClient.post('/api/v1/geofences', geofence),
  updateGeofence: (id: string, geofence: any) => apiClient.put(`/api/v1/geofences/${id}`, geofence),
  deleteGeofence: (id: string) => apiClient.delete(`/api/v1/geofences/${id}`),
  
  // Chat APIs
  getChatRooms: () => apiClient.get('/api/v1/chat/rooms'),
  getChatMessages: (roomId: string) => apiClient.get(`/api/v1/chat/rooms/${roomId}/messages`),
  sendChatMessage: (roomId: string, message: any) => apiClient.post(`/api/v1/chat/rooms/${roomId}/messages`, message),
  
  // System APIs
  getServerInfo: () => apiClient.get('/api/v1/info'),
  getMetrics: () => apiClient.get('/api/v1/metrics'),
};

export default api;
