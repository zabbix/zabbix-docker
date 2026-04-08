prepare_web_server() {
    local fcgi_read_timeout

    [[ -f "$NGINX_CONF_FILE" ]] || error "Missing configuration file: $NGINX_CONF_FILE"

    if [ "$(id -u)" -eq 0 ]; then
        sed -i -e "/^[#;] user/s/.*/user ${DAEMON_USER};/" "$NGINX_CONF_FILE"
    fi

    if [ ! -f "/proc/net/if_inet6" ]; then
        [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ] && {
            sed -i '/listen \[::\]/d' "${ZABBIX_CONF_DIR}/nginx.conf"
            sed -i '/allow ::1/d' "${ZABBIX_CONF_DIR}/nginx.conf"
        }
        [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ] && {
            sed -i '/listen \[::\]/d' "${ZABBIX_CONF_DIR}/nginx_ssl.conf"
            sed -i '/allow ::1/d' "${ZABBIX_CONF_DIR}/nginx_ssl.conf"
        }
    fi

    info "** Adding Zabbix virtual host (HTTP)"
    if [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ]; then
        ln -sfT "${ZABBIX_CONF_DIR}/nginx.conf" "${NGINX_CONFD_DIR}/nginx.conf"
    else
        info "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "${NGINX_SSL_CONFIG_DIR}/ssl.crt" ] && [ -f "${NGINX_SSL_CONFIG_DIR}/ssl.key" ] \
        && [ -f "${NGINX_SSL_CONFIG_DIR}/dhparam.pem" ]; then
        info "** Enable SSL support for Nginx"
        if [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ]; then
            ln -sfT "${ZABBIX_CONF_DIR}/nginx_ssl.conf" "${NGINX_CONFD_DIR}/nginx_ssl.conf"
        else
            info "**** Impossible to enable HTTPS virtual host"
        fi
    else
        info "**** Impossible to enable SSL support for Nginx. Certificates are missing."
    fi

    : "${ZBX_MAXEXECUTIONTIME:=3}"
    fcgi_read_timeout=$(( ZBX_MAXEXECUTIONTIME + 1 ))

    : "${HTTP_INDEX_FILE:=index.php}"

    [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ] && sed -i \
        -e "s/{FCGI_READ_TIMEOUT}/${fcgi_read_timeout}/g" \
        -e "s/{HTTP_INDEX_FILE}/${HTTP_INDEX_FILE}/g" \
        "${ZABBIX_CONF_DIR}/nginx.conf"

    [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ] && sed -i \
        -e "s/{FCGI_READ_TIMEOUT}/${fcgi_read_timeout}/g" \
        -e "s/{HTTP_INDEX_FILE}/${HTTP_INDEX_FILE}/g" \
        "${ZABBIX_CONF_DIR}/nginx_ssl.conf"

    : "${ENABLE_WEB_ACCESS_LOG:=true}"
    if [ "${ENABLE_WEB_ACCESS_LOG,,}" = "false" ]; then
        [ -f "${NGINX_CONF_FILE}" ] && sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "${NGINX_CONF_FILE}"

        [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ] && sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "${ZABBIX_CONF_DIR}/nginx.conf"

        [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ] && sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "${ZABBIX_CONF_DIR}/nginx_ssl.conf"
    fi

    : "${EXPOSE_WEB_SERVER_INFO:=on}"
    [[ "${EXPOSE_WEB_SERVER_INFO,,}" != "off" ]] && EXPOSE_WEB_SERVER_INFO="on"

    export EXPOSE_WEB_SERVER_INFO
    [ -f "${NGINX_CONF_FILE}" ] && sed -i \
        -e "s/{EXPOSE_WEB_SERVER_INFO}/${EXPOSE_WEB_SERVER_INFO}/g" \
        "${NGINX_CONF_FILE}"

    if [ -n "${WEB_REAL_IP_FROM:-}" ]; then
        WEB_REAL_IP_FROM="set_real_ip_from ${WEB_REAL_IP_FROM};"
    else
        WEB_REAL_IP_FROM=""
    fi
    WEB_REAL_IP_FROM="$(escape_special_chars "$WEB_REAL_IP_FROM")"

    if [ -n "${WEB_REAL_IP_HEADER:-}" ]; then
        WEB_REAL_IP_HEADER="real_ip_header ${WEB_REAL_IP_HEADER};"
    else
        WEB_REAL_IP_HEADER=""
    fi
    WEB_REAL_IP_HEADER="$(escape_special_chars "$WEB_REAL_IP_HEADER")"

    [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ] && sed -i \
        -e "s#{WEB_REAL_IP_FROM}#${WEB_REAL_IP_FROM}#g" \
        -e "s#{WEB_REAL_IP_HEADER}#${WEB_REAL_IP_HEADER}#g" \
        "${ZABBIX_CONF_DIR}/nginx.conf"

    [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ] && sed -i \
        -e "s#{WEB_REAL_IP_FROM}#${WEB_REAL_IP_FROM}#g" \
        -e "s#{WEB_REAL_IP_HEADER}#${WEB_REAL_IP_HEADER}#g" \
        "${ZABBIX_CONF_DIR}/nginx_ssl.conf"
}
