/**
 * Settings Page - Complete Mobile-First Redesign
 * Tactical interface with card-based layout and responsive design
 */

import React, { useState, useCallback, useEffect, useMemo } from 'react';
import { Icon } from '../components/ui/Icon';
import './Settings.css';

interface SettingSection {
  id: string;
  title: string;
  icon: string;
  description: string;
  badge?: string;
}

interface Setting {
  id: string;
  label: string;
  description?: string;
  type: 'toggle' | 'select' | 'input' | 'slider' | 'color' | 'link' | 'range';
  value: any;
  options?: { value: string; label: string }[];
  min?: number;
  max?: number;
  step?: number;
  unit?: string;
  disabled?: boolean;
  warning?: string;
}

// Settings sections with tactical icons - moved outside component to prevent recreation
const SETTINGS_SECTIONS: SettingSection[] = [
  { 
    id: 'general', 
    title: 'General', 
    icon: 'settings', 
    description: 'Basic application settings',
  },
  { 
    id: 'network', 
    title: 'Network', 
    icon: 'signal', 
    description: 'TAK server and connectivity',
    badge: 'CONNECTED'
  },
  { 
    id: 'display', 
    title: 'Display', 
    icon: 'monitor', 
    description: 'Interface and appearance'
  },
  { 
    id: 'map', 
    title: 'Map', 
    icon: 'map', 
    description: 'Map layers and visualization'
  },
  { 
    id: 'security', 
    title: 'Security', 
    icon: 'shield', 
    description: 'Authentication and encryption'
  },
  { 
    id: 'notifications', 
    title: 'Notifications', 
    icon: 'bell', 
    description: 'Alerts and sound preferences'
  },
  { 
    id: 'data', 
    title: 'Data & Storage', 
    icon: 'database', 
    description: 'Sync, backup, and export'
  },
  { 
    id: 'advanced', 
    title: 'Advanced', 
    icon: 'code', 
    description: 'Developer and debug options'
  },
];

const Settings: React.FC = () => {
  const [activeSection, setActiveSection] = useState('general');
  const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle');

  // Handle responsive design
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Settings data structure with tactical focus
  const [settings, setSettings] = useState<Record<string, Setting[]>>({
    general: [
      { 
        id: 'callsign', 
        label: 'Callsign', 
        description: 'Your tactical identifier for communications',
        type: 'input', 
        value: 'ALPHA-1' 
      },
      { 
        id: 'team', 
        label: 'Force Assignment', 
        description: 'Select your operational force designation',
        type: 'select', 
        value: 'blue', 
        options: [
          { value: 'blue', label: 'Blue Force (Friendly)' },
          { value: 'red', label: 'Red Force (OPFOR)' },
          { value: 'green', label: 'Green Force (Partner)' },
          { value: 'white', label: 'White (Neutral/Observer)' },
        ]
      },
      { 
        id: 'autoConnect', 
        label: 'Auto-connect on startup', 
        description: 'Automatically connect to TAK server when app launches',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'keepAlive', 
        label: 'Maintain connection', 
        description: 'Keep connection alive during inactivity',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'language', 
        label: 'Interface Language', 
        description: 'Select display language for the interface',
        type: 'select', 
        value: 'en',
        options: [
          { value: 'en', label: 'English' },
          { value: 'es', label: 'Español' },
          { value: 'fr', label: 'Français' },
          { value: 'de', label: 'Deutsch' },
        ]
      },
    ],
    network: [
      { 
        id: 'serverUrl', 
        label: 'TAK Server URL', 
        description: 'Primary TAK server endpoint',
        type: 'input', 
        value: 'wss://tak.server.mil:8089' 
      },
      { 
        id: 'protocol', 
        label: 'Connection Protocol', 
        description: 'Select connection protocol (WSS recommended)',
        type: 'select', 
        value: 'wss',
        options: [
          { value: 'wss', label: 'WebSocket Secure (WSS)' },
          { value: 'ws', label: 'WebSocket (WS)' },
          { value: 'tcp', label: 'TCP Direct' },
        ]
      },
      { 
        id: 'reconnectInterval', 
        label: 'Reconnect Interval', 
        description: 'Time between reconnection attempts',
        type: 'range', 
        value: 30, 
        min: 5, 
        max: 120, 
        step: 5,
        unit: 'seconds'
      },
      { 
        id: 'timeout', 
        label: 'Connection Timeout', 
        description: 'Maximum time to wait for connection',
        type: 'range', 
        value: 10, 
        min: 5, 
        max: 60, 
        step: 5,
        unit: 'seconds'
      },
      { 
        id: 'heartbeat', 
        label: 'Heartbeat Interval', 
        description: 'Keep-alive signal frequency',
        type: 'range', 
        value: 30, 
        min: 10, 
        max: 300, 
        step: 10,
        unit: 'seconds'
      },
      { 
        id: 'compression', 
        label: 'Enable Compression', 
        description: 'Compress data to reduce bandwidth usage',
        type: 'toggle', 
        value: true 
      },
    ],
    display: [
      { 
        id: 'theme', 
        label: 'Interface Theme', 
        description: 'Select visual theme (Tactical Dark recommended)',
        type: 'select', 
        value: 'dark',
        options: [
          { value: 'dark', label: 'Tactical Dark' },
          { value: 'light', label: 'Light Mode' },
          { value: 'auto', label: 'System Preference' },
        ]
      },
      { 
        id: 'accentColor', 
        label: 'Accent Color', 
        description: 'Primary accent color for highlights and indicators',
        type: 'color', 
        value: '#00d4aa' 
      },
      { 
        id: 'fontSize', 
        label: 'Font Size', 
        description: 'Interface text size for readability',
        type: 'select', 
        value: 'medium',
        options: [
          { value: 'small', label: 'Small' },
          { value: 'medium', label: 'Medium (Recommended)' },
          { value: 'large', label: 'Large' },
          { value: 'xlarge', label: 'Extra Large' },
        ]
      },
      { 
        id: 'animations', 
        label: 'Enable Animations', 
        description: 'Smooth transitions and micro-interactions',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'reducedMotion', 
        label: 'Reduce Motion', 
        description: 'Minimize animations for accessibility',
        type: 'toggle', 
        value: false 
      },
      { 
        id: 'compactMode', 
        label: 'Compact Interface', 
        description: 'Reduce spacing for information density',
        type: 'toggle', 
        value: false 
      },
    ],
    map: [
      { 
        id: 'mapProvider', 
        label: 'Map Provider', 
        description: 'Base map tile source for terrain visualization',
        type: 'select', 
        value: 'osm',
        options: [
          { value: 'osm', label: 'OpenStreetMap' },
          { value: 'satellite', label: 'Satellite Imagery' },
          { value: 'terrain', label: 'Terrain Topology' },
          { value: 'tactical', label: 'Tactical Overlay' },
        ]
      },
      { 
        id: 'defaultZoom', 
        label: 'Default Zoom Level', 
        description: 'Initial map zoom when loading',
        type: 'range', 
        value: 13, 
        min: 1, 
        max: 20, 
        step: 1 
      },
      { 
        id: 'trackingMode', 
        label: 'Position Tracking', 
        description: 'How frequently to update your position',
        type: 'select', 
        value: 'auto',
        options: [
          { value: 'auto', label: 'Automatic' },
          { value: 'continuous', label: 'Continuous' },
          { value: 'periodic', label: 'Periodic (5min)' },
          { value: 'manual', label: 'Manual Only' },
        ]
      },
      { 
        id: 'showGrid', 
        label: 'MGRS Grid Overlay', 
        description: 'Display military grid reference system',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'showCompass', 
        label: 'Compass Rose', 
        description: 'Show directional compass on map',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'entityLabels', 
        label: 'Entity Labels', 
        description: 'Display callsigns on map markers',
        type: 'toggle', 
        value: true 
      },
    ],
    notifications: [
      { 
        id: 'enableNotifications', 
        label: 'Enable Notifications', 
        description: 'Allow desktop and system notifications',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'emergencyAlerts', 
        label: 'Emergency Alerts', 
        description: 'High-priority alerts for critical situations',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'chatAlerts', 
        label: 'Message Notifications', 
        description: 'Notifications for new chat messages',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'entityAlerts', 
        label: 'Entity Movement Alerts', 
        description: 'Notifications when entities enter/exit areas',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'notificationSound', 
        label: 'Alert Sound', 
        description: 'Audio notification style',
        type: 'select', 
        value: 'tactical',
        options: [
          { value: 'tactical', label: 'Tactical Beep' },
          { value: 'subtle', label: 'Subtle Chime' },
          { value: 'urgent', label: 'Urgent Tone' },
          { value: 'none', label: 'Silent' },
        ]
      },
      { 
        id: 'soundVolume', 
        label: 'Notification Volume', 
        description: 'Audio alert volume level',
        type: 'range', 
        value: 75, 
        min: 0, 
        max: 100, 
        step: 5,
        unit: '%'
      },
    ],
    security: [
      { 
        id: 'requireAuth', 
        label: 'Require Authentication', 
        description: 'Force login before accessing tactical data',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'useCertificate', 
        label: 'Client Certificate Auth', 
        description: 'Use PKI certificate for server authentication',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'certificatePath', 
        label: 'Certificate Location', 
        description: 'Path to client certificate file (.p12/.pfx)',
        type: 'input', 
        value: '/certs/client.p12' 
      },
      { 
        id: 'encryptMessages', 
        label: 'Message Encryption', 
        description: 'Encrypt all tactical communications',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'sessionTimeout', 
        label: 'Session Timeout', 
        description: 'Auto-logout after period of inactivity',
        type: 'range', 
        value: 30, 
        min: 5, 
        max: 1440, 
        step: 5,
        unit: 'minutes'
      },
      { 
        id: 'lockOnIdle', 
        label: 'Auto-lock on Idle', 
        description: 'Lock screen when inactive for security',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'vaultIntegration', 
        label: 'HashiCorp Vault', 
        description: 'Manage certificates and secrets securely',
        type: 'link', 
        value: 'http://localhost:8200/ui/' 
      },
    ],
    data: [
      { 
        id: 'cacheEnabled', 
        label: 'Enable Local Cache', 
        description: 'Store frequently used data locally for faster access',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'cacheSize', 
        label: 'Cache Storage Limit', 
        description: 'Maximum disk space for cached data',
        type: 'range', 
        value: 500, 
        min: 100, 
        max: 5000, 
        step: 100,
        unit: 'MB'
      },
      { 
        id: 'offlineMode', 
        label: 'Offline Mode', 
        description: 'Allow app to function without network connectivity',
        type: 'toggle', 
        value: false 
      },
      { 
        id: 'syncInterval', 
        label: 'Data Sync Frequency', 
        description: 'How often to synchronize with the server',
        type: 'range', 
        value: 5, 
        min: 1, 
        max: 60, 
        step: 1,
        unit: 'minutes'
      },
      { 
        id: 'autoBackup', 
        label: 'Automatic Backup', 
        description: 'Regularly backup tactical data and settings',
        type: 'toggle', 
        value: true 
      },
      { 
        id: 'exportFormat', 
        label: 'Export Format', 
        description: 'Default format for data export',
        type: 'select', 
        value: 'kml',
        options: [
          { value: 'kml', label: 'KML (Keyhole Markup)' },
          { value: 'gpx', label: 'GPX (GPS Exchange)' },
          { value: 'json', label: 'JSON (JavaScript Object)' },
          { value: 'csv', label: 'CSV (Comma Separated)' },
        ]
      },
    ],
    advanced: [
      { 
        id: 'debugMode', 
        label: 'Debug Mode', 
        description: 'Enable detailed logging and diagnostics',
        type: 'toggle', 
        value: false,
        warning: 'May impact performance'
      },
      { 
        id: 'logLevel', 
        label: 'Logging Level', 
        description: 'Verbosity of system logs for troubleshooting',
        type: 'select', 
        value: 'info',
        options: [
          { value: 'error', label: 'Error Only' },
          { value: 'warn', label: 'Warning & Errors' },
          { value: 'info', label: 'Info, Warnings & Errors' },
          { value: 'debug', label: 'Debug (Verbose)' },
          { value: 'trace', label: 'Trace (Maximum)' },
        ]
      },
      { 
        id: 'performanceMonitor', 
        label: 'Performance Monitoring', 
        description: 'Track app performance metrics and resource usage',
        type: 'toggle', 
        value: false 
      },
      { 
        id: 'experimentalFeatures', 
        label: 'Experimental Features', 
        description: 'Enable beta features (may be unstable)',
        type: 'toggle', 
        value: false,
        warning: 'Use with caution in production'
      },
      { 
        id: 'developerMode', 
        label: 'Developer Console', 
        description: 'Show advanced developer tools and options',
        type: 'toggle', 
        value: false 
      },
      { 
        id: 'telemetry', 
        label: 'Anonymous Telemetry', 
        description: 'Help improve GoTAK by sharing usage statistics',
        type: 'toggle', 
        value: false 
      },
    ],
  });

  // Toggle card expansion
  const toggleCard = (sectionId: string) => {
    setExpandedCards(prev => {
      const newSet = new Set(prev);
      if (newSet.has(sectionId)) {
        newSet.delete(sectionId);
      } else {
        newSet.add(sectionId);
      }
      return newSet;
    });
  };

  // Handle setting changes with validation
  const handleSettingChange = useCallback((sectionId: string, settingId: string, newValue: any) => {
    setSettings(prev => ({
      ...prev,
      [sectionId]: prev[sectionId].map(setting =>
        setting.id === settingId ? { ...setting, value: newValue } : setting
      )
    }));
    setHasUnsavedChanges(true);
  }, []);

  // Save settings with status feedback
  const handleSave = async () => {
    setSaveStatus('saving');
    try {
      console.log('Saving settings:', settings);
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // In a real app, this would save to backend/localStorage
      setHasUnsavedChanges(false);
      setSaveStatus('saved');
      
      // Reset status after showing success
      setTimeout(() => setSaveStatus('idle'), 2000);
    } catch (error) {
      console.error('Failed to save settings:', error);
      setSaveStatus('error');
      setTimeout(() => setSaveStatus('idle'), 3000);
    }
  };

  // Reset settings with confirmation
  const handleReset = () => {
    const confirmed = window.confirm(
      'Are you sure you want to reset all settings to defaults?\n\nThis action cannot be undone.'
    );
    if (confirmed) {
      console.log('Resetting settings...');
      // Reset to defaults would go here
      setHasUnsavedChanges(false);
      setSaveStatus('idle');
    }
  };

  // Memoized filter functions to prevent unnecessary re-renders
  const filteredSections = useMemo(() => {
    if (!searchQuery.trim()) return SETTINGS_SECTIONS;
    
    const query = searchQuery.toLowerCase();
    return SETTINGS_SECTIONS.filter(section => {
      const sectionMatches = section.title.toLowerCase().includes(query) ||
                           section.description.toLowerCase().includes(query);
      
      const settingMatches = settings[section.id]?.some(setting =>
        setting.label.toLowerCase().includes(query) ||
        setting.description?.toLowerCase().includes(query)
      );
      
      return sectionMatches || settingMatches;
    });
  }, [searchQuery, settings]);

  const getFilteredSettings = useCallback((sectionId: string) => {
    if (!searchQuery.trim()) return settings[sectionId] || [];
    
    const query = searchQuery.toLowerCase();
    return (settings[sectionId] || []).filter(setting =>
      setting.label.toLowerCase().includes(query) ||
      setting.description?.toLowerCase().includes(query)
    );
  }, [searchQuery, settings]);

  // Render setting control based on type with improved styling
  const renderSettingControl = (section: string, setting: Setting) => {
    const commonProps = {
      disabled: setting.disabled,
      'aria-describedby': setting.description ? `${setting.id}-desc` : undefined,
    };

    switch (setting.type) {
      case 'toggle':
        return (
          <label className={`toggle-switch ${setting.disabled ? 'disabled' : ''}`}>
            <input
              id={setting.id}
              type="checkbox"
              checked={setting.value}
              onChange={(e) => handleSettingChange(section, setting.id, e.target.checked)}
              {...commonProps}
            />
            <span className="toggle-slider"></span>
          </label>
        );

      case 'select':
        return (
          <select
            id={setting.id}
            value={setting.value}
            onChange={(e) => handleSettingChange(section, setting.id, e.target.value)}
            className="setting-select"
            {...commonProps}
          >
            {setting.options?.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        );

      case 'input':
        return (
          <input
            id={setting.id}
            type="text"
            value={setting.value}
            onChange={(e) => handleSettingChange(section, setting.id, e.target.value)}
            className="setting-input"
            placeholder={setting.description}
            {...commonProps}
          />
        );

      case 'range':
        return (
          <div className="range-control">
            <input
              id={setting.id}
              type="range"
              min={setting.min}
              max={setting.max}
              step={setting.step}
              value={setting.value}
              onChange={(e) => handleSettingChange(section, setting.id, Number(e.target.value))}
              className="setting-range"
              {...commonProps}
            />
            <div className="range-value">
              <span className="value">{setting.value}</span>
              {setting.unit && <span className="unit">{setting.unit}</span>}
            </div>
          </div>
        );

      case 'color':
        return (
          <div className="color-control">
            <input
              id={setting.id}
              type="color"
              value={setting.value}
              onChange={(e) => handleSettingChange(section, setting.id, e.target.value)}
              className="setting-color"
              {...commonProps}
            />
            <div className="color-info">
              <span className="color-value">{setting.value}</span>
              <div 
                className="color-preview" 
                style={{ backgroundColor: setting.value }}
                aria-hidden="true"
              ></div>
            </div>
          </div>
        );

      case 'link':
        return (
          <a
            href={setting.value}
            target="_blank"
            rel="noopener noreferrer"
            className="setting-link"
          >
            <Icon name="external-link" size={16} />
            Open External
          </a>
        );

      default:
        return null;
    }
  };

  // filteredSections is now memoized above

  return (
    <div className="settings-page">
      {/* Fixed Header */}
      <header className="settings-header">
        <div className="header-content">
          <div className="header-title">
            <h1>
              <Icon name="settings" size={24} />
              Settings
            </h1>
            <p className="header-subtitle">System configuration and tactical preferences</p>
          </div>

          <div className="header-actions">
            <div className="search-container">
              <Icon name="search" size={16} className="search-icon" />
              <input
                type="text"
                placeholder="Search settings..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>

            <div className="action-buttons">
              {hasUnsavedChanges && (
                <div className="unsaved-indicator">
                  <div className="indicator-pulse"></div>
                  Unsaved changes
                </div>
              )}

              <button 
                className="btn btn-secondary" 
                onClick={handleReset}
                disabled={saveStatus === 'saving'}
              >
                <Icon name="refresh" size={16} />
                Reset All
              </button>
              
              <button 
                className="btn btn-primary" 
                onClick={handleSave}
                disabled={!hasUnsavedChanges || saveStatus === 'saving'}
              >
                {saveStatus === 'saving' && <Icon name="loader" size={16} className="spinning" />}
                {saveStatus === 'saved' && <Icon name="check" size={16} />}
                {saveStatus === 'error' && <Icon name="alert-triangle" size={16} />}
                {saveStatus === 'saving' ? 'Saving...' : 
                 saveStatus === 'saved' ? 'Saved!' : 
                 saveStatus === 'error' ? 'Error!' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="settings-main">
        {isMobile ? (
          // Mobile: Card-based accordion layout
          <div className="settings-cards">
            {filteredSections.map(section => {
              const sectionSettings = getFilteredSettings(section.id);
              const isExpanded = expandedCards.has(section.id);
              
              if (searchQuery && sectionSettings.length === 0) return null;
              
              return (
                <div key={section.id} className={`settings-card ${isExpanded ? 'expanded' : ''}`}>
                  <button 
                    className="card-header"
                    onClick={() => toggleCard(section.id)}
                    aria-expanded={isExpanded}
                  >
                    <div className="card-header-content">
                      <div className="section-icon">
                        <Icon name={section.icon} size={20} />
                      </div>
                      <div className="section-info">
                        <h3 className="section-title">{section.title}</h3>
                        <p className="section-description">{section.description}</p>
                      </div>
                      {section.badge && (
                        <span className="section-badge">{section.badge}</span>
                      )}
                    </div>
                    <Icon 
                      name="chevron-down" 
                      size={16} 
                      className={`expand-icon ${isExpanded ? 'expanded' : ''}`} 
                    />
                  </button>
                  
                  <div className={`card-content ${isExpanded ? 'expanded' : 'collapsed'}`}>
                    <div className="settings-grid">
                      {sectionSettings.map(setting => (
                        <div key={setting.id} className="setting-item">
                          <div className="setting-header">
                            <label className="setting-label" htmlFor={setting.id}>
                              {setting.label}
                            </label>
                            {setting.warning && (
                              <div className="setting-warning">
                                <Icon name="alert-triangle" size={12} />
                                {setting.warning}
                              </div>
                            )}
                          </div>
                          
                          {setting.description && (
                            <p className="setting-description" id={`${setting.id}-desc`}>
                              {setting.description}
                            </p>
                          )}
                          
                          <div className="setting-control">
                            {renderSettingControl(section.id, setting)}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          // Desktop: Sidebar + Panel layout
          <div className="settings-layout">
            <nav className="settings-sidebar">
              <div className="sidebar-content">
                <h2 className="sidebar-title">Categories</h2>
                <ul className="section-list">
                  {filteredSections.map(section => {
                    const sectionSettings = getFilteredSettings(section.id);
                    const hasResults = !searchQuery || sectionSettings.length > 0;
                    
                    if (!hasResults) return null;
                    
                    return (
                      <li key={section.id}>
                        <button
                          className={`section-button ${activeSection === section.id ? 'active' : ''}`}
                          onClick={() => setActiveSection(section.id)}
                        >
                          <div className="section-icon">
                            <Icon name={section.icon} size={18} />
                          </div>
                          <div className="section-content">
                            <span className="section-title">{section.title}</span>
                            <span className="section-description">{section.description}</span>
                          </div>
                          {section.badge && (
                            <span className="section-badge">{section.badge}</span>
                          )}
                        </button>
                      </li>
                    );
                  })}
                </ul>
              </div>
            </nav>

            <div className="settings-panel">
              <div className="panel-content">
                {(() => {
                  const currentSection = SETTINGS_SECTIONS.find(s => s.id === activeSection);
                  const filteredSettings = getFilteredSettings(activeSection);
                  
                  return (
                    <>
                      <div className="panel-header">
                        <div className="panel-title">
                          <Icon name={currentSection?.icon || 'settings'} size={24} />
                          <h2>{currentSection?.title}</h2>
                          {currentSection?.badge && (
                            <span className="section-badge">{currentSection.badge}</span>
                          )}
                        </div>
                        <p className="panel-description">{currentSection?.description}</p>
                      </div>

                      <div className="settings-grid">
                        {filteredSettings.length === 0 ? (
                          <div className="empty-state">
                            <Icon name="search" size={48} />
                            <h3>No settings found</h3>
                            <p>Try adjusting your search query or select a different category.</p>
                          </div>
                        ) : (
                          filteredSettings.map(setting => (
                            <div key={setting.id} className="setting-item">
                              <div className="setting-header">
                                <label className="setting-label" htmlFor={setting.id}>
                                  {setting.label}
                                </label>
                                {setting.warning && (
                                  <div className="setting-warning">
                                    <Icon name="alert-triangle" size={12} />
                                    {setting.warning}
                                  </div>
                                )}
                              </div>
                              
                              {setting.description && (
                                <p className="setting-description" id={`${setting.id}-desc`}>
                                  {setting.description}
                                </p>
                              )}
                              
                              <div className="setting-control">
                                {renderSettingControl(activeSection, setting)}
                              </div>
                            </div>
                          ))
                        )}
                      </div>
                    </>
                  );
                })()}
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default Settings;
