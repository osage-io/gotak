/**
 * GoTAK Main Application
 * Modern tactical awareness interface with professional typography and routing
 */

import React, { useEffect, useState } from 'react';
import { RouterProvider, RouterOutlet, Route, Navigation, Breadcrumb, useRouter } from './utils/router';
import { wsService } from './services/websocketService';
import './App.css';

// Import pages
import Dashboard from './pages/DashboardNew';
import Communications from './pages/Communications';
import TacticalMap from './pages/TacticalMap';
import Alerts from './pages/Alerts';
import Entities from './pages/Entities-new';
import Routes from './pages/Routes-new';
import RouteView from './pages/RouteView';
import RouteNavigate from './pages/RouteNavigate';
import Integrations from './pages/Integrations';
import Settings from './pages/Settings';
import Header from './components/layout/Header';
import Login from './pages/Login';
import { Icon } from './components/ui/Icon';

// Placeholder pages (will be built later)
const DocsPage = () => (
  <div className="page-placeholder">
    <div className="placeholder-content">
      <h1 className="font-display font-bold text-4xl text-primary mb-4">📖 Documentation</h1>
      <p className="text-secondary text-lg">System documentation and user guides...</p>
      <div className="placeholder-features">
        <div className="feature-item">• User Manual</div>
        <div className="feature-item">• API Documentation</div>
        <div className="feature-item">• System Architecture</div>
        <div className="feature-item">• Troubleshooting</div>
      </div>
    </div>
  </div>
);


// Protected route wrapper component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const router = useRouter();
  const isAuthenticated = localStorage.getItem('authToken');
  
  useEffect(() => {
    if (!isAuthenticated) {
      router.navigate('/login');
    }
  }, [isAuthenticated, router]);
  
  if (!isAuthenticated) {
    return null;
  }
  
  return <>{children}</>;
};

// Inner App component with routing
const AppContent: React.FC = () => {
  const [sideNavOpen, setSideNavOpen] = useState(false);
  const [sidebarVisible, setSidebarVisible] = useState(true); // Desktop sidebar visibility
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('authToken'));
  const router = useRouter();
  
  // Handle window resize to detect mobile/desktop
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth <= 768;
      setIsMobile(mobile);
      if (!mobile && sideNavOpen) {
        setSideNavOpen(false); // Close mobile nav when switching to desktop
      }
    };
    
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [sideNavOpen]);
  
  // Handle click outside to close sidenav
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (sideNavOpen && !target.closest('.side-nav') && !target.closest('.hamburger-btn')) {
        setSideNavOpen(false);
      }
    };
    
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [sideNavOpen]);
  
  // Handle escape key to close sidenav
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && sideNavOpen) {
        setSideNavOpen(false);
      }
    };
    
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [sideNavOpen]);
  
  // Initialize WebSocket connection on app start
  useEffect(() => {
    console.log('🚀 GoTAK Application Starting...');
    
    // Connect to WebSocket server
    const initializeConnection = async () => {
      try {
        await wsService.connect();
        console.log('✅ WebSocket connection established');
      } catch (error) {
        console.error('❌ Failed to connect to WebSocket:', error);
      }
    };

    initializeConnection();

    // Cleanup on unmount
    return () => {
      wsService.disconnect();
    };
  }, []);

  // Check authentication state and current route
  const currentPath = window.location.pathname;
  const isLoginPage = currentPath === '/login';
  
  // Update authentication state on storage changes
  useEffect(() => {
    const checkAuth = () => {
      const token = localStorage.getItem('authToken');
      setIsAuthenticated(!!token);
    };
    
    window.addEventListener('storage', checkAuth);
    // Also check when the component mounts
    checkAuth();
    
    return () => window.removeEventListener('storage', checkAuth);
  }, []);
  
  // If not authenticated or on login page without auth, show only login
  if (!isAuthenticated || isLoginPage) {
    if (!isAuthenticated && currentPath !== '/login') {
      router.navigate('/login');
    }
    return <RouterOutlet />;
  }
  
  return (
    <div className="app-container">
      {/* Side Navigation - Positioned First for Full Height */}
      <nav className={`side-nav ${isMobile ? `mobile ${sideNavOpen ? 'open' : ''}` : (sidebarVisible ? 'expanded' : 'hidden')}`}>
            <div className="side-nav-header">
              {/* GOTAK Branding as Home Button */}
              <button 
                className="nav-home-button"
                onClick={() => router.navigate('/')}
                title="Go to Dashboard"
              >
                <h1 className="nav-brand-title">GOTAK</h1>
                <span className="nav-brand-subtitle">Tactical Awareness Kit</span>
              </button>
            </div>
            
            <div className="side-nav-content">
              <Navigation 
                className="nav-menu" 
                onNavigate={() => setSideNavOpen(false)}
              />
            </div>
            
            <div className="side-nav-footer">
              {/* Show Logout button only when authenticated */}
              {isAuthenticated && (
                <div className="nav-menu">
                  <div className="nav-menu-item">
                    <button 
                      className="nav-menu-link"
                      onClick={(e) => {
                        e.preventDefault();
                        // Logout
                        localStorage.removeItem('authToken');
                        localStorage.removeItem('rememberUsername');
                        setIsAuthenticated(false);
                        router.navigate('/login');
                        window.location.href = '/login';
                      }}
                      title="Logout"
                    >
                      <span className="nav-icon"><Icon name="user" size={20} /></span>
                      <span>Logout</span>
                    </button>
                    <div className="tooltip">Logout</div>
                  </div>
                </div>
              )}
            </div>
          </nav>
          
          {/* Mobile overlay */}
          {isMobile && (
            <div 
              className={`mobile-overlay ${sideNavOpen ? 'visible' : ''}`} 
              onClick={() => setSideNavOpen(false)}
              aria-hidden="true" 
            />
          )}
          
          {/* Main Content Wrapper - Header + Content */}
          <div className={`main-wrapper ${!isMobile ? (sidebarVisible ? 'sidebar-expanded' : 'sidebar-hidden') : ''}`}>
            {/* Simple Clean Header */}
            <Header 
              onMenuToggle={() => isMobile ? setSideNavOpen(!sideNavOpen) : setSidebarVisible(!sidebarVisible)}
              menuOpen={isMobile ? sideNavOpen : sidebarVisible}
              isMobile={isMobile}
            />
            
            {/* Main Content Area */}
            <main className="main-content">
              <RouterOutlet />
            </main>
          </div>
      </div>
  );
};

// Main App component wrapper
const App: React.FC = () => {
  // Check if user is authenticated
  const isAuthenticated = !!localStorage.getItem('authToken');
  const initialRoute = isAuthenticated ? '/' : '/login';
  
  return (
    <RouterProvider initialRoute={initialRoute}>
      {/* Route definitions */}
      <Route path="/login" component={Login} title="Login" hideFromNav />
      <Route path="/" component={Dashboard} title="Dashboard" icon={<Icon name="dashboard" size={20} />} />
      <Route path="/map" component={TacticalMap} title="Map" icon={<Icon name="map" size={20} />} />
      <Route path="/communications" component={Communications} title="Comms" icon={<Icon name="chat" size={20} />} />
      <Route path="/alerts" component={Alerts} title="Alerts" icon={<Icon name="bell" size={20} />} />
      <Route path="/entities" component={Entities} title="Entities" icon={<Icon name="users" size={20} />} />
      <Route path="/routes" component={Routes} title="Routes" icon={<Icon name="route" size={20} />} />
      <Route path="/routes/view" component={RouteView} title="Route View" hideFromNav />
      <Route path="/routes/navigate" component={RouteNavigate} title="Route Navigate" hideFromNav />
      <Route path="/integrations" component={Integrations} title="Integrations" icon={<Icon name="link" size={20} />} />
      <Route path="/settings" component={Settings} title="Settings" icon={<Icon name="settings" size={20} />} />
      <Route path="/docs" component={DocsPage} title="Docs" icon={<Icon name="book" size={20} />} />
      
      <AppContent />
    </RouterProvider>
  );
};

export default App;
