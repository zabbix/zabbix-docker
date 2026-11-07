![logo](https://assets.zabbix.com/img/logo/zabbix_logo_500x131.png)

# What is Zabbix?

Zabbix is an enterprise-class open source distributed monitoring solution.

Zabbix is software that monitors numerous parameters of a network and the health and integrity of servers. Zabbix uses a flexible notification mechanism that allows users to configure e-mail based alerts for virtually any event. This allows a fast reaction to server problems. Zabbix offers excellent reporting and data visualisation features based on the stored data. This makes Zabbix ideal for capacity planning.

For more information and related downloads for Zabbix components, please visit https://hub.docker.com/u/zabbix/ and https://zabbix.com

# What is Zabbix Bridge Adapter?

Zabbix Bridge Adapter connects a Zabbix installation to Zabbix Bridge for use with the Zabbix mobile application.

# Zabbix Bridge Adapter images

Zabbix Bridge Adapter Docker images are based on the distroless `static` image (Debian 13). Images are updated when new releases are published.

# How to use this image

## Before starting `zabbix-bridge-adapter`

Zabbix server and the Zabbix frontend must be available. Obtain the Server ID from **Reports > System information** in the target Zabbix frontend.

The examples use an existing Docker network named `zabbix-backend`, shared with Zabbix server. The network must provide access to the Zabbix API and Zabbix Bridge.

## Reading and accepting the EULA

Read the bundled Zabbix Mobile App End User License Agreement (EULA) before accepting it:

```console
$ docker run --rm zabbix/zabbix-bridge-adapter:tag zabbix_bridge_adapter license show | less
```

Here and below, `tag` specifies the image version. Use the same image version for license acceptance and service startup.

If you agree to the terms, create `bridge-adapter.env` with the EULA version shown by the command and the values for your installation:

```text
ZBX_EULAVERSION=<reviewed EULA version>
ZBX_SERVERID=<Server ID from the Zabbix frontend>
ZBX_ZABBIXAPIURL=http://zabbix-web-nginx:8080/api_jsonrpc.php
ZBX_ZABBIXAPITLSCONNECT=disabled
```

For an HTTPS API URL, configure the TLS variables described below. Configure `ZBX_ALLOWEDAPIMETHODS_<N>` in the same file to allow the API methods required by the mobile application.

Accept the EULA using a persistent data volume:

```console
$ docker run --rm --network none --read-only \
    --env-file bridge-adapter.env \
    --volume zabbix-bridge-adapter-data:/var/lib/zabbix/bridge-adapter \
    zabbix/zabbix-bridge-adapter:tag accept-license
```

This command records acceptance of the specified EULA version and exits. Review and accept the terms again if the bundled EULA changes.

## Start `zabbix-bridge-adapter`

Start a Zabbix Bridge Adapter container using the same environment file and data volume:

```console
$ docker run --name some-zabbix-bridge-adapter \
    --network zabbix-backend \
    --read-only --tmpfs /tmp \
    --env-file bridge-adapter.env \
    --volume zabbix-bridge-adapter-data:/var/lib/zabbix/bridge-adapter \
    -d zabbix/zabbix-bridge-adapter:tag
```

Where `some-zabbix-bridge-adapter` is the name you want to assign to your container and `tag` specifies the image version.

## Connects from Zabbix server in other containers

Connect Zabbix server to the same network and set its `ZBX_BRIDGEADAPTERURL` variable to:

```text
http://some-zabbix-bridge-adapter:10054/rpc
```

Use an HTTPS URL when TLS is enabled on the Adapter listener.

## Container shell access and viewing Zabbix Bridge Adapter logs

The distroless image contains no shell. View the container log as follows:

```console
$ docker logs some-zabbix-bridge-adapter
```

## Environment Variables

When you start the `zabbix-bridge-adapter` image, you can adjust the configuration by passing environment variables on the `docker run` command line or using `--env-file`.

### `ZBX_EULAVERSION`

Exact version of the reviewed EULA to accept. The value must match the version displayed by `zabbix_bridge_adapter license show`. The variable is used only by the EULA acceptance step; setting it does not accept the EULA during normal Adapter startup.

### `ZBX_SERVERID`

Full Server ID displayed under **Reports > System information** in the target Zabbix frontend. This variable is required.

### `ZBX_ZABBIXAPIURL`

URL of the target Zabbix JSON-RPC API endpoint, for example `https://zabbix.example.com/api_jsonrpc.php`. This variable is required.

### `ZBX_ZABBIXBRIDGEURL`

URL used to retrieve the Zabbix Bridge configuration. By default, value is `https://config.mobilebridge.zabbix.com/.well-known/bridge-configuration`.

### `ZBX_ZABBIXAPITLSCONNECT`

TLS mode used for outbound connections to the Zabbix API. Supported values are:

- `disabled` - use plain HTTP;
- `required` - use HTTPS without certificate verification;
- `verify_ca` - use HTTPS and verify the certificate issuer;
- `verify_full` - use HTTPS and verify the certificate issuer and hostname.

Use `disabled` for an HTTP API URL. The default is `required` for HTTPS.

### `ZBX_ZABBIXAPITLSCAFILE`, `ZBX_ZABBIXAPITLSCA`

`ZBX_ZABBIXAPITLSCAFILE` specifies the CA file used to verify the Zabbix API certificate. A relative path is resolved against `/run/zabbix/secrets/tls`; the previous location remains available during migration with a warning. Alternatively, `ZBX_ZABBIXAPITLSCA` provides the CA certificate content directly. A CA is required when `ZBX_ZABBIXAPITLSCONNECT` is set to `verify_ca` or `verify_full`.

### `ZBX_LISTENIP`

IP address or hostname on which Bridge Adapter accepts requests from Zabbix server. By default, all available addresses are used.

### `ZBX_LISTENPORT`

Port on which Bridge Adapter accepts requests from Zabbix server. By default, value is `10054`. The RPC endpoint is available at `/rpc`.

### `ZBX_LISTENHEALTHPORT`

Port exposing the `/health` and `/ready` endpoints. By default, value is `8081`. The container health check uses `/ready`; the container becomes healthy when the Adapter is ready to handle requests.

### `ZBX_DEBUGLEVEL`

The variable is used to specify debug level. By default, value is ``3``. Allowed values are listed below:

- ``0`` - basic information about starting and stopping of Zabbix processes
- ``1`` - critical information
- ``2`` - error information
- ``3`` - warnings
- ``4`` - for debugging (produces lots of information)
- ``5`` - extended debugging (produces even more information)

### `ZBX_ALLOWEDAPIMETHODS_<N>`

Zabbix API methods that Bridge Adapter is allowed to call. Specify each method in a separate variable, using consecutive indexes starting with `0`. Each entry is either an exact method name, `resource.*` or `*.method`:

```text
ZBX_ALLOWEDAPIMETHODS_0=apiinfo.version
ZBX_ALLOWEDAPIMETHODS_1=host.get
```

If no variables are specified, the default deny-all policy applies and the Zabbix mobile application cannot use the Adapter. The minimal set of methods the mobile application currently needs is:

```text
ZBX_ALLOWEDAPIMETHODS_0=apiinfo.version
ZBX_ALLOWEDAPIMETHODS_1=device.onboard
ZBX_ALLOWEDAPIMETHODS_2=device.offboard
ZBX_ALLOWEDAPIMETHODS_3=user.get
ZBX_ALLOWEDAPIMETHODS_4=role.get
ZBX_ALLOWEDAPIMETHODS_5=problem.get
ZBX_ALLOWEDAPIMETHODS_6=event.get
ZBX_ALLOWEDAPIMETHODS_7=event.acknowledge
ZBX_ALLOWEDAPIMETHODS_8=trigger.get
ZBX_ALLOWEDAPIMETHODS_9=alert.get
ZBX_ALLOWEDAPIMETHODS_10=host.get
ZBX_ALLOWEDAPIMETHODS_11=hostgroup.get
```

### Listener TLS variables

The Adapter listener supports the following TLS variables:

```text
ZBX_TLSMODE=none
ZBX_TLSCAFILE=
ZBX_TLSCA=
ZBX_TLSCERTFILE=
ZBX_TLSCERT=
ZBX_TLSKEYFILE=
ZBX_TLSKEY=
ZBX_TLSALLOWEDCN=
```

## Allowed volumes for the Zabbix Bridge Adapter container

### ``/var/lib/zabbix/bridge-adapter``

The volume stores persistent Adapter data and EULA acceptance. Use the same writable volume for license acceptance and service startup. For a bind mount, create the directory beforehand and grant access to UID `1997`. Run both containers with the same UID.

### ``/run/zabbix/secrets/tls``

The volume provides TLS files referenced by the `ZBX_TLSCAFILE`, `ZBX_TLSCERTFILE`, `ZBX_TLSKEYFILE` and `ZBX_ZABBIXAPITLSCAFILE` variables. Mount it read-only.

# The image variants

## `zabbix-bridge-adapter:<version>`

This image is based on the [distroless `static` image](https://github.com/GoogleContainerTools/distroless). It has no shell or package manager. The Adapter runs as a non-root user (UID `1997`).

# User Feedback

## Documentation

Documentation for this image is stored in the [`bridge-adapter/` directory](https://github.com/zabbix/zabbix-docker/tree/trunk/Dockerfiles/bridge-adapter) of the [`zabbix/zabbix-docker` GitHub repo](https://github.com/zabbix/zabbix-docker/). Be sure to familiarize yourself with the [repository's `README.md` file](https://github.com/zabbix/zabbix-docker/blob/trunk/README.md) before attempting a pull request.

## Issues

If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues).

## Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull requests, and do our best to process them as fast as we can.

Before you start to code, we recommend discussing your plans through a [GitHub issue](https://github.com/zabbix/zabbix-docker/issues), especially for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you feedback on your design, and help you find out if someone else is working on the same thing.

## License

Zabbix Bridge Adapter is proprietary software licensed under the Zabbix Mobile App End User License Agreement (EULA), not under GPLv2 or AGPLv3. The source code of Zabbix Bridge Adapter is not made available or distributed.

The [current published EULA](https://www.zabbix.com/mobile_app_eula) may differ from the version bundled with your image. Review the bundled EULA at `/licenses/Zabbix-Mobile-App-EULA.txt`. Startup requires explicit acceptance of that version.

The image contains the following license files:

- `/licenses/Zabbix-Mobile-App-EULA.txt` - the EULA bundled with Zabbix Bridge Adapter;
- `/licenses/bridge-adapter-THIRD-PARTY-NOTICES.txt` - notices for the Adapter runtime dependencies;
- `/licenses/go-THIRD-PARTY-NOTICES.txt` - the Go runtime and standard library license and patent notice.

Unless otherwise noted, files in this repository are licensed under the [GNU Affero General Public License version 3 (AGPLv3)](https://www.gnu.org/licenses/agpl-3.0.html). Third-party components remain subject to their respective licenses.
