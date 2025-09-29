/**
 * Route Navigate Page - Real-time Navigation & Tracking
 * Operational navigation interface with turn-by-turn instructions
 */

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from '../utils/router';
import { Icon } from '../components/ui/Icon';
import './RouteNavigate.css';

interface Position {
  lat: number;
  lng: number;
  alt?: number;
  heading?: number;
  speed?: number;
}

interface NavigationInstruction {
  id: string;
  type: 'turn' | 'continue' | 'waypoint' | 'arrival';
  direction?: 'left' | 'right' | 'straight' | 'slight-left' | 'slight-right';
  instruction: string;
  distance: number;
  time: number;
  waypoint?: string;
}

interface NavigationState {
  currentPosition: Position;
  nextWaypoint: number;
  distanceToNext: number;
  timeToNext: number;
  totalDistanceRemaining: number;
  totalTimeRemaining: number;
  currentSpeed: number;
  averageSpeed: number;
  startTime: Date;
  eta: Date;
  instructions: NavigationInstruction[];
  currentInstruction: number;
}

const RouteNavigate: React.FC = () => {
  const router = useRouter();
  const id = router.params.id;
  const [isNavigating, setIsNavigating] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [navState, setNavState] = useState<NavigationState>({
    currentPosition: { lat: 38.8951, lng: -77.0364, alt: 120, heading: 45, speed: 30 },
    nextWaypoint: 1,
    distanceToNext: 2340,
    timeToNext: 8,
    totalDistanceRemaining: 12500,
    totalTimeRemaining: 45,
    currentSpeed: 30,
    averageSpeed: 28,
    startTime: new Date(),
    eta: new Date(Date.now() + 45 * 60 * 1000),
    instructions: [
      {
        id: 'inst-1',
        type: 'continue',
        direction: 'straight',
        instruction: 'Continue on Main Street',
        distance: 500,
        time: 2
      },
      {
        id: 'inst-2',
        type: 'turn',
        direction: 'right',
        instruction: 'Turn right onto Oak Avenue',
        distance: 800,
        time: 3
      },
      {
        id: 'inst-3',
        type: 'waypoint',
        instruction: 'Checkpoint Alpha - Radio check required',
        distance: 0,
        time: 0,
        waypoint: 'Checkpoint Alpha'
      },
      {
        id: 'inst-4',
        type: 'continue',
        direction: 'straight',
        instruction: 'Continue on Oak Avenue',
        distance: 1200,
        time: 4
      },
      {
        id: 'inst-5',
        type: 'turn',
        direction: 'left',
        instruction: 'Turn left onto Industrial Parkway',
        distance: 600,
        time: 2
      }
    ],
    currentInstruction: 0
  });

  const [showSettings, setShowSettings] = useState(false);
  const [navSettings, setNavSettings] = useState({
    voiceGuidance: true,
    autoReroute: true,
    nightMode: false,
    speedAlerts: true,
    threatAlerts: true
  });

  // Simulate position updates
  useEffect(() => {
    if (!isNavigating || isPaused) return;

    const interval = setInterval(() => {
      setNavState(prev => {
        const newDistance = Math.max(0, prev.distanceToNext - 50);
        const newTime = Math.max(0, prev.timeToNext - 0.05);
        
        return {
          ...prev,
          distanceToNext: newDistance,
          timeToNext: newTime,
          currentSpeed: 25 + Math.random() * 10,
          currentPosition: {
            ...prev.currentPosition,
            heading: prev.currentPosition.heading! + (Math.random() - 0.5) * 5
          }
        };
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [isNavigating, isPaused]);

  const formatDistance = (meters: number) => {
    if (meters < 1000) return `${Math.round(meters)}m`;
    return `${(meters / 1000).toFixed(1)}km`;
  };

  const formatTime = (minutes: number) => {
    if (minutes < 1) return `${Math.round(minutes * 60)}s`;
    if (minutes < 60) return `${Math.round(minutes)}min`;
    const hours = Math.floor(minutes / 60);
    const mins = Math.round(minutes % 60);
    return `${hours}h ${mins}m`;
  };

  const getDirectionIcon = (direction?: string): any => {
    switch (direction) {
      case 'left': return 'route';
      case 'right': return 'route';
      case 'straight': return 'route';
      case 'slight-left': return 'route';
      case 'slight-right': return 'route';
      default: return 'route';
    }
  };

  const handleStartNavigation = () => {
    setIsNavigating(true);
    setIsPaused(false);
  };

  const handlePauseNavigation = () => {
    setIsPaused(!isPaused);
  };

  const handleStopNavigation = () => {
    if (window.confirm('Are you sure you want to stop navigation?')) {
      setIsNavigating(false);
      router.navigate('/routes/view');
    }
  };

  return (
    <div className="route-navigate-container">
      {/* Navigation Header */}
      <header className="navigate-header">
        <div className="nav-controls">
          <button className="back-btn" onClick={() => router.navigate('/routes/view')}>
            <Icon name="x" size={20} />
          </button>
          
          <div className="route-name">
            Alpha Approach Navigation
          </div>

          <div className="nav-actions">
            {!isNavigating ? (
              <button className="btn-primary" onClick={handleStartNavigation}>
                <Icon name="send" size={16} />
                Start Navigation
              </button>
            ) : (
              <>
                <button 
                  className={`btn-control ${isPaused ? 'paused' : ''}`} 
                  onClick={handlePauseNavigation}
                >
                  <Icon name={isPaused ? 'send' : 'alert-circle'} size={16} />
                  {isPaused ? 'Resume' : 'Pause'}
                </button>
                <button className="btn-danger" onClick={handleStopNavigation}>
                  <Icon name="x" size={16} />
                  Stop
                </button>
              </>
            )}
            <button className="btn-settings" onClick={() => setShowSettings(!showSettings)}>
              <Icon name="settings" size={16} />
            </button>
          </div>
        </div>
      </header>

      {/* Main Navigation Display */}
      <div className="navigate-content">
        {/* Map Area */}
        <div className="map-container">
          <div className="map-placeholder">
            <Icon name="map" size={48} />
            <h3>Live Navigation Map</h3>
            <p>Real-time position tracking and route overlay</p>
            
            {/* Current Position Overlay */}
            <div className="position-overlay">
              <div className="position-info">
                <span className="label">Position:</span>
                <span className="value">
                  {navState.currentPosition.lat.toFixed(4)}, {navState.currentPosition.lng.toFixed(4)}
                </span>
              </div>
              <div className="position-info">
                <span className="label">Altitude:</span>
                <span className="value">{navState.currentPosition.alt}m</span>
              </div>
              <div className="position-info">
                <span className="label">Heading:</span>
                <span className="value">{Math.round(navState.currentPosition.heading || 0)}°</span>
              </div>
            </div>
          </div>

          {/* Navigation Stats Overlay */}
          <div className="nav-stats-overlay">
            <div className="stat-card primary">
              <Icon name="pin" size={20} />
              <div className="stat-info">
                <span className="value">{formatDistance(navState.distanceToNext)}</span>
                <span className="label">To Next</span>
              </div>
            </div>
            <div className="stat-card">
              <Icon name="speed" size={20} />
              <div className="stat-info">
                <span className="value">{Math.round(navState.currentSpeed)} km/h</span>
                <span className="label">Speed</span>
              </div>
            </div>
            <div className="stat-card">
              <Icon name="target" size={20} />
              <div className="stat-info">
                <span className="value">{formatTime(navState.totalTimeRemaining)}</span>
                <span className="label">ETA</span>
              </div>
            </div>
          </div>
        </div>

        {/* Instructions Panel */}
        <div className="instructions-panel">
          {/* Current Instruction */}
          <div className="current-instruction">
            <div className="instruction-icon">
              <Icon name={getDirectionIcon(navState.instructions[navState.currentInstruction]?.direction)} size={32} />
            </div>
            <div className="instruction-details">
              <h2>{navState.instructions[navState.currentInstruction]?.instruction}</h2>
              <div className="instruction-meta">
                <span>{formatDistance(navState.instructions[navState.currentInstruction]?.distance || 0)}</span>
                <span className="separator">•</span>
                <span>{formatTime(navState.instructions[navState.currentInstruction]?.time || 0)}</span>
              </div>
            </div>
          </div>

          {/* Upcoming Instructions */}
          <div className="upcoming-instructions">
            <h3>Upcoming Maneuvers</h3>
            <div className="instructions-list">
              {navState.instructions.slice(navState.currentInstruction + 1, navState.currentInstruction + 4).map((instruction, idx) => (
                <div key={instruction.id} className="instruction-item">
                  <div className="instruction-number">{idx + 1}</div>
                  <Icon name={getDirectionIcon(instruction.direction)} size={16} />
                  <div className="instruction-text">
                    <span className="text">{instruction.instruction}</span>
                    <span className="distance">{formatDistance(instruction.distance)}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Waypoint Info */}
          <div className="waypoint-info-panel">
            <h3>Next Waypoint</h3>
            <div className="waypoint-details">
              <div className="waypoint-header">
                <Icon name="check-circle" size={20} />
                <span className="waypoint-name">Checkpoint Alpha</span>
              </div>
              <div className="waypoint-stats">
                <div className="stat">
                  <span className="label">Distance:</span>
                  <span className="value">{formatDistance(navState.distanceToNext)}</span>
                </div>
                <div className="stat">
                  <span className="label">ETA:</span>
                  <span className="value">{navState.eta.toLocaleTimeString()}</span>
                </div>
                <div className="stat">
                  <span className="label">Type:</span>
                  <span className="value">Checkpoint</span>
                </div>
              </div>
              <div className="waypoint-notes">
                <Icon name="info" size={14} />
                Radio check required. Alternative route available if compromised.
              </div>
            </div>
          </div>

          {/* Route Progress */}
          <div className="route-progress">
            <h3>Route Progress</h3>
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: '35%' }} />
              <div className="progress-markers">
                <div className="marker completed" style={{ left: '0%' }} title="Start" />
                <div className="marker completed" style={{ left: '20%' }} title="Waypoint 1" />
                <div className="marker active" style={{ left: '35%' }} title="Current Position" />
                <div className="marker" style={{ left: '60%' }} title="Checkpoint Alpha" />
                <div className="marker" style={{ left: '80%' }} title="Rally Point" />
                <div className="marker" style={{ left: '100%' }} title="Objective" />
              </div>
            </div>
            <div className="progress-stats">
              <span>Completed: {formatDistance(12500 - navState.totalDistanceRemaining)}</span>
              <span>Remaining: {formatDistance(navState.totalDistanceRemaining)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <div className="settings-panel">
          <div className="settings-header">
            <h3>Navigation Settings</h3>
            <button onClick={() => setShowSettings(false)}>
              <Icon name="x" size={20} />
            </button>
          </div>
          <div className="settings-content">
            <label className="setting-item">
              <input 
                type="checkbox" 
                checked={navSettings.voiceGuidance}
                onChange={(e) => setNavSettings({...navSettings, voiceGuidance: e.target.checked})}
              />
              <span>Voice Guidance</span>
            </label>
            <label className="setting-item">
              <input 
                type="checkbox" 
                checked={navSettings.autoReroute}
                onChange={(e) => setNavSettings({...navSettings, autoReroute: e.target.checked})}
              />
              <span>Auto Reroute</span>
            </label>
            <label className="setting-item">
              <input 
                type="checkbox" 
                checked={navSettings.nightMode}
                onChange={(e) => setNavSettings({...navSettings, nightMode: e.target.checked})}
              />
              <span>Night Mode</span>
            </label>
            <label className="setting-item">
              <input 
                type="checkbox" 
                checked={navSettings.speedAlerts}
                onChange={(e) => setNavSettings({...navSettings, speedAlerts: e.target.checked})}
              />
              <span>Speed Alerts</span>
            </label>
            <label className="setting-item">
              <input 
                type="checkbox" 
                checked={navSettings.threatAlerts}
                onChange={(e) => setNavSettings({...navSettings, threatAlerts: e.target.checked})}
              />
              <span>Threat Alerts</span>
            </label>
          </div>
        </div>
      )}

      {/* Alert/Warning Banner */}
      {isNavigating && !isPaused && (
        <div className="alert-banner warning">
          <Icon name="alert-triangle" size={16} />
          <span>Approaching high-threat area in 500m - Stay alert</span>
        </div>
      )}
    </div>
  );
};

export default RouteNavigate;