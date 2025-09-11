import type { 
  EntityPosition, 
  PositionHistory, 
  PositionStatistics 
} from '../types/position';
import { POSITION_API_ENDPOINTS } from '../types/position';

export class PositionService {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
  }

  // Get all entity positions
  async getAllPositions(): Promise<EntityPosition[]> {
    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.ALL_POSITIONS}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch positions: ${response.statusText}`);
    }
    return response.json();
  }

  // Get only active (non-stale) positions
  async getActivePositions(): Promise<EntityPosition[]> {
    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.ACTIVE_POSITIONS}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch active positions: ${response.statusText}`);
    }
    return response.json();
  }

  // Get friendly positions only
  async getFriendlyPositions(): Promise<EntityPosition[]> {
    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.FRIENDLY_POSITIONS}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch friendly positions: ${response.statusText}`);
    }
    return response.json();
  }

  // Get hostile positions only
  async getHostilePositions(): Promise<EntityPosition[]> {
    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.HOSTILE_POSITIONS}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch hostile positions: ${response.statusText}`);
    }
    return response.json();
  }

  // Get positions within bounding box
  async getPositionsInBounds(
    north: number,
    south: number,
    east: number,
    west: number
  ): Promise<EntityPosition[]> {
    const params = new URLSearchParams({
      north: north.toString(),
      south: south.toString(),
      east: east.toString(),
      west: west.toString(),
    });

    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.POSITIONS_IN_BOUNDS}?${params}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch positions in bounds: ${response.statusText}`);
    }
    return response.json();
  }

  // Get specific entity position
  async getPosition(entityId: string): Promise<EntityPosition> {
    const url = `${this.baseUrl}${POSITION_API_ENDPOINTS.POSITION_BY_ID.replace(':entityId', entityId)}`;
    const response = await fetch(url);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Entity ${entityId} not found`);
      }
      throw new Error(`Failed to fetch position for ${entityId}: ${response.statusText}`);
    }
    return response.json();
  }

  // Get entity position trail/history
  async getPositionTrail(entityId: string): Promise<PositionHistory[]> {
    const url = `${this.baseUrl}${POSITION_API_ENDPOINTS.POSITION_TRAIL.replace(':entityId', entityId)}`;
    const response = await fetch(url);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Entity ${entityId} not found`);
      }
      throw new Error(`Failed to fetch trail for ${entityId}: ${response.statusText}`);
    }
    return response.json();
  }

  // Get position statistics
  async getPositionStatistics(): Promise<PositionStatistics> {
    const response = await fetch(`${this.baseUrl}${POSITION_API_ENDPOINTS.POSITION_STATISTICS}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch position statistics: ${response.statusText}`);
    }
    return response.json();
  }

  // Delete entity position
  async deletePosition(entityId: string): Promise<void> {
    const url = `${this.baseUrl}${POSITION_API_ENDPOINTS.POSITION_BY_ID.replace(':entityId', entityId)}`;
    const response = await fetch(url, {
      method: 'DELETE',
    });
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Entity ${entityId} not found`);
      }
      throw new Error(`Failed to delete position for ${entityId}: ${response.statusText}`);
    }
  }
}

// Default export for convenience
export const positionService = new PositionService();
