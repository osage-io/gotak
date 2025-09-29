# GoTAK Web Interface

Modern tactical awareness web interface built with React, TypeScript, and Vite. Features advanced search capabilities, keyboard shortcuts, and real-time tactical data visualization.

## Features

- **Global Search**: Intelligent search across all pages, entities, and actions
- **Keyboard Shortcuts**: Full keyboard navigation and command palette
- **Real-time Data**: Live tactical updates via WebSocket connection
- **Responsive Design**: Mobile and desktop optimized interface
- **Modern UI**: Glass morphism design with tactical theme

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## Keyboard Shortcuts

### Global Navigation
- `Ctrl/Cmd + K` - Open command palette
- `/` - Focus quick search
- `Ctrl + 1-9` - Navigate to pages:
  - `Ctrl + 1` - Dashboard
  - `Ctrl + 2` - Tactical Map
  - `Ctrl + 3` - Communications
  - `Ctrl + 4` - Alerts
  - `Ctrl + 5` - Entities
  - `Ctrl + 6` - Routes
  - `Ctrl + 7` - Settings

### Quick Actions
- `Ctrl + E` - Emergency alert
- `Ctrl + I` - AI Intel Officer
- `Ctrl + L` - Alerts
- `Ctrl + G` - Settings

### Search Navigation
- `↑↓` - Navigate results
- `Enter` - Select result
- `Escape` - Close search

## Search Capabilities

The global search system can find:

- **Pages**: All application pages with descriptions and keywords
- **Commands**: Quick actions and system commands
- **AI Actions**: AI Intel Officer commands
- **Entities**: Tactical entities and units
- **Settings**: Configuration sections

Search supports:
- Real-time filtering as you type
- Keyword and alias matching
- Categorized results
- Keyboard navigation
- Smart ranking based on relevance

## Development

### Architecture

- **Components**: Modular React components in `/src/components/`
- **Pages**: Full-page components in `/src/pages/`
- **Layout**: Header and navigation components in `/src/components/layout/`
- **Router**: Custom router implementation in `/src/utils/router.tsx`
- **Services**: WebSocket and API services in `/src/services/`

### Adding New Search Items

To add new searchable items to the global search:

1. **Edit Header Component**: Update `searchableItems` array in `/src/components/layout/Header.tsx`
2. **Add Search Item**:
   ```typescript
   {
     type: 'page' | 'command' | 'ai' | 'entity' | 'setting',
     title: 'Display Name',
     description: 'Short description',
     path: '/route/path', // for pages
     icon: 'icon-name',
     keywords: ['searchable', 'terms'],
     aliases: ['alternative', 'names'],
     shortcut: '1', // for Ctrl+N shortcuts
     action: () => console.log('Custom action') // for commands
   }
   ```

### Keyboard Shortcut System

The Header component handles all global keyboard shortcuts through:
- Event listeners on `window.addEventListener('keydown')`
- Smart conflict avoidance (shortcuts disabled in input fields)
- Cross-platform support (Cmd on Mac, Ctrl on Windows/Linux)

### Customizing Search

The search system can be extended by:
1. Adding new search types in the Header component
2. Implementing custom search logic in `performSearch()` function
3. Adding new result categories in the search modal

### Note on SimpleHeader Removal

**Important**: The `SimpleHeader` component has been removed in favor of the enhanced `Header` component with search functionality. Do not re-introduce `SimpleHeader` as it lacks the advanced search capabilities.
