#!/bin/bash
# GoTAK Security Audit Script
# Comprehensive security assessment including vulnerability scanning, configuration hardening, and compliance checks
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_ROOT/test-reports/security"
TEMP_DIR="/tmp/gotak-security-$$"

# Environment defaults
GOTAK_BASE_URL="${GOTAK_BASE_URL:-http://localhost:8080}"
GOTAK_WS_URL="${GOTAK_WS_URL:-ws://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Security severity levels
CRITICAL=0
HIGH=0
MEDIUM=0
LOW=0
INFO=0

# Logging functions
log() {
    echo -e "${BLUE}[SECURITY]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

critical() {
    echo -e "${RED}[CRITICAL]${NC} $1" >&2
    CRITICAL=$((CRITICAL + 1))
}

high() {
    echo -e "${PURPLE}[HIGH]${NC} $1"
    HIGH=$((HIGH + 1))
}

medium() {
    echo -e "${YELLOW}[MEDIUM]${NC} $1"
    MEDIUM=$((MEDIUM + 1))
}

low() {
    echo -e "${BLUE}[LOW]${NC} $1"
    LOW=$((LOW + 1))
}

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    INFO=$((INFO + 1))
}

# Setup security audit environment
setup_audit_environment() {
    log "Setting up security audit environment..."
    
    # Create directories
    mkdir -p "$REPORTS_DIR"
    mkdir -p "$TEMP_DIR"
    
    # Set trap for cleanup
    trap cleanup_temp_files EXIT
    
    success "Security audit environment ready"
}

# Cleanup temporary files
cleanup_temp_files() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Check security tools availability
check_security_tools() {
    log "Checking security tools availability..."
    
    local missing_tools=()
    local optional_tools=()
    
    # Check for core tools
    for tool in curl openssl nc; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    # Check for optional security tools
    for tool in nmap nikto sqlmap gobuster; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            optional_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        warn "Missing required tools: ${missing_tools[*]}"
        warn "Install missing tools for complete security audit"
    fi
    
    if [ ${#optional_tools[@]} -gt 0 ]; then
        info "Optional security tools not found: ${optional_tools[*]}"
        info "Install these for enhanced security scanning capabilities"
    fi
    
    success "Security tools check completed"
}

# Test service accessibility
test_service_accessibility() {
    log "Testing service accessibility..."
    
    local services=(
        "$GOTAK_BASE_URL:HTTP API"
        "${GOTAK_WS_URL}:WebSocket"
    )
    
    for service_info in "${services[@]}"; do
        local service_url="${service_info%:*}"
        local service_name="${service_info#*:}"
        
        if curl -s -f -m 10 "${service_url%ws:*}http:${service_url#*:}/health" >/dev/null 2>&1; then
            success "$service_name is accessible"
        else
            error "$service_name is not accessible at $service_url"
            return 1
        fi
    done
    
    success "Service accessibility test passed"
}

# HTTP Security Headers Check
check_http_security_headers() {
    log "Checking HTTP security headers..."
    
    local report_file="$REPORTS_DIR/http_headers_$(date +%Y%m%d_%H%M%S).txt"
    local temp_headers="$TEMP_DIR/headers.txt"
    
    {
        echo "HTTP Security Headers Audit Report"
        echo "Generated: $(date)"
        echo "Target: $GOTAK_BASE_URL"
        echo "=================================="
        echo ""
        
    } > "$report_file"
    
    # Get headers
    curl -I -s -m 10 "$GOTAK_BASE_URL/health" > "$temp_headers" 2>/dev/null || {
        error "Failed to retrieve HTTP headers"
        return 1
    }
    
    # Check critical security headers
    local security_headers=(
        "Strict-Transport-Security:HSTS header missing - enables HTTPS enforcement"
        "X-Content-Type-Options:X-Content-Type-Options header missing - prevents MIME sniffing"
        "X-Frame-Options:X-Frame-Options header missing - prevents clickjacking"
        "X-XSS-Protection:X-XSS-Protection header missing - XSS protection disabled"
        "Content-Security-Policy:CSP header missing - no content security policy"
        "Referrer-Policy:Referrer-Policy header missing - referrer information leakage"
        "Permissions-Policy:Permissions-Policy header missing - no feature policy"
    )
    
    for header_check in "${security_headers[@]}"; do
        local header_name="${header_check%:*}"
        local description="${header_check#*:}"
        
        if grep -qi "^$header_name:" "$temp_headers"; then
            local header_value
            header_value=$(grep -i "^$header_name:" "$temp_headers" | cut -d':' -f2- | tr -d '\r\n' | sed 's/^ *//')
            success "$header_name: $header_value"
            echo "✓ $header_name: $header_value" >> "$report_file"
        else
            medium "$description"
            echo "✗ $header_name: MISSING" >> "$report_file"
        fi
    done
    
    # Check for information disclosure headers
    local info_headers=(
        "Server"
        "X-Powered-By" 
        "X-AspNet-Version"
        "X-AspNetMvc-Version"
    )
    
    echo "" >> "$report_file"
    echo "Information Disclosure Check:" >> "$report_file"
    echo "----------------------------" >> "$report_file"
    
    for header in "${info_headers[@]}"; do
        if grep -qi "^$header:" "$temp_headers"; then
            local header_value
            header_value=$(grep -i "^$header:" "$temp_headers" | cut -d':' -f2- | tr -d '\r\n' | sed 's/^ *//')
            low "Information disclosure: $header: $header_value"
            echo "! $header: $header_value (Information Disclosure)" >> "$report_file"
        fi
    done
    
    # Raw headers for reference
    echo "" >> "$report_file"
    echo "Raw HTTP Headers:" >> "$report_file"
    echo "-----------------" >> "$report_file"
    cat "$temp_headers" >> "$report_file"
    
    success "HTTP security headers check completed"
    info "Report saved: $report_file"
}

# TLS/SSL Security Check
check_tls_security() {
    log "Checking TLS/SSL security configuration..."
    
    local report_file="$REPORTS_DIR/tls_security_$(date +%Y%m%d_%H%M%S).txt"
    local https_url="${GOTAK_BASE_URL/http:/https:}"
    
    {
        echo "TLS/SSL Security Audit Report"
        echo "Generated: $(date)"
        echo "Target: $https_url"
        echo "============================="
        echo ""
        
    } > "$report_file"
    
    # Check if HTTPS is available
    if curl -s -m 10 "$https_url/health" >/dev/null 2>&1; then
        success "HTTPS endpoint is accessible"
        echo "✓ HTTPS endpoint accessible" >> "$report_file"
        
        # Check TLS version and cipher
        local tls_info
        tls_info=$(openssl s_client -connect "${https_url#https://}/443" -servername "${https_url#https://}" </dev/null 2>/dev/null | openssl x509 -noout -text 2>/dev/null || echo "TLS info unavailable")
        
        echo "" >> "$report_file"
        echo "TLS Certificate Information:" >> "$report_file"
        echo "----------------------------" >> "$report_file"
        echo "$tls_info" >> "$report_file"
        
        # Check for weak TLS versions
        for version in ssl2 ssl3 tls1 tls1_1; do
            if openssl s_client -"$version" -connect "${https_url#https://}:443" </dev/null >/dev/null 2>&1; then
                high "Weak TLS version supported: $version"
                echo "✗ Weak TLS version: $version" >> "$report_file"
            fi
        done
        
        success "TLS security check completed"
        
    else
        warn "HTTPS endpoint not accessible - TLS checks skipped"
        echo "! HTTPS not available - TLS security cannot be verified" >> "$report_file"
    fi
    
    info "Report saved: $report_file"
}

# Authentication Security Check
check_authentication_security() {
    log "Checking authentication security..."
    
    local report_file="$REPORTS_DIR/auth_security_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Authentication Security Audit Report"
        echo "Generated: $(date)"
        echo "Target: $GOTAK_BASE_URL"
        echo "===================================="
        echo ""
        
    } > "$report_file"
    
    # Test authentication endpoints
    local auth_endpoint="$GOTAK_BASE_URL/api/auth/login"
    
    # Check if authentication is required
    local unauth_response
    unauth_response=$(curl -s -o /dev/null -w "%{http_code}" "$GOTAK_BASE_URL/api/routes" 2>/dev/null || echo "000")
    
    if [ "$unauth_response" = "401" ] || [ "$unauth_response" = "403" ]; then
        success "Authentication is properly enforced"
        echo "✓ Authentication properly enforced (HTTP $unauth_response)" >> "$report_file"
    else
        critical "Authentication bypass possible - unauthenticated access allowed"
        echo "✗ CRITICAL: Unauthenticated access allowed (HTTP $unauth_response)" >> "$report_file"
    fi
    
    # Test for common authentication vulnerabilities
    echo "" >> "$report_file"
    echo "Authentication Vulnerability Tests:" >> "$report_file"
    echo "-----------------------------------" >> "$report_file"
    
    # SQL injection in login
    local sqli_payloads=(
        "admin' OR '1'='1"
        "admin'; DROP TABLE users; --"
        "admin' UNION SELECT 1,2,3 --"
    )
    
    for payload in "${sqli_payloads[@]}"; do
        local response
        response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$auth_endpoint" \
            -H "Content-Type: application/json" \
            -d "{\"username\":\"$payload\",\"password\":\"password\"}" 2>/dev/null || echo "000")
        
        if [ "$response" = "200" ]; then
            critical "SQL injection vulnerability detected in authentication"
            echo "✗ CRITICAL: SQL injection vulnerability (payload: $payload)" >> "$report_file"
        fi
    done
    
    # Test rate limiting
    echo "" >> "$report_file"
    echo "Rate Limiting Test:" >> "$report_file"
    echo "------------------" >> "$report_file"
    
    local rate_limit_count=0
    for i in {1..10}; do
        local response
        response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$auth_endpoint" \
            -H "Content-Type: application/json" \
            -d '{"username":"nonexistent","password":"invalid"}' 2>/dev/null || echo "000")
        
        if [ "$response" = "429" ]; then
            rate_limit_count=$((rate_limit_count + 1))
        fi
        sleep 0.1
    done
    
    if [ $rate_limit_count -gt 0 ]; then
        success "Rate limiting is active"
        echo "✓ Rate limiting active (triggered $rate_limit_count times)" >> "$report_file"
    else
        medium "Rate limiting not detected - brute force attacks possible"
        echo "! Rate limiting not detected" >> "$report_file"
    fi
    
    success "Authentication security check completed"
    info "Report saved: $report_file"
}

# Input Validation Security Check
check_input_validation() {
    log "Checking input validation security..."
    
    local report_file="$REPORTS_DIR/input_validation_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Input Validation Security Audit Report"
        echo "Generated: $(date)"
        echo "Target: $GOTAK_BASE_URL"
        echo "======================================"
        echo ""
        
    } > "$report_file"
    
    # Get a valid token for testing (using test credentials)
    local token
    token=$(curl -s -X POST "$GOTAK_BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"test_user1","password":"test123"}' | \
        grep -o '"token":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "")
    
    if [ -n "$token" ]; then
        local headers="Authorization: Bearer $token"
        
        # Test XSS in various endpoints
        local xss_payloads=(
            "<script>alert('XSS')</script>"
            "javascript:alert('XSS')"
            "<img src=x onerror=alert('XSS')>"
            "'><script>alert('XSS')</script>"
        )
        
        echo "XSS Vulnerability Tests:" >> "$report_file"
        echo "------------------------" >> "$report_file"
        
        for payload in "${xss_payloads[@]}"; do
            # Test route creation with XSS payload
            local response
            response=$(curl -s -X POST "$GOTAK_BASE_URL/api/routes" \
                -H "Content-Type: application/json" \
                -H "$headers" \
                -d "{\"name\":\"$payload\",\"description\":\"test\"}" 2>/dev/null || echo "")
            
            if echo "$response" | grep -q "$payload"; then
                high "XSS vulnerability detected in route name field"
                echo "✗ XSS vulnerability in route name: $payload" >> "$report_file"
            fi
        done
        
        # Test for command injection
        echo "" >> "$report_file"
        echo "Command Injection Tests:" >> "$report_file"
        echo "------------------------" >> "$report_file"
        
        local cmd_payloads=(
            "; ls -la"
            "$(whoami)"
            "\`id\`"
            "| cat /etc/passwd"
        )
        
        for payload in "${cmd_payloads[@]}"; do
            local response
            response=$(curl -s -X POST "$GOTAK_BASE_URL/api/routes" \
                -H "Content-Type: application/json" \
                -H "$headers" \
                -d "{\"name\":\"test$payload\",\"description\":\"test\"}" 2>/dev/null || echo "")
            
            if echo "$response" | grep -E "(root:|bin/bash|uid=)" >/dev/null; then
                critical "Command injection vulnerability detected"
                echo "✗ CRITICAL: Command injection vulnerability: $payload" >> "$report_file"
            fi
        done
        
        success "Input validation check completed"
    else
        warn "Could not obtain authentication token - input validation tests skipped"
        echo "! Authentication failed - input validation tests skipped" >> "$report_file"
    fi
    
    info "Report saved: $report_file"
}

# Network Security Scan
check_network_security() {
    log "Checking network security configuration..."
    
    local report_file="$REPORTS_DIR/network_security_$(date +%Y%m%d_%H%M%S).txt"
    local host="${GOTAK_BASE_URL#http://}"
    local port="${host#*:}"
    host="${host%:*}"
    
    if [ "$port" = "$host" ]; then
        port="80"  # Default HTTP port
    fi
    
    {
        echo "Network Security Audit Report"
        echo "Generated: $(date)"
        echo "Target: $host:$port"
        echo "============================="
        echo ""
        
    } > "$report_file"
    
    # Port scanning with nc
    log "Scanning common ports..."
    local common_ports=(22 23 25 53 80 110 143 443 993 995 1433 3306 5432 6379 8080 8443 9090)
    
    echo "Port Scan Results:" >> "$report_file"
    echo "------------------" >> "$report_file"
    
    for port_num in "${common_ports[@]}"; do
        if nc -z -w3 "$host" "$port_num" 2>/dev/null; then
            if [ "$port_num" = "22" ]; then
                medium "SSH port 22 is open - ensure strong authentication"
            elif [ "$port_num" = "23" ]; then
                high "Telnet port 23 is open - unencrypted protocol"
            elif [ "$port_num" = "3306" ]; then
                medium "MySQL port 3306 is open - ensure proper access controls"
            elif [ "$port_num" = "5432" ]; then
                info "PostgreSQL port 5432 is open"
            elif [ "$port_num" = "6379" ]; then
                medium "Redis port 6379 is open - ensure authentication enabled"
            fi
            echo "✓ Port $port_num: OPEN" >> "$report_file"
        else
            echo "- Port $port_num: CLOSED" >> "$report_file"
        fi
    done
    
    # Check for unnecessary services
    echo "" >> "$report_file"
    echo "Service Analysis:" >> "$report_file"
    echo "----------------" >> "$report_file"
    
    # Use nmap if available for more detailed scanning
    if command -v nmap >/dev/null 2>&1; then
        log "Running nmap service detection..."
        nmap -sV -T4 "$host" 2>/dev/null >> "$report_file" || true
    fi
    
    success "Network security check completed"
    info "Report saved: $report_file"
}

# Configuration Security Audit
check_configuration_security() {
    log "Checking configuration security..."
    
    local report_file="$REPORTS_DIR/config_security_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Configuration Security Audit Report"
        echo "Generated: $(date)"
        echo "===================================="
        echo ""
        
    } > "$report_file"
    
    # Check for sensitive files in the project
    echo "Sensitive Files Check:" >> "$report_file"
    echo "---------------------" >> "$report_file"
    
    local sensitive_patterns=(
        "*.key"
        "*.pem"
        "*.p12"
        "*.pfx"
        "*password*"
        "*secret*"
        "*.env"
        ".env.*"
        "id_rsa"
        "id_dsa"
    )
    
    for pattern in "${sensitive_patterns[@]}"; do
        while IFS= read -r -d '' file; do
            if [ -f "$file" ]; then
                local perms
                perms=$(stat -c "%a" "$file" 2>/dev/null || stat -f "%A" "$file" 2>/dev/null || echo "unknown")
                if [[ "$perms" =~ [2367] ]]; then
                    high "Sensitive file has world-writable permissions: $file ($perms)"
                    echo "✗ World-writable sensitive file: $file ($perms)" >> "$report_file"
                else
                    info "Sensitive file found with safe permissions: $file ($perms)"
                    echo "! Sensitive file: $file ($perms)" >> "$report_file"
                fi
            fi
        done < <(find "$PROJECT_ROOT" -name "$pattern" -type f -print0 2>/dev/null)
    done
    
    # Check Docker configuration security
    echo "" >> "$report_file"
    echo "Docker Security Check:" >> "$report_file"
    echo "---------------------" >> "$report_file"
    
    if [ -f "$PROJECT_ROOT/Dockerfile" ]; then
        # Check for running as root
        if grep -q "USER root" "$PROJECT_ROOT/Dockerfile"; then
            high "Dockerfile runs as root user"
            echo "✗ Dockerfile runs as root" >> "$report_file"
        elif grep -q "USER " "$PROJECT_ROOT/Dockerfile"; then
            success "Dockerfile uses non-root user"
            echo "✓ Non-root user in Dockerfile" >> "$report_file"
        else
            medium "Dockerfile user not explicitly set"
            echo "! Dockerfile user not explicitly set" >> "$report_file"
        fi
        
        # Check for latest tag usage
        if grep -q "FROM.*:latest" "$PROJECT_ROOT/Dockerfile"; then
            medium "Dockerfile uses 'latest' tag - version pinning recommended"
            echo "! Uses 'latest' tag in base image" >> "$report_file"
        fi
        
        # Check for secrets in Dockerfile
        if grep -iE "(password|secret|key|token)" "$PROJECT_ROOT/Dockerfile" | grep -v "ARG\|ENV"; then
            critical "Potential secrets found in Dockerfile"
            echo "✗ CRITICAL: Potential secrets in Dockerfile" >> "$report_file"
        fi
    fi
    
    success "Configuration security check completed"
    info "Report saved: $report_file"
}

# Dependency Security Check
check_dependency_security() {
    log "Checking dependency security..."
    
    local report_file="$REPORTS_DIR/dependency_security_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Dependency Security Audit Report"
        echo "Generated: $(date)"
        echo "================================"
        echo ""
        
    } > "$report_file"
    
    # Go dependencies check
    if [ -f "$PROJECT_ROOT/go.mod" ]; then
        echo "Go Dependencies:" >> "$report_file"
        echo "----------------" >> "$report_file"
        
        # List all dependencies
        cd "$PROJECT_ROOT"
        go list -m all >> "$report_file" 2>/dev/null || echo "Failed to list Go dependencies" >> "$report_file"
        
        # Check for known vulnerabilities with govulncheck if available
        if command -v govulncheck >/dev/null 2>&1; then
            echo "" >> "$report_file"
            echo "Vulnerability Scan (govulncheck):" >> "$report_file"
            echo "---------------------------------" >> "$report_file"
            govulncheck ./... >> "$report_file" 2>&1 || true
        else
            info "govulncheck not available - install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
            echo "! govulncheck not available for vulnerability scanning" >> "$report_file"
        fi
    fi
    
    # Check for outdated dependencies
    echo "" >> "$report_file"
    echo "Dependency Update Check:" >> "$report_file"
    echo "-----------------------" >> "$report_file"
    
    if [ -f "$PROJECT_ROOT/go.mod" ]; then
        cd "$PROJECT_ROOT"
        go list -u -m all 2>/dev/null | grep -E "\[.*\]" >> "$report_file" || echo "All dependencies appear up to date" >> "$report_file"
    fi
    
    success "Dependency security check completed"
    info "Report saved: $report_file"
}

# Generate comprehensive security report
generate_security_report() {
    log "Generating comprehensive security report..."
    
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local main_report="$REPORTS_DIR/security_audit_${timestamp}.html"
    
    cat > "$main_report" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoTAK Security Audit Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #d32f2f; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; border-bottom: 1px solid #ddd; padding-bottom: 5px; }
        .severity-summary { display: flex; justify-content: space-around; margin: 20px 0; }
        .severity-card { text-align: center; padding: 15px; border-radius: 8px; min-width: 100px; }
        .critical { background: #ffebee; border: 2px solid #d32f2f; color: #d32f2f; }
        .high { background: #f3e5f5; border: 2px solid #7b1fa2; color: #7b1fa2; }
        .medium { background: #fff3e0; border: 2px solid #f57c00; color: #f57c00; }
        .low { background: #e3f2fd; border: 2px solid #1976d2; color: #1976d2; }
        .info { background: #e8f5e8; border: 2px solid #388e3c; color: #388e3c; }
        .severity-count { font-size: 2em; font-weight: bold; }
        .severity-label { font-size: 0.9em; margin-top: 5px; }
        .timestamp { color: #666; font-size: 0.9em; margin-bottom: 20px; }
        .section { margin: 20px 0; padding: 15px; border-radius: 4px; background: #fafafa; border-left: 4px solid #2196f3; }
        ul { margin: 10px 0; }
        li { margin: 5px 0; }
        .report-link { color: #1976d2; text-decoration: none; }
        .report-link:hover { text-decoration: underline; }
        .status-ok { color: #4caf50; font-weight: bold; }
        .status-warn { color: #ff9800; font-weight: bold; }
        .status-error { color: #f44336; font-weight: bold; }
        .recommendations { background: #e8f5e8; border: 1px solid #4caf50; border-radius: 4px; padding: 15px; margin: 20px 0; }
        .recommendations h3 { color: #2e7d32; margin-top: 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔒 GoTAK Security Audit Report</h1>
        <div class="timestamp">Generated: $(date)</div>
        <div class="timestamp">Target: $GOTAK_BASE_URL</div>
        
        <h2>Security Issue Summary</h2>
        <div class="severity-summary">
            <div class="severity-card critical">
                <div class="severity-count">$CRITICAL</div>
                <div class="severity-label">CRITICAL</div>
            </div>
            <div class="severity-card high">
                <div class="severity-count">$HIGH</div>
                <div class="severity-label">HIGH</div>
            </div>
            <div class="severity-card medium">
                <div class="severity-count">$MEDIUM</div>
                <div class="severity-label">MEDIUM</div>
            </div>
            <div class="severity-card low">
                <div class="severity-count">$LOW</div>
                <div class="severity-label">LOW</div>
            </div>
            <div class="severity-card info">
                <div class="severity-count">$INFO</div>
                <div class="severity-label">INFO</div>
            </div>
        </div>
        
        <h2>Audit Sections</h2>
        <div class="section">
            <h3>HTTP Security Headers</h3>
            <p>Analysis of security-related HTTP headers to prevent common web attacks.</p>
            <ul>$(find "$REPORTS_DIR" -name "http_headers_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>TLS/SSL Security</h3>
            <p>Evaluation of encryption protocols and certificate configuration.</p>
            <ul>$(find "$REPORTS_DIR" -name "tls_security_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>Authentication Security</h3>
            <p>Assessment of authentication mechanisms and access controls.</p>
            <ul>$(find "$REPORTS_DIR" -name "auth_security_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>Input Validation</h3>
            <p>Testing for injection vulnerabilities and input sanitization.</p>
            <ul>$(find "$REPORTS_DIR" -name "input_validation_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>Network Security</h3>
            <p>Network configuration and exposed services analysis.</p>
            <ul>$(find "$REPORTS_DIR" -name "network_security_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>Configuration Security</h3>
            <p>Review of system and application configuration security.</p>
            <ul>$(find "$REPORTS_DIR" -name "config_security_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="section">
            <h3>Dependency Security</h3>
            <p>Analysis of third-party dependencies for known vulnerabilities.</p>
            <ul>$(find "$REPORTS_DIR" -name "dependency_security_*.txt" -newer "$REPORTS_DIR" 2>/dev/null | head -1 | xargs -I {} echo "<li><a href=\"$(basename {})\" class=\"report-link\">{}</a></li>" || echo "<li>No reports generated</li>")</ul>
        </div>
        
        <div class="recommendations">
            <h3>🎯 Security Recommendations</h3>
            <ul>
EOF

    # Add recommendations based on findings
    if [ $CRITICAL -gt 0 ]; then
        echo "<li><strong>URGENT:</strong> Address all critical security issues immediately</li>" >> "$main_report"
    fi
    
    if [ $HIGH -gt 0 ]; then
        echo "<li><strong>HIGH PRIORITY:</strong> Resolve high-severity issues within 24 hours</li>" >> "$main_report"
    fi
    
    if [ $MEDIUM -gt 0 ]; then
        echo "<li>Address medium-severity issues in the next maintenance window</li>" >> "$main_report"
    fi
    
    cat >> "$main_report" << EOF
                <li>Implement security headers (HSTS, CSP, X-Frame-Options)</li>
                <li>Enable TLS 1.3 and disable older protocol versions</li>
                <li>Implement rate limiting for all API endpoints</li>
                <li>Regular security dependency updates</li>
                <li>Enable comprehensive audit logging</li>
                <li>Implement Web Application Firewall (WAF)</li>
                <li>Regular penetration testing</li>
            </ul>
        </div>
        
        <div class="section">
            <h3>📈 Next Steps</h3>
            <p>1. Review and address findings by severity level</p>
            <p>2. Implement security hardening recommendations</p>
            <p>3. Schedule regular security audits</p>
            <p>4. Update incident response procedures</p>
        </div>
    </div>
</body>
</html>
EOF

    success "Comprehensive security report generated: $main_report"
    echo "Security Audit Summary:"
    echo "======================"
    echo "Critical Issues: $CRITICAL"
    echo "High Issues: $HIGH"
    echo "Medium Issues: $MEDIUM"
    echo "Low Issues: $LOW"
    echo "Informational: $INFO"
    echo ""
    echo "Main Report: $main_report"
}

# Main security audit function
main() {
    local audit_type="${1:-full}"
    shift || true
    
    case "$audit_type" in
        "headers")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_http_security_headers
            ;;
        "tls")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_tls_security
            ;;
        "auth")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_authentication_security
            ;;
        "input")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_input_validation
            ;;
        "network")
            setup_audit_environment
            check_security_tools
            check_network_security
            ;;
        "config")
            setup_audit_environment
            check_configuration_security
            ;;
        "dependencies")
            setup_audit_environment
            check_dependency_security
            ;;
        "quick")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_http_security_headers
            check_authentication_security
            generate_security_report
            ;;
        "full")
            setup_audit_environment
            check_security_tools
            test_service_accessibility
            check_http_security_headers
            check_tls_security
            check_authentication_security
            check_input_validation
            check_network_security
            check_configuration_security
            check_dependency_security
            generate_security_report
            ;;
        "help"|*)
            cat << EOF
GoTAK Security Audit Tool

Usage: $0 <audit_type>

Audit Types:
  full            Complete security audit (default)
  quick           Quick security assessment
  headers         HTTP security headers check
  tls             TLS/SSL security assessment
  auth            Authentication security check
  input           Input validation testing
  network         Network security scan
  config          Configuration security review
  dependencies    Dependency vulnerability check
  help            Show this help message

Environment Variables:
  GOTAK_BASE_URL   Base URL for security tests (default: http://localhost:8080)
  GOTAK_WS_URL     WebSocket URL (default: ws://localhost:8080)

Examples:
  $0 full          # Complete security audit
  $0 quick         # Quick assessment
  $0 headers       # Only check HTTP headers
  $0 auth          # Only test authentication

Reports are saved to: $REPORTS_DIR/

Security Tools Recommendations:
  - nmap: Network scanning
  - nikto: Web vulnerability scanning  
  - sqlmap: SQL injection testing
  - govulncheck: Go vulnerability scanning

EOF
            ;;
    esac
}

# Run main function
main "$@"
