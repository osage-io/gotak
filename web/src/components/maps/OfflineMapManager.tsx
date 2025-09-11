import React, { useState, useEffect } from 'react';
import { formatBytes, formatNumber } from '../../utils/formatting';
import './OfflineMapManager.css';

interface TileSource {
  id: string;
  name: string;
  type: 'osm' | 'satellite' | 'terrain' | 'custom';
  url: string;
  maxZoom: number;
  attribution?: string;
}

interface DownloadJob {
  id: string;
  name: string;
  source: string;
  bounds: {
    north: number;
    south: number;
    east: number;
    west: number;
  };
  minZoom: number;
  maxZoom: number;
  status: 'pending' | 'downloading' | 'completed' | 'failed' | 'paused';
  progress: number;
  totalTiles: number;
  downloadedTiles: number;
  estimatedSize: number;
  actualSize?: number;
  createdAt: Date;
  completedAt?: Date;
  error?: string;
}

interface OfflineMapManagerProps {
  onClose?: () => void;
}

const DEFAULT_TILE_SOURCES: TileSource[] = [
  {
    id: 'osm',
    name: 'OpenStreetMap',
    type: 'osm',
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    maxZoom: 19,
    attribution: '© OpenStreetMap contributors'
  },
  {
    id: 'osm-topo',
    name: 'OpenTopoMap',
    type: 'terrain',
    url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
    maxZoom: 17,
    attribution: '© OpenTopoMap'
  },
  {
    id: 'satellite',
    name: 'Satellite Imagery',
    type: 'satellite',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    maxZoom: 19,
    attribution: '© Esri'
  }
];

export const OfflineMapManager: React.FC<OfflineMapManagerProps> = ({ onClose }) => {
  const [downloadJobs, setDownloadJobs] = useState<DownloadJob[]>([]);
  const [tileSources, setTileSources] = useState<TileSource[]>(DEFAULT_TILE_SOURCES);
  const [selectedSource, setSelectedSource] = useState<string>('osm');
  const [downloadName, setDownloadName] = useState<string>('');
  const [bounds, setBounds] = useState({
    north: 40.0,
    south: 39.0,
    east: -76.0,
    west: -77.0
  });
  const [minZoom, setMinZoom] = useState<number>(8);
  const [maxZoom, setMaxZoom] = useState<number>(15);
  const [storageInfo, setStorageInfo] = useState({
    used: 0,
    available: 0,
    total: 0
  });
  const [activeTab, setActiveTab] = useState<'download' | 'manage' | 'sources'>('download');
  const [isCreatingDownload, setIsCreatingDownload] = useState<boolean>(false);

  useEffect(() => {
    // Load existing download jobs and storage info
    loadDownloadJobs();
    loadStorageInfo();
    const interval = setInterval(updateJobProgress, 1000);
    return () => clearInterval(interval);
  }, []);

  const loadDownloadJobs = async () => {
    try {
      // In a real implementation, this would fetch from the backend
      const stored = localStorage.getItem('offline-map-jobs');
      if (stored) {
        const jobs = JSON.parse(stored).map((job: any) => ({
          ...job,
          createdAt: new Date(job.createdAt),
          completedAt: job.completedAt ? new Date(job.completedAt) : undefined
        }));
        setDownloadJobs(jobs);
      }
    } catch (error) {
      console.error('Failed to load download jobs:', error);
    }
  };

  const loadStorageInfo = () => {
    // Simulate storage info - in real implementation would query IndexedDB/filesystem
    const used = downloadJobs.reduce((total, job) => total + (job.actualSize || 0), 0);
    setStorageInfo({
      used,
      available: 5 * 1024 * 1024 * 1024 - used, // 5GB total
      total: 5 * 1024 * 1024 * 1024
    });
  };

  const updateJobProgress = () => {
    setDownloadJobs(prevJobs => 
      prevJobs.map(job => {
        if (job.status === 'downloading') {
          // Simulate progress updates
          const newProgress = Math.min(job.progress + Math.random() * 5, 100);
          const newDownloadedTiles = Math.floor((newProgress / 100) * job.totalTiles);
          
          if (newProgress >= 100) {
            return {
              ...job,
              status: 'completed' as const,
              progress: 100,
              downloadedTiles: job.totalTiles,
              actualSize: job.estimatedSize * (0.8 + Math.random() * 0.4),
              completedAt: new Date()
            };
          }
          
          return {
            ...job,
            progress: newProgress,
            downloadedTiles: newDownloadedTiles
          };
        }
        return job;
      })
    );
  };

  const calculateTileCount = (bounds: any, minZoom: number, maxZoom: number): number => {
    let totalTiles = 0;
    for (let z = minZoom; z <= maxZoom; z++) {
      const tilesPerSide = Math.pow(2, z);
      const latRange = Math.abs(bounds.north - bounds.south);
      const lngRange = Math.abs(bounds.east - bounds.west);
      const tilesInLat = Math.ceil((latRange / 180) * tilesPerSide);
      const tilesInLng = Math.ceil((lngRange / 360) * tilesPerSide);
      totalTiles += tilesInLat * tilesInLng;
    }
    return totalTiles;
  };

  const createDownloadJob = () => {
    if (!downloadName.trim()) {
      alert('Please enter a name for the download job');
      return;
    }

    const totalTiles = calculateTileCount(bounds, minZoom, maxZoom);
    const estimatedSize = totalTiles * 15000; // Estimate ~15KB per tile

    const newJob: DownloadJob = {
      id: Date.now().toString(),
      name: downloadName.trim(),
      source: selectedSource,
      bounds,
      minZoom,
      maxZoom,
      status: 'pending',
      progress: 0,
      totalTiles,
      downloadedTiles: 0,
      estimatedSize,
      createdAt: new Date()
    };

    setDownloadJobs(prev => [...prev, newJob]);
    setDownloadName('');
    setIsCreatingDownload(false);

    // Save to localStorage (in real implementation, would save to backend)
    localStorage.setItem('offline-map-jobs', JSON.stringify([...downloadJobs, newJob]));
  };

  const startDownload = (jobId: string) => {
    setDownloadJobs(prev =>
      prev.map(job =>
        job.id === jobId ? { ...job, status: 'downloading' } : job
      )
    );
  };

  const pauseDownload = (jobId: string) => {
    setDownloadJobs(prev =>
      prev.map(job =>
        job.id === jobId ? { ...job, status: 'paused' } : job
      )
    );
  };

  const deleteDownloadJob = (jobId: string) => {
    if (confirm('Are you sure you want to delete this download job? This will remove all downloaded tiles.')) {
      setDownloadJobs(prev => prev.filter(job => job.id !== jobId));
      // Update localStorage
      const updated = downloadJobs.filter(job => job.id !== jobId);
      localStorage.setItem('offline-map-jobs', JSON.stringify(updated));
    }
  };

  const clearAllCompleted = () => {
    if (confirm('Are you sure you want to clear all completed downloads?')) {
      setDownloadJobs(prev => prev.filter(job => job.status !== 'completed'));
    }
  };

  const getStatusIcon = (status: DownloadJob['status']) => {
    switch (status) {
      case 'pending': return '⏳';
      case 'downloading': return '⬇️';
      case 'completed': return '✅';
      case 'failed': return '❌';
      case 'paused': return '⏸️';
      default: return '❓';
    }
  };

  const getStatusColor = (status: DownloadJob['status']) => {
    switch (status) {
      case 'pending': return 'var(--warning-color)';
      case 'downloading': return 'var(--primary-color)';
      case 'completed': return 'var(--success-color)';
      case 'failed': return 'var(--danger-color)';
      case 'paused': return 'var(--text-secondary)';
      default: return 'var(--text-secondary)';
    }
  };

  return (
    <div className="offline-map-manager">
      <div className="manager-header">
        <h2>Offline Map Manager</h2>
        {onClose && (
          <button className="close-button" onClick={onClose} aria-label="Close manager">
            ×
          </button>
        )}
      </div>

      <div className="storage-info">
        <div className="storage-bar">
          <div 
            className="storage-used" 
            style={{ width: `${(storageInfo.used / storageInfo.total) * 100}%` }}
          />
        </div>
        <div className="storage-details">
          <span>Storage: {formatBytes(storageInfo.used)} used of {formatBytes(storageInfo.total)}</span>
          <span>Available: {formatBytes(storageInfo.available)}</span>
        </div>
      </div>

      <div className="manager-tabs">
        <button
          className={`tab-button ${activeTab === 'download' ? 'active' : ''}`}
          onClick={() => setActiveTab('download')}
        >
          📥 New Download
        </button>
        <button
          className={`tab-button ${activeTab === 'manage' ? 'active' : ''}`}
          onClick={() => setActiveTab('manage')}
        >
          📋 Manage ({downloadJobs.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'sources' ? 'active' : ''}`}
          onClick={() => setActiveTab('sources')}
        >
          🌍 Sources
        </button>
      </div>

      <div className="tab-content">
        {activeTab === 'download' && (
          <div className="download-tab">
            <div className="form-group">
              <label htmlFor="download-name">Download Name:</label>
              <input
                id="download-name"
                type="text"
                value={downloadName}
                onChange={(e) => setDownloadName(e.target.value)}
                placeholder="e.g., Washington DC Area"
                maxLength={100}
              />
            </div>

            <div className="form-group">
              <label htmlFor="tile-source">Map Source:</label>
              <select
                id="tile-source"
                value={selectedSource}
                onChange={(e) => setSelectedSource(e.target.value)}
              >
                {tileSources.map(source => (
                  <option key={source.id} value={source.id}>
                    {source.name} (Max Zoom: {source.maxZoom})
                  </option>
                ))}
              </select>
            </div>

            <div className="bounds-group">
              <label>Area Bounds (Decimal Degrees):</label>
              <div className="bounds-inputs">
                <div className="bound-input">
                  <label htmlFor="north">North:</label>
                  <input
                    id="north"
                    type="number"
                    step="0.001"
                    value={bounds.north}
                    onChange={(e) => setBounds(prev => ({ ...prev, north: parseFloat(e.target.value) }))}
                  />
                </div>
                <div className="bound-input">
                  <label htmlFor="south">South:</label>
                  <input
                    id="south"
                    type="number"
                    step="0.001"
                    value={bounds.south}
                    onChange={(e) => setBounds(prev => ({ ...prev, south: parseFloat(e.target.value) }))}
                  />
                </div>
                <div className="bound-input">
                  <label htmlFor="west">West:</label>
                  <input
                    id="west"
                    type="number"
                    step="0.001"
                    value={bounds.west}
                    onChange={(e) => setBounds(prev => ({ ...prev, west: parseFloat(e.target.value) }))}
                  />
                </div>
                <div className="bound-input">
                  <label htmlFor="east">East:</label>
                  <input
                    id="east"
                    type="number"
                    step="0.001"
                    value={bounds.east}
                    onChange={(e) => setBounds(prev => ({ ...prev, east: parseFloat(e.target.value) }))}
                  />
                </div>
              </div>
            </div>

            <div className="zoom-group">
              <div className="zoom-input">
                <label htmlFor="min-zoom">Min Zoom:</label>
                <input
                  id="min-zoom"
                  type="number"
                  min="0"
                  max="19"
                  value={minZoom}
                  onChange={(e) => setMinZoom(parseInt(e.target.value))}
                />
              </div>
              <div className="zoom-input">
                <label htmlFor="max-zoom">Max Zoom:</label>
                <input
                  id="max-zoom"
                  type="number"
                  min="0"
                  max="19"
                  value={maxZoom}
                  onChange={(e) => setMaxZoom(parseInt(e.target.value))}
                />
              </div>
            </div>

            <div className="download-estimate">
              <div className="estimate-info">
                <span>Estimated Tiles: {formatNumber(calculateTileCount(bounds, minZoom, maxZoom))}</span>
                <span>Estimated Size: {formatBytes(calculateTileCount(bounds, minZoom, maxZoom) * 15000)}</span>
              </div>
            </div>

            <button
              className="create-download-button"
              onClick={createDownloadJob}
              disabled={!downloadName.trim()}
            >
              Create Download Job
            </button>
          </div>
        )}

        {activeTab === 'manage' && (
          <div className="manage-tab">
            <div className="manage-actions">
              <button
                className="clear-completed-button"
                onClick={clearAllCompleted}
                disabled={!downloadJobs.some(job => job.status === 'completed')}
              >
                Clear Completed
              </button>
            </div>

            <div className="download-jobs">
              {downloadJobs.length === 0 ? (
                <div className="no-jobs">
                  <p>No download jobs yet. Create one in the "New Download" tab.</p>
                </div>
              ) : (
                downloadJobs.map(job => (
                  <div key={job.id} className="download-job">
                    <div className="job-header">
                      <div className="job-title">
                        <span className="job-icon">{getStatusIcon(job.status)}</span>
                        <span className="job-name">{job.name}</span>
                        <span 
                          className="job-status" 
                          style={{ color: getStatusColor(job.status) }}
                        >
                          {job.status.toUpperCase()}
                        </span>
                      </div>
                      <div className="job-actions">
                        {job.status === 'pending' && (
                          <button
                            className="start-button"
                            onClick={() => startDownload(job.id)}
                            title="Start download"
                          >
                            ▶️
                          </button>
                        )}
                        {job.status === 'downloading' && (
                          <button
                            className="pause-button"
                            onClick={() => pauseDownload(job.id)}
                            title="Pause download"
                          >
                            ⏸️
                          </button>
                        )}
                        {job.status === 'paused' && (
                          <button
                            className="start-button"
                            onClick={() => startDownload(job.id)}
                            title="Resume download"
                          >
                            ▶️
                          </button>
                        )}
                        <button
                          className="delete-button"
                          onClick={() => deleteDownloadJob(job.id)}
                          title="Delete job"
                        >
                          🗑️
                        </button>
                      </div>
                    </div>

                    <div className="job-details">
                      <div className="detail-row">
                        <span>Source: {tileSources.find(s => s.id === job.source)?.name || job.source}</span>
                        <span>Zoom: {job.minZoom}-{job.maxZoom}</span>
                      </div>
                      <div className="detail-row">
                        <span>Tiles: {formatNumber(job.downloadedTiles)} / {formatNumber(job.totalTiles)}</span>
                        <span>Size: {job.actualSize ? formatBytes(job.actualSize) : formatBytes(job.estimatedSize)}</span>
                      </div>
                      <div className="detail-row">
                        <span>Created: {job.createdAt.toLocaleDateString()}</span>
                        {job.completedAt && (
                          <span>Completed: {job.completedAt.toLocaleDateString()}</span>
                        )}
                      </div>
                    </div>

                    {(job.status === 'downloading' || job.status === 'paused') && (
                      <div className="progress-bar">
                        <div 
                          className="progress-fill" 
                          style={{ width: `${job.progress}%` }}
                        />
                        <span className="progress-text">{job.progress.toFixed(1)}%</span>
                      </div>
                    )}

                    {job.error && (
                      <div className="job-error">
                        <span>Error: {job.error}</span>
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>
        )}

        {activeTab === 'sources' && (
          <div className="sources-tab">
            <div className="sources-list">
              {tileSources.map(source => (
                <div key={source.id} className="source-item">
                  <div className="source-info">
                    <h4>{source.name}</h4>
                    <p>Type: {source.type} | Max Zoom: {source.maxZoom}</p>
                    <p className="source-url">{source.url}</p>
                    {source.attribution && (
                      <p className="source-attribution">{source.attribution}</p>
                    )}
                  </div>
                  <div className="source-preview">
                    <div className={`type-badge ${source.type}`}>
                      {source.type.toUpperCase()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
