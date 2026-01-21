# Configuration Examples

Real-world examples for common Zabbix Docker setups.

> **Note:** This guide assumes you followed the [Quick Start Guide](QUICKSTART.md). These examples build upon the basic setup.

## Simplifying Commands

To avoid typing the long compose file name repeatedly, set this environment variable:

```bash
export COMPOSE_FILE=docker-compose_v3_alpine_mysql_latest.yaml
```

Then you can use short commands like `docker compose up -d` instead of the full version.

Examples below show the full command for clarity.

---

## Table of Contents

- [Basic Monitoring Setup](#basic-monitoring-setup)
- [HTTPS with Let's Encrypt](#https-with-lets-encrypt)
- [External Database](#external-database)
- [High Availability Setup](#high-availability-setup)
- [Custom Configuration](#custom-configuration)
- [Monitoring Docker Containers](#monitoring-docker-containers)

---

## Basic Monitoring Setup

### Monitor the Docker Host

Start Zabbix with agent:

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml --profile full up -d
```

Add host in Zabbix UI:

- Configuration → Hosts → Create host
- Host name: `Docker Host`
- Groups: `Linux servers`
- Agent interface: `zabbix-agent:10050`
- Templates: `Linux by Zabbix agent`

Verify (wait 1-2 minutes):

- Monitoring → Latest data → Select host

---

## HTTPS with Let's Encrypt

### Using Nginx Proxy

Create `docker-compose.override.yaml`:

```yaml
version: '3.8'

services:
  nginx-proxy:
    image: nginxproxy/nginx-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/tmp/docker.sock:ro
      - ./certs:/etc/nginx/certs
    networks:
      - frontend

  zabbix-web-nginx-mysql:
    environment:
      - VIRTUAL_HOST=zabbix.example.com
      - LETSENCRYPT_HOST=zabbix.example.com
      - LETSENCRYPT_EMAIL=admin@example.com
    expose:
      - "8080"
    ports: []  # Remove direct port mapping
```

Start:

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

Access at: https://zabbix.example.com

---

## External Database

### Use existing MySQL server

Create `docker-compose.override.yaml`:

```yaml
version: '3.8'

services:
  zabbix-server:
    environment:
      - DB_SERVER_HOST=mysql.example.com
      - DB_SERVER_PORT=3306
      - MYSQL_DATABASE=zabbix
      - MYSQL_USER=zabbix
      - MYSQL_PASSWORD=your_password

  zabbix-web-nginx-mysql:
    environment:
      - DB_SERVER_HOST=mysql.example.com
      - DB_SERVER_PORT=3306
      - MYSQL_DATABASE=zabbix
      - MYSQL_USER=zabbix
      - MYSQL_PASSWORD=your_password

  mysql-server:
    # Disable local MySQL
    profiles:
      - disabled
```

Prepare external database:

```sql
CREATE DATABASE zabbix CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
CREATE USER 'zabbix'@'%' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON zabbix.* TO 'zabbix'@'%';
FLUSH PRIVILEGES;
```

Import schema:

```bash
# Download schema
wget https://cdn.zabbix.com/zabbix/sources/stable/7.4/zabbix-7.4.0.tar.gz
tar -xzf zabbix-7.4.0.tar.gz

# Import
mysql -h mysql.example.com -u zabbix -p zabbix < zabbix-7.4.0/database/mysql/schema.sql
mysql -h mysql.example.com -u zabbix -p zabbix < zabbix-7.4.0/database/mysql/images.sql
mysql -h mysql.example.com -u zabbix -p zabbix < zabbix-7.4.0/database/mysql/data.sql
```

---

## High Availability Setup

### Multiple Zabbix Servers

Create `docker-compose.ha.yaml`:

```yaml
version: '3.8'

services:
  zabbix-server-1:
    extends:
      file: docker-compose_v3_alpine_mysql_latest.yaml
      service: zabbix-server
    environment:
      - ZBX_NODEADDRESS=zabbix-server-1
      - ZBX_HANODENAME=zabbix-ha-node-1

  zabbix-server-2:
    extends:
      file: docker-compose_v3_alpine_mysql_latest.yaml
      service: zabbix-server
    environment:
      - ZBX_NODEADDRESS=zabbix-server-2
      - ZBX_HANODENAME=zabbix-ha-node-2
```

Start:

```bash
docker compose -f docker-compose.ha.yaml up -d
```

---

## Custom Configuration

### Override Zabbix Server Settings

Create `zabbix_server.conf`:

```ini
# Custom settings
CacheSize=128M
HistoryCacheSize=64M
TrendCacheSize=32M
ValueCacheSize=256M
```

Mount in `docker-compose.override.yaml`:

```yaml
version: '3.8'

services:
  zabbix-server:
    volumes:
      - ./zabbix_server.conf:/etc/zabbix/zabbix_server.conf:ro
```

Restart:

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

---

## Monitoring Docker Containers

### Using Zabbix Agent 2

Start with Docker monitoring:

```yaml
version: '3.8'

services:
  zabbix-agent2:
    image: zabbix/zabbix-agent2:alpine-latest
    environment:
      - ZBX_SERVER_HOST=zabbix-server
      - ZBX_HOSTNAME=Docker Host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    privileged: true
```

Add template in Zabbix UI:

- Configuration → Hosts → Your host
- Templates → Add: `Docker by Zabbix agent 2`

View metrics:

- Monitoring → Latest data
- Filter: Docker

---

## Environment-Specific Examples

### Development Environment

```yaml
version: '3.8'

services:
  zabbix-server:
    environment:
      - ZBX_DEBUGLEVEL=4  # Verbose logging
      - ZBX_TIMEOUT=30

  zabbix-web-nginx-mysql:
    environment:
      - PHP_TZ=America/New_York
      - ZBX_SERVER_NAME=Dev Environment
```

### Production Environment

```yaml
version: '3.8'

services:
  zabbix-server:
    environment:
      - ZBX_CACHESIZE=256M
      - ZBX_HISTORYCACHESIZE=128M
      - ZBX_TRENDCACHESIZE=64M
      - ZBX_VALUECACHESIZE=512M
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G

  mysql-server:
    environment:
      - MYSQL_INNODB_BUFFER_POOL_SIZE=1G
    deploy:
      resources:
        limits:
          memory: 2G
```

---

## More Examples

For more examples, see:
- [Official Documentation](https://www.zabbix.com/documentation/current/manual/installation/containers)
- [GitHub Wiki](https://github.com/zabbix/zabbix-docker/wiki)
- [Community Examples](https://www.zabbix.com/forum/)

