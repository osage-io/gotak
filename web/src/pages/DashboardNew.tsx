/**
 * Dashboard Page - Redesigned
 * Full-width tactical command overview with real-time metrics
 */

import React, { useEffect, useState } from 'react';
import EntityMap from '../components/Map/EntityMap';
import { useEntityStats, useBulkEntityOperations } from '../hooks/useEntityTracker';
import { wsService, ConnectionState } from '../services/websocketService';
import { Icon } from '../components/ui/Icon';

interface MetricCard {
  id: string;
  title: string;
  value: number | string;
  unit?: string;
  change?: number;
  iconName: string;
  color: string;
  trend?: 'up' | 'down' | 'neutral';
}

interface ActivityItem {
  id: string;
  type: 'entity' | 'alert' | 'system' | 'communication';
  title: string;
  description: string;
  timestamp: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
}

const DashboardNew: React.FC = () => {
  const [currentTime, setCurrentTime] = useState(new Date());
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);

  // Entity tracking
  const { counts, lastUpdate } = useEntityStats();
  const { loadInitialEntities, isLoading } = useBulkEntityOperations();

  // WebSocket connection status
  const [connectionStatus, setConnectionStatus] = useState<ConnectionState>(
    wsService.connectionState
  );

  // System metrics
  const [metrics, setMetrics] = useState<MetricCard[]>([
    {
      id: 'friendly',
      title: 'Friendly Forces',
      value: 0,
      iconName: 'shield',
      color: '#2ed573',
      trend: 'neutral'
    },
    {
      id: 'hostile',
      title: 'Hostile Forces',
      value: 0,
      iconName: 'warning',
      color: '#ff4757',
      trend: 'up'
    },
    {
      id: 'unknown',
      title: 'Unknown Contacts',
      value: 0,
      iconName: 'target',
      color: '#ffa502',
      trend: 'neutral'
    },
    {
      id: 'drones',
      title: 'Active Drones',
      value: 3,
      iconName: 'rocket',
      color: '#3742fa',
      trend: 'up'
    },
    {
      id: 'sensors',
      title: 'Online Sensors',
      value: 12,
      iconName: 'broadcast',
      color: '#00d4aa',
      trend: 'neutral'
    },
    {
      id: 'messages',
      title: 'Messages Today',
      value: 247,
      iconName: 'chat',
      color: '#5f27cd',
      trend: 'up'
    },
    {
      id: 'alerts',
      title: 'Active Alerts',
      value: 5,
      iconName: 'alert',
      color: '#ff6348',
      trend: 'down'
    },
    {
      id: 'bandwidth',
      title: 'Network Load',
      value: '82',
      unit: '%',
      iconName: 'chart',
      color: '#70a1ff',
      trend: 'up'
    }
  ]);

  // Update metrics with entity counts
  useEffect(() => {
    setMetrics(prev => prev.map(metric => {
      switch (metric.id) {
        case 'friendly':
          return { ...metric, value: counts.friendly };
        case 'hostile':
          return { ...metric, value: counts.hostile };
        case 'unknown':
          return { ...metric, value: counts.unknown };
        default:
          return metric;
      }
    }));
  }, [counts]);

  // Update current time
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

    // Listen for system alerts and create activities
    const unsubscribeAlerts = wsService.onSystemAlert((alert) => {
      const activity: ActivityItem = {
        id: alert.id,
        type: 'alert',
        title: alert.title,
        description: alert.message,
        timestamp: alert.timestamp,
        priority: alert.type === 'critical' ? 'critical' : 
                  alert.type === 'error' ? 'high' :
                  alert.type === 'warning' ? 'medium' : 'low'
      };
      
      setActivities(prev => [activity, ...prev.slice(0, 19)]); // Keep last 20
    });

    return () => {
      unsubscribeConnection();
      unsubscribeDisconnection();
      unsubscribeError();
      unsubscribeAlerts();
    };
  }, []);

  // Initialize dashboard
  useEffect(() => {
    const initializeDashboard = async () => {
      try {
        await wsService.connect();
        await loadInitialEntities();
        
        // Add some mock activities
        setActivities([
          {
            id: '1',
            type: 'entity',
            title: 'New Entity Detected',
            description: 'ALPHA-1 has joined the network at Grid 38S MC 12345 67890',
            timestamp: new Date().toISOString(),
            priority: 'medium'
          },
          {
            id: '2',
            type: 'communication',
            title: 'Emergency Message',
            description: 'DELTA-4 requesting immediate CASEVAC at position',
            timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
            priority: 'critical'
          },
          {
            id: '3',
            type: 'system',
            title: 'Sensor Alert',
            description: 'Motion detected by SENSOR-05 in sector 7',
            timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
            priority: 'high'
          },
          {
            id: '4',
            type: 'entity',
            title: 'Position Update',
            description: 'BRAVO-2 position stale for over 10 minutes',
            timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
            priority: 'medium'
          },
          {
            id: '5',
            type: 'system',
            title: 'Drone Deployed',
            description: 'EAGLE-EYE-1 launched for reconnaissance mission',
            timestamp: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
            priority: 'low'
          }
        ]);
      } catch (error) {
        console.error('Failed to initialize dashboard:', error);
      }
    };

    initializeDashboard();
  }, [loadInitialEntities]);

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
        return '#2ed573';
      case ConnectionState.CONNECTING:
      case ConnectionState.RECONNECTING:
        return '#ffa502';
      case ConnectionState.DISCONNECTED:
      case ConnectionState.ERROR:
        return '#ff4757';
      default:
        return '#57606f';
    }
  };

  const getPriorityColor = (priority: ActivityItem['priority']) => {
    switch (priority) {
      case 'critical': return '#ff4757';
      case 'high': return '#ffa502';
      case 'medium': return '#3742fa';
      case 'low': return '#70a1ff';
      default: return '#57606f';
    }
  };

  const getActivityIcon = (type: ActivityItem['type']) => {
    switch (type) {
      case 'entity': return <Icon name="users" size={20} color="var(--color-text-secondary)" />;
      case 'alert': return <Icon name="warning" size={20} color="var(--color-warning)" />;
      case 'system': return <Icon name="settings" size={20} color="var(--color-text-secondary)" />;
      case 'communication': return <Icon name="chat" size={20} color="var(--color-accent)" />;
      default: return <Icon name="pin" size={20} color="var(--color-text-muted)" />;
    }
  };

  const formatTimeAgo = (timestamp: string) => {
    const seconds = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  };

  return (
    <div className="dashboard-fullpage">
      {/* Dashboard Header */}
      <header className="dashboard-header">
        <div className="header-content">
          <div className="header-title">
            <h1>Command Dashboard</h1>
            <div className="system-time">
              <span className="time-label">System Time:</span>
              <span className="time-value">{currentTime.toLocaleTimeString()}</span>
              <span className="date-value">{currentTime.toLocaleDateString()}</span>
            </div>
          </div>

          <div className="header-status">
            <div className="connection-indicator">
              <span 
                className="status-dot"
                style={{ backgroundColor: getConnectionStatusColor() }}
              />
              <span className="status-text">{getConnectionStatusText()}</span>
            </div>
            <div className="last-update">
              Last Update: {lastUpdate ? new Date(lastUpdate).toLocaleTimeString() : 'Never'}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="dashboard-content">
        {/* Metrics Grid */}
        <section className="metrics-section">
          <div className="section-header">
            <h2>Operational Metrics</h2>
            <button className="refresh-btn">
              <Icon name="sync" size={16} color="currentColor" /> Refresh
            </button>
          </div>
          
          <div className="metrics-grid">
            {metrics.map(metric => {
              const isSelected = selectedMetric === metric.id;
              const trendValue = metric.trend === 'up' ? '+12%' : metric.trend === 'down' ? '-8%' : '0%';
              return (
                <div
                  key={metric.id}
                  className={`metric-card ${isSelected ? 'selected' : ''}`}
                  onClick={() => setSelectedMetric(metric.id === selectedMetric ? null : metric.id)}
                >
                  <div className="metric-glow" style={{ background: `radial-gradient(circle at center, ${metric.color}15 0%, transparent 70%)` }} />
                  
                  <div className="metric-icon-container">
                    <div className="metric-icon-bg" style={{ background: `linear-gradient(135deg, ${metric.color}20 0%, ${metric.color}10 100%)` }}>
                      <Icon name={metric.iconName as any} size={20} color={metric.color} />
                    </div>
                  </div>
                  
                  <div className="metric-content">
                    <div className="metric-label">
                      {metric.title}
                    </div>
                    
                    <div className="metric-value-row">
                      <span className="metric-value">
                        {metric.value}
                        {metric.unit && <span className="metric-unit">{metric.unit}</span>}
                      </span>
                      
                      {metric.trend && (
                        <div className={`metric-trend ${metric.trend}`}>
                          <Icon 
                            name={metric.trend === 'up' ? 'trending' : metric.trend === 'down' ? 'trending' : 'trending'} 
                            size={12} 
                            color={metric.trend === 'up' ? '#2ed573' : metric.trend === 'down' ? '#ff4757' : '#57606f'}
                          />
                          <span className="trend-value">{trendValue}</span>
                        </div>
                      )}
                    </div>
                    
                    <div className="metric-progress">
                      <div 
                        className="metric-progress-bar" 
                        style={{ 
                          width: `${Math.min(100, (typeof metric.value === 'number' ? (metric.value / 100) * 100 : parseInt(metric.value as string)))}%`,
                          background: `linear-gradient(90deg, ${metric.color}80 0%, ${metric.color} 100%)`
                        }}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        {/* Map and Activity Feed */}
        <div className="main-grid">
          {/* Map Section */}
          <section className="map-section">
            <div className="section-header">
              <h2>Tactical Map Overview</h2>
            </div>
            <div className="map-container">
              <EntityMap height="100%" />
            </div>
          </section>

          {/* Activity Feed */}
          <section className="activity-section">
            <div className="section-header">
              <h2>Activity Feed</h2>
              <span className="activity-count">{activities.length} events</span>
            </div>
            
            <div className="activity-feed">
              {activities.length === 0 ? (
                <div className="no-activity">
                  <span className="no-activity-icon">
                    <Icon name="bell" size={32} color="var(--color-text-muted)" />
                  </span>
                  <p>No recent activity</p>
                </div>
              ) : (
                activities.map(activity => (
                  <div key={activity.id} className="activity-item">
                    <div className="activity-icon">
                      {getActivityIcon(activity.type)}
                    </div>
                    
                    <div className="activity-content">
                      <div className="activity-header">
                        <h4>{activity.title}</h4>
                        <span 
                          className="priority-badge"
                          style={{ backgroundColor: getPriorityColor(activity.priority) + '20',
                                  color: getPriorityColor(activity.priority) }}
                        >
                          {activity.priority}
                        </span>
                      </div>
                      <p className="activity-description">{activity.description}</p>
                      <span className="activity-time">{formatTimeAgo(activity.timestamp)}</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </section>
        </div>

        {/* Quick Actions Bar */}
        <section className="actions-section">
          <div className="quick-actions">
            <button className="action-btn primary">
              <Icon name="alert" size={16} color="currentColor" /> Emergency Alert
            </button>
            <button className="action-btn">
              <Icon name="broadcast" size={16} color="currentColor" /> Deploy Sensor
            </button>
            <button className="action-btn">
              <Icon name="rocket" size={16} color="currentColor" /> Launch Drone
            </button>
            <button className="action-btn">
              <Icon name="chat" size={16} color="currentColor" /> Broadcast Message
            </button>
            <button className="action-btn">
              <Icon name="chart" size={16} color="currentColor" /> Generate Report
            </button>
          </div>
        </section>
      </div>

      {/* Styles */}
      <style jsx>{`
        .dashboard-fullpage {
          height: 100vh;
          width: 100vw;
          display: flex;
          flex-direction: column;
          background: var(--color-bg-primary);
          overflow: hidden;
        }

        /* Header */
        .dashboard-header {
          height: 72px;
          padding: 0 24px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border-bottom: 1px solid rgba(0, 212, 170, 0.15);
          backdrop-filter: blur(10px);
          display: flex;
          align-items: center;
        }

        .header-content {
          width: 100%;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .header-title h1 {
          margin: 0;
          color: var(--color-accent);
          text-shadow: 0 0 8px rgba(0, 212, 170, 0.3);
          font-size: 1.5rem;
        }

        .system-time {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-top: 4px;
          font-size: 0.85rem;
          color: var(--color-text-secondary);
        }

        .time-label {
          color: var(--color-text-muted);
        }

        .time-value {
          font-family: monospace;
          color: var(--color-accent);
          font-weight: 600;
        }

        .date-value {
          color: var(--color-text-secondary);
        }

        .header-status {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 4px;
        }

        .connection-indicator {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .status-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          animation: pulse 2s infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .status-text {
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .last-update {
          font-size: 0.75rem;
          color: var(--color-text-muted);
        }

        /* Content */
        .dashboard-content {
          flex: 1;
          overflow-y: auto;
          padding: 24px;
        }

        /* Sections */
        .section-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
        }

        .section-header h2 {
          margin: 0;
          color: var(--color-text-primary);
          font-size: 1.1rem;
        }

        .refresh-btn,
        .map-btn {
          padding: 6px 12px;
          background: rgba(0, 212, 170, 0.1);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          color: var(--color-accent);
          font-size: 0.85rem;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .refresh-btn:hover,
        .map-btn:hover {
          background: rgba(0, 212, 170, 0.2);
        }

        /* Metrics Grid */
        .metrics-section {
          margin-bottom: 32px;
        }

        .metrics-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 20px;
        }

        .metric-card {
          position: relative;
          padding: 24px;
          background: rgba(10, 12, 16, 0.6);
          backdrop-filter: blur(20px);
          border: 1px solid rgba(255, 255, 255, 0.05);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
          overflow: hidden;
        }
        
        .metric-glow {
          position: absolute;
          top: -50%;
          left: -50%;
          width: 200%;
          height: 200%;
          opacity: 0;
          transition: opacity 0.3s ease;
          pointer-events: none;
        }

        .metric-card:hover {
          transform: translateY(-4px) scale(1.02);
          border-color: rgba(255, 255, 255, 0.1);
          box-shadow: 
            0 10px 40px rgba(0, 0, 0, 0.3),
            0 0 0 1px rgba(255, 255, 255, 0.05) inset;
        }
        
        .metric-card:hover .metric-glow {
          opacity: 1;
        }

        .metric-card.selected {
          border-color: rgba(0, 212, 170, 0.3);
          background: rgba(0, 212, 170, 0.02);
          box-shadow: 
            0 10px 40px rgba(0, 0, 0, 0.4),
            0 0 20px rgba(0, 212, 170, 0.1),
            0 0 0 1px rgba(0, 212, 170, 0.2) inset;
        }
        
        .metric-icon-container {
          position: absolute;
          top: 20px;
          right: 20px;
        }
        
        .metric-icon-bg {
          width: 40px;
          height: 40px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.3s ease;
        }
        
        .metric-card:hover .metric-icon-bg {
          transform: rotate(5deg) scale(1.1);
        }
        
        .metric-content {
          position: relative;
          z-index: 1;
        }
        
        .metric-label {
          font-size: 0.75rem;
          font-weight: 500;
          color: var(--color-text-muted);
          text-transform: uppercase;
          letter-spacing: 0.1em;
          margin-bottom: 8px;
        }
        
        .metric-value-row {
          display: flex;
          align-items: baseline;
          gap: 12px;
          margin-bottom: 12px;
        }
        
        .metric-value {
          font-size: 2.25rem;
          font-weight: 300;
          color: var(--color-text-primary);
          line-height: 1;
          font-family: var(--font-display);
        }
        
        .metric-unit {
          font-size: 1rem;
          font-weight: 400;
          color: var(--color-text-secondary);
          margin-left: 4px;
        }
        
        .metric-trend {
          display: flex;
          align-items: center;
          gap: 4px;
          padding: 4px 8px;
          border-radius: 20px;
          font-size: 0.7rem;
          font-weight: 600;
        }
        
        .metric-trend.up {
          background: rgba(46, 213, 115, 0.1);
          color: #2ed573;
        }
        
        .metric-trend.down {
          background: rgba(255, 71, 87, 0.1);
          color: #ff4757;
        }
        
        .metric-trend.neutral {
          background: rgba(87, 96, 111, 0.1);
          color: #57606f;
        }
        
        .trend-value {
          font-family: var(--font-mono);
        }
        
        .metric-progress {
          position: relative;
          height: 3px;
          background: rgba(255, 255, 255, 0.05);
          border-radius: 3px;
          overflow: hidden;
        }
        
        .metric-progress-bar {
          height: 100%;
          border-radius: 3px;
          transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
          box-shadow: 0 0 10px currentColor;
        }

        /* Main Grid */
        .main-grid {
          display: grid;
          grid-template-columns: 2fr 1fr;
          gap: 24px;
          margin-bottom: 24px;
          min-height: 400px;
        }

        /* Map Section */
        .map-section {
          display: flex;
          flex-direction: column;
        }

        .map-controls {
          display: flex;
          gap: 8px;
        }

        .map-container {
          flex: 1;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.3) 0%, 
            rgba(15, 20, 25, 0.5) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
          overflow: hidden;
          min-height: 400px;
        }

        /* Activity Feed */
        .activity-section {
          display: flex;
          flex-direction: column;
        }

        .activity-count {
          font-size: 0.75rem;
          color: var(--color-text-muted);
          padding: 4px 8px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 12px;
        }

        .activity-feed {
          flex: 1;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.3) 0%, 
            rgba(15, 20, 25, 0.5) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
          padding: 16px;
          overflow-y: auto;
          max-height: 400px;
        }

        .no-activity {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 200px;
          color: var(--color-text-muted);
        }

        .no-activity-icon {
          font-size: 2rem;
          margin-bottom: 8px;
        }

        .activity-item {
          display: flex;
          gap: 12px;
          padding: 12px;
          margin-bottom: 12px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 6px;
          transition: all 0.2s ease;
        }

        .activity-item:hover {
          background: rgba(0, 212, 170, 0.05);
        }

        .activity-icon {
          font-size: 1.2rem;
          flex-shrink: 0;
        }

        .activity-content {
          flex: 1;
        }

        .activity-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 4px;
        }

        .activity-header h4 {
          margin: 0;
          color: var(--color-text-primary);
          font-size: 0.9rem;
        }

        .priority-badge {
          padding: 2px 8px;
          border-radius: 4px;
          font-size: 0.65rem;
          font-weight: 600;
          text-transform: uppercase;
        }

        .activity-description {
          margin: 0 0 4px 0;
          color: var(--color-text-secondary);
          font-size: 0.85rem;
          line-height: 1.4;
        }

        .activity-time {
          font-size: 0.75rem;
          color: var(--color-text-muted);
        }

        /* Quick Actions */
        .actions-section {
          margin-top: auto;
        }

        .quick-actions {
          display: flex;
          gap: 12px;
          padding: 20px;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.4) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-radius: 8px;
        }

        .action-btn {
          flex: 1;
          padding: 12px 16px;
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: 6px;
          color: var(--color-text-secondary);
          font-size: 0.85rem;
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
        }

        .action-btn:hover {
          background: rgba(0, 212, 170, 0.1);
          color: var(--color-accent);
        }

        .action-btn.primary {
          background: linear-gradient(135deg, 
            rgba(255, 71, 87, 0.2) 0%, 
            rgba(255, 71, 87, 0.1) 100%);
          border-color: rgba(255, 71, 87, 0.3);
          color: #ff4757;
        }

        .action-btn.primary:hover {
          background: linear-gradient(135deg, 
            rgba(255, 71, 87, 0.3) 0%, 
            rgba(255, 71, 87, 0.2) 100%);
        }

        .action-btn span {
          font-size: 1.1rem;
        }

        /* Responsive */
        @media (max-width: 1200px) {
          .main-grid {
            grid-template-columns: 1fr;
          }

          .map-container {
            min-height: 300px;
          }
        }

        @media (max-width: 768px) {
          .dashboard-content {
            padding: 16px;
          }

          .metrics-grid {
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          }

          .quick-actions {
            flex-wrap: wrap;
          }

          .action-btn {
            min-width: calc(50% - 6px);
          }
        }
      `}</style>
    </div>
  );
};

export default DashboardNew;
