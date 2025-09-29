/**
 * Entity Tracking Store
 * Manages tactical entities with real-time updates from the GoTAK server
 */

import { Entity, PositionPoint } from '../services/apiClient';
import { wsService, PositionUpdate } from '../services/websocketService';

// Entity state management
export interface EntityState {
  entities: Map<string, Entity>;
  selectedEntityId: string | null;
  filterType: EntityFilterType;
  lastUpdate: number;
}

export type EntityFilterType = 'all' | 'friendly' | 'hostile' | 'unknown' | 'selected';

// Entity tracking class
export class EntityTracker {
  private state: EntityState = {
    entities: new Map(),
    selectedEntityId: null,
    filterType: 'all',
    lastUpdate: Date.now()
  };

  private subscribers: Set<() => void> = new Set();
  private wsUnsubscribers: (() => void)[] = [];

  constructor() {
    this.setupWebSocketListeners();
  }

  // ===== STATE MANAGEMENT =====

  getState(): EntityState {
    return { ...this.state, entities: new Map(this.state.entities) };
  }

  subscribe(callback: () => void): () => void {
    this.subscribers.add(callback);
    return () => this.subscribers.delete(callback);
  }

  private notify(): void {
    this.state.lastUpdate = Date.now();
    this.subscribers.forEach(callback => {
      try {
        callback();
      } catch (error) {
        console.error('Error in EntityTracker subscriber:', error);
      }
    });
  }

  // ===== WEBSOCKET INTEGRATION =====

  private setupWebSocketListeners(): void {
    // Listen for position updates
    const unsubscribePositionUpdate = wsService.onPositionUpdate((update: PositionUpdate) => {
      this.updateEntityPosition(update.entityId, update.position);
    });
    
    // Listen for entity removal
    const unsubscribeEntityRemoved = wsService.onEntityRemoved(({ entityId }: { entityId: string }) => {
      this.removeEntity(entityId);
    });

    this.wsUnsubscribers.push(unsubscribePositionUpdate, unsubscribeEntityRemoved);
  }

  // ===== ENTITY OPERATIONS =====

  addEntity(entity: Entity): void {
    this.state.entities.set(entity.id, { ...entity });
    this.notify();
  }

  addEntities(entities: Entity[]): void {
    entities.forEach(entity => {
      this.state.entities.set(entity.id, { ...entity });
    });
    this.notify();
  }

  updateEntity(entityId: string, updates: Partial<Entity>): void {
    const existing = this.state.entities.get(entityId);
    if (existing) {
      this.state.entities.set(entityId, { ...existing, ...updates });
      this.notify();
    }
  }

  updateEntityPosition(entityId: string, position: PositionPoint): void {
    const existing = this.state.entities.get(entityId);
    if (existing) {
      this.state.entities.set(entityId, {
        ...existing,
        position,
        lastUpdate: new Date().toISOString()
      });
      this.notify();
    }
  }

  removeEntity(entityId: string): void {
    if (this.state.entities.delete(entityId)) {
      if (this.state.selectedEntityId === entityId) {
        this.state.selectedEntityId = null;
      }
      this.notify();
    }
  }

  clearEntities(): void {
    this.state.entities.clear();
    this.state.selectedEntityId = null;
    this.notify();
  }

  // ===== SELECTION =====

  selectEntity(entityId: string | null): void {
    this.state.selectedEntityId = entityId;
    this.notify();
  }

  getSelectedEntity(): Entity | null {
    return this.state.selectedEntityId 
      ? this.state.entities.get(this.state.selectedEntityId) || null
      : null;
  }

  // ===== FILTERING =====

  setFilter(filterType: EntityFilterType): void {
    this.state.filterType = filterType;
    this.notify();
  }

  getFilteredEntities(): Entity[] {
    const entities = Array.from(this.state.entities.values());
    
    switch (this.state.filterType) {
      case 'friendly':
        return entities.filter(e => e.entityType.startsWith('a-f'));
      case 'hostile':
        return entities.filter(e => e.entityType.startsWith('a-h'));
      case 'unknown':
        return entities.filter(e => e.entityType.startsWith('a-u'));
      case 'selected':
        return this.state.selectedEntityId 
          ? entities.filter(e => e.id === this.state.selectedEntityId)
          : [];
      case 'all':
      default:
        return entities;
    }
  }

  // ===== QUERIES =====

  getEntity(entityId: string): Entity | undefined {
    return this.state.entities.get(entityId);
  }

  getEntitiesInBounds(bounds: {
    north: number;
    south: number;
    east: number;
    west: number;
  }): Entity[] {
    return Array.from(this.state.entities.values()).filter(entity => {
      if (!entity.position) return false;
      const { lat, lng } = entity.position;
      return lat >= bounds.south && 
             lat <= bounds.north && 
             lng >= bounds.west && 
             lng <= bounds.east;
    });
  }

  getEntitiesByType(entityType: string): Entity[] {
    return Array.from(this.state.entities.values())
      .filter(entity => entity.entityType === entityType);
  }

  getEntitiesByCallsign(callsign: string): Entity[] {
    return Array.from(this.state.entities.values())
      .filter(entity => entity.callsign?.toLowerCase().includes(callsign.toLowerCase()));
  }

  // ===== STATISTICS =====

  getEntityCounts(): Record<string, number> {
    const entities = Array.from(this.state.entities.values());
    return {
      total: entities.length,
      friendly: entities.filter(e => e.entityType.startsWith('a-f')).length,
      hostile: entities.filter(e => e.entityType.startsWith('a-h')).length,
      unknown: entities.filter(e => e.entityType.startsWith('a-u')).length,
    };
  }

  getOldestEntity(): Entity | null {
    const entities = Array.from(this.state.entities.values());
    if (entities.length === 0) return null;
    
    return entities.reduce((oldest, current) => {
      const oldestTime = new Date(oldest.lastUpdate).getTime();
      const currentTime = new Date(current.lastUpdate).getTime();
      return currentTime < oldestTime ? current : oldest;
    });
  }

  getNewestEntity(): Entity | null {
    const entities = Array.from(this.state.entities.values());
    if (entities.length === 0) return null;
    
    return entities.reduce((newest, current) => {
      const newestTime = new Date(newest.lastUpdate).getTime();
      const currentTime = new Date(current.lastUpdate).getTime();
      return currentTime > newestTime ? current : newest;
    });
  }

  // ===== CLEANUP =====

  destroy(): void {
    this.wsUnsubscribers.forEach(unsubscribe => unsubscribe());
    this.wsUnsubscribers = [];
    this.subscribers.clear();
  }
}

// Export singleton instance
export const entityTracker = new EntityTracker();
export default entityTracker;
