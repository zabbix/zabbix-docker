![logo](http://www.zabbix.com/ru/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is the image?

The image is used to receive SNMP traps, store them to a log file and provide access to Zabbix to collected SNMP trap messsages.

# Zabbix snmptraps images

These are the only official Zabbix snmptraps Docker images. They are based on trusty Ubuntu. The available versions of Zabbix snmptraps are:

    Zabbix server 3.0 (tags: alpine-3.0-latest, ubuntu-3.0-latest)
    Zabbix server 3.0.* (tags: alpine-3.0.*, ubuntu-3.0.*)
    Zabbix server 3.2 (tags: alpine-3.2-latest, ubuntu-3.2-latest)
    Zabbix server 3.2.* (tags: alpine-3.2.*, ubuntu-3.2.*)
    Zabbix server 3.4 (tags: alpine-3.4-latest, ubuntu-3.4-latest, alpine-latest, ubuntu-latest, latest)
    Zabbix server 3.4.* (tags: alpine-3.4.*, ubuntu-3.4.*)
    Zabbix server 4.0 (tags: alpine-trunk, ubuntu-trunk)

Images are updated when new releases are published.

# How to use this image

## Start `zabbix-snmptraps`

Start a Zabbix snmptraps container as follows:

    docker run --name some-zabbix-snmptraps -d zabbix/zabbix-snmptraps:tag

Where `some-zabbix-snmptraps` is the name you want to assign to your container and `tag` is the tag specifying the version you want. See the list above for relevant tags, or look at the [full list of tags](https://hub.docker.com/r/zabbix/zabbix-snmptraps/tags/).

## Linking Zabbix server or Zabbix proxy with the container

    docker run --name some-zabbix-server --link some-zabbix-snmptraps:zabbix-snmptraps --volumes-from some-zabbix-snmptraps -d zabbix/zabbix-server:tag

## Container shell access and viewing Zabbix snmptraps logs

The `docker exec` command allows you to run commands inside a Docker container. The following command line will give you a bash shell inside your `zabbix-snmptraps` container:

```console
$ docker exec -ti some-zabbix-snmptraps /bin/bash
```

The Zabbix snmptraps log is available through Docker's container log:

```console
$ docker logs  some-zabbix-snmptraps
```

## Allowed volumes for the Zabbix snmptraps container

### ``/var/lib/zabbix/snmptraps``

The volume contains log file ``snmptraps.log`` named with received SNMP traps.

### ``/var/lib/zabbix/mibs``

The volume allows to add new MIB files. It does not support subdirectories, all MIBs must be placed to ``/var/lib/zabbix/mibs``.

# Supported Docker versions

This image is officially supported on Docker version 1.12.0.

Support for older versions (down to 1.6) is provided on a best-effort basis.

Please see [the Docker installation documentation](https://docs.docker.com/installation/) for details on how to upgrade your Docker daemon.

# User Feedback

## Documentation

Documentation for this image is stored in the [`snmptraps/` directory](https://github.com/zabbix/zabbix-docker/tree/3.0/snmptraps) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/master/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.
