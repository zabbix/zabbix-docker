# Database permissions

This example uses [`env_vars/mysql_init`](../../env_vars/mysql_init/) or [`env_vars/postgresql_init`](../../env_vars/postgresql_init/) to give Zabbix server and the frontend separate database accounts. The server account creates and upgrades the schema. The frontend account can read and write Zabbix data.

Run the commands from the repository root. The standard data directory for the chosen variant (`${DATA_DIRECTORY}/var/lib/mysql` or `${DATA_DIRECTORY}/var/lib/postgresql`) must be empty for the database image to run the initialization scripts.

## MySQL

```console
docker compose -p zabbix-db-permissions \
    -f compose.yaml -f examples/db_permissions/compose.mysql.yaml \
    up -d --wait zabbix-server zabbix-web-nginx-mysql
```

The server connects as `zabbix_server` through role `zbx_srv`. The frontend connects as `zabbix_web` through role `zbx_web`, which has no schema modification privileges.

## PostgreSQL

```console
docker compose -p zabbix-db-permissions \
    -f compose_pgsql.yaml -f examples/db_permissions/compose.postgresql.yaml \
    up -d --wait zabbix-server zabbix-web-nginx-pgsql
```

The server connects as `zabbix`, the owner of the Zabbix database and schema. The frontend connects as `zabbix_web` and gets access to tables and sequences through the default privileges set for objects created by `zabbix`.

Open the frontend to check access through the restricted account. To stop the example, use the same Compose files with `down`.
