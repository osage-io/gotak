# GoTAK Updated Sprint Plan - January 2025

## Current Status & Context

**Date:** January 2025  
**Current Sprint:** Sprint 3 (85% Complete)  
**Next Priority:** Complete Sprint 3 → Sprint 4 Frontend Foundation  

### Sprint Progress Summary
- ✅ **Sprint 1**: Database & Auth Foundation (95% Complete)
- ✅ **Sprint 2**: REST API & Mission Management (100% Complete) 
- 🚧 **Sprint 3**: Mission Planning Service (85% Complete)
- 📋 **Sprint 4**: Interactive Maps & Positioning (Ready to Start)
- 📋 **Sprint 5**: Mission Management UI (Planned)

## Immediate Next Steps (Week 1)

### 1. Complete Sprint 3 Remaining Items (15% - Estimated 8 hours)
**Priority: HIGH - Finish current sprint cleanly**

**Tasks to Complete:**
- [ ] Fix 4 failing tests in timeline/mocking complexity
- [ ] Resolve 10 skipped tests (QueryContext empty result mocking)
- [ ] Add event replay capabilities to event system  
- [ ] Optimize database queries for large mission datasets
- [ ] Update API documentation with new mission endpoints

**Timeline:** 2-3 days  
**Owner:** Backend team  

### 2. Sprint 4 Preparation (Parallel to Sprint 3 completion)
**Priority: HIGH - Foundation for frontend development**

**Tasks to Start:**
- [ ] Set up React/TypeScript frontend foundation
- [ ] Install and configure Leaflet.js mapping library
- [ ] Create basic map container component
- [ ] Set up WebSocket client connection
- [ ] Design tactical map UI components

**Timeline:** 3-4 days  
**Owner:** Frontend team  

## Sprint 4 Revised Plan: Interactive Maps & Real-Time UI

**Duration:** 2 weeks  
**Revised Goals:** Build interactive mapping with real-time mission integration

### Week 1: Foundation & Core Mapping
1. **React Frontend Setup**
   - Create-react-app with TypeScript
   - Material-UI tactical theme implementation
   - Basic authentication flow integration
   - WebSocket client setup

2. **Interactive Map Implementation**
   - Leaflet.js integration with multiple base layers
   - Basic entity marker system
   - Real-time position updates from WebSocket
   - Map controls and responsive design

### Week 2: Tactical Features & Integration  
1. **Mission Integration on Map**
   - Display mission locations and areas of interest
   - Show mission status and progress on map
   - Click-to-view mission details popup
   - Real-time mission updates

2. **Enhanced Map Features**
   - Drawing tools for tactical overlays
   - Layer management system  
   - Position history trails
   - Performance optimization for 100+ entities

## Sprint 5 Revised Plan: Mission Management Frontend

**Duration:** 2 weeks  
**Goals:** Complete mission management UI with full CRUD capabilities

### Sprint 5 Focus Areas:
1. **Mission Dashboard** - Overview of all missions with status
2. **Mission Planning Interface** - Create and edit missions
3. **Task Management UI** - Assign and track mission tasks  
4. **Resource Management** - Personnel and equipment allocation
5. **Real-time Collaboration** - Multi-user mission editing

## Updated Long-Term Roadmap

### Sprints 6-8: Core Platform Completion
- **Sprint 6**: Communication Systems (Chat, Alerts, Emergency)
- **Sprint 7**: Advanced Authentication & User Management  
- **Sprint 8**: Persistence Layer & Audit Logging

### Sprints 9-12: Enterprise Features
- **Sprint 9**: Observability & External APIs
- **Sprint 10**: Federation, Scalability & Hardening
- **Sprint 11**: Advanced Security & Compliance  
- **Sprint 12**: Performance & Production Deployment

## Technical Decisions & Architecture Updates

### Frontend Technology Stack
```json
{
  "core": {
    "react": "^18.2.0",
    "typescript": "^5.0.0", 
    "@mui/material": "^5.14.0"
  },
  "mapping": {
    "leaflet": "^1.9.4",
    "react-leaflet": "^4.2.1"
  },
  "state": {
    "zustand": "^4.4.0",
    "react-query": "^3.39.0"
  },
  "websockets": {
    "socket.io-client": "^4.7.0"
  }
}
```

### Backend Enhancements
```go
// Add to existing GoTAK server
type WebSocketManager struct {
    clients    map[string]*Client
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

// Enhanced API endpoints
// GET /api/v1/missions/{id}/map-data
// POST /api/v1/overlays  
// GET /api/v1/entities/positions
// WebSocket: /ws/tactical-updates
```

### Database Schema Updates
```sql
-- Add to existing schema
CREATE TABLE tactical_overlays (
    id UUID PRIMARY KEY,
    mission_id UUID REFERENCES missions(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100), -- point, line, polygon, text
    geometry JSONB,
    properties JSONB,
    created_by UUID,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE entity_positions (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(255) NOT NULL,
    callsign VARCHAR(255),
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    group_id VARCHAR(255),
    INDEX idx_positions_uid_time (uid, timestamp)
);
```

## Success Metrics & Goals

### Sprint 4 Success Criteria
- [ ] Interactive map loads and displays correctly
- [ ] Real-time entity positions update within 1 second
- [ ] Map handles 100+ entities without performance issues
- [ ] Mission locations display with status information
- [ ] Mobile-responsive design works on tablets

### Sprint 5 Success Criteria  
- [ ] Complete mission CRUD operations via UI
- [ ] Task assignment and tracking functional
- [ ] Real-time collaborative editing works
- [ ] Resource management interface complete
- [ ] Mission dashboard shows accurate status

## Risk Mitigation

### Technical Risks
1. **Frontend Complexity**: Start simple, iterate quickly
2. **Real-time Performance**: Implement throttling and optimization early
3. **Map Performance**: Use clustering and lazy loading
4. **WebSocket Reliability**: Implement reconnection and error handling

### Project Risks
1. **Scope Creep**: Strict sprint boundaries and MVP focus
2. **Integration Issues**: Regular testing and continuous integration
3. **UI/UX Complexity**: User testing and iterative design

## Development Workflow Updates

### Sprint 4 Development Process
```bash
# Frontend development
cd frontend/
npm install
npm run dev  # Hot reload development server

# Backend with WebSocket support  
cd /Users/dfedick/projects/gotak
make dev-ws  # Development with WebSocket support

# Full stack testing
make test-integration  # Backend + Frontend integration tests
```

### Quality Gates
- [ ] Code review required for all changes
- [ ] Unit tests maintain >80% coverage  
- [ ] Integration tests pass for all new features
- [ ] No security vulnerabilities in dependencies
- [ ] Performance benchmarks maintained

## Next Actions Required

### This Week (Immediate)
1. **Complete Sprint 3** - Fix remaining tests and technical debt
2. **Start Sprint 4 Setup** - Initialize React frontend foundation
3. **Review Sprint Plans** - Validate approach with team
4. **Update Documentation** - Ensure all APIs are documented

### Sprint 4 Week 1 Goals
- [ ] React frontend scaffolding complete
- [ ] Basic interactive map working
- [ ] WebSocket connection established
- [ ] First entity markers displaying on map

### Sprint 4 Week 2 Goals  
- [ ] Mission integration with map display
- [ ] Real-time updates functioning
- [ ] Drawing tools and overlays working
- [ ] Performance optimization complete

---

## Contact & Coordination

**Technical Lead:** Review architectural decisions  
**Product Owner:** Validate user story priorities  
**QA Lead:** Coordinate testing strategy  
**DevOps:** Ensure CI/CD pipeline supports frontend

---

**Status:** ✅ Plan Updated - Ready for Execution  
**Next Review:** End of Sprint 3 (Target: Within 3-4 days)  
**Sprint 4 Kickoff:** Immediately following Sprint 3 completion
