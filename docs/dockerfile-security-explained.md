# Dockerfile Deployment Stage Explained

This document explains the deployment stage of our Dockerfile and the critical security benefits of running containers as a non-root user.

## Deployment Stage Overview

The deployment stage (lines 37-63) is where your application runs in production. It's designed for security, minimal size, and production readiness.

```dockerfile
FROM debian:bookworm-slim AS deploy
```

## Command-by-Command Breakdown

### Line 38: Base Image Selection
```dockerfile
FROM debian:bookworm-slim AS deploy
```
- **`debian:bookworm-slim`**: Minimal Debian 12 image
- **Why slim?** Reduces attack surface - fewer packages = fewer vulnerabilities
- **Why not scratch or alpine?** You need `ca-certificates` for HTTPS, which requires a minimal OS
- **Why Debian?** Excellent security track record and compatibility with Go binaries

### Line 40: Set Working Directory
```dockerfile
WORKDIR /app
```
Sets `/app` as the current directory for all subsequent commands. Creates the directory if it doesn't exist.

### Lines 42-45: Install CA Certificates
```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```
**What it does:**
- **`apt-get update`**: Updates package lists
- **`ca-certificates`**: Installs root CA certificates for HTTPS/TLS verification
- **`--no-install-recommends`**: Skips suggested/recommended packages (keeps image small)
- **`rm -rf /var/lib/apt/lists/*`**: Deletes apt cache to reduce image size

**Why in one RUN command?**
- Creates a single Docker layer (smaller final image)
- Cleanup happens in the same layer as installation

**Why ca-certificates?**
Your app needs these to:
- Connect to PostgreSQL with SSL/TLS
- Verify HTTPS connections to external APIs
- Trust certificate chains from Certificate Authorities

### Line 48: Create Non-Root User
```dockerfile
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
```

**Command breakdown:**

**`useradd`** - Creates a new user
- **`-m`**: Create a home directory for the user (`/home/appuser`)
- **`-u 1000`**: Assign user ID 1000 (standard first non-root user ID in Linux)
- **`appuser`**: The username

**Why UID 1000?**
- Standard convention for the first regular user
- Matches typical host user IDs (easier for volume permissions)
- Predictable and debuggable

**`chown -R appuser:appuser /app`** - Changes ownership
- **`-R`**: Recursive (all files and subdirectories)
- **`appuser:appuser`**: Set owner to `appuser` user and `appuser` group
- Ensures the app can read/write in `/app`

### Lines 50-51: Prepare Certificate Directory
```dockerfile
RUN mkdir -p /app/certs && chown appuser:appuser /app/certs
```
Creates `/app/certs` directory and gives `appuser` ownership so it can:
- Read SSL certificates for PostgreSQL connections
- Write certificates if they're dynamically generated
- Access without permission errors

### Line 54: Copy Binary from Build Stage
```dockerfile
COPY --from=build-prod /app/bin/testers-admin-api ./testers-admin-api
```
**Multi-stage build magic:**
- **`--from=build-prod`**: Copy from the `build-prod` stage
- Copies **only** the compiled binary
- Leaves behind: Go compiler, source code, build tools, dependencies
- Result: Tiny, secure final image

**Security benefit:**
- No source code in production image
- No build tools attackers could use
- Minimal attack surface

### Line 57: Switch to Non-Root User
```dockerfile
USER appuser
```

**THIS IS THE CRITICAL SECURITY LINE!**

Changes the user context for:
- All subsequent Dockerfile commands
- **The running container** (when `CMD` executes)

After this line, everything runs as `appuser`, not `root`.

### Lines 60-63: Expose Port and Run
```dockerfile
EXPOSE 8080
CMD ["./testers-admin-api"]
```
- **`EXPOSE 8080`**: Documents that the app listens on port 8080
  - Doesn't actually publish the port (that's done with `-p` flag)
  - Acts as documentation for developers
- **`CMD`**: Runs the binary **as `appuser`** (because of line 57)

---

## Why Non-Root User Protects Your Container

### The Root Problem

By default, Docker containers run as **root (UID 0)** inside the container. This creates severe security risks.

### Security Threat Scenarios

#### 1. Container Breakout Risk

**The Attack:**
If an attacker finds a vulnerability in:
- Your application code (SQL injection, RCE, etc.)
- Docker daemon
- Linux kernel

They could escape the container with **root privileges**.

**Without non-root user:**
```
Attacker exploits vulnerability
  ↓
Gains root inside container
  ↓
Exploits kernel/Docker vulnerability
  ↓
Escapes container AS ROOT
  ↓
Owns the entire host machine 💀
```

**With non-root user:**
```
Attacker exploits vulnerability
  ↓
Gains appuser privileges (UID 1000)
  ↓
Attempts privilege escalation → Much harder
  ↓
If escapes, only has limited user permissions
  ↓
Can't install packages, modify system, or easily escalate
```

#### 2. File System Protection

**Running as root allows:**
- Modify any file in the container
- Delete system binaries (`rm -rf /bin/bash`)
- Change permissions on critical files
- Install malware system-wide

**Running as appuser:**
- Can only write to files owned by `appuser`
- Can't modify `/bin`, `/usr`, `/etc`
- Can't accidentally break the system
- Limits damage from bugs or exploits

#### 3. Process Capabilities

**Root user has dangerous capabilities:**
- `CAP_NET_ADMIN`: Modify network settings
- `CAP_SYS_ADMIN`: Mount filesystems, change namespaces
- `CAP_SYS_MODULE`: Load kernel modules
- `CAP_SETUID/SETGID`: Change user/group IDs

**Non-root user:**
- Has minimal capabilities
- Can't perform privileged operations
- Constrained by standard user limits

#### 4. Lateral Movement Prevention

**Attack scenario with root:**
```
Compromise Container A (as root)
  ↓
Access shared Docker socket
  ↓
Start new privileged container
  ↓
Mount host filesystem
  ↓
Compromise host + all containers
```

**With non-root user:**
- Can't access Docker socket (permission denied)
- Can't create privileged containers
- Isolated from other containers
- Lateral movement blocked

### Real-World Attack Example

**CVE-2019-5736 (runc Container Breakout)**

This vulnerability allowed an attacker to:
1. Execute code inside a container
2. Exploit runc to overwrite the host's runc binary
3. Gain root on the host machine

**Impact:**
- **With root container**: Full compromise of host
- **With non-root container**: Significantly harder to exploit, limited impact even if successful

---

## Principle of Least Privilege

Your application **doesn't need** root privileges to:
- ✅ Listen on port 8080 (ports > 1024 don't require root)
- ✅ Read configuration files
- ✅ Write logs
- ✅ Connect to PostgreSQL
- ✅ Process HTTP requests

Running as `appuser` means:
- App has **only** the permissions it needs
- Follows security best practice
- Limits blast radius of any compromise

---

## Defense in Depth Strategy

Your Dockerfile implements multiple security layers:

```
┌──────────────────────────────────────────┐
│ 1. Minimal base image (bookworm-slim)   │ ← Fewer packages = fewer vulnerabilities
├──────────────────────────────────────────┤
│ 2. Multi-stage build                    │ ← No build tools in final image
├──────────────────────────────────────────┤
│ 3. Static binary (CGO_ENABLED=0)        │ ← No dynamic library dependencies
├──────────────────────────────────────────┤
│ 4. Non-root user (appuser)              │ ← Limited permissions
├──────────────────────────────────────────┤
│ 5. Specific file permissions            │ ← Ownership and access control
└──────────────────────────────────────────┘
```

Each layer provides protection even if another layer fails.

---

## Additional Security Benefits

### 1. Kubernetes Security Policies

Many production Kubernetes clusters enforce:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
```

Your container automatically passes these checks!

### 2. Compliance Requirements

Meets security standards:
- **PCI-DSS**: Prohibits running as root
- **CIS Docker Benchmark**: Recommends non-root users
- **NIST**: Principle of least privilege
- **SOC 2**: Access control requirements

### 3. Volume Permission Issues

**Common problem:**
```dockerfile
# Container runs as root
# Creates files in mounted volume as root
# Host user can't modify them!
```

**Your solution:**
```dockerfile
# Container runs as UID 1000
# Creates files as UID 1000
# Host user (usually UID 1000) can access them
```

### 4. Audit and Forensics

When investigating security incidents:
- **Root**: "Who did this? Root. But which process?"
- **appuser**: "This was definitely our API, not system process"

Clear separation makes forensics easier.

---

## Best Practices Demonstrated

Your Dockerfile follows industry best practices:

| Practice | Implementation | Benefit |
|----------|---------------|---------|
| ✅ Multi-stage build | Separate build and deploy stages | Smaller, more secure image |
| ✅ Minimal base image | `debian:bookworm-slim` | Reduced attack surface |
| ✅ Static binary | `CGO_ENABLED=0` | No runtime dependencies |
| ✅ Non-root user | `USER appuser` | Principle of least privilege |
| ✅ Explicit ownership | `chown appuser:appuser` | Clear permissions |
| ✅ Single-purpose layers | Combined commands with `&&` | Optimized layer caching |
| ✅ Cache cleanup | `rm -rf /var/lib/apt/lists/*` | Smaller image size |

---

## Even Better: Distroless Alternative

For maximum security, consider Google's distroless images:

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot AS deploy

COPY --from=build-prod /app/bin/testers-admin-api /testers-admin-api
COPY --from=build-prod /app/certs /certs

EXPOSE 8080
CMD ["/testers-admin-api"]
```

**Benefits:**
- No shell (`/bin/sh` doesn't exist)
- No package manager
- Only your app and minimal runtime
- Pre-configured as non-root user (`nonroot` user, UID 65532)

**Trade-offs:**
- Can't `docker exec` into container for debugging
- No shell for troubleshooting
- Requires static binaries (already done with `CGO_ENABLED=0`)

---

## Testing Non-Root User

Verify your container runs as non-root:

```bash
# Build the image
docker build -t testers-admin:test .

# Run and check the user
docker run --rm testers-admin:test id

# Expected output:
# uid=1000(appuser) gid=1000(appuser) groups=1000(appuser)
```

Not `uid=0(root)` ✓

---

## Summary

The non-root user in your Dockerfile:

1. **Prevents privilege escalation** - Attackers can't gain root easily
2. **Limits file system damage** - Can't modify system files
3. **Blocks container escapes** - Even if successful, limited host access
4. **Follows least privilege** - App has only needed permissions
5. **Meets compliance** - Passes security audits and policies
6. **Improves forensics** - Clear separation of processes
7. **Fixes volume permissions** - Files created with proper ownership

**Bottom line:** This simple `USER appuser` command dramatically reduces your attack surface and is essential for production containers.
