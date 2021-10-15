![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix build base (MySQL)?

Zabbix build base (MySQL) image is used for building Zabbix components. It contains compiled Zabbix components.

# Zabbix build base (MySQL) images

These are the only official Zabbix build base (MySQL) Docker images. They are based on Alpine Linux v3.11, v3.14, Ubuntu 20.04 (focal), CentOS 8 and Oracle Linux 8 images. The available versions of the image are:

    Zabbix build base 4.0 (tags: alpine-4.0-latest, ubuntu-4.0-latest, centos-4.0-latest)
    Zabbix build base 4.0.* (tags: alpine-4.0.*, ubuntu-4.0.*, centos-4.0.*)
    Zabbix build base 5.0 (tags: alpine-5.0-latest, ubuntu-5.0-latest, ol-5.0-latest)
    Zabbix build base 5.0.* (tags: alpine-5.0.*, ubuntu-5.0.*, ol-5.0.*)
    Zabbix build base 5.4 (tags: alpine-5.4-latest, ubuntu-5.4-latest, ol-5.4-latest, alpine-latest, ubuntu-latest, ol-latest, latest)
    Zabbix build base 5.4.* (tags: alpine-5.4.*, ubuntu-5.4.*, ol-5.4.*)
    Zabbix build base 6.0 (tags: alpine-trunk, ubuntu-trunk, ol-trunk)

Images are updated when new releases are published. The image with ``latest`` tag is based on Alpine Linux.

# How to use this image

The image is used to build / compile Zabbix components. Components are prepared for usage in any other images.

The image uses [Zabbix build base](https://github.com/zabbix/zabbix-docker/tree/trunk/Dockerfiles/build-base) image with prepared build environment as base image and build / compile Zabbix components only.

# The image variants

The `zabbix-build-mysql` images come in many flavors, each designed for a specific use case.

## `zabbix-build-mysql:alpine-<version>`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in [the `alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images (~5MB), and thus leads to much slimmer images in general.

This variant is highly recommended when final image size being as small as possible is desired. The main caveat to note is that it does use [musl libc](http://www.musl-libc.org) instead of [glibc and friends](http://www.etalabs.net/compare_libcs.html), so certain software might run into issues depending on the depth of their libc requirements. However, most software doesn't have an issue with this, so this variant is usually a very safe choice. See [this Hacker News comment thread](https://news.ycombinator.com/item?id=10782897) for more discussion of the issues that might arise and some pro/con comparisons of using Alpine-based images.

To minimize image size, it's uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the things you need in your own Dockerfile (see the [`alpine` image description](https://hub.docker.com/_/alpine/) for examples of how to install packages if you are unfamiliar).

## `zabbix-build-mysql:ubuntu-<version>`

This is the defacto image. If you are unsure about what your needs are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code and start the container to start your app), as well as the base to build other images off of.

## `zabbix-build-mysql:ol-<version>`

Oracle Linux is an open-source operating system available under the GNU General Public License (GPLv2). Suitable for general purpose or Oracle workloads, it benefits from rigorous testing of more than 128,000 hours per day with real-world workloads and includes unique innovations such as Ksplice for zero-downtime kernel patching, DTrace for real-time diagnostics, the powerful Btrfs file system, and more.

# Supported Docker versions

This image is officially supported on Docker version 1.12.0.

Support for older versions (down to 1.6) is provided on a best-effort basis.

Please see [the Docker installation documentation](https://docs.docker.com/installation/) for details on how to upgrade your Docker daemon.

# User Feedback

## Documentation

Documentation for this image is stored in the [`build-base/` directory](https://github.com/zabbix/zabbix-docker/tree/trunk/Dockerfiles/build-base) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/master/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.
