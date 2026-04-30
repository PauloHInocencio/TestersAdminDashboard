# PostgreSQL Initialization Script Explained

## Overview
The `db/init/01-configure-ssl.sh` script is automatically executed by the PostgreSQL Docker container during database initialization (first startup only). It configures SSL enforcement by properly installing the custom `pg_hba.conf` file.

## Script Location
`db/init/01-configure-ssl.sh`

## Why This Script Exists
PostgreSQL needs to set ownership and permissions on files in its data directory (`/var/lib/postgresql/data`). Mounting `pg_hba.conf` directly as read-only caused conflicts because PostgreSQL couldn't modify file permissions. This initialization script solves that by copying the configuration file during startup with proper permissions.

## Command-by-Command Breakdown

### Line 1: `#!/bin/bash`
**Purpose**: Shebang line that specifies the script should be executed using the Bash shell.

**Why it matters**: Ensures the script runs with Bash, which supports all the syntax used in the script.

---

### Line 2: `set -e`
**Purpose**: Exit immediately if any command fails (returns non-zero exit code).

**Why it matters**:
- Prevents the script from continuing if something goes wrong
- Makes debugging easier by stopping at the first error
- Ensures PostgreSQL won't start with incorrect configuration

**Example**: If the `cp` command fails, the script stops immediately instead of trying to run `chown` and `chmod` on a file that doesn't exist.

---

### Line 5: `if [ -f /docker-entrypoint-initdb.d/pg_hba.conf ]; then`
**Purpose**: Check if the pg_hba.conf template file exists before proceeding.

**Breakdown**:
- `if [ ... ]; then` - Conditional statement
- `-f` - Test flag that checks if the path is a regular file (not a directory or symlink)
- `/docker-entrypoint-initdb.d/pg_hba.conf` - Path where the template file is mounted

**Why it matters**:
- Prevents errors if the file is missing
- Makes the script defensive and safe to run in different environments
- The PostgreSQL Docker image automatically executes scripts from `/docker-entrypoint-initdb.d/` during initialization

---

### Line 6: `cp /docker-entrypoint-initdb.d/pg_hba.conf "$PGDATA/pg_hba.conf"`
**Purpose**: Copy the pg_hba.conf template to PostgreSQL's data directory.

**Breakdown**:
- `cp` - Copy command
- `/docker-entrypoint-initdb.d/pg_hba.conf` - Source file (mounted read-only from host)
- `"$PGDATA/pg_hba.conf"` - Destination (PostgreSQL data directory)
- `$PGDATA` - Environment variable set by PostgreSQL Docker image, typically `/var/lib/postgresql/data`

**Why it matters**:
- Creates a writable copy in the data directory
- Preserves the original template file as read-only
- PostgreSQL looks for `pg_hba.conf` in `$PGDATA` by default

---

### Line 7: `chown postgres:postgres "$PGDATA/pg_hba.conf"`
**Purpose**: Change file ownership to the postgres user and group.

**Breakdown**:
- `chown` - Change ownership command
- `postgres:postgres` - Format is `user:group`
- First `postgres` - User name
- Second `postgres` - Group name
- `"$PGDATA/pg_hba.conf"` - Target file

**Why it matters**:
- PostgreSQL runs as the `postgres` user inside the container
- PostgreSQL needs to be able to read this file
- Proper ownership is a security requirement
- Without this, PostgreSQL might not be able to reload the configuration

---

### Line 8: `chmod 600 "$PGDATA/pg_hba.conf"`
**Purpose**: Set file permissions to read/write for owner only.

**Breakdown**:
- `chmod` - Change mode (permissions) command
- `600` - Permission octal notation
  - `6` (owner) = read (4) + write (2) = 6
  - `0` (group) = no permissions
  - `0` (others) = no permissions
- Equivalent to `-rw-------` (owner can read/write, no one else can access)

**Why it matters**:
- Security requirement: pg_hba.conf contains authentication rules
- PostgreSQL will refuse to start if pg_hba.conf has overly permissive permissions
- Prevents other users from reading or modifying the authentication configuration
- Follows the principle of least privilege

---

### Line 9: `fi`
**Purpose**: Closes the `if` statement from line 5.

**Why it matters**: Required Bash syntax to end the conditional block.

---

## Execution Flow

1. **When it runs**: Only during first database initialization when the data volume is empty
2. **What happens**:
   - Script checks if template exists
   - Copies template to data directory
   - Sets correct ownership (postgres user)
   - Sets secure permissions (600)
   - PostgreSQL starts and loads the configuration

3. **What doesn't happen**: The script does NOT run on subsequent container restarts (only on first initialization)

## Docker Integration

The script is mounted in `docker-compose.yml`:
```yaml
volumes:
  - ./db/init/01-configure-ssl.sh:/docker-entrypoint-initdb.d/01-configure-ssl.sh:ro
  - ./db/config/pg_hba.conf:/docker-entrypoint-initdb.d/pg_hba.conf:ro
```

The PostgreSQL official Docker image automatically:
1. Sorts and executes all `.sh` files in `/docker-entrypoint-initdb.d/` alphabetically
2. Runs them before starting PostgreSQL for the first time
3. Skips them on subsequent starts (unless the volume is recreated)

## Updating pg_hba.conf After Initial Setup

Since the init script only runs once, to update pg_hba.conf you must:

**Option 1: Recreate the volume** (data will be lost)
```bash
docker compose down -v
make run
```

**Option 2: Manual copy** (preserves data)
```bash
docker cp ./db/config/pg_hba.conf postgres17-testers:/var/lib/postgresql/data/pg_hba.conf
docker exec postgres17-testers chown postgres:postgres /var/lib/postgresql/data/pg_hba.conf
docker exec postgres17-testers chmod 600 /var/lib/postgresql/data/pg_hba.conf
docker exec postgres17-testers psql -U $DB_USER -c "SELECT pg_reload_conf();"
```

## Security Benefits

1. **Template Protection**: Source file remains read-only and unchanged
2. **Proper Permissions**: Ensures pg_hba.conf has secure permissions (600)
3. **Correct Ownership**: File owned by postgres user, not root
4. **Fail-Safe**: Script exits on any error, preventing misconfiguration
5. **Defense in Depth**: Combines with SSL enforcement for layered security

## Related Documentation

- [pg_hba.conf Explained](./pg_hba_conf_explained.md) - Details on the configuration file itself
- [SSL Certificates Explained](./SSL_CERTIFICATES_EXPLAINED.md) - How SSL certificates work
- [Docker Compose Explained](./docker-compose-explained.md) - Overall Docker setup
