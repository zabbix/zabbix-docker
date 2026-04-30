# shellcheck shell=bash

: "${APACHE_RUN_DIR:=/tmp/apache2}"

: "${HTTPD_CONF_FILE:=/etc/apache2/httpd.conf}"
: "${APACHE_SITES_DIR:=/etc/apache2/conf.d}"
: "${APACHE_SSL_CONFIG_DIR:=/etc/ssl/apache2}"

prepare_web_server() {
    if [ "$(id -u)" -eq 0 ]; then
        APACHE_RUN_USER="${DAEMON_USER}"
        export APACHE_RUN_USER
    else
        APACHE_RUN_USER="$(id -un)"
        export APACHE_RUN_USER
    fi
    APACHE_RUN_GROUP="${DAEMON_GROUP}"
    export APACHE_RUN_GROUP

    info "** Adding Zabbix virtual host (HTTP)"
    if [ -f "${ZABBIX_CONF_DIR}/apache.conf" ]; then
        ln -sfT "${ZABBIX_CONF_DIR}/apache.conf" "${APACHE_SITES_DIR}/zabbix.conf"
    else
        info "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "${APACHE_SSL_CONFIG_DIR}/ssl.crt" ] && [ -f "${APACHE_SSL_CONFIG_DIR}/ssl.key" ]; then
        info "** Adding Zabbix virtual host (HTTPS)"
        if [ -f "${ZABBIX_CONF_DIR}/apache_ssl.conf" ]; then
            ln -sfT "${ZABBIX_CONF_DIR}/apache_ssl.conf" "${APACHE_SITES_DIR}/zabbix_ssl.conf"
        else
            info "**** Impossible to enable HTTPS virtual host"
        fi
    else
        info "**** Impossible to enable SSL support for Apache2. Certificates are missing."
    fi

    : "${HTTP_INDEX_FILE:=index.php}"
    export HTTP_INDEX_FILE

    : "${ENABLE_WEB_ACCESS_LOG:=true}"
    export APACHE_CUSTOM_LOG="/proc/self/fd/1"
    if [ "${ENABLE_WEB_ACCESS_LOG,,}" = "false" ]; then
        export APACHE_CUSTOM_LOG="/dev/null"
    fi

    : "${EXPOSE_WEB_SERVER_INFO:=on}"
    export APACHE_SERVER_TOKENS="OS"
    export APACHE_SERVER_SIGNATURE="On"
    if [ "${EXPOSE_WEB_SERVER_INFO,,}" = "off" ]; then
        export APACHE_SERVER_TOKENS="Prod"
        export APACHE_SERVER_SIGNATURE="Off"
    fi

    if [ -z "${WEB_REAL_IP_FROM:-}" ]; then
        [ -f "${APACHE_SITES_DIR}/server-common.inc" ] && sed -i '/WEB_REAL_IP_FROM/d' "${APACHE_SITES_DIR}/server-common.inc"
    fi

    if [ -z "${WEB_REAL_IP_HEADER:-}" ]; then
        [ -f "${APACHE_SITES_DIR}/server-common.inc" ] && sed -i '/WEB_REAL_IP_HEADER/d' "${APACHE_SITES_DIR}/server-common.inc"
    fi

    mkdir -p "${APACHE_RUN_DIR}"
}
