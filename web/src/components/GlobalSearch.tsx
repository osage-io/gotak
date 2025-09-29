/**
 * Global Search Component
 * Rich search engine with keyboard navigation for all pages
 */

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useRouter } from '../utils/router';
import { Icon } from '../components/ui/Icon';

interface SearchResult {
  id: string;
  type: 'page' | 'entity' | 'alert' | 'setting' | 'action' | 'message';
  title: string;
  description?: string;
  path?: string;
  iconName: string;
  metadata?: any;
  aliases?: string[];
}

interface SearchCategory {
  name: string;
  iconName: string;
  results: SearchResult[];
}

const GlobalSearch: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [categories, setCategories] = useState<SearchCategory[]>([]);
  
  const searchRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  // All searchable content
  const searchableContent: SearchResult[] = [
    // Pages
    { id: 'dashboard', type: 'page', title: 'Dashboard', description: 'Command overview and metrics', path: '/', iconName: 'dashboard', aliases: ['home', 'overview'] },
    { id: 'map', type: 'page', title: 'Tactical Map', description: 'Real-time entity positions', path: '/map', iconName: 'map', aliases: ['map view', 'positions'] },
    { id: 'comms', type: 'page', title: 'Communications', description: 'Chat and messaging', path: '/communications', iconName: 'chat', aliases: ['comms', 'chat', 'messages', 'messaging'] },
    { id: 'ai-intel', type: 'page', title: 'AI Intel Officer', description: 'Intelligence officer AI assistant', path: '/communications?ai=true', iconName: 'bot', aliases: ['ai', 'intelligence', 'intel', 'ai chat', 'ai officer', 'artificial intelligence', 'claude'] },
    { id: 'alerts', type: 'page', title: 'Alerts', description: 'System notifications', path: '/alerts', iconName: 'bell', aliases: ['notifications', 'warnings'] },
    { id: 'entities', type: 'page', title: 'Entities', description: 'Manage tactical entities', path: '/entities', iconName: 'users', aliases: ['units', 'forces', 'contacts'] },
    { id: 'routes', type: 'page', title: 'Routes', description: 'Route planning', path: '/routes', iconName: 'route', aliases: ['paths', 'navigation', 'waypoints'] },
    { id: 'settings', type: 'page', title: 'Settings', description: 'System configuration', path: '/settings', iconName: 'settings', aliases: ['config', 'configuration', 'preferences'] },
    
    // Quick Actions
    { id: 'ai-brief', type: 'action', title: 'Get Mission Briefing', description: 'Request AI mission briefing', iconName: 'bot', aliases: ['ai brief', 'intel brief', 'mission intel'] },
    { id: 'ai-threat', type: 'action', title: 'Get Threat Assessment', description: 'Request AI threat analysis', iconName: 'bot', aliases: ['threat assessment', 'threat intel'] },
    { id: 'emergency', type: 'action', title: 'Send Emergency Alert', description: 'Broadcast emergency message', iconName: 'alert' },
    { id: 'drone', type: 'action', title: 'Launch Drone', description: 'Deploy UAV for reconnaissance', iconName: 'rocket' },
    { id: 'sensor', type: 'action', title: 'Deploy Sensor', description: 'Activate sensor network', iconName: 'broadcast' },
    { id: 'broadcast', type: 'action', title: 'Broadcast Message', description: 'Send message to all units', iconName: 'broadcast' },
    { id: 'report', type: 'action', title: 'Generate Report', description: 'Create tactical report', iconName: 'chart' },
    
    // Common Entities
    { id: 'alpha1', type: 'entity', title: 'ALPHA-1', description: 'Friendly - Squad Leader', iconName: 'shield' },
    { id: 'bravo2', type: 'entity', title: 'BRAVO-2', description: 'Friendly - Rifleman', iconName: 'shield' },
    { id: 'eagle1', type: 'entity', title: 'EAGLE-EYE-1', description: 'Drone - ISR Platform', iconName: 'rocket' },
    
    // Settings shortcuts
    { id: 'server-config', type: 'setting', title: 'Server Configuration', description: 'TAK server settings', iconName: 'server' },
    { id: 'map-settings', type: 'setting', title: 'Map Settings', description: 'Map display options', iconName: 'map' },
    { id: 'security', type: 'setting', title: 'Security Settings', description: 'Authentication and encryption', iconName: 'lock' },
  ];

  // Global keyboard shortcuts and search shortcuts
  const [gPressed, setGPressed] = useState(false);
  
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isTyping = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';
      
      // Don't handle global shortcuts when typing in input fields
      if (isTyping) return;
      
      // Cmd/Ctrl + K shortcut to open search
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsOpen(true);
        return;
      }
      
      // Forward slash shortcut (only when not typing in an input)
      if (e.key === '/' && !isOpen) {
        e.preventDefault();
        setIsOpen(true);
        return;
      }
      
      // Escape to close
      if (e.key === 'Escape' && isOpen) {
        setIsOpen(false);
        return;
      }
      
      // Handle 'g' key for global navigation shortcuts
      if (e.key === 'g' || e.key === 'G') {
        if (!gPressed) {
          e.preventDefault();
          setGPressed(true);
          // Clear the 'g' state after 1 second if no second key is pressed
          setTimeout(() => setGPressed(false), 1000);
        }
        return;
      }
      
      // Handle global navigation shortcuts (g+key combinations)
      if (gPressed) {
        e.preventDefault();
        setGPressed(false);
        
        switch (e.key.toLowerCase()) {
          case 'm': // g+m = Maps
            router.navigate('/map');
            break;
          case 'd': // g+d = Dashboard
            router.navigate('/');
            break;
          case 'a': // g+a = Alerts
            router.navigate('/alerts');
            break;
          case 'e': // g+e = Entities
            router.navigate('/entities');
            break;
          case 'r': // g+r = Routes
            router.navigate('/routes');
            break;
          case 'i': // g+i = Integrations
            router.navigate('/integrations');
            break;
          case 's': // g+s = Settings
            router.navigate('/settings');
            break;
          case 'c': // g+c = Comms
            router.navigate('/communications');
            break;
          default:
            // Invalid combination, do nothing
            break;
        }
      }
    };
    
    const handleKeyUp = (e: KeyboardEvent) => {
      // Reset 'g' state if user releases any key (allows for quick typing)
      if (gPressed && e.key !== 'g' && e.key !== 'G') {
        setGPressed(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [isOpen, gPressed, router]);

  // Focus input when opened
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
      setQuery('');
      setSelectedIndex(0);
    }
  }, [isOpen]);

  // Click outside to close
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  // Search logic
  const performSearch = useCallback((searchQuery: string) => {
    if (!searchQuery.trim()) {
      setResults([]);
      setCategories([]);
      return;
    }

    const query = searchQuery.toLowerCase();
    const searchResults = searchableContent.filter(item => 
      item.title.toLowerCase().includes(query) ||
      item.description?.toLowerCase().includes(query) ||
      item.aliases?.some(alias => alias.toLowerCase().includes(query))
    );

    // Group by type
    const grouped = searchResults.reduce((acc, result) => {
      const category = result.type;
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(result);
      return acc;
    }, {} as Record<string, SearchResult[]>);

    // Create categories
    const categoryList: SearchCategory[] = [];
    
    if (grouped.page) {
      categoryList.push({ name: 'Pages', iconName: 'book', results: grouped.page });
    }
    if (grouped.action) {
      categoryList.push({ name: 'Actions', iconName: 'rocket', results: grouped.action });
    }
    if (grouped.entity) {
      categoryList.push({ name: 'Entities', iconName: 'users', results: grouped.entity });
    }
    if (grouped.setting) {
      categoryList.push({ name: 'Settings', iconName: 'settings', results: grouped.setting });
    }

    setCategories(categoryList);
    setResults(searchResults);
  }, []);

  // Handle search input
  const handleSearch = (value: string) => {
    setQuery(value);
    setSelectedIndex(0);
    performSearch(value);
  };

  // Navigate results with keyboard
  const handleKeyNavigation = (e: React.KeyboardEvent) => {
    if (!results.length) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(prev => (prev + 1) % results.length);
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(prev => (prev - 1 + results.length) % results.length);
        break;
      case 'Enter':
        e.preventDefault();
        if (results[selectedIndex]) {
          handleSelectResult(results[selectedIndex]);
        }
        break;
    }
  };

  // Handle result selection
  const handleSelectResult = (result: SearchResult) => {
    setIsOpen(false);
    
    switch (result.type) {
      case 'page':
        if (result.path) {
          router.navigate(result.path);
        }
        break;
      case 'entity':
        router.navigate('/entities');
        // Could also pass entity ID as state
        break;
      case 'setting':
        router.navigate('/settings');
        // Could also pass setting section as state
        break;
      case 'action':
        console.log('Execute action:', result.id);
        // Execute the action
        break;
    }
  };

  return (
    <>
      {/* Global 'g' Key Indicator */}
      {gPressed && (
        <div className="g-key-indicator">
          <div className="g-key-popup">
            <span className="g-key-title">Quick Navigation</span>
            <div className="g-key-shortcuts">
              <span><kbd>M</kbd> Maps</span>
              <span><kbd>D</kbd> Dashboard</span>
              <span><kbd>A</kbd> Alerts</span>
              <span><kbd>E</kbd> Entities</span>
              <span><kbd>R</kbd> Routes</span>
              <span><kbd>I</kbd> Integrations</span>
              <span><kbd>S</kbd> Settings</span>
              <span><kbd>C</kbd> Comms</span>
            </div>
          </div>
        </div>
      )}
      
      {/* Search Trigger Button */}
      <button
        className="search-trigger"
        onClick={() => setIsOpen(true)}
        title="Search (⌘K or /)"
      >
        <span className="search-text">Search</span>
        <span className="search-shortcut">/</span>
      </button>

      {/* Search Modal */}
      {isOpen && (
        <div className="search-overlay">
          <div className="search-modal" ref={searchRef}>
            <div className="search-header">
              <span className="search-modal-icon">
                <Icon name="search" size={20} color="var(--color-text-secondary)" />
              </span>
              <input
                ref={inputRef}
                type="text"
                className="search-input"
                placeholder="Search pages, entities, actions..."
                value={query}
                onChange={(e) => handleSearch(e.target.value)}
                onKeyDown={handleKeyNavigation}
              />
              <button 
                className="search-close"
                onClick={() => setIsOpen(false)}
              >
                ESC
              </button>
            </div>

            {/* Search Results */}
            {query && (
              <div className="search-results">
                {categories.length === 0 ? (
                  <div className="no-results">
                    <span className="no-results-icon">
                      <Icon name="search" size={32} color="var(--color-text-muted)" />
                    </span>
                    <p>No results found for "{query}"</p>
                  </div>
                ) : (
                  categories.map((category, categoryIndex) => (
                    <div key={category.name} className="result-category">
                      <div className="category-header">
                        <span className="category-icon">
                          <Icon name={category.iconName as any} size={14} color="var(--color-text-muted)" />
                        </span>
                        <span className="category-name">{category.name}</span>
                      </div>
                      
                      <div className="category-results">
                        {category.results.map((result, index) => {
                          const globalIndex = results.indexOf(result);
                          return (
                            <div
                              key={result.id}
                              className={`search-result ${selectedIndex === globalIndex ? 'selected' : ''}`}
                              onClick={() => handleSelectResult(result)}
                              onMouseEnter={() => setSelectedIndex(globalIndex)}
                            >
                              <span className="result-icon">
                                <Icon name={result.iconName as any} size={16} color="var(--color-text-secondary)" />
                              </span>
                              <div className="result-content">
                                <span className="result-title">{result.title}</span>
                                {result.description && (
                                  <span className="result-description">{result.description}</span>
                                )}
                              </div>
                              {selectedIndex === globalIndex && (
                                <span className="result-action">↵</span>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))
                )}
              </div>
            )}

            {/* Search Footer */}
            <div className="search-footer">
              <div className="search-hints">
                <span className="hint">
                  <kbd>↑↓</kbd> Navigate
                </span>
                <span className="hint">
                  <kbd>↵</kbd> Select
                </span>
                <span className="hint">
                  <kbd>ESC</kbd> Close
                </span>
                <span className="hint">
                  <kbd>G</kbd> <kbd>M/D/A/E/R/I/S/C</kbd> Go to
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      <style jsx>{`
        /* Search Trigger Button - Glass Morphism */
        .search-trigger {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 10px 24px;
          min-width: 400px;
          background: rgba(0, 0, 0, 0.3);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(255, 255, 255, 0.08);
          border-radius: 10px;
          color: var(--color-text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .search-trigger:hover {
          background: rgba(0, 0, 0, 0.4);
          border-color: rgba(255, 255, 255, 0.12);
          color: var(--color-accent);
          transform: translateY(-1px);
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
        }

        .search-text {
          flex: 1;
          text-align: left;
          font-size: 0.95rem;
          font-weight: 500;
        }

        .search-shortcut {
          padding: 2px 6px;
          background: rgba(0, 0, 0, 0.5);
          border-radius: 4px;
          font-size: 0.75rem;
          font-family: monospace;
          color: var(--color-text-muted);
        }

        /* Search Overlay */
        .search-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.8);
          backdrop-filter: blur(4px);
          z-index: 10000;
          display: flex;
          align-items: flex-start;
          justify-content: center;
          padding-top: 100px;
          animation: fadeIn 0.2s ease;
        }

        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }

        /* Search Modal */
        .search-modal {
          width: 600px;
          max-width: 90vw;
          max-height: 70vh;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.95) 0%, 
            rgba(15, 20, 25, 0.95) 100%);
          border: 1px solid rgba(0, 212, 170, 0.3);
          border-radius: 12px;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
          display: flex;
          flex-direction: column;
          animation: slideIn 0.3s ease;
        }

        @keyframes slideIn {
          from { 
            opacity: 0;
            transform: translateY(-20px);
          }
          to { 
            opacity: 1;
            transform: translateY(0);
          }
        }

        /* Search Header */
        .search-header {
          display: flex;
          align-items: center;
          padding: 20px;
          border-bottom: 1px solid rgba(0, 212, 170, 0.1);
        }

        .search-modal-icon {
          font-size: 1.5rem;
          margin-right: 12px;
        }

        .search-input {
          flex: 1;
          background: transparent;
          border: none;
          outline: none;
          font-size: 1.2rem;
          color: var(--color-text-primary);
        }

        .search-input::placeholder {
          color: var(--color-text-muted);
        }

        .search-close {
          padding: 4px 8px;
          background: rgba(255, 255, 255, 0.1);
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 4px;
          color: var(--color-text-secondary);
          font-size: 0.75rem;
          font-weight: 600;
          cursor: pointer;
        }

        /* Search Results */
        .search-results {
          flex: 1;
          overflow-y: auto;
          padding: 8px;
        }

        .no-results {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 40px;
          color: var(--color-text-muted);
        }

        .no-results-icon {
          font-size: 2rem;
          margin-bottom: 12px;
          opacity: 0.5;
        }

        /* Result Category */
        .result-category {
          margin-bottom: 16px;
        }

        .category-header {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          color: var(--color-text-muted);
          font-size: 0.8rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }

        .category-icon {
          font-size: 0.9rem;
        }

        /* Search Result Item */
        .search-result {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 12px 16px;
          margin: 2px 0;
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.15s ease;
        }

        .search-result:hover {
          background: rgba(0, 212, 170, 0.05);
        }

        .search-result.selected {
          background: rgba(0, 212, 170, 0.1);
          border: 1px solid rgba(0, 212, 170, 0.2);
        }

        .result-icon {
          font-size: 1.2rem;
          flex-shrink: 0;
        }

        .result-content {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .result-title {
          color: var(--color-text-primary);
          font-weight: 500;
        }

        .result-description {
          color: var(--color-text-muted);
          font-size: 0.85rem;
        }

        .result-action {
          padding: 2px 8px;
          background: rgba(0, 212, 170, 0.2);
          border-radius: 4px;
          color: var(--color-accent);
          font-size: 0.9rem;
        }

        /* Search Footer */
        .search-footer {
          padding: 12px 20px;
          border-top: 1px solid rgba(0, 212, 170, 0.1);
          background: rgba(0, 0, 0, 0.2);
        }

        .search-hints {
          display: flex;
          gap: 16px;
          font-size: 0.75rem;
          color: var(--color-text-muted);
        }

        .hint {
          display: flex;
          align-items: center;
          gap: 6px;
        }

        kbd {
          padding: 2px 6px;
          background: rgba(255, 255, 255, 0.1);
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 3px;
          font-family: monospace;
          font-size: 0.7rem;
        }

        /* G Key Indicator */
        .g-key-indicator {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          z-index: 9999;
          display: flex;
          align-items: center;
          justify-content: center;
          pointer-events: none;
        }
        
        .g-key-popup {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.95) 0%, 
            rgba(15, 20, 25, 0.95) 100%);
          border: 2px solid rgba(0, 212, 170, 0.4);
          border-radius: 12px;
          padding: 20px;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
          animation: popIn 0.2s ease;
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
        }
        
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.8);
          }
          to {
            opacity: 1;
            transform: scale(1);
          }
        }
        
        .g-key-title {
          display: block;
          color: var(--color-accent);
          font-weight: 600;
          font-size: 0.9rem;
          margin-bottom: 12px;
          text-align: center;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }
        
        .g-key-shortcuts {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 8px;
          min-width: 200px;
        }
        
        .g-key-shortcuts span {
          display: flex;
          align-items: center;
          gap: 8px;
          color: var(--color-text-secondary);
          font-size: 0.85rem;
          padding: 4px 0;
        }
        
        .g-key-shortcuts kbd {
          min-width: 20px;
          text-align: center;
          background: rgba(0, 212, 170, 0.2);
          border: 1px solid rgba(0, 212, 170, 0.3);
          color: var(--color-accent);
        }

        /* Responsive */
        @media (max-width: 768px) {
          .search-modal {
            width: 95vw;
            max-height: 80vh;
          }

          .search-trigger {
            padding: 8px 16px;
            min-width: unset;
            width: 100%;
          }

          .search-text {
            display: none;
          }
          
          .g-key-popup {
            margin: 0 16px;
            max-width: calc(100vw - 32px);
          }
          
          .g-key-shortcuts {
            grid-template-columns: 1fr;
            min-width: unset;
          }
        }
      `}</style>
    </>
  );
};

export default GlobalSearch;
