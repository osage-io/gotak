# GoTAK Security Hardening Guide

This document provides comprehensive security hardening recommendations and configurations for production deployment of GoTAK.

## Table of Contents

- [Network Security](#network-security)
- [TLS/SSL Configuration](#tlsssl-configuration)
- [Authentication & Authorization](#authentication--authorization)
- [Input Validation & Sanitization](#input-validation--sanitization)
- [HTTP Security Headers](#http-security-headers)
- [Database Security](#database-security)
- [Container Security](#container-security)
- [Monitoring & Logging](#monitoring--logging)
- [Infrastructure Security](#infrastructure-security)
- [Compliance & Standards](#compliance--standards)

## Network Security

### Firewall Configuration

#### iptables Rules (Linux)
```bash
# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# SSH (restrict to management network)
iptables -A INPUT -p tcp --dport 22 -s 10.0.0.0/8 -j ACCEPT

# GoTAK services
iptables -A INPUT -p tcp --dport 8087 -j ACCEPT  # TAK TCP
iptables -A INPUT -p udp --dport 8087 -j ACCEPT  # TAK UDP
iptables -A INPUT -p tcp --dport 8089 -j ACCEPT  # TAK TLS
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT  # Web interface

# Rate limiting for authentication endpoints
iptables -A INPUT -p tcp --dport 8080 -m limit --limit 10/min -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
```

#### UFW Configuration (Ubuntu)
```bash
# Reset and default policies
ufw --force reset
ufw default deny incoming
ufw default allow outgoing

# SSH with rate limiting
ufw limit ssh

# GoTAK services
ufw allow 8087
ufw allow 8089
ufw allow 8080/tcp

# Enable firewall
ufw enable
```

### Network Segmentation

Create separate network zones:
- **DMZ**: Public-facing services (web interface)
- **Application Tier**: GoTAK server instances
- **Database Tier**: PostgreSQL instances
- **Management**: Administrative access

## TLS/SSL Configuration

### Certificate Management

#### Generate Production Certificates
```bash
# Create CA certificate
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -days 365 -key ca-key.pem -out ca.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=GoTAK-CA"

# Generate server certificate
openssl genrsa -out server-key.pem 4096
openssl req -new -key server-key.pem -out server.csr \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=gotak.example.com"

# Sign server certificate
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca-key.pem \
    -CAcreateserial -out server.pem -days 365

# Set secure permissions
chmod 600 *-key.pem
chmod 644 *.pem
```

#### TLS Configuration (server.yaml)
```yaml
security:
  tls:
    enabled: true
    cert_file: "/etc/gotak/certs/server.pem"
    key_file: "/etc/gotak/certs/server-key.pem"
    ca_file: "/etc/gotak/certs/ca.pem"
    client_auth: true  # Require client certificates
    min_version: "1.3"  # TLS 1.3 only
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"
      - "TLS_AES_128_GCM_SHA256"
    curve_preferences:
      - "CurveP384"
      - "CurveP256"
```

### NGINX TLS Proxy Configuration
```nginx
server {
    listen 443 ssl http2;
    server_name gotak.example.com;
    
    # TLS Configuration
    ssl_certificate /etc/nginx/ssl/server.pem;
    ssl_certificate_key /etc/nginx/ssl/server-key.pem;
    ssl_client_certificate /etc/nginx/ssl/ca.pem;
    ssl_verify_client optional;
    
    ssl_protocols TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    
    # Security Headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'" always;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;
    
    location /api/auth/login {
        limit_req zone=login burst=5 nodelay;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Authentication & Authorization

### Strong Authentication Configuration
```yaml
auth:
  methods:
    certificate:
      enabled: true
      require_client_cert: true
      ca_file: "/etc/gotak/certs/ca.pem"
      verify_depth: 3
    
    local:
      enabled: true
      password_policy:
        min_length: 12
        require_uppercase: true
        require_lowercase: true
        require_numbers: true
        require_symbols: true
        max_age_days: 90
        history_count: 5
      
      lockout_policy:
        max_attempts: 5
        lockout_duration: "15m"
        reset_time: "1h"
    
    ldap:
      enabled: false  # Enable if using LDAP
      server: "ldaps://ldap.example.com:636"
      base_dn: "dc=example,dc=com"
      bind_dn: "cn=gotak,ou=service,dc=example,dc=com"
      user_filter: "(&(objectClass=person)(uid=%s))"
      group_filter: "(&(objectClass=group)(member=%s))"

  session:
    timeout: "4h"
    max_sessions_per_user: 3
    secure_cookies: true
    same_site: "Strict"
```

### Role-Based Access Control (RBAC)
```yaml
authorization:
  roles:
    admin:
      permissions:
        - "system:*"
        - "users:*"
        - "routes:*"
        - "logs:read"
    
    operator:
      permissions:
        - "routes:read"
        - "routes:create"
        - "routes:update"
        - "users:read"
    
    observer:
      permissions:
        - "routes:read"
        - "users:read"
  
  default_role: "observer"
  
  groups:
    tactical_operations:
      roles: ["operator"]
      ldap_groups: ["cn=tactical,ou=groups,dc=example,dc=com"]
    
    system_administrators:
      roles: ["admin"]
      ldap_groups: ["cn=admins,ou=groups,dc=example,dc=com"]
```

## Input Validation & Sanitization

### API Input Validation
```go
// Example validation middleware
func ValidateInput() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // Content-Type validation
        if !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
            c.JSON(400, gin.H{"error": "Invalid content type"})
            c.Abort()
            return
        }
        
        // Request size limiting
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024) // 1MB limit
        
        // XSS prevention
        for key, values := range c.Request.Header {
            for _, value := range values {
                if containsXSS(value) {
                    c.JSON(400, gin.H{"error": "Invalid header content"})
                    c.Abort()
                    return
                }
            }
        }
        
        c.Next()
    })
}

func containsXSS(input string) bool {
    patterns := []string{
        `<script`,
        `javascript:`,
        `on\w+\s*=`,
        `<iframe`,
        `<object`,
        `<embed`,
    }
    
    lower := strings.ToLower(input)
    for _, pattern := range patterns {
        if matched, _ := regexp.MatchString(pattern, lower); matched {
            return true
        }
    }
    return false
}
```

### CoT Message Validation
```yaml
cot:
  validation:
    max_message_size: 65536  # 64KB
    max_elements: 100
    allowed_types:
      - "a-f-*"  # Friendly forces
      - "a-h-*"  # Hostile forces
      - "b-t-f"  # Chat
      - "t-x-*"  # System
    
    sanitization:
      strip_html: true
      escape_xml: true
      validate_coordinates: true
      max_string_length: 1024
    
    blacklist_patterns:
      - "<script"
      - "javascript:"
      - "data:text/html"
      - "vbscript:"
```

## HTTP Security Headers

### Security Headers Configuration
```yaml
server:
  security_headers:
    strict_transport_security:
      enabled: true
      max_age: 31536000  # 1 year
      include_subdomains: true
      preload: true
    
    content_security_policy:
      enabled: true
      policy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss://gotak.example.com"
    
    x_frame_options:
      enabled: true
      value: "DENY"
    
    x_content_type_options:
      enabled: true
      value: "nosniff"
    
    x_xss_protection:
      enabled: true
      value: "1; mode=block"
    
    referrer_policy:
      enabled: true
      value: "strict-origin-when-cross-origin"
    
    permissions_policy:
      enabled: true
      value: "geolocation=(), microphone=(), camera=()"
    
    hide_server_info: true
```

## Database Security

### PostgreSQL Hardening
```sql
-- Create dedicated user with minimal privileges
CREATE USER gotak_app WITH PASSWORD 'secure_password_here';

-- Create database and grant minimal permissions
CREATE DATABASE gotak OWNER gotak_app;
GRANT CONNECT ON DATABASE gotak TO gotak_app;
GRANT USAGE ON SCHEMA public TO gotak_app;
GRANT CREATE ON SCHEMA public TO gotak_app;

-- Revoke unnecessary privileges
REVOKE ALL ON pg_user FROM public;
REVOKE ALL ON pg_group FROM public;
REVOKE ALL ON pg_authid FROM public;
REVOKE ALL ON pg_auth_members FROM public;
REVOKE ALL ON pg_database FROM public;
REVOKE ALL ON pg_tablespace FROM public;
REVOKE ALL ON pg_settings FROM public;
```

### PostgreSQL Configuration (postgresql.conf)
```ini
# Connection security
ssl = on
ssl_cert_file = '/etc/postgresql/ssl/server.crt'
ssl_key_file = '/etc/postgresql/ssl/server.key'
ssl_ca_file = '/etc/postgresql/ssl/ca.crt'

# Authentication
password_encryption = scram-sha-256

# Logging
log_connections = on
log_disconnections = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_min_messages = warning
log_min_error_statement = error

# Resource limits
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
```

### Database Connection Security
```yaml
database:
  postgres:
    host: "localhost"
    port: 5432
    database: "gotak"
    username: "gotak_app"
    password: "${DB_PASSWORD}"  # From environment variable
    
    ssl_mode: "require"
    ssl_cert_file: "/etc/gotak/certs/client.pem"
    ssl_key_file: "/etc/gotak/certs/client-key.pem"
    ssl_ca_file: "/etc/gotak/certs/ca.pem"
    
    connection_pool:
      max_open: 25
      max_idle: 10
      max_lifetime: "1h"
    
    query_timeout: "30s"
    transaction_timeout: "5m"
```

## Container Security

### Secure Dockerfile
```dockerfile
# Use specific version, not latest
FROM golang:1.21-alpine3.18 AS builder

# Create non-root user
RUN addgroup -g 1001 -S gotak && \
    adduser -u 1001 -S gotak -G gotak

# Install security updates
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gotak-server ./cmd/gotak-server

# Runtime image
FROM scratch

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy user from builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy binary
COPY --from=builder /app/gotak-server /usr/local/bin/

# Use non-root user
USER gotak:gotak

# Security settings
EXPOSE 8080 8087 8089
ENTRYPOINT ["/usr/local/bin/gotak-server"]
```

### Docker Compose Security
```yaml
version: '3.8'

services:
  gotak:
    build: .
    container_name: gotak-server
    restart: unless-stopped
    
    # Security options
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
    read_only: true
    
    # Temporary filesystem for writable areas
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
      - /var/run:noexec,nosuid,size=100m
    
    # Resource limits
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
    
    # User namespace
    user: "1001:1001"
    
    # Environment variables from secrets
    env_file:
      - .env.production
    
    ports:
      - "8080:8080"
      - "8087:8087"
      - "8089:8089"
    
    volumes:
      - gotak_config:/etc/gotak:ro
      - gotak_certs:/etc/gotak/certs:ro
      - gotak_logs:/var/log/gotak

volumes:
  gotak_config:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/gotak/config
  
  gotak_certs:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/gotak/certs
  
  gotak_logs:
    driver: local
```

## Monitoring & Logging

### Comprehensive Logging Configuration
```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/gotak/server.log"
  
  audit:
    enabled: true
    file: "/var/log/gotak/audit.log"
    events:
      - "authentication"
      - "authorization"
      - "configuration_changes"
      - "user_management"
      - "route_management"
    
  security:
    enabled: true
    file: "/var/log/gotak/security.log"
    events:
      - "failed_auth"
      - "brute_force"
      - "suspicious_activity"
      - "privilege_escalation"
    
  rotation:
    max_size: 100  # MB
    max_backups: 10
    max_age: 30  # days
    compress: true
```

### Security Monitoring Script
```bash
#!/bin/bash
# Security monitoring and alerting

SECURITY_LOG="/var/log/gotak/security.log"
ALERT_EMAIL="security@example.com"
THRESHOLD_FAILED_AUTH=5
THRESHOLD_TIMEFRAME=300  # 5 minutes

# Monitor failed authentication attempts
check_failed_auth() {
    local count
    count=$(grep -c "failed_auth" "$SECURITY_LOG" | tail -n "$THRESHOLD_TIMEFRAME" | wc -l)
    
    if [ "$count" -gt "$THRESHOLD_FAILED_AUTH" ]; then
        echo "SECURITY ALERT: $count failed authentication attempts in last 5 minutes" | \
        mail -s "GoTAK Security Alert" "$ALERT_EMAIL"
    fi
}

# Monitor suspicious patterns
check_suspicious_activity() {
    # Check for SQL injection attempts
    if grep -q "sql injection" "$SECURITY_LOG"; then
        echo "SECURITY ALERT: SQL injection attempt detected" | \
        mail -s "GoTAK Critical Security Alert" "$ALERT_EMAIL"
    fi
    
    # Check for XSS attempts
    if grep -q "xss attempt" "$SECURITY_LOG"; then
        echo "SECURITY ALERT: XSS attempt detected" | \
        mail -s "GoTAK Security Alert" "$ALERT_EMAIL"
    fi
}

# Run checks
check_failed_auth
check_suspicious_activity
```

### Prometheus Metrics
```yaml
metrics:
  prometheus:
    enabled: true
    path: "/metrics"
    port: 9090
    
    custom_metrics:
      - name: "gotak_authentication_failures_total"
        help: "Total number of authentication failures"
        type: "counter"
        labels: ["source_ip", "username"]
      
      - name: "gotak_security_events_total"
        help: "Total number of security events"
        type: "counter"
        labels: ["event_type", "severity"]
      
      - name: "gotak_active_connections"
        help: "Number of active client connections"
        type: "gauge"
        labels: ["connection_type"]
```

## Infrastructure Security

### System Hardening Checklist

#### OS-Level Security
```bash
# Disable unnecessary services
systemctl disable bluetooth
systemctl disable cups
systemctl disable avahi-daemon

# Set secure file permissions
chmod 700 /root
chmod 600 /etc/ssh/sshd_config
chmod 600 /etc/crontab
chmod 600 /etc/shadow

# Configure SSH hardening
echo "
Protocol 2
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
PermitEmptyPasswords no
X11Forwarding no
MaxAuthTries 3
ClientAliveInterval 300
ClientAliveCountMax 2
" >> /etc/ssh/sshd_config

# Install fail2ban
apt-get install fail2ban
cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
```

#### Kernel Security
```bash
# Sysctl security settings
cat >> /etc/sysctl.conf << EOF
# IP Spoofing protection
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP redirects
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.all.secure_redirects = 0
net.ipv6.conf.all.accept_redirects = 0

# Ignore send redirects
net.ipv4.conf.all.send_redirects = 0

# Ignore ICMP ping requests
net.ipv4.icmp_echo_ignore_all = 1

# Log Martians
net.ipv4.conf.all.log_martians = 1

# Ignore source routed packets
net.ipv4.conf.all.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0

# TCP SYN flood protection
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_synack_retries = 2

# Disable IPv6 if not needed
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
EOF

sysctl -p
```

## Compliance & Standards

### Security Standards Alignment

#### NIST Cybersecurity Framework
- **Identify**: Asset inventory, risk assessment
- **Protect**: Access controls, data security, protective technology
- **Detect**: Monitoring, detection processes
- **Respond**: Incident response planning, communications
- **Recover**: Recovery planning, improvements

#### OWASP Top 10 Mitigation
1. **Injection**: Input validation, parameterized queries
2. **Broken Authentication**: Strong authentication, session management
3. **Sensitive Data Exposure**: Encryption, secure transmission
4. **XML External Entities**: Disable XML external entity processing
5. **Broken Access Control**: Implement RBAC, principle of least privilege
6. **Security Misconfiguration**: Security hardening, regular updates
7. **Cross-Site Scripting**: Input validation, output encoding
8. **Insecure Deserialization**: Validate serialized data
9. **Components with Known Vulnerabilities**: Regular updates, dependency scanning
10. **Insufficient Logging**: Comprehensive audit logging

### Regular Security Tasks

#### Daily
- Monitor security logs and alerts
- Review authentication failures
- Check system resource usage
- Verify backup integrity

#### Weekly
- Review user access and permissions
- Update security patches
- Analyze security metrics
- Test incident response procedures

#### Monthly
- Conduct vulnerability scans
- Review and update security policies
- Perform access control audits
- Update threat intelligence

#### Quarterly
- Penetration testing
- Security architecture review
- Incident response plan testing
- Compliance assessment

## Emergency Response

### Security Incident Response Plan

#### Phase 1: Detection and Analysis
1. Monitor security alerts and logs
2. Validate and classify incidents
3. Determine scope and impact
4. Activate incident response team

#### Phase 2: Containment
1. Isolate affected systems
2. Preserve evidence
3. Implement temporary fixes
4. Document all actions

#### Phase 3: Eradication and Recovery
1. Remove malicious components
2. Apply security patches
3. Restore systems from clean backups
4. Monitor for signs of compromise

#### Phase 4: Post-Incident Activity
1. Document lessons learned
2. Update security procedures
3. Conduct root cause analysis
4. Implement preventive measures

### Contact Information
```yaml
incident_response:
  team:
    security_lead: "security@example.com"
    system_admin: "admin@example.com"
    manager: "manager@example.com"
  
  external:
    cert_team: "+1-888-282-0870"
    legal: "legal@example.com"
    pr_team: "pr@example.com"
```

This security hardening guide provides comprehensive protection measures for GoTAK deployments. Regular reviews and updates of these configurations are essential to maintain security posture against evolving threats.
