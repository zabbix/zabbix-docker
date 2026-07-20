# Web TLS example

Generate a self-signed certificate for the web server being tested:

```console
./tools/web_tls/generate_nginx.sh
./tools/web_tls/generate_apache.sh
```

Each script writes only the files expected by its web container:

```text
zbx_env/etc/ssl/nginx/
zbx_env/etc/ssl/apache2/
```

Start the selected frontend normally. HTTPS is enabled automatically when its
certificate files are present:

```console
docker compose up -d zabbix-web-nginx-mysql
docker compose --profile all up -d zabbix-web-apache-mysql
```

The default endpoints are `https://localhost/` and
`https://localhost:8443/`. The certificate is intended only for local tests.
