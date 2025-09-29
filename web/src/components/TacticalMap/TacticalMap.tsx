/**
 * Tactical Map Component (Stub)
 * High-performance map component for tactical visualization
 */

import React, { memo } from 'react';
import { useEntityCount } from '../../stores/simpleTacticalStore';

interface TacticalMapProps {
  fullscreen?: boolean;
}

export const TacticalMap: React.FC<TacticalMapProps> = memo(({ fullscreen = false }) => {
  const entityCount = useEntityCount();
  
  // Default map viewport for display
  const mapViewport = {
    center: { lat: 38.9072, lng: -77.0369 },
    zoom: 12
  };

  return (
    <div className={`tactical-map ${fullscreen ? 'fullscreen' : ''}`}>
      <div className="map-placeholder">
        <div className="map-info">
          <h2>🗺️ Tactical Map</h2>
          <p>Interactive tactical map will be implemented here</p>
          <div className="map-stats">
            <div>Center: {mapViewport.center.lat.toFixed(4)}, {mapViewport.center.lng.toFixed(4)}</div>
            <div>Zoom: {mapViewport.zoom}</div>
            <div>Entities: {entityCount}</div>
            <div>Mode: {fullscreen ? 'Fullscreen' : 'Normal'}</div>
          </div>
        </div>
        <div className="map-controls">
          <button className="map-btn">🔍 Zoom In</button>
          <button className="map-btn">🔍 Zoom Out</button>
          <button className="map-btn">📍 Center</button>
          <button className="map-btn">📐 Measure</button>
        </div>
      </div>
      
      <style>{`
        .tactical-map {
          width: 100%;
          height: 100%;
          background: linear-gradient(135deg, #1a1d20 0%, #0a0d10 100%);
          border: 1px solid #404040;
          border-radius: 8px;
          overflow: hidden;
          position: relative;
        }
        
        .tactical-map.fullscreen {
          border-radius: 0;
          border: none;
        }
        
        .map-placeholder {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100%;
          padding: 2rem;
          text-align: center;
          color: #dcddde;
        }
        
        .map-info h2 {
          color: #00d4aa;
          margin-bottom: 1rem;
          font-size: 2rem;
        }
        
        .map-info p {
          color: #b0b3b8;
          margin-bottom: 2rem;
          font-size: 1.1rem;
        }
        
        .map-stats {
          background: rgba(0, 0, 0, 0.3);
          border: 1px solid #404040;
          border-radius: 8px;
          padding: 1rem;
          margin-bottom: 2rem;
          font-family: 'Roboto Mono', monospace;
          font-size: 0.9rem;
        }
        
        .map-stats div {
          margin: 0.25rem 0;
        }
        
        .map-controls {
          display: flex;
          gap: 1rem;
          flex-wrap: wrap;
          justify-content: center;
        }
        
        .map-btn {
          background: #2a2d30;
          border: 1px solid #404040;
          color: #dcddde;
          padding: 0.75rem 1.5rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.15s ease;
          font-size: 0.9rem;
        }
        
        .map-btn:hover {
          background: #36393f;
          border-color: #00d4aa;
          color: #00d4aa;
        }
      `}</style>
    </div>
  );
});

TacticalMap.displayName = 'TacticalMap';
