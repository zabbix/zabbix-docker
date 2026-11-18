# Zabbix TLS and PSK

Run an active proxy and an agent with certificate-based TLS for active and passive checks, and mutual TLS between the frontend and Zabbix server:

```console
./examples/zabbix_tls/run.sh cert
```

Use pre-shared keys for the proxy and agent connections. The frontend connects to Zabbix server without TLS in this mode:

```console
./examples/zabbix_tls/run.sh psk
```

The example registers the proxy and monitors the agent through it using the built-in `Zabbix server` host.
