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
func RequireAdmin(store admin.AdminStore) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
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
            next(w, r.WithContext(ctx))
        }
    }
}
```

### Three Layers of Functions

This pattern has **three nested functions**. Let's understand each layer:

#### Layer 1: The Outer Function
```go
func RequireAdmin(store admin.AdminStore) func(http.HandlerFunc) http.HandlerFunc
```

**Purpose**: Accept dependencies and configuration

- Takes parameters your middleware needs (like `AdminStore` for database access)
- Returns the middleware function
- This is called once during server setup, not on every request

**Why?** It allows dependency injection - your middleware can access databases, configs, etc.

#### Layer 2: The Middleware Function
```go
func(next http.HandlerFunc) http.HandlerFunc
```

**Purpose**: Wrap the next handler in the chain

- Takes `next` - the handler function that should run after this middleware
- Returns a new handler function that wraps it
- Creates the "chain" of handlers

**Why?** This is a cleaner middleware signature when working with function handlers. It works directly with `http.HandlerFunc`, which is what `router.HandleFunc` expects.

**Note**: The difference between `http.Handler` and `http.HandlerFunc`:
- `http.Handler` is an interface with a `ServeHTTP` method
- `http.HandlerFunc` is a function type `func(ResponseWriter, *Request)` that implements `http.Handler`
- Using `http.HandlerFunc` in the middleware signature is simpler when you know you're working with function handlers

#### Layer 3: The Anonymous Handler
```go
func(w http.ResponseWriter, r *http.Request) {
    // Your middleware logic
    next(w, r)  // Call the function directly
}
```

**Purpose**: Do the actual middleware work

- Runs on **every request**
- Contains your middleware logic
- Calls `next(w, r)` directly to continue the chain (since `next` is a function, not an interface)
- No need to wrap in `http.HandlerFunc()` since we're already returning a function

## Understanding `next(w, r)` - Calling the Next Handler

This is the key line that makes middleware work:

```go
next(w, r.WithContext(ctx))
```

**What it means**: "I'm done with my middleware logic, now let the next handler process this request"

Since `next` is of type `http.HandlerFunc` (a function type), we call it directly as a function. This is cleaner than using `next.ServeHTTP(w, r)`, which would also work but adds unnecessary indirection.

### The Flow

```
Request comes in
    ↓
Your middleware logic BEFORE next()
    ↓
next(w, r)  ← Pass control to next handler
    ↓
Next handler (or middleware) runs
    ↓
Your middleware logic AFTER next() (if any)
    ↓
Response sent
```

### Important Details

1. **You control if the chain continues**
   ```go
   if err != nil {
       http.Error(w, "Unauthorized", http.StatusUnauthorized)
       return  // Don't call next() - stop here!
   }
   next(w, r)  // Only reached if no error
   ```

2. **You can modify the request**
   ```go
   ctx := context.WithValue(r.Context(), AdminEmailKey, email)
   next(w, r.WithContext(ctx))  // Pass modified request
   ```

3. **You can do cleanup after**
   ```go
   next(w, r)
   // Code here runs AFTER the next handler completes
   log.Println("Request completed")
   ```

### Why `next(w, r)` Instead of `next.ServeHTTP(w, r)`?

Both work, but `next(w, r)` is better:
- **Direct function call**: Since `next` is an `http.HandlerFunc`, calling it directly is more natural
- **Less indirection**: `next.ServeHTTP(w, r)` internally just calls `next(w, r)` anyway
- **Clearer code**: Makes it obvious you're calling a function, not invoking an interface method

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

### Real-World Example from Our Codebase

Here's how we use `RequireAdmin` in `services/admin/route.go`:

```go
func (h *Handler) RegisterRoutes(router *http.ServeMux) {
    // Step 1: Create the middleware once, passing the dependency
    requireAdmin := middleware.RequireAdmin(h.adminStore)

    // Step 2: Public routes (no middleware)
    router.HandleFunc("POST /admin/request-magic-link", h.requestMagicLink)
    router.HandleFunc("GET /admin/callback", h.callback)

    // Step 3: Protected routes (wrapped with middleware)
    router.HandleFunc("GET /admin/testers", requireAdmin(h.getTesters))
    router.HandleFunc("POST /admin/testers/{id}/approve", requireAdmin(h.approveTester))
    router.HandleFunc("POST /admin/testers/{id}/reject", requireAdmin(h.rejectTester))
}
```

**Key points**:
1. `requireAdmin := middleware.RequireAdmin(h.adminStore)` creates the middleware once
2. `requireAdmin(h.getTesters)` wraps the handler - this is the function returned by Layer 2
3. The wrapped handler is passed to `router.HandleFunc`
4. Public endpoints don't use the middleware
5. All protected endpoints simply wrap their handler: `requireAdmin(handler)`

### Pattern Breakdown

```go
// Step 1: Initialize with dependencies (called once)
requireAdmin := middleware.RequireAdmin(h.adminStore)

// Step 2: Wrap handlers that need protection
router.HandleFunc("GET /admin/testers", requireAdmin(h.getTesters))
//                                      └─────────┬────────────┘
//                                        Wraps the handler
```

### Inline Usage
You can also use it inline without storing in a variable:

```go
router.HandleFunc("GET /admin/testers",
    middleware.RequireAdmin(adminStore)(h.getTesters))
```

### Chaining Multiple Middlewares
```go
// If you had multiple middlewares
router.HandleFunc("GET /admin/testers",
    loggingMiddleware(
        requireAdmin(
            h.getTesters,
        ),
    ),
)
```

The request flows through: logging → admin auth → handler

## Accessing Context Values

In your handler, retrieve the admin email:

```go
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    email := r.Context().Value(middleware.AdminEmailKey).(string)
    fmt.Fprintf(w, "Welcome, admin: %s", email)
}
```

## Why `http.HandlerFunc` Instead of `http.Handler`?

Our implementation uses `http.HandlerFunc` in the signature instead of `http.Handler`:

```go
// Our pattern (simpler)
func(http.HandlerFunc) http.HandlerFunc

// Alternative pattern (more generic)
func(http.Handler) http.Handler
```

**Advantages of our approach**:

1. **Simpler code**: No need to wrap the return value in `http.HandlerFunc()`
2. **More explicit**: Makes it clear we're working with function handlers
3. **Perfect for ServeMux**: `router.HandleFunc` expects a function, not an interface
4. **Still compatible**: `http.HandlerFunc` implements `http.Handler`, so it works everywhere

**When to use which**:
- Use `http.HandlerFunc` when you know you're working with function handlers (most cases)
- Use `http.Handler` when you need to work with any type that implements the `http.Handler` interface

In practice, the `http.HandlerFunc` pattern is simpler and covers 99% of use cases.

## Key Takeaways

| Concept | Simple Explanation |
|---------|-------------------|
| **Outer function** | Takes dependencies, returns middleware (called once at setup) |
| **Middleware function** | Takes next handler, returns wrapped handler (creates the chain) |
| **Anonymous handler** | The actual logic that runs on every request |
| **`next(w, r)`** | "I'm done, let the next handler take over" - direct function call |
| **Stopping the chain** | Just `return` without calling `next(w, r)` |
| **Context** | Use `r.Context()` to pass data between middlewares and handlers |
| **`http.HandlerFunc` vs `http.Handler`** | `HandlerFunc` is simpler and allows direct function calls |

## The Power of This Pattern

1. **Reusable**: Write once, use on any handler
2. **Composable**: Chain multiple middlewares together
3. **Clean**: Keep your handlers focused on business logic
4. **Testable**: Test middleware independently
5. **Dependency injection**: Pass what each middleware needs
6. **Type-safe**: The compiler ensures correct usage

This pattern is the foundation of building modular, maintainable HTTP servers in Go.
