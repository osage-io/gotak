# Sprint 2: REST API & Mission Management Core

**Duration:** 2 weeks  
**Sprint Goal:** Build comprehensive REST API with mission management and enhanced CoT protocol support

## Sprint Objectives

1. Expand REST API beyond authentication to cover all tactical operations
2. Implement comprehensive mission planning and management system
3. Enhance CoT protocol support with database persistence
4. Add real-time WebSocket API for frontend integration
5. Create data classification and group-based visibility controls

## User Stories

### Epic 1: Mission Management System

**As a** Mission Commander  
**I want** to create, manage, and track military missions with personnel and resource assignments  
**So that** operations are properly coordinated and documented

#### Story 1.1: Mission Lifecycle Management
- [ ] Create missions with objectives, timelines, and classifications
- [ ] Update mission status (Planning, Active, Completed, Cancelled)
- [ ] Assign personnel to missions with roles
- [ ] Track mission progress and completion metrics
- [ ] Archive completed missions with after-action reports

#### Story 1.2: Task Management within Missions
- [ ] Break missions into discrete tasks and subtasks
- [ ] Assign tasks to specific personnel with deadlines
- [ ] Track task dependencies and critical path
- [ ] Update task status and completion percentage
- [ ] Generate task completion reports

#### Story 1.3: Resource Allocation
- [ ] Assign equipment and assets to missions
- [ ] Track resource availability and conflicts
- [ ] Manage resource checkout and return processes
- [ ] Monitor resource utilization and maintenance schedules
- [ ] Generate resource allocation reports

#### Story 1.4: Mission Templates and SOPs
- [ ] Create reusable mission templates
- [ ] Define Standard Operating Procedures (SOPs) per mission type
- [ ] Apply templates to new missions with customization
- [ ] Version control for templates and SOPs
- [ ] Share templates across organizations

### Epic 2: Enhanced CoT Protocol & Persistence

**As an** Operator  
**I want** all CoT messages properly stored and accessible for historical analysis  
**So that** patterns can be identified and lessons learned

#### Story 2.1: CoT Message Persistence
- [ ] Store all incoming CoT messages with full metadata
- [ ] Create time-series indexes for position and telemetry data
- [ ] Implement message deduplication and conflict resolution
- [ ] Add message versioning and update tracking
- [ ] Support message deletion with audit trails

#### Story 2.2: Historical Position Tracking
- [ ] Maintain complete position history for all assets
- [ ] Generate movement trails and tracks
- [ ] Calculate speed, bearing, and movement patterns
- [ ] Support position interpolation for gaps
- [ ] Create position-based alerts and geofences

#### Story 2.3: Enhanced Message Types
- [ ] Support additional CoT message types (routes, boundaries, etc.)
- [ ] Add custom metadata fields for organizational needs
- [ ] Implement message priorities and urgency levels
- [ ] Support message encryption and secure channels
- [ ] Add message expiration and cleanup policies

#### Story 2.4: Message Search and Analysis
- [ ] Full-text search across message content
- [ ] Filter messages by type, time, user, location
- [ ] Generate message statistics and patterns
- [ ] Export message data for external analysis
- [ ] Create message replay capabilities

### Epic 3: Group-Based Visibility and Data Classification

**As a** Security Administrator  
**I want** fine-grained control over data visibility based on classification and group membership  
**So that** sensitive information is properly protected

#### Story 3.1: Data Classification System
- [ ] Implement classification levels (UNCLASSIFIED, RESTRICTED, SECRET, etc.)
- [ ] Auto-classify data based on content and source
- [ ] Enforce classification-based access controls
- [ ] Add declassification workflows and schedules
- [ ] Generate classification compliance reports

#### Story 3.2: Group-Based Visibility
- [ ] Create organizational groups and teams
- [ ] Control data visibility based on group membership
- [ ] Support hierarchical group structures
- [ ] Allow cross-group collaboration with approval
- [ ] Track group-based data access patterns

#### Story 3.3: Need-to-Know Access Controls
- [ ] Implement compartmentalized information access
- [ ] Support project-based access controls
- [ ] Add temporary access grants with expiration
- [ ] Create access request and approval workflows
- [ ] Monitor and audit privileged access

### Epic 4: Real-Time WebSocket API

**As a** Frontend Developer  
**I want** a comprehensive WebSocket API for real-time updates  
**So that** the web interface can provide live tactical awareness

#### Story 4.1: WebSocket Connection Management
- [ ] Implement secure WebSocket authentication
- [ ] Support connection multiplexing and routing
- [ ] Add automatic reconnection with backoff
- [ ] Handle connection scaling and load balancing
- [ ] Monitor connection health and metrics

#### Story 4.2: Real-Time Data Streaming
- [ ] Stream position updates in real-time
- [ ] Push mission status changes immediately
- [ ] Deliver chat messages with minimal latency
- [ ] Send system alerts and notifications
- [ ] Support selective data streaming based on subscriptions

#### Story 4.3: Client-Side Caching and Synchronization
- [ ] Implement client-side data caching
- [ ] Support offline operation with sync on reconnect
- [ ] Handle conflict resolution for concurrent edits
- [ ] Add optimistic updates with rollback
- [ ] Maintain data consistency across multiple clients

## Technical Implementation

### Database Schema Extensions

```sql
-- Missions table
CREATE TABLE missions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    classification TEXT DEFAULT 'UNCLASSIFIED',
    status TEXT DEFAULT 'planning',
    priority INTEGER DEFAULT 3,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    commander_id TEXT REFERENCES users(id),
    created_by TEXT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Mission tasks
CREATE TABLE mission_tasks (
    id TEXT PRIMARY KEY,
    mission_id TEXT REFERENCES missions(id) ON DELETE CASCADE,
    parent_task_id TEXT REFERENCES mission_tasks(id),
    name TEXT NOT NULL,
    description TEXT,
    assigned_to TEXT REFERENCES users(id),
    status TEXT DEFAULT 'assigned',
    priority INTEGER DEFAULT 3,
    due_date TIMESTAMP,
    completion_percent INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CoT message persistence (time-series optimized)
CREATE TABLE cot_messages (
    id TEXT PRIMARY KEY,
    uid TEXT NOT NULL,
    message_type TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT REFERENCES users(id),
    classification TEXT DEFAULT 'UNCLASSIFIED',
    raw_xml TEXT NOT NULL,
    parsed_data TEXT, -- JSON
    position_lat REAL,
    position_lon REAL,
    position_hae REAL,
    callsign TEXT,
    group_name TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Position history (time-series)
CREATE TABLE position_history (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    altitude REAL DEFAULT 0,
    accuracy REAL,
    speed REAL,
    bearing REAL,
    source TEXT DEFAULT 'cot'
);

-- Groups and teams
CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    classification TEXT DEFAULT 'UNCLASSIFIED',
    parent_group_id TEXT REFERENCES groups(id),
    created_by TEXT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Group memberships
CREATE TABLE group_memberships (
    user_id TEXT REFERENCES users(id),
    group_id TEXT REFERENCES groups(id),
    role TEXT DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, group_id)
);
```

### REST API Endpoints

```yaml
# Mission Management
GET    /api/v1/missions                    # List missions
POST   /api/v1/missions                    # Create mission
GET    /api/v1/missions/{id}               # Get mission details
PUT    /api/v1/missions/{id}               # Update mission
DELETE /api/v1/missions/{id}               # Delete mission
POST   /api/v1/missions/{id}/status        # Update mission status
GET    /api/v1/missions/{id}/personnel     # List mission personnel
POST   /api/v1/missions/{id}/personnel     # Assign personnel to mission
DELETE /api/v1/missions/{id}/personnel/{user_id} # Remove personnel

# Task Management
GET    /api/v1/missions/{id}/tasks         # List mission tasks
POST   /api/v1/missions/{id}/tasks         # Create task
GET    /api/v1/tasks/{id}                  # Get task details
PUT    /api/v1/tasks/{id}                  # Update task
DELETE /api/v1/tasks/{id}                  # Delete task
POST   /api/v1/tasks/{id}/assign           # Assign task
POST   /api/v1/tasks/{id}/complete         # Mark task complete

# CoT Messages and History
GET    /api/v1/cot/messages                # List CoT messages
GET    /api/v1/cot/messages/{id}           # Get message details
POST   /api/v1/cot/search                  # Search messages
GET    /api/v1/positions                   # Get current positions
GET    /api/v1/positions/history           # Get position history
GET    /api/v1/positions/trails            # Get movement trails

# Groups and Teams
GET    /api/v1/groups                      # List groups
POST   /api/v1/groups                      # Create group
GET    /api/v1/groups/{id}                 # Get group details
PUT    /api/v1/groups/{id}                 # Update group
DELETE /api/v1/groups/{id}                 # Delete group
GET    /api/v1/groups/{id}/members         # List group members
POST   /api/v1/groups/{id}/members         # Add member to group
DELETE /api/v1/groups/{id}/members/{user_id} # Remove member

# Real-time subscriptions (WebSocket)
/ws/positions                              # Position updates
/ws/missions                               # Mission status changes
/ws/messages                               # Chat and system messages
/ws/alerts                                 # Emergency alerts and notifications
```

### WebSocket Message Protocol

```json
{
  "type": "position_update",
  "timestamp": "2025-09-08T18:30:00Z",
  "data": {
    "user_id": "user-123",
    "callsign": "Alpha-1",
    "position": {
      "lat": 37.7749,
      "lon": -122.4194,
      "alt": 100.0
    },
    "metadata": {
      "accuracy": 5.0,
      "speed": 25.0,
      "bearing": 180.0
    }
  },
  "classification": "UNCLASSIFIED"
}

{
  "type": "mission_update",
  "timestamp": "2025-09-08T18:30:00Z",
  "data": {
    "mission_id": "mission-456",
    "status": "active",
    "updated_by": "commander-789"
  },
  "classification": "RESTRICTED"
}

{
  "type": "chat_message",
  "timestamp": "2025-09-08T18:30:00Z",
  "data": {
    "from": "Alpha-1",
    "to": "all",
    "message": "Objective secured",
    "channel": "tactical"
  },
  "classification": "UNCLASSIFIED"
}
```

## Acceptance Criteria

### Mission Management
- [ ] Commanders can create missions with full details and classifications
- [ ] Personnel can be assigned to missions with appropriate roles
- [ ] Mission status changes are tracked and audited
- [ ] Tasks can be created and assigned with dependencies
- [ ] Resource allocation prevents conflicts and over-allocation

### CoT Protocol Enhancement
- [ ] All CoT messages are stored with complete metadata
- [ ] Position history is maintained with configurable retention
- [ ] Message search returns results within 1 second
- [ ] Historical data can be exported in standard formats
- [ ] Message replay functionality works accurately

### Data Classification
- [ ] Classification levels are enforced on all data access
- [ ] Group-based visibility restricts data appropriately
- [ ] Classification changes are audited and logged
- [ ] Declassification workflows function correctly
- [ ] Cross-group collaboration requires proper authorization

### WebSocket API
- [ ] WebSocket connections authenticate properly
- [ ] Real-time updates are delivered within 100ms
- [ ] Connection failures are handled gracefully with reconnect
- [ ] Multiple clients can subscribe to different data streams
- [ ] Message delivery is reliable and ordered

### Performance Requirements
- [ ] API endpoints respond within 200ms for 95% of requests
- [ ] WebSocket can handle 1000+ concurrent connections
- [ ] Database queries are optimized with proper indexes
- [ ] Large result sets are paginated appropriately
- [ ] Memory usage remains stable under load

## Development Tasks

### Week 1: Mission Management & Database
- [ ] Implement mission CRUD operations
- [ ] Add task management functionality
- [ ] Enhance CoT message persistence
- [ ] Create group and team management
- [ ] Add data classification system

### Week 2: WebSocket API & Integration
- [ ] Implement WebSocket server and protocol
- [ ] Add real-time data streaming
- [ ] Create position history and trails
- [ ] Implement message search functionality
- [ ] Add comprehensive API testing

## Testing Strategy

### Unit Tests
- [ ] Mission management operations
- [ ] CoT message parsing and storage
- [ ] Classification enforcement
- [ ] WebSocket message handling

### Integration Tests
- [ ] End-to-end mission workflows
- [ ] Real-time data streaming
- [ ] Multi-user collaboration scenarios
- [ ] Cross-group data access controls

### Performance Tests
- [ ] API load testing with 1000+ concurrent users
- [ ] WebSocket scalability testing
- [ ] Database performance with large datasets
- [ ] Memory and CPU usage profiling

## Dependencies

### New Go Packages
- `github.com/gorilla/websocket` - WebSocket support
- `github.com/go-chi/chi/v5` - HTTP router and middleware
- `github.com/go-chi/cors` - CORS middleware
- `github.com/patrickmn/go-cache` - In-memory caching
- `github.com/google/uuid` - UUID generation

### External Services (Optional)
- Redis for WebSocket scaling across multiple servers
- TimescaleDB extension for PostgreSQL time-series optimization

---

## Sprint Retrospective Template

### What went well?
- 

### What could be improved?
- 

### Action items for next sprint:
- 

### Blockers encountered:
- 

### Technical debt created:
- 
