import React, { useState } from 'react';
import { Icon } from '../components/ui/Icon';
import './Settings.css';

const Settings: React.FC = () => {
  const [settings, setSettings] = useState({
    // General
    callsign: 'ALPHA-1',
    autoConnect: true,
    language: 'en',
    
    // Network
    serverUrl: 'wss://tak.server.mil:8089',
    protocol: 'wss',
    reconnectInterval: 30,
    
    // Display
    theme: 'dark',
    fontSize: 'medium',
    showGrid: true,
    
    // Map
    mapProvider: 'osm',
    showCoordinates: true,
    trackHistory: 24,
    
    // Security
    encryptComms: true,
    sessionTimeout: 30,
    requireAuth: true,
    
    // Notifications
    soundEnabled: true,
    desktopNotifications: true,
    alertVolume: 75,
    
    // Advanced
    debugMode: false,
    logLevel: 'info',
    experimentalFeatures: false
  });

  const handleChange = (key: string, value: any) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleSave = () => {
    console.log('Saving settings:', settings);
    alert('Settings saved successfully!');
  };

  const handleReset = () => {
    if (confirm('Reset all settings to defaults?')) {
      setSettings({
        // General
        callsign: 'ALPHA-1',
        autoConnect: true,
        language: 'en',
        
        // Network
        serverUrl: 'wss://tak.server.mil:8089',
        protocol: 'wss',
        reconnectInterval: 30,
        
        // Display
        theme: 'dark',
        fontSize: 'medium',
        showGrid: true,
        
        // Map
        mapProvider: 'osm',
        showCoordinates: true,
        trackHistory: 24,
        
        // Security
        encryptComms: true,
        sessionTimeout: 30,
        requireAuth: true,
        
        // Notifications
        soundEnabled: true,
        desktopNotifications: true,
        alertVolume: 75,
        
        // Advanced
        debugMode: false,
        logLevel: 'info',
        experimentalFeatures: false
      });
    }
  };

  return (
    <div className="settings-fullpage">
      {/* Header */}
      <header className="settings-header">
        <div className="header-title">
          <h1>System Settings</h1>
          <div className="settings-stats">
            <span className="stat active">{Object.values(settings).filter(v => v === true).length} Enabled</span>
            <span className="stat inactive">{Object.values(settings).filter(v => v === false).length} Disabled</span>
            <span className="stat total">{Object.keys(settings).length} Total</span>
          </div>
        </div>

        <div className="header-controls">
          <div className="settings-info">
            <span className="info-text">Configure your GoTAK application preferences</span>
          </div>

          <div className="action-buttons">
            <button className="btn-secondary" onClick={handleReset}>
              <Icon name="refresh" size={16} />
              Reset to Defaults
            </button>
            <button className="btn-primary" onClick={handleSave}>
              <Icon name="save" size={16} />
              Save Configuration
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="settings-content">
        <div className="settings-display">
        {/* General Settings */}
        <div className="settings-section">
          <h2><Icon name="settings" size={18} /> General</h2>
          
          <div className="setting-item">
            <label htmlFor="callsign">Callsign</label>
            <input
              id="callsign"
              type="text"
              value={settings.callsign}
              onChange={(e) => handleChange('callsign', e.target.value)}
              placeholder="Your tactical identifier"
            />
          </div>

          <div className="setting-item">
            <label htmlFor="autoConnect">Auto-connect on startup</label>
            <label className="toggle">
              <input
                id="autoConnect"
                type="checkbox"
                checked={settings.autoConnect}
                onChange={(e) => handleChange('autoConnect', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="language">Interface Language</label>
            <select
              id="language"
              value={settings.language}
              onChange={(e) => handleChange('language', e.target.value)}
            >
              <option value="en">English</option>
              <option value="es">Español</option>
              <option value="fr">Français</option>
              <option value="de">Deutsch</option>
            </select>
          </div>
        </div>

        {/* Network Settings */}
        <div className="settings-section">
          <h2><Icon name="signal" size={18} /> Network</h2>
          
          <div className="setting-item">
            <label htmlFor="serverUrl">TAK Server URL</label>
            <input
              id="serverUrl"
              type="text"
              value={settings.serverUrl}
              onChange={(e) => handleChange('serverUrl', e.target.value)}
              placeholder="wss://server.mil:8089"
            />
          </div>

          <div className="setting-item">
            <label htmlFor="protocol">Connection Protocol</label>
            <select
              id="protocol"
              value={settings.protocol}
              onChange={(e) => handleChange('protocol', e.target.value)}
            >
              <option value="wss">WebSocket Secure (WSS)</option>
              <option value="ws">WebSocket (WS)</option>
              <option value="tcp">TCP Direct</option>
            </select>
          </div>

          <div className="setting-item">
            <label htmlFor="reconnectInterval">Reconnect Interval (seconds)</label>
            <input
              id="reconnectInterval"
              type="number"
              min="5"
              max="300"
              value={settings.reconnectInterval}
              onChange={(e) => handleChange('reconnectInterval', parseInt(e.target.value))}
            />
          </div>
        </div>

        {/* Display Settings */}
        <div className="settings-section">
          <h2><Icon name="monitor" size={18} /> Display</h2>
          
          <div className="setting-item">
            <label htmlFor="theme">Theme</label>
            <select
              id="theme"
              value={settings.theme}
              onChange={(e) => handleChange('theme', e.target.value)}
            >
              <option value="dark">Dark (Tactical)</option>
              <option value="light">Light</option>
            </select>
          </div>

          <div className="setting-item">
            <label htmlFor="fontSize">Font Size</label>
            <select
              id="fontSize"
              value={settings.fontSize}
              onChange={(e) => handleChange('fontSize', e.target.value)}
            >
              <option value="small">Small</option>
              <option value="medium">Medium</option>
              <option value="large">Large</option>
            </select>
          </div>

          <div className="setting-item">
            <label htmlFor="showGrid">Show Grid Lines</label>
            <label className="toggle">
              <input
                id="showGrid"
                type="checkbox"
                checked={settings.showGrid}
                onChange={(e) => handleChange('showGrid', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>
        </div>

        {/* Map Settings */}
        <div className="settings-section">
          <h2><Icon name="map" size={18} /> Map</h2>
          
          <div className="setting-item">
            <label htmlFor="mapProvider">Map Provider</label>
            <select
              id="mapProvider"
              value={settings.mapProvider}
              onChange={(e) => handleChange('mapProvider', e.target.value)}
            >
              <option value="osm">OpenStreetMap</option>
              <option value="satellite">Satellite</option>
              <option value="terrain">Terrain</option>
              <option value="tactical">Tactical Grid</option>
            </select>
          </div>

          <div className="setting-item">
            <label htmlFor="showCoordinates">Show Coordinates</label>
            <label className="toggle">
              <input
                id="showCoordinates"
                type="checkbox"
                checked={settings.showCoordinates}
                onChange={(e) => handleChange('showCoordinates', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="trackHistory">Track History (hours)</label>
            <input
              id="trackHistory"
              type="number"
              min="1"
              max="168"
              value={settings.trackHistory}
              onChange={(e) => handleChange('trackHistory', parseInt(e.target.value))}
            />
          </div>
        </div>

        {/* Security Settings */}
        <div className="settings-section">
          <h2><Icon name="shield" size={18} /> Security</h2>
          
          <div className="setting-item">
            <label htmlFor="encryptComms">Encrypt Communications</label>
            <label className="toggle">
              <input
                id="encryptComms"
                type="checkbox"
                checked={settings.encryptComms}
                onChange={(e) => handleChange('encryptComms', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="requireAuth">Require Authentication</label>
            <label className="toggle">
              <input
                id="requireAuth"
                type="checkbox"
                checked={settings.requireAuth}
                onChange={(e) => handleChange('requireAuth', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="sessionTimeout">Session Timeout (minutes)</label>
            <input
              id="sessionTimeout"
              type="number"
              min="5"
              max="480"
              value={settings.sessionTimeout}
              onChange={(e) => handleChange('sessionTimeout', parseInt(e.target.value))}
            />
          </div>
        </div>

        {/* Notifications Settings */}
        <div className="settings-section">
          <h2><Icon name="bell" size={18} /> Notifications</h2>
          
          <div className="setting-item">
            <label htmlFor="soundEnabled">Sound Notifications</label>
            <label className="toggle">
              <input
                id="soundEnabled"
                type="checkbox"
                checked={settings.soundEnabled}
                onChange={(e) => handleChange('soundEnabled', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="desktopNotifications">Desktop Notifications</label>
            <label className="toggle">
              <input
                id="desktopNotifications"
                type="checkbox"
                checked={settings.desktopNotifications}
                onChange={(e) => handleChange('desktopNotifications', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="alertVolume">Alert Volume (%)</label>
            <input
              id="alertVolume"
              type="range"
              min="0"
              max="100"
              value={settings.alertVolume}
              onChange={(e) => handleChange('alertVolume', parseInt(e.target.value))}
            />
            <span className="volume-display">{settings.alertVolume}%</span>
          </div>
        </div>

        {/* Advanced Settings */}
        <div className="settings-section">
          <h2><Icon name="code" size={18} /> Advanced</h2>
          
          <div className="setting-item">
            <label htmlFor="debugMode">Debug Mode</label>
            <label className="toggle">
              <input
                id="debugMode"
                type="checkbox"
                checked={settings.debugMode}
                onChange={(e) => handleChange('debugMode', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>

          <div className="setting-item">
            <label htmlFor="logLevel">Logging Level</label>
            <select
              id="logLevel"
              value={settings.logLevel}
              onChange={(e) => handleChange('logLevel', e.target.value)}
            >
              <option value="error">Error Only</option>
              <option value="warn">Warnings & Errors</option>
              <option value="info">Info, Warnings & Errors</option>
              <option value="debug">Debug (Verbose)</option>
            </select>
          </div>

          <div className="setting-item">
            <label htmlFor="experimentalFeatures">Experimental Features</label>
            <label className="toggle">
              <input
                id="experimentalFeatures"
                type="checkbox"
                checked={settings.experimentalFeatures}
                onChange={(e) => handleChange('experimentalFeatures', e.target.checked)}
              />
              <span className="toggle-slider"></span>
            </label>
          </div>
        </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;