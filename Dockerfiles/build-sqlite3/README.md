![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix build base (SQLite3)?

Zabbix build base (SQLite3) image is used for building Zabbix components. It contains compiled Zabbix components.

# Zabbix build base (SQLite3) images

These are the only official Zabbix build base (SQLite3) Docker images. They are based on Alpine Linux v3.24, Ubuntu 26.04 (resolute), CentOS Stream 10 and Oracle Linux 10 images. The available versions of the image are:

    Zabbix build base 6.0 (tags: alpine-6.0-latest, ubuntu-6.0-latest, ol-6.0-latest)
    Zabbix build base 6.0.* (tags: alpine-6.0.*, ubuntu-6.0.*, ol-6.0.*)
    Zabbix build base 7.0 (tags: alpine-7.0-latest, ubuntu-7.0-latest, ol-7.0-latest)
    Zabbix build base 7.0.* (tags: alpine-7.0.*, ubuntu-7.0.*, ol-7.0.*)
    Zabbix build base 7.4 (tags: alpine-7.4-latest, ubuntu-7.4-latest, ol-7.4-latest)
    Zabbix build base 7.4.* (tags: alpine-7.4.*, ubuntu-7.4.*, ol-7.4.*)
    Zabbix build base 8.0 (tags: alpine-8.0-latest, ubuntu-8.0-latest, ol-8.0-latest, alpine-latest, ubuntu-latest, ol-latest, latest)
    Zabbix build base 8.0.* (tags: alpine-8.0.*, ubuntu-8.0.*, ol-8.0.*)
    Zabbix build base 8.2 (tags: alpine-trunk, ubuntu-trunk, ol-trunk)

The image with ``latest`` tag is based on Alpine Linux.

# How to use this image

The image is used to build / compile Zabbix components. Components are prepared for usage in any other images.

The image uses [Zabbix build base](https://github.com/zabbix/zabbix-docker/tree/8.0/Dockerfiles/build-base) image with prepared build environment as base image and build / compile Zabbix components only.

It contains limited prepared Zabbix components while MySQL and PostgreSQL build base images contain all possible components:
* zabbix-agent
* zabbix-agent2
* zabbix-proxy-sqlite3
* zabbix-java-gateway

# User Feedback

## Documentation

Documentation for this image is stored in the [`build-sqlite3/` directory](https://github.com/zabbix/zabbix-docker/tree/8.0/Dockerfiles/build-sqlite3) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/8.0/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

### Known issues

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.

## License

Starting from Zabbix version 7.0, all subsequent Zabbix versions will be released under the GNU Affero General Public License version 3 (AGPLv3).
You can modify the relevant version and propagate such modified version under the terms of the AGPLv3 as published by the Free Software Foundation.
For additional details, including answers to common questions about the AGPLv3, see the generic FAQ from the [Free Software Foundation](http://www.fsf.org/licenses/gpl-faq.html).

Zabbix is Open Source Software, however, if you use Zabbix in a commercial context we kindly ask you to support the development of Zabbix by purchasing some level of technical support.
All previous Zabbix software versions up to 6.4 are released under the GNU General Public License version 2 (GPLv2). The formal terms of the GPLv2 and AGPLv3 can be found at http://www.fsf.org/licenses/.
