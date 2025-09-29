import React from 'react';

// Integration logo components - SVG logos for each service
export const IntegrationLogos: Record<string, React.ReactElement> = {
  // Intelligence & OSINT
  janes: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#003366"/>
      <text x="50" y="60" fontSize="32" fontWeight="bold" fill="white" textAnchor="middle" fontFamily="Arial, sans-serif">
        JANES
      </text>
    </svg>
  ),
  
  shodan: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#BE0F34"/>
      <text x="50" y="50" fontSize="20" fontWeight="bold" fill="white" textAnchor="middle" fontFamily="monospace">
        SHODAN
      </text>
      <circle cx="50" cy="70" r="15" fill="none" stroke="white" strokeWidth="2"/>
      <circle cx="50" cy="70" r="8" fill="white"/>
    </svg>
  ),
  
  'osint-framework': (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#2C3E50"/>
      <circle cx="50" cy="40" r="20" fill="none" stroke="#3498DB" strokeWidth="3"/>
      <path d="M50 40 L65 55 L35 55 Z" fill="#3498DB"/>
      <text x="50" y="80" fontSize="14" fill="#3498DB" textAnchor="middle" fontFamily="Arial, sans-serif">
        OSINT
      </text>
    </svg>
  ),
  
  maltego: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#FF6B00"/>
      <circle cx="30" cy="30" r="8" fill="white"/>
      <circle cx="70" cy="30" r="8" fill="white"/>
      <circle cx="30" cy="70" r="8" fill="white"/>
      <circle cx="70" cy="70" r="8" fill="white"/>
      <circle cx="50" cy="50" r="10" fill="white"/>
      <path d="M30 30 L50 50 M70 30 L50 50 M30 70 L50 50 M70 70 L50 50" stroke="white" strokeWidth="2"/>
    </svg>
  ),
  
  misp: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#0088CC"/>
      <text x="50" y="55" fontSize="24" fontWeight="bold" fill="white" textAnchor="middle" fontFamily="Arial, sans-serif">
        MISP
      </text>
      <path d="M20 70 L35 70 L40 80 L45 65 L50 75 L55 70 L80 70" stroke="white" strokeWidth="2" fill="none"/>
    </svg>
  ),
  
  opencti: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#001F3F"/>
      <circle cx="50" cy="50" r="30" fill="none" stroke="#00D4FF" strokeWidth="3"/>
      <circle cx="50" cy="20" r="5" fill="#00D4FF"/>
      <circle cx="75" cy="40" r="5" fill="#00D4FF"/>
      <circle cx="75" cy="60" r="5" fill="#00D4FF"/>
      <circle cx="50" cy="80" r="5" fill="#00D4FF"/>
      <circle cx="25" cy="60" r="5" fill="#00D4FF"/>
      <circle cx="25" cy="40" r="5" fill="#00D4FF"/>
      <path d="M50 20 L75 40 L75 60 L50 80 L25 60 L25 40 Z" stroke="#00D4FF" strokeWidth="1" fill="none"/>
    </svg>
  ),
  
  // Security
  vault: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#000000"/>
      <polygon points="50,70 28,30 72,30" fill="#FFD700" stroke="#FFD700" strokeWidth="1"/>
    </svg>
  ),
  
  keycloak: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#4D5061"/>
      <path d="M30 30 L70 30 L60 50 L70 70 L30 70 L40 50 Z" fill="#00B9E4"/>
      <path d="M40 50 L50 30 L60 50 L50 70 Z" fill="#EC7A08"/>
    </svg>
  ),
  
  okta: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#007DC1"/>
      <circle cx="50" cy="50" r="25" fill="none" stroke="white" strokeWidth="6"/>
      <circle cx="50" cy="25" r="6" fill="white"/>
    </svg>
  ),
  
  // Cloud Storage
  s3: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#FF9900"/>
      <path d="M30 25 L50 15 L70 25 L70 50 L50 60 L30 50 Z" fill="white"/>
      <path d="M30 50 L50 60 L70 50 L70 75 L50 85 L30 75 Z" fill="white" opacity="0.7"/>
      <text x="50" y="40" fontSize="16" fontWeight="bold" fill="#FF9900" textAnchor="middle" fontFamily="Arial, sans-serif">
        S3
      </text>
    </svg>
  ),
  
  'azure-storage': (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#0078D4"/>
      <path d="M20 40 L45 15 L80 30 L65 60 L35 60 Z" fill="white"/>
      <path d="M35 60 L65 60 L55 85 L20 70 Z" fill="white" opacity="0.7"/>
    </svg>
  ),
  
  minio: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#C72E49"/>
      <rect x="20" y="30" width="25" height="40" fill="white"/>
      <rect x="55" y="30" width="25" height="40" fill="white"/>
      <rect x="30" y="40" width="40" height="5" fill="white"/>
    </svg>
  ),
  
  // Messaging
  slack: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#4A154B"/>
      <path d="M35 20 a8 8 0 0 1 0 16 h-8 v-8 a8 8 0 0 1 8-8" fill="#E01E5A"/>
      <path d="M43 36 a8 8 0 0 1 -16 0 v-8 h8 a8 8 0 0 1 8 8" fill="#E01E5A"/>
      <path d="M65 80 a8 8 0 0 1 0-16 h8 v8 a8 8 0 0 1 -8 8" fill="#36C5F0"/>
      <path d="M57 64 a8 8 0 0 1 16 0 v8 h-8 a8 8 0 0 1 -8-8" fill="#36C5F0"/>
      <path d="M80 35 a8 8 0 0 1 -16 0 v-8 h8 a8 8 0 0 1 8 8" fill="#2EB67D"/>
      <path d="M64 43 a8 8 0 0 1 0 16 h-8 v-8 a8 8 0 0 1 8-8" fill="#2EB67D"/>
      <path d="M20 65 a8 8 0 0 1 16 0 v8 h-8 a8 8 0 0 1 -8-8" fill="#ECB22E"/>
      <path d="M36 57 a8 8 0 0 1 0-16 h8 v8 a8 8 0 0 1 -8 8" fill="#ECB22E"/>
    </svg>
  ),
  
  mattermost: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#0058CC"/>
      <circle cx="50" cy="50" r="30" fill="white"/>
      <circle cx="40" cy="45" r="4" fill="#0058CC"/>
      <circle cx="60" cy="45" r="4" fill="#0058CC"/>
      <path d="M35 55 Q50 65 65 55" stroke="#0058CC" strokeWidth="3" fill="none" strokeLinecap="round"/>
    </svg>
  ),
  
  matrix: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#000000"/>
      <rect x="10" y="20" width="5" height="60" fill="#FFFFFF"/>
      <rect x="85" y="20" width="5" height="60" fill="#FFFFFF"/>
      <rect x="10" y="20" width="20" height="5" fill="#FFFFFF"/>
      <rect x="10" y="75" width="20" height="5" fill="#FFFFFF"/>
      <rect x="70" y="20" width="20" height="5" fill="#FFFFFF"/>
      <rect x="70" y="75" width="20" height="5" fill="#FFFFFF"/>
      <circle cx="35" cy="50" r="8" fill="#FFFFFF"/>
      <circle cx="50" cy="50" r="8" fill="#FFFFFF"/>
      <circle cx="65" cy="50" r="8" fill="#FFFFFF"/>
    </svg>
  ),
  
  // Monitoring
  prometheus: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#E6522C"/>
      <circle cx="50" cy="45" r="25" fill="white"/>
      <path d="M50 30 L55 45 L70 45 L57 55 L62 70 L50 60 L38 70 L43 55 L30 45 L45 45 Z" fill="#E6522C"/>
      <rect x="35" y="75" width="30" height="5" fill="white"/>
      <rect x="40" y="80" width="20" height="5" fill="white"/>
    </svg>
  ),
  
  elasticsearch: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#005571"/>
      <circle cx="50" cy="30" r="15" fill="#FEC514"/>
      <rect x="20" y="45" width="60" height="10" fill="#24BBB1"/>
      <circle cx="50" cy="70" r="15" fill="#EF5098"/>
    </svg>
  ),
  
  datadog: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#632CA6"/>
      <path d="M30 60 L30 40 Q30 30 40 30 L60 30 Q70 30 70 40 L70 60 Q70 70 60 70 L40 70 Q30 70 30 60" fill="white"/>
      <path d="M45 45 L45 55 M55 45 L55 55" stroke="#632CA6" strokeWidth="5" strokeLinecap="round"/>
      <circle cx="45" cy="45" r="3" fill="#632CA6"/>
      <circle cx="55" cy="45" r="3" fill="#632CA6"/>
    </svg>
  ),
  
  // Maps & GIS
  mapbox: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#4264FB"/>
      <circle cx="50" cy="50" r="30" fill="white"/>
      <circle cx="50" cy="40" r="8" fill="#4264FB"/>
      <path d="M50 48 L40 65 L50 60 L60 65 Z" fill="#4264FB"/>
    </svg>
  ),
  
  arcgis: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#007AC2"/>
      <path d="M25 50 Q50 20 75 50 Q50 80 25 50" fill="white"/>
      <circle cx="50" cy="50" r="8" fill="#007AC2"/>
    </svg>
  ),
  
  'google-maps': (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#34A853"/>
      <path d="M50 20 C35 20 25 32 25 45 C25 65 50 80 50 80 S75 65 75 45 C75 32 65 20 50 20" fill="#EA4335"/>
      <circle cx="50" cy="43" r="8" fill="white"/>
    </svg>
  ),
  
  // Data Export
  kafka: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#000000"/>
      <circle cx="30" cy="30" r="8" fill="white"/>
      <circle cx="70" cy="30" r="8" fill="white"/>
      <circle cx="50" cy="50" r="12" fill="white"/>
      <circle cx="30" cy="70" r="8" fill="white"/>
      <circle cx="70" cy="70" r="8" fill="white"/>
      <path d="M30 30 L50 50 M70 30 L50 50 M30 70 L50 50 M70 70 L50 50" stroke="white" strokeWidth="2"/>
    </svg>
  ),
  
  rabbitmq: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#FF6600"/>
      <ellipse cx="50" cy="45" rx="25" ry="20" fill="white"/>
      <ellipse cx="40" cy="40" rx="4" ry="8" fill="#FF6600"/>
      <ellipse cx="60" cy="40" rx="4" ry="8" fill="#FF6600"/>
      <path d="M35 65 L35 75 M45 65 L45 75 M55 65 L55 75 M65 65 L65 75" stroke="white" strokeWidth="3" strokeLinecap="round"/>
    </svg>
  ),
  
  webhook: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#2E3440"/>
      <path d="M30 50 L45 35 L45 45 L70 45 L70 55 L45 55 L45 65 Z" fill="#88C0D0"/>
      <circle cx="20" cy="50" r="8" fill="#5E81AC"/>
      <circle cx="80" cy="50" r="8" fill="#BF616A"/>
    </svg>
  ),
  
  // AI & ML
  openai: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#000000"/>
      <path d="M50 15 L70 27.5 L70 52.5 L50 65 L30 52.5 L30 27.5 Z" fill="none" stroke="#00D9A3" strokeWidth="3"/>
      <path d="M50 35 L62 42.5 L62 57.5 L50 65 L38 57.5 L38 42.5 Z" fill="none" stroke="#00D9A3" strokeWidth="2"/>
      <circle cx="50" cy="50" r="8" fill="#00D9A3"/>
    </svg>
  ),
  
  tensorflow: (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#FF6F00"/>
      <path d="M25 30 L50 15 L75 30 L75 40 L50 55 L25 40 Z" fill="white"/>
      <path d="M50 55 L50 85 L60 80 L60 60 Z" fill="white"/>
      <path d="M40 50 L40 70 L30 75 L30 55 Z" fill="white" opacity="0.7"/>
    </svg>
  ),
  
  'amazon-rekognition': (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <rect width="100" height="100" fill="#FF9900"/>
      <circle cx="50" cy="40" r="20" fill="none" stroke="white" strokeWidth="3"/>
      <circle cx="50" cy="40" r="8" fill="white"/>
      <rect x="30" y="60" width="40" height="4" fill="white"/>
      <rect x="35" y="68" width="30" height="4" fill="white"/>
      <rect x="40" y="76" width="20" height="4" fill="white"/>
    </svg>
  ),
};

export default IntegrationLogos;