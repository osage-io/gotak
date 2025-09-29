/**
 * Simplified Tactical State Management Store
 * Basic Zustand store without Immer to avoid infinite loops
 */

import React from 'react';
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

// ===== CORE TYPES =====

export interface SystemStatus {
  connectionStatus: 'connected' | 'connecting' | 'disconnected' | 'error';
  serverLatency: number;
  lastSync: number;
  gpsStatus: 'active' | 'searching' | 'disabled' | 'error';
  batteryLevel?: number;
  networkQuality: 'excellent' | 'good' | 'fair' | 'poor' | 'offline';
  encryption: boolean;
  authentication: boolean;
}

export interface TacticalAlert {
  id: string;
  type: 'threat' | 'intel' | 'communication' | 'system' | 'mission';
  priority: 'low' | 'medium' | 'high' | 'critical' | 'immediate';
  title: string;
  message: string;
  source?: string;
  timestamp: number;
  acknowledged: boolean;
  dismissed: boolean;
}

export interface PerformanceMetrics {
  entityUpdateCount: number;
  lastUpdateTime: number;
  frameRate: number;
  memoryUsage?: number;
}

// ===== STORE STATE =====

interface SimpleTacticalState {
  // System State
  systemStatus: SystemStatus;
  
  // Core data
  entityCount: number;
  alerts: TacticalAlert[];
  
  // Performance tracking
  performance: PerformanceMetrics;
}

// ===== STORE ACTIONS =====

interface SimpleTacticalActions {
  // System Operations
  updateSystemStatus: (status: Partial<SystemStatus>) => void;
  updatePerformanceMetrics: (metrics: Partial<PerformanceMetrics>) => void;
  
  // Alerts & Notifications
  createAlert: (alert: Omit<TacticalAlert, 'id' | 'timestamp' | 'acknowledged' | 'dismissed'>) => void;
  acknowledgeAlert: (id: string) => void;
  dismissAlert: (id: string) => void;
  
  // Entity Management (simplified)
  setEntityCount: (count: number) => void;
  
  // Bulk Operations
  reset: () => void;
}

// ===== INITIAL STATE =====

const initialState: SimpleTacticalState = {
  systemStatus: {
    connectionStatus: 'disconnected',
    serverLatency: 0,
    lastSync: 0,
    gpsStatus: 'disabled',
    networkQuality: 'offline',
    encryption: false,
    authentication: false,
  },
  
  entityCount: 0,
  alerts: [],
  
  performance: {
    entityUpdateCount: 0,
    lastUpdateTime: Date.now(),
    frameRate: 60,
  },
};

// ===== STORE IMPLEMENTATION =====

export const useTacticalStore = create<SimpleTacticalState & SimpleTacticalActions>()(
  subscribeWithSelector((set, get) => ({
    ...initialState,
    
    // System Operations
    updateSystemStatus: (status) => set((state) => ({
      ...state,
      systemStatus: { ...state.systemStatus, ...status }
    })),
    
    updatePerformanceMetrics: (metrics) => set((state) => ({
      ...state,
      performance: { ...state.performance, ...metrics }
    })),
    
    // Alerts & Notifications
    createAlert: (alertData) => set((state) => {
      const id = `alert_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      const alert: TacticalAlert = {
        ...alertData,
        id,
        timestamp: Date.now(),
        acknowledged: false,
        dismissed: false,
      };
      return {
        ...state,
        alerts: [...state.alerts, alert]
      };
    }),
    
    acknowledgeAlert: (id) => set((state) => ({
      ...state,
      alerts: state.alerts.map(alert => 
        alert.id === id ? { ...alert, acknowledged: true } : alert
      )
    })),
    
    dismissAlert: (id) => set((state) => ({
      ...state,
      alerts: state.alerts.map(alert => 
        alert.id === id ? { ...alert, dismissed: true } : alert
      )
    })),
    
    // Entity Management (simplified)
    setEntityCount: (count) => set((state) => ({
      ...state,
      entityCount: count
    })),
    
    // Bulk Operations
    reset: () => set(() => initialState),
  }))
);

// ===== SELECTORS FOR OPTIMIZED ACCESS =====

export const useSystemStatus = () => useTacticalStore((state) => state.systemStatus);

// Memoized active alerts selector to prevent infinite re-renders
export const useActiveAlerts = () => {
  const alerts = useTacticalStore((state) => state.alerts);
  return React.useMemo(
    () => alerts.filter(alert => !alert.dismissed && !alert.acknowledged),
    [alerts]
  );
};

export const useEntityCount = () => useTacticalStore((state) => state.entityCount);
export const usePerformanceMetrics = () => useTacticalStore((state) => state.performance);

// Export store for direct access
export default useTacticalStore;
