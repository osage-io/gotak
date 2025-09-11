import { useState, useEffect, useCallback } from 'react';
import type { Entity } from '../types/entity';
import { mockEntities, mockWebSocketService } from '../services/mockData';

// Environment variables
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true';
const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1';
const DEBUG = import.meta.env.VITE_DEBUG === 'true';

interface UsePositionsResult {
  entities: Entity[];
  loading: boolean;
  error: string | null;
  refreshPositions: () => void;
  getEntityById: (id: string) => Entity | undefined;
  getEntitiesByAffiliation: (affiliation: Entity['affiliation']) => Entity[];
}

export const usePositions = (enableRealTime: boolean = true): UsePositionsResult => {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch initial positions
  const fetchPositions = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      if (USE_MOCK) {
        // Use mock data
        await new Promise(resolve => setTimeout(resolve, 300));
        setEntities([...mockEntities]);
        if (DEBUG) console.log('Loaded mock entities:', mockEntities.length);
      } else {
        // Fetch from real API
        const response = await fetch(`${API_BASE}/entities`);
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        const entitiesData = data.data?.entities || data.entities || [];
        setEntities(entitiesData);
        if (DEBUG) console.log('Loaded real entities:', entitiesData.length);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(errorMessage);
      console.error('Error fetching positions:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Process WebSocket position updates
  const handlePositionUpdate = useCallback((message: any) => {
    if (message.type === 'position_update' && message.payload) {
      const { entityId, position } = message.payload;
      console.log('Position update received:', entityId, position);
      
      setEntities(prevEntities => {
        const existingIndex = prevEntities.findIndex(e => e.id === entityId);
        
        if (existingIndex >= 0) {
          // Update existing entity
          const updated = [...prevEntities];
          updated[existingIndex] = {
            ...updated[existingIndex],
            lat: position.lat,
            lng: position.lng,
            altitude: position.altitude,
            speed: position.speed,
            course: position.course,
            lastUpdate: position.timestamp ? new Date(position.timestamp) : new Date(),
          };
          return updated;
        } else {
          // Add new entity (this would need more complete entity data in practice)
          const newEntity: Entity = {
            id: entityId,
            callsign: `Unknown-${entityId.slice(-4)}`,
            type: 'unknown',
            affiliation: 'unknown',
            lat: position.lat,
            lng: position.lng,
            altitude: position.altitude,
            speed: position.speed,
            course: position.course,
            lastUpdate: position.timestamp ? new Date(position.timestamp) : new Date(),
            classification: 'unclassified',
            status: 'active',
          };
          return [...prevEntities, newEntity];
        }
      });
    }
  }, []);

  // Set up WebSocket when real-time is enabled
  useEffect(() => {
    if (enableRealTime) {
      if (USE_MOCK) {
        mockWebSocketService.onMessage(handlePositionUpdate);
        mockWebSocketService.connect();
        
        return () => {
          mockWebSocketService.disconnect();
        };
      } else {
        // Real WebSocket will be handled by the useWebSocket hook
        if (DEBUG) console.log('Real-time positioning enabled with real WebSocket');
      }
    }
  }, [enableRealTime, handlePositionUpdate]);

  // Load positions on mount
  useEffect(() => {
    fetchPositions();
  }, [fetchPositions]);

  // Helper functions
  const getEntityById = useCallback((id: string): Entity | undefined => {
    return entities.find(entity => entity.id === id);
  }, [entities]);

  const getEntitiesByAffiliation = useCallback((affiliation: Entity['affiliation']): Entity[] => {
    return entities.filter(entity => entity.affiliation === affiliation);
  }, [entities]);

  const refreshPositions = useCallback(() => {
    if (!enableRealTime) {
      fetchPositions();
    }
  }, [enableRealTime, fetchPositions]);

  return {
    entities,
    loading,
    error,
    refreshPositions,
    getEntityById,
    getEntitiesByAffiliation,
  };
};
