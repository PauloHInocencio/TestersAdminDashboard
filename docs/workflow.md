### 1. Copy environment template
```bash
cp .env.example .env
```

### 2. Edit .env with actual values
(DB_PASSWORD, etc.)

### 3. Generate SSL certificates
```bash
make ssl-certs
```
OR manually: bash scripts/generate-ssl-certs.sh

### 4. Create config directory and required files

```bash
mkdir -p config
```

### 5. Create pg_hba.conf (REQUIRED - enforces SSL)
Copy the complete code from section 7 above into config/pg_hba.conf

### 6. Create postgresql.conf (OPTIONAL - for advanced tuning)
- Copy the code from section 6 above into config/postgresql.conf if desired
- SSL settings are already configured via docker-compose command arguments

## 7. Start services
```bash
make docker-up
```

OR manually: 
```bash
docker compose up -d
```

## 8. View logs
```bash
docker compose logs -f api
```
