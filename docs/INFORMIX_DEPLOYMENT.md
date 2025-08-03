# Bytebase with Informix Support - Deployment Guide

This guide explains how to deploy Bytebase with IBM Informix database support using Docker containers.

## Prerequisites

- Docker and Docker Compose installed
- At least 8GB of available RAM
- 20GB of free disk space
- x86_64 (AMD64) architecture (ARM/Apple Silicon requires Rosetta emulation)

## Quick Start

1. **Clone the repository** (if not already done):
   ```bash
   git clone https://github.com/bytebase/bytebase.git
   cd bytebase
   ```

2. **Start Bytebase with Informix support**:
   ```bash
   ./scripts/start-informix.sh
   ```

3. **Access Bytebase**:
   - Web UI: http://localhost:8080
   - Default credentials will be shown on first access

## Architecture Overview

The deployment consists of three main containers:

1. **PostgreSQL** - Stores Bytebase metadata
2. **Informix** - Test database instance
3. **Bytebase** - Main application with Informix ODBC driver

## Manual Deployment

### Build the Docker image

```bash
./scripts/build-informix.sh
```

Or manually:

```bash
docker build -f docker/Dockerfile.informix -t bytebase-informix:latest .
```

### Start services

```bash
docker-compose -f docker-compose.informix.yml up -d
```

### Stop services

```bash
docker-compose -f docker-compose.informix.yml down
```

## Configuration

### Environment Variables

- `PG_URL` - PostgreSQL connection string for Bytebase metadata
- `INFORMIXDIR` - Informix installation directory (default: /opt/IBM/informix)
- `INFORMIXSERVER` - Informix server name (default: informix_tcp)

### ODBC Configuration

The ODBC configuration files are located in `docker/informix/`:
- `odbcinst.ini` - ODBC driver configuration
- `odbc.ini` - Data source configuration
- `sqlhosts` - Informix server configuration

### Custom Informix Connection

To connect to an external Informix database:

1. Update `docker/informix/sqlhosts` with your server details
2. Modify `docker/informix/odbc.ini` with your connection parameters
3. Rebuild and restart the containers

## Adding Informix Instances in Bytebase

1. Navigate to **Environments** in Bytebase
2. Click **Add Instance**
3. Select **Informix** as the database type
4. Configure connection:
   - Host: `informix` (for the bundled instance) or your external host
   - Port: `9088` (default)
   - Username: `informix`
   - Password: `in4mix`
   - Database: `sysmaster` (or your target database)

## Troubleshooting

### View logs

```bash
# All services
docker-compose -f docker-compose.informix.yml logs

# Bytebase only
docker-compose -f docker-compose.informix.yml logs bytebase

# Follow logs
docker-compose -f docker-compose.informix.yml logs -f
```

### Test ODBC connection

```bash
# Enter Bytebase container
docker-compose -f docker-compose.informix.yml exec bytebase bash

# Test ODBC connection
isql -v informix_default informix in4mix
```

### Common Issues

1. **Connection refused**: Ensure Informix container is healthy
   ```bash
   docker-compose -f docker-compose.informix.yml ps
   ```

2. **ODBC driver not found**: Check driver installation
   ```bash
   docker-compose -f docker-compose.informix.yml exec bytebase odbcinst -q -d
   ```

3. **Permission issues**: Ensure proper file permissions
   ```bash
   chmod +x scripts/*.sh
   ```

## Development Mode

For development with hot reload:

```bash
# Start infrastructure only
docker-compose -f docker-compose.informix.yml up -d postgres informix

# Run Bytebase locally with Informix tag
CGO_ENABLED=1 go run -tags informix ./backend/bin/server/main.go \
  --data /tmp/bytebase \
  --port 8080 \
  --pg postgresql://bbdev:bbdev@localhost:5432/bbdev
```

## Production Considerations

1. **Security**:
   - Change default passwords
   - Use SSL/TLS for connections
   - Implement network isolation

2. **Performance**:
   - Allocate sufficient resources to containers
   - Monitor ODBC connection pool
   - Consider using persistent volumes for data

3. **Backup**:
   - Regular backups of PostgreSQL metadata
   - Informix database backups as needed

## Limitations

- Currently supports Linux x86_64 only
- ODBC driver adds some performance overhead
- Some advanced Informix features may not be fully supported

## Support

For issues specific to Informix support:
1. Check the [Bytebase documentation](https://www.bytebase.com/docs)
2. Report issues on [GitHub](https://github.com/bytebase/bytebase/issues)
3. Join the [Bytebase community](https://bytebase.com/community)