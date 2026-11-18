# Mixed history providers

This example stores `uint`, `dbl`, and `str` history in ClickHouse and `log`, `text`, and `json` history in Elasticsearch.

Run from the repository root:

```console
./examples/history_providers/run.sh
```

Use mutual TLS between Zabbix server and ClickHouse:

```console
./examples/history_providers/run.sh --tls
```

Elasticsearch remains on HTTP because its history provider does not support TLS certificate settings.
