# Sprint 5: Mission Management UI

**Duration:** 2 weeks  
**Sprint Goals:** Build comprehensive mission planning and management interface

## Objectives

1. **Mission Planning Interface**: Create missions with tasks, resources, and timelines
2. **Task Management System**: Assign and track mission tasks and objectives
3. **Resource Allocation**: Manage personnel and equipment assignments
4. **Mission Timeline**: Visual timeline with milestones and dependencies
5. **Collaborative Planning**: Multi-user mission planning capabilities

## User Stories

### Epic: Mission Management Platform

**US-5.1: Mission Planning Dashboard**
```
As a mission commander
I want to create and plan tactical missions
So that I can coordinate complex operations with clear objectives
```

**Acceptance Criteria:**
- Mission creation form with all required fields
- Mission templates for common operation types
- Drag-and-drop mission builder interface
- Save draft missions and templates
- Mission summary and overview dashboard

**US-5.2: Task and Objective Management**
```
As a mission planner
I want to create tasks with objectives and assignments
So that every team member knows their responsibilities
```

**Acceptance Criteria:**
- Task creation with descriptions, priorities, and deadlines
- Assignment of personnel and equipment to tasks
- Task dependencies and sequencing
- Progress tracking and status updates
- Subtask breakdown and organization

**US-5.3: Resource Management Interface**
```
As a logistics coordinator
I want to track and allocate mission resources
So that I can ensure optimal resource utilization
```

**Acceptance Criteria:**
- Personnel roster with skills and availability
- Equipment inventory with status tracking
- Resource allocation conflict detection
- Resource scheduling and timeline view
- Automatic resource optimization suggestions

**US-5.4: Mission Timeline and Scheduling**
```
As an operations officer
I want to view mission timeline and critical path
So that I can identify bottlenecks and adjust schedules
```

**Acceptance Criteria:**
- Interactive Gantt chart for mission timeline
- Critical path analysis and highlighting
- Timeline drag-and-drop editing
- Milestone markers and dependency lines
- Time zone support for global operations

**US-5.5: Collaborative Mission Planning**
```
As a mission team member
I want to collaborate on mission planning in real-time
So that we can create comprehensive plans together
```

**Acceptance Criteria:**
- Real-time collaborative editing
- Comments and annotations on mission elements
- Change tracking and version history
- Role-based editing permissions
- Notification system for changes and updates

## Technical Implementation

### Frontend Components

**MissionDashboard.tsx**
```typescript
import React, { useState } from 'react';
import { Grid, Card, CardContent, Typography, Button } from '@mui/material';
import { MissionList } from './MissionList';
import { MissionTimeline } from './MissionTimeline';
import { ResourcePanel } from './ResourcePanel';
import { useMissions } from '../hooks/useMissions';

export const MissionDashboard: React.FC = () => {
  const { missions, createMission, updateMission } = useMissions();
  const [selectedMission, setSelectedMission] = useState<string | null>(null);

  return (
    <Grid container spacing={3}>
      <Grid item xs={12} md={4}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Active Missions
            </Typography>
            <MissionList
              missions={missions}
              onSelectMission={setSelectedMission}
              selectedMission={selectedMission}
            />
            <Button
              variant="contained"
              color="primary"
              fullWidth
              onClick={() => createMission()}
              sx={{ mt: 2 }}
            >
              Create New Mission
            </Button>
          </CardContent>
        </Card>
      </Grid>
      
      <Grid item xs={12} md={8}>
        {selectedMission ? (
          <MissionPlanningView missionId={selectedMission} />
        ) : (
          <MissionOverview missions={missions} />
        )}
      </Grid>
    </Grid>
  );
};
```

**MissionPlanningView.tsx**
```typescript
import React, { useState } from 'react';
import { Tabs, Tab, Box, Paper } from '@mui/material';
import { MissionEditor } from './MissionEditor';
import { TaskManager } from './TaskManager';
import { ResourceAllocator } from './ResourceAllocator';
import { MissionTimeline } from './MissionTimeline';
import { useMission } from '../hooks/useMission';

interface MissionPlanningViewProps {
  missionId: string;
}

export const MissionPlanningView: React.FC<MissionPlanningViewProps> = ({ missionId }) => {
  const [currentTab, setCurrentTab] = useState(0);
  const { mission, updateMission, loading } = useMission(missionId);

  if (loading) return <div>Loading...</div>;
  if (!mission) return <div>Mission not found</div>;

  return (
    <Paper elevation={2}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={currentTab} onChange={(e, newValue) => setCurrentTab(newValue)}>
          <Tab label="Overview" />
          <Tab label="Tasks" />
          <Tab label="Resources" />
          <Tab label="Timeline" />
        </Tabs>
      </Box>
      
      <Box sx={{ p: 3 }}>
        {currentTab === 0 && <MissionEditor mission={mission} onUpdate={updateMission} />}
        {currentTab === 1 && <TaskManager mission={mission} onUpdate={updateMission} />}
        {currentTab === 2 && <ResourceAllocator mission={mission} onUpdate={updateMission} />}
        {currentTab === 3 && <MissionTimeline mission={mission} onUpdate={updateMission} />}
      </Box>
    </Paper>
  );
};
```

**TaskManager.tsx**
```typescript
import React, { useState } from 'react';
import {
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip
} from '@mui/material';
import { Add, Edit, Delete, Assignment } from '@mui/icons-material';
import { Task, Mission } from '../types/mission';
import { usePersonnel } from '../hooks/usePersonnel';

interface TaskManagerProps {
  mission: Mission;
  onUpdate: (mission: Mission) => void;
}

export const TaskManager: React.FC<TaskManagerProps> = ({ mission, onUpdate }) => {
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const { personnel } = usePersonnel();

  const handleCreateTask = () => {
    setEditingTask({
      id: '',
      title: '',
      description: '',
      priority: 'medium',
      status: 'planning',
      assignees: [],
      dueDate: new Date(),
      subtasks: []
    });
    setIsDialogOpen(true);
  };

  const handleSaveTask = (task: Task) => {
    const updatedTasks = editingTask?.id
      ? mission.tasks.map(t => t.id === editingTask.id ? task : t)
      : [...mission.tasks, { ...task, id: generateTaskId() }];
    
    onUpdate({ ...mission, tasks: updatedTasks });
    setIsDialogOpen(false);
    setEditingTask(null);
  };

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3>Mission Tasks</h3>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={handleCreateTask}
        >
          Add Task
        </Button>
      </div>
      
      <List>
        {mission.tasks.map((task) => (
          <ListItem key={task.id} divider>
            <ListItemText
              primary={task.title}
              secondary={
                <div>
                  <div>{task.description}</div>
                  <div style={{ marginTop: 8 }}>
                    <Chip label={task.priority} size="small" color={getPriorityColor(task.priority)} />
                    <Chip label={task.status} size="small" style={{ marginLeft: 8 }} />
                    {task.assignees.map(assignee => (
                      <Chip key={assignee} label={assignee} size="small" style={{ marginLeft: 4 }} />
                    ))}
                  </div>
                </div>
              }
            />
            <ListItemSecondaryAction>
              <IconButton onClick={() => handleEditTask(task)}>
                <Edit />
              </IconButton>
              <IconButton onClick={() => handleDeleteTask(task.id)}>
                <Delete />
              </IconButton>
            </ListItemSecondaryAction>
          </ListItem>
        ))}
      </List>

      <TaskEditDialog
        open={isDialogOpen}
        task={editingTask}
        personnel={personnel}
        onSave={handleSaveTask}
        onCancel={() => {
          setIsDialogOpen(false);
          setEditingTask(null);
        }}
      />
    </>
  );
};
```

### Mission Management Hook

**useMissions.ts**
```typescript
import { useState, useEffect } from 'react';
import { useAuth } from './useAuth';
import { useWebSocket } from './useWebSocket';
import { Mission } from '../types/mission';

export const useMissions = () => {
  const [missions, setMissions] = useState<Mission[]>([]);
  const [loading, setLoading] = useState(true);
  const { token } = useAuth();
  const { lastMessage } = useWebSocket();

  useEffect(() => {
    fetchMissions();
  }, []);

  useEffect(() => {
    if (lastMessage?.type === 'mission_update') {
      setMissions(prev => 
        prev.map(m => 
          m.id === lastMessage.data.id ? lastMessage.data : m
        )
      );
    }
  }, [lastMessage]);

  const fetchMissions = async () => {
    try {
      const response = await fetch('/api/v1/missions', {
        headers: { Authorization: `Bearer ${token}` }
      });
      const data = await response.json();
      setMissions(data.missions || []);
    } catch (error) {
      console.error('Failed to fetch missions:', error);
    } finally {
      setLoading(false);
    }
  };

  const createMission = async (missionData?: Partial<Mission>) => {
    try {
      const response = await fetch('/api/v1/missions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          title: 'New Mission',
          description: '',
          status: 'planning',
          priority: 'medium',
          startDate: new Date(),
          endDate: new Date(Date.now() + 24 * 60 * 60 * 1000), // +1 day
          tasks: [],
          resources: [],
          ...missionData
        })
      });
      
      if (response.ok) {
        const newMission = await response.json();
        setMissions(prev => [...prev, newMission]);
        return newMission.id;
      }
    } catch (error) {
      console.error('Failed to create mission:', error);
    }
  };

  const updateMission = async (missionId: string, updates: Partial<Mission>) => {
    try {
      const response = await fetch(`/api/v1/missions/${missionId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify(updates)
      });
      
      if (response.ok) {
        const updatedMission = await response.json();
        setMissions(prev => 
          prev.map(m => m.id === missionId ? updatedMission : m)
        );
      }
    } catch (error) {
      console.error('Failed to update mission:', error);
    }
  };

  const deleteMission = async (missionId: string) => {
    try {
      const response = await fetch(`/api/v1/missions/${missionId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });
      
      if (response.ok) {
        setMissions(prev => prev.filter(m => m.id !== missionId));
      }
    } catch (error) {
      console.error('Failed to delete mission:', error);
    }
  };

  return {
    missions,
    loading,
    createMission,
    updateMission,
    deleteMission,
    refreshMissions: fetchMissions
  };
};
```

### Backend Mission API

**Mission Model**
```go
type Mission struct {
    ID           string                 `json:"id" db:"id"`
    Title        string                 `json:"title" db:"title"`
    Description  string                 `json:"description" db:"description"`
    Status       string                 `json:"status" db:"status"` // planning, active, completed, cancelled
    Priority     string                 `json:"priority" db:"priority"` // low, medium, high, critical
    StartDate    time.Time              `json:"start_date" db:"start_date"`
    EndDate      time.Time              `json:"end_date" db:"end_date"`
    CreatedBy    string                 `json:"created_by" db:"created_by"`
    GroupID      string                 `json:"group_id" db:"group_id"`
    Tasks        []Task                 `json:"tasks"`
    Resources    []ResourceAllocation   `json:"resources"`
    Classification string               `json:"classification" db:"classification"`
    CreatedAt    time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

type Task struct {
    ID          string    `json:"id" db:"id"`
    MissionID   string    `json:"mission_id" db:"mission_id"`
    Title       string    `json:"title" db:"title"`
    Description string    `json:"description" db:"description"`
    Priority    string    `json:"priority" db:"priority"`
    Status      string    `json:"status" db:"status"`
    Assignees   []string  `json:"assignees"`
    DueDate     time.Time `json:"due_date" db:"due_date"`
    ParentTask  *string   `json:"parent_task" db:"parent_task"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ResourceAllocation struct {
    ID           string    `json:"id" db:"id"`
    MissionID    string    `json:"mission_id" db:"mission_id"`
    ResourceType string    `json:"resource_type" db:"resource_type"` // personnel, equipment
    ResourceID   string    `json:"resource_id" db:"resource_id"`
    Quantity     int       `json:"quantity" db:"quantity"`
    AllocatedAt  time.Time `json:"allocated_at" db:"allocated_at"`
    ReleasedAt   *time.Time `json:"released_at" db:"released_at"`
}
```

**Mission Handlers**
```go
// GET /api/v1/missions
func (s *Server) handleGetMissions(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := getUserIDFromContext(ctx)
    groupID := getGroupIDFromContext(ctx)
    
    missions, err := s.db.GetMissionsByGroup(ctx, groupID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Load tasks and resources for each mission
    for i, mission := range missions {
        tasks, err := s.db.GetTasksByMission(ctx, mission.ID)
        if err != nil {
            continue
        }
        missions[i].Tasks = tasks
        
        resources, err := s.db.GetResourcesByMission(ctx, mission.ID)
        if err != nil {
            continue
        }
        missions[i].Resources = resources
    }
    
    response := map[string]interface{}{
        "missions": missions,
        "total":    len(missions),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// POST /api/v1/missions
func (s *Server) handleCreateMission(w http.ResponseWriter, r *http.Request) {
    var mission Mission
    if err := json.NewDecoder(r.Body).Decode(&mission); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    ctx := r.Context()
    userID := getUserIDFromContext(ctx)
    groupID := getGroupIDFromContext(ctx)
    
    mission.ID = generateUUID()
    mission.CreatedBy = userID
    mission.GroupID = groupID
    mission.CreatedAt = time.Now()
    mission.UpdatedAt = time.Now()
    
    if err := s.db.CreateMission(ctx, &mission); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Broadcast mission creation
    s.broadcastMissionUpdate(&mission, "created")
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(mission)
}
```

### Database Schema

```sql
-- Missions table
CREATE TABLE missions (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'planning',
    priority VARCHAR(50) DEFAULT 'medium',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    group_id VARCHAR(255) NOT NULL,
    classification VARCHAR(50) DEFAULT 'UNCLASSIFIED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table
CREATE TABLE tasks (
    id VARCHAR(255) PRIMARY KEY,
    mission_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) DEFAULT 'medium',
    status VARCHAR(50) DEFAULT 'pending',
    assignees JSONB,
    due_date TIMESTAMP,
    parent_task VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (mission_id) REFERENCES missions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_task) REFERENCES tasks(id) ON DELETE SET NULL
);

-- Resource allocations table
CREATE TABLE resource_allocations (
    id VARCHAR(255) PRIMARY KEY,
    mission_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    quantity INTEGER DEFAULT 1,
    allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP,
    FOREIGN KEY (mission_id) REFERENCES missions(id) ON DELETE CASCADE
);

-- Personnel table
CREATE TABLE personnel (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    rank VARCHAR(100),
    unit VARCHAR(255),
    specialties JSONB,
    availability_status VARCHAR(50) DEFAULT 'available',
    group_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Equipment table  
CREATE TABLE equipment (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    status VARCHAR(50) DEFAULT 'available',
    specifications JSONB,
    group_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_missions_group_status ON missions(group_id, status);
CREATE INDEX idx_tasks_mission ON tasks(mission_id);
CREATE INDEX idx_tasks_assignee ON tasks USING GIN(assignees);
CREATE INDEX idx_resource_allocations_mission ON resource_allocations(mission_id);
```

## API Specifications

### REST Endpoints

**Missions API**
```
GET /api/v1/missions
Response: {
  "missions": [Mission],
  "total": number
}

POST /api/v1/missions
Body: Mission (without ID)
Response: Mission

PUT /api/v1/missions/{id}
Body: Partial<Mission>
Response: Mission

DELETE /api/v1/missions/{id}
Response: 204 No Content

GET /api/v1/missions/{id}/tasks
Response: Task[]

POST /api/v1/missions/{id}/tasks
Body: Task (without ID)
Response: Task

PUT /api/v1/tasks/{id}
Body: Partial<Task>
Response: Task

DELETE /api/v1/tasks/{id}
Response: 204 No Content
```

**Resources API**
```
GET /api/v1/personnel
Response: Personnel[]

GET /api/v1/equipment
Response: Equipment[]

POST /api/v1/missions/{id}/allocate
Body: {
  "resource_type": "personnel" | "equipment",
  "resource_id": string,
  "quantity": number
}
Response: ResourceAllocation
```

### WebSocket Messages

**Mission Update**
```json
{
  "type": "mission_update",
  "data": {
    "action": "created" | "updated" | "deleted",
    "mission": Mission
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Task Update**
```json
{
  "type": "task_update", 
  "data": {
    "action": "created" | "updated" | "deleted",
    "task": Task,
    "mission_id": string
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Testing Strategy

### Unit Tests
```typescript
describe('useMissions', () => {
  test('creates new mission', async () => {
    const { result } = renderHook(() => useMissions());
    
    await act(async () => {
      await result.current.createMission({
        title: 'Test Mission',
        description: 'Test Description'
      });
    });
    
    expect(result.current.missions).toHaveLength(1);
    expect(result.current.missions[0].title).toBe('Test Mission');
  });

  test('updates mission via WebSocket', () => {
    const { result } = renderHook(() => useMissions());
    
    act(() => {
      mockWebSocketMessage({
        type: 'mission_update',
        data: {
          action: 'updated',
          mission: updatedMission
        }
      });
    });
    
    expect(result.current.missions[0]).toEqual(updatedMission);
  });
});
```

```go
func TestCreateMission(t *testing.T) {
    server := setupTestServer()
    
    mission := &Mission{
        Title:       "Test Mission",
        Description: "Test Description",
        Status:      "planning",
        StartDate:   time.Now(),
        EndDate:     time.Now().Add(24 * time.Hour),
    }
    
    missionJSON, _ := json.Marshal(mission)
    req, _ := http.NewRequest("POST", "/api/v1/missions", bytes.NewBuffer(missionJSON))
    req.Header.Set("Content-Type", "application/json")
    req = req.WithContext(contextWithUser("user-123", "group-456"))
    
    rr := httptest.NewRecorder()
    server.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusCreated, rr.Code)
    
    var response Mission
    json.Unmarshal(rr.Body.Bytes(), &response)
    assert.Equal(t, "Test Mission", response.Title)
    assert.NotEmpty(t, response.ID)
}
```

## Acceptance Criteria

### Mission Planning
- [ ] Mission creation form with all required fields
- [ ] Mission templates for common operations
- [ ] Draft missions save automatically
- [ ] Mission overview dashboard displays key metrics
- [ ] Search and filter missions by status, priority, date

### Task Management
- [ ] Task creation with full CRUD operations
- [ ] Task assignment to personnel with notification
- [ ] Task dependencies and sequencing
- [ ] Progress tracking with visual indicators
- [ ] Subtask breakdown and hierarchical organization

### Resource Management
- [ ] Personnel roster with skills and availability
- [ ] Equipment inventory with status tracking
- [ ] Resource allocation conflict detection and warnings
- [ ] Resource scheduling timeline view
- [ ] Automated resource optimization suggestions

### Collaborative Features
- [ ] Real-time collaborative editing of missions
- [ ] Comments and annotations system
- [ ] Change tracking and version history
- [ ] Role-based permissions for editing
- [ ] Notification system for updates and changes

### Timeline and Scheduling
- [ ] Interactive Gantt chart for mission timeline
- [ ] Critical path analysis visualization
- [ ] Drag-and-drop timeline editing
- [ ] Milestone markers and dependency visualization
- [ ] Time zone support for global operations

## Dependencies

### Frontend Dependencies
```json
{
  "dependencies": {
    "@mui/x-date-pickers": "^6.19.0",
    "@mui/x-data-grid": "^6.19.0",
    "react-beautiful-dnd": "^13.1.1",
    "gantt-schedule-timeline-calendar": "^2.24.0",
    "date-fns": "^2.30.0"
  }
}
```

### Backend Dependencies
```go
require (
    github.com/lib/pq v1.10.9
    github.com/google/uuid v1.5.0
)
```

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 80%+ coverage
- [ ] Integration tests pass
- [ ] No security vulnerabilities
- [ ] Performance requirements met

### Functionality  
- [ ] All user stories completed and accepted
- [ ] Manual testing on desktop and mobile
- [ ] Error handling for all scenarios
- [ ] Real-time collaboration working
- [ ] Data persistence verified

### Documentation
- [ ] API documentation updated
- [ ] User interface documented
- [ ] Deployment guides updated
- [ ] Training materials created

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
