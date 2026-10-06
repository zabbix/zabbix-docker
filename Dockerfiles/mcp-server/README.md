![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix MCP server?

Zabbix MCP server gives AI agents access to Zabbix over the [Model Context Protocol](https://modelcontextprotocol.io/) (MCP). An agent can browse hosts, triggers, problems and other monitoring data, and act on incidents: acknowledge, suppress or close problems and manage maintenance periods.

The server keeps no credentials of its own. Every request carries the Zabbix API token of the user, so the agent sees and changes only what that user may, and its actions appear in the Zabbix audit log under that user.

For the list of tools, security recommendations and other details, see the [Zabbix MCP server documentation](https://git.zabbix.com/projects/ZT/repos/mcp-server/browse).

# Zabbix MCP server images

These are the only official Zabbix MCP server Docker images. They are based on Alpine Linux v3.24, Ubuntu 26.04 (resolute), CentOS Stream 10 and Oracle Linux 10 images. The available versions of Zabbix MCP server are:

    Zabbix MCP server 8.0 (tags: alpine-trunk, ubuntu-trunk, ol-trunk)

Images are updated when new releases are published. The image with ``latest`` tag is based on Alpine Linux.

# How to use this image

## Start `zabbix-mcp-server`

Start a Zabbix MCP server container as follows:

```console
$ docker run --name some-zabbix-mcp-server -e ZBX_FRONTEND_URL="https://zabbix.example.com/api_jsonrpc.php" -e ZBX_ALLOWEDIP="192.0.2.0/24" -p 8443:8443 -d zabbix/zabbix-mcp-server:tag
```

Where `some-zabbix-mcp-server` is the name you want to assign to your container, `zabbix.example.com` is the Zabbix frontend, `192.0.2.0/24` is the network from which MCP clients may connect and `tag` is the tag specifying the version you want. See the list above for relevant tags, or look at the [full list of tags](https://hub.docker.com/r/zabbix/zabbix-mcp-server/tags/).

## Connect an MCP client

Configure the client to use the MCP Streamable HTTP transport and connect to:

    http://host:8443/mcp

Every request must include the following header:

    Authorization: Bearer <Zabbix API token>

The token must belong to the user on whose behalf the agent operates. The server is stateless and can run behind a load balancer without session affinity.

The endpoint can be checked with a single request:

```console
$ curl -sS http://host:8443/mcp \
    -H "Authorization: Bearer $ZABBIX_API_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json, text/event-stream" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```
or
```console
$ npx @modelcontextprotocol/inspector \
    --cli http://host:8443/mcp \
    --transport http \
    --header "Authorization: Bearer $ZABBIX_API_TOKEN" \
    --method tools/list
```

A successful response lists the tools exposed to the client.

## Container shell access and viewing Zabbix MCP server logs

The `docker exec` command allows you to run commands inside a Docker container. The following command line will give you a shell inside your `zabbix-mcp-server` container:

```console
$ docker exec -ti some-zabbix-mcp-server /bin/sh
```

The Zabbix MCP server log is available through Docker's container log:

```console
$ docker logs some-zabbix-mcp-server
```

## Environment Variables

When you start the `zabbix-mcp-server` image, you can adjust the configuration of Zabbix MCP server by passing one or more environment variables on the `docker run` command line.

### `ZBX_FRONTEND_URL`

Zabbix API endpoint URL ending with `api_jsonrpc.php`, for example `https://zabbix.example.com/api_jsonrpc.php`. This variable is required.

### `ZBX_ALLOWEDIP`

Comma-delimited list of IP addresses, CIDR networks or DNS names from which MCP client connections are accepted. By default, value is `127.0.0.1,::1`.

### `ZBX_LISTENIP`

IP address the MCP server listens on. By default, value is `0.0.0.0`.

### `ZBX_LISTENPORT`

Port the MCP server listens on. Allowed values are from `1024` to `32767`. By default, value is `8443`. The MCP endpoint is available at `/mcp`.

### `ZBX_DEBUGLEVEL`

The variable is used to specify debug level. By default, value is ``3``. Allowed values are listed below:

- ``0`` - basic information about starting and stopping of Zabbix processes
- ``1`` - critical information
- ``2`` - error information
- ``3`` - warnings
- ``4`` - for debugging (produces lots of information)
- ``5`` - extended debugging (produces even more information)

### `ZBX_TIMEOUT`

Full timeout for Zabbix API requests, in seconds. Allowed values are from `3` to `30`. By default, value is `10`.

### `ZBX_IGNOREURLCERTERRORS`

Disables TLS certificate validation for requests to the Zabbix API. By default, value is `false`. This should only be enabled in a development environment.

### `ZBX_MAX_RESPONSE_SIZE`

Maximum size of the text of a read tool result, not the entire HTTP/MCP message. Allowed values are from `16K` to `256K`. By default, value is `64K`.

### `ZBX_TOOL_PREFIX`

Prefix prepended to every exposed MCP tool name. By default, no prefix is used. Tool access rules always match the original tool names without this prefix.

### Tool access variables

The image supports an ordered sequence of tool access rules through the following indexed variables:

```
ZBX_ALLOWTOOL_<0-N>=
ZBX_DENYTOOL_<0-N>=
ZBX_ALLOWTOOL_REGEXP_<0-N>=
ZBX_DENYTOOL_REGEXP_<0-N>=
```

The indices form one sequence shared by all four variable types. They must start at `0`, must not contain gaps and each index must be used by exactly one variable. Rules are evaluated in order and the first matching rule decides. Wildcard rules support `*`; regular expression rules use RE2 syntax.

For example, expose only read tools used for incident triage:

```console
$ docker run --name some-zabbix-mcp-server \
    -e ZBX_FRONTEND_URL="https://zabbix.example.com/api_jsonrpc.php" \
    -e ZBX_ALLOWEDIP="192.0.2.0/24" \
    -e ZBX_ALLOWTOOL_REGEXP_0='^(host|hostgroup|trigger|problem)_get$' \
    -e ZBX_DENYTOOL_1='*' \
    -p 8443:8443 \
    -d zabbix/zabbix-mcp-server:tag
```

By default, the image denies `item_get`, `lld_get`, `interface_get`, `macro_get` and `history_get`. Indexed environment rules replace this default policy entirely. Removing these variables restores the default policy on the next container start.

Write tools (`problem_*` and `maintenance_*`) are enabled by default; if the MCP client does not support form elicitation, changes are applied without user approval.

### TLS variables

Additionally the image allows the following TLS variables:

```
ZBX_TLSACCEPT=unencrypted
ZBX_TLSCERTFILE=
ZBX_TLSCERT=
ZBX_TLSKEYFILE=
ZBX_TLSKEY=
```

Set `ZBX_TLSACCEPT=cert` and specify the server certificate and private key using `ZBX_TLSCERTFILE` and `ZBX_TLSKEYFILE`. Relative file names are resolved under `/var/lib/zabbix/enc`. Alternatively, the certificate and key contents may be passed directly through `ZBX_TLSCERT` and `ZBX_TLSKEY`.

## Allowed volumes for the Zabbix MCP server container

### ``/var/lib/zabbix/enc``

The volume is used to store the TLS certificate and private key specified with `ZBX_TLSCERTFILE` and `ZBX_TLSKEYFILE`. Mount the volume read-only.

# The image variants

The `zabbix-mcp-server` images come in many flavors, each designed for a specific use case.

## `zabbix-mcp-server:alpine-<version>`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in the [`alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images and thus leads to much slimmer images in general.

To minimize image size, it is uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the tools you need in your own Dockerfile.

## `zabbix-mcp-server:ubuntu-<version>`

This is the defacto image. If you are unsure about what your needs are, you probably want to use this one.

## `zabbix-mcp-server:ol-<version>`

Oracle Linux is an open-source operating system available under the GNU General Public License (GPLv2). It is suitable for general purpose or Oracle workloads.

# User Feedback

## Documentation

Documentation for this image is stored in the [`mcp-server/` directory](https://github.com/zabbix/zabbix-docker/tree/trunk/Dockerfiles/mcp-server) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/trunk/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

## Contributing

You are invited to contribute new features, fixes or updates, large or small; we are always thrilled to receive pull requests and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design and help you find out if someone else is working on the same thing.

## License

Zabbix MCP server is released under the GNU Affero General Public License version 3 (AGPLv3).
You can modify it and propagate such a modified version under the terms of the AGPLv3 as published by the Free Software Foundation.
For additional details, including answers to common questions about the AGPLv3, see the generic FAQ from the [Free Software Foundation](http://www.fsf.org/licenses/gpl-faq.html).

Zabbix is Open Source Software; however, if you use Zabbix in a commercial context, we kindly ask you to support the development of Zabbix by purchasing some level of technical support.
