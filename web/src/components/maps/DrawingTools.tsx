import { useState, useCallback } from 'react';
import type { DrawingTool, DrawingToolType, TacticalOverlayType } from '../../types/tactical';
import './DrawingTools.css';

export interface DrawingToolsProps {
  activeToolType?: DrawingToolType;
  onToolSelect?: (toolType: DrawingToolType) => void;
  onToolDeselect?: () => void;
  disabled?: boolean;
  className?: string;
}

// Available drawing tools configuration
const DRAWING_TOOLS: DrawingTool[] = [
  {
    id: 'marker',
    name: 'Marker',
    icon: '📍',
    type: 'marker',
    enabled: true,
    options: {
      style: { color: '#f97316', weight: 2, opacity: 1 },
    },
  },
  {
    id: 'line',
    name: 'Line',
    icon: '📏',
    type: 'line',
    enabled: true,
    options: {
      style: { color: '#8b5cf6', weight: 3, opacity: 0.8 },
      showLength: true,
      metric: true,
    },
  },
  {
    id: 'polygon',
    name: 'Area',
    icon: '⬡',
    type: 'polygon',
    enabled: true,
    options: {
      style: { 
        color: '#ec4899', 
        fillColor: '#ec4899', 
        weight: 2, 
        opacity: 0.8, 
        fillOpacity: 0.15 
      },
      showArea: true,
      metric: true,
    },
  },
  {
    id: 'rectangle',
    name: 'Rectangle',
    icon: '⬛',
    type: 'rectangle',
    enabled: true,
    options: {
      style: { 
        color: '#06b6d4', 
        fillColor: '#06b6d4', 
        weight: 2, 
        opacity: 0.8, 
        fillOpacity: 0.1 
      },
      showArea: true,
    },
  },
  {
    id: 'circle',
    name: 'Circle',
    icon: '⭕',
    type: 'circle',
    enabled: true,
    options: {
      style: { 
        color: '#06b6d4', 
        fillColor: '#06b6d4', 
        weight: 2, 
        opacity: 0.8, 
        fillOpacity: 0.1 
      },
      showArea: true,
      metric: true,
    },
  },
  {
    id: 'route',
    name: 'Route',
    icon: '🛣️',
    type: 'route',
    enabled: true,
    options: {
      style: { 
        color: '#3b82f6', 
        weight: 4, 
        opacity: 0.8,
        lineCap: 'round',
        lineJoin: 'round'
      },
      showLength: true,
      metric: true,
    },
  },
  {
    id: 'boundary',
    name: 'Boundary',
    icon: '🚧',
    type: 'boundary',
    enabled: true,
    options: {
      style: { 
        color: '#f59e0b', 
        weight: 3, 
        opacity: 0.9,
        dashArray: '15,5',
        lineCap: 'butt'
      },
      showLength: true,
    },
  },
  {
    id: 'threat_circle',
    name: 'Threat Circle',
    icon: '⚠️',
    type: 'threat_circle',
    enabled: true,
    options: {
      style: { 
        color: '#ef4444', 
        fillColor: '#ef4444', 
        weight: 2, 
        opacity: 0.8,
        fillOpacity: 0.1,
        dashArray: '5,5'
      },
      showArea: true,
      metric: true,
    },
  },
  {
    id: 'range_ring',
    name: 'Range Ring',
    icon: '🎯',
    type: 'range_ring',
    enabled: true,
    options: {
      style: { 
        color: '#6366f1', 
        weight: 2, 
        opacity: 0.7,
        dashArray: '10,10',
        fillOpacity: 0
      },
      metric: true,
    },
  },
  {
    id: 'symbol',
    name: 'Symbol',
    icon: '🔰',
    type: 'symbol',
    enabled: true,
    options: {
      style: { color: '#ffffff', weight: 2, opacity: 1 },
    },
  },
];

// Tool groups for better organization
const TOOL_GROUPS = [
  {
    id: 'basic',
    name: 'Basic',
    tools: ['marker', 'line', 'polygon', 'rectangle', 'circle'],
  },
  {
    id: 'tactical',
    name: 'Tactical',
    tools: ['route', 'boundary', 'threat_circle', 'range_ring', 'symbol'],
  },
];

export function DrawingTools({ 
  activeToolType, 
  onToolSelect, 
  onToolDeselect, 
  disabled = false,
  className = ''
}: DrawingToolsProps) {
  const [expandedGroup, setExpandedGroup] = useState<string>('basic');
  const [isCollapsed, setIsCollapsed] = useState<boolean>(false);

  const handleToolClick = useCallback((tool: DrawingTool) => {
    if (disabled) return;
    
    if (activeToolType === tool.type) {
      // Deselect if clicking the same tool
      onToolDeselect?.();
    } else {
      // Select the new tool
      onToolSelect?.(tool.type);
    }
  }, [activeToolType, onToolSelect, onToolDeselect, disabled]);

  const handleGroupToggle = useCallback((groupId: string) => {
    setExpandedGroup(expandedGroup === groupId ? '' : groupId);
  }, [expandedGroup]);

  const getToolsForGroup = useCallback((groupId: string) => {
    const group = TOOL_GROUPS.find(g => g.id === groupId);
    if (!group) return [];
    
    return DRAWING_TOOLS.filter(tool => 
      group.tools.includes(tool.id) && tool.enabled
    );
  }, []);

  return (
    <div className={`drawing-tools ${className} ${isCollapsed ? 'collapsed' : ''}`}>
      <div className="drawing-tools-header">
        <h3>Drawing Tools</h3>
        <button
          className="collapse-btn"
          onClick={() => setIsCollapsed(!isCollapsed)}
          title={isCollapsed ? 'Expand tools' : 'Collapse tools'}
        >
          {isCollapsed ? '▶' : '▼'}
        </button>
      </div>

      {!isCollapsed && (
        <div className="drawing-tools-content">
          {/* Clear/Stop Drawing Button */}
          <div className="tool-group-header">
            <button
              className={`clear-tool-btn ${activeToolType ? 'active' : ''}`}
              onClick={() => onToolDeselect?.()}
              disabled={disabled || !activeToolType}
              title="Stop drawing / Clear selection"
            >
              🚫 Stop Drawing
            </button>
          </div>

          {/* Tool Groups */}
          {TOOL_GROUPS.map((group) => {
            const groupTools = getToolsForGroup(group.id);
            if (groupTools.length === 0) return null;

            return (
              <div key={group.id} className="tool-group">
                <button
                  className={`tool-group-header ${expandedGroup === group.id ? 'expanded' : ''}`}
                  onClick={() => handleGroupToggle(group.id)}
                >
                  <span>{group.name}</span>
                  <span className="expand-icon">
                    {expandedGroup === group.id ? '▼' : '▶'}
                  </span>
                </button>

                {expandedGroup === group.id && (
                  <div className="tool-group-content">
                    {groupTools.map((tool) => (
                      <button
                        key={tool.id}
                        className={`tool-btn ${activeToolType === tool.type ? 'active' : ''}`}
                        onClick={() => handleToolClick(tool)}
                        disabled={disabled || !tool.enabled}
                        title={tool.name}
                        data-tool-type={tool.type}
                      >
                        <span className="tool-icon">{tool.icon}</span>
                        <span className="tool-name">{tool.name}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            );
          })}

          {/* Active Tool Info */}
          {activeToolType && (
            <div className="active-tool-info">
              <div className="active-tool-header">Active Tool:</div>
              <div className="active-tool-details">
                {(() => {
                  const activeTool = DRAWING_TOOLS.find(t => t.type === activeToolType);
                  return activeTool ? (
                    <>
                      <span className="active-tool-icon">{activeTool.icon}</span>
                      <span className="active-tool-name">{activeTool.name}</span>
                    </>
                  ) : (
                    <span>Unknown Tool</span>
                  );
                })()}
              </div>
              
              {/* Tool-specific instructions */}
              <div className="tool-instructions">
                {getToolInstructions(activeToolType)}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Helper function to get tool-specific instructions
function getToolInstructions(toolType: DrawingToolType): string {
  switch (toolType) {
    case 'marker':
      return 'Click on the map to place a marker';
    case 'line':
      return 'Click to start, click additional points, double-click to finish';
    case 'polygon':
      return 'Click to start drawing area, click additional points, double-click to close';
    case 'rectangle':
      return 'Click and drag to draw rectangle';
    case 'circle':
      return 'Click center then drag to set radius';
    case 'route':
      return 'Click waypoints in order, double-click to finish route';
    case 'boundary':
      return 'Click to draw boundary line, double-click to finish';
    case 'threat_circle':
      return 'Click center of threat area, drag to set threat radius';
    case 'range_ring':
      return 'Click center point, drag to set first range ring';
    case 'symbol':
      return 'Click to place tactical symbol on map';
    default:
      return 'Follow on-screen prompts for this tool';
  }
}

// Get tool configuration by type
export function getDrawingTool(toolType: DrawingToolType): DrawingTool | undefined {
  return DRAWING_TOOLS.find(tool => tool.type === toolType);
}

// Get all available drawing tools
export function getAllDrawingTools(): DrawingTool[] {
  return DRAWING_TOOLS.filter(tool => tool.enabled);
}

// Get tools by group
export function getDrawingToolsByGroup(groupId: string): DrawingTool[] {
  const group = TOOL_GROUPS.find(g => g.id === groupId);
  if (!group) return [];
  
  return DRAWING_TOOLS.filter(tool => 
    group.tools.includes(tool.id) && tool.enabled
  );
}
