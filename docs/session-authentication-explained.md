# Session-Based Authentication Security Analysis

## Overview

This project implements a passwordless, session-based authentication system using "magic links" sent via email. The authentication flow combines one-time tokens with persistent session cookies.

## Authentication Flow

1. **Magic Link Request**: Admin requests access via email
2. **Magic Link Callback**: Admin clicks email link with token
3. **Session Creation**: System creates a session and sets a cookie
4. **Protected Routes**: Middleware validates session cookie on each request

---

## The Callback Function (Lines 151-159)

### Code Breakdown

```go
http.SetCookie(w, &http.Cookie{
    Name:     "admin_session",           // Cookie identifier
    Value:    sessionToken,              // Random 32-byte token
    Path:     "/",                       // Available across entire domain
    Expires:  expiresAt,                 // 1 hour from creation
    Secure:   true,                      // HTTPS only
    HttpOnly: true,                      // No JavaScript access
    SameSite: http.SameSiteLaxMode,     // CSRF protection
})
```

### What This Does

**Session Cookie Creation**: After validating the magic link token, the server creates a persistent session by setting a cookie in the user's browser.

**The Session Token**:
- Generated using `utils.GenerateToken()` → 32 cryptographically random bytes (256 bits)
- Only the SHA-256 hash is stored in the database
- The plain token is sent to the browser as the cookie value

---

## Set-Cookie Header Explained

When this code executes, the HTTP response includes a header like:

```
Set-Cookie: admin_session=<token>; Path=/; Expires=<date>; Secure; HttpOnly; SameSite=Lax
```

**What happens:**
1. Browser receives this header in the HTTP response
2. Browser stores the cookie locally
3. On every subsequent request to this domain, the browser automatically sends:
   ```
   Cookie: admin_session=<token>
   ```
4. The server validates this token via middleware (`RequireAdmin`)

---

## Security Analysis

### ✅ Strong Security Measures

#### 1. **Token Strength**
- **32 random bytes** (256 bits) from `crypto/rand`
- Cryptographically secure random number generator
- Probability of guessing: 1 in 2^256 (astronomically impossible)

#### 2. **Secure Flag** (`Secure: true`)
- Cookie only sent over HTTPS
- Prevents interception on insecure HTTP connections
- **Critical**: Ensure production uses HTTPS/TLS

#### 3. **HttpOnly Flag** (`HttpOnly: true`)
- JavaScript cannot access `document.cookie`
- Prevents XSS attacks from stealing sessions
- Most important cookie security flag

#### 4. **SameSite Protection** (`SameSite: http.SameSiteLaxMode`)
- Prevents CSRF attacks
- Cookie not sent on cross-site POST requests
- Sent on safe methods (GET) and same-site requests

#### 5. **Token Hashing**
- Only SHA-256 hash stored in database (`utils.HashToken`)
- If database is compromised, attackers can't use hashed values
- Must steal active cookie to hijack session

#### 6. **Time-Limited Sessions**
- Sessions expire after 1 hour
- Reduces attack window
- Forces periodic re-authentication

#### 7. **Magic Link Expiration**
- Magic links expire after 2 minutes
- Single-use tokens (consumed on first use)
- Limits replay attack window

---

### ⚠️ Security Considerations & Recommendations

#### 1. **Path Scope** (`Path: "/"`)
- **Current**: Cookie sent to ALL routes on the domain
- **Risk**: If domain hosts other applications, they receive the cookie
- **Recommendation**: Narrow scope to `/api/v1/admin` or `/admin` if possible

#### 2. **Domain Attribute Missing**
- **Current**: Cookie applies to current domain and all subdomains
- **Risk**: Subdomain takeover could steal cookies
- **Recommendation**: Add explicit `Domain` attribute to limit scope

#### 3. **Session Renewal**
- **Current**: Fixed 1-hour expiration, no renewal mechanism
- **Risk**: Active users forced to re-authenticate every hour
- **Recommendation**: Implement sliding expiration (refresh on activity)

#### 4. **No Session Invalidation Endpoint**
- **Current**: No logout endpoint visible in routes
- **Risk**: Users cannot explicitly end sessions
- **Recommendation**: Add `/admin/logout` to delete session from DB and cookie

#### 5. **HTTPS Enforcement**
- **Current**: `Secure: true` requires HTTPS
- **Critical**: Verify production uses TLS certificates
- **Check**: Ensure no HTTP fallback in production

#### 6. **Token Timing Attacks**
- **Current**: Hash comparison may be vulnerable
- **Risk**: Timing attacks could leak token information
- **Recommendation**: Use `crypto/subtle.ConstantTimeCompare` for hash validation

---

## How Middleware Validates Sessions

**From `middleware/admin_auth.go`:**

```go
func RequireAdmin(store admin.AdminStore) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Extract cookie from request
            cookie, err := r.Cookie("admin_session")

            // 2. Hash the cookie value
            tokenHash := utils.HashToken(cookie.Value)

            // 3. Validate hash exists and not expired in database
            email, err := store.FindValidSession(r.Context(), tokenHash)

            // 4. Add admin email to request context
            ctx := context.WithValue(r.Context(), AdminEmailKey, email)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**Security Flow:**
1. Browser automatically sends `Cookie: admin_session=<token>`
2. Server hashes received token
3. Looks up hash in database
4. Checks expiration timestamp
5. If valid: grants access + adds admin email to context
6. If invalid/expired: returns 401 Unauthorized

---

## Overall Security Rating

**7.5/10 - Good with Room for Improvement**

**Strengths:**
- Strong token generation (crypto-secure)
- Proper cookie security flags (HttpOnly, Secure, SameSite)
- Token hashing in database
- Time-limited sessions
- Magic link one-time use

**Improvements Needed:**
- Add explicit Domain attribute to cookies
- Implement session logout
- Reduce cookie Path scope
- Add session renewal mechanism
- Use constant-time comparison for tokens
- Ensure HTTPS is enforced in production

---

## Quick Security Checklist

- [x] Cryptographically secure token generation
- [x] HttpOnly flag (prevents XSS)
- [x] Secure flag (requires HTTPS)
- [x] SameSite protection (prevents CSRF)
- [x] Token hashing in database
- [x] Session expiration
- [ ] Explicit Domain attribute
- [ ] Logout endpoint
- [ ] Session renewal/sliding expiration
- [ ] Constant-time comparison
- [ ] Narrow cookie Path scope
- [ ] Production HTTPS verification

---

## Good YouTube videos
- [Cookies, Sessions, & Tokens Explained in 12 Minutes](https://www.youtube.com/watch?v=NlvngHl0cdc)
- [HTTP Cookies Crash Course](https://www.youtube.com/watch?v=sovAIX4doOE)
---

## Conclusion

The authentication system is **fundamentally secure** with strong cryptographic foundations. The callback function correctly implements session cookie creation with appropriate security flags. The primary areas for improvement are operational (logout, session management) rather than cryptographic vulnerabilities.

**Most Critical**: Ensure production environment uses HTTPS/TLS, as the `Secure: true` flag makes cookies unusable over HTTP.