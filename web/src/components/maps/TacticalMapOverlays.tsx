import React, { useState, useEffect, useRef } from 'react';
import { useTacticalOverlays } from '../../hooks/useTacticalOverlays';
import useDrawingHandler from '../../hooks/useDrawingHandler';
import DrawingTools from './DrawingTools';
import LayerPanel from './LayerPanel';
import type { DrawingToolType } from '../../types/tactical';
import './TacticalMapOverlays.css';

interface TacticalMapOverlaysProps {
  map: L.Map | null;
  className?: string;
}

export const TacticalMapOverlays: React.FC<TacticalMapOverlaysProps> = ({ 
  map,
  className = '',
}) => {
  // State for UI panels
  const [layerPanelCollapsed, setLayerPanelCollapsed] = useState(false);
  const [drawingToolsCollapsed, setDrawingToolsCollapsed] = useState(false);
  
  // Initialize tactical overlay manager
  const overlayManager = useTacticalOverlays({ 
    map,
    onOverlayCreated: (overlay) => {
      console.log('Overlay created:', overlay);
    },
    onOverlayUpdated: (overlay) => {
      console.log('Overlay updated:', overlay);
    },
    onOverlayDeleted: (overlayId) => {
      console.log('Overlay deleted:', overlayId);
    },
  });
  
  // Initialize drawing handler
  const drawingHandler = useDrawingHandler({
    map,
    overlayManager,
    activeDrawingTool: overlayManager.overlayManager.activeDrawingTool,
    onDrawComplete: (overlay) => {
      console.log('Drawing completed:', overlay);
    },
  });
  
  // Handle drawing tool selection
  const handleToolSelect = (toolType: DrawingToolType) => {
    overlayManager.setActiveDrawingTool(toolType);
  };
  
  // Handle tool deselection
  const handleToolDeselect = () => {
    overlayManager.setActiveDrawingTool(undefined);
    drawingHandler.stopDrawing();
  };
  
  // Keyboard shortcuts for drawing tools
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // ESC key to cancel drawing
      if (e.key === 'Escape') {
        if (overlayManager.overlayManager.drawingMode) {
          drawingHandler.cancelDrawing();
        }
      }
      
      // Delete key to delete selected overlay
      if (e.key === 'Delete') {
        if (overlayManager.selectedOverlay) {
          overlayManager.deleteOverlay(overlayManager.selectedOverlay.id);
        }
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [overlayManager, drawingHandler]);
  
  return (
    <div className={`tactical-map-overlays ${className}`}>
      {/* Drawing Tools Panel */}
      <DrawingTools 
        onToolSelect={handleToolSelect}
        onToolDeselect={handleToolDeselect}
        activeToolType={overlayManager.overlayManager.activeDrawingTool}
        isDrawing={drawingHandler.isDrawing}
        collapsed={drawingToolsCollapsed}
        onToggle={() => setDrawingToolsCollapsed(!drawingToolsCollapsed)}
      />
      
      {/* Layer Management Panel */}
      <LayerPanel 
        overlayManager={overlayManager}
        collapsed={layerPanelCollapsed}
        onToggle={() => setLayerPanelCollapsed(!layerPanelCollapsed)}
      />
      
      {/* Drawing Instructions Overlay (shown when drawing) */}
      {drawingHandler.isDrawing && (
        <div className="drawing-instructions">
          <div className="instructions-content">
            <p>
              <strong>Drawing: {formatToolName(overlayManager.overlayManager.activeDrawingTool)}</strong>
            </p>
            <p>
              {getDrawingInstructions(overlayManager.overlayManager.activeDrawingTool)}
            </p>
            <p className="instructions-shortcuts">
              <kbd>ESC</kbd> to cancel • <kbd>Click</kbd> to place points • <kbd>Double-click</kbd> to finish
            </p>
          </div>
        </div>
      )}
      
      {/* Overlay Info Panel (shown when overlay selected) */}
      {overlayManager.selectedOverlay && (
        <div className="overlay-info-panel">
          <div className="overlay-info-header">
            <h3>{overlayManager.selectedOverlay.name}</h3>
            <button 
              className="close-btn" 
              onClick={() => overlayManager.deselectOverlay()}
            >
              ×
            </button>
          </div>
          
          <div className="overlay-info-content">
            <div className="info-row">
              <span className="info-label">Type:</span>
              <span className="info-value">{formatToolName(overlayManager.selectedOverlay.type as any)}</span>
            </div>
            
            {overlayManager.selectedOverlay.description && (
              <div className="info-row">
                <span className="info-label">Description:</span>
                <span className="info-value">{overlayManager.selectedOverlay.description}</span>
              </div>
            )}
            
            <div className="info-row">
              <span className="info-label">Priority:</span>
              <span className="info-value priority-badge">{overlayManager.selectedOverlay.metadata.priority}</span>
            </div>
            
            {overlayManager.selectedOverlay.metadata.classification && (
              <div className="info-row">
                <span className="info-label">Classification:</span>
                <span className="info-value classification-badge">
                  {overlayManager.selectedOverlay.metadata.classification}
                </span>
              </div>
            )}
            
            <div className="overlay-actions">
              <button 
                className="action-btn" 
                onClick={() => {
                  overlayManager.updateOverlay(overlayManager.selectedOverlay!.id, {
                    visible: !overlayManager.selectedOverlay!.visible
                  });
                }}
              >
                {overlayManager.selectedOverlay.visible ? 'Hide' : 'Show'}
              </button>
              
              <button 
                className="action-btn danger" 
                onClick={() => {
                  if (window.confirm(`Delete overlay "${overlayManager.selectedOverlay!.name}"?`)) {
                    overlayManager.deleteOverlay(overlayManager.selectedOverlay!.id);
                  }
                }}
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Helper function to format tool name for display
function formatToolName(toolType?: DrawingToolType): string {
  if (!toolType) return '';
  
  switch (toolType) {
    case 'threat_circle': return 'Threat Circle';
    case 'range_ring': return 'Range Ring';
    case 'tactical_route': return 'Route';
    case 'tactical_area': return 'Area';
    case 'mil_symbol': return 'Military Symbol';
    default: return toolType.charAt(0).toUpperCase() + toolType.slice(1).replace('_', ' ');
  }
}

// Helper function to get drawing instructions for a tool
function getDrawingInstructions(toolType?: DrawingToolType): string {
  if (!toolType) return '';
  
  switch (toolType) {
    case 'marker':
    case 'symbol':
      return 'Click on the map to place a marker.';
    case 'line':
    case 'route':
      return 'Click to add points. Double-click to finish the line.';
    case 'polygon':
    case 'boundary':
      return 'Click to add points. Double-click to close the polygon.';
    case 'rectangle':
      return 'Click and drag to draw a rectangle.';
    case 'circle':
    case 'threat_circle':
    case 'range_ring':
      return 'Click on the center point, then drag to set the radius.';
    default:
      return 'Click on the map to start drawing.';
  }
}

export default TacticalMapOverlays;
