# GoTAK Keyboard Shortcuts

## Global Navigation Shortcuts

Press `g` followed by one of these keys to quickly navigate between pages:

| Shortcut | Page | Description |
|----------|------|-------------|
| `g` + `m` | **Maps** | Tactical Map view |
| `g` + `d` | **Dashboard** | Command overview and metrics |
| `g` + `a` | **Alerts** | System notifications and warnings |
| `g` + `e` | **Entities** | Manage tactical entities and units |
| `g` + `r` | **Routes** | Route planning and navigation |
| `g` + `i` | **Integrations** | External system integrations |
| `g` + `s` | **Settings** | System configuration |
| `g` + `c` | **Comms** | Communications and chat |
| `g` + `o` | **AI Officer** | Open AI Intel Officer |

## Search Shortcuts

| Shortcut | Action |
|----------|---------|
| `/` | Open global search |
| `Cmd/Ctrl` + `K` | Open global search |
| `↑` `↓` | Navigate search results |
| `Enter` | Select search result |
| `Escape` | Close search |

## How Global Navigation Works

1. **Press `g`** - A quick navigation popup will appear showing available shortcuts
2. **Press a letter key** (m, d, a, e, r, i, s, c, o) - You'll be instantly navigated to that page
3. **Wait 2 seconds** - The popup will automatically dismiss if no key is pressed

## Features

- **Smart Context Detection**: Shortcuts are disabled when typing in input fields
- **Visual Feedback**: A popup shows available shortcuts when `g` is pressed
- **Fast & Intuitive**: Inspired by popular applications like GitHub, Gmail, etc.
- **Mobile Friendly**: Works on all devices and screen sizes

## Implementation

The global shortcuts are implemented in the `Header` component and use the browser's native keyboard event system. They integrate seamlessly with the GoTAK router system for instant navigation.

## Examples

- Quick access to tactical map: `g` → `m`
- Jump to alerts: `g` → `a` 
- Open settings: `g` → `s`
- Search for anything: `/` or `Cmd+K`

These shortcuts significantly improve navigation efficiency for tactical operators who need rapid access to different system functions.