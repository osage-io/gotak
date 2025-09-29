/**
 * Integrations Marketplace Page
 * Central hub for configuring third-party service integrations
 */

import React, { useState, useCallback } from 'react';
import { IntegrationLogos } from '../components/IntegrationLogos';
import { Icon } from '../components/ui/Icon';
import './Integrations.css';

interface Integration {
  id: string;
  name: string;
  category: 'security' | 'cloud' | 'messaging' | 'monitoring' | 'maps' | 'data' | 'ai' | 'intelligence';
  icon: string;
  description: string;
  features: string[];
  status: 'available' | 'configured' | 'connected' | 'error';
  provider: string;
  docsUrl?: string;
  configurable: boolean;
}

interface VaultConfig {
  url: string;
  namespace: string;
  authMethod: 'token' | 'userpass' | 'ldap' | 'kubernetes' | 'oidc';
  token?: string;
  username?: string;
  password?: string;
  // OIDC Configuration
  oidcEnabled?: boolean;
  oidcDiscoveryUrl?: string;
  oidcClientId?: string;
  oidcClientSecret?: string;
  oidcDefaultRole?: string;
  oidcRedirectUri?: string;
  // GoTAK Integration
  gotakLoginIntegration?: boolean;
  gotakAuthPath?: string;
  tlsEnabled: boolean;
  pkiEngine: string;
  encryptionEnabled: boolean;
  transitEngine: string;
}

const Integrations: React.FC = () => {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [viewMode, setViewMode] = useState<'list' | 'cards'>('list');
  const [selectedIntegration, setSelectedIntegration] = useState<Integration | null>(null);
  const [anthropicApiKey, setAnthropicApiKey] = useState('');
  const [vaultConfig, setVaultConfig] = useState<VaultConfig>({
    url: 'http://localhost:8200',
    namespace: '',
    authMethod: 'token',
    token: '',
    oidcEnabled: false,
    oidcDiscoveryUrl: '',
    oidcClientId: '',
    oidcClientSecret: '',
    oidcDefaultRole: 'management',
    oidcRedirectUri: '',
    gotakLoginIntegration: false,
    gotakAuthPath: 'auth/userpass',
    tlsEnabled: false,
    pkiEngine: 'pki',
    encryptionEnabled: false,
    transitEngine: 'transit',
  });

  // Integration categories with Icon component names
  const categories = [
    { id: 'all', label: 'All Integrations', iconName: 'link' },
    { id: 'intelligence', label: 'Intelligence', iconName: 'search' },
    { id: 'security', label: 'Security', iconName: 'lock' },
    { id: 'cloud', label: 'Cloud Storage', iconName: 'database' },
    { id: 'messaging', label: 'Messaging', iconName: 'chat' },
    { id: 'monitoring', label: 'Monitoring', iconName: 'chart' },
    { id: 'maps', label: 'Maps & GIS', iconName: 'map' },
    { id: 'data', label: 'Data Export', iconName: 'package' },
    { id: 'ai', label: 'AI & ML', iconName: 'sparkle' },
  ] as const;

  // Available integrations
  const [integrations, setIntegrations] = useState<Integration[]>([
    // Intelligence & OSINT
    {
      id: 'janes',
      name: 'Janes',
      category: 'intelligence',
      icon: '🎯',
      description: 'Military intelligence, defense capabilities, and equipment identification database',
      features: [
        'Equipment Recognition',
        'Military Capabilities Database',
        'Threat Assessment',
        'Order of Battle (ORBAT)',
        'Defense Industry Analysis',
        'Country Risk Intelligence'
      ],
      status: 'available',
      provider: 'Janes Information Group',
      docsUrl: 'https://www.janes.com/api-documentation',
      configurable: true,
    },
    {
      id: 'shodan',
      name: 'Shodan',
      category: 'intelligence',
      icon: '🌐',
      description: 'Internet-connected device search engine for infrastructure reconnaissance',
      features: [
        'Device Discovery',
        'Network Mapping',
        'Vulnerability Detection',
        'Port Scanning Results',
        'SSL Certificate Analysis',
        'Industrial Control Systems'
      ],
      status: 'available',
      provider: 'Shodan',
      docsUrl: 'https://developer.shodan.io/api',
      configurable: true,
    },
    {
      id: 'osint-framework',
      name: 'OSINT Framework',
      category: 'intelligence',
      icon: '🕵️',
      description: 'Comprehensive open source intelligence gathering and analysis tools',
      features: [
        'Social Media Analysis',
        'Domain Investigation',
        'Image Intelligence',
        'Geolocation Services',
        'Dark Web Monitoring',
        'Threat Intelligence Feeds'
      ],
      status: 'available',
      provider: 'OSINT Community',
      configurable: true,
    },
    {
      id: 'maltego',
      name: 'Maltego',
      category: 'intelligence',
      icon: '🕸️',
      description: 'Link analysis and data mining for investigative intelligence gathering',
      features: [
        'Entity Relationship Mapping',
        'Transform Hub',
        'Graph Analysis',
        'Data Correlation',
        'Pattern Recognition',
        'Automated Reconnaissance'
      ],
      status: 'available',
      provider: 'Maltego Technologies',
      docsUrl: 'https://docs.maltego.com/',
      configurable: true,
    },
    {
      id: 'misp',
      name: 'MISP',
      category: 'intelligence',
      icon: '🔒',
      description: 'Malware Information Sharing Platform for threat intelligence sharing',
      features: [
        'Threat Intelligence Sharing',
        'IOC Management',
        'Malware Analysis',
        'Attack Pattern Database',
        'Event Correlation',
        'STIX/TAXII Support'
      ],
      status: 'available',
      provider: 'MISP Project',
      docsUrl: 'https://www.misp-project.org/documentation/',
      configurable: true,
    },
    {
      id: 'opencti',
      name: 'OpenCTI',
      category: 'intelligence',
      icon: '📡',
      description: 'Open Cyber Threat Intelligence platform for knowledge management and analysis',
      features: [
        'Threat Actor Tracking',
        'Campaign Analysis',
        'TTPs Mapping',
        'Knowledge Graph',
        'MITRE ATT&CK Framework',
        'Intelligence Reports'
      ],
      status: 'available',
      provider: 'Filigran',
      docsUrl: 'https://docs.opencti.io/',
      configurable: true,
    },
    
    // Security
    {
      id: 'vault',
      name: 'HashiCorp Vault',
      category: 'security',
      icon: '🔐',
      description: 'Enterprise-grade secrets management, encryption as a service, and dynamic PKI infrastructure',
      features: [
        'TLS Certificate Management',
        'Encryption as a Service',
        'Dynamic Secrets',
        'PKI Infrastructure',
        'Auto-unseal',
        'Audit Logging'
      ],
      status: 'available',
      provider: 'HashiCorp',
      docsUrl: 'https://www.vaultproject.io/docs',
      configurable: true,
    },
    {
      id: 'keycloak',
      name: 'Keycloak',
      category: 'security',
      icon: '🛡️',
      description: 'Open source identity and access management with SSO, LDAP, and OAuth2 support',
      features: [
        'Single Sign-On (SSO)',
        'LDAP Integration',
        'OAuth2/OIDC',
        'User Federation',
        'Multi-factor Auth',
        'Role-based Access'
      ],
      status: 'available',
      provider: 'Red Hat',
      configurable: true,
    },
    {
      id: 'okta',
      name: 'Okta',
      category: 'security',
      icon: '🔑',
      description: 'Cloud-based identity management and authentication service',
      features: [
        'Universal Directory',
        'Adaptive MFA',
        'Lifecycle Management',
        'API Access Management',
        'B2B Integration',
        'Compliance Reports'
      ],
      status: 'available',
      provider: 'Okta Inc.',
      configurable: true,
    },

    // Cloud Storage
    {
      id: 's3',
      name: 'AWS S3',
      category: 'cloud',
      icon: '📦',
      description: 'Scalable object storage for data archiving, backup, and analytics',
      features: [
        'Unlimited Storage',
        'High Durability',
        'Lifecycle Policies',
        'Versioning',
        'Cross-region Replication',
        'Server-side Encryption'
      ],
      status: 'available',
      provider: 'Amazon Web Services',
      configurable: true,
    },
    {
      id: 'azure-storage',
      name: 'Azure Storage',
      category: 'cloud',
      icon: '☁️',
      description: 'Microsoft cloud storage for unstructured data and file sharing',
      features: [
        'Blob Storage',
        'File Shares',
        'Queue Storage',
        'Geo-redundancy',
        'Hot/Cool Tiers',
        'Azure AD Integration'
      ],
      status: 'available',
      provider: 'Microsoft Azure',
      configurable: true,
    },
    {
      id: 'minio',
      name: 'MinIO',
      category: 'cloud',
      icon: '🗄️',
      description: 'High-performance, S3-compatible object storage for on-premises deployment',
      features: [
        'S3 Compatible API',
        'Erasure Coding',
        'Bitrot Protection',
        'Lambda Compute',
        'Encryption',
        'Multi-tenancy'
      ],
      status: 'available',
      provider: 'MinIO Inc.',
      configurable: true,
    },

    // Messaging
    {
      id: 'slack',
      name: 'Slack',
      category: 'messaging',
      icon: '💬',
      description: 'Team collaboration and real-time messaging integration',
      features: [
        'Channel Notifications',
        'Direct Messages',
        'File Sharing',
        'Alert Webhooks',
        'Bot Commands',
        'Thread Replies'
      ],
      status: 'available',
      provider: 'Slack Technologies',
      configurable: true,
    },
    {
      id: 'mattermost',
      name: 'Mattermost',
      category: 'messaging',
      icon: '🗨️',
      description: 'Open source, self-hosted team collaboration platform',
      features: [
        'Self-hosted',
        'End-to-end Encryption',
        'Compliance Features',
        'Custom Webhooks',
        'Plugin System',
        'Mobile Apps'
      ],
      status: 'available',
      provider: 'Mattermost Inc.',
      configurable: true,
    },
    {
      id: 'matrix',
      name: 'Matrix/Element',
      category: 'messaging',
      icon: '🔐',
      description: 'Decentralized, encrypted communication protocol and client',
      features: [
        'Decentralized',
        'E2E Encryption',
        'Federation',
        'Voice/Video Calls',
        'Bridging',
        'Self-hosting'
      ],
      status: 'available',
      provider: 'Matrix.org',
      configurable: true,
    },

    // Monitoring
    {
      id: 'prometheus',
      name: 'Prometheus',
      category: 'monitoring',
      icon: '📈',
      description: 'Open source monitoring and alerting toolkit for reliability and scalability',
      features: [
        'Time-series Database',
        'PromQL Queries',
        'Alert Manager',
        'Service Discovery',
        'Grafana Integration',
        'Multi-dimensional Data'
      ],
      status: 'available',
      provider: 'CNCF',
      configurable: true,
    },
    {
      id: 'elasticsearch',
      name: 'Elasticsearch',
      category: 'monitoring',
      icon: '🔍',
      description: 'Distributed search and analytics engine for log aggregation',
      features: [
        'Full-text Search',
        'Log Aggregation',
        'Real-time Analytics',
        'Kibana Dashboards',
        'Machine Learning',
        'Anomaly Detection'
      ],
      status: 'available',
      provider: 'Elastic',
      configurable: true,
    },
    {
      id: 'datadog',
      name: 'Datadog',
      category: 'monitoring',
      icon: '🐕',
      description: 'Cloud-scale monitoring and security platform for complete observability',
      features: [
        'APM & Tracing',
        'Log Management',
        'Infrastructure Monitoring',
        'Real User Monitoring',
        'Synthetic Monitoring',
        'Security Monitoring'
      ],
      status: 'available',
      provider: 'Datadog Inc.',
      configurable: true,
    },

    // Maps & GIS
    {
      id: 'mapbox',
      name: 'Mapbox',
      category: 'maps',
      icon: '🗺️',
      description: 'Custom maps and location services with vector tiles and geocoding',
      features: [
        'Vector Tiles',
        'Custom Styles',
        'Geocoding API',
        'Navigation SDK',
        'Offline Maps',
        'Satellite Imagery'
      ],
      status: 'available',
      provider: 'Mapbox Inc.',
      configurable: true,
    },
    {
      id: 'arcgis',
      name: 'ArcGIS',
      category: 'maps',
      icon: '🌍',
      description: 'Enterprise GIS platform for spatial analysis and mapping',
      features: [
        'Spatial Analysis',
        'Feature Services',
        'Geoprocessing',
        'Story Maps',
        'Field Operations',
        'Real-time GIS'
      ],
      status: 'available',
      provider: 'Esri',
      configurable: true,
    },
    {
      id: 'google-maps',
      name: 'Google Maps',
      category: 'maps',
      icon: '📍',
      description: 'Comprehensive mapping platform with Street View and real-time traffic',
      features: [
        'Street View',
        'Real-time Traffic',
        'Places API',
        'Directions API',
        'Distance Matrix',
        'Elevation API'
      ],
      status: 'available',
      provider: 'Google',
      configurable: true,
    },

    // Data Export
    {
      id: 'kafka',
      name: 'Apache Kafka',
      category: 'data',
      icon: '📊',
      description: 'Distributed event streaming platform for high-throughput data pipelines',
      features: [
        'Event Streaming',
        'Pub/Sub Messaging',
        'Stream Processing',
        'Exactly-once Delivery',
        'Horizontal Scaling',
        'Fault Tolerance'
      ],
      status: 'available',
      provider: 'Apache Foundation',
      configurable: true,
    },
    {
      id: 'rabbitmq',
      name: 'RabbitMQ',
      category: 'data',
      icon: '🐰',
      description: 'Message broker for reliable message delivery and routing',
      features: [
        'Message Queuing',
        'Topic Exchange',
        'Dead Letter Queues',
        'Priority Queues',
        'Federation',
        'Management UI'
      ],
      status: 'available',
      provider: 'VMware',
      configurable: true,
    },
    {
      id: 'webhook',
      name: 'Webhooks',
      category: 'data',
      icon: '🔗',
      description: 'HTTP callbacks for real-time event notifications to external systems',
      features: [
        'Event Triggers',
        'JSON Payloads',
        'Retry Logic',
        'Authentication',
        'Rate Limiting',
        'Batch Processing'
      ],
      status: 'available',
      provider: 'Generic',
      configurable: true,
    },

    // AI & ML
    {
      id: 'anthropic',
      name: 'Anthropic Claude',
      category: 'ai',
      icon: '🎖️',
      description: 'Advanced AI intelligence officer for tactical support and mission analysis',
      features: [
        'Intel Officer Persona',
        'Mission Briefings',
        'Threat Assessment',
        'Area Intelligence',
        'Weather Analysis',
        'Tactical Recommendations'
      ],
      status: 'available',
      provider: 'Anthropic',
      configurable: true,
    },
    {
      id: 'openai',
      name: 'OpenAI',
      category: 'ai',
      icon: '🤖',
      description: 'AI models for natural language processing and image analysis',
      features: [
        'GPT Text Generation',
        'Vision Analysis',
        'Embeddings',
        'Fine-tuning',
        'Assistants API',
        'Function Calling'
      ],
      status: 'available',
      provider: 'OpenAI',
      configurable: true,
    },
    {
      id: 'tensorflow',
      name: 'TensorFlow Serving',
      category: 'ai',
      icon: '🧠',
      description: 'Production-ready ML model serving for object detection and classification',
      features: [
        'Model Serving',
        'Object Detection',
        'Image Classification',
        'REST/gRPC APIs',
        'Model Versioning',
        'GPU Acceleration'
      ],
      status: 'available',
      provider: 'Google',
      configurable: true,
    },
    {
      id: 'amazon-rekognition',
      name: 'AWS Rekognition',
      category: 'ai',
      icon: '👁️',
      description: 'Computer vision service for image and video analysis',
      features: [
        'Object Detection',
        'Facial Analysis',
        'Text Extraction',
        'Activity Detection',
        'Content Moderation',
        'Custom Labels'
      ],
      status: 'available',
      provider: 'Amazon Web Services',
      configurable: true,
    },
  ]);

  // Check URL parameters on mount and listen for changes
  React.useEffect(() => {
    const checkUrlParams = () => {
      const urlParams = new URLSearchParams(window.location.search);
      const setupParam = urlParams.get('setup');
      const searchParam = urlParams.get('search');
      
      // Handle setup parameter
      if (setupParam === 'anthropic') {
        const anthropicIntegration = integrations.find(i => i.id === 'anthropic');
        if (anthropicIntegration) {
          setSelectedCategory('ai');
          handleConfigureIntegration(anthropicIntegration);
        }
      }
      
      // Handle search parameter from header
      if (searchParam && searchParam !== searchQuery) {
        setSearchQuery(searchParam);
        setSelectedCategory('all'); // Show all categories when searching
      }
    };
    
    checkUrlParams();
    
    // Listen for URL changes (when header search is used)
    const handlePopState = () => checkUrlParams();
    window.addEventListener('popstate', handlePopState);
    
    return () => window.removeEventListener('popstate', handlePopState);
  }, [integrations, searchQuery]);

  // Filter integrations
  const filteredIntegrations = integrations.filter(integration => {
    const matchesCategory = selectedCategory === 'all' || integration.category === selectedCategory;
    const matchesSearch = !searchQuery || 
      integration.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.provider.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  // Handle integration configuration
  const handleConfigureIntegration = (integration: Integration) => {
    setSelectedIntegration(integration);
    setShowConfigModal(true);
    
    // Load saved API key for Anthropic if available
    if (integration.id === 'anthropic') {
      const savedKey = localStorage.getItem('anthropic_api_key') || '';
      setAnthropicApiKey(savedKey);
    }
  };

  // Handle Vault configuration save
  const handleSaveVaultConfig = () => {
    // Save Vault configuration to localStorage
    localStorage.setItem('vaultConfig', JSON.stringify(vaultConfig));
    
    // Update integration status
    setIntegrations(prev => prev.map(int => 
      int.id === 'vault' ? { ...int, status: 'configured' } : int
    ));
    
    // Close modal
    setShowConfigModal(false);
    
    // Show success message
    console.log('Vault configuration saved successfully!', vaultConfig);
  };

  // Handle Anthropic configuration save
  const handleSaveAnthropicConfig = () => {
    // Save API key to localStorage (in production, this would be saved securely to backend)
    localStorage.setItem('anthropic_api_key', anthropicApiKey);
    
    // Update integration status
    setIntegrations(prev => prev.map(int => 
      int.id === 'anthropic' ? { ...int, status: anthropicApiKey ? 'connected' : 'available' } : int
    ));
    
    // Update environment variable for the AI service
    (window as any).VITE_ANTHROPIC_API_KEY = anthropicApiKey;
    
    setShowConfigModal(false);
    
    // Show success message
    console.log('Anthropic API key saved');
  };

  // Test connection
  const handleTestConnection = async () => {
    // In a real app, this would test the connection to Vault
    console.log('Testing Vault connection...');
    setIntegrations(prev => prev.map(int => 
      int.id === 'vault' ? { ...int, status: 'connected' } : int
    ));
  };

  // Get status badge
  const getStatusBadge = (status: Integration['status']) => {
    switch (status) {
      case 'connected':
        return { label: 'Connected', class: 'badge-success' };
      case 'configured':
        return { label: 'Configured', class: 'badge-warning' };
      case 'error':
        return { label: 'Error', class: 'badge-error' };
      default:
        return { label: 'Available', class: 'badge-default' };
    }
  };

  return (
    <div className="integrations-fullpage">
      {/* Header */}
      <header className="integrations-header">
        <div className="header-title">
          <h1>Integration Marketplace</h1>
          <div className="integration-stats">
            <span className="stat connected">{integrations.filter(i => i.status === 'connected').length} Connected</span>
            <span className="stat available">{filteredIntegrations.length} Available</span>
            <span className="stat total">{integrations.length} Total</span>
          </div>
        </div>

        <div className="header-controls">
          <div className="search-box">
            <input
              type="text"
              placeholder="Search integrations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
            />
          </div>

          <div className="filter-group">
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="filter-select"
            >
              {categories.map(category => (
                <option key={category.id} value={category.id}>
                  {category.label}
                </option>
              ))}
            </select>
          </div>

          <div className="view-toggles">
            <button
              className={`view-btn ${viewMode === 'list' ? 'active' : ''}`}
              onClick={() => setViewMode('list')}
              title="List View"
            >
              ☰
            </button>
            <button
              className={`view-btn ${viewMode === 'cards' ? 'active' : ''}`}
              onClick={() => setViewMode('cards')}
              title="Card View"
            >
              ⊞
            </button>
          </div>

          <button className="btn-primary">
            + Browse Marketplace
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div className="integrations-content">
        <div className={`integrations-display view-${viewMode}`}>
          {filteredIntegrations.length === 0 ? (
            <div className="no-integrations">
              <span className="no-integrations-icon">🔗</span>
              <h3>No integrations found</h3>
              <p>{searchQuery ? `No results for "${searchQuery}"` : 'No integrations match the current filter'}</p>
            </div>
          ) : viewMode === 'list' ? (
            <div className="integrations-list">
              <table className="integrations-table">
                <thead>
                  <tr>
                    <th>Service</th>
                    <th>Category</th>
                    <th>Provider</th>
                    <th>Status</th>
                    <th>Features</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredIntegrations.map(integration => {
                    const statusBadge = getStatusBadge(integration.status);
                    
                    return (
                      <tr key={integration.id} className="integration-row">
                        <td>
                          <div className="service-info">
                            <div className="service-icon">
                              {IntegrationLogos[integration.id] || (
                                <Icon name="package" size={24} color="var(--color-accent)" />
                              )}
                            </div>
                            <div>
                              <div className="service-name">{integration.name}</div>
                              <div className="service-description">{integration.description}</div>
                            </div>
                          </div>
                        </td>
                        <td>
                          <span className="category-tag">
                            {categories.find(c => c.id === integration.category)?.label || integration.category}
                          </span>
                        </td>
                        <td className="provider">{integration.provider}</td>
                        <td>
                          <span 
                            className="status-badge"
                            style={{ backgroundColor: statusBadge.class === 'badge-success' ? '#2ed573' : 
                                                     statusBadge.class === 'badge-warning' ? '#ffa502' : 
                                                     statusBadge.class === 'badge-error' ? '#ff6348' : '#57606f' }}
                          >
                            {statusBadge.label}
                          </span>
                        </td>
                        <td className="features">
                          {integration.features.slice(0, 2).join(', ')}
                          {integration.features.length > 2 && ` +${integration.features.length - 2} more`}
                        </td>
                        <td>
                          <div className="action-buttons">
                            {integration.status === 'connected' ? (
                              <button 
                                className="action-btn manage"
                                onClick={() => handleConfigureIntegration(integration)}
                              >
                                Manage
                              </button>
                            ) : (
                              <button 
                                className="action-btn configure"
                                onClick={() => handleConfigureIntegration(integration)}
                              >
                                Configure
                              </button>
                            )}
                            
                            {integration.docsUrl && (
                              <a 
                                href={integration.docsUrl} 
                                target="_blank" 
                                rel="noopener noreferrer"
                                className="action-btn docs"
                                title="View Documentation"
                              >
                                📚
                              </a>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="integrations-grid">
              {filteredIntegrations.map(integration => {
                const statusBadge = getStatusBadge(integration.status);
                
                return (
                  <div key={integration.id} className="integration-card">
                    <div className="card-header">
                      <span className="integration-icon">
                        {IntegrationLogos[integration.id] || (
                          <Icon name="package" size={32} color="var(--color-accent)" />
                        )}
                      </span>
                      <h3 className="integration-name">{integration.name}</h3>
                      <span 
                        className="status-indicator"
                        style={{ backgroundColor: statusBadge.class === 'badge-success' ? '#2ed573' : 
                                                 statusBadge.class === 'badge-warning' ? '#ffa502' : 
                                                 statusBadge.class === 'badge-error' ? '#ff6348' : '#57606f' }}
                        title={statusBadge.label}
                      />
                    </div>

                    <div className="card-content">
                      <div className="integration-info">
                        <span className="info-label">Provider:</span>
                        <span className="info-value">{integration.provider}</span>
                      </div>
                      <div className="integration-info">
                        <span className="info-label">Category:</span>
                        <span className="info-value">
                          {categories.find(c => c.id === integration.category)?.label || integration.category}
                        </span>
                      </div>
                      <div className="integration-description">
                        {integration.description}
                      </div>
                      <div className="integration-features">
                        {integration.features.slice(0, 3).map((feature, idx) => (
                          <span key={idx} className="feature-tag">
                            {feature}
                          </span>
                        ))}
                        {integration.features.length > 3 && (
                          <span className="feature-more">
                            +{integration.features.length - 3} more
                          </span>
                        )}
                      </div>
                    </div>

                    <div className="card-footer">
                      {integration.status === 'connected' ? (
                        <button className="btn-manage" onClick={() => handleConfigureIntegration(integration)}>
                          Manage
                        </button>
                      ) : (
                        <button className="btn-configure" onClick={() => handleConfigureIntegration(integration)}>
                          Configure
                        </button>
                      )}
                      
                      {integration.docsUrl && (
                        <a 
                          href={integration.docsUrl} 
                          target="_blank" 
                          rel="noopener noreferrer"
                          className="btn-docs"
                        >
                          Docs
                        </a>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Configuration Modal - Vault Example */}
      {showConfigModal && selectedIntegration?.id === 'vault' && (
        <div className="modal-overlay" onClick={() => setShowConfigModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>
                <span className="modal-icon">
                  {IntegrationLogos[selectedIntegration.id] || (
                    <Icon name="package" size={24} color="var(--color-accent)" />
                  )}
                </span>
                Configure {selectedIntegration.name}
              </h2>
              <button className="modal-close" onClick={() => setShowConfigModal(false)}>×</button>
            </div>

            <div className="modal-body">
              <div className="config-section">
                <h3>Connection Settings</h3>
                
                <div className="form-group">
                  <label>Vault Server URL</label>
                  <input
                    type="text"
                    value={vaultConfig.url}
                    onChange={(e) => setVaultConfig(prev => ({ ...prev, url: e.target.value }))}
                    placeholder="http://localhost:8200"
                    className="form-input"
                  />
                </div>

                <div className="form-group">
                  <label>Namespace (Enterprise)</label>
                  <input
                    type="text"
                    value={vaultConfig.namespace}
                    onChange={(e) => setVaultConfig(prev => ({ ...prev, namespace: e.target.value }))}
                    placeholder="Optional namespace"
                    className="form-input"
                  />
                </div>

                <div className="form-group">
                  <label>Authentication Method</label>
                  <select
                    value={vaultConfig.authMethod}
                    onChange={(e) => setVaultConfig(prev => ({ 
                      ...prev, 
                      authMethod: e.target.value as any,
                      oidcEnabled: e.target.value === 'oidc'
                    }))}
                    className="form-select"
                  >
                    <option value="token">Token</option>
                    <option value="userpass">Username/Password</option>
                    <option value="oidc">OIDC (OpenID Connect)</option>
                    <option value="ldap">LDAP</option>
                    <option value="kubernetes">Kubernetes</option>
                  </select>
                </div>

                {vaultConfig.authMethod === 'token' && (
                  <div className="form-group">
                    <label>Vault Token</label>
                    <input
                      type="password"
                      value={vaultConfig.token}
                      onChange={(e) => setVaultConfig(prev => ({ ...prev, token: e.target.value }))}
                      placeholder="s.xxxxxxxxxx"
                      className="form-input"
                    />
                  </div>
                )}

                {vaultConfig.authMethod === 'userpass' && (
                  <>
                    <div className="form-group">
                      <label>Username</label>
                      <input
                        type="text"
                        value={vaultConfig.username}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, username: e.target.value }))}
                        className="form-input"
                      />
                    </div>
                    <div className="form-group">
                      <label>Password</label>
                      <input
                        type="password"
                        value={vaultConfig.password}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, password: e.target.value }))}
                        className="form-input"
                      />
                    </div>
                  </>
                )}

                {vaultConfig.authMethod === 'oidc' && (
                  <div className="oidc-config">
                    <div className="form-group">
                      <label>OIDC Discovery URL</label>
                      <input
                        type="text"
                        value={vaultConfig.oidcDiscoveryUrl}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, oidcDiscoveryUrl: e.target.value }))}
                        placeholder="http://keycloak:8080/auth/realms/your-realm"
                        className="form-input"
                      />
                      <p className="form-help">The OIDC provider's discovery endpoint URL</p>
                    </div>
                    <div className="form-group">
                      <label>Client ID</label>
                      <input
                        type="text"
                        value={vaultConfig.oidcClientId}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, oidcClientId: e.target.value }))}
                        placeholder="vault-client"
                        className="form-input"
                      />
                    </div>
                    <div className="form-group">
                      <label>Client Secret</label>
                      <input
                        type="password"
                        value={vaultConfig.oidcClientSecret}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, oidcClientSecret: e.target.value }))}
                        placeholder="Client secret from OIDC provider"
                        className="form-input"
                      />
                    </div>
                    <div className="form-group">
                      <label>Default Role</label>
                      <input
                        type="text"
                        value={vaultConfig.oidcDefaultRole}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, oidcDefaultRole: e.target.value }))}
                        placeholder="management"
                        className="form-input"
                      />
                      <p className="form-help">The default Vault role to use for OIDC authentication</p>
                    </div>
                    <div className="form-group">
                      <label>Redirect URI (Optional)</label>
                      <input
                        type="text"
                        value={vaultConfig.oidcRedirectUri}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, oidcRedirectUri: e.target.value }))}
                        placeholder="https://vault.example.com:8200/ui/vault/auth/oidc/oidc/callback"
                        className="form-input"
                      />
                      <p className="form-help">Custom redirect URI if different from default</p>
                    </div>
                  </div>
                )}
              </div>

              <div className="config-section">
                <h3>Features</h3>
                
                <div className="form-group feature-toggle">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={vaultConfig.tlsEnabled}
                      onChange={(e) => setVaultConfig(prev => ({ ...prev, tlsEnabled: e.target.checked }))}
                    />
                    <span>Enable TLS Certificate Management</span>
                  </label>
                  <p className="form-help">Use Vault PKI engine for dynamic certificate generation and management</p>
                  
                  {vaultConfig.tlsEnabled && (
                    <div className="sub-config">
                      <label>PKI Engine Path</label>
                      <input
                        type="text"
                        value={vaultConfig.pkiEngine}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, pkiEngine: e.target.value }))}
                        placeholder="pki"
                        className="form-input"
                      />
                    </div>
                  )}
                </div>

                <div className="form-group feature-toggle">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={vaultConfig.encryptionEnabled}
                      onChange={(e) => setVaultConfig(prev => ({ ...prev, encryptionEnabled: e.target.checked }))}
                    />
                    <span>Enable Encryption as a Service</span>
                  </label>
                  <p className="form-help">
                    Use Vault Transit engine for end-to-end encryption of all Communications (Comms) traffic.
                    {vaultConfig.encryptionEnabled && (
                      <span className="encryption-status">
                        <span className="lock-icon">🔒</span> Comms will be encrypted
                      </span>
                    )}
                    {!vaultConfig.encryptionEnabled && (
                      <span className="encryption-status warning">
                        <span className="lock-icon">🔓</span> Comms will be unencrypted
                      </span>
                    )}
                  </p>
                  
                  {vaultConfig.encryptionEnabled && (
                    <div className="sub-config">
                      <label>Transit Engine Path</label>
                      <input
                        type="text"
                        value={vaultConfig.transitEngine}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, transitEngine: e.target.value }))}
                        placeholder="transit"
                        className="form-input"
                      />
                      <p className="form-help-sub">Transit secrets engine path for encrypting Comms messages</p>
                    </div>
                  )}
                </div>

                <div className="form-group feature-toggle">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={vaultConfig.gotakLoginIntegration}
                      onChange={(e) => setVaultConfig(prev => ({ ...prev, gotakLoginIntegration: e.target.checked }))}
                    />
                    <span>Enable GoTAK Login Integration</span>
                  </label>
                  <p className="form-help">
                    Allow users to authenticate to GoTAK using Vault credentials.
                    {vaultConfig.gotakLoginIntegration && (
                      <span className="encryption-status">
                        <span className="lock-icon">🔐</span> Vault Login option will appear on GoTAK login page
                      </span>
                    )}
                  </p>
                  
                  {vaultConfig.gotakLoginIntegration && (
                    <div className="sub-config">
                      <label>Vault Auth Path</label>
                      <input
                        type="text"
                        value={vaultConfig.gotakAuthPath}
                        onChange={(e) => setVaultConfig(prev => ({ ...prev, gotakAuthPath: e.target.value }))}
                        placeholder="auth/userpass"
                        className="form-input"
                      />
                      <p className="form-help-sub">Vault authentication path for GoTAK users (e.g., auth/userpass, auth/ldap)</p>
                      
                      <div className="integration-note">
                        <strong>Note:</strong> When enabled, users can select "Vault Login" on the GoTAK login page 
                        and authenticate using their Vault {vaultConfig.authMethod === 'userpass' ? 'username/password' : vaultConfig.authMethod.toUpperCase()} credentials.
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <div className="config-actions">
                <button className="btn-test" onClick={handleTestConnection}>
                  Test Connection
                </button>
                <div className="config-status">
                  {selectedIntegration.status === 'connected' && (
                    <span className="status-connected">✓ Connected</span>
                  )}
                </div>
              </div>
            </div>

            <div className="modal-footer">
              <button className="btn-cancel" onClick={() => setShowConfigModal(false)}>
                Cancel
              </button>
              <button className="btn-save" onClick={handleSaveVaultConfig}>
                Save Configuration
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Configuration Modal - Anthropic */}
      {showConfigModal && selectedIntegration?.id === 'anthropic' && (
        <div className="modal-overlay" onClick={() => setShowConfigModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>
                <span className="modal-icon">🎖️</span>
                Configure {selectedIntegration.name}
              </h2>
              <button className="modal-close" onClick={() => setShowConfigModal(false)}>×</button>
            </div>

            <div className="modal-body">
              <div className="config-section">
                <h3>AI Intelligence Officer Setup</h3>
                <p className="config-description">
                  Enable advanced AI capabilities for tactical support, mission briefings, and real-time intelligence analysis.
                </p>
                
                <div className="form-group">
                  <label>Anthropic API Key</label>
                  <input
                    type="password"
                    value={anthropicApiKey}
                    onChange={(e) => setAnthropicApiKey(e.target.value)}
                    placeholder="sk-ant-api03-xxxxx"
                    className="form-input"
                  />
                  <p className="form-help">
                    Get your API key from <a href="https://console.anthropic.com/" target="_blank" rel="noopener noreferrer" style={{color: 'var(--color-accent)'}}>Anthropic Console</a>
                  </p>
                </div>

                <div className="feature-highlights">
                  <h4>Enabled Features:</h4>
                  <ul className="feature-list">
                    <li>✓ AI Intelligence Officer with military persona</li>
                    <li>✓ Tactical mission briefings and analysis</li>
                    <li>✓ Real-time threat assessment</li>
                    <li>✓ Area intelligence reports</li>
                    <li>✓ Weather and terrain analysis</li>
                    <li>✓ 24/7 tactical support via chat</li>
                  </ul>
                </div>

                <div className="security-notice">
                  <Icon name="shield" size={16} color="var(--color-warning)" />
                  <span>Your API key is stored locally and never sent to our servers</span>
                </div>
              </div>

              <div className="config-actions">
                <button 
                  className="btn-test" 
                  onClick={() => {
                    if (anthropicApiKey) {
                      // Test the API key
                      console.log('Testing Anthropic connection...');
                      setIntegrations(prev => prev.map(int => 
                        int.id === 'anthropic' ? { ...int, status: 'connected' } : int
                      ));
                    }
                  }}
                  disabled={!anthropicApiKey}
                >
                  Test Connection
                </button>
                <div className="config-status">
                  {selectedIntegration.status === 'connected' && (
                    <span className="status-connected">✓ Connected</span>
                  )}
                </div>
              </div>
            </div>

            <div className="modal-footer">
              <button className="btn-cancel" onClick={() => setShowConfigModal(false)}>
                Cancel
              </button>
              <button 
                className="btn-save" 
                onClick={handleSaveAnthropicConfig}
                disabled={!anthropicApiKey}
              >
                Save Configuration
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Integrations;
