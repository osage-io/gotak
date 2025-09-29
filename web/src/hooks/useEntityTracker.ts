/**
 * React Hooks for Entity Tracking
 * Provides React integration for the EntityTracker store
 */

import { useState, useEffect, useCallback } from 'react';
import { entityTracker, EntityState, EntityFilterType } from '../stores/entityTracker';
import { Entity } from '../services/apiClient';
import { apiClient } from '../services/apiClient';

// ===== MAIN ENTITY TRACKER HOOK =====

export function useEntityTracker(): {
  state: EntityState;
  actions: {
    selectEntity: (entityId: string | null) => void;
    setFilter: (filter: EntityFilterType) => void;
    refreshEntities: () => Promise<void>;
    clearEntities: () => void;
  };
} {
  const [state, setState] = useState<EntityState>(entityTracker.getState());

  useEffect(() => {
    const unsubscribe = entityTracker.subscribe(() => {
      setState(entityTracker.getState());
    });
    return unsubscribe;
  }, []);

  const refreshEntities = useCallback(async () => {
    try {
      const entities = await apiClient.getEntities();
      entityTracker.addEntities(entities);
    } catch (error) {
      console.error('Failed to refresh entities:', error);
    }
  }, []);

  const actions = {
    selectEntity: entityTracker.selectEntity.bind(entityTracker),
    setFilter: entityTracker.setFilter.bind(entityTracker),
    refreshEntities,
    clearEntities: entityTracker.clearEntities.bind(entityTracker),
  };

  return { state, actions };
}

// ===== FILTERED ENTITIES HOOK =====

export function useFilteredEntities(): {
  entities: Entity[];
  counts: Record<string, number>;
  filter: EntityFilterType;
  setFilter: (filter: EntityFilterType) => void;
} {
  const { state, actions } = useEntityTracker();

  const entities = entityTracker.getFilteredEntities();
  const counts = entityTracker.getEntityCounts();

  return {
    entities,
    counts,
    filter: state.filterType,
    setFilter: actions.setFilter,
  };
}

// ===== SELECTED ENTITY HOOK =====

export function useSelectedEntity(): {
  selectedEntity: Entity | null;
  selectEntity: (entityId: string | null) => void;
  clearSelection: () => void;
} {
  const { actions } = useEntityTracker();
  const [selectedEntity, setSelectedEntity] = useState<Entity | null>(
    entityTracker.getSelectedEntity()
  );

  useEffect(() => {
    const unsubscribe = entityTracker.subscribe(() => {
      setSelectedEntity(entityTracker.getSelectedEntity());
    });
    return unsubscribe;
  }, []);

  const clearSelection = useCallback(() => {
    actions.selectEntity(null);
  }, [actions]);

  return {
    selectedEntity,
    selectEntity: actions.selectEntity,
    clearSelection,
  };
}

// ===== ENTITY QUERIES HOOK =====

export function useEntityQueries(): {
  getEntity: (entityId: string) => Entity | undefined;
  getEntitiesInBounds: (bounds: {
    north: number;
    south: number;
    east: number;
    west: number;
  }) => Entity[];
  getEntitiesByType: (entityType: string) => Entity[];
  getEntitiesByCallsign: (callsign: string) => Entity[];
} {
  return {
    getEntity: entityTracker.getEntity.bind(entityTracker),
    getEntitiesInBounds: entityTracker.getEntitiesInBounds.bind(entityTracker),
    getEntitiesByType: entityTracker.getEntitiesByType.bind(entityTracker),
    getEntitiesByCallsign: entityTracker.getEntitiesByCallsign.bind(entityTracker),
  };
}

// ===== ENTITY STATISTICS HOOK =====

export function useEntityStats(): {
  counts: Record<string, number>;
  oldestEntity: Entity | null;
  newestEntity: Entity | null;
  lastUpdate: number;
} {
  const { state } = useEntityTracker();
  const [stats, setStats] = useState(() => ({
    counts: entityTracker.getEntityCounts(),
    oldestEntity: entityTracker.getOldestEntity(),
    newestEntity: entityTracker.getNewestEntity(),
    lastUpdate: state.lastUpdate,
  }));

  useEffect(() => {
    setStats({
      counts: entityTracker.getEntityCounts(),
      oldestEntity: entityTracker.getOldestEntity(),
      newestEntity: entityTracker.getNewestEntity(),
      lastUpdate: state.lastUpdate,
    });
  }, [state.lastUpdate]);

  return stats;
}

// ===== ENTITY REAL-TIME UPDATES HOOK =====

export function useEntityUpdates(entityId?: string): {
  entity: Entity | undefined;
  isUpdating: boolean;
  lastUpdate: string | null;
} {
  const [entity, setEntity] = useState<Entity | undefined>(
    entityId ? entityTracker.getEntity(entityId) : undefined
  );
  const [isUpdating, setIsUpdating] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<string | null>(null);

  useEffect(() => {
    if (!entityId) return;

    let updateTimeout: number;

    const unsubscribe = entityTracker.subscribe(() => {
      const updatedEntity = entityTracker.getEntity(entityId);
      if (updatedEntity && (!entity || updatedEntity.lastUpdate !== entity.lastUpdate)) {
        setEntity(updatedEntity);
        setIsUpdating(true);
        setLastUpdate(updatedEntity.lastUpdate);

        // Clear updating state after a short delay
        if (updateTimeout) clearTimeout(updateTimeout);
        updateTimeout = window.setTimeout(() => {
          setIsUpdating(false);
        }, 1000);
      }
    });

    return () => {
      unsubscribe();
      if (updateTimeout) clearTimeout(updateTimeout);
    };
  }, [entityId, entity]);

  return { entity, isUpdating, lastUpdate };
}

// ===== ENTITY PERSISTENCE HOOK =====

export function useEntityPersistence(): {
  saveEntity: (entity: Entity) => Promise<void>;
  deleteEntity: (entityId: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
} {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const saveEntity = useCallback(async (entity: Entity) => {
    setIsLoading(true);
    setError(null);
    try {
      if (entityTracker.getEntity(entity.id)) {
        await apiClient.updateEntity(entity.id, entity);
      } else {
        await apiClient.createEntity(entity);
      }
      entityTracker.addEntity(entity);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save entity';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const deleteEntity = useCallback(async (entityId: string) => {
    setIsLoading(true);
    setError(null);
    try {
      await apiClient.deleteEntity(entityId);
      entityTracker.removeEntity(entityId);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete entity';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { saveEntity, deleteEntity, isLoading, error };
}

// ===== BULK ENTITY OPERATIONS HOOK =====

export function useBulkEntityOperations(): {
  loadInitialEntities: () => Promise<void>;
  refreshFromServer: () => Promise<void>;
  exportEntities: (filter?: EntityFilterType) => void;
  isLoading: boolean;
  error: string | null;
} {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadInitialEntities = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const entities = await apiClient.getEntities();
      entityTracker.clearEntities();
      entityTracker.addEntities(entities);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load entities';
      setError(errorMessage);
      console.error('Failed to load initial entities:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const refreshFromServer = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const entities = await apiClient.getEntities();
      // Update existing entities and add new ones
      entities.forEach(entity => {
        entityTracker.addEntity(entity);
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to refresh entities';
      setError(errorMessage);
      console.error('Failed to refresh entities from server:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const exportEntities = useCallback((filter?: EntityFilterType) => {
    try {
      const originalFilter = entityTracker.getState().filterType;
      if (filter && filter !== originalFilter) {
        entityTracker.setFilter(filter);
      }
      
      const entities = entityTracker.getFilteredEntities();
      const dataStr = JSON.stringify(entities, null, 2);
      const dataBlob = new Blob([dataStr], { type: 'application/json' });
      
      const link = document.createElement('a');
      link.href = URL.createObjectURL(dataBlob);
      link.download = `entities-${filter || 'all'}-${new Date().toISOString().split('T')[0]}.json`;
      link.click();
      
      // Restore original filter
      if (filter && filter !== originalFilter) {
        entityTracker.setFilter(originalFilter);
      }
    } catch (err) {
      console.error('Failed to export entities:', err);
      setError('Failed to export entities');
    }
  }, []);

  return { loadInitialEntities, refreshFromServer, exportEntities, isLoading, error };
}
