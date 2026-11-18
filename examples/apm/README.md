# Zabbix APM example

Run the example from the repository root:

```console
./examples/apm/run.sh
```

Open `http://localhost:8090/checkout?order=1` and inspect `apm-demo-checkout` and `apm-demo-inventory` in Zabbix. The example includes successful requests and inventory errors so both normal and failed operations are visible.

## ClickHouse cluster

Run APM against a ClickHouse cluster with two shards and two replicas per shard:

```console
./examples/apm/run.sh --cluster
```
