import React, { useState } from 'react';
import type { 
  OverlayLayer, 
  TacticalOverlay,
  UseTacticalOverlaysReturn 
} from '../../hooks/useTacticalOverlays';
import './LayerPanel.css';

interface LayerPanelProps {
  overlayManager: UseTacticalOverlaysReturn;
  className?: string;
  collapsed?: boolean;
  onToggle?: () => void;
}

interface LayerItemProps {
  layer: OverlayLayer;
  isActive: boolean;
  overlayManager: UseTacticalOverlaysReturn;
}

const LayerItem: React.FC<LayerItemProps> = ({ layer, isActive, overlayManager }) => {
  const [expanded, setExpanded] = useState(false);
  const [showRename, setShowRename] = useState(false);
  const [newName, setNewName] = useState(layer.name);

  const handleRename = (e: React.FormEvent) => {
    e.preventDefault();
    if (newName.trim() && newName !== layer.name) {
      // Update layer name through overlay manager
      // This would require adding an updateLayer function to the hook
      console.log('Rename layer:', layer.id, 'to:', newName);
    }
    setShowRename(false);
  };

  const handleDeleteLayer = () => {
    if (window.confirm(`Are you sure you want to delete layer "${layer.name}"?`)) {
      overlayManager.deleteLayer(layer.id);
    }
  };

  const formatOverlayType = (overlay: TacticalOverlay): string => {
    switch (overlay.type) {
      case 'threat_circle': return 'Threat Circle';
      case 'range_ring': return 'Range Ring';
      case 'tactical_route': return 'Route';
      case 'tactical_area': return 'Area';
      case 'mil_symbol': return 'Symbol';
      default: return overlay.type.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
    }
  };

  const getOverlayIcon = (overlay: TacticalOverlay): string => {
    switch (overlay.type) {
      case 'threat_circle': return '⚠️';
      case 'range_ring': return '📡';
      case 'tactical_route': return '🛣️';
      case 'tactical_area': return '🏢';
      case 'mil_symbol': return '🎖️';
      case 'marker': return '📍';
      case 'line': return '📏';
      case 'polygon': return '⬟';
      case 'rectangle': return '▭';
      case 'circle': return '⭕';
      default: return '📍';
    }
  };

  return (
    <div className={`layer-item ${isActive ? 'active' : ''} ${layer.locked ? 'locked' : ''}`}>
      <div className="layer-header">
        <div className="layer-info">
          <button
            className="layer-toggle"
            onClick={() => setExpanded(!expanded)}
            title={expanded ? 'Collapse layer' : 'Expand layer'}
          >
            {expanded ? '▼' : '▶'}
          </button>
          
          <button
            className="visibility-toggle"
            onClick={() => overlayManager.toggleLayerVisibility(layer.id)}
            title={layer.visible ? 'Hide layer' : 'Show layer'}
          >
            {layer.visible ? '👁️' : '🚫'}
          </button>

          {showRename ? (
            <form onSubmit={handleRename} className="rename-form">
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onBlur={() => setShowRename(false)}
                className="rename-input"
                autoFocus
              />
            </form>
          ) : (
            <span 
              className="layer-name"
              onClick={() => overlayManager.setActiveLayer(layer.id)}
              onDoubleClick={() => setShowRename(true)}
              title={layer.description || 'Double-click to rename'}
            >
              {layer.name}
            </span>
          )}

          <span className="overlay-count">({layer.overlays.length})</span>
        </div>

        <div className="layer-controls">
          <input
            type="range"
            min="0"
            max="1"
            step="0.1"
            value={layer.opacity}
            onChange={(e) => overlayManager.setLayerOpacity(layer.id, parseFloat(e.target.value))}
            className="opacity-slider"
            title={`Opacity: ${Math.round(layer.opacity * 100)}%`}
          />
          
          <div className="layer-actions">
            {layer.id !== 'default' && (
              <button
                className="delete-layer-btn"
                onClick={handleDeleteLayer}
                title="Delete layer"
              >
                🗑️
              </button>
            )}
          </div>
        </div>
      </div>

      {expanded && (
        <div className="overlay-list">
          {layer.overlays.length === 0 ? (
            <div className="empty-layer">No overlays in this layer</div>
          ) : (
            layer.overlays.map(overlay => (
              <div 
                key={overlay.id} 
                className={`overlay-item ${overlay.id === overlayManager.selectedOverlay?.id ? 'selected' : ''}`}
                onClick={() => overlayManager.selectOverlay(overlay.id)}
              >
                <div className="overlay-info">
                  <span className="overlay-icon">{getOverlayIcon(overlay)}</span>
                  <div className="overlay-details">
                    <div className="overlay-name">{overlay.name}</div>
                    <div className="overlay-type">{formatOverlayType(overlay)}</div>
                    {overlay.metadata.classification && (
                      <div className="overlay-classification">{overlay.metadata.classification}</div>
                    )}
                  </div>
                </div>

                <div className="overlay-controls">
                  <button
                    className="visibility-toggle"
                    onClick={(e) => {
                      e.stopPropagation();
                      overlayManager.updateOverlay(overlay.id, { visible: !overlay.visible });
                    }}
                    title={overlay.visible ? 'Hide overlay' : 'Show overlay'}
                  >
                    {overlay.visible ? '👁️' : '🚫'}
                  </button>
                  
                  <button
                    className="delete-overlay-btn"
                    onClick={(e) => {
                      e.stopPropagation();
                      if (window.confirm(`Delete overlay "${overlay.name}"?`)) {
                        overlayManager.deleteOverlay(overlay.id);
                      }
                    }}
                    title="Delete overlay"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export const LayerPanel: React.FC<LayerPanelProps> = ({ 
  overlayManager,
  className = '',
  collapsed = false,
  onToggle 
}) => {
  const [showCreateLayer, setShowCreateLayer] = useState(false);
  const [newLayerName, setNewLayerName] = useState('');
  const [newLayerDescription, setNewLayerDescription] = useState('');
  const [showImportExport, setShowImportExport] = useState(false);

  const handleCreateLayer = (e: React.FormEvent) => {
    e.preventDefault();
    if (newLayerName.trim()) {
      overlayManager.createLayer(
        newLayerName.trim(), 
        newLayerDescription.trim() || undefined
      );
      setNewLayerName('');
      setNewLayerDescription('');
      setShowCreateLayer(false);
    }
  };

  const handleExport = () => {
    const data = overlayManager.exportOverlays();
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `tactical-overlays-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const data = event.target?.result as string;
        overlayManager.importOverlays(data);
      };
      reader.readAsText(file);
    }
  };

  const totalOverlays = overlayManager.overlays.length;
  const visibleLayers = overlayManager.overlayManager.layers.filter(layer => layer.visible).length;

  if (collapsed) {
    return (
      <div className={`layer-panel collapsed ${className}`}>
        <button 
          className="panel-toggle"
          onClick={onToggle}
          title="Expand layer panel"
        >
          📋 ({totalOverlays})
        </button>
      </div>
    );
  }

  return (
    <div className={`layer-panel ${className}`}>
      <div className="panel-header">
        <div className="panel-title">
          <h3>Layers</h3>
          <button 
            className="panel-toggle"
            onClick={onToggle}
            title="Collapse panel"
          >
            ×
          </button>
        </div>
        
        <div className="panel-stats">
          <span>{overlayManager.overlayManager.layers.length} layers</span>
          <span>{totalOverlays} overlays</span>
          <span>{visibleLayers} visible</span>
        </div>
      </div>

      <div className="panel-actions">
        <button
          className="action-btn primary"
          onClick={() => setShowCreateLayer(true)}
          title="Create new layer"
        >
          + New Layer
        </button>
        
        <button
          className="action-btn"
          onClick={() => setShowImportExport(!showImportExport)}
          title="Import/Export layers"
        >
          ⚙️ I/E
        </button>
        
        <button
          className="action-btn danger"
          onClick={() => {
            if (window.confirm('Clear all overlays? This cannot be undone.')) {
              overlayManager.clearAllOverlays();
            }
          }}
          title="Clear all overlays"
        >
          🗑️ Clear
        </button>
      </div>

      {showCreateLayer && (
        <div className="create-layer-form">
          <form onSubmit={handleCreateLayer}>
            <input
              type="text"
              placeholder="Layer name"
              value={newLayerName}
              onChange={(e) => setNewLayerName(e.target.value)}
              className="form-input"
              autoFocus
            />
            <textarea
              placeholder="Description (optional)"
              value={newLayerDescription}
              onChange={(e) => setNewLayerDescription(e.target.value)}
              className="form-textarea"
              rows={2}
            />
            <div className="form-actions">
              <button type="submit" className="btn primary">Create</button>
              <button 
                type="button" 
                onClick={() => setShowCreateLayer(false)}
                className="btn secondary"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {showImportExport && (
        <div className="import-export-controls">
          <button 
            className="btn secondary full-width"
            onClick={handleExport}
          >
            📤 Export All Layers
          </button>
          
          <label className="btn secondary full-width file-input-label">
            📥 Import Layers
            <input
              type="file"
              accept=".json"
              onChange={handleImport}
              className="file-input"
            />
          </label>
        </div>
      )}

      <div className="layer-list">
        {overlayManager.overlayManager.layers.map(layer => (
          <LayerItem
            key={layer.id}
            layer={layer}
            isActive={layer.id === overlayManager.activeLayer?.id}
            overlayManager={overlayManager}
          />
        ))}
      </div>
    </div>
  );
};

export default LayerPanel;
