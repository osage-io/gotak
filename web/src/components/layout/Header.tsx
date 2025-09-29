/**
 * Enhanced Header Component
 * Tactical command center header with intelligent search and quick access
 */

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useRouter } from '../../utils/router';
import { Icon } from '../ui/Icon';
import './Header.css';

interface QuickAction {
  id: string;
  title: string;
  icon: string;
  shortcut: string;
  action: () => void;
  color?: string;
}

interface HeaderProps {
  onMenuToggle?: () => void;
  menuOpen?: boolean;
  isMobile?: boolean;
}

const Header: React.FC<HeaderProps> = ({ onMenuToggle, menuOpen, isMobile }) => {
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [currentTime, setCurrentTime] = useState(new Date());
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'connecting'>('connected');
  const [commandMode, setCommandMode] = useState(false);
  const [gKeyPressed, setGKeyPressed] = useState(false);
  
  const searchInputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  // Update time every second
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // Quick actions with keyboard shortcuts (available in search, not in header)
  const quickActions: QuickAction[] = [
    {
      id: 'ai-officer',
      title: 'AI Officer',
      icon: 'bot',
      shortcut: 'O',
      action: () => {
        router.navigate('/communications?ai=true');
      },
      color: 'var(--color-accent)'
    }
  ];

  // Searchable pages and commands
  const searchableItems = [
    // Main Pages
    { type: 'page', title: 'Dashboard', path: '/', icon: 'dashboard', keywords: ['home', 'overview', 'metrics'], shortcut: 'd' },
    { type: 'page', title: 'Tactical Map', path: '/map', icon: 'map', keywords: ['map', 'positions', 'tracking'], shortcut: 'm' },
    { type: 'page', title: 'Communications', path: '/communications', icon: 'chat', keywords: ['chat', 'messages', 'comms'], shortcut: 'c' },
    { type: 'page', title: 'Alerts', path: '/alerts', icon: 'bell', keywords: ['notifications', 'warnings'], shortcut: 'a' },
    { type: 'page', title: 'Entities', path: '/entities', icon: 'users', keywords: ['units', 'forces', 'contacts'], shortcut: 'e' },
    { type: 'page', title: 'Routes', path: '/routes', icon: 'route', keywords: ['navigation', 'waypoints', 'paths'], shortcut: 'r' },
    { type: 'page', title: 'Settings', path: '/settings', icon: 'settings', keywords: ['config', 'preferences'], shortcut: 's' },
    
    // Commands
    { type: 'command', title: 'Open AI Intel Officer', icon: 'bot', keywords: ['ai', 'claude', 'intelligence'], action: () => router.navigate('/communications?ai=true') },
    { type: 'command', title: 'Create New Route', icon: 'route', keywords: ['route', 'plan'], action: () => router.navigate('/routes?action=new') },
    { type: 'command', title: 'Send Emergency Alert', icon: 'alert', keywords: ['emergency', '911', 'sos'], action: () => console.log('Emergency!') },
    { type: 'command', title: 'Toggle Dark Mode', icon: 'palette', keywords: ['theme', 'dark', 'light'], action: () => console.log('Toggle theme') },
    { type: 'command', title: 'Export Data', icon: 'download', keywords: ['export', 'download', 'save'], action: () => console.log('Export') },
    { type: 'command', title: 'View Documentation', icon: 'book', keywords: ['docs', 'help', 'manual'], action: () => router.navigate('/docs') },
    
    // AI Commands
    { type: 'ai', title: 'Get Mission Briefing', icon: 'bot', keywords: ['brief', 'mission', 'intel'], action: () => console.log('AI Brief') },
    { type: 'ai', title: 'Analyze Threat Level', icon: 'bot', keywords: ['threat', 'risk', 'danger'], action: () => console.log('AI Threat') },
    { type: 'ai', title: 'Optimize Route', icon: 'bot', keywords: ['route', 'optimize', 'best'], action: () => console.log('AI Route') },
    { type: 'ai', title: 'Predict Enemy Movement', icon: 'bot', keywords: ['predict', 'enemy', 'movement'], action: () => console.log('AI Predict') },
  ];

  // Global keyboard shortcuts with g+letter pattern
  useEffect(() => {
    let gTimeout: NodeJS.Timeout;
    
    const handleKeyDown = (e: KeyboardEvent) => {
      // Command palette (Cmd/Ctrl + K)
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearchOpen(true);
        setCommandMode(true);
      }
      
      // Quick search (/)
      if (e.key === '/' && !searchOpen) {
        const target = e.target as HTMLElement;
        if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
          e.preventDefault();
          setSearchOpen(true);
          setCommandMode(false);
        }
      }
      
      // Letter shortcuts for pages (g + m/d/a/e/r/i/s/c)
      if (gKeyPressed && !searchOpen) {
        const letterShortcuts: Record<string, string> = {
          'm': '/map',        // g+m = Maps
          'd': '/',           // g+d = Dashboard
          'a': '/alerts',     // g+a = Alerts
          'e': '/entities',   // g+e = Entities
          'r': '/routes',     // g+r = Routes
          'i': '/integrations', // g+i = Integrations
          's': '/settings',   // g+s = Settings
          'c': '/communications' // g+c = Comms
        };
        
        const key = e.key.toLowerCase();
        if (letterShortcuts[key]) {
          e.preventDefault();
          router.navigate(letterShortcuts[key]);
          setGKeyPressed(false);
        }
      }
      
      // Handle 'g' key press (start of vim-style shortcuts)
      if (e.key === 'g' && !searchOpen && !gKeyPressed) {
        const target = e.target as HTMLElement;
        if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
          e.preventDefault();
          setGKeyPressed(true);
          // Reset g key press after 2 seconds
          gTimeout = setTimeout(() => setGKeyPressed(false), 2000);
        }
      }
      
      // Special quick action shortcut (g + o for AI Officer)
      if (gKeyPressed && !searchOpen && e.key.toLowerCase() === 'o') {
        const action = quickActions.find(a => a.shortcut.toLowerCase() === 'o');
        if (action) {
          e.preventDefault();
          action.action();
          setGKeyPressed(false);
        }
      }
      
      // Escape to close search or reset g key
      if (e.key === 'Escape') {
        if (searchOpen) {
          setSearchOpen(false);
          setSearchQuery('');
          setCommandMode(false);
        }
        if (gKeyPressed) {
          setGKeyPressed(false);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      if (gTimeout) clearTimeout(gTimeout);
    };
  }, [searchOpen, gKeyPressed, router]);

  // Focus search input when opened
  useEffect(() => {
    if (searchOpen && searchInputRef.current) {
      searchInputRef.current.focus();
      setSelectedIndex(0);
    }
  }, [searchOpen]);

  // Search functionality
  const performSearch = useCallback((query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }

    const lowerQuery = query.toLowerCase();
    const results = searchableItems.filter(item => {
      return item.title.toLowerCase().includes(lowerQuery) ||
             item.keywords.some(k => k.toLowerCase().includes(lowerQuery));
    });

    setSearchResults(results);
  }, []);

  // Handle search input
  const handleSearchInput = (value: string) => {
    setSearchQuery(value);
    performSearch(value);
    setSelectedIndex(0);
  };

  // Navigate search results
  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(prev => (prev + 1) % searchResults.length);
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(prev => (prev - 1 + searchResults.length) % searchResults.length);
        break;
      case 'Enter':
        e.preventDefault();
        if (searchResults[selectedIndex]) {
          executeSearchResult(searchResults[selectedIndex]);
        }
        break;
    }
  };

  // Execute search result
  const executeSearchResult = (result: any) => {
    setSearchOpen(false);
    setSearchQuery('');
    setCommandMode(false);
    
    if (result.type === 'page' && result.path) {
      router.navigate(result.path);
    } else if (result.action) {
      result.action();
    }
  };

  return (
    <>
      <header className="enhanced-header">
        <div className="header-section header-left">
          {/* Menu Toggle */}
          <button
            className="header-menu-toggle"
            onClick={onMenuToggle}
            aria-label="Toggle menu"
          >
            <Icon name={menuOpen ? 'x' : 'menu'} size={20} />
          </button>

          {/* Brand - Only on Desktop */}
          {!isMobile && (
            <div className="header-brand">
              <span className="brand-text">GOTAK</span>
              <span className="brand-badge">v2.0</span>
            </div>
          )}
        </div>

        <div className="header-section header-center">
          {/* Enhanced Search Bar */}
          <button
            className="header-search-trigger"
            onClick={() => {
              setSearchOpen(true);
              setCommandMode(false);
            }}
          >
            <Icon name="search" size={18} />
            <span className="search-placeholder">Search or type command...</span>
            <div className="search-shortcuts">
              <kbd>/</kbd>
              <span className="shortcut-separator">or</span>
              <kbd>⌘K</kbd>
            </div>
          </button>
        </div>

        <div className="header-section header-right">
          {/* Quick Actions - Hidden from header, available in search */}

          {/* Status Indicators */}
          <div className="header-status">
            <div className="status-item">
              <span className={`status-indicator ${connectionStatus}`}></span>
              <span className="status-label">{connectionStatus.toUpperCase()}</span>
            </div>
            <div className="status-item status-time">
              <span className="time-label">ZULU</span>
              <span className="time-value">{currentTime.toISOString().slice(11, 19)}Z</span>
            </div>
          </div>
        </div>
      </header>

      {/* Enhanced Search Overlay */}
      {searchOpen && (
        <div className="search-overlay" onClick={() => setSearchOpen(false)}>
          <div className="search-modal" onClick={(e) => e.stopPropagation()}>
            <div className="search-modal-header">
              <Icon name="search" size={20} />
              <input
                ref={searchInputRef}
                type="text"
                className="search-modal-input"
                placeholder={commandMode ? "Type a command..." : "Search pages, actions, or type '>' for commands"}
                value={searchQuery}
                onChange={(e) => handleSearchInput(e.target.value)}
                onKeyDown={handleSearchKeyDown}
              />
              <button className="search-close-btn" onClick={() => setSearchOpen(false)}>
                <kbd>ESC</kbd>
              </button>
            </div>

            {searchQuery && searchResults.length > 0 && (
              <div className="search-results">
                {/* Group results by type */}
                {['page', 'command', 'ai'].map(type => {
                  const typeResults = searchResults.filter(r => r.type === type);
                  if (typeResults.length === 0) return null;
                  
                  return (
                    <div key={type} className="search-group">
                      <div className="search-group-header">
                        {type === 'page' && 'Pages'}
                        {type === 'command' && 'Commands'}
                        {type === 'ai' && 'AI Actions'}
                      </div>
                      {typeResults.map((result, index) => {
                        const globalIndex = searchResults.indexOf(result);
                        return (
                          <button
                            key={`${result.type}-${result.title}`}
                            className={`search-result-item ${selectedIndex === globalIndex ? 'selected' : ''}`}
                            onClick={() => executeSearchResult(result)}
                            onMouseEnter={() => setSelectedIndex(globalIndex)}
                          >
                            <Icon name={result.icon as any} size={18} />
                            <span className="search-result-title">{result.title}</span>
                            {result.shortcut && (
                              <kbd className="search-result-shortcut">g{result.shortcut.toLowerCase()}</kbd>
                            )}
                            {selectedIndex === globalIndex && (
                              <span className="search-result-action">↵</span>
                            )}
                          </button>
                        );
                      })}
                    </div>
                  );
                })}
              </div>
            )}

            {searchQuery && searchResults.length === 0 && (
              <div className="search-empty">
                <Icon name="search" size={32} />
                <p>No results for "{searchQuery}"</p>
                <span>Try different keywords or commands</span>
              </div>
            )}

            {!searchQuery && (
              <div className="search-help">
                <div className="search-help-section">
                  <h4>Quick Navigation (g + letter)</h4>
                  <div className="search-help-items">
                    <div className="help-item">
                      <kbd>g</kbd><kbd>m</kbd> Maps
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>d</kbd> Dashboard
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>a</kbd> Alerts
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>e</kbd> Entities
                    </div>
                  </div>
                </div>
                
                <div className="search-help-section">
                  <h4>More Navigation</h4>
                  <div className="search-help-items">
                    <div className="help-item">
                      <kbd>g</kbd><kbd>r</kbd> Routes
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>i</kbd> Integrations
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>s</kbd> Settings
                    </div>
                    <div className="help-item">
                      <kbd>g</kbd><kbd>c</kbd> Comms
                    </div>
                  </div>
                </div>
              </div>
            )}

            <div className="search-modal-footer">
              <div className="search-footer-hints">
                <span><kbd>↑↓</kbd> Navigate</span>
                <span><kbd>↵</kbd> Select</span>
                <span><kbd>ESC</kbd> Close</span>
                <span><kbd>/</kbd> Search</span>
                <span><kbd>⌘K</kbd> Commands</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Shortcut Popup when 'g' is pressed */}
      {gKeyPressed && (
        <div className="shortcut-popup">
          <div className="shortcut-popup-content">
            <div className="shortcut-popup-header">
              <span className="shortcut-popup-title">Quick Navigation</span>
              <span className="shortcut-popup-hint">Press a key to navigate</span>
            </div>
            <div className="shortcut-popup-grid">
              <div className="shortcut-item">
                <kbd>m</kbd>
                <span>Maps</span>
              </div>
              <div className="shortcut-item">
                <kbd>d</kbd>
                <span>Dashboard</span>
              </div>
              <div className="shortcut-item">
                <kbd>a</kbd>
                <span>Alerts</span>
              </div>
              <div className="shortcut-item">
                <kbd>e</kbd>
                <span>Entities</span>
              </div>
              <div className="shortcut-item">
                <kbd>r</kbd>
                <span>Routes</span>
              </div>
              <div className="shortcut-item">
                <kbd>i</kbd>
                <span>Integrations</span>
              </div>
              <div className="shortcut-item">
                <kbd>s</kbd>
                <span>Settings</span>
              </div>
              <div className="shortcut-item">
                <kbd>c</kbd>
                <span>Comms</span>
              </div>
              <div className="shortcut-item">
                <kbd>o</kbd>
                <span>AI Officer</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default Header;
