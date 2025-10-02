# Sprint 3: Frontend Foundation & Core UI

**Duration:** 2 weeks  
**Sprint Goal:** Create React/TypeScript frontend foundation with authentication, routing, and basic tactical UI components

## Sprint Objectives

1. Set up modern React/TypeScript project with Vite
2. Implement authentication flow with OIDC and fallback support
3. Create responsive, mobile-first tactical UI design system
4. Build core navigation and layout components
5. Add WebSocket integration for real-time updates
6. Implement embedded static file serving in Go backend

## User Stories

### Epic 1: Frontend Project Setup & Tooling

**As a** Frontend Developer  
**I want** a modern, well-configured React development environment  
**So that** the team can build efficiently with proper tooling and standards

#### Story 1.1: Project Structure and Configuration
- [ ] Initialize React + TypeScript + Vite project in `web/` directory
- [ ] Configure ESLint, Prettier, and TypeScript for code quality
- [ ] Set up Husky for pre-commit hooks and automated testing
- [ ] Add path aliases and import organization
- [ ] Configure environment-specific builds and variables

#### Story 1.2: Build System Integration
- [ ] Create build process that outputs to `web/dist/`
- [ ] Embed static files in Go binary using `embed` package
- [ ] Set up development proxy to Go backend
- [ ] Configure hot reload for efficient development
- [ ] Add production build optimization and asset hashing

#### Story 1.3: UI Framework and Styling
- [ ] Set up Material-UI (MUI) with tactical/military theme
- [ ] Create custom theme with appropriate colors and typography
- [ ] Add responsive breakpoints for mobile, tablet, desktop
- [ ] Set up CSS-in-JS with emotion/styled-components
- [ ] Create reusable component library foundation

### Epic 2: Authentication Integration

**As a** Military Operator  
**I want** seamless authentication through the web interface  
**So that** I can access the tactical system securely from any device

#### Story 2.1: Authentication Context and State Management
- [ ] Create React Context for authentication state
- [ ] Implement Redux Toolkit for global state management
- [ ] Add authentication actions and reducers
- [ ] Create authentication hooks and utilities
- [ ] Handle token storage and automatic refresh

#### Story 2.2: Login Interface
- [ ] Create responsive login form with multiple auth methods
- [ ] Add OIDC login flow with redirect handling
- [ ] Implement fallback username/password login
- [ ] Add certificate-based authentication support
- [ ] Create loading states and error handling

#### Story 2.3: Protected Routes and Navigation
- [ ] Implement route protection with authentication checks
- [ ] Add role-based route access controls
- [ ] Create automatic redirect for unauthenticated users
- [ ] Handle session expiration gracefully
- [ ] Add logout functionality with session cleanup

#### Story 2.4: User Profile Management
- [ ] Create user profile display and editing interface
- [ ] Add password change functionality
- [ ] Display user roles and permissions
- [ ] Show session information and activity
- [ ] Add account settings and preferences

### Epic 3: Core Layout and Navigation

**As a** Tactical User  
**I want** intuitive navigation that works well on mobile and desktop  
**So that** I can efficiently access all system features

#### Story 3.1: Responsive Layout System
- [ ] Create main application layout with sidebar and header
- [ ] Implement mobile-first responsive design
- [ ] Add collapsible sidebar for mobile devices
- [ ] Create breadcrumb navigation
- [ ] Add consistent spacing and typography

#### Story 3.2: Navigation Components
- [ ] Create main navigation menu with role-based visibility
- [ ] Add quick action buttons and shortcuts
- [ ] Implement search functionality in header
- [ ] Add notification center and alerts
- [ ] Create user menu with profile and logout options

#### Story 3.3: Dashboard Layout
- [ ] Create dashboard home page with tactical overview
- [ ] Add widget system for customizable dashboards
- [ ] Display key metrics and status indicators
- [ ] Show recent activity and notifications
- [ ] Add quick access to common actions

### Epic 4: Real-Time WebSocket Integration

**As a** Operator  
**I want** real-time updates throughout the interface  
**So that** I have the most current tactical information

#### Story 4.1: WebSocket Connection Management
- [ ] Create WebSocket service with automatic reconnection
- [ ] Implement authentication for WebSocket connections
- [ ] Add connection status indicators throughout UI
- [ ] Handle connection failures and retries gracefully
- [ ] Add WebSocket event logging and debugging

#### Story 4.2: Real-Time Data Integration
- [ ] Connect WebSocket events to Redux state updates
- [ ] Add selective subscription management
- [ ] Implement optimistic updates with rollback
- [ ] Create real-time notification system
- [ ] Add conflict resolution for concurrent edits

#### Story 4.3: Live Status Indicators
- [ ] Show real-time connection status
- [ ] Display online/offline user indicators
- [ ] Add live data timestamps and freshness indicators
- [ ] Create real-time activity feeds
- [ ] Show system health and performance metrics

### Epic 5: Tactical UI Components Library

**As a** UI Developer  
**I want** reusable tactical-themed components  
**So that** the interface has consistent military styling and functionality

#### Story 5.1: Military-Themed Design System
- [ ] Create tactical color palette (olive, tan, blue, red)
- [ ] Design military-style iconography and symbols
- [ ] Add classification banners and labels
- [ ] Create tactical-style buttons, forms, and inputs
- [ ] Add military time display and formatting

#### Story 5.2: Data Display Components
- [ ] Create data tables with sorting and filtering
- [ ] Add tactical status indicators and badges
- [ ] Create timeline components for events
- [ ] Add progress indicators and completion meters
- [ ] Create expandable detail panels

#### Story 5.3: Form Components
- [ ] Build tactical-styled form inputs and validation
- [ ] Add classification level selectors
- [ ] Create date/time pickers with military formatting
- [ ] Add file upload with security scanning indicators
- [ ] Create multi-step form wizards

#### Story 5.4: Communication Components
- [ ] Create chat message components
- [ ] Add alert and notification components
- [ ] Build status update display components
- [ ] Create emergency alert styling
- [ ] Add message composition interfaces

## Technical Implementation

### Project Structure

```
web/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── auth/           # Authentication components
│   │   ├── common/         # Common UI elements
│   │   ├── forms/          # Form components
│   │   ├── layout/         # Layout components
│   │   └── tactical/       # Military-specific components
│   ├── pages/              # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Login.tsx
│   │   ├── Missions/
│   │   └── Profile.tsx
│   ├── hooks/              # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useWebSocket.ts
│   │   └── useApi.ts
│   ├── services/           # API and service layers
│   │   ├── api.ts          # HTTP API client
│   │   ├── websocket.ts    # WebSocket client
│   │   └── auth.ts         # Authentication service
│   ├── store/              # Redux store and slices
│   │   ├── authSlice.ts
│   │   ├── uiSlice.ts
│   │   └── index.ts
│   ├── types/              # TypeScript type definitions
│   │   ├── api.ts
│   │   ├── auth.ts
│   │   └── tactical.ts
│   ├── utils/              # Utility functions
│   │   ├── format.ts       # Formatting utilities
│   │   ├── validation.ts   # Form validation
│   │   └── constants.ts    # Application constants
│   ├── theme/              # MUI theme configuration
│   │   ├── index.ts
│   │   └── tactical.ts     # Military theme
│   └── App.tsx             # Root component
├── public/                 # Static assets
├── dist/                   # Build output
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js      # If using Tailwind CSS
```

### Authentication Flow

```typescript
// Authentication service
export class AuthService {
  private api: ApiClient;

  async loginWithOIDC(): Promise<AuthResult> {
    // Redirect to Vault OIDC provider
    const authUrl = await this.api.get('/auth/oidc/url');
    window.location.href = authUrl.url;
  }

  async loginWithCredentials(username: string, password: string): Promise<AuthResult> {
    const response = await this.api.post('/auth/login', { username, password });
    this.setTokens(response.tokens);
    return response;
  }

  async loginWithCertificate(): Promise<AuthResult> {
    // Use client certificates for authentication
    const response = await this.api.post('/auth/cert');
    this.setTokens(response.tokens);
    return response;
  }

  private setTokens(tokens: TokenPair): void {
    localStorage.setItem('access_token', tokens.access);
    localStorage.setItem('refresh_token', tokens.refresh);
  }
}

// Authentication hook
export function useAuth() {
  const dispatch = useAppDispatch();
  const auth = useAppSelector(state => state.auth);

  const login = useCallback(async (method: AuthMethod, credentials?: any) => {
    dispatch(loginStart());
    try {
      const result = await authService.login(method, credentials);
      dispatch(loginSuccess(result));
    } catch (error) {
      dispatch(loginFailure(error.message));
    }
  }, [dispatch]);

  return { ...auth, login };
}
```

### WebSocket Integration

```typescript
// WebSocket service
export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private subscriptions = new Map<string, (data: any) => void>();

  connect(token: string): void {
    this.ws = new WebSocket(`wss://${window.location.host}/ws?token=${token}`);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onclose = () => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => this.reconnect(), 1000 * Math.pow(2, this.reconnectAttempts));
        this.reconnectAttempts++;
      }
    };
  }

  subscribe(channel: string, callback: (data: any) => void): void {
    this.subscriptions.set(channel, callback);
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'subscribe', channel }));
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    const callback = this.subscriptions.get(message.type);
    if (callback) {
      callback(message.data);
    }
  }
}

// WebSocket hook
export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const wsRef = useRef<WebSocketService | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token && !wsRef.current) {
      wsRef.current = new WebSocketService();
      wsRef.current.connect(token);
    }
  }, []);

  const subscribe = useCallback((channel: string, callback: (data: any) => void) => {
    wsRef.current?.subscribe(channel, callback);
  }, []);

  return { isConnected, lastMessage, subscribe };
}
```

### Tactical Theme Configuration

```typescript
// Material-UI theme for tactical interface
export const tacticalTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#4CAF50',     // Military green
      dark: '#388E3C',
      light: '#81C784',
    },
    secondary: {
      main: '#FF9800',     // Amber for warnings
      dark: '#F57C00',
      light: '#FFB74D',
    },
    error: {
      main: '#F44336',     // Red for alerts
    },
    warning: {
      main: '#FF9800',     // Orange for cautions
    },
    info: {
      main: '#2196F3',     // Blue for information
    },
    success: {
      main: '#4CAF50',     // Green for success
    },
    background: {
      default: '#121212',   // Dark background
      paper: '#1E1E1E',     // Card backgrounds
    },
    text: {
      primary: '#E0E0E0',   // Light text
      secondary: '#B0B0B0', // Secondary text
    },
  },
  typography: {
    fontFamily: '"Roboto Mono", "Courier New", monospace',
    h1: { fontSize: '2rem', fontWeight: 600 },
    h2: { fontSize: '1.75rem', fontWeight: 600 },
    h3: { fontSize: '1.5rem', fontWeight: 600 },
    body1: { fontSize: '1rem', lineHeight: 1.5 },
    body2: { fontSize: '0.875rem', lineHeight: 1.43 },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          textTransform: 'uppercase',
          fontWeight: 600,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          border: '1px solid #333',
        },
      },
    },
  },
});

// Classification banner component
export function ClassificationBanner({ level }: { level: string }) {
  const colors = {
    UNCLASSIFIED: '#4CAF50',
    RESTRICTED: '#FF9800',
    SECRET: '#F44336',
    TOPSECRET: '#9C27B0',
  };

  return (
    <Box
      sx={{
        backgroundColor: colors[level] || colors.UNCLASSIFIED,
        color: 'white',
        padding: '4px 8px',
        textAlign: 'center',
        fontWeight: 'bold',
        fontSize: '0.75rem',
        letterSpacing: '1px',
      }}
    >
      {level}
    </Box>
  );
}
```

## Go Backend Integration

### Static File Embedding

```go
// Embed frontend build files
//go:embed web/dist/*
var staticFiles embed.FS

// Serve static files
func (s *Server) setupStaticRoutes() {
    // Serve embedded static files
    staticFS, err := fs.Sub(staticFiles, "web/dist")
    if err != nil {
        log.Fatal(err)
    }
    
    s.router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
    
    // Serve index.html for SPA routes
    s.router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
        // Check if file exists in static files
        if file, err := staticFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
            file.Close()
            http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
            return
        }
        
        // Serve index.html for SPA routes
        indexFile, err := staticFS.Open("index.html")
        if err != nil {
            http.Error(w, "Not found", 404)
            return
        }
        defer indexFile.Close()
        
        w.Header().Set("Content-Type", "text/html")
        io.Copy(w, indexFile)
    })
}
```

## Acceptance Criteria

### Frontend Setup
- [ ] React app builds and runs in development mode
- [ ] Production build creates optimized bundles under 2MB
- [ ] Static files are properly embedded in Go binary
- [ ] Hot reload works in development environment
- [ ] TypeScript compilation has zero errors

### Authentication
- [ ] Users can log in with OIDC, certificates, or username/password
- [ ] Authentication state persists across browser refreshes
- [ ] Token refresh works automatically before expiration
- [ ] Protected routes redirect unauthenticated users
- [ ] Logout clears all authentication data

### Responsive Design
- [ ] Interface works on mobile devices (320px width minimum)
- [ ] Tablet layout provides optimal user experience
- [ ] Desktop interface utilizes full screen real estate
- [ ] Navigation collapses appropriately on small screens
- [ ] All interactive elements are touch-friendly (44px minimum)

### WebSocket Integration
- [ ] Real-time connection established on login
- [ ] Connection status is visible to users
- [ ] Automatic reconnection works after network interruptions
- [ ] Real-time updates appear immediately in UI
- [ ] No memory leaks from WebSocket connections

### UI Components
- [ ] All components follow tactical theme consistently
- [ ] Classification levels are displayed prominently
- [ ] Loading states provide appropriate user feedback
- [ ] Error handling displays user-friendly messages
- [ ] Components are accessible (WCAG 2.1 AA compliance)

## Development Tasks

### Week 1: Foundation & Authentication
- [ ] Set up React/TypeScript/Vite project structure
- [ ] Configure build system and static file embedding
- [ ] Implement authentication flows and state management
- [ ] Create responsive layout and navigation
- [ ] Add tactical theme and base UI components

### Week 2: WebSocket & Advanced UI
- [ ] Integrate WebSocket real-time functionality
- [ ] Build tactical UI component library
- [ ] Add comprehensive error handling and loading states
- [ ] Create user profile and settings interfaces
- [ ] Add comprehensive testing and documentation

## Testing Strategy

### Unit Tests (Jest + React Testing Library)
- [ ] Authentication service and hooks
- [ ] WebSocket service functionality
- [ ] UI component rendering and interactions
- [ ] Form validation and submission
- [ ] Utility functions and helpers

### Integration Tests
- [ ] Authentication flow end-to-end
- [ ] WebSocket connection and messaging
- [ ] API integration with error handling
- [ ] Responsive design across breakpoints
- [ ] Cross-browser compatibility

### Performance Tests
- [ ] Bundle size analysis and optimization
- [ ] Runtime performance profiling
- [ ] Memory leak detection
- [ ] WebSocket connection scalability
- [ ] Accessibility compliance testing

## Dependencies

### Frontend Packages

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@mui/material": "^5.14.0",
    "@mui/icons-material": "^5.14.0",
    "@reduxjs/toolkit": "^1.9.0",
    "react-redux": "^8.1.0",
    "react-router-dom": "^6.15.0",
    "axios": "^1.5.0",
    "@emotion/react": "^11.11.0",
    "@emotion/styled": "^11.11.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "@vitejs/plugin-react": "^4.0.0",
    "eslint": "^8.45.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "prettier": "^3.0.0",
    "typescript": "^5.0.0",
    "vite": "^4.4.0",
    "@testing-library/react": "^13.4.0",
    "@testing-library/jest-dom": "^6.0.0",
    "jest": "^29.6.0"
  }
}
```

### Go Backend Updates
- Add `embed` package for static files
- Update router to serve SPA routes
- Add WebSocket authentication middleware
- Configure CORS for development

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
