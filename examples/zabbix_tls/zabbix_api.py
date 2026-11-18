#!/usr/bin/env python3

import json
import os
import urllib.request
from pathlib import Path

API_URL = "http://zabbix-web:8080/api_jsonrpc.php"
API_USERNAME = "Admin"
API_PASSWORD = "zabbix"
MODE = os.environ["EXAMPLE_TLS_MODE"]
TLS_MODES = {"psk": 2, "cert": 4}
TLS_CA_ISSUER = "CN=Zabbix TLS example CA"
PROXY_NAME = "zabbix-proxy-tls"
HOST_NAME = "Zabbix server"


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


def encryption_settings(component):
    if MODE == "cert":
        return {
            "tls_issuer": TLS_CA_ISSUER,
            "tls_subject": f"CN=zabbix-{component}",
        }

    return {
        "tls_issuer": "",
        "tls_subject": "",
        "tls_psk_identity": f"zabbix-{component}-tls",
        "tls_psk": Path(f"/run/zabbix/secrets/tls/{component}.psk")
        .read_text(encoding="ascii")
        .strip(),
    }


def configure_proxy(auth):
    proxies = api_request(
        "proxy.get",
        {"output": ["proxyid"], "filter": {"name": PROXY_NAME}},
        auth,
    )
    settings = {
        "name": PROXY_NAME,
        "operating_mode": 0,
        "tls_accept": TLS_MODES[MODE],
        **encryption_settings("proxy"),
    }
    if proxies:
        proxy_id = proxies[0]["proxyid"]
        api_request("proxy.update", {"proxyid": proxy_id, **settings}, auth)
        return proxy_id

    result = api_request("proxy.create", settings, auth)
    return result["proxyids"][0]


def configure_host(auth, proxy_id):
    host = api_request(
        "host.get",
        {
            "output": ["hostid"],
            "selectInterfaces": ["interfaceid"],
            "filter": {"host": HOST_NAME},
        },
        auth,
    )[0]
    host_id = host["hostid"]
    api_request(
        "host.update",
        {
            "hostid": host_id,
            "monitored_by": 1,
            "proxyid": proxy_id,
            "tls_connect": TLS_MODES[MODE],
            "tls_accept": TLS_MODES[MODE],
            "interfaces": [
                {
                    "interfaceid": host["interfaces"][0]["interfaceid"],
                    "type": 1,
                    "main": 1,
                    "useip": 0,
                    "ip": "",
                    "dns": "zabbix-agent",
                    "port": "10050",
                }
            ],
            **encryption_settings("agent"),
        },
        auth,
    )
    return host_id


def main():
    auth = api_request(
        "user.login",
        {"username": API_USERNAME, "password": API_PASSWORD},
    )
    proxy_id = configure_proxy(auth)
    host_id = configure_host(auth, proxy_id)
    print(f"Configured {MODE} encryption for proxy {proxy_id} and host {host_id}")


if __name__ == "__main__":
    main()
