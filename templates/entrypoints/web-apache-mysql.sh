#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/mysql.sh"
source "${ENTRYPOINT_LIBS}/php.sh"
source "${ENTRYPOINT_LIBS}/web.sh"
source "${ENTRYPOINT_LIBS}/apache.sh"

readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

: "${ZBX_SERVER_NAME:=Zabbix docker}"
: "${ZBX_SERVER_PORT:=10051}"
: "${PHP_TZ:=Europe/Riga}"
: "${DAEMON_USER:=apache}"
: "${DAEMON_GROUP:=apache}"

: "${APACHE_RUN_DIR:=/tmp/apache2}"

: "${HTTPD_CONF_FILE:=/etc/apache2/httpd.conf}"
: "${APACHE_SITES_DIR:=/etc/apache2/conf.d}"
: "${APACHE_SSL_CONFIG_DIR:=/etc/ssl/apache2}"
: "${PHP_CONFIG_FILE:=/etc/php85/php-fpm.d/zabbix.conf}"

#################################################

info "** Deploying Zabbix web-interface (Apache) with MySQL database"

check_db_variables "zabbix"
check_db_connect "true"
prepare_php_config "MYSQL"
prepare_web_server
prepare_zbx_config

info "########################################################"

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
