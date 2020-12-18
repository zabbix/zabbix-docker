![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is the image?

The image is used to receive SNMP traps, store them to a log file and provide access to Zabbix to collected SNMP trap messsages.

# Zabbix snmptraps images

These are the only official Zabbix snmptraps Docker images. They are based on Alpine Linux v3.11, Ubuntu 18.04 (bionic) and CentOS 7 images. The available versions of Zabbix snmptraps are:

    Zabbix snmptraps 3.0 (tags: alpine-3.0-latest, ubuntu-3.0-latest, centos-3.0-latest)
    Zabbix snmptraps 3.0.* (tags: alpine-3.0.*, ubuntu-3.0.*, centos-3.0.*)
    Zabbix snmptraps 3.2 (tags: alpine-3.2-latest, ubuntu-3.2-latest, centos-3.2-latest) (unsupported)
    Zabbix snmptraps 3.2.* (tags: alpine-3.2.*, ubuntu-3.2.*, centos-3.2.*) (unsupported)
    Zabbix snmptraps 3.4 (tags: alpine-3.4-latest, ubuntu-3.4-latest, centos-3.4-latest) (unsupported)
    Zabbix snmptraps 3.4.* (tags: alpine-3.4.*, ubuntu-3.4.*, centos-3.4.*) (unsupported)
    Zabbix snmptraps 4.0 (tags: alpine-4.0-latest, ubuntu-4.0-latest, centos-4.0-latest)
    Zabbix snmptraps 4.0.* (tags: alpine-4.0.*, ubuntu-4.0.*, centos-4.0.*)
    Zabbix snmptraps 4.2 (tags: alpine-4.2-latest, ubuntu-4.2-latest, centos-4.2-latest) (unsupported)
    Zabbix snmptraps 4.2.* (tags: alpine-4.2.*, ubuntu-4.2.*, centos-4.2.*) (unsupported)
    Zabbix snmptraps 4.4 (tags: alpine-4.4-latest, ubuntu-4.4-latest, centos-4.4-latest) (unsupported)
    Zabbix snmptraps 4.4.* (tags: alpine-4.4.*, ubuntu-4.4.*, centos-4.4.*) (unsupported)
    Zabbix snmptraps 5.0 (tags: alpine-5.0-latest, ubuntu-5.0-latest, centos-5.0-latest)
    Zabbix snmptraps 5.0.* (tags: alpine-5.0.*, ubuntu-5.0.*, centos-5.0.*)
    Zabbix snmptraps 5.2 (tags: alpine-5.2-latest, ubuntu-5.2-latest, centos-5.2-latest, alpine-latest, ubuntu-latest, centos-latest, latest)
    Zabbix snmptraps 5.2.* (tags: alpine-5.2.*, ubuntu-5.2.*, centos-5.2.*)
    Zabbix snmptraps 5.4 (tags: alpine-trunk, ubuntu-trunk, centos-trunk)

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

# The image variants

The `zabbix-snmptraps` images come in many flavors, each designed for a specific use case.

## `zabbix-snmptraps:ubuntu-<version>`

This is the defacto image. If you are unsure about what your needs are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code and start the container to start your app), as well as the base to build other images off of.

## `zabbix-snmptraps:centos-<version>`

This is the defacto image also. If you are unsure about what your needs are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code and start the container to start your app), as well as the base to build other images off of.

## `zabbix-snmptraps:alpine-<version>`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in [the `alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images (~5MB), and thus leads to much slimmer images in general.

This variant is highly recommended when final image size being as small as possible is desired. The main caveat to note is that it does use [musl libc](http://www.musl-libc.org) instead of [glibc and friends](http://www.etalabs.net/compare_libcs.html), so certain software might run into issues depending on the depth of their libc requirements. However, most software doesn't have an issue with this, so this variant is usually a very safe choice. See [this Hacker News comment thread](https://news.ycombinator.com/item?id=10782897) for more discussion of the issues that might arise and some pro/con comparisons of using Alpine-based images.

To minimize image size, it's uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the things you need in your own Dockerfile (see the [`alpine` image description](https://hub.docker.com/_/alpine/) for examples of how to install packages if you are unfamiliar).

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
