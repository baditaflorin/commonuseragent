# Common User Agent Library

A secure, production-ready Go library and demo application for managing user agent strings with comprehensive security features, input validation, and a practical web-based test harness.

## Features

### Core Library
- **Cryptographically Secure Random Selection** - Uses `crypto/rand` instead of `math/rand`
- **Thread-Safe Operations** - All operations are protected with proper synchronization
- **Comprehensive Input Validation** - Validates all user agent data with strict rules
- **Proper Error Handling** - No panics, all errors are returned and handled gracefully
- **Zero External Dependencies** - Core library has no external dependencies

### Demo Application
- **Web-Based GUI** - Interactive test harness for API testing and configuration
- **SQLite Database** - Tracks user agent requests with full history
- **Parameterized Queries** - 100% protection against SQL injection
- **Environment-Based Configuration** - Secure configuration via environment variables
- **Rate Limiting** - Built-in protection against abuse
- **Input Sanitization** - All inputs are validated and sanitized
- **Secure Error Messages** - No internal information leakage
- **Docker Support** - Production-ready containerization

## Security Features

✅ **No SQL Injection** - All database queries use parameterized statements
✅ **Crypto-Secure Random** - Uses `crypto/rand` for unpredictable randomness
✅ **Input Validation** - Strict validation on all inputs with length limits
✅ **Error Sanitization** - No sensitive information in error messages
✅ **Rate Limiting** - Configurable request rate limits
✅ **Security Headers** - CSP, X-Frame-Options, X-Content-Type-Options
✅ **Non-Root Container** - Docker runs as unprivileged user
✅ **Resource Limits** - CPU and memory constraints
✅ **Health Checks** - Built-in health monitoring

## Installation

### Library Only

```bash
go get github.com/baditaflorin/commonuseragent
```

### Full Development Environment

```bash
git clone https://github.com/baditaflorin/commonuseragent.git
cd commonuseragent
make deps
make build
```

## Quick Start

### Using the Library

```go
package main

import (
    "fmt"
    "log"

    "github.com/baditaflorin/commonuseragent"
)

func main() {
    // Get a random desktop user agent
    ua, err := commonuseragent.GetRandomDesktopUA()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Desktop UA:", ua)

    // Get a random mobile user agent
    mobileUA, err := commonuseragent.GetRandomMobileUA()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Mobile UA:", mobileUA)

    // Get any random user agent
    randomUA, err := commonuseragent.GetRandomUA()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Random UA:", randomUA)

    // Get all desktop user agents
    allDesktop, err := commonuseragent.GetAllDesktop()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Total desktop agents: %d\n", len(allDesktop))
}
```

### Running the Demo Application

```bash
# Using Make
make run-demo-dev

# Using Go directly
go run ./cmd/demo

# Using Docker
docker-compose up

# Or with custom configuration
SERVER_PORT=9000 DB_PATH=./custom.db go run ./cmd/demo
```

Access the web interface at: http://localhost:8080

## Configuration

The demo application is configured entirely through environment variables. See `.env.example` for all options.

### Key Configuration Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server port |
| `DB_PATH` | `./useragent.db` | SQLite database path |
| `APP_ENV` | `development` | Environment (development, staging, production) |
| `MAX_REQUESTS_PER_MINUTE` | `100` | Rate limit per IP |

See full configuration documentation in [CONFIGURATION.md](CONFIGURATION.md)

## API Endpoints

### User Agent Endpoints

- `GET /api/desktop` - Get random desktop user agent
- `GET /api/mobile` - Get random mobile user agent
- `GET /api/random` - Get random user agent (any type)
- `GET /api/all/desktop` - Get all desktop user agents
- `GET /api/all/mobile` - Get all mobile user agents

### Monitoring Endpoints

- `GET /api/logs?limit=N` - Get recent requests (max 1000)
- `GET /api/stats` - Get aggregated statistics
- `GET /api/health` - Health check endpoint

### Example Response

```json
{
  "success": true,
  "data": {
    "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...",
    "type": "desktop"
  }
}
```

## Development

### Building

```bash
make build          # Build library
make build-demo     # Build demo application
make all            # Build everything
```

### Testing

```bash
make test           # Run tests
make test-race      # Run tests with race detector
make test-coverage  # Run tests with coverage
make bench          # Run benchmarks
```

### Code Quality

```bash
make fmt            # Format code
make lint           # Run linter
make security-scan  # Run security scanner
make check          # Run all checks
```

## Docker Deployment

```bash
# Using Docker Compose (Recommended)
docker-compose up -d

# Using Docker directly
docker build -t commonuseragent:latest .
docker run -p 8080:8080 commonuseragent:latest
```

## Security

See [SECURITY.md](SECURITY.md) for security best practices, vulnerability reporting, and security architecture details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

UserAgent data comes from https://useragents.me/

This project is licensed under the MIT License - see the LICENSE file for details.

## Changelog

### v2.0.0 (Latest)

**Breaking Changes:**
- All API functions now return errors instead of panicking
- Thread-safe Manager pattern introduced

**Security Improvements:**
- ✅ Replaced `math/rand` with `crypto/rand`
- ✅ Added comprehensive input validation
- ✅ Implemented parameterized SQL queries
- ✅ Added error sanitization
- ✅ Thread-safety with proper locking

**New Features:**
- Web-based test harness and GUI
- SQLite request logging with analytics
- Environment-based configuration
- Rate limiting and security headers
- Docker support with non-root execution
- Comprehensive test suite (78%+ coverage)

See [CHANGELOG.md](CHANGELOG.md) for full version history.
