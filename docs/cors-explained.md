# CORS Behavior Explanation

## Understanding Why Curl Succeeds (Educational)

### How CORS Actually Works

**CORS (Cross-Origin Resource Sharing) is a browser security mechanism, not a server security mechanism.** Here's what's happening with your current implementation:

#### Current Behavior (Working Correctly ✓)

1. **Your curl command:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/testers/signup \
     -H "Origin: http://malicious-site.com" \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","name":"Test User","platform":"android"}'
   ```

2. **Server response:**
   - Status: `200 OK` ✓
   - Headers: `Vary: Origin` ✓
   - **Missing header:** `Access-Control-Allow-Origin` ✓ (This is correct!)
   - Body: Success message ✓

3. **What's happening:**
   - The server **processes the request** normally (returns 200)
   - The rs/cors library **does NOT add** `Access-Control-Allow-Origin` header because `http://malicious-site.com` is not in the allowed origins list
   - **curl doesn't care** - it displays the response because curl is not a browser
   - **A browser WOULD block** this response because there's no `Access-Control-Allow-Origin` header

### Why Curl Succeeds but Browsers Don't

| Client | Behavior | Reason |
|--------|----------|--------|
| **curl** | ✅ Shows response | curl is a command-line tool that doesn't enforce CORS |
| **Browser** | ❌ Blocks response | Browser checks for `Access-Control-Allow-Origin` header and blocks if missing |
| **Postman** | ✅ Shows response | Postman disables CORS checks for testing purposes |

### The Key Insight

**CORS is enforced by the browser, not the server.**

The server will:
- Always process valid HTTP requests (return 200, 400, 500, etc.)
- Add CORS headers only for allowed origins
- Not add CORS headers for non-allowed origins

The browser will:
- Make the request to the server
- Receive the response
- Check for `Access-Control-Allow-Origin` header
- **Block the response** if header is missing or doesn't match the origin
- Throw a CORS error in the console

---

## Testing CORS Properly

### Test 1: Check for CORS Headers with Malicious Origin

```bash
curl -X POST http://localhost:8080/api/v1/testers/signup \
  -H "Origin: http://malicious-site.com" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test","platform":"android"}' \
  -v 2>&1 | grep -i "access-control"
```

**Expected:** No output (no `Access-Control-Allow-Origin` header) ✓

### Test 2: Check for CORS Headers with Allowed Origin

```bash
curl -X POST http://localhost:8080/api/v1/testers/signup \
  -H "Origin: http://localhost:8081" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test","platform":"android"}' \
  -v 2>&1 | grep -i "access-control"
```

**Expected:**
```
< Access-Control-Allow-Origin: http://localhost:8081
< Access-Control-Allow-Credentials: true
```

### Test 3: Browser Testing (Real CORS Test)

Create a simple HTML file and test from a browser:

```html
<!-- test-cors.html -->
<!DOCTYPE html>
<html>
<head>
    <title>CORS Test</title>
</head>
<body>
    <h1>CORS Testing</h1>
    <button onclick="testCORS()">Test Tester Signup</button>
    <pre id="result"></pre>

    <script>
        async function testCORS() {
            const resultEl = document.getElementById('result');
            resultEl.textContent = 'Testing...';

            try {
                const response = await fetch('http://localhost:8080/api/v1/testers/signup', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        email: 'test@example.com',
                        name: 'Test User',
                        platform: 'android'
                    })
                });

                const data = await response.json();
                resultEl.textContent = 'SUCCESS:\\n' + JSON.stringify(data, null, 2);
                resultEl.style.color = 'green';
            } catch (error) {
                resultEl.textContent = 'CORS ERROR:\\n' + error.message +
                    '\\n\\nThis is expected if served from non-allowed origin!';
                resultEl.style.color = 'red';
            }
        }
    </script>
</body>
</html>
```

**How to test:**

1. **Allowed origin test (should succeed):**
   - Serve the file from `http://localhost:8081`
   - Click the button
   - Should see success message

2. **Blocked origin test (should fail with CORS error):**
   - Serve the file from `http://localhost:9999` or any other port
   - Click the button
   - Should see CORS error in browser console and result area

---

## Current CORS Configuration

**File:** `api/server.go` (lines 42-48)

```go
c := cors.New(cors.Options{
    AllowedOrigins:   []string{"http://localhost:8081"},
    AllowedMethods:   []string{"GET", "POST", "DELETE"},
    AllowCredentials: true,
})

handler := c.Handler(router)
```

### Configuration Breakdown

| Option | Value | Purpose |
|--------|-------|---------|
| `AllowedOrigins` | `["http://localhost:8081"]` | Only requests from this origin will receive CORS headers |
| `AllowedMethods` | `["GET", "POST", "DELETE"]` | HTTP methods that are allowed for cross-origin requests |
| `AllowCredentials` | `true` | Allows cookies and authentication headers to be sent |

### What's Working ✓

- ✅ Restricts cross-origin access to `http://localhost:8081` only
- ✅ Only allows specific HTTP methods (GET, POST, DELETE)
- ✅ Enables credentials for authenticated requests
- ✅ Properly wraps the router with CORS middleware
- ✅ Uses the well-maintained `github.com/rs/cors` library

### What Could Be Improved

- ⚠️ **Hardcoded origins:** Not configurable via environment variables (challenging for different environments)
- ⚠️ **No preflight caching:** Missing `MaxAge` option (could improve performance)
- ⚠️ **No explicit allowed headers:** Could cause issues with custom headers
- ⚠️ **No debug logging:** Harder to troubleshoot CORS issues in development

---

## Optional Improvements

### 1. Environment-Based Configuration

**Current problem:** Origins are hardcoded, making it difficult to use different origins for dev/staging/production.

**Solution:** Load allowed origins from environment variables.

```go
// Get allowed origins from environment or use default
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
if allowedOrigins == "" {
    allowedOrigins = "http://localhost:8081"
}

// Split comma-separated origins
origins := strings.Split(allowedOrigins, ",")
for i := range origins {
    origins[i] = strings.TrimSpace(origins[i])
}

c := cors.New(cors.Options{
    AllowedOrigins:   origins,
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowCredentials: true,
})
```

**.env.example:**
```bash
# CORS Configuration
# Comma-separated list of allowed origins
# Example: http://localhost:8081,https://myapp.com
ALLOWED_ORIGINS=http://localhost:8081
```

**Required imports:**
```go
import (
    "os"
    "strings"
    // ... existing imports
)
```

### 2. Add Explicit Allowed Headers

**Why:** Prevents browsers from blocking requests with custom headers during preflight.

```go
c := cors.New(cors.Options{
    AllowedOrigins:   origins,
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept"},
    AllowCredentials: true,
})
```

### 3. Add Preflight Caching

**Why:** Reduces the number of preflight OPTIONS requests, improving performance.

```go
c := cors.New(cors.Options{
    AllowedOrigins:   origins,
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept"},
    AllowCredentials: true,
    MaxAge:           300, // Cache preflight response for 5 minutes
})
```

### 4. Add Debug Logging (Development Only)

**Why:** Makes CORS issues easier to troubleshoot during development.

```go
c := cors.New(cors.Options{
    AllowedOrigins:   origins,
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept"},
    AllowCredentials: true,
    MaxAge:           300,
    Debug:            os.Getenv("ENV") == "development", // Enable debug in dev only
})
```

### Complete Enhanced Configuration

```go
// api/server.go

// Get allowed origins from environment or use default
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
if allowedOrigins == "" {
    allowedOrigins = "http://localhost:8081"
}

// Split comma-separated origins and trim whitespace
origins := strings.Split(allowedOrigins, ",")
for i := range origins {
    origins[i] = strings.TrimSpace(origins[i])
}

// Configure CORS with enhanced options
c := cors.New(cors.Options{
    AllowedOrigins:   origins,
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept"},
    AllowCredentials: true,
    MaxAge:           300,
    Debug:            os.Getenv("ENV") == "development",
})

handler := c.Handler(router)
```

---

## Summary

### ✅ Your CORS Is Working Correctly

The current CORS configuration is **functioning as designed**:

1. **Malicious origins are blocked** (by browsers, not servers)
2. The `rs/cors` library correctly **omits** the `Access-Control-Allow-Origin` header for non-allowed origins
3. Browsers will **block the response** when the header is missing
4. curl succeeds because **it doesn't enforce CORS** (it's not a browser)

### 🎯 Key Takeaways

1. **CORS is browser-enforced, not server-enforced**
   - Servers process all valid HTTP requests
   - Servers add CORS headers for allowed origins only
   - Browsers block responses without proper CORS headers

2. **curl is not a CORS test**
   - Use browser testing or check for CORS headers
   - curl will always show the response regardless of origin

3. **Your configuration is secure**
   - Only `http://localhost:8081` receives CORS headers
   - Other origins are blocked by browsers
   - Credentials are enabled for authenticated requests

### 🚀 Optional Next Steps

If you want to enhance your CORS setup:

1. ✅ Make origins environment-configurable (recommended for production)
2. ✅ Add explicit allowed headers (prevents preflight issues)
3. ✅ Enable preflight caching (improves performance)
4. ✅ Add debug logging for development (easier troubleshooting)

---

## References

- **rs/cors Library:** https://github.com/rs/cors
- **MDN CORS Guide:** https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
- **CORS Specification:** https://fetch.spec.whatwg.org/#http-cors-protocol

---

## Related Files

- `api/server.go:42-48` - CORS middleware configuration
- `services/tester/route.go:24` - Tester signup endpoint registration