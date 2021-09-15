![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix web service?

Zabbix web service for performing various tasks using headless web browser (for example, reporting).

# Zabbix web service images

These are the only official Zabbix web service Docker images. They are based on Alpine Linux v3.13, Ubuntu 20.04 (focal) and Oracle Linux 8 images. The available versions of Zabbix web service are:

    Zabbix web service 5.4 (tags: alpine-5.4-latest, ubuntu-5.4-latest, ol-5.4-latest, alpine-latest, ubuntu-latest, ol-latest, latest)
    Zabbix web service 5.4.* (tags: alpine-5.4.*, ubuntu-5.4.*, ol-5.4.*)
    Zabbix web service 6.0 (tags: alpine-trunk, ubuntu-trunk, ol-trunk)

Images are updated when new releases are published. The image with ``latest`` tag is based on Alpine Linux.

# How to use this image

## Start `zabbix-web-service`

Start a Zabbix web service container as follows:

    docker run --name some-zabbix-web-service -e ZBX_ALLOWEDIP="some-zabbix-server" --cap-add=SYS_ADMIN -d zabbix/zabbix-web-service:tag

Where `some-zabbix-web-service` is the name you want to assign to your container, `some-zabbix-server` is IP or DNS name of Zabbix server and `tag` is the tag specifying the version you want. See the list above for relevant tags, or look at the [full list of tags](https://hub.docker.com/r/zabbix/zabbix-web-service/tags/).

## Connects from Zabbix server in other containers

This image exposes the standard Zabbix web service port (``10053``) to perform communication, so container linking makes Zabbix web service instance available to Zabbix server containers. Start your application container like this in order to link it to the Zabbix web service container:

```console
$ docker run --name some-zabbix-server --link some-zabbix-web-service:zabbix-web-service -e ZBX_STARTREPORTWRITERS="2" -e ZBX_WEBSERVICEURL="http://some-zabbix-web-service:10053/report" -d zabbix/zabbix-server:latest
```

## Container shell access and viewing Zabbix web service logs

The `docker exec` command allows you to run commands inside a Docker container. The following command line will give you a bash shell inside your `zabbix-web-service` container:

```console
$ docker exec -ti some-zabbix-web-service /bin/bash
```

The Zabbix web service log is available through Docker's container log:

```console
$ docker logs some-zabbix-web-service
```

## Environment Variables

When you start the `zabbix-web-service` image, you can adjust the configuration of the Zabbix web service by passing one or more environment variables on the `docker run` command line.

### `ZBX_ALLOWEDIP`

This variable is IP or DNS name or list of IP / DNS names of Zabbix server. By default, value is `zabbix-server`.

### `ZBX_LISTENPORT`

Listen port for incoming request. By default, value is `10053`.

### `ZBX_DEBUGLEVEL`

The variable is used to specify debug level. By default, value is ``3``. It is ``DebugLevel`` parameter in ``zabbix_web_service.conf``. Allowed values are listed below:
- ``0`` - basic information about starting and stopping of Zabbix processes;
- ``1`` - critical information
- ``2`` - error information
- ``3`` - warnings
- ``4`` - for debugging (produces lots of information)
- ``5`` - extended debugging (produces even more information)

### `ZBX_TIMEOUT`

The variable is used to specify timeout for processing requests. By default, value is ``3``.

### Other variables

Additionally the image allows to specify many other environment variables listed below:

```
ZBX_TLSACCEPT=unencrypted
ZBX_TLSCAFILE=
ZBX_TLSCERTFILE=
ZBX_TLSKEYFILE=
```

Default values of these variables are specified after equal sign.

Please use official documentation for [``zabbix_web_service.conf``](https://www.zabbix.com/documentation/current/manual/appendix/config/zabbix_web_service) to get more information about the variables.

## Allowed volumes for the Zabbix web service container

### ``/var/lib/zabbix/enc``
    
The volume is used to store TLS related files. These file names are specified using ``ZBX_TLSCAFILE``, ``ZBX_TLSCERTFILE`` and ``ZBX_TLSKEY_FILE`` variables.

# The image variants

The `zabbix-web-service` images come in many flavors, each designed for a specific use case.

## `zabbix-web-service:alpine-<version>`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in [the `alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images (~5MB), and thus leads to much slimmer images in general.

This variant is highly recommended when final image size being as small as possible is desired. The main caveat to note is that it does use [musl libc](http://www.musl-libc.org) instead of [glibc and friends](http://www.etalabs.net/compare_libcs.html), so certain software might run into issues depending on the depth of their libc requirements. However, most software doesn't have an issue with this, so this variant is usually a very safe choice. See [this Hacker News comment thread](https://news.ycombinator.com/item?id=10782897) for more discussion of the issues that might arise and some pro/con comparisons of using Alpine-based images.

To minimize image size, it's uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the things you need in your own Dockerfile (see the [`alpine` image description](https://hub.docker.com/_/alpine/) for examples of how to install packages if you are unfamiliar).

## `zabbix-web-service:ubuntu-<version>`

This is the defacto image. If you are unsure about what your needs are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code and start the container to start your app), as well as the base to build other images off of.

## `zabbix-web-service:ol-<version>`

Oracle Linux is an open-source operating system available under the GNU General Public License (GPLv2). Suitable for general purpose or Oracle workloads, it benefits from rigorous testing of more than 128,000 hours per day with real-world workloads and includes unique innovations such as Ksplice for zero-downtime kernel patching, DTrace for real-time diagnostics, the powerful Btrfs file system, and more.

# Supported Docker versions

This image is officially supported on Docker version 1.12.0.

Support for older versions (down to 1.6) is provided on a best-effort basis.

Please see [the Docker installation documentation](https://docs.docker.com/installation/) for details on how to upgrade your Docker daemon.

# User Feedback

## Documentation

Documentation for this image is stored in the [`web-service/` directory](https://github.com/zabbix/zabbix-docker/tree/5.4/web-service) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/master/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

Zabbix web services uses Google Chromium with headless mode. Because of restrictions you may see the following error during report generation:
```
Failed to move to new namespace: PID namespaces supported, Network namespace supported, but failed: errno = Operation not permitted
```

To avoid the issue it is required to add ``SYS_ADMIN`` capability for Zabbix web service. The capability is redundant and allow too much.

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.
