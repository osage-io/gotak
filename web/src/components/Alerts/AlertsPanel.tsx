/**
 * Alerts Panel Component (Stub)
 * Shows tactical alerts and notifications
 */

import React, { memo } from 'react';
import { useActiveAlerts, useTacticalStore } from '../../stores/simpleTacticalStore';

export const AlertsPanel: React.FC = memo(() => {
  const activeAlerts = useActiveAlerts();
  const createAlert = useTacticalStore(state => state.createAlert);

  const handleCreateTestAlert = () => {
    createAlert({
      type: 'threat',
      priority: Math.random() > 0.5 ? 'high' : 'medium',
      title: 'Test Alert',
      message: 'This is a test alert generated from the UI',
      source: 'System Test'
    });
  };

  return (
    <div className="alerts-panel">
      <div className="panel-header">
        <h2>🚨 Tactical Alerts</h2>
        <button className="test-btn" onClick={handleCreateTestAlert}>
          Create Test Alert
        </button>
      </div>
      
      <div className="panel-content">
        {activeAlerts.length > 0 ? (
          <div className="alerts-list">
            {activeAlerts.map(alert => (
              <div key={alert.id} className={`alert-item priority-${alert.priority}`}>
                <div className="alert-header">
                  <span className="alert-type">{alert.type.toUpperCase()}</span>
                  <span className="alert-priority">{alert.priority.toUpperCase()}</span>
                </div>
                <h3>{alert.title}</h3>
                <p>{alert.message}</p>
                <div className="alert-meta">
                  <span>Source: {alert.source || 'Unknown'}</span>
                  <span>{new Date(alert.timestamp).toLocaleTimeString()}</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="no-alerts">
            <div className="empty-state">
              <div className="empty-icon">✅</div>
              <h3>All Clear</h3>
              <p>No active alerts at this time</p>
            </div>
          </div>
        )}
      </div>
      
      <style>{`
        .alerts-panel {
          height: 100%;
          background-color: #0a0d10;
          color: #dcddde;
          display: flex;
          flex-direction: column;
        }
        
        .panel-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1.5rem;
          border-bottom: 1px solid #404040;
          background-color: #1a1d20;
        }
        
        .panel-header h2 {
          margin: 0;
          color: #00d4aa;
          font-size: 1.5rem;
        }
        
        .test-btn {
          background: #2a2d30;
          border: 1px solid #404040;
          color: #dcddde;
          padding: 0.5rem 1rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.15s ease;
          font-size: 0.8rem;
        }
        
        .test-btn:hover {
          background: #36393f;
          border-color: #00d4aa;
          color: #00d4aa;
        }
        
        .panel-content {
          flex: 1;
          overflow-y: auto;
          padding: 1rem;
        }
        
        .alerts-list {
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }
        
        .alert-item {
          background: #1a1d20;
          border: 1px solid #404040;
          border-radius: 8px;
          padding: 1rem;
          transition: all 0.15s ease;
        }
        
        .alert-item.priority-critical {
          border-left: 4px solid #d32f2f;
        }
        
        .alert-item.priority-high {
          border-left: 4px solid #f57c00;
        }
        
        .alert-item.priority-medium {
          border-left: 4px solid #fbc02d;
        }
        
        .alert-item.priority-low {
          border-left: 4px solid #388e3c;
        }
        
        .alert-header {
          display: flex;
          gap: 1rem;
          margin-bottom: 0.5rem;
        }
        
        .alert-type,
        .alert-priority {
          background: #2a2d30;
          padding: 0.25rem 0.5rem;
          border-radius: 4px;
          font-size: 0.75rem;
          font-weight: 600;
        }
        
        .alert-priority {
          background: #f44336;
          color: white;
        }
        
        .alert-item h3 {
          margin: 0 0 0.5rem 0;
          color: #dcddde;
        }
        
        .alert-item p {
          margin: 0 0 1rem 0;
          color: #b0b3b8;
          line-height: 1.4;
        }
        
        .alert-meta {
          display: flex;
          justify-content: space-between;
          font-size: 0.8rem;
          color: #72767d;
        }
        
        .no-alerts {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 100%;
        }
        
        .empty-state {
          text-align: center;
          padding: 3rem;
        }
        
        .empty-icon {
          font-size: 4rem;
          margin-bottom: 1rem;
        }
        
        .empty-state h3 {
          color: #00d4aa;
          margin-bottom: 0.5rem;
        }
        
        .empty-state p {
          color: #b0b3b8;
        }
      `}</style>
    </div>
  );
});

AlertsPanel.displayName = 'AlertsPanel';
