![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)


[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/zabbix/zabbix-docker/badge)](https://securityscorecards.dev/viewer/?uri=github.com/zabbix/zabbix-docker)
<a href="https://bestpractices.coreinfrastructure.org/projects/8395" style="display: inline;"><img src="https://bestpractices.coreinfrastructure.org/projects/8395/badge" style="display: inline;"></a>
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=zabbix_zabbix-docker&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=zabbix_zabbix-docker)

[![Build images (DockerHub)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build.yml/badge.svg?branch=7.0&event=push)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build.yml)
[![Build images (DockerHub, Windows)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build_windows.yml/badge.svg?branch=7.0&event=push)](https://github.com/zabbix/zabbix-docker/actions/workflows/images_build_windows.yml)

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
* [rhel](https://catalog.redhat.com/software/container-stacks/detail/663b05ecf9ba071d2ff93618)

### Usage

There is some documentation and examples in the [official Zabbix Documentation](https://www.zabbix.com/documentation/current/manual/installation/containers)!

Please also follow usage instructions of each Zabbix component image:

* [zabbix-appliance](https://hub.docker.com/r/zabbix/zabbix-appliance/) - Zabbix appliance with built-in MySQL server, Zabbix server, Zabbix Java Gateway and Zabbix frontend based on Nginx web-server

* [zabbix-agent](https://hub.docker.com/r/zabbix/zabbix-agent/) - Zabbix agent
* [zabbix-agent2](https://hub.docker.com/r/zabbix/zabbix-agent2/) - Zabbix agent 2
* [zabbix-server-mysql](https://hub.docker.com/r/zabbix/zabbix-server-mysql/) - Zabbix server with MySQL database support
* [zabbix-server-pgsql](https://hub.docker.com/r/zabbix/zabbix-server-pgsql/) - Zabbix server with PostgreSQL database support
* [zabbix-web-apache-mysql](https://hub.docker.com/r/zabbix/zabbix-web-apache-mysql/) - Zabbix web interface on Apache2 web server with MySQL database support
* [zabbix-web-apache-pgsql](https://hub.docker.com/r/zabbix/zabbix-web-apache-pgsql/) - Zabbix web interface on Apache2 web server with PostgreSQL database support
* [zabbix-web-nginx-mysql](https://hub.docker.com/r/zabbix/zabbix-web-nginx-mysql/) - Zabbix web interface on Nginx web server with MySQL database support
* [zabbix-web-nginx-pgsql](https://hub.docker.com/r/zabbix/zabbix-web-nginx-pgsql/) - Zabbix web interface on Nginx web server with PostgreSQL database support
* [zabbix-proxy-sqlite3](https://hub.docker.com/r/zabbix/zabbix-proxy-sqlite3/) - Zabbix proxy with SQLite3 database support
* [zabbix-proxy-mysql](https://hub.docker.com/r/zabbix/zabbix-proxy-mysql/) - Zabbix proxy with MySQL database support
* [zabbix-java-gateway](https://hub.docker.com/r/zabbix/zabbix-java-gateway/) - Zabbix Java Gateway
* [zabbix-web-service](https://hub.docker.com/r/zabbix/zabbix-web-service/) - Zabbix web service for performing various tasks using headless web browser (for example, reporting)
* [zabbix-snmptraps](https://hub.docker.com/r/zabbix/zabbix-snmptraps/) - Additional container image for Zabbix server and Zabbix proxy to support SNMP traps

### Docker Compose

There is provided Docker Compose files for each set of base Operating System and Database engine.

Templates support several [Compose  profiles](https://docs.docker.com/compose/profiles/). Minimal set of services is brought up by default, to start additional components e.g. Zabbix Agent use profile 'full' or 'all'. Additionally, it is possible to start only required components.

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
