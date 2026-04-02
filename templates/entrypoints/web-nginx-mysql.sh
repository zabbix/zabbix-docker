#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/${DB_ENGINE}.sh"
source "${ENTRYPOINT_LIBS}/php.sh"
source "${ENTRYPOINT_LIBS}/web.sh"
source "${ENTRYPOINT_LIBS}/nginx.sh"

readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

: "${ZBX_SERVER_NAME:=Zabbix docker}"
: "${ZBX_SERVER_PORT:=10051}"
: "${PHP_TZ:=Europe/Riga}"
: "${DAEMON_USER:=nginx}"
: "${DAEMON_GROUP:=nginx}"

: "${NGINX_CONF_FILE:=/etc/nginx/nginx.conf}"
: "${NGINX_CONFD_DIR:=/etc/nginx/http.d}"
: "${NGINX_SSL_CONFIG_DIR:=/etc/ssl/nginx}"
: "${PHP_CONFIG_FILE:=/etc/php84/php-fpm.d/zabbix.conf}"

#################################################

info "** Deploying Zabbix web-interface (Nginx) with MySQL database"

check_db_variables
check_db_connect "true"
prepare_php_config "MYSQL"
prepare_web_server
prepare_zbx_config

########################################################

if [ $# -gt 0 ]; then
    info "** Executing '$*'"
    exec "$@"
elif [ -f "/usr/bin/supervisord" ]; then
    info "** Executing supervisord"
    exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
else
    error "Unknown instructions. Exiting..."
fi

#################################################
