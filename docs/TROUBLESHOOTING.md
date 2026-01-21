# Troubleshooting Guide

Common issues and solutions for Zabbix Docker.

> **Note:** This guide assumes you followed the [Quick Start Guide](QUICKSTART.md) and are using `docker-compose_v3_alpine_mysql_latest.yaml`. Adjust the compose file name if you're using a different configuration.

## Simplifying Commands

To avoid typing the long compose file name repeatedly, set this environment variable:

```bash
export COMPOSE_FILE=docker-compose_v3_alpine_mysql_latest.yaml
```

Then use short commands:
```bash
docker compose ps              # instead of: docker compose -f docker-compose_v3_alpine_mysql_latest.yaml ps
docker compose logs -f         # instead of: docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs -f
docker compose restart         # instead of: docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart
```

All commands below show the full version for clarity.

---

## Containers Not Starting

### Check logs

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs
```

### Common causes

**Port conflicts:**
```bash
# Check what's using port 80
sudo lsof -i :80

# Use a different port
export ZBX_WEB_SERVER_PORT=8080
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

**Insufficient memory:**
```bash
# Check Docker resources
docker stats

# Increase Docker memory limit (Docker Desktop)
# Settings → Resources → Memory → Increase to 4GB+
```

**Database initialization failed:**
```bash
# Remove everything and start fresh
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml down -v
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d

# Watch database initialization
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs -f mysql-server
```

---

## Cannot Access Web Interface

### Port 80 already in use

Solution 1: Use different port
```bash
# Edit .env file or set environment variable
export ZBX_WEB_SERVER_PORT=8080
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

Access at: http://localhost:8080

Solution 2: Stop conflicting service
```bash
# Find what's using port 80
sudo lsof -i :80

# Stop it (example: Apache)
sudo systemctl stop apache2
```

### Firewall blocking access

```bash
# Allow port 80 (Linux)
sudo ufw allow 80/tcp

# Check if port is accessible
curl http://localhost
```

---

## Cannot Login

### Default credentials not working

Check database initialization:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs mysql-server | grep "ready for connections"
```

Restart web container:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart zabbix-web-nginx-mysql
```

Reset to defaults (WARNING: This deletes all data):
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml down -v
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

### "GUI access disabled" error

Check server configuration:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs zabbix-server | grep "GUI"
```

---

## Database Connection Errors

### "MySQL server is not available"

This is normal during startup. Wait 30-60 seconds for the database to initialize.

Watch the initialization:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs -f zabbix-server
```

You should see: `database is working`

### Connection timeout

Check database container:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml ps mysql-server
```

Restart database:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart mysql-server
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart zabbix-server
```

---

## Performance Issues

### Web interface is slow

Increase PHP memory:

Create `.env_web` file:
```bash
PHP_MEMORY_LIMIT=256M
```

Restart:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart zabbix-web-nginx-mysql
```

### Database is slow

Increase MySQL resources:

Edit `docker-compose.yaml`:
```yaml
mysql-server:
  environment:
    - MYSQL_INNODB_BUFFER_POOL_SIZE=512M
```

---

## Agent Connection Issues

### Agent cannot connect to server

Check server is listening:
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml exec zabbix-server netstat -ln | grep 10051
```

Check firewall:
```bash
# Allow Zabbix server port
sudo ufw allow 10051/tcp
```

Test connection from agent:
```bash
telnet <server-ip> 10051
```

---

## Upgrade Issues

### Data lost after upgrade

Always backup before upgrading:

```bash
# Backup database
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml exec mysql-server mysqldump -u root -p zabbix > backup.sql

# Backup volumes
docker run --rm -v zabbix_data:/data -v $(pwd):/backup alpine tar czf /backup/zabbix-backup.tar.gz /data
```

### Containers won't start after upgrade

Check compatibility:
```bash
# View current version
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml exec zabbix-server zabbix_server -V

# Check logs for errors
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs zabbix-server
```

Rollback if needed:
```bash
# Use specific version
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml down
# Edit docker-compose_v3_alpine_mysql_latest.yaml to use previous version tag
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

---

## SSL/TLS Issues

### Cannot enable HTTPS

See [EXAMPLES.md](EXAMPLES.md#https-with-lets-encrypt) for SSL configuration.

---

## Diagnostic Commands

### Check all container status
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml ps
```

### View all logs
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs
```

### Check resource usage
```bash
docker stats
```

### Inspect specific container
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml exec zabbix-server bash
```

### Check network connectivity
```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml exec zabbix-server ping mysql-server
```

---

## Getting Help

If you're still experiencing issues:

1. Check the [Wiki](https://github.com/zabbix/zabbix-docker/wiki) for known issues
2. Search [existing issues](https://github.com/zabbix/zabbix-docker/issues)
3. Ask in the [Zabbix Forum](https://www.zabbix.com/forum/)
4. Create a [new issue](https://github.com/zabbix/zabbix-docker/issues/new) with:
   - Docker version: `docker --version`
   - Compose version: `docker compose version`
   - OS: `uname -a`
   - Logs: `docker compose logs`

