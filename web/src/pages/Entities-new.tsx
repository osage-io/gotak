/**
 * Entities Page - Modern Redesign
 * Comprehensive entity management with advanced visualization
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Icon } from '../components/ui/Icon';
import './Entities-new.css';

// --- Vault PKI device certificates (DEMO ONLY) ----------------------------
// Issue / revoke per-device mTLS client certs from Vault's PKI engine. As with
// the rest of the demo we talk to Vault directly with the dev root token — do
// NOT ship a root token in frontend code outside a throwaway demo.
const VAULT_ADDR = 'http://127.0.0.1:8200';
const VAULT_TOKEN = 'root';
const PKI_ROLE = 'gotak-device';
const CERT_TTL = '168h'; // 7 days
const CERT_TTL_LABEL = '7 days';
const CERTS_STORAGE_KEY = 'gotak_device_certs';

interface DeviceCert {
  serial: string;
  commonName: string;
  certificate: string;   // PEM
  privateKey: string;    // PEM
  issuingCa: string;     // PEM
  expiration: number;    // unix seconds
  issuedAt: number;      // unix seconds (client clock)
  fingerprint: string;   // SHA-256, colon-separated hex
  revoked: boolean;
}

// Derive a stable cert common name from a device callsign.
const deviceCommonName = (callsign: string) =>
  `${callsign.replace(/[^A-Za-z0-9.-]/g, '-')}.device.gotak.local`;

async function vaultPki(path: string, body?: unknown): Promise<any> {
  const res = await fetch(`${VAULT_ADDR}/v1/${path}`, {
    method: body === undefined ? 'GET' : 'POST',
    headers: { 'X-Vault-Token': VAULT_TOKEN, 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = json?.errors?.join?.(', ') || `HTTP ${res.status}`;
    throw new Error(`Vault PKI: ${msg}`);
  }
  return json;
}

// SHA-256 fingerprint of the certificate DER, formatted like openssl.
async function certFingerprint(pem: string): Promise<string> {
  const b64 = pem.replace(/-----[^-]+-----/g, '').replace(/\s+/g, '');
  const der = Uint8Array.from(atob(b64), c => c.charCodeAt(0));
  const hash = await crypto.subtle.digest('SHA-256', der);
  return Array.from(new Uint8Array(hash))
    .map(b => b.toString(16).padStart(2, '0').toUpperCase())
    .join(':');
}

async function issueDeviceCert(callsign: string): Promise<DeviceCert> {
  const commonName = deviceCommonName(callsign);
  const { data } = await vaultPki(`pki/issue/${PKI_ROLE}`, { common_name: commonName, ttl: CERT_TTL });
  return {
    serial: data.serial_number,
    commonName,
    certificate: data.certificate,
    privateKey: data.private_key,
    issuingCa: data.issuing_ca,
    expiration: data.expiration,
    issuedAt: Math.floor(Date.now() / 1000),
    fingerprint: await certFingerprint(data.certificate),
    revoked: false,
  };
}

interface Entity {
  id: string;
  callsign: string;
  type: 'friendly' | 'hostile' | 'neutral' | 'unknown' | 'drone' | 'sensor' | 'camera' | 'vehicle' | 'equipment';
  subType?: 'uav' | 'ugv' | 'thermal' | 'motion' | 'acoustic' | 'radiation' | 'cbrn' | 'weather' | 'surveillance' | 'tactical';
  status: 'active' | 'inactive' | 'stale' | 'lost' | 'standby' | 'maintenance';
  team: string;
  role: string;
  position: {
    lat: number;
    lng: number;
    alt?: number;
    speed?: number;
    heading?: number;
  };
  lastUpdate: string;
  battery?: number;
  signal?: number;
  temperature?: number;
  fuelLevel?: number;
  operatingTime?: number;
  sensorData?: {
    type: string;
    value: any;
    unit: string;
    timestamp: string;
  }[];
  equipment?: string[];
  capabilities?: string[];
  notes?: string;
}

type EntityView = 'grid' | 'list' | 'tactical';
type EntityFilter = 'all' | 'personnel' | 'drones' | 'sensors' | 'vehicles' | 'friendly' | 'hostile' | 'neutral' | 'unknown';

const EntitiesNew: React.FC = () => {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedEntity, setSelectedEntity] = useState<string | null>(null);
  // Per-entity device certificates (keyed by entity id), issued from Vault PKI.
  const [certs, setCerts] = useState<Record<string, DeviceCert>>({});
  const [certBusy, setCertBusy] = useState(false);
  const [certError, setCertError] = useState<string | null>(null);
  // Entity for which the "Issue Certificate" security dialog is open.
  const [issueDialogEntity, setIssueDialogEntity] = useState<Entity | null>(null);

  // Load persisted certs on mount, then reconcile each against Vault so the
  // indicator reflects reality (revoked/expired) — survives page refresh.
  useEffect(() => {
    let saved: Record<string, DeviceCert> = {};
    try {
      saved = JSON.parse(localStorage.getItem(CERTS_STORAGE_KEY) || '{}');
    } catch { saved = {}; }
    if (Object.keys(saved).length === 0) return;
    setCerts(saved);

    // Best-effort: ask Vault about each serial; mark revoked if Vault says so.
    (async () => {
      const updated: Record<string, DeviceCert> = { ...saved };
      let changed = false;
      await Promise.all(Object.entries(saved).map(async ([id, cert]) => {
        try {
          const { data } = await vaultPki(`pki/cert/${cert.serial}`);
          const revoked = !!data.revocation_time && data.revocation_time !== 0;
          if (revoked !== cert.revoked) { updated[id] = { ...cert, revoked }; changed = true; }
        } catch { /* leave as-is if Vault is unreachable */ }
      }));
      if (changed) setCerts(updated);
    })();
  }, []);

  // Persist certs whenever they change.
  useEffect(() => {
    try { localStorage.setItem(CERTS_STORAGE_KEY, JSON.stringify(certs)); } catch { /* ignore */ }
  }, [certs]);
  const [view, setView] = useState<EntityView>('grid');
  const [filter, setFilter] = useState<EntityFilter>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showOffline, setShowOffline] = useState(true);

  // Initialize with mock data
  useEffect(() => {
    const mockEntities: Entity[] = [
      {
        id: '1',
        callsign: 'ALPHA-1',
        type: 'friendly',
        status: 'active',
        team: 'Blue Force',
        role: 'Squad Leader',
        position: {
          lat: 38.8951,
          lng: -77.0364,
          alt: 125,
          speed: 0,
          heading: 45
        },
        lastUpdate: new Date().toISOString(),
        battery: 85,
        signal: 92,
        equipment: ['Radio', 'NVG', 'GPS', 'Medical Kit'],
        capabilities: ['Command', 'Medical', 'Navigation'],
      },
      {
        id: '2',
        callsign: 'BRAVO-2',
        type: 'friendly',
        status: 'active',
        team: 'Blue Force',
        role: 'Rifleman',
        position: {
          lat: 38.8977,
          lng: -77.0365,
          alt: 130,
          speed: 5.2,
          heading: 90
        },
        lastUpdate: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
        battery: 67,
        signal: 88,
        equipment: ['Radio', 'GPS'],
        capabilities: ['Reconnaissance'],
      },
      {
        id: '3',
        callsign: 'CHARLIE-3',
        type: 'friendly',
        status: 'stale',
        team: 'Blue Force',
        role: 'Medic',
        position: {
          lat: 38.8923,
          lng: -77.0389,
          alt: 110,
          speed: 0,
          heading: 180
        },
        lastUpdate: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        battery: 45,
        signal: 65,
        equipment: ['Radio', 'Medical Supplies', 'GPS'],
        capabilities: ['Medical', 'CASEVAC'],
      },
      {
        id: '4',
        callsign: 'HOSTILE-X1',
        type: 'hostile',
        status: 'active',
        team: 'Red Force',
        role: 'Unknown',
        position: {
          lat: 38.9012,
          lng: -77.0423,
          alt: 150,
          speed: 8.3,
          heading: 270
        },
        lastUpdate: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        equipment: ['Unknown'],
        capabilities: ['Armed'],
      },
      {
        id: '5',
        callsign: 'NEUTRAL-N1',
        type: 'neutral',
        status: 'active',
        team: 'Civilian',
        role: 'Observer',
        position: {
          lat: 38.8889,
          lng: -77.0298,
          alt: 100,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 95,
        signal: 100,
        equipment: ['Radio', 'Camera'],
        capabilities: ['Observation'],
      },
      {
        id: '6',
        callsign: 'DELTA-4',
        type: 'friendly',
        status: 'lost',
        team: 'Blue Force',
        role: 'Scout',
        position: {
          lat: 38.9072,
          lng: -77.0369,
          alt: 140,
          speed: 0,
          heading: 315
        },
        lastUpdate: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
        battery: 12,
        signal: 0,
        equipment: ['Radio', 'NVG', 'GPS'],
        capabilities: ['Reconnaissance', 'Sniper'],
        notes: 'Last contact 45 minutes ago. Possible equipment failure.',
      },
      // Drone entities
      {
        id: '7',
        callsign: 'EAGLE-EYE-1',
        type: 'drone',
        subType: 'uav',
        status: 'active',
        team: 'Air Assets',
        role: 'ISR Platform',
        position: {
          lat: 38.9021,
          lng: -77.0367,
          alt: 450,
          speed: 25.5,
          heading: 135
        },
        lastUpdate: new Date().toISOString(),
        battery: 78,
        signal: 95,
        operatingTime: 2.5,
        equipment: ['EO/IR Camera', '4K Video', 'Thermal Imaging'],
        capabilities: ['Surveillance', 'Reconnaissance', 'Target Acquisition'],
      },
      {
        id: '8',
        callsign: 'RAVEN-2',
        type: 'drone',
        subType: 'uav',
        status: 'standby',
        team: 'Air Assets',
        role: 'Tactical UAV',
        position: {
          lat: 38.8901,
          lng: -77.0321,
          alt: 0,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
        battery: 100,
        signal: 100,
        operatingTime: 0,
        equipment: ['HD Camera', 'Night Vision'],
        capabilities: ['Quick Deploy', 'Low Altitude Recon'],
      },
      // Sensor entities
      {
        id: '9',
        callsign: 'THERMAL-CAM-01',
        type: 'sensor',
        subType: 'thermal',
        status: 'active',
        team: 'Perimeter Security',
        role: 'Thermal Imaging',
        position: {
          lat: 38.8955,
          lng: -77.0401,
          alt: 15,
          speed: 0,
          heading: 270
        },
        lastUpdate: new Date().toISOString(),
        temperature: 22.5,
        signal: 100,
        sensorData: [
          { type: 'heat_signature', value: 3, unit: 'targets', timestamp: new Date().toISOString() },
          { type: 'temperature_range', value: '18-37', unit: '°C', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Heat Detection', '360° Coverage', 'Auto-Tracking'],
      },
      {
        id: '10',
        callsign: 'MOTION-SENSOR-05',
        type: 'sensor',
        subType: 'motion',
        status: 'active',
        team: 'Perimeter Security',
        role: 'Motion Detection',
        position: {
          lat: 38.8912,
          lng: -77.0378,
          alt: 2,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 92,
        signal: 87,
        sensorData: [
          { type: 'motion_events', value: 2, unit: 'detections/hr', timestamp: new Date().toISOString() },
          { type: 'sensitivity', value: 85, unit: '%', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Motion Detection', 'Vibration Sensing', 'Alert Triggering'],
      },
      // Camera entity
      {
        id: '11',
        callsign: 'OVERWATCH-CAM-3',
        type: 'camera',
        subType: 'surveillance',
        status: 'active',
        team: 'Surveillance',
        role: 'Fixed Camera',
        position: {
          lat: 38.8945,
          lng: -77.0355,
          alt: 25,
          speed: 0,
          heading: 45
        },
        lastUpdate: new Date().toISOString(),
        signal: 100,
        equipment: ['4K Resolution', 'PTZ Control', 'IR Illumination'],
        capabilities: ['24/7 Recording', 'Motion Tracking', 'Face Detection'],
      },
      // Vehicle entity
      {
        id: '12',
        callsign: 'VICTOR-1',
        type: 'vehicle',
        subType: 'tactical',
        status: 'active',
        team: 'Mobile Command',
        role: 'Command Vehicle',
        position: {
          lat: 38.8967,
          lng: -77.0342,
          alt: 105,
          speed: 15.2,
          heading: 180
        },
        lastUpdate: new Date().toISOString(),
        fuelLevel: 68,
        signal: 90,
        temperature: 85,
        equipment: ['Communications Hub', 'Satellite Uplink', 'Command Console'],
        capabilities: ['Mobile Command', 'Communications Relay', 'Power Generation'],
      },
      // CBRN Sensor
      {
        id: '13',
        callsign: 'CBRN-DETECT-1',
        type: 'sensor',
        subType: 'cbrn',
        status: 'active',
        team: 'HAZMAT',
        role: 'CBRN Detection',
        position: {
          lat: 38.8934,
          lng: -77.0398,
          alt: 5,
          speed: 0,
          heading: 0
        },
        lastUpdate: new Date().toISOString(),
        battery: 88,
        signal: 92,
        sensorData: [
          { type: 'radiation_level', value: 0.12, unit: 'μSv/h', timestamp: new Date().toISOString() },
          { type: 'chemical_agents', value: 'None', unit: 'detection', timestamp: new Date().toISOString() },
          { type: 'biological_threat', value: 'Clear', unit: 'status', timestamp: new Date().toISOString() }
        ],
        capabilities: ['Chemical Detection', 'Radiation Monitoring', 'Biological Agent Detection'],
        notes: 'Continuous monitoring active. All readings within safe parameters.',
      },
    ];

    setEntities(mockEntities);
  }, []);

  // Get icon for entity type
  const getEntityIcon = (entity: Entity) => {
    if (entity.type === 'friendly' && entity.role === 'Medic') return 'medic';
    if (entity.type === 'friendly') return 'soldier';
    if (entity.type === 'hostile') return 'hostile';
    if (entity.type === 'neutral') return 'neutral';
    if (entity.type === 'unknown') return 'unknown';
    if (entity.type === 'drone') return 'drone';
    if (entity.type === 'sensor') {
      if (entity.subType === 'thermal' || entity.subType === 'motion') return 'sensor';
      if (entity.subType === 'cbrn' || entity.subType === 'radiation') return 'radar';
      return 'sensor';
    }
    if (entity.type === 'camera') return 'camera';
    if (entity.type === 'vehicle') return 'vehicle';
    if (entity.type === 'equipment') return 'equipment';
    return 'radio';
  };

  // Get status color
  const getStatusColor = (status: Entity['status']) => {
    switch (status) {
      case 'active': return '#4ade80';
      case 'inactive': return '#94a3b8';
      case 'stale': return '#fbbf24';
      case 'lost': return '#f87171';
      case 'standby': return '#60a5fa';
      case 'maintenance': return '#fb923c';
      default: return '#94a3b8';
    }
  };

  // Get entity type color
  const getEntityColor = (type: Entity['type']) => {
    switch (type) {
      case 'friendly': return '#00d4aa';
      case 'hostile': return '#ef4444';
      case 'neutral': return '#10b981';
      case 'unknown': return '#f59e0b';
      case 'drone': return '#8b5cf6';
      case 'sensor': return '#00b894';
      case 'camera': return '#ec4899';
      case 'vehicle': return '#6366f1';
      case 'equipment': return '#78716c';
      default: return '#64748b';
    }
  };

  // Format time ago
  const formatTimeAgo = (timestamp: string) => {
    const seconds = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  };

  // Handle entity selection
  const handleSelectEntity = useCallback((entityId: string) => {
    setSelectedEntity(entityId === selectedEntity ? null : entityId);
  }, [selectedEntity]);

  // Handle entity deletion
  const handleDeleteEntity = useCallback((entityId: string) => {
    if (confirm('Are you sure you want to remove this entity?')) {
      setEntities(prev => prev.filter(e => e.id !== entityId));
      if (selectedEntity === entityId) {
        setSelectedEntity(null);
      }
    }
  }, [selectedEntity]);

  // Filter and search entities
  const filteredEntities = useMemo(() => {
    let filtered = entities;

    // Apply type filter
    if (filter !== 'all') {
      switch (filter) {
        case 'personnel':
          filtered = filtered.filter(e => ['friendly', 'hostile', 'neutral', 'unknown'].includes(e.type));
          break;
        case 'drones':
          filtered = filtered.filter(e => e.type === 'drone');
          break;
        case 'sensors':
          filtered = filtered.filter(e => e.type === 'sensor');
          break;
        case 'vehicles':
          filtered = filtered.filter(e => e.type === 'vehicle');
          break;
        default:
          filtered = filtered.filter(e => e.type === filter);
      }
    }

    // Apply offline filter
    if (!showOffline) {
      filtered = filtered.filter(e => e.status === 'active');
    }

    // Apply search
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(entity =>
        entity.callsign.toLowerCase().includes(query) ||
        entity.team.toLowerCase().includes(query) ||
        entity.role.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [entities, filter, showOffline, searchQuery]);

  // Certificate posture for an entity, for the subtle list/grid indicator.
  const certState = useCallback((entityId: string): 'secure' | 'revoked' | 'unsecured' => {
    const c = certs[entityId];
    if (!c) return 'unsecured';
    return c.revoked ? 'revoked' : 'secure';
  }, [certs]);

  const certIndicatorMeta = {
    secure: { icon: 'lock', title: 'Secured — Vault-issued certificate active' },
    revoked: { icon: 'unlock', title: 'Certificate revoked — not secured' },
    unsecured: { icon: 'unlock', title: 'No certificate issued' },
  } as const;

  // --- Device certificate actions (Vault PKI) ---
  const handleIssueCert = useCallback(async (entity: Entity) => {
    setCertBusy(true);
    setCertError(null);
    try {
      const cert = await issueDeviceCert(entity.callsign);
      setCerts(prev => ({ ...prev, [entity.id]: cert }));
    } catch (err) {
      setCertError(err instanceof Error ? err.message : 'Failed to issue certificate');
    } finally {
      setCertBusy(false);
    }
  }, []);

  const handleRevokeCert = useCallback(async (entityId: string) => {
    const cert = certs[entityId];
    if (!cert) return;
    setCertBusy(true);
    setCertError(null);
    try {
      await vaultPki('pki/revoke', { serial_number: cert.serial });
      setCerts(prev => ({ ...prev, [entityId]: { ...cert, revoked: true } }));
    } catch (err) {
      setCertError(err instanceof Error ? err.message : 'Failed to revoke certificate');
    } finally {
      setCertBusy(false);
    }
  }, [certs]);

  const handleDownloadBundle = useCallback((cert: DeviceCert) => {
    const bundle =
      `# GoTAK device certificate for ${cert.commonName}\n` +
      `# serial: ${cert.serial}\n\n` +
      `${cert.certificate.trim()}\n\n${cert.privateKey.trim()}\n\n${cert.issuingCa.trim()}\n`;
    const url = URL.createObjectURL(new Blob([bundle], { type: 'application/x-pem-file' }));
    const a = document.createElement('a');
    a.href = url;
    a.download = `${cert.commonName}.pem`;
    a.click();
    URL.revokeObjectURL(url);
  }, []);

  // Get selected entity details
  const selectedEntityDetails = selectedEntity
    ? entities.find(e => e.id === selectedEntity)
    : null;

  // Statistics
  const stats = useMemo(() => {
    const active = entities.filter(e => e.status === 'active').length;
    const personnel = entities.filter(e => ['friendly', 'hostile', 'neutral', 'unknown'].includes(e.type)).length;
    const drones = entities.filter(e => e.type === 'drone').length;
    const sensors = entities.filter(e => e.type === 'sensor').length;
    const vehicles = entities.filter(e => e.type === 'vehicle').length;
    const alerts = entities.filter(e => ['lost', 'stale'].includes(e.status)).length;

    return { active, personnel, drones, sensors, vehicles, alerts };
  }, [entities]);

  return (
    <div className="entities-page">
      {/* Header */}
      <div className="entities-header">
        <div className="header-left">
          <h1>Entity Management</h1>
          <div className="entity-stats">
            <div className="stat">
              <Icon name="radio" size={16} />
              <span>{stats.active}/{entities.length} Active</span>
            </div>
            <div className="stat">
              <Icon name="user" size={16} />
              <span>{stats.personnel} Personnel</span>
            </div>
            <div className="stat">
              <Icon name="drone" size={16} />
              <span>{stats.drones} Drones</span>
            </div>
            <div className="stat">
              <Icon name="sensor" size={16} />
              <span>{stats.sensors} Sensors</span>
            </div>
            <div className="stat">
              <Icon name="vehicle" size={16} />
              <span>{stats.vehicles} Vehicles</span>
            </div>
            {stats.alerts > 0 && (
              <div className="stat alert">
                <Icon name="alert-triangle" size={16} />
                <span>{stats.alerts} Alerts</span>
              </div>
            )}
          </div>
        </div>

        <div className="header-right">
          <div className="search-box">
            <Icon name="search" size={18} />
            <input
              type="text"
              placeholder="Search entities..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          <div className="filter-controls">
            <select 
              className="filter-dropdown"
              value={filter}
              onChange={(e) => setFilter(e.target.value as EntityFilter)}
            >
              <option value="all">All Types</option>
              <option value="personnel">Personnel</option>
              <option value="drones">Drones</option>
              <option value="sensors">Sensors</option>
              <option value="vehicles">Vehicles</option>
              <option value="friendly">Friendly</option>
              <option value="hostile">Hostile</option>
              <option value="neutral">Neutral</option>
            </select>
            
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={showOffline}
                onChange={(e) => setShowOffline(e.target.checked)}
              />
              <span>Show Offline</span>
            </label>
          </div>

          <div className="view-toggle">
            <button
              className={view === 'grid' ? 'active' : ''}
              onClick={() => setView('grid')}
              title="Grid View"
            >
              <Icon name="grid" size={20} />
            </button>
            <button
              className={view === 'list' ? 'active' : ''}
              onClick={() => setView('list')}
              title="List View"
            >
              <Icon name="list" size={20} />
            </button>
            <button
              className={view === 'tactical' ? 'active' : ''}
              onClick={() => setView('tactical')}
              title="Tactical View"
            >
              <Icon name="tactical" size={20} />
            </button>
          </div>

          <button className="btn-primary">
            <Icon name="plus" size={18} />
            Add Entity
          </button>
        </div>
      </div>


      {/* Main Content */}
      <div className="entities-content">
        {/* Entity List/Grid */}
        <div className={`entities-container ${view}`}>
          {view === 'grid' && (
            <div className="entities-grid">
              {filteredEntities.map(entity => (
                <div
                  key={entity.id}
                  className={`entity-card ${selectedEntity === entity.id ? 'selected' : ''}`}
                  onClick={() => handleSelectEntity(entity.id)}
                >
                  <div className="entity-card-header">
                    <div 
                      className="entity-icon"
                      style={{ backgroundColor: getEntityColor(entity.type) }}
                    >
                      <Icon name={getEntityIcon(entity)} size={24} />
                    </div>
                    <div
                      className="entity-status"
                      style={{ backgroundColor: getStatusColor(entity.status) }}
                    />
                    <span
                      className={`cert-indicator ${certState(entity.id)}`}
                      title={certIndicatorMeta[certState(entity.id)].title}
                    >
                      <Icon name={certIndicatorMeta[certState(entity.id)].icon} size={12} />
                    </span>
                  </div>

                  <div className="entity-card-body">
                    <h3>{entity.callsign}</h3>
                    <p className="entity-role">{entity.role}</p>
                    <p className="entity-team">{entity.team}</p>
                  </div>

                  <div className="entity-card-stats">
                    <div className="stat-row">
                      <Icon name="map-pin" size={14} />
                      <span>{entity.position.lat.toFixed(4)}, {entity.position.lng.toFixed(4)}</span>
                    </div>
                    {entity.position.alt && (
                      <div className="stat-row">
                        <Icon name="altitude" size={14} />
                        <span>{entity.position.alt}m</span>
                      </div>
                    )}
                    {entity.position.speed && entity.position.speed > 0 && (
                      <div className="stat-row">
                        <Icon name="speed" size={14} />
                        <span>{entity.position.speed.toFixed(1)} m/s</span>
                      </div>
                    )}
                    {entity.battery !== undefined && (
                      <div className="stat-row">
                        <Icon name={entity.battery < 30 ? 'battery-low' : 'battery'} size={14} />
                        <span>{entity.battery}%</span>
                      </div>
                    )}
                    {entity.signal !== undefined && (
                      <div className="stat-row">
                        <Icon name="wifi" size={14} />
                        <span>{entity.signal}%</span>
                      </div>
                    )}
                  </div>

                  <div className="entity-card-footer">
                    <span className="last-update">{formatTimeAgo(entity.lastUpdate)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}

          {view === 'list' && (
            <div className="entities-list">
              <table>
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Callsign</th>
                    <th>Team</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Position</th>
                    <th>Battery</th>
                    <th>Signal</th>
                    <th>Last Update</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredEntities.map(entity => (
                    <tr
                      key={entity.id}
                      className={selectedEntity === entity.id ? 'selected' : ''}
                      onClick={() => handleSelectEntity(entity.id)}
                    >
                      <td>
                        <div 
                          className="entity-icon-small"
                          style={{ backgroundColor: getEntityColor(entity.type) }}
                        >
                          <Icon name={getEntityIcon(entity)} size={16} />
                        </div>
                      </td>
                      <td className="callsign">
                        <span
                          className={`cert-indicator ${certState(entity.id)}`}
                          title={certIndicatorMeta[certState(entity.id)].title}
                        >
                          <Icon name={certIndicatorMeta[certState(entity.id)].icon} size={12} />
                        </span>
                        {entity.callsign}
                      </td>
                      <td>{entity.team}</td>
                      <td>{entity.role}</td>
                      <td>
                        <span 
                          className="status-badge"
                          style={{ backgroundColor: getStatusColor(entity.status) }}
                        >
                          {entity.status}
                        </span>
                      </td>
                      <td className="position">
                        {entity.position.lat.toFixed(4)}, {entity.position.lng.toFixed(4)}
                        {entity.position.alt && ` @ ${entity.position.alt}m`}
                      </td>
                      <td>
                        {entity.battery !== undefined && (
                          <div className="battery-indicator">
                            <Icon name={entity.battery < 30 ? 'battery-low' : 'battery'} size={14} />
                            <span>{entity.battery}%</span>
                          </div>
                        )}
                      </td>
                      <td>
                        {entity.signal !== undefined && (
                          <div className="signal-indicator">
                            <Icon name="wifi" size={14} />
                            <span>{entity.signal}%</span>
                          </div>
                        )}
                      </td>
                      <td className="last-update">{formatTimeAgo(entity.lastUpdate)}</td>
                      <td className="actions">
                        <button 
                          className="btn-icon"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleSelectEntity(entity.id);
                          }}
                        >
                          <Icon name="eye" size={16} />
                        </button>
                        <button 
                          className="btn-icon"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteEntity(entity.id);
                          }}
                        >
                          <Icon name="trash-2" size={16} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {view === 'tactical' && (
            <div className="entities-tactical">
              <div className="tactical-grid">
                {filteredEntities.map(entity => (
                  <div
                    key={entity.id}
                    className={`tactical-unit ${selectedEntity === entity.id ? 'selected' : ''}`}
                    onClick={() => handleSelectEntity(entity.id)}
                    style={{
                      borderColor: getEntityColor(entity.type),
                      backgroundColor: selectedEntity === entity.id 
                        ? `${getEntityColor(entity.type)}20` 
                        : 'transparent'
                    }}
                  >
                    <div 
                      className="tactical-icon"
                      style={{ color: getEntityColor(entity.type) }}
                    >
                      <Icon name={getEntityIcon(entity)} size={32} />
                    </div>
                    <div className="tactical-info">
                      <div className="tactical-callsign">{entity.callsign}</div>
                      <div className="tactical-status">
                        <span 
                          className="status-dot"
                          style={{ backgroundColor: getStatusColor(entity.status) }}
                        />
                        {entity.status}
                      </div>
                    </div>
                    {entity.notes && (
                      <div className="tactical-notes">
                        <Icon name="info" size={12} />
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Details Panel */}
        {selectedEntityDetails && (
          <div className="entity-details">
            <div className="details-header">
              <div 
                className="entity-icon-large"
                style={{ backgroundColor: getEntityColor(selectedEntityDetails.type) }}
              >
                <Icon name={getEntityIcon(selectedEntityDetails)} size={32} />
              </div>
              <div className="details-title">
                <h2>{selectedEntityDetails.callsign}</h2>
                <p>{selectedEntityDetails.role} • {selectedEntityDetails.team}</p>
              </div>
              <button 
                className="btn-close"
                onClick={() => setSelectedEntity(null)}
              >
                <Icon name="x" size={20} />
              </button>
            </div>

            <div className="details-content">
              {/* Status Section */}
              <div className="details-section">
                <h3>Status</h3>
                <div className="status-info">
                  <div className="status-item">
                    <span className="label">Status:</span>
                    <span 
                      className="status-badge"
                      style={{ backgroundColor: getStatusColor(selectedEntityDetails.status) }}
                    >
                      {selectedEntityDetails.status}
                    </span>
                  </div>
                  <div className="status-item">
                    <span className="label">Last Update:</span>
                    <span>{formatTimeAgo(selectedEntityDetails.lastUpdate)}</span>
                  </div>
                </div>
              </div>

              {/* Position Section */}
              <div className="details-section">
                <h3>Position</h3>
                <div className="position-info">
                  <div className="position-item">
                    <Icon name="map-pin" size={16} />
                    <span>{selectedEntityDetails.position.lat.toFixed(6)}, {selectedEntityDetails.position.lng.toFixed(6)}</span>
                  </div>
                  {selectedEntityDetails.position.alt !== undefined && (
                    <div className="position-item">
                      <Icon name="altitude" size={16} />
                      <span>{selectedEntityDetails.position.alt}m altitude</span>
                    </div>
                  )}
                  {selectedEntityDetails.position.speed !== undefined && (
                    <div className="position-item">
                      <Icon name="speed" size={16} />
                      <span>{selectedEntityDetails.position.speed.toFixed(1)} m/s</span>
                    </div>
                  )}
                  {selectedEntityDetails.position.heading !== undefined && (
                    <div className="position-item">
                      <Icon name="navigation" size={16} />
                      <span>{selectedEntityDetails.position.heading}° heading</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Vitals Section */}
              {(selectedEntityDetails.battery !== undefined || 
                selectedEntityDetails.signal !== undefined ||
                selectedEntityDetails.temperature !== undefined ||
                selectedEntityDetails.fuelLevel !== undefined) && (
                <div className="details-section">
                  <h3>Vitals</h3>
                  <div className="vitals-grid">
                    {selectedEntityDetails.battery !== undefined && (
                      <div className="vital-item">
                        <Icon name={selectedEntityDetails.battery < 30 ? 'battery-low' : 'battery'} size={20} />
                        <div className="vital-value">
                          <span className="value">{selectedEntityDetails.battery}%</span>
                          <span className="label">Battery</span>
                        </div>
                      </div>
                    )}
                    {selectedEntityDetails.signal !== undefined && (
                      <div className="vital-item">
                        <Icon name="wifi" size={20} />
                        <div className="vital-value">
                          <span className="value">{selectedEntityDetails.signal}%</span>
                          <span className="label">Signal</span>
                        </div>
                      </div>
                    )}
                    {selectedEntityDetails.temperature !== undefined && (
                      <div className="vital-item">
                        <Icon name="thermometer" size={20} />
                        <div className="vital-value">
                          <span className="value">{selectedEntityDetails.temperature}°</span>
                          <span className="label">Temp</span>
                        </div>
                      </div>
                    )}
                    {selectedEntityDetails.fuelLevel !== undefined && (
                      <div className="vital-item">
                        <Icon name="droplet" size={20} />
                        <div className="vital-value">
                          <span className="value">{selectedEntityDetails.fuelLevel}%</span>
                          <span className="label">Fuel</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Sensor Data Section */}
              {selectedEntityDetails.sensorData && selectedEntityDetails.sensorData.length > 0 && (
                <div className="details-section">
                  <h3>Sensor Data</h3>
                  <div className="sensor-data-list">
                    {selectedEntityDetails.sensorData.map((data, index) => (
                      <div key={index} className="sensor-data-item">
                        <span className="sensor-type">{data.type.replace(/_/g, ' ')}</span>
                        <span className="sensor-value">{data.value} {data.unit}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Equipment Section */}
              {selectedEntityDetails.equipment && selectedEntityDetails.equipment.length > 0 && (
                <div className="details-section">
                  <h3>Equipment</h3>
                  <div className="equipment-list">
                    {selectedEntityDetails.equipment.map((item, index) => (
                      <div key={index} className="equipment-item">
                        <Icon name="equipment" size={14} />
                        <span>{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Capabilities Section */}
              {selectedEntityDetails.capabilities && selectedEntityDetails.capabilities.length > 0 && (
                <div className="details-section">
                  <h3>Capabilities</h3>
                  <div className="capabilities-list">
                    {selectedEntityDetails.capabilities.map((capability, index) => (
                      <span key={index} className="capability-badge">
                        {capability}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Notes Section */}
              {selectedEntityDetails.notes && (
                <div className="details-section">
                  <h3>Notes</h3>
                  <div className="notes-content">
                    {selectedEntityDetails.notes}
                  </div>
                </div>
              )}

              {/* Security / Device Certificate (Vault PKI) */}
              <div className="details-section">
                <h3>Security</h3>
                {(() => {
                  const cert = certs[selectedEntityDetails.id];
                  if (!cert) {
                    return (
                      <div className="cert-block">
                        <span className="cert-badge none">
                          <Icon name="unlock" size={14} /> No Certificate
                        </span>
                        <p className="cert-help">
                          Issue an mTLS client certificate from Vault PKI for this device.
                        </p>
                        <button
                          className="btn-cert"
                          disabled={certBusy}
                          onClick={() => setIssueDialogEntity(selectedEntityDetails)}
                        >
                          <Icon name="shield" size={16} />
                          {certBusy ? 'Issuing…' : 'Issue Certificate'}
                        </button>
                      </div>
                    );
                  }
                  return (
                    <div className="cert-block">
                      {cert.revoked ? (
                        <span className="cert-badge revoked">
                          <Icon name="alert-triangle" size={14} /> Revoked
                        </span>
                      ) : (
                        <span className="cert-badge active">
                          <Icon name="lock" size={14} /> Certificate Active
                        </span>
                      )}
                      <dl className="cert-fields">
                        <div><dt>Common Name</dt><dd>{cert.commonName}</dd></div>
                        <div><dt>Serial</dt><dd className="mono">{cert.serial}</dd></div>
                        <div><dt>Issued</dt><dd>{new Date(cert.issuedAt * 1000).toLocaleString()}</dd></div>
                        <div><dt>Expires</dt><dd>{new Date(cert.expiration * 1000).toLocaleString()}</dd></div>
                        <div><dt>SHA-256</dt><dd className="mono cert-fp">{cert.fingerprint}</dd></div>
                      </dl>
                      {cert.revoked ? (
                        <button
                          className="btn-cert"
                          disabled={certBusy}
                          onClick={() => setIssueDialogEntity(selectedEntityDetails)}
                        >
                          <Icon name="shield" size={16} />
                          {certBusy ? 'Issuing…' : 'Re-issue Certificate'}
                        </button>
                      ) : (
                        <div className="cert-actions">
                          <button className="btn-secondary" onClick={() => handleDownloadBundle(cert)}>
                            <Icon name="download" size={16} /> Download
                          </button>
                          <button
                            className="btn-danger"
                            disabled={certBusy}
                            onClick={() => handleRevokeCert(selectedEntityDetails.id)}
                          >
                            <Icon name="alert-triangle" size={16} />
                            {certBusy ? 'Revoking…' : 'Revoke'}
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })()}
                {certError && <div className="cert-error" role="alert">{certError}</div>}
              </div>

              {/* Actions */}
              <div className="details-actions">
                <button className="btn-secondary">
                  <Icon name="map-pin" size={16} />
                  Track on Map
                </button>
                <button className="btn-secondary">
                  <Icon name="message-circle" size={16} />
                  Send Message
                </button>
                <button className="btn-danger" onClick={() => handleDeleteEntity(selectedEntityDetails.id)}>
                  <Icon name="trash-2" size={16} />
                  Remove
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Issue Certificate — duration + operational-security dialog */}
      {issueDialogEntity && (
        <div className="cert-dialog-overlay" onClick={() => !certBusy && setIssueDialogEntity(null)}>
          <div className="cert-dialog" onClick={e => e.stopPropagation()}>
            <div className="cert-dialog-header">
              <span className="cert-dialog-shield"><Icon name="shield" size={22} /></span>
              <div>
                <h2>Issue Device Certificate</h2>
                <p className="cert-dialog-device">
                  {issueDialogEntity.callsign} • {issueDialogEntity.role} • {issueDialogEntity.team}
                </p>
              </div>
            </div>

            <div className="cert-dialog-body">
              <div className="cert-dialog-duration">
                <Icon name="lock" size={18} />
                <div>
                  <strong>Valid for {CERT_TTL_LABEL}</strong>
                  <span>
                    A short-lived mTLS client certificate will be issued from HashiCorp Vault and will
                    auto-expire after {CERT_TTL_LABEL}. Short lifetimes limit exposure if a device is lost,
                    captured, or compromised — re-issue before expiry to keep the unit trusted.
                  </span>
                </div>
              </div>

              <div className="cert-dialog-warning">
                <Icon name="alert-triangle" size={18} />
                <div>
                  <strong>Secure communications are not optional on the tactical edge.</strong>
                  <p>
                    Every radio, drone, sensor, and handheld is a potential foothold for an adversary.
                    A device certificate cryptographically proves this unit is who it claims to be — so
                    spoofed contacts, intercepted positions, and rogue nodes cannot poison the common
                    operating picture. An unsecured device is an unverified device: treat it as untrusted
                    until a certificate is issued.
                  </p>
                </div>
              </div>
            </div>

            <div className="cert-dialog-actions">
              <button
                className="btn-secondary"
                disabled={certBusy}
                onClick={() => setIssueDialogEntity(null)}
              >
                Cancel
              </button>
              <button
                className="btn-cert"
                disabled={certBusy}
                onClick={async () => {
                  const target = issueDialogEntity;
                  await handleIssueCert(target);
                  setIssueDialogEntity(null);
                }}
              >
                <Icon name="shield" size={16} />
                {certBusy ? 'Issuing…' : `Issue Certificate (${CERT_TTL_LABEL})`}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default EntitiesNew;