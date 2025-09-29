/**
 * Alerts & Notifications Page - Redesigned
 * Full-width alert management system with modern tactical interface
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { wsService, SystemAlert } from '../services/websocketService';
import { Icon } from '../components/ui/Icon';

// Alert types and interfaces
interface Alert extends SystemAlert {
  acknowledged: boolean;
  acknowledgedBy?: string;
  acknowledgedAt?: string;
  source: string;
  category: string;
  location?: {
    lat: number;
    lng: number;
    description?: string;
  };
}

type AlertFilterType = 'all' | 'unread' | 'critical' | 'error' | 'warning' | 'info';

const Alerts: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [selectedAlert, setSelectedAlert] = useState<string | null>(null);
  const [filter, setFilter] = useState<AlertFilterType>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Initialize with mock data and WebSocket listeners
  useEffect(() => {
    // Mock initial alerts
    const initialAlerts: Alert[] = [
      {
        id: '1',
        type: 'critical',
        title: 'Communication Lost',
        message: 'Lost connection to tactical unit ALPHA-1. Last known position: Grid 38S MC 12345 67890. Signal strength dropped below threshold at 14:32:15Z.',
        timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        acknowledged: false,
        source: 'COMMS',
        category: 'Network',
        requiresAck: true,
        location: {
          lat: 38.8951,
          lng: -77.0364,
          description: 'Washington, DC'
        }
      },
      {
        id: '2',
        type: 'error',
        title: 'Entity Position Stale',
        message: 'Entity BRAVO-2 position data is over 10 minutes old. GPS signal may be degraded or unit may be in RF-denied environment.',
        timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        acknowledged: true,
        acknowledgedBy: 'TAC-OPS-01',
        acknowledgedAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        source: 'TRACKING',
        category: 'Position',
        requiresAck: false
      },
      {
        id: '3',
        type: 'warning',
        title: 'Server Load High',
        message: 'TAK server CPU usage at 85%. Performance degradation possible if load continues to increase.',
        timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
        acknowledged: false,
        source: 'SYSTEM',
        category: 'Performance',
        requiresAck: false
      },
      {
        id: '4',
        type: 'info',
        title: 'New Entity Connected',
        message: 'Tactical entity CHARLIE-3 has joined the network. Authentication verified. Assigned to Blue Force.',
        timestamp: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
        acknowledged: false,
        source: 'AUTH',
        category: 'Connection',
        requiresAck: false
      },
      {
        id: '5',
        type: 'critical',
        title: 'Emergency Beacon Activated',
        message: 'Emergency beacon activated by DELTA-4. Immediate assistance required. Automated SOS protocol initiated.',
        timestamp: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
        acknowledged: false,
        source: 'EMERGENCY',
        category: 'SOS',
        requiresAck: true,
        location: {
          lat: 38.9072,
          lng: -77.0369,
          description: 'Georgetown Area'
        }
      }
    ];

    setAlerts(initialAlerts);

    // Set up WebSocket listener for real-time alerts
    const unsubscribeAlerts = wsService.onSystemAlert((systemAlert) => {
      const newAlert: Alert = {
        ...systemAlert,
        acknowledged: false,
        source: 'SYSTEM',
        category: 'Real-time',
      };
      
      setAlerts(prev => [newAlert, ...prev]);
    });

    return () => {
      unsubscribeAlerts();
    };
  }, []);

  // Handle acknowledge
  const handleAcknowledge = useCallback((alertId: string) => {
    setAlerts(prev => prev.map(alert => 
      alert.id === alertId 
        ? { 
            ...alert, 
            acknowledged: true,
            acknowledgedBy: 'Current User',
            acknowledgedAt: new Date().toISOString()
          }
        : alert
    ));
  }, []);

  // Handle delete
  const handleDelete = useCallback((alertId: string) => {
    setAlerts(prev => prev.filter(alert => alert.id !== alertId));
    if (selectedAlert === alertId) {
      setSelectedAlert(null);
    }
  }, [selectedAlert]);

  // Filter alerts
  const filteredAlerts = useMemo(() => {
    let filtered = alerts;

    // Apply type filter
    if (filter !== 'all') {
      if (filter === 'unread') {
        filtered = filtered.filter(a => !a.acknowledged);
      } else {
        filtered = filtered.filter(a => a.type === filter);
      }
    }

    // Apply search
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(alert =>
        alert.title.toLowerCase().includes(query) ||
        alert.message.toLowerCase().includes(query) ||
        alert.source.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [alerts, filter, searchQuery]);

  // Get counts
  const counts = useMemo(() => ({
    all: alerts.length,
    unread: alerts.filter(a => !a.acknowledged).length,
    critical: alerts.filter(a => a.type === 'critical').length,
    error: alerts.filter(a => a.type === 'error').length,
    warning: alerts.filter(a => a.type === 'warning').length,
    info: alerts.filter(a => a.type === 'info').length,
  }), [alerts]);

  const selectedAlertData = alerts.find(a => a.id === selectedAlert);

  const getAlertIconName = (type: Alert['type']): string => {
    switch (type) {
      case 'critical': return 'cross';
      case 'error': return 'alert-circle';
      case 'warning': return 'warning';
      case 'info': return 'info';
      default: return 'bell';
    }
  };

  const getAlertIconColor = (type: Alert['type']): string => {
    switch (type) {
      case 'critical': return '#ef4444'; // danger red
      case 'error': return '#f59e0b'; // warning amber
      case 'warning': return '#f59e0b'; // warning amber
      case 'info': return '#14b8a6'; // primary teal
      default: return '#94a3b8'; // muted gray
    }
  };

  return (
    <div className="alerts-fullpage">
      {/* Header Bar */}
      <header className="alerts-header">
        <div className="header-title">
          <h1>Alerts & Notifications</h1>
          <div className="alert-stats">
            <span className={`stat ${counts.critical > 0 ? 'critical' : ''}`}>
              {counts.critical} Critical
            </span>
            <span className={`stat ${counts.unread > 0 ? 'unread' : ''}`}>
              {counts.unread} Unread
            </span>
            <span className="stat">{counts.all} Total</span>
          </div>
          <div className="alert-type-legend">
            <div className="legend-item">
              <div className="legend-icon critical">
                <Icon name="cross" size={16} color="#ef4444" />
              </div>
              <span>Critical</span>
            </div>
            <div className="legend-item">
              <div className="legend-icon error">
                <Icon name="alert-circle" size={16} color="#f59e0b" />
              </div>
              <span>Error</span>
            </div>
            <div className="legend-item">
              <div className="legend-icon warning">
                <Icon name="warning" size={16} color="#f59e0b" />
              </div>
              <span>Warning</span>
            </div>
            <div className="legend-item">
              <div className="legend-icon info">
                <Icon name="info" size={16} color="#14b8a6" />
              </div>
              <span>Info</span>
            </div>
          </div>
        </div>

        <div className="header-controls">
          <div className="search-box">
            <Icon name="search" size={16} color="var(--color-text-muted)" />
            <input
              type="text"
              placeholder="Search alerts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
            />
          </div>

          <div className="filter-tabs">
            {(['all', 'unread', 'critical', 'error', 'warning', 'info'] as AlertFilterType[]).map(f => (
              <button
                key={f}
                className={`filter-tab ${filter === f ? 'active' : ''}`}
                onClick={() => setFilter(f)}
              >
                {f === 'all' ? 'All' : f.charAt(0).toUpperCase() + f.slice(1)}
                <span className="count">{counts[f]}</span>
              </button>
            ))}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="alerts-content">
        {/* Alerts List */}
        <div className="alerts-list">
          {filteredAlerts.length === 0 ? (
            <div className="no-alerts">
              <div className="no-alerts-icon">
                <Icon name="inbox" size={48} color="var(--color-text-muted)" />
              </div>
              <h3>No alerts found</h3>
              <p>{searchQuery ? `No results for "${searchQuery}"` : 'All clear'}</p>
            </div>
          ) : (
            filteredAlerts.map(alert => (
              <div
                key={alert.id}
                className={`alert-item ${selectedAlert === alert.id ? 'selected' : ''} ${!alert.acknowledged ? 'unread' : ''} ${alert.type}`}
                onClick={() => setSelectedAlert(alert.id)}
              >
                <div className="alert-icon">
                  <Icon 
                    name={getAlertIconName(alert.type)} 
                    size={24} 
                    color={getAlertIconColor(alert.type)} 
                  />
                </div>

                <div className="alert-main">
                  <div className="alert-header">
                    <h3 className="alert-title">{alert.title}</h3>
                    <span className={`alert-type ${alert.type}`}>
                      {alert.type.toUpperCase()}
                    </span>
                  </div>

                  <p className="alert-message">{alert.message}</p>

                  <div className="alert-meta">
                    <span className="meta-item">
                      <span className="label">Source:</span> {alert.source}
                    </span>
                    <span className="meta-item">
                      <span className="label">Category:</span> {alert.category}
                    </span>
                    <span className="meta-item">
                      <span className="label">Time:</span> {new Date(alert.timestamp).toLocaleTimeString()}
                    </span>
                    {alert.location && (
                      <span className="meta-item location">
                        <Icon name="map-pin" size={14} color="var(--color-accent)" />
                        {alert.location.description}
                      </span>
                    )}
                  </div>

                  {alert.acknowledged && (
                    <div className="ack-info">
                      <Icon name="check" size={16} color="#10b981" />
                      Acknowledged by {alert.acknowledgedBy} at {new Date(alert.acknowledgedAt!).toLocaleTimeString()}
                    </div>
                  )}
                </div>

                <div className="alert-actions">
                  {!alert.acknowledged && (
                    <button
                      className="action-btn ack"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAcknowledge(alert.id);
                      }}
                      title="Acknowledge"
                    >
                      <Icon name="check" size={16} color="#ffffff" />
                    </button>
                  )}
                  <button
                    className="action-btn delete"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDelete(alert.id);
                    }}
                    title="Delete"
                  >
                    <Icon name="x" size={16} color="#ffffff" />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Detail Panel */}
        {selectedAlertData && (
          <div className="alert-details-panel">
            <div className="details-header">
              <h2>Alert Details</h2>
              <button 
                className="close-btn"
                onClick={() => setSelectedAlert(null)}
              >
                <Icon name="x" size={20} color="var(--color-text-secondary)" />
              </button>
            </div>

            <div className="details-content">
              <div className="detail-section">
                <h3>{selectedAlertData.title}</h3>
                <p className="detail-message">{selectedAlertData.message}</p>
              </div>

              <div className="detail-grid">
                <div className="detail-item">
                  <span className="label">Alert ID</span>
                  <span className="value">{selectedAlertData.id}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Type</span>
                  <span className={`value ${selectedAlertData.type}`}>
                    {selectedAlertData.type.toUpperCase()}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="label">Source</span>
                  <span className="value">{selectedAlertData.source}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Category</span>
                  <span className="value">{selectedAlertData.category}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Timestamp</span>
                  <span className="value">{new Date(selectedAlertData.timestamp).toLocaleString()}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Requires Ack</span>
                  <span className="value">{selectedAlertData.requiresAck ? 'Yes' : 'No'}</span>
                </div>
                {selectedAlertData.location && (
                  <>
                    <div className="detail-item">
                      <span className="label">Location</span>
                      <span className="value">{selectedAlertData.location.description}</span>
                    </div>
                    <div className="detail-item">
                      <span className="label">Coordinates</span>
                      <span className="value">
                        {selectedAlertData.location.lat.toFixed(4)}, {selectedAlertData.location.lng.toFixed(4)}
                      </span>
                    </div>
                  </>
                )}
                {selectedAlertData.acknowledged && (
                  <>
                    <div className="detail-item">
                      <span className="label">Acknowledged By</span>
                      <span className="value">{selectedAlertData.acknowledgedBy}</span>
                    </div>
                    <div className="detail-item">
                      <span className="label">Acknowledged At</span>
                      <span className="value">{new Date(selectedAlertData.acknowledgedAt!).toLocaleString()}</span>
                    </div>
                  </>
                )}
              </div>

              <div className="detail-actions">
                {!selectedAlertData.acknowledged && (
                  <button
                    className="btn-primary"
                    onClick={() => handleAcknowledge(selectedAlertData.id)}
                  >
                    Acknowledge Alert
                  </button>
                )}
                <button
                  className="btn-danger"
                  onClick={() => handleDelete(selectedAlertData.id)}
                >
                  Delete Alert
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Styles */}
      <style jsx>{`
        .alerts-fullpage {
          height: 100vh;
          width: 100vw;
          display: flex;
          flex-direction: column;
          background: var(--color-bg-primary);
          overflow: hidden;
        }

        /* Header */
        .alerts-header {
          padding: 20px 24px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border-bottom: 1px solid rgba(0, 212, 170, 0.15);
          backdrop-filter: blur(10px);
          display: flex;
          justify-content: space-between;
          align-items: center;
          flex-wrap: wrap;
          gap: 20px;
        }

        .header-title h1 {
          margin: 0;
          color: var(--color-accent);
          text-shadow: 0 0 8px rgba(0, 212, 170, 0.3);
          font-size: 1.5rem;
        }

        .alert-stats {
          display: flex;
          gap: 16px;
          margin-top: 8px;
        }

        .stat {
          font-size: 0.85rem;
          color: var(--color-text-secondary);
        }

        .stat.critical {
          color: #ef4444;
          font-weight: 600;
        }

        .stat.unread {
          color: #f59e0b;
          font-weight: 600;
        }

        /* Alert Type Legend */
        .alert-type-legend {
          display: flex;
          gap: 16px;
          margin-top: 12px;
          padding-top: 12px;
          border-top: 1px solid rgba(0, 212, 170, 0.1);
        }

        .legend-item {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 0.85rem;
          color: var(--color-text-secondary);
        }

        .legend-icon {
          width: 28px;
          height: 28px;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 6px;
          border: 1px solid;
        }

        .legend-icon.critical {
          background: rgba(239, 68, 68, 0.1);
          border-color: rgba(239, 68, 68, 0.3);
        }

        .legend-icon.error {
          background: rgba(245, 158, 11, 0.1);
          border-color: rgba(245, 158, 11, 0.3);
        }

        .legend-icon.warning {
          background: rgba(245, 158, 11, 0.1);
          border-color: rgba(245, 158, 11, 0.3);
        }

        .legend-icon.info {
          background: rgba(20, 184, 166, 0.1);
          border-color: rgba(20, 184, 166, 0.3);
        }

        .header-controls {
          display: flex;
          align-items: center;
          gap: 20px;
        }

        .search-box {
          position: relative;
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          transition: all 0.2s ease;
        }

        .search-box:focus-within {
          border-color: rgba(0, 212, 170, 0.4);
          background: rgba(0, 0, 0, 0.4);
        }

        .search-input {
          background: transparent;
          border: none;
          color: var(--color-text-primary);
          width: 200px;
          outline: none;
        }

        .search-input:focus {
          outline: none;
          border-color: rgba(0, 212, 170, 0.4);
          background: rgba(0, 0, 0, 0.4);
        }

        .filter-tabs {
          display: flex;
          gap: 8px;
        }

        .filter-tab {
          padding: 8px 16px;
          background: rgba(0, 0, 0, 0.2);
          border: 1px solid transparent;
          border-radius: 6px;
          color: var(--color-text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .filter-tab:hover {
          background: rgba(0, 212, 170, 0.05);
          border-color: rgba(0, 212, 170, 0.2);
        }

        .filter-tab.active {
          background: rgba(0, 212, 170, 0.1);
          border-color: rgba(0, 212, 170, 0.3);
          color: var(--color-accent);
        }

        .filter-tab .count {
          font-size: 0.75rem;
          padding: 2px 6px;
          background: rgba(0, 0, 0, 0.3);
          border-radius: 10px;
        }

        /* Content */
        .alerts-content {
          flex: 1;
          display: flex;
          overflow: hidden;
        }

        /* Alerts List */
        .alerts-list {
          flex: 1;
          overflow-y: auto;
          padding: 24px;
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .no-alerts {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          color: var(--color-text-muted);
        }

        .no-alerts-icon {
          margin-bottom: 16px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .no-alerts h3 {
          margin: 0 0 8px 0;
          color: var(--color-text-primary);
        }

        /* Alert Item */
        .alert-item {
          display: flex;
          gap: 16px;
          padding: 20px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.3) 0%, 
            rgba(15, 20, 25, 0.5) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.2s ease;
          position: relative;
        }

        .alert-item:hover {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.4) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          transform: translateX(4px);
        }

        .alert-item.selected {
          border-color: rgba(0, 212, 170, 0.3);
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.05) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
        }

        .alert-item.unread {
          border-left: 3px solid var(--color-accent);
        }

        .alert-item.critical {
          border-left-color: #ef4444;
        }

        .alert-item.error {
          border-left-color: #f59e0b;
        }

        .alert-item.warning {
          border-left-color: #f59e0b;
        }

        .alert-item.info {
          border-left-color: #14b8a6;
        }

        .alert-icon {
          flex-shrink: 0;
          display: flex;
          align-items: center;
          justify-content: center;
          width: 36px;
          height: 36px;
          border-radius: 6px;
          border: 1px solid;
        }

        .alert-item.critical .alert-icon {
          background: rgba(239, 68, 68, 0.1);
          border-color: rgba(239, 68, 68, 0.3);
        }

        .alert-item.error .alert-icon {
          background: rgba(245, 158, 11, 0.1);
          border-color: rgba(245, 158, 11, 0.3);
        }

        .alert-item.warning .alert-icon {
          background: rgba(245, 158, 11, 0.1);
          border-color: rgba(245, 158, 11, 0.3);
        }

        .alert-item.info .alert-icon {
          background: rgba(20, 184, 166, 0.1);
          border-color: rgba(20, 184, 166, 0.3);
        }

        .alert-main {
          flex: 1;
        }

        .alert-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 8px;
        }

        .alert-title {
          margin: 0;
          font-size: 1.1rem;
          color: var(--color-text-primary);
        }

        .alert-type {
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 0.7rem;
          font-weight: 600;
          text-transform: uppercase;
        }

        .alert-type.critical {
          background: rgba(239, 68, 68, 0.2);
          color: #ef4444;
        }

        .alert-type.error {
          background: rgba(245, 158, 11, 0.2);
          color: #f59e0b;
        }

        .alert-type.warning {
          background: rgba(245, 158, 11, 0.2);
          color: #f59e0b;
        }

        .alert-type.info {
          background: rgba(20, 184, 166, 0.2);
          color: #14b8a6;
        }

        .alert-message {
          margin: 0 0 12px 0;
          color: var(--color-text-secondary);
          line-height: 1.5;
        }

        .alert-meta {
          display: flex;
          flex-wrap: wrap;
          gap: 16px;
          font-size: 0.85rem;
        }

        .meta-item {
          color: var(--color-text-muted);
        }

        .meta-item .label {
          font-weight: 600;
          color: var(--color-text-secondary);
        }

        .meta-item.location {
          color: var(--color-accent);
          display: flex;
          align-items: center;
          gap: 4px;
        }

        .ack-info {
          margin-top: 8px;
          padding: 6px 12px;
          background: rgba(16, 185, 129, 0.1);
          border-radius: 4px;
          font-size: 0.85rem;
          color: #10b981;
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .alert-actions {
          display: flex;
          gap: 8px;
          align-items: flex-start;
        }

        .action-btn {
          width: 32px;
          height: 32px;
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(0, 0, 0, 0.2);
          color: var(--color-text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .action-btn:hover {
          transform: scale(1.1);
        }

        .action-btn.ack:hover {
          background: #10b981;
          color: white;
        }

        .action-btn.delete:hover {
          background: #ef4444;
          color: white;
        }

        /* Details Panel */
        .alert-details-panel {
          width: 400px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border-left: 1px solid rgba(0, 212, 170, 0.2);
          display: flex;
          flex-direction: column;
        }

        .details-header {
          padding: 20px;
          border-bottom: 1px solid rgba(0, 212, 170, 0.1);
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .details-header h2 {
          margin: 0;
          color: var(--color-accent);
          font-size: 1.2rem;
        }

        .close-btn {
          width: 32px;
          height: 32px;
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(0, 0, 0, 0.2);
          color: var(--color-text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .close-btn:hover {
          background: rgba(255, 255, 255, 0.1);
        }

        .details-content {
          flex: 1;
          padding: 20px;
          overflow-y: auto;
        }

        .detail-section {
          margin-bottom: 24px;
        }

        .detail-section h3 {
          margin: 0 0 12px 0;
          color: var(--color-text-primary);
        }

        .detail-message {
          color: var(--color-text-secondary);
          line-height: 1.6;
        }

        .detail-grid {
          display: grid;
          gap: 16px;
        }

        .detail-item {
          display: flex;
          justify-content: space-between;
          padding: 12px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 6px;
        }

        .detail-item .label {
          font-weight: 600;
          color: var(--color-text-secondary);
          font-size: 0.85rem;
        }

        .detail-item .value {
          color: var(--color-text-primary);
          font-size: 0.85rem;
          text-align: right;
        }

        .detail-item .value.critical {
          color: var(--color-error);
        }

        .detail-item .value.error {
          color: var(--color-warning);
        }

        .detail-item .value.warning {
          color: #fbc02d;
        }

        .detail-item .value.info {
          color: var(--color-info);
        }

        .detail-actions {
          margin-top: 24px;
          display: flex;
          gap: 12px;
        }

        .btn-primary,
        .btn-danger {
          flex: 1;
          padding: 12px;
          border-radius: 6px;
          border: 1px solid;
          font-weight: 600;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .btn-primary {
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.2) 0%, 
            rgba(0, 212, 170, 0.1) 100%);
          border-color: rgba(0, 212, 170, 0.3);
          color: var(--color-accent);
        }

        .btn-primary:hover {
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.3) 0%, 
            rgba(0, 212, 170, 0.2) 100%);
        }

        .btn-danger {
          background: linear-gradient(135deg, 
            rgba(211, 47, 47, 0.2) 0%, 
            rgba(211, 47, 47, 0.1) 100%);
          border-color: rgba(211, 47, 47, 0.3);
          color: var(--color-error);
        }

        .btn-danger:hover {
          background: linear-gradient(135deg, 
            rgba(211, 47, 47, 0.3) 0%, 
            rgba(211, 47, 47, 0.2) 100%);
        }

        /* Responsive */
        @media (max-width: 1200px) {
          .alert-details-panel {
            position: absolute;
            right: 0;
            top: 0;
            bottom: 0;
            box-shadow: -4px 0 24px rgba(0, 0, 0, 0.6);
          }
        }

        @media (max-width: 768px) {
          .alerts-header {
            flex-direction: column;
            align-items: stretch;
          }

          .header-controls {
            flex-direction: column;
            gap: 12px;
          }

          .search-input {
            width: 100%;
          }

          .filter-tabs {
            overflow-x: auto;
          }

          .alert-details-panel {
            width: 100%;
          }

          .alert-meta {
            flex-direction: column;
            gap: 8px;
          }
        }
      `}</style>
    </div>
  );
};

export default Alerts;
