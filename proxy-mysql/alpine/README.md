![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix proxy?

Zabbix proxy is a process that may collect monitoring data from one or more monitored devices and send the information to the Zabbix server, essentially working on behalf of the server. All collected data is buffered locally and then transferred to the Zabbix server the proxy belongs to.

# Zabbix proxy images

These are the only official Zabbix proxy Docker images. They are based on Alpine Linux v3.9, Ubuntu 18.04 (bionic) and CentOS 7 images. The available versions of Zabbix proxy are:

    Zabbix proxy 3.0 (tags: alpine-3.0-latest, ubuntu-3.0-latest, centos-3.0-latest)
    Zabbix proxy 3.0.* (tags: alpine-3.0.*, ubuntu-3.0.*, centos-3.0.*)
    Zabbix proxy 3.2 (tags: alpine-3.2-latest, ubuntu-3.2-latest, centos-3.2-latest)
    Zabbix proxy 3.2.* (tags: alpine-3.2.*, ubuntu-3.2.*, centos-3.2.*)
    Zabbix proxy 3.4 (tags: alpine-3.4-latest, ubuntu-3.4-latest, centos-3.4-latest)
    Zabbix proxy 3.4.* (tags: alpine-3.4.*, ubuntu-3.4.*, centos-3.4.*)
    Zabbix proxy 4.0 (tags: alpine-4.0-latest, ubuntu-4.0-latest, centos-4.0-latest)
    Zabbix proxy 4.0.* (tags: alpine-4.0.*, ubuntu-4.0.*, centos-4.0.*)
    Zabbix proxy 4.2 (tags: alpine-4.2-latest, ubuntu-4.2-latest, centos-4.2-latest)
    Zabbix proxy 4.2.* (tags: alpine-4.2.*, ubuntu-4.2.*, centos-4.2.*)
    Zabbix proxy 4.4 (tags: alpine-4.4-latest, ubuntu-4.4-latest, centos-4.4-latest, alpine-latest, ubuntu-latest, centos-latest, latest)
    Zabbix proxy 4.4.* (tags: alpine-4.4.*, ubuntu-4.4.*, centos-4.4.*)
    Zabbix proxy 5.0 (tags: alpine-trunk, ubuntu-trunk, centos-trunk)

Images are updated when new releases are published. The image with ``latest`` tag is based on Alpine Linux.

The image uses MySQL database to store collected data before sending it to Zabbix server. It uses the next procedure to start:
- Checking database availability
- If ``MYSQL_ROOT_PASSWORD`` or ``MYSQL_ALLOW_EMPTY_PASSWORD`` are specified, the instance tries to create ``MYSQL_USER`` user with ``MYSQL_PASSWORD`` to use these credentials then for Zabbix server.
- Checking of having `MYSQL_DATABASE` database. Creating `MYSQL_DATABASE` database name if it does not exist
- Checking of having `dbversion` table. Creating Zabbix proxy database schema if no `dbversion` table

# How to use this image

## Start `zabbix-proxy-mysql`

Start a Zabbix proxy container as follows:

    docker run --name some-zabbix-proxy-mysql -e DB_SERVER_HOST="some-mysql-server" -e MYSQL_USER="some-user" -e MYSQL_PASSWORD="some-password" -e ZBX_HOSTNAME=some-hostname -e ZBX_SERVER_HOST=some-zabbix-server -d zabbix/zabbix-proxy-mysql:tag

Where `some-zabbix-proxy-mysql` is the name you want to assign to your container, `some-mysql-server` is IP or DNS name of MySQL server, `some-user` is user to connect to Zabbix database on MySQL server, `some-password` is the password to connect to MySQL server, `some-hostname` is the hostname, it is Hostname parameter in Zabbix proxy configuration file, `some-zabbix-server` is IP or DNS name of Zabbix server and `tag` is the tag specifying the version you want. See the list above for relevant tags, or look at the [full list of tags](https://hub.docker.com/r/zabbix/zabbix-proxy-mysql/tags/).

## Connects from Zabbix server (Passive proxy)

This image exposes the standard Zabbix proxy port (10051) and can operate as Passive proxy in case `ZBX_PROXYMODE` = `1`. Start Zabbix server container like this in order to link it to the Zabbix proxy container:

```console
$ docker run --name some-zabbix-server --link some-zabbix-proxy-mysql:zabbix-proxy-mysql -d zabbix/zabbix-server:latest
```

## Connect to Zabbix server (Active proxy)

This image can operate as Active proxy (`default` mode). Start your application container like this in order to link Zabbix proxy to Zabbix server containters:

```console
$ docker run --name some-zabbix-proxy-mysql --link some-zabbix-server:zabbix-server -d zabbix/zabbix-proxy-mysql:latest
```

## Container shell access and viewing Zabbix proxy logs

The `docker exec` command allows you to run commands inside a Docker container. The following command line will give you a bash shell inside your `zabbix-proxy-mysql` container:

```console
$ docker exec -ti some-zabbix-proxy-mysql /bin/bash
```

The Zabbix proxy log is available through Docker's container log:

```console
$ docker logs some-zabbix-proxy-mysql
```

## Environment Variables

When you start the `zabbix-proxy-mysql` image, you can adjust the configuration of the Zabbix proxy by passing one or more environment variables on the `docker run` command line.

### `ZBX_PROXYMODE`

The variable allows to switch Zabbix proxy mode. Bu default, value is `0` - active proxy. Allowed values are `0` - active proxy and `1` - passive proxy.

### `ZBX_HOSTNAME`

This variable is unique, case sensitive hostname. By default, value is `zabbix-proxy-mysql` of the container. It is ``Hostname`` parameter in ``zabbix_proxy.conf``.

### `ZBX_SERVER_HOST`

This variable is IP or DNS name of Zabbix server or Zabbix proxy. By default, value is `zabbix-server`. It is ``Server`` parameter in ``zabbix_proxy.conf``. It is allowed to specify Zabbix server or Zabbix proxy port number using ``ZBX_SERVER_PORT`` variable. It make sense in case of non-default port for active checks.

### `ZBX_SERVER_PORT`

This variable is port Zabbix server listening on. By default, value is `10051`.

### `DB_SERVER_HOST`

This variable is IP or DNS name of MySQL server. By default, value is 'mysql-server'

### `DB_SERVER_PORT`
    
This variable is port of MySQL server. By default, value is '3306'.

### `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_USER_FILE`, `MYSQL_PASSWORD_FILE`

These variables are used by Zabbix proxy to connect to Zabbix database. With the `_FILE` variables you can instead provide the path to a file which contains the user / the password instead. Without Docker Swarm or Kubernetes you also have to map the files. Those are exclusive so you can just provide one type - either `MYSQL_USER` or `MYSQL_USER_FILE`!

```console
docker run --name some-zabbix-proxy-mysql -e DB_SERVER_HOST="some-mysql-server" -v ./.MYSQL_USER:/run/secrets/MYSQL_USER -e MYSQL_USER_FILE=/run/secrets/MYSQL_USER -v ./.MYSQL_PASSWORD:/run/secrets/MYSQL_PASSWORD -e MYSQL_PASSWORD_FILE=/var/run/secrets/MYSQL_PASSWORD -e ZBX_HOSTNAME=some-hostname -e ZBX_SERVER_HOST=some-zabbix-server -d zabbix/zabbix-proxy-mysql:tag
```

With Docker Swarm or Kubernetes this works with secrets. That way it is replicated in your cluster!

```console
printf "zabbix" | docker secret create MYSQL_USER -
printf "zabbix" | docker secret create MYSQL_PASSWORD -
docker run --name some-zabbix-proxy-mysql -e DB_SERVER_HOST="some-mysql-server" -e MYSQL_USER_FILE=/run/secrets/MYSQL_USER -e MYSQL_PASSWORD_FILE=/run/secrets/MYSQL_PASSWORD -e ZBX_SERVER_HOST="some-zabbix-server" -e ZBX_HOSTNAME=some-hostname -e ZBX_SERVER_HOST=some-zabbix-server -d zabbix/zabbix-proxy-mysql:tag
```

This method is also applicable for `MYSQL_ROOT_PASSWORD` with `MYSQL_ROOT_PASSWORD_FILE`.

By default, values for `MYSQL_USER` and `MYSQL_PASSWORD` are `zabbix`, `zabbix`.

### `MYSQL_DATABASE`

The variable is Zabbix database name. By default, value is `zabbix_proxy`.

### `ZBX_LOADMODULE`

The variable is list of comma separated loadable Zabbix modules. It works with  volume ``/var/lib/zabbix/modules``. The syntax of the variable is ``dummy1.so,dummy2.so``.

### ``ZBX_DEBUGLEVEL``

The variable is used to specify debug level. By default, value is ``3``. It is ``DebugLevel`` parameter in ``zabbix_server.conf``. Allowed values are listed below:
- ``0`` - basic information about starting and stopping of Zabbix processes;
- ``1`` - critical information
- ``2`` - error information
- ``3`` - warnings
- ``4`` - for debugging (produces lots of information)
- ``5`` - extended debugging (produces even more information)

### `ZBX_TIMEOUT`

The variable is used to specify timeout for processing checks. By default, value is ``4``.

### `ZBX_JAVAGATEWAY_ENABLE`

The variable enable communication with Zabbix Java Gateway to collect Java related checks. By default, value is `false`.

### Other variables

Additionally the image allows to specify many other environment variables listed below:

```
ZBX_ENABLEREMOTECOMMANDS=0 # Available since 3.4.0
ZBX_LOGREMOTECOMMANDS=0 # Available since 3.4.0
ZBX_HOSTNAMEITEM=system.hostname
ZBX_SOURCEIP=
ZBX_PROXYLOCALBUFFER=0
ZBX_PROXYOFFLINEBUFFER=1
ZBX_PROXYHEARTBEATFREQUENCY=60
ZBX_CONFIGFREQUENCY=3600
ZBX_DATASENDERFREQUENCY=1
ZBX_STARTPOLLERS=5
ZBX_IPMIPOLLERS=0
ZBX_STARTPOLLERSUNREACHABLE=1
ZBX_STARTTRAPPERS=5
ZBX_STARTPINGERS=1
ZBX_STARTDISCOVERERS=1
ZBX_STARTHTTPPOLLERS=1
ZBX_JAVAGATEWAY=zabbix-java-gateway
ZBX_JAVAGATEWAYPORT=10052
ZBX_STARTJAVAPOLLERS=0
ZBX_STARTVMWARECOLLECTORS=0
ZBX_VMWAREFREQUENCY=60
ZBX_VMWAREPERFFREQUENCY=60
ZBX_VMWARECACHESIZE=8M
ZBX_VMWARETIMEOUT=10
ZBX_ENABLE_SNMP_TRAPS=false
ZBX_LISTENIP=
ZBX_HOUSEKEEPINGFREQUENCY=1
ZBX_CACHESIZE=8M
ZBX_STARTDBSYNCERS=4
ZBX_HISTORYCACHESIZE=16M
ZBX_HISTORYINDEXCACHESIZE=4M
ZBX_TRAPPERIMEOUT=300
ZBX_UNREACHABLEPERIOD=45
ZBX_UNAVAILABLEDELAY=60
ZBX_UNREACHABLEDELAY=15
ZBX_LOGSLOWQUERIES=3000
ZBX_TLSCONNECT=unencrypted
ZBX_TLSACCEPT=unencrypted
ZBX_TLSCAFILE=
ZBX_TLSCRLFILE=
ZBX_TLSSERVERCERTISSUER=
ZBX_TLSSERVERCERTSUBJECT=
ZBX_TLSCERTFILE=
ZBX_TLSKEYFILE=
ZBX_TLSPSKIDENTITY=
ZBX_TLSPSKFILE=
```

Default values of these variables are specified after equal sign.

The allowed variables are identical of parameters in official ``zabbix_proxy.conf``. For example, ``ZBX_LOGSLOWQUERIES`` = ``LogSlowQueries``.

Please use official documentation for [``zabbix_proxy.conf``](https://www.zabbix.com/documentation/current/manual/appendix/config/zabbix_proxy) to get more information about the variables.

## Allowed volumes for the Zabbix proxy container

### ``/usr/lib/zabbix/externalscripts``

The volume is used by External checks (type of items). It is `ExternalScripts` parameter in ``zabbix_proxy.conf``.

### ``/var/lib/zabbix/modules``

The volume allows load additional modules and extend Zabbix proxy using ``LoadModule`` feature.

### ``/var/lib/zabbix/enc``

The volume is used to store TLS related files. These file names are specified using ``ZBX_TLSCAFILE``, ``ZBX_TLSCRLFILE``, ``ZBX_TLSKEY_FILE`` and ``ZBX_TLSPSKFILE`` variables.

### ``/var/lib/zabbix/ssh_keys``

The volume is used as location of public and private keys for SSH checks and actions. It is `SSHKeyLocation` parameter in ``zabbix_proxy.conf``.

### ``/var/lib/zabbix/ssl/certs``

The volume is used as location of of SSL client certificate files for client authentication. It is `SSLCertLocation` parameter in ``zabbix_proxy.conf``.

### ``/var/lib/zabbix/ssl/keys``

The volume is used as location of SSL private key files for client authentication. It is `SSLKeyLocation` parameter in ``zabbix_proxy.conf``.

### ``/var/lib/zabbix/ssl/ssl_ca``

The volume is used as location of certificate authority (CA) files for SSL server certificate verification. It is `SSLCALocation` parameter in ``zabbix_proxy.conf``.

### ``/var/lib/zabbix/snmptraps``

The volume is used as location of ``snmptraps.log`` file. It could be shared by ``zabbix-snmptraps`` container and inherited using `volumes_from` Docker option while creating new instance of Zabbix proxy.
SNMP traps processing feature could be enabled using shared volume and switched ``ZBX_ENABLE_SNMP_TRAPS`` environment variable to `true`.

### ``/var/lib/zabbix/mibs``

The volume allows to add new MIB files. It does not support subdirectories, all MIBs must be placed to ``/var/lib/zabbix/mibs``.

# The image variants

The `zabbix-proxy-mysql` images come in many flavors, each designed for a specific use case.

## `zabbix-proxy-mysql:ubuntu-<version>`

This is the defacto image. If you are unsure about what your needs are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code and start the container to start your app), as well as the base to build other images off of.

## `zabbix-proxy-mysql:alpine-<version>`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in [the `alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images (~5MB), and thus leads to much slimmer images in general.

This variant is highly recommended when final image size being as small as possible is desired. The main caveat to note is that it does use [musl libc](http://www.musl-libc.org) instead of [glibc and friends](http://www.etalabs.net/compare_libcs.html), so certain software might run into issues depending on the depth of their libc requirements. However, most software doesn't have an issue with this, so this variant is usually a very safe choice. See [this Hacker News comment thread](https://news.ycombinator.com/item?id=10782897) for more discussion of the issues that might arise and some pro/con comparisons of using Alpine-based images.

To minimize image size, it's uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the things you need in your own Dockerfile (see the [`alpine` image description](https://hub.docker.com/_/alpine/) for examples of how to install packages if you are unfamiliar).

# Supported Docker versions

This image is officially supported on Docker version 1.12.0.

Support for older versions (down to 1.6) is provided on a best-effort basis.

Please see [the Docker installation documentation](https://docs.docker.com/installation/) for details on how to upgrade your Docker daemon.

# User Feedback

## Documentation

Documentation for this image is stored in the [`proxy-mysql/` directory](https://github.com/zabbix/zabbix-docker/tree/3.0/proxy-mysql) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/master/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

Zabbix proxy does not support Jabber notifications on Alpine Linux because of `iksemel` package is in testing repository and not available in stable repository.

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.
