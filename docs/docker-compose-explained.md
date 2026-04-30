# Docker Compose Configuration Explained

This document explains every command and configuration in your `docker-compose.yml` file to help you understand how Docker Compose orchestrates your application.

## Table of Contents
- [What is Docker Compose?](#what-is-docker-compose)
- [File Structure Overview](#file-structure-overview)
- [Services Section](#services-section)
  - [API Service](#api-service)
  - [Database Service](#database-service)
- [Volumes Section](#volumes-section)
- [Networks Section](#networks-section)

---

## What is Docker Compose?

Docker Compose is a tool for defining and running multi-container Docker applications. With a single `docker-compose.yml` file, you can configure all your application's services, networks, and volumes, then start everything with one command: `docker-compose up`.

---

## File Structure Overview

Your docker-compose.yml has three main sections:
1. **services**: Defines the containers that make up your application
2. **volumes**: Defines persistent data storage
3. **networks**: Defines how containers communicate with each other

---

## Services Section

### API Service

#### `api:`
The name of your first service. This creates a container for your API application.

#### `build:`
Tells Docker to build an image from a Dockerfile instead of pulling a pre-built image.

- **`context: .`**: The build context is the current directory (`.`). Docker will look for a Dockerfile here and include all files from this directory in the build context.

- **`target: development`**: Specifies which build stage to use from a multi-stage Dockerfile. This targets the stage named "development", allowing you to have different configurations for development vs production.

#### `container_name: testers-admin-api`
Assigns a specific name to the container. Without this, Docker Compose would generate a name automatically (like `project_api_1`). A custom name makes it easier to reference the container in commands like `docker logs testers-admin-api`.

#### `ports:`
Maps ports between your host machine and the container.

- **`"${PORT:-8080}:8080"`**:
  - Format: `"HOST_PORT:CONTAINER_PORT"`
  - `${PORT:-8080}` reads the PORT variable from your environment (or .env file), defaulting to 8080 if not set
  - Maps that port on your host to port 8080 inside the container
  - Example: If PORT=3000 in .env, you'd access the API at `localhost:3000`, which forwards to port 8080 inside the container

#### `environment:`
Sets environment variables inside the container. These are available to your application at runtime.

- **`PORT: ${PORT:-8080}`**: Sets the PORT variable inside the container, using the value from your .env file or defaulting to 8080
- **`DB_NAME: ${DB_NAME}`**: Database name from your .env file
- **`DB_USER: ${DB_USER}`**: Database username from your .env file
- **`DB_PASSWORD: ${DB_PASSWORD}`**: Database password from your .env file
- **`DB_HOST: ${DB_HOST}`**: Database host (likely "db" to connect to the database service)
- **`DB_SSLMODE: ${DB_SSLMODE:-require}`**: PostgreSQL SSL mode, defaults to "require" for secure connections

#### `env_file:`
Loads environment variables from a file.

- **`./.env`**: Loads variables from the .env file in the current directory. Variables here are available in the docker-compose.yml (using `${VAR}` syntax) and inside the container.

#### `depends_on:`
Controls the startup order of services.

- **`db:`**: This service depends on the "db" service
  - **`condition: service_healthy`**: The API container won't start until the database passes its healthcheck. This prevents your API from trying to connect to a database that isn't ready yet.

#### `volumes:`
Mounts directories or files from your host machine into the container.

- **`.:/app`**:
  - Format: `HOST_PATH:CONTAINER_PATH`
  - Mounts your current directory (`.`) to `/app` inside the container
  - Changes to files on your host are immediately reflected in the container (useful for development)

- **`/app/tmp`**:
  - Creates an anonymous volume for `/app/tmp`
  - This prevents the host's tmp directory from being mounted, keeping temporary files separate

- **`./certs/root.crt:/app/certs/root.crt:ro`**:
  - Mounts the SSL root certificate from your host to the container
  - `:ro` means "read-only" - the container can read but not modify this file

#### `networks:`
Connects this service to specified networks.

- **`testers-admin-network`**: Connects the API container to this custom network, allowing it to communicate with other services on the same network (like the database).

#### `restart: unless-stopped`
Container restart policy. The container will automatically restart if it crashes or Docker restarts, unless you manually stop it with `docker stop`.

---

### Database Service

#### `db:`
The name of your database service.

#### `image: postgres:17-alpine`
Uses a pre-built image instead of building from a Dockerfile.
- `postgres:17-alpine` is PostgreSQL version 17 on Alpine Linux (a minimal Linux distribution, resulting in smaller image sizes)

#### `container_name: postgres17-testers`
Names the database container for easy reference.

#### `command:`
Overrides the default command that runs when the container starts. This passes configuration flags to PostgreSQL.

- **`"postgres"`**: The PostgreSQL server executable
- **`"-c"`**: Flag to set a configuration parameter
- **`"ssl=on"`**: Enables SSL/TLS encryption for database connections
- **`"ssl_cert_file=/var/lib/postgresql/certs/server.crt"`**: Path to the server's SSL certificate
- **`"ssl_key_file=/var/lib/postgresql/certs/server.key"`**: Path to the server's private key
- **`"ssl_ca_file=/var/lib/postgresql/certs/root.crt"`**: Path to the Certificate Authority (CA) root certificate
- **`"ssl_prefer_server_ciphers=on"`**: The server's cipher preferences take priority over the client's
- **`"ssl_min_protocol_version=TLSv1.2"`**: Requires at least TLS 1.2 for security (blocks older, vulnerable protocols)

#### `environment:`
Environment variables for PostgreSQL initialization.

- **`POSTGRES_USER: ${DB_USER}`**: Sets the PostgreSQL superuser name
- **`POSTGRES_PASSWORD: ${DB_PASSWORD}`**: Sets the superuser password
- **`POSTGRES_DB: ${DB_NAME}`**: Creates this database on first startup
- **`POSTGRES_HOST_SSL: "on"`**: Enforces SSL for host connections

#### `env_file:`
- **`./.env`**: Loads environment variables from the .env file

#### `ports:`
- **`"5432:5432"`**: Maps PostgreSQL's default port (5432) from the container to your host machine. This allows you to connect to the database from your host using tools like `psql` or database GUIs.

#### `volumes:`
Mounts for persistent data and configuration.

- **`db_data:/var/lib/postgresql/data`**:
  - Named volume for database data persistence
  - Without this, all database data would be lost when the container stops
  - `db_data` refers to the named volume defined in the volumes section

- **`./certs/server.crt:/var/lib/postgresql/certs/server.crt:ro`**: Mounts the server SSL certificate (read-only)
- **`./certs/server.key:/var/lib/postgresql/certs/server.key:ro`**: Mounts the server private key (read-only)
- **`./certs/root.crt:/var/lib/postgresql/certs/root.crt:ro`**: Mounts the CA root certificate (read-only)
- **`./config/pg_hba.conf:/var/lib/postgresql/data/pg_hba.conf:ro`**:
  - Mounts the PostgreSQL host-based authentication configuration
  - This file controls which hosts can connect and how they must authenticate

#### `networks:`
- **`testers-admin-network`**: Connects to the same network as the API service, enabling communication between containers

#### `healthcheck:`
Defines how Docker checks if the container is healthy.

- **`test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]`**:
  - Runs the `pg_isready` command to check if PostgreSQL is accepting connections
  - `pg_isready` is a PostgreSQL utility that checks server status
  - Returns 0 (success) if the database is ready

- **`interval: 5s`**: Run the healthcheck every 5 seconds

- **`timeout: 5s`**: If the healthcheck command doesn't respond within 5 seconds, consider it failed

- **`retries: 5`**: The container is considered unhealthy after 5 consecutive failures

#### `restart: unless-stopped`
Same as the API service - automatically restarts unless manually stopped.

---

## Volumes Section

### `volumes:`
Defines named volumes that can be referenced by services.

#### `db_data:`
A named volume for PostgreSQL data persistence.

- **`name: testers_admin_db_data`**: The actual name of the volume on your system. Without this, Docker Compose would prefix it with the project name.

**Why use named volumes?**
- Data persists even if you delete and recreate containers
- Better performance than bind mounts on some systems
- Docker manages the volume's location on disk
- Can be backed up, moved between hosts, etc.

---

## Networks Section

### `networks:`
Defines custom networks for container communication.

#### `testers-admin-network:`
A custom network that both services connect to.

- **`name: testers-admin-network`**: The actual network name on your system

- **`driver: bridge`**:
  - The network type
  - Bridge networks create an isolated network on your host
  - Containers on the same bridge network can communicate using service names as hostnames
  - Example: Your API container can connect to the database using `DB_HOST=db` because both are on this network

**Why use custom networks?**
- Containers can communicate using service names (e.g., `db` instead of IP addresses)
- Provides isolation from other Docker containers not on this network
- Allows you to control which containers can talk to each other

---

## Common Docker Compose Commands

```bash
# Start all services (builds images if needed)
docker-compose up

# Start in detached mode (runs in background)
docker-compose up -d

# Stop all services
docker-compose down

# Stop and remove volumes (deletes all data!)
docker-compose down -v

# View logs for all services
docker-compose logs

# View logs for specific service
docker-compose logs api

# Follow logs in real-time
docker-compose logs -f

# Rebuild images
docker-compose build

# Rebuild and start
docker-compose up --build

# List running containers
docker-compose ps

# Execute command in running container
docker-compose exec api sh
docker-compose exec db psql -U ${DB_USER} -d ${DB_NAME}
```

---

## Key Concepts Summary

### Environment Variables
- Defined in `.env` file
- Referenced in docker-compose.yml using `${VAR_NAME}`
- Can have defaults: `${VAR_NAME:-default_value}`

### Volumes
- **Bind mounts** (`./host/path:/container/path`): Links host directory to container
- **Named volumes** (`volume_name:/container/path`): Managed by Docker for data persistence
- **Anonymous volumes** (`/container/path`): Temporary, deleted with container

### Networks
- Allow containers to communicate using service names
- Provide isolation between different applications
- Bridge driver is most common for single-host setups

### Health Checks
- Verify a service is actually working, not just running
- Used with `depends_on` to ensure proper startup order
- Prevent cascading failures when services aren't ready

### Restart Policies
- `no`: Never restart (default)
- `always`: Always restart
- `on-failure`: Restart only if container exits with error
- `unless-stopped`: Always restart unless manually stopped

---

## Security Features in Your Configuration

1. **SSL/TLS Encryption**: Database connections are encrypted
2. **Read-only mounts** (`:ro`): Certificates can't be modified by containers
3. **Isolated network**: Only containers on the network can communicate
4. **Environment file**: Keeps secrets out of docker-compose.yml
5. **Health checks**: Ensures database is properly configured before API connects
6. **TLS 1.2 minimum**: Blocks outdated, vulnerable encryption protocols