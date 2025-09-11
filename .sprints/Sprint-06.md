# Sprint 6: Communication Systems

**Duration:** 2 weeks  
**Theme:** Real-time Communication & Alerts  
**Sprint Goals:** Implement comprehensive communication systems for tactical coordination

## Objectives

1. **Real-time Chat System**: Multi-room chat with tactical messaging
2. **Emergency Alerts**: Priority alert system with escalation
3. **Broadcast Messages**: System-wide announcements and notifications  
4. **Message Classification**: Security classification and handling
5. **Communication History**: Message archival and search

## User Stories

### Epic: Tactical Communication Platform

**US-6.1: Multi-Room Chat System**
```
As an operator
I want to communicate in different chat rooms for different operations
So that I can coordinate with specific teams without cluttering other channels
```

**Acceptance Criteria:**
- Create and manage multiple chat rooms
- Join/leave rooms with proper permissions
- Room-specific message history and participants
- Real-time message delivery with typing indicators
- Message reactions and acknowledgments

**US-6.2: Emergency Alert System**
```
As a mission commander  
I want to send emergency alerts that bypass normal communication
So that I can immediately notify all personnel of critical situations
```

**Acceptance Criteria:**
- Priority alert levels (Low, Medium, High, Critical, Emergency)
- Alert broadcast to all or selected groups
- Visual and audio alert indicators in UI
- Alert acknowledgment tracking
- Alert escalation if not acknowledged

**US-6.3: Tactical Message Classification**
```
As a security officer
I want all messages to be properly classified and handled
So that sensitive information is protected according to policy
```

**Acceptance Criteria:**
- Message classification levels (UNCLASSIFIED, RESTRICTED, CONFIDENTIAL, SECRET)
- Automatic classification based on content and context
- Classification-based access controls
- Audit trail for all classified communications
- Warning indicators for classification violations

**US-6.4: Broadcast and Announcements**
```
As a system administrator
I want to send system-wide broadcasts and announcements
So that I can inform all users of important information
```

**Acceptance Criteria:**
- System-wide broadcast messages
- Scheduled announcements and notifications
- Message priority and expiration settings
- User notification preferences
- Broadcast message history and analytics

## Technical Implementation

### Enhanced Chat System Architecture

**Real-time Chat Service**
```go
// internal/chat/enhanced_service.go
type EnhancedChatService struct {
    baseService     *Service
    alertManager    *AlertManager
    classifier      *MessageClassifier
    broadcaster     *BroadcastManager
    archiver        *MessageArchiver
    wsHub          *handlers.TacticalWSHub
}

type ChatRoom struct {
    ID              uuid.UUID              `json:"id" db:"id"`
    Name            string                 `json:"name" db:"name"`
    Description     string                 `json:"description" db:"description"`
    Type            RoomType               `json:"type" db:"type"`
    Classification  Classification         `json:"classification" db:"classification"`
    CreatedBy       uuid.UUID              `json:"created_by" db:"created_by"`
    GroupID         string                 `json:"group_id" db:"group_id"`
    Participants    []RoomParticipant      `json:"participants"`
    Settings        RoomSettings           `json:"settings"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

type RoomType string
const (
    RoomTypeOperational RoomType = "operational"  // Mission-specific rooms
    RoomTypeGeneral     RoomType = "general"      // General discussion
    RoomTypeEmergency   RoomType = "emergency"    // Emergency coordination
    RoomTypeCommand     RoomType = "command"      // Command staff only
    RoomTypeIntel       RoomType = "intel"        // Intelligence sharing
)
```

**Emergency Alert System**
```go
// internal/alerts/manager.go
type AlertManager struct {
    db          database.DB
    logger      *logger.Logger
    wsHub       *handlers.TacticalWSHub
    escalator   *AlertEscalator
}

type TacticalAlert struct {
    ID              uuid.UUID         `json:"id" db:"id"`
    Type            AlertType         `json:"type" db:"type"`
    Priority        AlertPriority     `json:"priority" db:"priority"`
    Title           string            `json:"title" db:"title"`
    Message         string            `json:"message" db:"message"`
    Classification  Classification    `json:"classification" db:"classification"`
    
    // Targeting
    Recipients      []uuid.UUID       `json:"recipients"`
    Groups          []string          `json:"groups"`
    Broadcast       bool              `json:"broadcast" db:"broadcast"`
    
    // Lifecycle
    CreatedBy       uuid.UUID         `json:"created_by" db:"created_by"`
    ExpiresAt       *time.Time        `json:"expires_at" db:"expires_at"`
    AcknowledgedBy  []AlertAck        `json:"acknowledged_by"`
    EscalatedAt     *time.Time        `json:"escalated_at" db:"escalated_at"`
    
    // Metadata
    Location        *Location         `json:"location,omitempty"`
    RelatedMission  *uuid.UUID        `json:"related_mission,omitempty"`
    Attachments     []Attachment      `json:"attachments,omitempty"`
    
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

type AlertType string
const (
    AlertTypeSystem     AlertType = "system"       // System notifications
    AlertTypeTactical   AlertType = "tactical"     // Tactical updates
    AlertTypeEmergency  AlertType = "emergency"    // Emergency situations
    AlertTypeSecurity   AlertType = "security"     // Security incidents
    AlertTypeWeather    AlertType = "weather"      // Weather alerts
    AlertTypeMedical    AlertType = "medical"      // Medical emergencies
)

type AlertPriority string
const (
    PriorityLow       AlertPriority = "low"
    PriorityMedium    AlertPriority = "medium"  
    PriorityHigh      AlertPriority = "high"
    PriorityCritical  AlertPriority = "critical"
    PriorityEmergency AlertPriority = "emergency"
)
```

**Message Classification System**
```go
// internal/classification/classifier.go
type MessageClassifier struct {
    rules        []ClassificationRule
    keywords     map[Classification][]string
    patterns     map[Classification][]*regexp.Regexp
    mlModel      *ClassificationModel  // Optional ML-based classification
}

type ClassificationRule struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Pattern     string          `json:"pattern"`
    Keywords    []string        `json:"keywords"`
    Level       Classification  `json:"level"`
    Priority    int             `json:"priority"`
    Active      bool            `json:"active"`
}

func (c *MessageClassifier) ClassifyMessage(message *ChatMessage) Classification {
    // 1. Check for explicit classification markers
    if explicit := c.extractExplicitClassification(message.Text); explicit != ClassificationUnclassified {
        return explicit
    }
    
    // 2. Apply keyword-based classification
    if keywordLevel := c.classifyByKeywords(message.Text); keywordLevel != ClassificationUnclassified {
        return keywordLevel
    }
    
    // 3. Apply pattern-based classification  
    if patternLevel := c.classifyByPatterns(message.Text); patternLevel != ClassificationUnclassified {
        return patternLevel
    }
    
    // 4. Context-based classification (room, participants, time)
    if contextLevel := c.classifyByContext(message); contextLevel != ClassificationUnclassified {
        return contextLevel
    }
    
    // Default to room classification or UNCLASSIFIED
    return message.Room.Classification
}
```

### Frontend Communication UI

**Chat Interface Component**
```typescript
// src/components/communication/ChatInterface.tsx
import { useState, useEffect } from 'react';
import { Card, CardContent, Tabs, Tab, Badge } from '@mui/material';
import { ChatRoom } from './ChatRoom';
import { AlertPanel } from './AlertPanel';  
import { BroadcastPanel } from './BroadcastPanel';
import { useChat } from '../hooks/useChat';
import { useAlerts } from '../hooks/useAlerts';

export const ChatInterface: React.FC = () => {
  const { rooms, activeRoom, setActiveRoom, unreadCounts } = useChat();
  const { alerts, unacknowledgedCount } = useAlerts();
  const [currentTab, setCurrentTab] = useState(0);

  return (
    <Card className="chat-interface" sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Tabs value={currentTab} onChange={(e, newValue) => setCurrentTab(newValue)}>
        <Tab 
          label={
            <span>
              Chat Rooms
              {unreadCounts.total > 0 && (
                <Badge badgeContent={unreadCounts.total} color="error" sx={{ ml: 1 }} />
              )}
            </span>
          }
        />
        <Tab 
          label={
            <span>
              Alerts
              {unacknowledgedCount > 0 && (
                <Badge badgeContent={unacknowledgedCount} color="warning" sx={{ ml: 1 }} />
              )}
            </span>
          }
        />
        <Tab label="Broadcasts" />
      </Tabs>
      
      <CardContent sx={{ flex: 1, p: 0 }}>
        {currentTab === 0 && (
          <div className="chat-rooms-panel">
            <RoomList 
              rooms={rooms}
              activeRoom={activeRoom}
              onRoomSelect={setActiveRoom}
              unreadCounts={unreadCounts}
            />
            {activeRoom && <ChatRoom room={activeRoom} />}
          </div>
        )}
        
        {currentTab === 1 && (
          <AlertPanel alerts={alerts} />
        )}
        
        {currentTab === 2 && (
          <BroadcastPanel />
        )}
      </CardContent>
    </Card>
  );
};
```

**Alert System Component**
```typescript
// src/components/communication/AlertPanel.tsx
import { Alert, Button, Chip, List, ListItem } from '@mui/material';
import { formatDistanceToNow } from 'date-fns';
import type { TacticalAlert } from '../../types/alerts';

interface AlertPanelProps {
  alerts: TacticalAlert[];
}

export const AlertPanel: React.FC<AlertPanelProps> = ({ alerts }) => {
  const { acknowledgeAlert, createAlert } = useAlerts();
  
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'emergency': return '#FF0000';
      case 'critical': return '#FF4500';
      case 'high': return '#FFA500';
      case 'medium': return '#FFFF00';
      case 'low': return '#90EE90';
      default: return '#CCCCCC';
    }
  };

  return (
    <div className="alert-panel">
      <div className="alert-controls">
        <Button variant="contained" color="error" onClick={() => createAlert('emergency')}>
          🚨 Emergency Alert
        </Button>
        <Button variant="outlined" onClick={() => createAlert('tactical')}>
          📢 Tactical Update
        </Button>
      </div>
      
      <List className="alert-list">
        {alerts.map((alert) => (
          <ListItem key={alert.id} className={`alert-item priority-${alert.priority}`}>
            <div className="alert-content">
              <div className="alert-header">
                <Chip 
                  label={alert.priority.toUpperCase()}
                  size="small"
                  sx={{ 
                    backgroundColor: getPriorityColor(alert.priority),
                    color: '#000',
                    fontWeight: 'bold'
                  }}
                />
                <Chip 
                  label={alert.classification}
                  size="small"
                  variant="outlined"
                />
                <span className="alert-time">
                  {formatDistanceToNow(new Date(alert.createdAt))} ago
                </span>
              </div>
              
              <h4>{alert.title}</h4>
              <p>{alert.message}</p>
              
              {!alert.acknowledgedBy.some(ack => ack.userId === currentUser.id) && (
                <Button 
                  variant="contained" 
                  size="small"
                  onClick={() => acknowledgeAlert(alert.id)}
                >
                  Acknowledge
                </Button>
              )}
              
              <div className="alert-acknowledgments">
                {alert.acknowledgedBy.length > 0 && (
                  <span>✓ {alert.acknowledgedBy.length} acknowledged</span>
                )}
              </div>
            </div>
          </ListItem>
        ))}
      </List>
    </div>
  );
};
```

### Database Schema Extensions

```sql
-- Enhanced chat rooms
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'general',
    classification VARCHAR(50) DEFAULT 'UNCLASSIFIED',
    created_by UUID REFERENCES users(id),
    group_id VARCHAR(255) NOT NULL,
    max_participants INTEGER DEFAULT 100,
    settings JSONB DEFAULT '{}',
    archived BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Room participants with roles
CREATE TABLE room_participants (
    room_id UUID REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'participant', -- admin, moderator, participant
    joined_at TIMESTAMP DEFAULT NOW(),
    last_read_at TIMESTAMP DEFAULT NOW(),
    muted BOOLEAN DEFAULT false,
    PRIMARY KEY (room_id, user_id)
);

-- Enhanced chat messages with classification
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(255) NOT NULL,
    message_text TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text',
    classification VARCHAR(50) DEFAULT 'UNCLASSIFIED',
    priority VARCHAR(50) DEFAULT 'normal',
    reply_to_id UUID REFERENCES chat_messages(id),
    edited_at TIMESTAMP,
    requires_ack BOOLEAN DEFAULT false,
    location_lat DOUBLE PRECISION,
    location_lng DOUBLE PRECISION,
    attachments JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tactical alerts system
CREATE TABLE tactical_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    classification VARCHAR(50) DEFAULT 'UNCLASSIFIED',
    broadcast BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id),
    expires_at TIMESTAMP,
    escalated_at TIMESTAMP,
    related_mission UUID REFERENCES missions(id),
    location_lat DOUBLE PRECISION,
    location_lng DOUBLE PRECISION,
    attachments JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Alert recipients and acknowledgments
CREATE TABLE alert_recipients (
    alert_id UUID REFERENCES tactical_alerts(id) ON DELETE CASCADE,
    recipient_type VARCHAR(50) NOT NULL, -- user, group, broadcast
    recipient_id VARCHAR(255) NOT NULL,
    acknowledged_at TIMESTAMP,
    acknowledged_by UUID REFERENCES users(id),
    PRIMARY KEY (alert_id, recipient_type, recipient_id)
);

-- Message classification rules
CREATE TABLE classification_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pattern TEXT,
    keywords TEXT[],
    classification_level VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_chat_rooms_group ON chat_rooms(group_id);
CREATE INDEX idx_chat_messages_room_time ON chat_messages(room_id, created_at DESC);
CREATE INDEX idx_tactical_alerts_priority ON tactical_alerts(priority, created_at DESC);
CREATE INDEX idx_alert_recipients_type_id ON alert_recipients(recipient_type, recipient_id);
```

## API Specifications

### Chat System Endpoints
```
POST   /api/v1/chat/rooms                    # Create chat room
GET    /api/v1/chat/rooms                    # List chat rooms
GET    /api/v1/chat/rooms/{id}               # Get room details
PUT    /api/v1/chat/rooms/{id}               # Update room
DELETE /api/v1/chat/rooms/{id}               # Archive room
POST   /api/v1/chat/rooms/{id}/join          # Join room
POST   /api/v1/chat/rooms/{id}/leave         # Leave room
GET    /api/v1/chat/rooms/{id}/messages      # Get room messages
POST   /api/v1/chat/rooms/{id}/messages      # Send message
PUT    /api/v1/chat/messages/{id}            # Edit message
DELETE /api/v1/chat/messages/{id}            # Delete message
POST   /api/v1/chat/messages/{id}/ack        # Acknowledge message
```

### Alert System Endpoints  
```
POST   /api/v1/alerts                        # Create alert
GET    /api/v1/alerts                        # List alerts
GET    /api/v1/alerts/{id}                   # Get alert details
POST   /api/v1/alerts/{id}/acknowledge       # Acknowledge alert
POST   /api/v1/alerts/{id}/escalate          # Escalate alert
DELETE /api/v1/alerts/{id}                   # Cancel alert
GET    /api/v1/alerts/statistics             # Alert statistics
```

### WebSocket Message Types
```json
// Chat messages
{
  "type": "chat_message",
  "payload": {
    "roomId": "uuid",
    "message": ChatMessage,
    "action": "new|update|delete"
  }
}

// Alert notifications
{
  "type": "tactical_alert", 
  "payload": {
    "alert": TacticalAlert,
    "action": "created|acknowledged|escalated"
  }
}

// Typing indicators
{
  "type": "user_typing",
  "payload": {
    "roomId": "uuid",
    "userId": "uuid", 
    "typing": true
  }
}
```

## Testing Strategy

### Unit Tests
```go
func TestAlertManager_CreateEmergencyAlert(t *testing.T) {
    manager := setupTestAlertManager()
    
    alert := &TacticalAlert{
        Type:      AlertTypeEmergency,
        Priority:  PriorityEmergency,
        Title:     "Medical Emergency",
        Message:   "Medical assistance required at CP Alpha",
        Broadcast: true,
    }
    
    createdAlert, err := manager.CreateAlert(context.Background(), alert)
    assert.NoError(t, err)
    assert.Equal(t, PriorityEmergency, createdAlert.Priority)
    
    // Verify broadcast was sent
    assert.True(t, manager.broadcaster.LastBroadcast().Emergency)
}
```

### Integration Tests
```typescript
describe('Chat System Integration', () => {
  test('sends and receives messages in real-time', async () => {
    const room = await createTestRoom();
    const client1 = await connectWebSocketClient('user1');
    const client2 = await connectWebSocketClient('user2');
    
    await client1.joinRoom(room.id);
    await client2.joinRoom(room.id);
    
    const message = 'Test tactical message';
    await client1.sendMessage(room.id, message);
    
    const receivedMessage = await client2.waitForMessage();
    expect(receivedMessage.text).toBe(message);
  });
});
```

## Acceptance Criteria

### Chat System
- [ ] Multiple chat rooms with different purposes
- [ ] Real-time message delivery (< 500ms latency)
- [ ] Message classification and security controls
- [ ] Typing indicators and user presence
- [ ] Message history and search capabilities

### Alert System  
- [ ] Emergency alerts bypass normal communication
- [ ] Alert acknowledgment tracking and escalation
- [ ] Visual and audio alert indicators
- [ ] Alert broadcasting to groups or system-wide
- [ ] Alert analytics and reporting

### Security & Classification
- [ ] All messages properly classified
- [ ] Access controls based on classification
- [ ] Audit trail for all communications
- [ ] Classification violation warnings
- [ ] Secure message storage and transmission

### Performance Requirements
- [ ] Support 100+ concurrent chat users
- [ ] Message delivery within 500ms
- [ ] Alert broadcast to 1000+ users within 2 seconds
- [ ] Chat history queries complete within 1 second
- [ ] Classification processing adds <100ms overhead

## Dependencies

### Backend Dependencies
```go
require (
    github.com/gorilla/websocket v1.5.0
    github.com/lib/pq v1.10.9
    github.com/google/uuid v1.5.0
    github.com/rs/zerolog v1.31.0
)
```

### Frontend Dependencies
```json
{
  "dependencies": {
    "@mui/material": "^5.14.0",
    "@mui/icons-material": "^5.14.0", 
    "date-fns": "^2.30.0",
    "react-virtualized": "^9.22.0"
  }
}
```

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 85%+ coverage
- [ ] Integration tests for real-time features
- [ ] Security testing for classification system
- [ ] Performance benchmarks meet requirements

### Functionality
- [ ] All user stories completed and accepted
- [ ] Real-time communication working reliably
- [ ] Alert system functions under load
- [ ] Classification system accurately categorizes messages
- [ ] Mobile-responsive chat interface

### Security & Compliance
- [ ] Message classification system operational
- [ ] Access controls properly enforced
- [ ] Audit logging captures all communications
- [ ] Security review completed
- [ ] Classification handling compliant with policies

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
