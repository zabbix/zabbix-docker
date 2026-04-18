# shellcheck shell=bash

: "${DAEMON_USER:=nginx}"

: "${NGINX_CONF_FILE:=/etc/nginx/nginx.conf}"
: "${NGINX_CONFD_DIR:=/etc/nginx/http.d}"
: "${NGINX_INCLUDES_DIR:=/etc/nginx/includes}"
: "${NGINX_SSL_CONFIG_DIR:=/etc/ssl/nginx}"

prepare_web_server() {
    local fcgi_read_timeout

    : > "${NGINX_INCLUDES_DIR}/access-log.conf"
    : > "${NGINX_INCLUDES_DIR}/user.conf"
    : > "${NGINX_INCLUDES_DIR}/real-ip.conf"
    : > "${NGINX_INCLUDES_DIR}/listen-ipv6.conf"
    : > "${NGINX_INCLUDES_DIR}/listen-ipv6-ssl.conf"

    if [ "$(id -u)" -eq 0 ]; then
        printf 'user %s;\n' "$DAEMON_USER" > "${NGINX_INCLUDES_DIR}/user.conf"
    fi

    info "** Adding Zabbix virtual host (HTTP)"

    if [ -f "${ZABBIX_CONF_DIR}/nginx.conf" ]; then
        ln -sfT "${ZABBIX_CONF_DIR}/nginx.conf" "${NGINX_CONFD_DIR}/nginx.conf"
        if [ -f "/proc/net/if_inet6" ]; then
            printf 'listen [::]:8080;\n' > "${NGINX_INCLUDES_DIR}/listen-ipv6.conf"
            printf 'allow ::1;\n' >> "${NGINX_INCLUDES_DIR}/listen-ipv6.conf"
        fi
    else
        warn "**** Impossible to enable HTTP virtual host"
        : > "${NGINX_INCLUDES_DIR}/listen-ipv6.conf"
    fi

    if [ -f "${NGINX_SSL_CONFIG_DIR}/ssl.crt" ] && [ -f "${NGINX_SSL_CONFIG_DIR}/ssl.key" ] \
        && [ -f "${NGINX_SSL_CONFIG_DIR}/dhparam.pem" ]; then
        info "** Enable SSL support for Nginx"
        if [ -f "${ZABBIX_CONF_DIR}/nginx_ssl.conf" ]; then
            ln -sfT "${ZABBIX_CONF_DIR}/nginx_ssl.conf" "${NGINX_CONFD_DIR}/nginx_ssl.conf"
            if [ -f "/proc/net/if_inet6" ]; then
                printf 'listen [::]:8443 ssl;\n' > "${NGINX_INCLUDES_DIR}/listen-ipv6-ssl.conf"
                printf 'allow ::1;\n' >> "${NGINX_INCLUDES_DIR}/listen-ipv6-ssl.conf"
            fi
        else
            warn "**** Impossible to enable HTTPS virtual host"
        fi
    else
        warn "**** Impossible to enable SSL support for Nginx. Certificates are missing."
    fi

    : "${ZBX_MAXEXECUTIONTIME:=3}"
    fcgi_read_timeout=$(( ZBX_MAXEXECUTIONTIME + 1 ))

    : "${HTTP_INDEX_FILE:=index.php}"

    [ -f "${NGINX_INCLUDES_DIR}/server-common.conf" ] && sed -i \
        -e "s/{FCGI_READ_TIMEOUT}/${fcgi_read_timeout}/g" \
        -e "s/{HTTP_INDEX_FILE}/${HTTP_INDEX_FILE}/g" \
        "${NGINX_INCLUDES_DIR}/server-common.conf"

    : "${EXPOSE_WEB_SERVER_INFO:=on}"
    [[ "${EXPOSE_WEB_SERVER_INFO,,}" != "off" ]] && EXPOSE_WEB_SERVER_INFO="on"

    [ -f "${NGINX_CONF_FILE}" ] && sed -i \
        -e "s/{EXPOSE_WEB_SERVER_INFO}/${EXPOSE_WEB_SERVER_INFO}/g" \
        "${NGINX_CONF_FILE}"

    : "${ENABLE_WEB_ACCESS_LOG:=true}"
    if [ "${ENABLE_WEB_ACCESS_LOG,,}" = "false" ]; then
        printf 'access_log off;\n' > "${NGINX_INCLUDES_DIR}/access-log.conf"
    else
        printf 'access_log /var/log/nginx/access.log main;\n' > "${NGINX_INCLUDES_DIR}/access-log.conf"
    fi

    if [ -n "${WEB_REAL_IP_FROM:-}" ]; then
        printf 'set_real_ip_from %s;\n' "${WEB_REAL_IP_FROM}" > "${NGINX_INCLUDES_DIR}/real-ip.conf"

        if [ -n "${WEB_REAL_IP_HEADER:-}" ]; then
            printf 'real_ip_header %s;\n' "${WEB_REAL_IP_HEADER}" >> "${NGINX_INCLUDES_DIR}/real-ip.conf"
        fi
    fi
}
