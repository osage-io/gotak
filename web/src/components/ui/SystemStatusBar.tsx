/**
 * System Status Bar Component
 * Shows connection status, entity count, alerts, and system info
 */

import React, { memo } from 'react';
import type { SystemStatus } from '../../stores/simpleTacticalStore';

interface SystemStatusBarProps {
  status: SystemStatus;
  entityCount: number;
  alertCount: number;
  onToggleDebug?: () => void;
}

const SystemStatusBar: React.FC<SystemStatusBarProps> = memo(({ 
  status, 
  entityCount, 
  alertCount, 
  onToggleDebug 
}) => {
  const formatLatency = (latency: number) => {
    return latency < 1000 ? `${latency}ms` : `${(latency / 1000).toFixed(1)}s`;
  };

  const formatLastSync = (timestamp: number) => {
    const now = Date.now();
    const diff = now - timestamp;
    
    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return `${Math.floor(diff / 86400000)}d ago`;
  };

  return (
    <div className="system-status-bar">
      <div className="status-left">
        {/* Connection Status */}
        <div className={`connection-status ${status.connectionStatus}`}>
          <div className="status-indicator" />
          <span>
            {status.connectionStatus === 'connected' ? 'CONNECTED' : 
             status.connectionStatus === 'connecting' ? 'CONNECTING...' : 
             status.connectionStatus === 'error' ? 'ERROR' : 'DISCONNECTED'}
          </span>
          {status.connectionStatus === 'connected' && status.serverLatency > 0 && (
            <span className="latency">({formatLatency(status.serverLatency)})</span>
          )}
        </div>

        {/* Network Quality */}
        {status.networkQuality && status.networkQuality !== 'offline' && (
          <div className="network-quality">
            📡 {status.networkQuality.toUpperCase()}
          </div>
        )}

        {/* GPS Status */}
        <div className={`gps-status ${status.gpsStatus}`}>
          🛰️ {status.gpsStatus.toUpperCase()}
        </div>

        {/* Security Indicators */}
        <div className="security-indicators">
          {status.encryption && <span className="security-badge">🔒 ENC</span>}
          {status.authentication && <span className="security-badge">🗝️ AUTH</span>}
        </div>
      </div>

      <div className="status-right">
        {/* Entity Count */}
        <div className="entity-count">
          👥 {entityCount} {entityCount === 1 ? 'Entity' : 'Entities'}
        </div>

        {/* Alert Count */}
        {alertCount > 0 && (
          <div className="alert-count">
            🚨 {alertCount} {alertCount === 1 ? 'Alert' : 'Alerts'}
          </div>
        )}

        {/* Last Sync */}
        <div className="last-sync">
          🔄 {formatLastSync(status.lastSync)}
        </div>

        {/* Battery Level (if available) */}
        {status.batteryLevel !== undefined && (
          <div className={`battery-level ${status.batteryLevel < 20 ? 'low' : ''}`}>
            🔋 {status.batteryLevel}%
          </div>
        )}

        {/* Debug Toggle */}
        {onToggleDebug && (
          <button className="debug-toggle" onClick={onToggleDebug} title="Toggle Debug Mode">
            🐛
          </button>
        )}
      </div>
    </div>
  );
});

SystemStatusBar.displayName = 'SystemStatusBar';

export default SystemStatusBar;
