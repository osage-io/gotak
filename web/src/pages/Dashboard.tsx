/**
 * Dashboard Page - Tactical Command Overview
 * Main operational dashboard for GoTAK tactical awareness system
 */

import React, { useEffect, useState, useRef } from 'react';
import EntityMap from '../components/Map/EntityMap';
import { useEntityStats, useBulkEntityOperations } from '../hooks/useEntityTracker';
import { wsService, ConnectionState } from '../services/websocketService';
import { Entity } from '../services/apiClient';
import { Icon } from '../components/ui/Icon';

// Dashboard stats interface
interface DashboardStats {
  totalEntities: number;
  friendly: number;
  hostile: number;
  unknown: number;
  connectionStatus: ConnectionState;
  lastUpdate: string;
}

const Dashboard: React.FC = () => {
  const [currentTime, setCurrentTime] = useState(new Date());
  const [alerts, setAlerts] = useState<Array<{
    id: string;
    type: 'info' | 'warning' | 'error' | 'critical';
    message: string;
    timestamp: string;
  }>>([]);

  // Entity tracking
  const { counts, lastUpdate } = useEntityStats();
  const { loadInitialEntities, isLoading } = useBulkEntityOperations();

  // WebSocket connection status
  const [connectionStatus, setConnectionStatus] = useState<ConnectionState>(
    wsService.connectionState
  );

  // Update current time every second
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // Monitor WebSocket connection
  useEffect(() => {
    const updateConnectionStatus = () => {
      setConnectionStatus(wsService.connectionState);
    };

    const unsubscribeConnection = wsService.onConnection(updateConnectionStatus);
    const unsubscribeDisconnection = wsService.onDisconnection(updateConnectionStatus);
    const unsubscribeError = wsService.onError(updateConnectionStatus);

    // Listen for system alerts
    const unsubscribeAlerts = wsService.onSystemAlert((alert) => {
      setAlerts(prev => [
        {
          id: alert.id,
          type: alert.type,
          message: alert.message,
          timestamp: alert.timestamp,
        },
        ...prev.slice(0, 9) // Keep only last 10 alerts
      ]);
    });

    return () => {
      unsubscribeConnection();
      unsubscribeDisconnection();
      unsubscribeError();
      unsubscribeAlerts();
    };
  }, []);

  // Initialize WebSocket connection and load entities
  useEffect(() => {
    const initializeDashboard = async () => {
      try {
        // Connect WebSocket
        await wsService.connect();
        // Load initial entities
        await loadInitialEntities();
      } catch (error) {
        console.error('Failed to initialize dashboard:', error);
      }
    };

    initializeDashboard();
  }, [loadInitialEntities]);

  // Format connection status
  const getConnectionStatusText = (): string => {
    switch (connectionStatus) {
      case ConnectionState.CONNECTED:
        return 'ONLINE';
      case ConnectionState.CONNECTING:
        return 'CONNECTING';
      case ConnectionState.RECONNECTING:
        return 'RECONNECTING';
      case ConnectionState.DISCONNECTED:
        return 'OFFLINE';
      case ConnectionState.ERROR:
        return 'ERROR';
      default:
        return 'UNKNOWN';
    }
  };

  const getConnectionStatusColor = (): string => {
    switch (connectionStatus) {
      case ConnectionState.CONNECTED:
        return 'var(--color-success)';
      case ConnectionState.CONNECTING:
      case ConnectionState.RECONNECTING:
        return 'var(--color-warning)';
      case ConnectionState.DISCONNECTED:
      case ConnectionState.ERROR:
        return 'var(--color-error)';
      default:
        return 'var(--color-neutral)';
    }
  };

  return (
    <div className="dashboard-container">
      {/* Dashboard Header */}
      <header className="dashboard-header">
        <div className="header-left">
          <h1 className="font-display font-bold text-2xl tracking-tight text-primary">
            GoTAK DASHBOARD
          </h1>
          <div className="header-time font-mono text-sm text-secondary">
            {currentTime.toLocaleString('en-US', {
              weekday: 'short',
              year: 'numeric',
              month: 'short',
              day: '2-digit',
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
              timeZoneName: 'short'
            })}
          </div>
        </div>

        <div className="header-right">
          <div className="connection-status-indicator">
            <Icon 
              name={connectionStatus === ConnectionState.CONNECTED ? "signal" : "broadcast"}
              size={16}
              color={getConnectionStatusColor()}
            />
            <span className="font-mono text-sm font-medium uppercase tracking-wide">
              {getConnectionStatusText()}
            </span>
          </div>
        </div>
      </header>

      {/* Main Dashboard Content */}
      <div className="dashboard-content">
        {/* Stats Cards */}
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-value" style={{ color: 'var(--color-success)' }}>
              {counts.friendly}
            </div>
            <div className="stat-label text-secondary font-medium text-sm uppercase tracking-wide">
              Friendly Forces
            </div>
            <div className="stat-icon">
              <Icon name="shield" size={24} color="var(--color-success)" />
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-value" style={{ color: 'var(--color-error)' }}>
              {counts.hostile}
            </div>
            <div className="stat-label text-secondary font-medium text-sm uppercase tracking-wide">
              Hostile Forces
            </div>
            <div className="stat-icon">
              <Icon name="warning" size={24} color="var(--color-error)" />
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-value" style={{ color: 'var(--color-warning)' }}>
              {counts.unknown}
            </div>
            <div className="stat-label text-secondary font-medium text-sm uppercase tracking-wide">
              Unknown Contacts
            </div>
            <div className="stat-icon">
              <Icon name="target" size={24} color="var(--color-warning)" />
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-value" style={{ color: 'var(--color-accent)' }}>
              {counts.total}
            </div>
            <div className="stat-label text-secondary font-medium text-sm uppercase tracking-wide">
              Total Entities
            </div>
            <div className="stat-icon">
              <Icon name="broadcast" size={24} color="var(--color-accent)" />
            </div>
          </div>
        </div>

        {/* Main Content Grid */}
        <div className="main-grid">
          {/* Map Section */}
          <div className="map-section">
            <div className="section-header">
              <div className="section-title-wrapper">
                <Icon name="map" size={20} color="var(--color-accent)" />
                <h2 className="section-title font-display font-semibold text-xl text-primary">
                  Tactical Map
                </h2>
              </div>
              {isLoading && (
                <div className="loading-indicator text-warning font-mono text-sm">
                  <Icon name="sync" size={14} color="var(--color-warning)" className="spinning" />
                  Loading entities...
                </div>
              )}
            </div>
            <div className="map-container">
              <EntityMap
                style={{ height: '500px' }}
                showEntityLabels={true}
                showEntityTrails={false}
                enableClustering={true}
              />
            </div>
          </div>

          {/* Alerts Panel */}
          <div className="alerts-section">
            <div className="section-header">
              <div className="section-title-wrapper">
                <Icon name="alert" size={20} color="var(--color-accent)" />
                <h2 className="section-title font-display font-semibold text-xl text-primary">
                  System Alerts
                </h2>
              </div>
              <div className="alert-count">
                <Icon name="bell" size={12} color="var(--color-text-secondary)" />
                {alerts.length} Active
              </div>
            </div>
            <div className="alerts-container">
              {alerts.length === 0 ? (
                <div className="no-alerts text-muted text-center font-medium">
                  No active alerts
                </div>
              ) : (
                alerts.map((alert) => (
                  <div key={alert.id} className={`alert-item alert-${alert.type}`}>
                    <div className="alert-header">
                      <div className={`alert-type text-${alert.type} font-semibold text-xs uppercase tracking-widest`}>
                        {alert.type}
                      </div>
                      <div className="alert-time font-mono text-xs text-muted">
                        {new Date(alert.timestamp).toLocaleTimeString()}
                      </div>
                    </div>
                    <div className="alert-message text-sm text-secondary">
                      {alert.message}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Status Bar */}
        <div className="dashboard-status-bar">
          <div className="status-left">
            <div className="status-item">
              <span className="status-label">Last Update:</span>
              <span className="status-value font-mono">
                {lastUpdate ? new Date(lastUpdate).toLocaleTimeString() : 'Never'}
              </span>
            </div>
            <div className="status-item">
              <span className="status-label">Data Source:</span>
              <span className="status-value font-mono">GoTAK Server</span>
            </div>
          </div>
          <div className="status-right">
            <div className="status-item">
              <span className="status-label">System Status:</span>
              <span 
                className="status-value font-mono font-semibold"
                style={{ color: getConnectionStatusColor() }}
              >
                {getConnectionStatusText()}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Dashboard Styles */}
      <style jsx>{`
        .dashboard-container {
          height: 100%;
          width: 100%;
          padding: var(--spacing-xl);
          background: var(--color-bg-primary);
          overflow-y: auto;
        }

        .dashboard-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: var(--spacing-lg) var(--spacing-xl);
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border: 1px solid rgba(0, 212, 170, 0.15);
          border-radius: var(--radius-lg);
          box-shadow: 
            0 4px 24px rgba(0, 0, 0, 0.4),
            inset 0 1px 0 rgba(255, 255, 255, 0.02);
          backdrop-filter: blur(10px);
          margin-bottom: var(--spacing-xl);
        }

        .header-left {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }
        
        .header-left h1 {
          color: var(--color-accent);
          text-shadow: 0 0 12px rgba(0, 212, 170, 0.3);
        }

        .header-time {
          opacity: 0.7;
          font-size: 0.8rem;
        }

        .header-right {
          display: flex;
          align-items: center;
          gap: var(--space-4);
        }

        .connection-status-indicator {
          display: flex;
          align-items: center;
          gap: var(--spacing-sm);
          padding: var(--spacing-sm) var(--spacing-md);
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: var(--radius-lg);
          backdrop-filter: blur(8px);
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .dashboard-content {
          display: flex;
          flex-direction: column;
          gap: var(--spacing-xl);
        }

        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
          gap: var(--spacing-lg);
        }

        .stat-card {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.4) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: var(--radius-lg);
          padding: var(--spacing-lg);
          position: relative;
          transition: all var(--transition-fast);
          backdrop-filter: blur(8px);
          box-shadow: 
            0 2px 12px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.02);
        }

        .stat-card:hover {
          border-color: rgba(0, 212, 170, 0.2);
          box-shadow: 
            0 4px 20px rgba(0, 0, 0, 0.4),
            0 0 0 1px rgba(0, 212, 170, 0.1);
          transform: translateY(-2px);
        }

        .stat-value {
          margin-bottom: var(--spacing-sm);
          text-shadow: 0 0 8px currentColor;
          font-family: var(--font-display);
          font-weight: 700;
          font-size: 2rem;
        }

        .stat-icon {
          position: absolute;
          top: var(--spacing-md);
          right: var(--spacing-md);
          opacity: 0.3;
          transition: opacity var(--transition-fast);
        }
        
        .stat-card:hover .stat-icon {
          opacity: 0.5;
        }

        .main-grid {
          display: grid;
          grid-template-columns: 2fr 1fr;
          gap: var(--spacing-xl);
          min-height: 500px;
        }

        .map-section,
        .alerts-section {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.4) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: var(--radius-lg);
          padding: var(--spacing-lg);
          display: flex;
          flex-direction: column;
          backdrop-filter: blur(8px);
          box-shadow: 
            0 2px 12px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.02);
        }

        .section-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: var(--spacing-md);
          padding-bottom: var(--spacing-md);
          border-bottom: 1px solid rgba(0, 212, 170, 0.1);
        }
        
        .section-title-wrapper {
          display: flex;
          align-items: center;
          gap: var(--spacing-sm);
        }
        
        .section-title {
          color: var(--color-accent);
          text-shadow: 0 0 6px rgba(0, 212, 170, 0.2);
        }

        .loading-indicator {
          display: flex;
          align-items: center;
          gap: var(--spacing-xs);
        }
        
        .loading-indicator .spinning {
          animation: spin 1s linear infinite;
        }
        
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }

        .alert-count {
          display: flex;
          align-items: center;
          gap: 6px;
          background: rgba(0, 0, 0, 0.3);
          padding: 4px 12px;
          border-radius: var(--radius-md);
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--color-text-secondary);
          border: 1px solid rgba(255, 255, 255, 0.08);
        }

        .map-container {
          flex: 1;
          min-height: 400px;
        }

        .alerts-container {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: var(--spacing-sm);
          max-height: 500px;
          overflow-y: auto;
        }

        .no-alerts {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: var(--space-12);
        }

        .alert-item {
          padding: var(--spacing-md);
          border-radius: var(--radius-md);
          border-left: 3px solid;
          background: rgba(0, 0, 0, 0.2);
          backdrop-filter: blur(4px);
          transition: all var(--transition-fast);
        }
        
        .alert-item:hover {
          background: rgba(0, 0, 0, 0.3);
          transform: translateX(2px);
        }

        .alert-item.alert-info {
          border-left-color: var(--color-info);
        }

        .alert-item.alert-warning {
          border-left-color: var(--color-warning);
        }

        .alert-item.alert-error {
          border-left-color: var(--color-error);
        }

        .alert-item.alert-critical {
          border-left-color: var(--color-error);
          background-color: rgba(255, 71, 87, 0.1);
        }

        .alert-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: var(--space-2);
        }

        .dashboard-status-bar {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: var(--spacing-md) var(--spacing-xl);
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.4) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: var(--radius-lg);
          font-size: 0.75rem;
          backdrop-filter: blur(8px);
          margin-top: var(--spacing-lg);
        }

        .status-left,
        .status-right {
          display: flex;
          gap: var(--spacing-xl);
        }

        .status-item {
          display: flex;
          gap: var(--spacing-sm);
        }

        .status-label {
          color: var(--color-text-muted);
        }

        .status-value {
          color: var(--color-text-secondary);
        }

        /* Responsive Design */
        @media (max-width: 1200px) {
          .main-grid {
            grid-template-columns: 1fr;
          }
          
          .stats-grid {
            grid-template-columns: repeat(2, 1fr);
          }
        }

        @media (max-width: 768px) {
          .dashboard-header {
            flex-direction: column;
            gap: var(--space-4);
          }
          
          .stats-grid {
            grid-template-columns: 1fr;
          }
          
          .dashboard-content {
            padding: var(--space-4);
          }
          
          .status-left,
          .status-right {
            flex-direction: column;
            gap: var(--space-2);
          }
        }
      `}</style>
    </div>
  );
};

export default Dashboard;
