/**
 * Performance Monitor Component
 * Shows real-time performance metrics in debug mode
 */

import React, { memo, useState, useEffect } from 'react';

interface PerformanceMetrics {
  entityUpdateCount: number;
  lastUpdateTime: number;
  frameRate: number;
  memoryUsage?: number;
}

interface PerformanceMonitorProps {
  metrics: PerformanceMetrics;
}

const PerformanceMonitor: React.FC<PerformanceMonitorProps> = memo(({ metrics }) => {
  const [isMinimized, setIsMinimized] = useState(false);
  const [memoryInfo, setMemoryInfo] = useState<any>(null);

  useEffect(() => {
    // Get memory usage if available (Chrome)
    const updateMemoryInfo = () => {
      if ('memory' in performance) {
        setMemoryInfo((performance as any).memory);
      }
    };

    updateMemoryInfo();
    const interval = setInterval(updateMemoryInfo, 5000); // Update every 5 seconds

    return () => clearInterval(interval);
  }, []);

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTime = (timestamp: number) => {
    const now = Date.now();
    const diff = now - timestamp;
    return diff < 1000 ? `${diff}ms ago` : `${(diff / 1000).toFixed(1)}s ago`;
  };

  if (isMinimized) {
    return (
      <div className="performance-monitor minimized">
        <button onClick={() => setIsMinimized(false)}>
          📊 Performance
        </button>
      </div>
    );
  }

  return (
    <div className="performance-monitor">
      <div className="monitor-header">
        <h3>🔧 Performance Monitor</h3>
        <button onClick={() => setIsMinimized(true)}>−</button>
      </div>
      
      <div className="perf-metric">
        <span className="label">Frame Rate:</span>
        <span className="value">{metrics.frameRate.toFixed(1)} FPS</span>
      </div>
      
      <div className="perf-metric">
        <span className="label">Entity Updates:</span>
        <span className="value">{metrics.entityUpdateCount}</span>
      </div>
      
      <div className="perf-metric">
        <span className="label">Last Update:</span>
        <span className="value">{formatTime(metrics.lastUpdateTime)}</span>
      </div>
      
      {memoryInfo && (
        <>
          <div className="perf-metric">
            <span className="label">Used Memory:</span>
            <span className="value">{formatBytes(memoryInfo.usedJSHeapSize)}</span>
          </div>
          
          <div className="perf-metric">
            <span className="label">Total Memory:</span>
            <span className="value">{formatBytes(memoryInfo.totalJSHeapSize)}</span>
          </div>
          
          <div className="perf-metric">
            <span className="label">Memory Limit:</span>
            <span className="value">{formatBytes(memoryInfo.jsHeapSizeLimit)}</span>
          </div>
        </>
      )}
      
      <div className="performance-bars">
        {memoryInfo && (
          <div className="memory-bar">
            <div className="bar-label">Memory Usage</div>
            <div className="bar-container">
              <div 
                className="bar-fill" 
                style={{ 
                  width: `${(memoryInfo.usedJSHeapSize / memoryInfo.jsHeapSizeLimit) * 100}%`,
                  backgroundColor: memoryInfo.usedJSHeapSize / memoryInfo.jsHeapSizeLimit > 0.8 ? '#f44336' : '#00d4aa'
                }}
              />
            </div>
          </div>
        )}
        
        <div className="fps-bar">
          <div className="bar-label">Frame Rate</div>
          <div className="bar-container">
            <div 
              className="bar-fill" 
              style={{ 
                width: `${(metrics.frameRate / 60) * 100}%`,
                backgroundColor: metrics.frameRate < 30 ? '#f44336' : metrics.frameRate < 45 ? '#ff9800' : '#00d4aa'
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
});

PerformanceMonitor.displayName = 'PerformanceMonitor';

export default PerformanceMonitor;
