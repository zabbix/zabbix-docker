# Database TLS examples

Generate a local CA plus server and client TLS certificates for database being tested:

```console
./tools/db_tls/generate_mysql.sh
./tools/db_tls/generate_postgresql.sh
```

The MySQL certificate is valid for `mysql-server`;
The PostgreSQL certificate is valid for `postgres-server`.

Both scripts write the filenames expected by Compose directly to `env_vars`, so run the one matching the selected database.

Enable verified database TLS with the matching Compose override:

```console
docker compose -f compose.yaml -f tools/db_tls/compose.mysql.yaml up -d
docker compose -f compose_pgsql.yaml -f tools/db_tls/compose.postgresql.yaml up -d
```

The overrides verify the database certificate and hostname. The generated client certificate is available for tests that additionally configure MySQL `REQUIRE X509` or PostgreSQL client-certificate authentication.

All generated files are ignored by Git and are intended only for local tests.
