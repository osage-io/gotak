import React, { useState } from 'react';
import './IconShowcase.css';
import { icons } from '../components/ui/Icon';

// Icon component for consistent rendering
const IconCard = ({ 
  name, 
  emoji, 
  category 
}: { 
  name: string; 
  emoji: string; 
  category: string;
}) => {
  const [copied, setCopied] = useState(false);
  const svg = icons[name as keyof typeof icons];

  const handleCopy = () => {
    const svgString = `<Icon name="${name}" />`;
    navigator.clipboard.writeText(svgString);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="icon-card" onClick={handleCopy}>
      <div className="icon-svg">
        {svg}
      </div>
      <div className="icon-emoji">{emoji}</div>
      <div className="icon-name">{name}</div>
      <div className="icon-category">{category}</div>
      {copied && <div className="icon-copied">Copied!</div>}
    </div>
  );
};

const iconData = [
  // Navigation & Location
  { name: 'target', emoji: '🎯', category: 'Navigation' },
  { name: 'pin', emoji: '📍', category: 'Navigation' },
  { name: 'map-pin', emoji: '📍', category: 'Navigation' },
  { name: 'map', emoji: '🗺️', category: 'Navigation' },
  { name: 'compass', emoji: '🧭', category: 'Navigation' },
  { name: 'navigation', emoji: '🧭', category: 'Navigation' },
  
  // Communication
  { name: 'chat', emoji: '💬', category: 'Communication' },
  { name: 'broadcast', emoji: '📡', category: 'Communication' },
  { name: 'signal', emoji: '📶', category: 'Communication' },
  { name: 'send', emoji: '➤', category: 'Communication' },
  { name: 'message-circle', emoji: '💬', category: 'Communication' },
  { name: 'radio', emoji: '📻', category: 'Communication' },
  { name: 'wifi', emoji: '📶', category: 'Communication' },
  
  // Security
  { name: 'lock', emoji: '🔒', category: 'Security' },
  { name: 'unlock', emoji: '🔓', category: 'Security' },
  { name: 'shield', emoji: '🛡️', category: 'Security' },
  { name: 'key', emoji: '🔑', category: 'Security' },
  
  // Actions & Status
  { name: 'rocket', emoji: '🚀', category: 'Actions' },
  { name: 'check', emoji: '✅', category: 'Status' },
  { name: 'check-circle', emoji: '✅', category: 'Status' },
  { name: 'cross', emoji: '❌', category: 'Status' },
  { name: 'x', emoji: '✖️', category: 'Status' },
  { name: 'warning', emoji: '⚠️', category: 'Status' },
  { name: 'alert', emoji: '⚡', category: 'Status' },
  { name: 'info', emoji: 'ℹ️', category: 'Status' },
  { name: 'error', emoji: '⚠️', category: 'Status' },
  { name: 'alert-circle', emoji: '⚠️', category: 'Status' },
  { name: 'alert-triangle', emoji: '⚠️', category: 'Status' },
  
  // Entity Types
  { name: 'user', emoji: '👤', category: 'Entities' },
  { name: 'users', emoji: '👥', category: 'Entities' },
  { name: 'soldier', emoji: '💂', category: 'Entities' },
  { name: 'medic', emoji: '⚕️', category: 'Entities' },
  { name: 'hostile', emoji: '⚔️', category: 'Entities' },
  { name: 'neutral', emoji: '🏳️', category: 'Entities' },
  { name: 'unknown', emoji: '❓', category: 'Entities' },
  { name: 'drone', emoji: '🚁', category: 'Entities' },
  { name: 'vehicle', emoji: '🚗', category: 'Entities' },
  { name: 'sensor', emoji: '📡', category: 'Entities' },
  { name: 'camera', emoji: '📷', category: 'Entities' },
  { name: 'radar', emoji: '📡', category: 'Entities' },
  { name: 'equipment', emoji: '🎒', category: 'Entities' },
  
  // View Types
  { name: 'grid', emoji: '⬛', category: 'Views' },
  { name: 'list', emoji: '☰', category: 'Views' },
  { name: 'tactical', emoji: '🗺️', category: 'Views' },
  { name: 'eye', emoji: '👁️', category: 'Views' },
  { name: 'eye-off', emoji: '👁️‍🗨️', category: 'Views' },
  
  // Status Indicators
  { name: 'battery', emoji: '🔋', category: 'Status' },
  { name: 'battery-low', emoji: '🪫', category: 'Status' },
  { name: 'speed', emoji: '⚡', category: 'Status' },
  { name: 'altitude', emoji: '⛰️', category: 'Status' },
  { name: 'thermometer', emoji: '🌡️', category: 'Status' },
  { name: 'droplet', emoji: '💧', category: 'Status' },
  
  // Data & Analytics
  { name: 'chart', emoji: '📊', category: 'Analytics' },
  { name: 'trending', emoji: '📈', category: 'Analytics' },
  { name: 'dashboard', emoji: '📋', category: 'Analytics' },
  
  // System & Settings
  { name: 'settings', emoji: '⚙️', category: 'System' },
  { name: 'sync', emoji: '🔄', category: 'System' },
  { name: 'world', emoji: '🌐', category: 'System' },
  { name: 'server', emoji: '🖥️', category: 'System' },
  { name: 'log-out', emoji: '🚪', category: 'System' },
  
  // Storage & Database
  { name: 'database', emoji: '💾', category: 'Storage' },
  { name: 'save', emoji: '💾', category: 'Storage' },
  { name: 'package', emoji: '📦', category: 'Storage' },
  
  // Tools & Operations
  { name: 'tools', emoji: '🔧', category: 'Tools' },
  { name: 'wrench', emoji: '🔧', category: 'Tools' },
  { name: 'search', emoji: '🔍', category: 'Tools' },
  { name: 'plus', emoji: '➕', category: 'Tools' },
  { name: 'trash-2', emoji: '🗑️', category: 'Tools' },
  
  // UI & Display
  { name: 'palette', emoji: '🎨', category: 'Display' },
  
  // Miscellaneous
  { name: 'bell', emoji: '🔔', category: 'Misc' },
  { name: 'route', emoji: '🛤️', category: 'Misc' },
  { name: 'link', emoji: '🔗', category: 'Misc' },
  { name: 'book', emoji: '📖', category: 'Misc' },
  { name: 'sparkle', emoji: '✨', category: 'Misc' },
  { name: 'inbox', emoji: '📥', category: 'Misc' },
  { name: 'bot', emoji: '🤖', category: 'Misc' }
];

const IconShowcase: React.FC = () => {
  const [selectedCategory, setSelectedCategory] = useState<string>('All');
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  const categories = ['All', ...Array.from(new Set(iconData.map(icon => icon.category)))];
  
  const filteredIcons = iconData.filter(icon => {
    const matchesCategory = selectedCategory === 'All' || icon.category === selectedCategory;
    const matchesSearch = icon.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                          icon.emoji.includes(searchTerm) ||
                          icon.category.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  return (
    <div className="icon-showcase">
      <div className="showcase-header">
        <h1>GoTAK Icon System</h1>
        <p>Clean, stencil-style icons to replace emojis throughout the application</p>
      </div>

      <div className="showcase-controls">
        <div className="search-box">
          <svg className="search-icon" viewBox="0 0 24 24" fill="none">
            <circle cx="10" cy="10" r="7" stroke="currentColor" strokeWidth="2"/>
            <path d="M15 15L21 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
          </svg>
          <input
            type="text"
            placeholder="Search icons..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>

        <div className="category-filters">
          {categories.map(category => (
            <button
              key={category}
              className={`category-btn ${selectedCategory === category ? 'active' : ''}`}
              onClick={() => setSelectedCategory(category)}
            >
              {category}
            </button>
          ))}
        </div>

        <div className="view-toggle">
          <button
            className={`view-btn ${viewMode === 'grid' ? 'active' : ''}`}
            onClick={() => setViewMode('grid')}
          >
            Grid
          </button>
          <button
            className={`view-btn ${viewMode === 'list' ? 'active' : ''}`}
            onClick={() => setViewMode('list')}
          >
            List
          </button>
        </div>
      </div>

      <div className="showcase-stats">
        <div className="stat">
          <span className="stat-value">{filteredIcons.length}</span>
          <span className="stat-label">Icons</span>
        </div>
        <div className="stat">
          <span className="stat-value">{categories.length - 1}</span>
          <span className="stat-label">Categories</span>
        </div>
        <div className="stat">
          <span className="stat-value">24px</span>
          <span className="stat-label">Base Size</span>
        </div>
      </div>

      <div className={`icons-container ${viewMode}`}>
        {filteredIcons.map((icon) => (
          <IconCard
            key={icon.name}
            name={icon.name}
            emoji={icon.emoji}
            category={icon.category}
          />
        ))}
      </div>

      <div className="implementation-guide">
        <h2>Implementation Guide</h2>
        <div className="guide-section">
          <h3>Usage</h3>
          <pre>
{`import { Icon } from '@/components/ui/Icon';

// Basic usage
<Icon name="target" />

// With size
<Icon name="shield" size={32} />

// With color
<Icon name="rocket" color="#00ff00" />

// With className
<Icon name="chat" className="custom-icon" />`}
          </pre>
        </div>

        <div className="guide-section">
          <h3>Design Principles</h3>
          <ul>
            <li>• 2px stroke width for consistency</li>
            <li>• No filled shapes, stencil style only</li>
            <li>• 24x24 viewBox for all icons</li>
            <li>• Uses currentColor for easy theming</li>
            <li>• Optimized paths for performance</li>
          </ul>
        </div>

        <div className="guide-section">
          <h3>Color System</h3>
          <div className="color-examples">
            <div className="color-example">
              <div className="icon-preview primary">{icons.rocket}</div>
              <span>Primary Teal</span>
            </div>
            <div className="color-example">
              <div className="icon-preview success">{icons['check-circle']}</div>
              <span>Success</span>
            </div>
            <div className="color-example">
              <div className="icon-preview info">{icons.info}</div>
              <span>Info</span>
            </div>
            <div className="color-example">
              <div className="icon-preview warning">{icons.warning}</div>
              <span>Warning</span>
            </div>
            <div className="color-example">
              <div className="icon-preview error">{icons.error}</div>
              <span>Error</span>
            </div>
            <div className="color-example">
              <div className="icon-preview danger">{icons.cross}</div>
              <span>Danger/Critical</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default IconShowcase;