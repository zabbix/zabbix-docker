![Zabbix Logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)


[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/zabbix/zabbix-docker/badge)](https://securityscorecards.dev/viewer/?uri=github.com/zabbix/zabbix-docker)
<a href="https://bestpractices.coreinfrastructure.org/projects/8395" style="display: inline;"><img src="https://bestpractices.coreinfrastructure.org/projects/8395/badge" style="display: inline;"></a>
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=zabbix-docker&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=zabbix-docker)

[![Build images (DockerHub)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build.yml/badge.svg?branch=trunk&event=push)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build.yml)
[![Build images (DockerHub, Windows)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build_windows.yml/badge.svg?branch=trunk&event=push)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build_windows.yml)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com


## Zabbix Dockerfiles

This repository contains **Dockerfile** of [Zabbix](https://zabbix.com/) for [Docker](https://www.docker.com/)'s [automated build](https://registry.hub.docker.com/u/zabbix/) published to the public [Docker Hub Registry](https://registry.hub.docker.com/).

### Base Docker Image

* [alpine](https://hub.docker.com/_/alpine/)
* [centos](https://quay.io/repository/centos/centos?tab=info)
* [oracle linux](https://hub.docker.com/_/oraclelinux/)
* [ubuntu](https://hub.docker.com/_/ubuntu/)

### Compatibility notes

> [!IMPORTANT]
> Starting with Zabbix 8.0, `ZBX_SSO_SP_KEY`, `ZBX_SSO_SP_CERT` and `ZBX_SSO_IDP_CERT` contain inline key or certificate data instead of file paths. Use `ZBX_SSO_SP_KEYFILE`, `ZBX_SSO_SP_CERTFILE` and `ZBX_SSO_IDP_CERTFILE` to specify file paths. Files with the previously recognized fixed names (`sp.key`, `sp.crt`, `idp.crt`, `vault.crt` and `vault.key`) are no longer detected automatically; specify them explicitly using the corresponding `*FILE` variables.

> [!IMPORTANT]
> Starting with Zabbix 8.0, mounted secret files have a common layout under `/run/zabbix/secrets`: `tls`, `apm`, `database`, `vault`, `sso`, `ssl` and `providers`. Mount only the required subdirectories. The bundled Compose configuration maps them from the matching `${DATA_DIRECTORY}/run/zabbix/secrets` subdirectories and keeps the deprecated mounts as commented examples for temporary migration. An absolute `ZBX_*FILE` path is used unchanged; a relative path is resolved in its component-specific subdirectory. Where a previous relative-location contract existed, it remains supported with a warning. Database secrets and all `*_FILE` value secrets also support the standard `/run/secrets` directory, which takes precedence when the same relative name exists in both locations.

> A `ZBX_<NAME>` file-secret variable contains inline data, while `ZBX_<NAME>FILE` identifies a mounted file. For raw credentials, `<NAME>` contains the value and `<NAME>_FILE` identifies a file whose contents are read by the entrypoint. The two sources in each pair are mutually exclusive. A zero-length mounted file is accepted as an explicitly empty value.

> Depending on the configured secrets, the entrypoint may write generated files to `/run/zabbix/secrets_internal`. When an image is run with a read-only root filesystem, mount a writable tmpfs there, for example `--tmpfs /run/zabbix/secrets_internal:mode=770,uid=1997,gid=0`. The bundled Compose configuration already provides this mount.

ClickHouse TLS files can be supplied separately from provider JSON through `ZBX_HISTORYPROVIDER_<N>_SSL_{CA,CERT,KEY}` and `ZBX_TELEMETRYPROVIDER_0_SSL_{CA,CERT,KEY}`, or their corresponding `*FILE` variables. The older `ssl_cert_file` and `ssl_key_file` JSON fields are deprecated and cannot be combined with the new variables.

### Usage

There is some documentation and examples in the [official Zabbix Documentation](https://www.zabbix.com/documentation/current/manual/installation/containers)!

Please also follow usage instructions of each Zabbix component image:

* [zabbix-agent](https://hub.docker.com/r/zabbix/zabbix-agent/) - Zabbix agent
* [zabbix-agent2](https://hub.docker.com/r/zabbix/zabbix-agent2/) - Zabbix agent 2
* [zabbix-server-mysql](https://hub.docker.com/r/zabbix/zabbix-server-mysql/) - Zabbix server with MySQL or MariaDB database support
* [zabbix-server-pgsql](https://hub.docker.com/r/zabbix/zabbix-server-pgsql/) - Zabbix server with PostgreSQL database support
* [zabbix-web-apache-mysql](https://hub.docker.com/r/zabbix/zabbix-web-apache-mysql/) - Zabbix web interface on Apache2 web server with MySQL or MariaDB database support
* [zabbix-web-apache-pgsql](https://hub.docker.com/r/zabbix/zabbix-web-apache-pgsql/) - Zabbix web interface on Apache2 web server with PostgreSQL database support
* [zabbix-web-nginx-mysql](https://hub.docker.com/r/zabbix/zabbix-web-nginx-mysql/) - Zabbix web interface on Nginx web server with MySQL or MariaDB database support
* [zabbix-web-nginx-pgsql](https://hub.docker.com/r/zabbix/zabbix-web-nginx-pgsql/) - Zabbix web interface on Nginx web server with PostgreSQL database support
* [zabbix-proxy-sqlite3](https://hub.docker.com/r/zabbix/zabbix-proxy-sqlite3/) - Zabbix proxy with SQLite3 database support
* [zabbix-proxy-mysql](https://hub.docker.com/r/zabbix/zabbix-proxy-mysql/) - Zabbix proxy with MySQL or MariaDB database support
* [zabbix-java-gateway](https://hub.docker.com/r/zabbix/zabbix-java-gateway/) - Zabbix Java Gateway
* [zabbix-mcp-server](https://hub.docker.com/r/zabbix/zabbix-mcp-server/) - Zabbix MCP server connects AI agents to the Zabbix API through the Model Context Protocol (MCP)
* [zabbix-web-service](https://hub.docker.com/r/zabbix/zabbix-web-service/) - Zabbix web service for performing various tasks using headless web browser (for example, reporting)
* [zabbix-snmptraps](https://hub.docker.com/r/zabbix/zabbix-snmptraps/) - Additional container image for Zabbix server and Zabbix proxy to support SNMP traps

### Docker Compose

> [!NOTE]
> The following requires `make` and `docker compose` version 2.24.0 or greater.

You can use the Docker Compose files for MySQL or MariaDB, and PostgreSQL (**compose.yaml** and **compose_pgsql.yaml**, respectively) directly with the ``docker compose`` command. The required images and the Compose file behavior can be configured through environment variables and the .env file. By default, Alpine-based images and the latest version from the current branch are used.

In addition, a Makefile is provided to configure and run the required Docker Compose files, supporting a variety of base operating system and database engine combinations, as well as to build Zabbix components locally.

To get started clone this repository and then from the root directory of the repository run:
```
make help
```
To start with the default configuration run:
```
make up
```
Use variables to customize the configuration to suit your needs.  For example, the following selects Postgres as the underlying database, alpine as the base OS, and exposes the web interface on ports 8282 and 8443:
```
make up DB=pgsql OS=alpine ZABBIX_WEB_NGINX_HTTP_PORT=8282 ZABBIX_WEB_NGINX_HTTPS_PORT=8443
```
Then to clean up run:
```
make down DB=pgsql
```

Provided compose files support several [Compose  profiles](https://docs.docker.com/compose/profiles/). Minimal set of services is brought up by default, to start additional components e.g. Zabbix Agent use profile 'full' or 'all'. Additionally, it is possible to start only required components.

For local TLS testing, see the [web](tools/web_tls/) and [database](tools/db_tls/) examples.

## Issues and Wiki

Be sure to check [the Wiki-page](https://github.com/zabbix/zabbix-docker/wiki) on common problems and questions. If you still have problems with or questions about the images, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

> [!NOTE]
> Please report here issues and feature requests related to Docker images only. If you have issues or ideas how to improve Zabbix, use official [bug tracker](https://support.zabbix.com/).

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.

## License

Starting from Zabbix version 7.0, all subsequent Zabbix versions will be released under the GNU Affero General Public License version 3 (AGPLv3).
You can modify the relevant version and propagate such modified version under the terms of the AGPLv3 as published by the Free Software Foundation.

For additional details, including answers to common questions about the AGPLv3, see the generic FAQ from the [Free Software Foundation](http://www.fsf.org/licenses/gpl-faq.html).

Zabbix is Open Source Software, however, if you use Zabbix in a commercial context we kindly ask you to support the development of Zabbix by purchasing some level of technical support.
All previous Zabbix software versions up to 6.4 are released under the GNU General Public License version 2 (GPLv2). The formal terms of the GPLv2 and AGPLv3 can be found at http://www.fsf.org/licenses/.

Zabbix Bridge Adapter is proprietary software licensed under the Zabbix Mobile App End User License Agreement (EULA), not under GPLv2 or AGPLv3. The source code of Zabbix Bridge Adapter is not made available or distributed.

See the [current published EULA](https://www.zabbix.com/mobile_app_eula) and the [Zabbix Bridge Adapter image documentation](Dockerfiles/bridge-adapter/README.md#license) for license acceptance instructions.

Unless otherwise noted, files in this repository are licensed under the [GNU Affero General Public License version 3 (AGPLv3)](https://www.gnu.org/licenses/agpl-3.0.html). Third-party components remain subject to their respective licenses.
