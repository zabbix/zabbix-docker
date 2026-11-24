# Zabbix MCP server

Start the local stack from the repository root:

```console
docker compose -p zabbix-mcp \
    -f compose.yaml -f examples/mcp/compose.yaml \
    up -d --wait \
    zabbix-server zabbix-web-nginx-mysql zabbix-mcp-server
```

Create an expiring [Zabbix API token](https://www.zabbix.com/documentation/8.0/en/manual/web_interface/frontend_sections/users/api_tokens) for a user with API access limited to `host.get` and **Read** permission on the **Zabbix servers** host group.

Replace `<token>` with your API token to call `host_get` over MCP Streamable HTTP.

```console
curl -fsS http://127.0.0.1:9443/mcp \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"host_get","arguments":{"limit":5}}}'
```

The result includes the built-in **Zabbix server** host. The MCP port is published only on localhost for this HTTP example.

Stop the stack:

```console
docker compose -p zabbix-mcp -f compose.yaml -f examples/mcp/compose.yaml down
```
