![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix server?

Zabbix server is the central process of Zabbix software.

The server performs the polling and trapping of data, it calculates triggers, sends notifications to users. It is the central component to which Zabbix agents and proxies report data on availability and integrity of systems. The server can itself remotely check networked services (such as web servers and mail servers) using simple service checks.

# Zabbix server images

Images are updated when new releases are published.

The image uses MySQL database. It uses the next procedure to start:
- Checking database availability
- If ``MYSQL_ROOT_PASSWORD`` or ``MYSQL_ALLOW_EMPTY_PASSWORD`` are specified, the instance tries to create ``MYSQL_USER`` user with ``MYSQL_PASSWORD`` to use these credentials then for Zabbix server.
- Checking of having `MYSQL_DATABASE` database. Creating `MYSQL_DATABASE` database name if it does not exist
- Checking of having `dbversion` table. Creating Zabbix server database schema and upload initial data sample if no `dbversion` table

# How to use this image

## Start `zabbix-server-mysql`

Start a Zabbix server container as follows:

    podman run --name some-zabbix-server-mysql -e DB_SERVER_HOST="some-mysql-server" -e MYSQL_USER="some-user" -e MYSQL_PASSWORD="some-password" --init -d zabbix/zabbix-server-mysql-trunk:tag

Where `some-zabbix-server-mysql` is the name you want to assign to your container, `some-mysql-server` is IP or DNS name of MySQL server, `some-user` is user to connect to Zabbix database on MySQL server, `some-password` is the password to connect to MySQL server and `tag` is the tag specifying the version you want.

> [!NOTE]
> Zabbix server has possibility to execute `fping` utility to perform ICMP checks. When containers are running in rootless mode or with specific restrictions environment, you may face errors related to fping:
> `fping: Operation not permitted`
> or
> lost all packets to all resources
> in this case add `--cap-add=net_raw` to `docker run` or `podman run` commands.
> Additionally fping executing in non-root environments can require sysctl modification:
> `net.ipv4.ping_group_range=0 1995`
> where 1995 is `zabbix` GID.

## Container shell access and viewing Zabbix server logs

The `podman exec` command allows you to run commands inside a Podman container. The following command line will give you a bash shell inside your `zabbix-server-mysql` container:

```console
$ podman exec -ti some-zabbix-server-mysql /bin/bash
```

The Zabbix server log is available through Podman's container log:

```console
$ podman logs some-zabbix-server-mysql
```

## Environment Variables

When you start the `zabbix-server-mysql` image, you can adjust the configuration of the Zabbix server by passing one or more environment variables on the `podman run` command line.

### `DB_SERVER_HOST`

This variable is IP or DNS name of MySQL server. By default, value is 'mysql-server'

### `DB_SERVER_PORT`

This variable is port of MySQL server. By default, value is '3306'.

### `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_USER_FILE`, `MYSQL_PASSWORD_FILE`

These variables are used by Zabbix server to connect to Zabbix database. With the `_FILE` variables you can instead provide the path to a file which contains the user / the password instead. Without Docker Swarm or Kubernetes you also have to map the files. Those are exclusive so you can just provide one type - either `MYSQL_USER` or `MYSQL_USER_FILE`!

```console
podman run --name some-zabbix-server-mysql -e DB_SERVER_HOST="some-mysql-server" -v ./.MYSQL_USER:/run/secrets/MYSQL_USER -e MYSQL_USER_FILE=/run/secrets/MYSQL_USER -v ./.MYSQL_PASSWORD:/run/secrets/MYSQL_PASSWORD -e MYSQL_PASSWORD_FILE=/var/run/secrets/MYSQL_PASSWORD --init -d zabbix/zabbix-server-mysql:tag
```

With Docker Swarm or Kubernetes this works with secrets. That way it is replicated in your cluster!

```console
printf "zabbix" | docker secret create MYSQL_USER -
printf "zabbix" | docker secret create MYSQL_PASSWORD -
podman run --name some-zabbix-server-mysql -e DB_SERVER_HOST="some-mysql-server" -e MYSQL_USER_FILE=/run/secrets/MYSQL_USER -e MYSQL_PASSWORD_FILE=/run/secrets/MYSQL_PASSWORD --init -d zabbix/zabbix-server-mysql:tag
```

This method is also applicable for `MYSQL_ROOT_PASSWORD` with `MYSQL_ROOT_PASSWORD_FILE`.

By default, values for `MYSQL_USER` and `MYSQL_PASSWORD` are `zabbix`, `zabbix`.

### `MYSQL_DATABASE`

The variable is Zabbix database name. By default, value is `zabbix`.

### `ZBX_LOADMODULE`

The variable is list of comma separated loadable Zabbix modules. It works with  volume ``/var/lib/zabbix/modules``. The syntax of the variable is ``dummy1.so,dummy2.so``.

### `ZBX_DEBUGLEVEL`

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
ZBX_ALLOWUNSUPPORTEDDBVERSIONS=0 # Available since 6.0.0
ZBX_DBTLSCONNECT= # Available since 5.0.0
ZBX_DBTLSCAFILE= # Available since 5.0.0
ZBX_DBTLSCERTFILE= # Available since 5.0.0
ZBX_DBTLSKEYFILE= # Available since 5.0.0
ZBX_DBTLSCIPHER= # Available since 5.0.0
ZBX_DBTLSCIPHER13= # Available since 5.0.0
ZBX_VAULTDBPATH= # Available since 5.2.0
ZBX_VAULTURL=https://127.0.0.1:8200 # Available since 5.2.0
VAULT_TOKEN= # Available since 5.2.0
ZBX_LISTENIP=
ZBX_LISTENPORT=10051
ZBX_LISTENBACKLOG=
ZBX_STARTREPORTWRITERS=0 # Available since 5.4.0
ZBX_WEBSERVICEURL=http://zabbix-web-service:10053/report # Available since 5.4.0
ZBX_SERVICEMANAGERSYNCFREQUENCY=60 # Available since 6.0.0
ZBX_HISTORYSTORAGEURL= # Available since 3.4.0
ZBX_HISTORYSTORAGETYPES=uint,dbl,str,log,text # Available since 3.4.0
ZBX_STARTPOLLERS=5
ZBX_IPMIPOLLERS=0
ZBX_STARTPREPROCESSORS=3 # Available since 3.4.0
ZBX_STARTCONNECTORS=0 # Available since 6.4.0
ZBX_STARTPOLLERSUNREACHABLE=1
ZBX_STARTTRAPPERS=5
ZBX_STARTPINGERS=1
ZBX_STARTDISCOVERERS=1
ZBX_STARTHISTORYPOLLERS=5 # Available since 5.4.0
ZBX_STARTHTTPPOLLERS=1
ZBX_STARTODBCPOLLERS=1 # Available since 6.0.0
ZBX_STARTTIMERS=1
ZBX_STARTESCALATORS=1
ZBX_STARTALERTERS=3 # Available since 3.4.0
ZBX_JAVAGATEWAY=zabbix-java-gateway
ZBX_JAVAGATEWAYPORT=10052
ZBX_STARTJAVAPOLLERS=5
ZBX_STARTLLDPROCESSORS=2 # Available since 4.2.0
ZBX_STATSALLOWEDIP= # Available since 4.0.5
ZBX_STARTVMWARECOLLECTORS=0
ZBX_VMWAREFREQUENCY=60
ZBX_VMWAREPERFFREQUENCY=60
ZBX_VMWARECACHESIZE=8M
ZBX_VMWARETIMEOUT=10
ZBX_ENABLE_SNMP_TRAPS=false
ZBX_SOURCEIP=
ZBX_HOUSEKEEPINGFREQUENCY=1
ZBX_MAXHOUSEKEEPERDELETE=5000
ZBX_PROBLEMHOUSEKEEPINGFREQUENCY=60 # Available since 6.0.0
ZBX_SENDERFREQUENCY=30 # Depcrecated since 3.4.0
ZBX_CACHESIZE=8M
ZBX_CACHEUPDATEFREQUENCY=10
ZBX_STARTDBSYNCERS=4
ZBX_EXPORTFILESIZE=1G # Available since 4.0.0
ZBX_EXPORTTYPE= # Available since 5.0.10 and 5.2.6
ZBX_AUTOHANODENAME=fqdn # Allowed values: fqdn, hostname. Available since 6.0.0
ZBX_HANODENAME= # Available since 6.0.0
ZBX_AUTONODEADDRESS=fqdn # Allowed values: fqdn, hostname. Available since 6.0.0
ZBX_NODEADDRESSPORT=10051 # Allowed to use with ZBX_AUTONODEADDRESS variable only. Available since 6.0.0
ZBX_NODEADDRESS=localhost # Available since 6.0.0
ZBX_HISTORYCACHESIZE=16M
ZBX_HISTORYINDEXCACHESIZE=4M
ZBX_HISTORYSTORAGEDATEINDEX=0 # Available since 4.0.0
ZBX_TRENDCACHESIZE=4M
ZBX_TRENDFUNCTIONCACHESIZE=4M
ZBX_VALUECACHESIZE=8M
ZBX_TRAPPERTIMEOUT=300
ZBX_UNREACHABLEPERIOD=45
ZBX_UNAVAILABLEDELAY=60
ZBX_UNREACHABLEDELAY=15
ZBX_LOGSLOWQUERIES=3000
ZBX_STARTPROXYPOLLERS=1
ZBX_PROXYCONFIGFREQUENCY=10
ZBX_PROXYDATAFREQUENCY=1
ZBX_TLSCAFILE=
ZBX_TLSCRLFILE=
ZBX_TLSCERTFILE=
ZBX_TLSKEYFILE=
ZBX_TLSCIPHERALL= # Available since 4.4.7
ZBX_TLSCIPHERALL13= # Available since 4.4.7
ZBX_TLSCIPHERCERT= # Available since 4.4.7
ZBX_TLSCIPHERCERT13= # Available since 4.4.7
ZBX_TLSCIPHERPSK= # Available since 4.4.7
ZBX_TLSCIPHERPSK13= # Available since 4.4.7

```

Default values of these variables are specified after equal sign.

The allowed variables are identical of parameters in official ``zabbix_server.conf``. For example, ``ZBX_LOGSLOWQUERIES`` = ``LogSlowQueries``.

Please use official documentation for [``zabbix_server.conf``](https://www.zabbix.com/documentation/current/manual/appendix/config/zabbix_server) to get more information about the variables.

## Allowed volumes for the Zabbix server container

### ``/usr/lib/zabbix/alertscripts``

The volume is used for custom alert scripts. It is `AlertScriptsPath` parameter in ``zabbix_server.conf``.

### ``/usr/lib/zabbix/externalscripts``

The volume is used by External checks (type of items). It is `ExternalScripts` parameter in ``zabbix_server.conf``.

### ``/var/lib/zabbix/modules``

The volume allows load additional modules and extend Zabbix server using ``LoadModule`` feature.

### ``/var/lib/zabbix/enc``

The volume is used to store TLS related files. These file names are specified using ``ZBX_TLSCAFILE``, ``ZBX_TLSCRLFILE``, ``ZBX_TLSKEY_FILE`` and ``ZBX_TLSPSKFILE`` variables.

### ``/var/lib/zabbix/ssh_keys``

The volume is used as location of public and private keys for SSH checks and actions. It is `SSHKeyLocation` parameter in ``zabbix_server.conf``.

### ``/var/lib/zabbix/ssl/certs``

The volume is used as location of of SSL client certificate files for client authentication. It is `SSLCertLocation` parameter in ``zabbix_server.conf``.

### ``/var/lib/zabbix/ssl/keys``

The volume is used as location of SSL private key files for client authentication. It is `SSLKeyLocation` parameter in ``zabbix_server.conf``.

### ``/var/lib/zabbix/ssl/ssl_ca``

The volume is used as location of certificate authority (CA) files for SSL server certificate verification. It is `SSLCALocation` parameter in ``zabbix_server.conf``.

### ``/var/lib/zabbix/snmptraps``

The volume is used as location of ``snmptraps.log`` file. It could be shared by ``zabbix-snmptraps`` container and inherited using `volumes_from` Podman option while creating new instance of Zabbix server.
SNMP traps processing feature could be enabled using shared volume and switched ``ZBX_ENABLE_SNMP_TRAPS`` environment variable to `true`.

### ``/var/lib/zabbix/mibs``

The volume allows to add new MIB files. It does not support subdirectories, all MIBs must be placed to ``/var/lib/zabbix/mibs``.

### ``/var/lib/zabbix/export``

Directory for real-time export of events, history and trends in newline-delimited JSON format. Could be enabled using ``ZBX_EXPORTFILESIZE`` environment variable.

# User Feedback

## Documentation

Documentation for this image is stored in the [`server-mysql/` directory](https://github.com/zabbix/zabbix-docker/tree/trunk/Dockerfiles/server-mysql/rhel/) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/trunk/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.
