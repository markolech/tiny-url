# Tiny URL Service

A URL shortening service built in Go with comprehensive testing. This is a development/learning project.

⚠️ **Important**: This service stores all data in memory only. All URLs and statistics are lost when the service restarts.

## 🚀 Features

- **URL shortening** with base62 encoding
- **Thread-safe in-memory storage** with atomic operations
- **HTTP server** with Gin framework
- **Middleware** (CORS, logging, recovery, validation)
- **Environment-based configuration**
- **Graceful shutdown** with timeout handling
- **Test coverage** (unit, integration, benchmark, race detection)
- **URL expiration support**
- **Access statistics tracking**

## 📋 API Endpoints

### Create Short URL
```bash
POST /urls
Content-Type: application/json

{
  "long_url": "https://www.example.com/very/long/path",
  "expiration_date": "2025-12-31T23:59:59Z"  // optional
}
```

**Response:**
```json
{
  "short_url": "http://localhost:8080/1"
}
```

### Redirect to Long URL
```bash
GET /{shortCode}
```
Returns `302 Found` with `Location` header pointing to the original URL.

### Get URL Statistics
```bash
GET /urls/{shortCode}/stats
```

**Response:**
```json
{
  "short_code": "1",
  "long_url": "https://www.example.com/very/long/path",
  "access_count": 42,
  "created_at": "2025-07-19T17:00:00Z",
  "expiration_date": "2025-12-31T23:59:59Z"
}
```

### Health Check
```bash
GET /health
```

## 🛠️ Installation & Usage

### Prerequisites
- Go 1.21 or later

### Setup
```bash
# Clone the repository
git clone <repository-url>
cd tiny-url

# Install dependencies
go mod tidy

# Build the service
go build -o tiny-url-service
```

### Running the Service

```bash
# Default configuration (port 8080)
./tiny-url-service

# Custom configuration
PORT=9000 \
GIN_MODE=release \
BASE_URL=https://yourdomain.com \
./tiny-url-service
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | Gin mode (`debug`, `release`, `test`) |
| `BASE_URL` | `http://localhost:8080` | Base URL for short links |
| `READ_TIMEOUT` | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### Integration Tests
```bash
# Run integration tests
go test ./tests/ -v

# Run specific test categories
go test ./tests/ -run TestCreateShortURL
go test ./tests/ -run TestConcurrentAccess
```

### Benchmark Tests
```bash
# Run performance benchmarks
go test ./tests/ -bench=. -benchmem

# Storage-specific benchmarks
go test ./tests/ -bench=BenchmarkMemoryStorage
```

### Automated Test Scripts
```bash
# Run comprehensive test suite (requires running server)
./scripts/tests/run_all_tests.sh

# Individual test categories
./scripts/tests/basic_tests.sh
./scripts/tests/error_tests.sh
./scripts/tests/concurrent_tests.sh
```

## 🏗️ Architecture

### Project Structure
```
tiny-url/
├── main.go                    # Application entry point
├── config/
│   └── config.go             # Environment configuration
├── models/
│   └── url.go                # Data models
├── storage/
│   ├── interface.go          # Storage interface
│   └── memory.go             # In-memory implementation
├── handlers/
│   ├── url_handlers.go       # HTTP request handlers
│   └── server.go             # Server setup and middleware
├── utils/
│   ├── encoding.go           # Base62 encoding/decoding
│   └── validation.go         # URL validation
├── tests/
│   ├── integration_test.go   # Integration tests
│   └── benchmark_test.go     # Performance benchmarks
└── scripts/
    └── tests/                # Automated test scripts
```

### Key Components

#### Base62 Encoding
- Converts numeric IDs to short, URL-safe strings
- Character set: `0-9A-Za-z` (62 characters)
- Collision-free through atomic counter incrementation

#### Thread-Safe Storage
- Mutex-protected concurrent access
- Atomic counter for unique ID generation
- Zero-allocation retrieval operations

#### Server Features
- Environment-based configuration
- Structured logging with Gin
- Panic recovery middleware
- CORS support
- Content-type validation
- Graceful shutdown with signal handling

## 📊 Performance

Based on benchmark tests:

| Operation | Throughput | Memory per Op |
|-----------|------------|---------------|
| Store URL | ~2M ops/sec | 201 B |
| Retrieve URL | ~13M ops/sec | 0 B |
| Create Short URL (HTTP) | ~50K req/sec | - |
| Redirect (HTTP) | ~100K req/sec | - |

**Race Condition Testing**: ✅ All concurrent access tests pass with `-race` flag

## Code Quality
- All code includes unit tests
- Integration tests cover API endpoints
- Benchmark tests ensure performance
- Race detection prevents concurrency bugs

## TODOs

- **Persistent Storage**: Currently uses in-memory storage - needs Redis/PostgreSQL
- **Distributed Counter**: Single-instance only - needs distributed counter for scaling
- **Security**: Add rate limiting, authentication, input sanitization
- **Monitoring**: Add metrics, health checks, observability
- **Error Handling**: More robust error responses and logging
- **Configuration**: More comprehensive config validation
- **Performance**: Connection pooling, caching, optimization
- **Deployment**: Containerization, CI/CD, infrastructure