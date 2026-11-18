#!/usr/bin/env python3

import json
import urllib.request


API_URL = "http://zabbix-web:8080/api_jsonrpc.php"
API_USERNAME = "Admin"
API_PASSWORD = "zabbix"
PROXY_NAME = "zabbix-proxy-apm"


def api_request(method, params, auth=None):
    payload = {
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": 1,
    }
    headers = {"Content-Type": "application/json-rpc"}
    if auth is not None:
        headers["Authorization"] = f"Bearer {auth}"

    request = urllib.request.Request(
        API_URL,
        data=json.dumps(payload).encode(),
        headers=headers,
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        result = json.load(response)

    if "error" in result:
        error = result["error"]
        raise RuntimeError(f'{error.get("message")}: {error.get("data")}')
    return result["result"]


def main():
    auth = api_request(
        "user.login",
        {"username": API_USERNAME, "password": API_PASSWORD},
    )
    proxies = api_request(
        "proxy.get",
        {"output": ["proxyid"], "filter": {"name": PROXY_NAME}},
        auth,
    )
    settings = {
        "name": PROXY_NAME,
        "operating_mode": 0,
        "apm": {
            "data_collection_status": 1,
        },
    }
    if proxies:
        proxy_id = proxies[0]["proxyid"]
        api_request("proxy.update", {"proxyid": proxy_id, **settings}, auth)
    else:
        proxy_id = api_request("proxy.create", settings, auth)["proxyids"][0]

    print(f"Configured APM for proxy {proxy_id}")


if __name__ == "__main__":
    main()
