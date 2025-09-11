import type { Entity } from '../types/entity';

// Mock entities for testing
export const mockEntities: Entity[] = [
  {
    id: 'entity-1',
    callsign: 'ALPHA-6',
    type: 'ground-infantry',
    affiliation: 'friendly',
    lat: 38.9072,
    lng: -77.0369,
    altitude: 15,
    speed: 2.5,
    course: 45,
    lastUpdate: new Date(Date.now() - 30000), // 30 seconds ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-2',
    callsign: 'BRAVO-3',
    type: 'ground-vehicle-wheeled',
    affiliation: 'friendly',
    lat: 38.9122,
    lng: -77.0319,
    altitude: 12,
    speed: 15.2,
    course: 180,
    lastUpdate: new Date(Date.now() - 45000), // 45 seconds ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-3',
    callsign: 'CHARLIE-1',
    type: 'air-rotorcraft',
    affiliation: 'friendly',
    lat: 38.9022,
    lng: -77.0419,
    altitude: 150,
    speed: 25.0,
    course: 270,
    lastUpdate: new Date(Date.now() - 60000), // 1 minute ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-4',
    callsign: 'TANGO-9',
    type: 'ground-infantry',
    affiliation: 'hostile',
    lat: 38.9172,
    lng: -77.0269,
    altitude: 18,
    speed: 1.8,
    course: 90,
    lastUpdate: new Date(Date.now() - 120000), // 2 minutes ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-5',
    callsign: 'ECHO-2',
    type: 'ground-vehicle-tracked',
    affiliation: 'hostile',
    lat: 38.9222,
    lng: -77.0219,
    altitude: 20,
    speed: 8.3,
    course: 225,
    lastUpdate: new Date(Date.now() - 90000), // 1.5 minutes ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-6',
    callsign: 'NEUTRAL-1',
    type: 'ground-vehicle-wheeled',
    affiliation: 'neutral',
    lat: 38.8972,
    lng: -77.0469,
    altitude: 10,
    speed: 12.0,
    course: 135,
    lastUpdate: new Date(Date.now() - 180000), // 3 minutes ago
    classification: 'unclassified',
    status: 'active',
  },
  {
    id: 'entity-7',
    callsign: 'UNKNOWN-X',
    type: 'ground-infantry',
    affiliation: 'unknown',
    lat: 38.8922,
    lng: -77.0519,
    altitude: 8,
    speed: 0.5,
    course: 0,
    lastUpdate: new Date(Date.now() - 300000), // 5 minutes ago
    classification: 'unclassified',
    status: 'active',
  },
];

// Generate random position updates for mock entities
export const generateRandomUpdate = (entity: Entity): Entity => {
  const latVariation = (Math.random() - 0.5) * 0.001; // ~100m variation
  const lngVariation = (Math.random() - 0.5) * 0.001;
  const speedVariation = (Math.random() - 0.5) * 2; // ±1 m/s
  const courseVariation = (Math.random() - 0.5) * 20; // ±10 degrees

  return {
    ...entity,
    lat: entity.lat + latVariation,
    lng: entity.lng + lngVariation,
    speed: Math.max(0, (entity.speed ?? 0) + speedVariation),
    course: ((entity.course ?? 0) + courseVariation + 360) % 360,
    lastUpdate: new Date(),
  };
};

// Mock WebSocket message generator
export class MockWebSocketService {
  private entities: Map<string, Entity> = new Map();
  private intervalId: number | null = null;
  private callbacks: ((message: any) => void)[] = [];

  constructor() {
    // Initialize with mock entities
    mockEntities.forEach(entity => {
      this.entities.set(entity.id, { ...entity });
    });
  }

  connect() {
    console.log('Mock WebSocket connected');
    
    // Send initial positions
    this.entities.forEach(entity => {
      this.sendMessage({
        type: 'position_update',
        payload: {
          entityId: entity.id,
          position: {
            lat: entity.lat,
            lng: entity.lng,
            altitude: entity.altitude,
            speed: entity.speed,
            course: entity.course,
            timestamp: entity.lastUpdate,
          }
        }
      });
    });

    // Start periodic updates
    this.intervalId = window.setInterval(() => {
      this.generateRandomUpdates();
    }, 5000); // Update every 5 seconds

    return true;
  }

  disconnect() {
    console.log('Mock WebSocket disconnected');
    
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  onMessage(callback: (message: any) => void) {
    this.callbacks.push(callback);
  }

  private sendMessage(message: any) {
    this.callbacks.forEach(callback => {
      try {
        callback(message);
      } catch (error) {
        console.error('Error in WebSocket message callback:', error);
      }
    });
  }

  private generateRandomUpdates() {
    // Update 2-3 random entities
    const entityIds = Array.from(this.entities.keys());
    const updateCount = Math.floor(Math.random() * 2) + 2; // 2-3 updates
    
    for (let i = 0; i < updateCount; i++) {
      const randomId = entityIds[Math.floor(Math.random() * entityIds.length)];
      const entity = this.entities.get(randomId);
      
      if (entity) {
        const updatedEntity = generateRandomUpdate(entity);
        this.entities.set(randomId, updatedEntity);
        
        this.sendMessage({
          type: 'position_update',
          payload: {
            entityId: randomId,
            position: {
              lat: updatedEntity.lat,
              lng: updatedEntity.lng,
              altitude: updatedEntity.altitude,
              speed: updatedEntity.speed,
              course: updatedEntity.course,
              timestamp: updatedEntity.lastUpdate,
            }
          }
        });
      }
    }
  }
}

// Export singleton instance
export const mockWebSocketService = new MockWebSocketService();
