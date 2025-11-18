# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it by creating a private security advisory on GitHub or by emailing the maintainers directly. Please do not create public issues for security vulnerabilities.

**Please include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will respond to security reports within 48 hours and provide a fix within 7 days for critical vulnerabilities.

## Security Features

### Core Library Security

#### 1. Cryptographically Secure Random Number Generation
- **Implementation:** Uses `crypto/rand` instead of `math/rand`
- **Impact:** Prevents predictability in user agent selection
- **Location:** `useragent.go:256-267` (secureRandomInt function)

```go
func secureRandomInt(max int) (int, error) {
    if max <= 0 {
        return 0, errors.New("max must be positive")
    }
    nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
    if err != nil {
        return 0, err
    }
    return int(nBig.Int64()), nil
}
```

#### 2. Thread-Safe Operations
- **Implementation:** All Manager operations protected with RWMutex
- **Impact:** Prevents race conditions in concurrent access
- **Location:** `useragent.go:39-45` (Manager struct)

```go
type Manager struct {
    mu            sync.RWMutex
    desktopAgents []UserAgent
    mobileAgents  []UserAgent
    config        Config
}
```

#### 3. Comprehensive Input Validation
- **Implementation:** Validates all user agent data on load
- **Checks:**
  - User agent string not empty
  - String length between 10 and 1000 characters
  - Percentage between 0 and 100
  - Agent type is valid (desktop, mobile, random)
- **Location:** `useragent.go:145-158` (validateAgent function)

#### 4. Immutable Data Returns
- **Implementation:** Returns copies of internal data, not references
- **Impact:** Prevents external modification of internal state
- **Location:** `useragent.go:160-180` (GetAllDesktop, GetAllMobile)

### Demo Application Security

#### 1. SQL Injection Prevention
All database queries use parameterized statements:

```go
// SECURE - Parameterized query
query := `INSERT INTO request_logs (user_agent, agent_type, ...) VALUES (?, ?, ?)`
result, err := db.conn.ExecContext(ctx, query, log.UserAgent, log.AgentType, ...)

// INSECURE - Don't do this
query := fmt.Sprintf("INSERT INTO request_logs VALUES ('%s', '%s')", ua, type)
```

**Location:** `internal/database/database.go:85-103`

#### 2. Input Validation and Sanitization

**Database Inputs:**
- User agent max length: 1000 characters
- IP address max length: 45 characters (IPv6)
- Endpoint max length: 255 characters
- Agent type: enum validation (desktop, mobile, random)

**Location:** `internal/database/database.go:338-368`

**API Inputs:**
- Limit parameter: 1-1000 range validation
- IP address: validated with `net.ParseIP`
- Error messages: HTML escaped and sanitized

**Location:** `internal/api/handlers.go:340-378`

#### 3. Error Message Sanitization
Error messages are sanitized to prevent information leakage:

```go
func sanitizeErrorMessage(message string) string {
    message = html.EscapeString(message)

    sensitivePatterns := []string{
        "/home/", "/usr/", "/var/",
        "password", "token", "secret", "key",
    }

    for _, pattern := range sensitivePatterns {
        if strings.Contains(strings.ToLower(message), pattern) {
            return "an error occurred"
        }
    }

    if len(message) > 200 {
        return "an error occurred"
    }

    return message
}
```

**Location:** `internal/api/handlers.go:355-378`

#### 4. Rate Limiting
Prevents abuse with configurable per-IP rate limiting:

```go
func RateLimitMiddleware(maxRequests int, window time.Duration)
```

**Configuration:** `MAX_REQUESTS_PER_MINUTE` environment variable
**Location:** `internal/api/handlers.go:380-402`

#### 5. Security Headers
All responses include security headers:

```go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("X-XSS-Protection", "1; mode=block")
w.Header().Set("Content-Security-Policy", "default-src 'self'; ...")
```

**Location:** `internal/web/server.go:27-30`, `internal/api/handlers.go:246-248`

#### 6. Client IP Extraction
Safely extracts client IP with validation:

```go
func getClientIP(r *http.Request) string {
    // Validates X-Forwarded-For
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        if len(ips) > 0 {
            ip := strings.TrimSpace(ips[0])
            if net.ParseIP(ip) != nil {  // Validation!
                return ip
            }
        }
    }
    // Falls back to RemoteAddr
    ...
}
```

**Location:** `internal/api/handlers.go:322-343`

### Docker Security

#### 1. Non-Root User Execution
Container runs as unprivileged user `appuser`:

```dockerfile
USER appuser
```

#### 2. Read-Only Root Filesystem
```yaml
read_only: true
```

#### 3. No New Privileges
```yaml
security_opt:
  - no-new-privileges:true
```

#### 4. Resource Limits
```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 256M
```

#### 5. Minimal Base Image
Uses Alpine Linux for minimal attack surface.

## Security Best Practices for Deployment

### 1. Use HTTPS in Production
Deploy behind a reverse proxy (nginx, Caddy, Traefik) with TLS:

```nginx
server {
    listen 443 ssl http2;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 2. Set Strong Configuration
```bash
# Production settings
export APP_ENV=production
export LOG_LEVEL=warn
export MAX_REQUESTS_PER_MINUTE=50

# Secure database
chmod 600 /data/useragent.db
```

### 3. Use Docker Security Scanning
```bash
# Scan image for vulnerabilities
docker scan commonuseragent:latest

# Run Trivy scan
trivy image commonuseragent:latest
```

### 4. Monitor and Audit
- Enable application logging
- Monitor rate limit hits
- Set up alerts for unusual patterns
- Regularly review database logs

### 5. Keep Dependencies Updated
```bash
# Check for updates
go list -m -u all

# Update dependencies
go get -u ./...
go mod tidy
```

## Security Checklist

Before deploying to production:

- [ ] Use HTTPS/TLS
- [ ] Set `APP_ENV=production`
- [ ] Configure appropriate rate limits
- [ ] Set strong file permissions on database
- [ ] Run security scanner: `make security-scan`
- [ ] Run tests with race detector: `make test-race`
- [ ] Review and set resource limits
- [ ] Enable monitoring and logging
- [ ] Use non-root user in Docker
- [ ] Scan Docker image for vulnerabilities
- [ ] Set up backup for database
- [ ] Configure firewall rules
- [ ] Review and minimize exposed ports

## Threat Model

### Threats Mitigated

✅ **SQL Injection** - Parameterized queries
✅ **XSS** - HTML escaping, CSP headers
✅ **Information Leakage** - Error sanitization
✅ **DoS** - Rate limiting, resource limits
✅ **CSRF** - Same-origin policy
✅ **Predictable Random** - Crypto-secure RNG
✅ **Race Conditions** - Thread-safe operations
✅ **Container Escape** - Non-root user, security options

### Residual Risks

⚠️ **DDoS** - Application-level rate limiting is not sufficient for large-scale DDoS
   - **Mitigation:** Use infrastructure-level DDoS protection (CloudFlare, AWS Shield)

⚠️ **Database Backup Security** - Database file may contain sensitive data
   - **Mitigation:** Encrypt backups, secure backup storage

⚠️ **Log Injection** - User-controlled data in logs
   - **Mitigation:** Sanitize before logging, use structured logging

## Security Testing

### Running Security Tests

```bash
# Full security test suite
make security-scan

# Race condition detection
make test-race

# Static analysis
make staticcheck

# All checks
make check
```

### Manual Security Testing

1. **Test SQL Injection:**
```bash
curl "http://localhost:8080/api/logs?limit='; DROP TABLE request_logs; --"
# Should be safely handled
```

2. **Test Rate Limiting:**
```bash
for i in {1..150}; do curl http://localhost:8080/api/random; done
# Should get 429 after limit
```

3. **Test Error Handling:**
```bash
curl http://localhost:8080/api/logs?limit=abc
# Should return sanitized error
```

## Compliance

This application follows:
- OWASP Top 10 security guidelines
- CWE/SANS Top 25 mitigation strategies
- Docker security best practices
- Go secure coding guidelines

## Security Updates

Security updates will be released as patch versions (e.g., v2.0.1) and communicated via:
- GitHub Security Advisories
- Release notes
- README changelog

## Contact

For security concerns, contact the maintainers through GitHub's private security advisory feature.
