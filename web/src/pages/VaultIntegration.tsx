/**
 * HashiCorp Vault Integration Settings
 * Zero-Trust security configuration for GoTAK
 */

import React, { useState, useEffect } from 'react';
import './VaultIntegration.css';

// Types
interface VaultConfig {
  enabled: boolean;
  address: string;
  namespace?: string;
  authMethod: 'approle' | 'token' | 'oidc' | 'kubernetes';
  roleId?: string;
  secretId?: string;
  token?: string;
  tlsSkipVerify: boolean;
}

interface PKIConfig {
  enabled: boolean;
  mountPath: string;
  roleName: string;
  commonName: string;
  ttl: string;
  autoRenew: boolean;
  renewBeforeExpiry: number; // hours
}

interface TransitConfig {
  enabled: boolean;
  mountPath: string;
  keyPrefix: string;
  autoCreateKeys: boolean;
  keyRotationDays: number;
  convergentEncryption: boolean;
}

interface EntityMapping {
  enabled: boolean;
  autoCreateUsers: boolean;
  autoCreateGroups: boolean;
  defaultPolicies: string[];
}

const VaultIntegration: React.FC = () => {
  // State
  const [activeTab, setActiveTab] = useState<'connection' | 'pki' | 'transit' | 'entities'>('connection');
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connected' | 'error'>('disconnected');
  const [testResult, setTestResult] = useState<string>('');
  
  // Vault Configuration
  const [vaultConfig, setVaultConfig] = useState<VaultConfig>({
    enabled: false,
    address: 'https://vault.example.com:8200',
    namespace: '',
    authMethod: 'approle',
    roleId: '',
    secretId: '',
    token: '',
    tlsSkipVerify: false,
  });

  // PKI Configuration
  const [pkiConfig, setPkiConfig] = useState<PKIConfig>({
    enabled: false,
    mountPath: 'pki_int',
    roleName: 'gotak-server',
    commonName: 'gotak.local',
    ttl: '8760h', // 1 year
    autoRenew: true,
    renewBeforeExpiry: 720, // 30 days
  });

  // Transit Configuration  
  const [transitConfig, setTransitConfig] = useState<TransitConfig>({
    enabled: false,
    mountPath: 'transit',
    keyPrefix: 'gotak',
    autoCreateKeys: true,
    keyRotationDays: 90,
    convergentEncryption: true,
  });

  // Entity Mapping
  const [entityMapping, setEntityMapping] = useState<EntityMapping>({
    enabled: false,
    autoCreateUsers: true,
    autoCreateGroups: true,
    defaultPolicies: ['gotak-default', 'transit-user'],
  });

  // Test connection to Vault
  const testConnection = async () => {
    setTestResult('Testing connection...');
    try {
      // In production, this would call your backend API
      const response = await fetch('/api/vault/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(vaultConfig),
      });
      
      if (response.ok) {
        setConnectionStatus('connected');
        setTestResult('✅ Successfully connected to Vault');
      } else {
        setConnectionStatus('error');
        setTestResult('❌ Failed to connect to Vault');
      }
    } catch (error) {
      setConnectionStatus('error');
      setTestResult(`❌ Error: ${error}`);
    }
  };

  // Save configuration
  const saveConfiguration = async () => {
    try {
      const config = {
        vault: vaultConfig,
        pki: pkiConfig,
        transit: transitConfig,
        entities: entityMapping,
      };
      
      // In production, save to backend
      localStorage.setItem('vault-config', JSON.stringify(config));
      alert('Configuration saved successfully');
    } catch (error) {
      alert(`Failed to save configuration: ${error}`);
    }
  };

  // Initialize Transit key for user
  const initializeUserKey = async (userId: string) => {
    const keyName = `${transitConfig.keyPrefix}/users/${userId}`;
    try {
      // API call to create transit key
      console.log(`Creating transit key: ${keyName}`);
    } catch (error) {
      console.error(`Failed to create key for user ${userId}:`, error);
    }
  };

  // Initialize Transit key for group
  const initializeGroupKey = async (groupId: string) => {
    const keyName = `${transitConfig.keyPrefix}/groups/${groupId}`;
    try {
      // API call to create transit key
      console.log(`Creating transit key: ${keyName}`);
    } catch (error) {
      console.error(`Failed to create key for group ${groupId}:`, error);
    }
  };

  return (
    <div className="vault-integration">
      {/* Header */}
      <header className="vault-header">
        <div className="header-content">
          <div className="header-left">
            <h1 className="page-title">
              <span className="vault-logo">🔐</span>
              HashiCorp Vault Integration
            </h1>
            <span className="subtitle">Zero-Trust Security Configuration</span>
          </div>
          <div className="header-right">
            <div className={`connection-badge ${connectionStatus}`}>
              <span className="status-dot" />
              <span className="status-text">
                {connectionStatus === 'connected' ? 'Connected' : 
                 connectionStatus === 'error' ? 'Error' : 'Disconnected'}
              </span>
            </div>
            <button className="btn-save" onClick={saveConfiguration}>
              💾 Save Configuration
            </button>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <nav className="vault-tabs">
        <button
          className={`tab ${activeTab === 'connection' ? 'active' : ''}`}
          onClick={() => setActiveTab('connection')}
        >
          🔌 Connection
        </button>
        <button
          className={`tab ${activeTab === 'pki' ? 'active' : ''}`}
          onClick={() => setActiveTab('pki')}
        >
          🔏 PKI/TLS
        </button>
        <button
          className={`tab ${activeTab === 'transit' ? 'active' : ''}`}
          onClick={() => setActiveTab('transit')}
        >
          🔐 Transit Encryption
        </button>
        <button
          className={`tab ${activeTab === 'entities' ? 'active' : ''}`}
          onClick={() => setActiveTab('entities')}
        >
          👥 Entity Mapping
        </button>
      </nav>

      {/* Content */}
      <div className="vault-content">
        {/* Connection Tab */}
        {activeTab === 'connection' && (
          <div className="tab-content">
            <div className="config-section">
              <h2>Vault Connection Settings</h2>
              
              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={vaultConfig.enabled}
                    onChange={(e) => setVaultConfig({...vaultConfig, enabled: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Enable Vault Integration
                </label>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Vault Address</label>
                  <input
                    type="text"
                    value={vaultConfig.address}
                    onChange={(e) => setVaultConfig({...vaultConfig, address: e.target.value})}
                    placeholder="https://vault.example.com:8200"
                  />
                </div>
                <div className="form-group">
                  <label>Namespace (Optional)</label>
                  <input
                    type="text"
                    value={vaultConfig.namespace}
                    onChange={(e) => setVaultConfig({...vaultConfig, namespace: e.target.value})}
                    placeholder="admin/gotak"
                  />
                </div>
              </div>

              <div className="form-group">
                <label>Authentication Method</label>
                <select
                  value={vaultConfig.authMethod}
                  onChange={(e) => setVaultConfig({...vaultConfig, authMethod: e.target.value as any})}
                >
                  <option value="approle">AppRole (Recommended)</option>
                  <option value="token">Token</option>
                  <option value="oidc">OIDC</option>
                  <option value="kubernetes">Kubernetes</option>
                </select>
              </div>

              {vaultConfig.authMethod === 'approle' && (
                <div className="form-row">
                  <div className="form-group">
                    <label>Role ID</label>
                    <input
                      type="text"
                      value={vaultConfig.roleId}
                      onChange={(e) => setVaultConfig({...vaultConfig, roleId: e.target.value})}
                      placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                    />
                  </div>
                  <div className="form-group">
                    <label>Secret ID</label>
                    <input
                      type="password"
                      value={vaultConfig.secretId}
                      onChange={(e) => setVaultConfig({...vaultConfig, secretId: e.target.value})}
                      placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                    />
                  </div>
                </div>
              )}

              {vaultConfig.authMethod === 'token' && (
                <div className="form-group">
                  <label>Vault Token</label>
                  <input
                    type="password"
                    value={vaultConfig.token}
                    onChange={(e) => setVaultConfig({...vaultConfig, token: e.target.value})}
                    placeholder="hvs.xxxxxxxxxxxxxxxxxxxxxxxxxx"
                  />
                </div>
              )}

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={vaultConfig.tlsSkipVerify}
                    onChange={(e) => setVaultConfig({...vaultConfig, tlsSkipVerify: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Skip TLS Verification (Development Only)
                </label>
              </div>

              <div className="test-section">
                <button className="btn-test" onClick={testConnection}>
                  🧪 Test Connection
                </button>
                {testResult && (
                  <div className="test-result">{testResult}</div>
                )}
              </div>
            </div>

            {/* Vault Policy Templates */}
            <div className="config-section">
              <h2>Required Vault Policies</h2>
              <div className="policy-templates">
                <div className="policy-card">
                  <h3>GoTAK Server Policy</h3>
                  <pre className="policy-code">{`# GoTAK Server Policy
path "pki_int/issue/gotak-server" {
  capabilities = ["create", "update"]
}

path "transit/encrypt/gotak/*" {
  capabilities = ["create", "update"]
}

path "transit/decrypt/gotak/*" {
  capabilities = ["create", "update"]
}

path "transit/keys/gotak/*" {
  capabilities = ["create", "read", "list"]
}

path "identity/entity/*" {
  capabilities = ["create", "read", "update", "list"]
}

path "identity/group/*" {
  capabilities = ["create", "read", "update", "list"]
}`}</pre>
                  <button className="btn-copy">📋 Copy Policy</button>
                </div>

                <div className="policy-card">
                  <h3>User Transit Policy</h3>
                  <pre className="policy-code">{`# User Transit Encryption Policy
path "transit/encrypt/gotak/users/{{identity.entity.id}}" {
  capabilities = ["create", "update"]
}

path "transit/decrypt/gotak/users/{{identity.entity.id}}" {
  capabilities = ["create", "update"]
}

path "transit/encrypt/gotak/groups/*" {
  capabilities = ["create", "update"]
}

path "transit/decrypt/gotak/groups/*" {
  capabilities = ["create", "update"]
}`}</pre>
                  <button className="btn-copy">📋 Copy Policy</button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* PKI/TLS Tab */}
        {activeTab === 'pki' && (
          <div className="tab-content">
            <div className="config-section">
              <h2>PKI Certificate Management</h2>
              
              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={pkiConfig.enabled}
                    onChange={(e) => setPkiConfig({...pkiConfig, enabled: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Enable PKI Integration
                </label>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>PKI Mount Path</label>
                  <input
                    type="text"
                    value={pkiConfig.mountPath}
                    onChange={(e) => setPkiConfig({...pkiConfig, mountPath: e.target.value})}
                    placeholder="pki_int"
                  />
                </div>
                <div className="form-group">
                  <label>Role Name</label>
                  <input
                    type="text"
                    value={pkiConfig.roleName}
                    onChange={(e) => setPkiConfig({...pkiConfig, roleName: e.target.value})}
                    placeholder="gotak-server"
                  />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Common Name</label>
                  <input
                    type="text"
                    value={pkiConfig.commonName}
                    onChange={(e) => setPkiConfig({...pkiConfig, commonName: e.target.value})}
                    placeholder="gotak.local"
                  />
                </div>
                <div className="form-group">
                  <label>Certificate TTL</label>
                  <input
                    type="text"
                    value={pkiConfig.ttl}
                    onChange={(e) => setPkiConfig({...pkiConfig, ttl: e.target.value})}
                    placeholder="8760h"
                  />
                </div>
              </div>

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={pkiConfig.autoRenew}
                    onChange={(e) => setPkiConfig({...pkiConfig, autoRenew: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Auto-Renew Certificates
                </label>
              </div>

              {pkiConfig.autoRenew && (
                <div className="form-group">
                  <label>Renew Before Expiry (hours)</label>
                  <input
                    type="number"
                    value={pkiConfig.renewBeforeExpiry}
                    onChange={(e) => setPkiConfig({...pkiConfig, renewBeforeExpiry: parseInt(e.target.value)})}
                  />
                </div>
              )}

              <div className="info-box">
                <span className="info-icon">ℹ️</span>
                <div>
                  <strong>Certificate Chain:</strong>
                  <ul>
                    <li>Root CA → Intermediate CA → Server Certificate</li>
                    <li>Automatic rotation before expiry</li>
                    <li>OCSP responder integration</li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Certificate Status */}
            <div className="config-section">
              <h2>Current Certificate Status</h2>
              <div className="cert-status">
                <div className="status-item">
                  <span className="label">Common Name:</span>
                  <span className="value">gotak.local</span>
                </div>
                <div className="status-item">
                  <span className="label">Serial Number:</span>
                  <span className="value">4a:7c:3d:89:15:2b:6e:9f</span>
                </div>
                <div className="status-item">
                  <span className="label">Valid From:</span>
                  <span className="value">2024-01-01 00:00:00 UTC</span>
                </div>
                <div className="status-item">
                  <span className="label">Valid Until:</span>
                  <span className="value">2025-01-01 00:00:00 UTC</span>
                </div>
                <div className="status-item">
                  <span className="label">Days Remaining:</span>
                  <span className="value success">352 days</span>
                </div>
              </div>
              <button className="btn-action">🔄 Renew Certificate Now</button>
            </div>
          </div>
        )}

        {/* Transit Encryption Tab */}
        {activeTab === 'transit' && (
          <div className="tab-content">
            <div className="config-section">
              <h2>Transit Encryption Settings</h2>
              
              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={transitConfig.enabled}
                    onChange={(e) => setTransitConfig({...transitConfig, enabled: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Enable Transit Encryption
                </label>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Transit Mount Path</label>
                  <input
                    type="text"
                    value={transitConfig.mountPath}
                    onChange={(e) => setTransitConfig({...transitConfig, mountPath: e.target.value})}
                    placeholder="transit"
                  />
                </div>
                <div className="form-group">
                  <label>Key Prefix</label>
                  <input
                    type="text"
                    value={transitConfig.keyPrefix}
                    onChange={(e) => setTransitConfig({...transitConfig, keyPrefix: e.target.value})}
                    placeholder="gotak"
                  />
                </div>
              </div>

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={transitConfig.autoCreateKeys}
                    onChange={(e) => setTransitConfig({...transitConfig, autoCreateKeys: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Auto-Create Keys for New Users/Groups
                </label>
              </div>

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={transitConfig.convergentEncryption}
                    onChange={(e) => setTransitConfig({...transitConfig, convergentEncryption: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Enable Convergent Encryption (Deduplication)
                </label>
              </div>

              <div className="form-group">
                <label>Key Rotation Period (days)</label>
                <input
                  type="number"
                  value={transitConfig.keyRotationDays}
                  onChange={(e) => setTransitConfig({...transitConfig, keyRotationDays: parseInt(e.target.value)})}
                />
              </div>

              <div className="info-box">
                <span className="info-icon">🔐</span>
                <div>
                  <strong>Zero-Trust Key Hierarchy:</strong>
                  <ul>
                    <li><strong>User Keys:</strong> {transitConfig.keyPrefix}/users/[user-id]</li>
                    <li><strong>Group Keys:</strong> {transitConfig.keyPrefix}/groups/[group-id]</li>
                    <li><strong>Session Keys:</strong> Ephemeral, derived from user keys</li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Key Management */}
            <div className="config-section">
              <h2>Key Management</h2>
              <div className="key-stats">
                <div className="stat-card">
                  <div className="stat-value">42</div>
                  <div className="stat-label">Active User Keys</div>
                </div>
                <div className="stat-card">
                  <div className="stat-value">8</div>
                  <div className="stat-label">Group Keys</div>
                </div>
                <div className="stat-card">
                  <div className="stat-value">3</div>
                  <div className="stat-label">Keys Pending Rotation</div>
                </div>
                <div className="stat-card">
                  <div className="stat-value">0</div>
                  <div className="stat-label">Compromised Keys</div>
                </div>
              </div>
              
              <div className="action-buttons">
                <button className="btn-action">🔄 Rotate All Keys</button>
                <button className="btn-action">🗑️ Purge Inactive Keys</button>
                <button className="btn-action">📊 View Key Usage Stats</button>
              </div>
            </div>
          </div>
        )}

        {/* Entity Mapping Tab */}
        {activeTab === 'entities' && (
          <div className="tab-content">
            <div className="config-section">
              <h2>Entity & Group Mapping</h2>
              
              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={entityMapping.enabled}
                    onChange={(e) => setEntityMapping({...entityMapping, enabled: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Enable Automatic Entity Mapping
                </label>
              </div>

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={entityMapping.autoCreateUsers}
                    onChange={(e) => setEntityMapping({...entityMapping, autoCreateUsers: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Auto-Create Vault Entities for New Users
                </label>
              </div>

              <div className="form-group">
                <label className="toggle-label">
                  <input
                    type="checkbox"
                    checked={entityMapping.autoCreateGroups}
                    onChange={(e) => setEntityMapping({...entityMapping, autoCreateGroups: e.target.checked})}
                  />
                  <span className="toggle-slider" />
                  Auto-Create Vault Groups for Teams
                </label>
              </div>

              <div className="form-group">
                <label>Default User Policies</label>
                <div className="policy-list">
                  {entityMapping.defaultPolicies.map((policy, index) => (
                    <div key={index} className="policy-item">
                      <input
                        type="text"
                        value={policy}
                        onChange={(e) => {
                          const newPolicies = [...entityMapping.defaultPolicies];
                          newPolicies[index] = e.target.value;
                          setEntityMapping({...entityMapping, defaultPolicies: newPolicies});
                        }}
                      />
                      <button
                        className="btn-remove"
                        onClick={() => {
                          const newPolicies = entityMapping.defaultPolicies.filter((_, i) => i !== index);
                          setEntityMapping({...entityMapping, defaultPolicies: newPolicies});
                        }}
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                  <button
                    className="btn-add"
                    onClick={() => {
                      setEntityMapping({
                        ...entityMapping,
                        defaultPolicies: [...entityMapping.defaultPolicies, '']
                      });
                    }}
                  >
                    + Add Policy
                  </button>
                </div>
              </div>
            </div>

            {/* Entity Sync Status */}
            <div className="config-section">
              <h2>Entity Synchronization Status</h2>
              <div className="sync-status">
                <div className="sync-item">
                  <span className="sync-label">GoTAK Users:</span>
                  <span className="sync-value">42</span>
                </div>
                <div className="sync-item">
                  <span className="sync-label">Vault Entities:</span>
                  <span className="sync-value">42</span>
                  <span className="sync-badge success">✓ Synced</span>
                </div>
                <div className="sync-item">
                  <span className="sync-label">GoTAK Groups:</span>
                  <span className="sync-value">8</span>
                </div>
                <div className="sync-item">
                  <span className="sync-label">Vault Groups:</span>
                  <span className="sync-value">8</span>
                  <span className="sync-badge success">✓ Synced</span>
                </div>
                <div className="sync-item">
                  <span className="sync-label">Last Sync:</span>
                  <span className="sync-value">2 minutes ago</span>
                </div>
              </div>
              
              <div className="action-buttons">
                <button className="btn-action">🔄 Sync Now</button>
                <button className="btn-action">🔍 View Sync Log</button>
                <button className="btn-action">⚠️ Resolve Conflicts</button>
              </div>
            </div>

            {/* Security Recommendations */}
            <div className="config-section">
              <h2>Zero-Trust Security Recommendations</h2>
              <div className="recommendations">
                <div className="recommendation">
                  <span className="rec-icon">✅</span>
                  <div className="rec-content">
                    <strong>Per-User Encryption Keys</strong>
                    <p>Each user has a unique transit key for end-to-end encryption</p>
                  </div>
                </div>
                <div className="recommendation">
                  <span className="rec-icon">✅</span>
                  <div className="rec-content">
                    <strong>Group Key Derivation</strong>
                    <p>Group keys are derived and shared only with authorized members</p>
                  </div>
                </div>
                <div className="recommendation">
                  <span className="rec-icon">⚠️</span>
                  <div className="rec-content">
                    <strong>Enable Key Rotation</strong>
                    <p>Configure automatic key rotation every 90 days</p>
                  </div>
                </div>
                <div className="recommendation">
                  <span className="rec-icon">ℹ️</span>
                  <div className="rec-content">
                    <strong>Audit Logging</strong>
                    <p>All encryption operations are logged in Vault audit logs</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default VaultIntegration;
