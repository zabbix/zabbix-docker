# Zabbix server HA

Start two Zabbix server nodes with a shared MySQL database and one frontend:

```console
docker compose -p zabbix-ha \
    -f compose.yaml -f examples/zabbix_ha/compose.yaml \
    up -d --wait \
    zabbix-server-1 zabbix-server-2 zabbix-web-nginx-mysql
```

To observe failover, stop the active server node, for example:

```console
docker compose -p zabbix-ha \
    -f compose.yaml -f examples/zabbix_ha/compose.yaml \
    stop zabbix-server-1
```

Remove the example stack:

```console
docker compose -p zabbix-ha down --remove-orphans
```
