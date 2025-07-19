# API Reference

## Base URL
```
http://localhost:8080
```

## Endpoints

### Create Short URL
```http
POST /urls
Content-Type: application/json

{
  "long_url": "https://www.example.com",
  "expiration_date": "2025-12-31T23:59:59Z"  // optional
}
```

**Response (200)**
```json
{
  "short_url": "http://localhost:8080/1"
}
```

### Redirect to Long URL
```http
GET /{shortCode}
```
Returns `302 Found` redirect to the original URL.

### Get URL Statistics  
```http
GET /urls/{shortCode}/stats
```

**Response (200)**
```json
{
  "short_code": "1",
  "long_url": "https://www.example.com",
  "created_at": "2025-07-19T17:30:00Z",
  "expiration_date": "2025-12-31T23:59:59Z",
  "id": 1
}
```

### Health Check
```http
GET /health
```

**Response (200)**
```json
{
  "status": "healthy",
  "stats": {
    "total_urls": 1,
    "current_counter": 1,
    "storage_type": "redis"
  }
}
```

## Examples

### cURL
```bash
# Create short URL
curl -X POST http://localhost:8080/urls \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://www.github.com"}'

# Access short URL
curl -L http://localhost:8080/1

# Get statistics
curl http://localhost:8080/urls/1/stats

# Health check
curl http://localhost:8080/health
```

### JavaScript
```javascript
// Create short URL
const response = await fetch('http://localhost:8080/urls', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ long_url: 'https://www.github.com' })
});

const data = await response.json();
console.log(data.short_url); // http://localhost:8080/1
```

## Error Responses

```http
400 Bad Request - Invalid URL format or JSON
404 Not Found - Short code doesn't exist  
500 Internal Server Error - Storage error
```

## Notes

- URLs must start with `http://` or `https://`
- Short codes use Base62 encoding (`0-9A-Za-z`)
- Expired URLs return 404 when accessed
- CORS enabled for browser requests 