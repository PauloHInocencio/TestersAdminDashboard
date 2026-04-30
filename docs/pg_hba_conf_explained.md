# PostgreSQL pg_hba.conf File Explained

## What is pg_hba.conf?

The `pg_hba.conf` file is PostgreSQL's **Host-Based Authentication** configuration file. It controls client authentication - determining which users can connect to which databases from which hosts and what authentication methods they must use.

## Purpose

The file acts as a security gateway for your PostgreSQL database by defining access rules based on:

- **Database name** - which database(s) the rule applies to
- **User/role** - which PostgreSQL user(s) can connect
- **Host/IP address** - where the connection is coming from
- **Authentication method** - how the user must authenticate

## How It Works

Rules are evaluated **top to bottom**, and the **first matching rule** is used. This means:

1. PostgreSQL reads the file from top to bottom
2. When a connection attempt is made, it checks each rule in order
3. The first rule that matches the connection parameters is applied
4. All subsequent rules are ignored for that connection

### Rule Format

Each line in the file follows this format:

```
TYPE  DATABASE  USER  ADDRESS  METHOD
```

**Fields:**
- **TYPE**: Connection type (`local`, `host`, `hostssl`, `hostnossl`)
- **DATABASE**: Database name(s) or `all`
- **USER**: PostgreSQL username(s) or `all`
- **ADDRESS**: IP address/CIDR (for `host` types) or omitted for `local`
- **METHOD**: Authentication method to use

### Example Configuration

```conf
# Allow local connections without password (for development only!)
local   all       all                          trust

# Require password for remote connections
host    all       all        0.0.0.0/0         scram-sha-256

# Allow specific user from specific IP with MD5
host    mydb      john       192.168.1.100/32  md5

# Require SSL for remote admin connections
hostssl postgres  postgres   192.168.1.0/24    scram-sha-256

# Reject all other connections
host    all       all        0.0.0.0/0         reject
```

## Common Authentication Methods

| Method | Description | Security Level | Use Case |
|--------|-------------|----------------|----------|
| `trust` | No password required | ⚠️ Very Low | Local development only |
| `reject` | Reject connection | N/A | Explicitly deny access |
| `md5` | MD5-hashed password | Medium (deprecated) | Legacy systems |
| `scram-sha-256` | Modern secure hashing | ✅ High | Production (recommended) |
| `peer` | Use OS username | Medium | Local connections only |
| `password` | Plain text password | ⚠️ Very Low | Never use in production |
| `cert` | SSL certificate | ✅ High | High-security environments |

## Connection Types

- **`local`**: Unix-domain socket connections (same machine only)
- **`host`**: TCP/IP connections (with or without SSL)
- **`hostssl`**: TCP/IP connections with SSL **required**
- **`hostnossl`**: TCP/IP connections **without** SSL

## Is This Standard in Backend Engineering?

While `pg_hba.conf` is **PostgreSQL-specific**, the concept of host-based authentication is common across database systems:

### Similar Concepts in Other Databases

- **MySQL**: Uses `GRANT` statements and `bind-address` configuration
- **MongoDB**: IP whitelisting and authentication configuration
- **Redis**: `bind` and `requirepass` directives
- **Most databases**: Some form of connection access control

### Industry Standard Practice

This is a **database administration fundamental**. Every production PostgreSQL deployment requires proper `pg_hba.conf` configuration as part of:

- **Defense in depth** security model
- **Principle of least privilege**
- **Network security** best practices
- **Compliance requirements** (PCI-DSS, HIPAA, etc.)

## Best Practices

### ✅ DO

1. **Use strong authentication methods**
   ```conf
   host    all    all    0.0.0.0/0    scram-sha-256
   ```

2. **Restrict by IP when possible**
   ```conf
   host    mydb   myuser   192.168.1.0/24   scram-sha-256
   ```

3. **Require SSL for remote connections**
   ```conf
   hostssl   all   all   0.0.0.0/0   scram-sha-256
   ```

4. **Place more specific rules first**
   ```conf
   # Specific rule first
   host    sensitive_db   admin   10.0.0.5/32      scram-sha-256
   # General rule later
   host    all           all     10.0.0.0/8       scram-sha-256
   ```

5. **Explicitly reject unwanted connections**
   ```conf
   host    all   all   192.168.1.0/24   scram-sha-256
   host    all   all   0.0.0.0/0        reject
   ```

### ⚠️ DON'T

1. **Never use `trust` for remote connections**
   ```conf
   # DANGEROUS!
   host    all   all   0.0.0.0/0   trust
   ```

2. **Don't use `md5` for new deployments** (use `scram-sha-256` instead)

3. **Don't allow connections from everywhere without good reason**
   ```conf
   # Too permissive unless you have firewall rules
   host    all   all   0.0.0.0/0   scram-sha-256
   ```

4. **Don't forget to reload PostgreSQL after changes**

## Applying Changes

After modifying `pg_hba.conf`, reload PostgreSQL:

```bash
# Using pg_ctl
pg_ctl reload

# Using systemctl (Linux)
sudo systemctl reload postgresql

# Using SQL
SELECT pg_reload_conf();

# Docker
docker exec <container> pg_ctl reload
```

## Common Issues

### Issue: "No pg_hba.conf entry for host..."

**Cause**: No matching rule found for the connection attempt

**Solution**: Add appropriate rule or check IP address/username

### Issue: Changes not taking effect

**Cause**: PostgreSQL not reloaded after configuration change

**Solution**: Reload PostgreSQL (see above)

### Issue: Authentication fails even with correct password

**Cause**: Wrong authentication method or password encoding

**Solution**: Ensure client and server agree on authentication method (prefer `scram-sha-256`)

## Location of pg_hba.conf

Default locations by operating system:

- **Linux**: `/etc/postgresql/{version}/main/pg_hba.conf`
- **macOS (Homebrew)**: `/usr/local/var/postgres/pg_hba.conf`
- **Windows**: `C:\Program Files\PostgreSQL\{version}\data\pg_hba.conf`
- **Docker**: Usually mounted or in `/var/lib/postgresql/data/pg_hba.conf`

Find current location:
```sql
SHOW hba_file;
```

## Summary

The `pg_hba.conf` file is a critical security component in PostgreSQL that:

- Controls **who** can connect (users)
- From **where** they can connect (hosts/IPs)
- To **which databases** they can connect
- Using **what authentication method**

Proper configuration is essential for database security and is a fundamental skill for backend engineers working with PostgreSQL.
