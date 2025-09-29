import React, { useState } from 'react';
import { useRouter } from '../utils/router';
import { Icon } from '../components/ui/Icon';
import './Login.css';

const Login: React.FC = () => {
  const router = useRouter();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    rememberMe: false
  });
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [loginMethod, setLoginMethod] = useState<'standard' | 'vault'>('standard');
  const [vaultConfig, setVaultConfig] = useState<any>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
    // Clear error when user starts typing
    if (error) setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    // Validate input
    if (!formData.username || !formData.password) {
      setError('Please enter both username and password');
      setIsLoading(false);
      return;
    }

    // Check login method
    if (loginMethod === 'vault' && vaultConfig) {
      // Vault authentication
      try {
        const vaultUrl = vaultConfig.url.replace(/\/$/, '');
        const authPath = vaultConfig.gotakAuthPath || 'auth/userpass';
        
        const response = await fetch(`${vaultUrl}/v1/${authPath}/login/${formData.username}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            password: formData.password
          }),
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('authToken', data.auth.client_token);
          localStorage.setItem('authMethod', 'vault');
          if (formData.rememberMe) {
            localStorage.setItem('rememberUsername', formData.username);
          }
          router.navigate('/');
          window.location.href = '/';
        } else {
          setError('Invalid Vault credentials');
        }
      } catch (err) {
        console.error('Vault auth error:', err);
        setError('Failed to connect to Vault server');
      }
    } else if (formData.username === 'admin' && formData.password === 'admin') {
      // Standard demo authentication
      await new Promise(resolve => setTimeout(resolve, 500));
      
      localStorage.setItem('authToken', 'demo-token');
      localStorage.setItem('authMethod', 'standard');
      if (formData.rememberMe) {
        localStorage.setItem('rememberUsername', formData.username);
      } else {
        localStorage.removeItem('rememberUsername');
      }
      router.navigate('/');
      window.location.href = '/';
    } else {
      // Try actual API if available
      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: formData.username,
            password: formData.password,
            rememberMe: formData.rememberMe
          }),
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('authToken', data.token);
          if (formData.rememberMe) {
            localStorage.setItem('rememberUsername', formData.username);
          } else {
            localStorage.removeItem('rememberUsername');
          }
          router.navigate('/');
          window.location.href = '/';
        } else {
          setError('Invalid username or password');
        }
      } catch (err) {
        // If API fails, show error
        setError('Invalid username or password');
      }
    }
    
    setIsLoading(false);
  };

  // Check for remembered username and vault config on mount
  React.useEffect(() => {
    const rememberedUsername = localStorage.getItem('rememberUsername');
    if (rememberedUsername) {
      setFormData(prev => ({
        ...prev,
        username: rememberedUsername,
        rememberMe: true
      }));
    }
    
    // Check if Vault integration is enabled
    const savedVaultConfig = localStorage.getItem('vaultConfig');
    if (savedVaultConfig) {
      try {
        const config = JSON.parse(savedVaultConfig);
        if (config.gotakLoginIntegration) {
          setVaultConfig(config);
        }
      } catch (e) {
        console.error('Failed to parse vault config', e);
      }
    }
  }, []);

  return (
    <div className="login-container">
      <div className="login-background">
        <div className="login-background-pattern"></div>
        <div className="login-background-gradient"></div>
      </div>

      <div className="login-content">
        <div className="login-card">
          <div className="login-header">
            <div className="login-logo">
              <div className="logo-icon">
                <Icon name="shield" className="logo-shield" size={32} />
                <Icon name="radio" className="logo-radio" size={24} />
              </div>
              <h1 className="logo-text">GoTAK</h1>
            </div>
            <p className="login-subtitle">Team Awareness Kit Server</p>
          </div>

          <form className="login-form" onSubmit={handleSubmit}>
            {/* Login Method Selector - Only show if Vault is configured */}
            {vaultConfig && (
              <div className="login-method-selector">
                <button
                  type="button"
                  className={`method-btn ${loginMethod === 'standard' ? 'active' : ''}`}
                  onClick={() => setLoginMethod('standard')}
                  disabled={isLoading}
                >
                  <Icon name="shield" size={18} />
                  <span>Standard Login</span>
                </button>
                <button
                  type="button"
                  className={`method-btn ${loginMethod === 'vault' ? 'active' : ''}`}
                  onClick={() => setLoginMethod('vault')}
                  disabled={isLoading}
                >
                  <Icon name="lock" size={18} />
                  <span>Vault Login</span>
                </button>
              </div>
            )}
            
            {error && (
              <div className="login-error">
                <Icon name="alert-circle" size={16} />
                <span>{error}</span>
              </div>
            )}

            <div className="form-group">
              <label htmlFor="username">
                {loginMethod === 'vault' ? 'Vault Username' : 'Username'}
              </label>
              <div className="input-wrapper">
                <Icon name="users" className="input-icon" size={18} />
                <input
                  type="text"
                  id="username"
                  name="username"
                  value={formData.username}
                  onChange={handleInputChange}
                  placeholder={loginMethod === 'vault' ? 'Enter your Vault username' : 'Enter your username'}
                  autoComplete="username"
                  autoFocus
                  disabled={isLoading}
                />
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="password">
                {loginMethod === 'vault' ? 'Vault Password' : 'Password'}
              </label>
              <div className="input-wrapper">
                <Icon name="shield" className="input-icon" size={18} />
                <input
                  type={showPassword ? 'text' : 'password'}
                  id="password"
                  name="password"
                  value={formData.password}
                  onChange={handleInputChange}
                  placeholder={loginMethod === 'vault' ? 'Enter your Vault password' : 'Enter your password'}
                  autoComplete="current-password"
                  disabled={isLoading}
                />
                <button
                  type="button"
                  className="password-toggle"
                  onClick={() => setShowPassword(!showPassword)}
                  tabIndex={-1}
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? <Icon name="eye-off" size={18} /> : <Icon name="eye" size={18} />}
                </button>
              </div>
            </div>

            <div className="form-options">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="rememberMe"
                  checked={formData.rememberMe}
                  onChange={handleInputChange}
                  disabled={isLoading}
                />
                <span>Remember me</span>
              </label>
              <a href="#" className="forgot-link" onClick={(e) => e.preventDefault()}>
                Forgot password?
              </a>
            </div>

            <button type="submit" className="login-button" disabled={isLoading}>
              {isLoading ? (
                <>
                  <span className="spinner"></span>
                  Authenticating...
                </>
              ) : (
                'Sign In'
              )}
            </button>
          </form>

          <div className="login-features">
            <div className="feature">
              <Icon name="map-pin" size={16} />
              <span>Real-time Tracking</span>
            </div>
            <div className="feature">
              <Icon name="users" size={16} />
              <span>Team Coordination</span>
            </div>
            <div className="feature">
              <Icon name="shield" size={16} />
              <span>Secure Communications</span>
            </div>
          </div>

          <div className="login-footer">
            <p>© 2024 GoTAK Server. All rights reserved.</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;