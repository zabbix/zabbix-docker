# Quick Start Guide

This guide will help you get Zabbix running with Docker Compose.

## Prerequisites

- Docker Engine 20.10.0 or later
- Docker Compose 2.0.0 or later

## Installation

Clone the repository and start the containers:

```bash
git clone https://github.com/zabbix/zabbix-docker.git
cd zabbix-docker
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

This compose file starts the essential Zabbix components:
- Zabbix Server (monitoring engine)
- Zabbix Web Interface (Nginx)
- MySQL Database

> **Note:** Other compose files in this repository may include additional services like Proxy, Elasticsearch, or Selenium. This file provides a minimal working setup.

The initial startup takes approximately 1-2 minutes while the database is initialized.

## Accessing the Web Interface

Once the containers are running, access the web interface at:

```
http://localhost
```

Default credentials:
- Username: `Admin`
- Password: `zabbix`

> **Important:** Change the default password after your first login. Go to User menu (top right) → User settings → Change password.

## Verifying the Installation

Check container status:

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml ps
```

View server logs:

```bash
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml logs -f zabbix-server
```

## Troubleshooting

### Port 80 already in use

If you encounter a "port is already allocated" error:

```bash
export ZBX_WEB_SERVER_PORT=8080
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d
```

Access the web interface at `http://localhost:8080`.

For additional troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

---

## Next Steps

After logging in, you can:

1. Add your first host: Configuration → Hosts → Create host
2. Explore pre-configured templates for Linux, MySQL, Apache, Docker, and more
3. Set up alert notifications: Administration → Media types

## Common Commands

```bash
# Stop Zabbix
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml down

# Restart
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml restart

# Update to latest
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml pull
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml up -d

# Remove everything (including data!)
docker compose -f docker-compose_v3_alpine_mysql_latest.yaml down -v
```

### Simplifying commands

To avoid typing the full compose file name, you can use an environment variable:

```bash
export COMPOSE_FILE=docker-compose_v3_alpine_mysql_latest.yaml
docker compose up -d
docker compose down
```

Alternatively, create a shell alias:

```bash
alias zabbix='docker compose -f docker-compose_v3_alpine_mysql_latest.yaml'
zabbix up -d
zabbix logs -f
zabbix down
```

## Alternative Configurations

| OS | MySQL | PostgreSQL |
|----|-------|------------|
| Alpine (recommended) | `docker-compose_v3_alpine_mysql_latest.yaml` | `docker-compose_v3_alpine_pgsql_latest.yaml` |
| Ubuntu | `docker-compose_v3_ubuntu_mysql_latest.yaml` | `docker-compose_v3_ubuntu_pgsql_latest.yaml` |
| CentOS Stream | `docker-compose_v3_centos_mysql_latest.yaml` | `docker-compose_v3_centos_pgsql_latest.yaml` |
| Oracle Linux | `docker-compose_v3_ol_mysql_latest.yaml` | `docker-compose_v3_ol_pgsql_latest.yaml` |

Example:
```bash
docker compose -f docker-compose_v3_ubuntu_pgsql_latest.yaml up -d
```

## Additional Resources

- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [Configuration Examples](EXAMPLES.md) - HTTPS, HA, external database, etc.
- [Official Documentation](https://www.zabbix.com/documentation/current/manual/installation/containers)
- [Docker Hub](https://hub.docker.com/u/zabbix/) - All available images
- [Zabbix Forum](https://www.zabbix.com/forum/) - Community support

