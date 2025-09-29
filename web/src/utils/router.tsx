/**
 * Simple Router System for GoTAK
 * Lightweight routing without external dependencies
 */

import React, { createContext, useContext, useState, useEffect, ReactNode, useCallback, useMemo } from 'react';

// Route definition
interface Route {
  path: string;
  component: React.ComponentType;
  title: string;
  icon?: string | React.ReactElement;
  hideFromNav?: boolean;
}

// Router context
interface RouterContext {
  currentRoute: string;
  navigate: (path: string) => void;
  routes: Route[];
  registerRoute: (route: Route) => void;
}

const RouterContext = createContext<RouterContext | null>(null);

// Router provider
interface RouterProviderProps {
  children: ReactNode;
  initialRoute?: string;
}

export const RouterProvider: React.FC<RouterProviderProps> = ({ 
  children, 
  initialRoute = '/' 
}) => {
  // Get the current path from the URL on initial load
  const getInitialPath = () => {
    const path = window.location.pathname;
    // If we're on a valid path, use it; otherwise use the provided initialRoute
    return path && path !== '/' ? path : initialRoute;
  };

  const [currentRoute, setCurrentRoute] = useState<string>(getInitialPath());
  const [routes, setRoutes] = useState<Route[]>([]);

  // Set up initial browser history state
  useEffect(() => {
    // Only push state if we're not already on the correct path
    if (window.location.pathname !== currentRoute) {
      window.history.replaceState({ path: currentRoute }, '', currentRoute);
    }
  }, []);

  // Handle browser navigation
  useEffect(() => {
    const handlePopState = (event: PopStateEvent) => {
      const path = event.state?.path || window.location.pathname || '/';
      setCurrentRoute(path);
    };

    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  // Navigation function - memoized for performance
  const navigate = useCallback((path: string) => {
    // Only navigate if we're not already on this path
    if (path !== currentRoute) {
      setCurrentRoute(path);
      window.history.pushState({ path }, '', path);
    }
  }, [currentRoute]);

  // Register route - memoized to prevent infinite re-renders
  const registerRoute = useCallback((route: Route) => {
    setRoutes(prev => {
      const exists = prev.find(r => r.path === route.path);
      if (exists) {
        return prev.map(r => r.path === route.path ? route : r);
      }
      return [...prev, route];
    });
  }, []);

  const contextValue: RouterContext = useMemo(() => ({
    currentRoute,
    navigate,
    routes,
    registerRoute,
  }), [currentRoute, navigate, routes, registerRoute]);

  return (
    <RouterContext.Provider value={contextValue}>
      {children}
    </RouterContext.Provider>
  );
};

// Router hook
export const useRouter = () => {
  const context = useContext(RouterContext);
  if (!context) {
    throw new Error('useRouter must be used within a RouterProvider');
  }
  
  // Extract params from the current route
  const params: Record<string, string> = {};
  const pathSegments = context.currentRoute.split('/');
  
  // Simple param extraction for /routes/{id} or /routes/{id}/navigate patterns
  if (pathSegments[1] === 'routes' && pathSegments[2] && pathSegments[2] !== 'view' && pathSegments[2] !== 'navigate') {
    params.id = pathSegments[2];
  }
  
  return {
    ...context,
    params
  };
};

// Route component
interface RouteProps {
  path: string;
  component: React.ComponentType;
  title: string;
  icon?: string | React.ReactElement;
  hideFromNav?: boolean;
}

export const Route: React.FC<RouteProps> = ({ path, component: Component, title, icon, hideFromNav }) => {
  const { registerRoute } = useRouter();

  useEffect(() => {
    registerRoute({ path, component: Component, title, icon, hideFromNav });
  }, [path, Component, title, icon, hideFromNav, registerRoute]);

  return null;
};

// Router outlet component
export const RouterOutlet: React.FC = () => {
  const { currentRoute, routes } = useRouter();
  
  const activeRoute = routes.find(route => {
    if (route.path === currentRoute) return true;
    
    // Handle dynamic routes for /routes/{id} and /routes/{id}/navigate
    if (currentRoute.match(/^\/routes\/[^/]+$/)) {
      return route.path === '/routes/view';
    }
    if (currentRoute.match(/^\/routes\/[^/]+\/navigate$/)) {
      return route.path === '/routes/navigate';
    }
    
    if (route.path !== '/' && currentRoute.startsWith(route.path)) return true;
    return false;
  });

  if (!activeRoute) {
    return (
      <div className="route-not-found">
        <div className="not-found-content">
          <h1 className="font-display font-bold text-3xl text-primary">404</h1>
          <p className="text-secondary">Route not found: {currentRoute}</p>
        </div>
      </div>
    );
  }

  const Component = activeRoute.component;
  return <Component />;
};

// Navigation component
interface NavigationProps {
  className?: string;
  onNavigate?: () => void; // Callback for when navigation occurs
}

export const Navigation: React.FC<NavigationProps> = ({ className, onNavigate }) => {
  const { routes, currentRoute, navigate } = useRouter();

  const handleNavigation = (path: string) => {
    navigate(path);
    onNavigate?.(); // Close side nav or perform other actions
  };

  return (
    <div className={`nav-menu ${className || ''}`}>
      {routes
        .filter(route => !route.hideFromNav) // Filter out routes that should be hidden from navigation
        .map((route) => (
          <div key={route.path} className={`nav-menu-item ${currentRoute === route.path ? 'active' : ''}`}>
            <button
              onClick={() => handleNavigation(route.path)}
              className="nav-menu-link"
              title={route.title}
            >
              <span className="nav-icon">{route.icon}</span>
              <span>{route.title}</span>
            </button>
            <div className="tooltip">{route.title}</div>
          </div>
        ))}
    </div>
  );
};

// Breadcrumb component
export const Breadcrumb: React.FC = () => {
  const { currentRoute, routes } = useRouter();
  const activeRoute = routes.find(r => r.path === currentRoute);

  if (!activeRoute) return null;

  return (
    <div className="breadcrumb">
      <span className="breadcrumb-icon">{activeRoute.icon}</span>
      <span className="breadcrumb-title">{activeRoute.title}</span>
    </div>
  );
};

export default RouterProvider;
