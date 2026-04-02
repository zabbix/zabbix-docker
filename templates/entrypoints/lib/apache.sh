prepare_web_server() {
    if [ "$(id -u)" -eq 0 ]; then
        export APACHE_RUN_USER="${DAEMON_USER}"
    else
        export APACHE_RUN_USER="$(id -un)"
    fi
    export APACHE_RUN_GROUP="${DAEMON_GROUP}"

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
        [ -f "${ZABBIX_CONF_DIR}/apache.conf" ] && sed -i '/WEB_REAL_IP_FROM/d' "${ZABBIX_CONF_DIR}/apache.conf"
        [ -f "${ZABBIX_CONF_DIR}/apache_ssl.conf" ] && sed -i '/WEB_REAL_IP_FROM/d' "${ZABBIX_CONF_DIR}/apache_ssl.conf"
    fi

    if [ -z "${WEB_REAL_IP_HEADER:-}" ]; then
        [ -f "${ZABBIX_CONF_DIR}/apache.conf" ] && sed -i '/WEB_REAL_IP_HEADER/d' "${ZABBIX_CONF_DIR}/apache.conf"
        [ -f "${ZABBIX_CONF_DIR}/apache_ssl.conf" ] && sed -i '/WEB_REAL_IP_HEADER/d' "${ZABBIX_CONF_DIR}/apache_ssl.conf"
    fi

    mkdir -p "${APACHE_RUN_DIR}"
}
