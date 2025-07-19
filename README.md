# Tiny URL Service

A URL shortening service built in Go with comprehensive testing. This is a development/learning project.

⚠️ **Important**: This service supports both in-memory and Redis storage. With in-memory storage (default), all URLs and statistics are lost when the service restarts. Use Redis for data persistence.

## 🚀 Features

- **URL shortening** with base62 encoding
- **Dual storage backends** (in-memory or Redis) with atomic operations
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
- Docker (for Redis storage)

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

#### Local Development
```bash
# In-memory storage (quick start)
go run .

# With Redis persistence
docker-compose up -d
STORAGE_TYPE=redis go run .
```

#### Live Demo
The service is deployed on Railway with managed Redis:
- **Demo URL**: https://tiny-url-production.up.railway.app
- **Health Check**: https://tiny-url-production.up.railway.app/health
- **API Testing**: Use the live API for demonstrations

#### Docker Commands
```bash
# Start Redis locally
docker-compose up -d

# Stop Redis
docker-compose down

# Access Redis CLI
docker exec -it tiny-url-redis redis-cli
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | Gin mode (`debug`, `release`, `test`) |
| `BASE_URL` | `http://localhost:8080` | Base URL for short links |
| `STORAGE_TYPE` | `memory` | Storage backend (`memory` or `redis`) |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |
| `READ_TIMEOUT` | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |

## 🐳 Redis Setup

### Docker Compose (Recommended)
The project includes a `docker-compose.yml` file for easy Redis setup:

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: tiny-url-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes --appendfsync everysec
    restart: unless-stopped
```

### Storage Comparison

| Feature | In-Memory | Redis |
|---------|-----------|-------|
| **Data Persistence** | ❌ Lost on restart | ✅ Persists across restarts |
| **Multiple Instances** | ❌ Single instance only | ✅ Supports multiple instances |
| **Performance** | ⚡ Fastest | 🚀 Fast (network overhead) |
| **Memory Usage** | 💾 Process memory | 💾 Redis memory |
| **Setup Complexity** | ✅ Zero setup | 🐳 Requires Docker/Redis |
| **Production Ready** | ❌ Development only | ✅ Production ready |

### Redis Data Structure
URLs are stored as JSON in Redis with the following structure:
```bash
# Keys
counter              # Atomic counter for unique IDs
url:{shortCode}      # URL mapping data

# Example data
GET url:1
> {"id":1,"short_code":"1","long_url":"https://example.com","created_at":"2025-07-19T17:30:00Z"}
```

## 🧪 Testing

### Run All Tests
```bash
# Run complete test suite (39 tests)
go test ./... -v

# With coverage report
go test ./... -cover

# Race condition detection
go test ./... -race
```

### Test Coverage
- **Storage**: 92.4% coverage (Redis + Memory)
- **Middleware**: 100% coverage (Rate limiting)
- **Utils**: 96.8% coverage (Encoding + Validation)
- **Integration**: Full API coverage

### Test Categories
```bash
# Storage tests (with Redis mocking)
go test ./storage -v

# Rate limiter tests
go test ./middleware -v

# Integration tests
go test ./tests -v

# Utility tests
go test ./utils -v
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
│   ├── memory.go             # In-memory implementation
│   └── redis.go              # Redis implementation
├── handlers/
│   ├── url_handlers.go       # HTTP request handlers
│   └── server.go             # Server setup and middleware
├── utils/
│   ├── encoding.go           # Base62 encoding/decoding
│   └── validation.go         # URL validation
├── tests/
│   ├── integration_test.go   # Integration tests
│   └── benchmark_test.go     # Performance benchmarks
├── middleware/
│   └── rate_limiter_test.go  # Rate limiting tests
├── storage/
│   ├── redis_test.go         # Redis storage tests (with mocking)
│   └── memory_test.go        # Memory storage tests
└── docs/
    ├── ARCHITECTURE.md       # System architecture documentation
    └── API.md                # API reference and examples
```

### Key Components

#### Base62 Encoding
- Converts numeric IDs to short, URL-safe strings
- Character set: `0-9A-Za-z` (62 characters)
- Collision-free through atomic counter incrementation

#### Storage Backends

**In-Memory Storage:**
- Mutex-protected concurrent access
- Atomic counter for unique ID generation
- Zero-allocation retrieval operations
- Fast development/testing, data lost on restart

**Redis Storage:**
- Persistent data across restarts
- Atomic counters using Redis INCR
- JSON serialization of URL mappings
- Support for multiple app instances
- Production-ready with data durability

#### Server Features
- Environment-based configuration
- Structured logging with Gin
- Panic recovery middleware
- CORS support
- Content-type validation
- Graceful shutdown with signal handling

## 📊 Performance

Based on benchmark tests:

**In-Memory Storage:**
| Operation | Throughput | Memory per Op |
|-----------|------------|---------------|
| Store URL | ~2M ops/sec | 201 B |
| Retrieve URL | ~13M ops/sec | 0 B |
| Create Short URL (HTTP) | ~50K req/sec | - |
| Redirect (HTTP) | ~100K req/sec | - |

**Redis Storage:**
| Operation | Throughput | Memory per Op |
|-----------|------------|---------------|
| Store URL | ~50K ops/sec | 512 B |
| Retrieve URL | ~100K ops/sec | 256 B |
| Create Short URL (HTTP) | ~10K req/sec | - |
| Redirect (HTTP) | ~20K req/sec | - |

**Race Condition Testing**: ✅ All concurrent access tests pass with `-race` flag for both storage backends

## 📚 Documentation

- **[Architecture Guide](docs/ARCHITECTURE.md)** - System design, components, and data flow
- **[API Reference](docs/API.md)** - Complete API documentation with examples
- **[Project Plan](plan.md)** - Development phases and Railway deployment setup

## Code Quality
- **39 comprehensive tests** with Redis mocking
- **92.4% storage coverage**, 100% middleware coverage
- Integration tests cover all API endpoints
- Race detection prevents concurrency bugs
- Benchmark tests ensure performance

## TODOs

- ✅ **Persistent Storage**: Redis implementation complete
- ✅ **Distributed Counter**: Redis INCR provides atomic counters across instances
- ✅ **Rate Limiting**: Per-IP token bucket (20 req/min) implemented
- **Security**: Add authentication, input sanitization
- **Monitoring**: Add metrics, health checks, observability
- **Error Handling**: More robust error responses and logging
- **Configuration**: More comprehensive config validation
- **Performance**: Connection pooling, caching, optimization
- **Deployment**: Containerization, CI/CD, infrastructure