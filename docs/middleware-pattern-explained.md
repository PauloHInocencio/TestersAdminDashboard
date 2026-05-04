# Go HTTP Middleware Pattern Explained

## What is Middleware?

Middleware is code that runs **before** your main handler processes a request. It's like a checkpoint or filter that requests pass through.

Common uses:
- Authentication (check if user is logged in)
- Logging (record request details)
- CORS handling
- Request validation
- Adding data to the request context

## The Pattern Breakdown

Let's use our `RequireAdmin` middleware from `middleware/admin_auth.go` as an example:

```go
func RequireAdmin(store admin.AdminStore) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Middleware logic here
            cookie, err := r.Cookie("admin_session")
            if err != nil || cookie.Value == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            tokenHash := utils.HashToken(cookie.Value)
            email, err := store.FindValidSession(r.Context(), tokenHash)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), AdminEmailKey, email)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Three Layers of Functions

This pattern has **three nested functions**. Let's understand each layer:

#### Layer 1: The Outer Function
```go
func RequireAdmin(store admin.AdminStore) func(http.Handler) http.Handler
```

**Purpose**: Accept dependencies and configuration

- Takes parameters your middleware needs (like `AdminStore` for database access)
- Returns the middleware function
- This is called once during server setup, not on every request

**Why?** It allows dependency injection - your middleware can access databases, configs, etc.

#### Layer 2: The Middleware Function
```go
func(next http.Handler) http.Handler
```

**Purpose**: Wrap the next handler in the chain

- Takes `next` - the handler that should run after this middleware
- Returns a new handler that wraps it
- Creates the "chain" of handlers

**Why?** This is the standard Go middleware signature. It allows middlewares to be composed together.

#### Layer 3: The Anonymous Handler
```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Your middleware logic
    next.ServeHTTP(w, r)
})
```

**Purpose**: Do the actual middleware work

- Runs on **every request**
- Contains your middleware logic
- Calls `next.ServeHTTP()` to continue the chain

## Understanding `next.ServeHTTP`

This is the key line that makes middleware work:

```go
next.ServeHTTP(w, r.WithContext(ctx))
```

**What it means**: "I'm done with my middleware logic, now let the next handler process this request"

### The Flow

```
Request comes in
    ↓
Your middleware logic BEFORE next.ServeHTTP
    ↓
next.ServeHTTP(w, r)  ← Pass control to next handler
    ↓
Next handler (or middleware) runs
    ↓
Your middleware logic AFTER next.ServeHTTP (if any)
    ↓
Response sent
```

### Important Details

1. **You control if the chain continues**
   ```go
   if err != nil {
       http.Error(w, "Unauthorized", http.StatusUnauthorized)
       return  // Don't call next.ServeHTTP - stop here!
   }
   next.ServeHTTP(w, r)  // Only reached if no error
   ```

2. **You can modify the request**
   ```go
   ctx := context.WithValue(r.Context(), AdminEmailKey, email)
   next.ServeHTTP(w, r.WithContext(ctx))  // Pass modified request
   ```

3. **You can do cleanup after**
   ```go
   next.ServeHTTP(w, r)
   // Code here runs AFTER the next handler completes
   log.Println("Request completed")
   ```

## Complete Example Walkthrough

Let's trace a request through `RequireAdmin`:

### Step 1: Cookie Extraction
```go
cookie, err := r.Cookie("admin_session")
if err != nil || cookie.Value == "" {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return  // Stop here - don't continue the chain
}
```
- Try to read the session cookie
- If missing or empty → respond with 401 and **stop**
- The chain never continues, `next.ServeHTTP` is never called

### Step 2: Token Validation
```go
tokenHash := utils.HashToken(cookie.Value)
email, err := store.FindValidSession(r.Context(), tokenHash)
if err != nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return  // Stop here - invalid session
}
```
- Hash the token (security best practice)
- Check if the session is valid in the database
- If invalid → respond with 401 and **stop**

### Step 3: Context Enrichment
```go
ctx := context.WithValue(r.Context(), AdminEmailKey, email)
next.ServeHTTP(w, r.WithContext(ctx))
```
- Add the admin's email to the request context
- Pass the enriched request to the next handler
- Now any handler down the chain can access the admin email

## How to Use Middleware

### Basic Usage
```go
// Create the middleware (pass dependencies)
adminAuth := RequireAdmin(adminStore)

// Wrap your handler
protectedHandler := adminAuth(yourHandler)

// Or in one line
protectedHandler := RequireAdmin(adminStore)(yourHandler)
```

### With a Router
```go
router.Handle("/admin/dashboard", RequireAdmin(adminStore)(dashboardHandler))
```

### Chaining Multiple Middlewares
```go
handler := loggingMiddleware(
    RequireAdmin(adminStore)(
        dashboardHandler,
    ),
)
```

The request flows through: logging → admin auth → dashboard handler

## Accessing Context Values

In your handler, retrieve the admin email:

```go
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    email := r.Context().Value(middleware.AdminEmailKey).(string)
    fmt.Fprintf(w, "Welcome, admin: %s", email)
}
```

## Key Takeaways

| Concept | Simple Explanation |
|---------|-------------------|
| **Outer function** | Takes dependencies, returns middleware (called once at setup) |
| **Middleware function** | Takes next handler, returns wrapped handler (creates the chain) |
| **Anonymous handler** | The actual logic that runs on every request |
| **`next.ServeHTTP`** | "I'm done, let the next handler take over" |
| **Stopping the chain** | Just `return` without calling `next.ServeHTTP` |
| **Context** | Use `r.Context()` to pass data between middlewares and handlers |

## The Power of This Pattern

1. **Reusable**: Write once, use on any handler
2. **Composable**: Chain multiple middlewares together
3. **Clean**: Keep your handlers focused on business logic
4. **Testable**: Test middleware independently
5. **Dependency injection**: Pass what each middleware needs

This pattern is the foundation of building modular, maintainable HTTP servers in Go.
