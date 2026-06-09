import React from 'react';
import './Icon.css';

// Icon definitions - same as in IconShowcase but exported for reuse
export const icons = {
  // Navigation & Location
  target: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
      <circle cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 2V6M12 18V22M2 12H6M18 12H22" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  pin: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2C8.13 2 5 5.13 5 9C5 14.25 12 22 12 22S19 14.25 19 9C19 5.13 15.87 2 12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <circle cx="12" cy="9" r="3" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  'map-pin': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2C8.13 2 5 5.13 5 9C5 14.25 12 22 12 22S19 14.25 19 9C19 5.13 15.87 2 12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <circle cx="12" cy="9" r="3" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  map: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M1 6L8 3L16 6L23 3V18L16 21L8 18L1 21V6Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M8 3V18M16 6V21" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  compass: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M16.24 7.76L14.12 14.12L7.76 16.24L9.88 9.88L16.24 7.76Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),

  // Communication
  chat: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M20 2H4C2.9 2 2 2.9 2 4V16C2 17.1 2.9 18 4 18H7L12 22V18H20C21.1 18 22 17.1 22 16V4C22 2.9 21.1 2 20 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),
  
  broadcast: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M16.24 7.76C18.07 9.59 19.07 11.99 19.07 14.49C19.07 17 18.07 19.4 16.24 21.24" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M7.76 21.24C5.93 19.4 4.93 17 4.93 14.49C4.93 11.99 5.93 9.59 7.76 7.76" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  signal: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M5 12V20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M9 9V20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M13 6V20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M17 3V20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  wifi: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M2 8.82C7.88 4.18 16.12 4.18 22 8.82" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M5 12.55C8.97 9.4 15.03 9.4 19 12.55" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M8.5 16.05C10.65 14.34 13.35 14.34 15.5 16.05" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M12 20H12.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),

  // Security & Authentication
  lock: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M7 11V7C7 4.24 9.24 2 12 2C14.76 2 17 4.24 17 7V11" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="16" r="1" fill="currentColor"/>
    </svg>
  ),
  
  unlock: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M7 11V7C7 5.34 8.34 4 10 4C11.66 4 13 5.34 13 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="16" r="1" fill="currentColor"/>
    </svg>
  ),
  
  shield: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L4 7V11C4 16.52 7.64 21.44 12 22C16.36 21.44 20 16.52 20 11V7L12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M9 12L11 14L15 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),

  // Actions & Status
  rocket: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2C12 2 17 7 17 13C17 15 16 17 14 18L12 20L10 18C8 17 7 15 7 13C7 7 12 2 12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M12 11C12.55 11 13 11.45 13 12C13 12.55 12.55 13 12 13C11.45 13 11 12.55 11 12C11 11.45 11.45 11 12 11Z" fill="currentColor"/>
      <path d="M7 20L10 17M17 20L14 17" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  check: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M5 12L10 17L20 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  'check-circle': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M8 12L11 15L16 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  cross: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M15 9L9 15M9 9L15 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  x: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M18 6L6 18M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  warning: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L2 20H22L12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M12 9V13" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="17" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),
  
  alert: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M10 3H14L13 9H16L8 21L11 13H7L10 3Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),
  
  info: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 16V12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="8" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),
  
  error: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 8V12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="16" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),
  
  'alert-circle': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 8V12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="16" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),
  
  'alert-triangle': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L2 20H22L12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M12 9V13" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="17" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),

  // Data & Analytics
  chart: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="12" width="4" height="9" stroke="currentColor" strokeWidth="2"/>
      <rect x="10" y="7" width="4" height="14" stroke="currentColor" strokeWidth="2"/>
      <rect x="17" y="3" width="4" height="18" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  trending: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M3 17L9 11L13 15L21 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M14 7H21V14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  dashboard: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="3" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
      <rect x="14" y="3" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
      <rect x="3" y="14" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
      <rect x="14" y="14" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  // System & Settings
  settings: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <line x1="4" y1="21" x2="4" y2="14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <line x1="4" y1="10" x2="4" y2="3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <line x1="12" y1="21" x2="12" y2="12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <line x1="12" y1="8" x2="12" y2="3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <line x1="20" y1="21" x2="20" y2="16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <line x1="20" y1="12" x2="20" y2="3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="4" cy="12" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="12" cy="10" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="20" cy="14" r="2" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  sync: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M4 12C4 7.58 7.58 4 12 4C14.95 4 17.53 5.61 18.93 8" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M20 12C20 16.42 16.42 20 12 20C9.05 20 6.47 18.39 5.07 16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M16 8H19V5M8 16H5V19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  world: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M2 12H22M12 2C9.5 2 7.5 6.5 7.5 12C7.5 17.5 9.5 22 12 22C14.5 22 16.5 17.5 16.5 12C16.5 6.5 14.5 2 12 2Z" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  // Storage & Database
  database: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <ellipse cx="12" cy="5" rx="8" ry="3" stroke="currentColor" strokeWidth="2"/>
      <path d="M4 5V12C4 13.66 7.58 15 12 15C16.42 15 20 13.66 20 12V5" stroke="currentColor" strokeWidth="2"/>
      <path d="M4 12V19C4 20.66 7.58 22 12 22C16.42 22 20 20.66 20 19V12" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  save: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M17 21H7C5.9 21 5 20.1 5 19V5C5 3.9 5.9 3 7 3H14L19 8V19C19 20.1 18.1 21 17 21Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M15 3V8H19" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <rect x="8" y="13" width="8" height="5" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  package: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L2 7V17L12 22L22 17V7L12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M12 22V12M2 7L12 12L22 7M17 4.5L7 9.5" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),

  plus: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  download: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M21 15V19C21 20.1 20.1 21 19 21H5C3.9 21 3 20.1 3 19V15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M7 10L12 15L17 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M12 15V3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  // Tools & Operations
  tools: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M14.7 6.3C15.1 5.9 15.1 5.3 14.7 4.9C14.3 4.5 13.7 4.5 13.3 4.9L4.9 13.3C4.5 13.7 4.5 14.3 4.9 14.7L9.3 19.1C9.7 19.5 10.3 19.5 10.7 19.1L19.1 10.7C19.5 10.3 19.5 9.7 19.1 9.3L14.7 6.3Z" stroke="currentColor" strokeWidth="2"/>
      <path d="M7 17L3 21M21 3L17 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  search: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="10" cy="10" r="7" stroke="currentColor" strokeWidth="2"/>
      <path d="M15 15L21 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  // More icons...
  bell: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M18 8C18 6.4 17.3 4.9 16.1 3.7C14.9 2.5 13.4 2 12 2C10.6 2 9.1 2.5 7.9 3.7C6.7 4.9 6 6.4 6 8C6 15 3 17 3 17H21C21 17 18 15 18 8Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M10 21C10.3 21.6 10.8 22 11.4 22.3C12 22.5 12.5 22.5 13 22.3C13.6 22 14.1 21.6 14 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  users: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
      <path d="M3 21V19C3 16.24 5.24 14 8 14H10C12.76 14 15 16.24 15 19V21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="17" cy="7" r="3" stroke="currentColor" strokeWidth="2" opacity="0.5"/>
      <path d="M21 21V19C21 17 19.5 15.5 17.5 15.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.5"/>
    </svg>
  ),
  
  route: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M3 17L9 11L13 15L21 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <circle cx="3" cy="17" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="21" cy="7" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="9" cy="11" r="1.5" stroke="currentColor" strokeWidth="2"/>
      <circle cx="13" cy="15" r="1.5" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  link: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M15 7H19C20.66 7 22 8.34 22 10C22 11.66 20.66 13 19 13H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M9 17H5C3.34 17 2 15.66 2 14C2 12.34 3.34 11 5 11H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M8 12H16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  book: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M4 19.5C4 18.67 4.67 18 5.5 18H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M5.5 2H20V22H5.5C4.67 22 4 21.33 4 20.5V3.5C4 2.67 4.67 2 5.5 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),
  
  sparkle: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L13.09 8.91L20 10L13.09 11.09L12 18L10.91 11.09L4 10L10.91 8.91L12 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M18 6L18.5 8.5L21 9L18.5 9.5L18 12L17.5 9.5L15 9L17.5 8.5L18 6Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" opacity="0.7"/>
      <path d="M6 16L6.5 18L8.5 18.5L6.5 19L6 21L5.5 19L3.5 18.5L5.5 18L6 16Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" opacity="0.7"/>
    </svg>
  ),
  
  // Additional icons for Settings
  monitor: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="2" y="4" width="20" height="12" rx="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M8 20H16M12 16V20" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="18" cy="7" r="1" fill="currentColor"/>
    </svg>
  ),
  
  code: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M16 18L22 12L16 6M8 6L2 12L8 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  'chevron-down': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M6 9L12 15L18 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  'external-link': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M18 13V19C18 20.1 17.1 21 16 21H5C3.9 21 3 20.1 3 19V8C3 6.9 3.9 6 5 6H11" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M15 3H21V9M10 14L21 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  
  refresh: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M1 4V10H7M23 20V14H17" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M20.49 9C19.9 5.99 17.24 3.64 14 3.64C10.76 3.64 8.1 5.99 7.51 9M3.51 15C4.1 18.01 6.76 20.36 10 20.36C13.24 20.36 15.9 18.01 16.49 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  loader: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2V6M12 18V22M4.93 4.93L7.76 7.76M16.24 16.24L19.07 19.07M2 12H6M18 12H22M4.93 19.07L7.76 16.24M16.24 7.76L19.07 4.93" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  server: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="4" width="18" height="5" rx="1" stroke="currentColor" strokeWidth="2"/>
      <rect x="3" y="11" width="18" height="5" rx="1" stroke="currentColor" strokeWidth="2"/>
      <rect x="3" y="18" width="18" height="3" rx="1" stroke="currentColor" strokeWidth="2"/>
      <circle cx="6" cy="6.5" r="1" fill="currentColor"/>
      <circle cx="6" cy="13.5" r="1" fill="currentColor"/>
      <circle cx="6" cy="19.5" r="0.5" fill="currentColor"/>
      <line x1="9" y1="6.5" x2="18" y2="6.5" stroke="currentColor" strokeWidth="1" opacity="0.5"/>
      <line x1="9" y1="13.5" x2="18" y2="13.5" stroke="currentColor" strokeWidth="1" opacity="0.5"/>
    </svg>
  ),
  
  palette: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C13.19 22 14.34 21.78 15.42 21.37C16.16 21.12 16.71 20.46 16.81 19.68C16.91 18.84 16.55 18.03 15.89 17.54C15.54 17.28 15.33 16.88 15.33 16.44C15.33 15.64 15.97 15 16.77 15H18C20.76 15 23 12.76 23 10C23 5.58 18.42 2 12 2Z" stroke="currentColor" strokeWidth="2"/>
      <circle cx="6.5" cy="11.5" r="1.5" fill="currentColor"/>
      <circle cx="9.5" cy="7.5" r="1.5" fill="currentColor"/>
      <circle cx="14.5" cy="7.5" r="1.5" fill="currentColor"/>
      <circle cx="17.5" cy="11.5" r="1.5" fill="currentColor"/>
    </svg>
  ),
  
  inbox: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M22 12H16L14 15H10L8 12H2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M5 7L4 12V18C4 19.1 4.9 20 6 20H18C19.1 20 20 19.1 20 18V12L19 7C19 5.9 18.1 5 17 5H7C5.9 5 5 5.9 5 7Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),
  
  bot: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="5" y="8" width="14" height="10" rx="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="9" cy="12" r="1" fill="currentColor"/>
      <circle cx="15" cy="12" r="1" fill="currentColor"/>
      <path d="M8 14H16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M12 8V5M12 5C12 3.9 12.9 3 14 3M12 5C12 3.9 11.1 3 10 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  
  wrench: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M14.7 6.3C15.1 5.9 15.1 5.3 14.7 4.9C14.3 4.5 13.7 4.5 13.3 4.9L4.9 13.3C4.5 13.7 4.5 14.3 4.9 14.7L9.3 19.1C9.7 19.5 10.3 19.5 10.7 19.1L19.1 10.7C19.5 10.3 19.5 9.7 19.1 9.3L14.7 6.3Z" stroke="currentColor" strokeWidth="2"/>
      <circle cx="17" cy="7" r="3" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),
  
  send: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M22 2L11 13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  ),

  // User & Entity icons
  user: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
      <path d="M4 21V19C4 16.24 6.24 14 9 14H15C17.76 14 20 16.24 20 19V21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  'log-out': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M9 21H5C3.9 21 3 20.1 3 19V5C3 3.9 3.9 3 5 3H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  eye: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M1 12C1 12 5 4 12 4C19 4 23 12 23 12C23 12 19 20 12 20C5 20 1 12 1 12Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  'eye-off': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M17.94 17.94C16.23 19.24 14.23 20 12 20C5 20 1 12 1 12C2.24 9.68 3.75 7.62 5.47 5.93M9.88 9.88C10.13 9.14 10.67 8.52 11.36 8.16C12.05 7.8 12.85 7.72 13.59 7.94C14.33 8.16 14.95 8.66 15.32 9.32C15.68 9.98 15.76 10.77 15.54 11.5M9.88 9.88L15.54 11.5M9.88 9.88L6 6M15.54 11.5L18 18M3 3L21 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),

  radio: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 2V8M12 16V22M4.93 4.93L9.17 9.17M14.83 14.83L19.07 19.07M2 12H8M16 12H22M4.93 19.07L9.17 14.83M14.83 9.17L19.07 4.93" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  // Entity Types
  drone: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="8" y="10" width="8" height="4" rx="1" stroke="currentColor" strokeWidth="2"/>
      <path d="M4 2L8 6M20 2L16 6M4 22L8 18M20 22L16 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="4" cy="2" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="20" cy="2" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="4" cy="22" r="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="20" cy="22" r="2" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  vehicle: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M5 11L3 6H7L9 11M19 11L21 6H17L15 11" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <rect x="3" y="11" width="18" height="7" rx="1" stroke="currentColor" strokeWidth="2"/>
      <circle cx="7" cy="20" r="1.5" stroke="currentColor" strokeWidth="2"/>
      <circle cx="17" cy="20" r="1.5" stroke="currentColor" strokeWidth="2"/>
      <rect x="6" y="13" width="4" height="3" stroke="currentColor" strokeWidth="2"/>
      <rect x="14" y="13" width="4" height="3" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  sensor: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 2V5M12 19V22M2 12H5M19 12H22" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M5.64 5.64L7.76 7.76M16.24 16.24L18.36 18.36M5.64 18.36L7.76 16.24M16.24 7.76L18.36 5.64" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  camera: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="6" width="18" height="13" rx="2" stroke="currentColor" strokeWidth="2"/>
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
      <path d="M9 3L10 6H14L15 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <circle cx="18" cy="9" r="1" fill="currentColor"/>
    </svg>
  ),

  radar: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <circle cx="12" cy="12" r="6" stroke="currentColor" strokeWidth="1" opacity="0.5"/>
      <circle cx="12" cy="12" r="2" stroke="currentColor" strokeWidth="1" opacity="0.5"/>
      <path d="M12 2L12 12L20 8" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" fill="currentColor" opacity="0.2"/>
      <circle cx="8" cy="8" r="1" fill="currentColor"/>
      <circle cx="16" cy="14" r="1" fill="currentColor"/>
    </svg>
  ),

  battery: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="7" width="15" height="10" rx="1" stroke="currentColor" strokeWidth="2"/>
      <rect x="20" y="10" width="2" height="4" rx="1" stroke="currentColor" strokeWidth="2" fill="currentColor"/>
      <rect x="5" y="9" width="8" height="6" fill="currentColor" opacity="0.8"/>
    </svg>
  ),

  'battery-low': (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="7" width="15" height="10" rx="1" stroke="currentColor" strokeWidth="2"/>
      <rect x="20" y="10" width="2" height="4" rx="1" stroke="currentColor" strokeWidth="2" fill="currentColor"/>
      <rect x="5" y="9" width="3" height="6" fill="currentColor" opacity="0.8"/>
    </svg>
  ),

  equipment: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="4" y="8" width="16" height="12" rx="2" stroke="currentColor" strokeWidth="2"/>
      <path d="M8 8V6C8 4.9 8.9 4 10 4H14C15.1 4 16 4.9 16 6V8" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 12V16M10 14H14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  soldier: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="7" r="3" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 10V14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M8 11L12 14L16 11" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M12 14L9 21M12 14L15 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M9 3H15V5C15 6.1 14.1 7 13 7H11C9.9 7 9 6.1 9 5V3Z" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  medic: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 8V16M8 12H16" stroke="currentColor" strokeWidth="3" strokeLinecap="round"/>
    </svg>
  ),

  hostile: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2L3 7V12C3 16.55 5.84 20.74 10 22C10.35 21.91 10.69 21.8 11 21.68" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M13 22C17.16 20.74 20 16.55 20 12V7L12 2" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
      <path d="M9 9L15 15M15 9L9 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  neutral: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="4" y="4" width="16" height="16" stroke="currentColor" strokeWidth="2"/>
      <path d="M8 12H16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),

  unknown: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M9 9C9 7.34 10.34 6 12 6C13.66 6 15 7.34 15 9C15 10.3 14.3 11.4 13.2 12C12.5 12.4 12 13.1 12 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="12" cy="17" r="0.5" fill="currentColor" stroke="currentColor"/>
    </svg>
  ),

  // View Types
  grid: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="3" y="3" width="8" height="8" stroke="currentColor" strokeWidth="2"/>
      <rect x="13" y="3" width="8" height="8" stroke="currentColor" strokeWidth="2"/>
      <rect x="3" y="13" width="8" height="8" stroke="currentColor" strokeWidth="2"/>
      <rect x="13" y="13" width="8" height="8" stroke="currentColor" strokeWidth="2"/>
    </svg>
  ),

  list: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M8 6H21M8 12H21M8 18H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="4" cy="6" r="1" fill="currentColor"/>
      <circle cx="4" cy="12" r="1" fill="currentColor"/>
      <circle cx="4" cy="18" r="1" fill="currentColor"/>
    </svg>
  ),

  tactical: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 2V7M12 17V22M2 12H7M17 12H22" stroke="currentColor" strokeWidth="1" strokeLinecap="round" opacity="0.5"/>
      <circle cx="8" cy="8" r="2" fill="currentColor"/>
      <circle cx="16" cy="10" r="2" fill="currentColor"/>
      <circle cx="10" cy="16" r="2" fill="currentColor"/>
    </svg>
  ),

  // Stats icons  
  speed: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2Z" stroke="currentColor" strokeWidth="2"/>
      <path d="M12 6V12L16 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),

  altitude: (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M7 14L12 4L17 14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M5 20H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
      <path d="M12 4V20" stroke="currentColor" strokeWidth="1" strokeLinecap="round" opacity="0.5"/>
    </svg>
  )
};

// Icon type definitions
export type IconName = keyof typeof icons;

interface IconProps {
  name: IconName;
  size?: number;
  color?: string;
  className?: string;
  onClick?: () => void;
}

// Main Icon component
export const Icon: React.FC<IconProps> = ({ 
  name, 
  size = 24, 
  color = 'currentColor', 
  className = '',
  onClick 
}) => {
  const iconSvg = icons[name];
  
  if (!iconSvg) {
    console.warn(`Icon "${name}" not found`);
    return null;
  }

  return (
    <span 
      className={`gotak-icon ${className}`}
      style={{ 
        width: size, 
        height: size, 
        color,
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center'
      }}
      onClick={onClick}
    >
      {React.cloneElement(iconSvg as React.ReactElement, {
        width: size,
        height: size,
        style: { width: size, height: size }
      })}
    </span>
  );
};

export default Icon;