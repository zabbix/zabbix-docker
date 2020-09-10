<?php
// Zabbix GUI configuration file.
global $DB;

$DB['TYPE']     = getenv('DB_SERVER_TYPE');
$DB['SERVER']   = getenv('DB_SERVER_HOST');
$DB['PORT']     = getenv('DB_SERVER_PORT');
$DB['DATABASE'] = getenv('DB_SERVER_DBNAME');
$DB['USER']     = getenv('DB_SERVER_USER');
$DB['PASSWORD'] = getenv('DB_SERVER_PASS');

// Schema name. Used for IBM DB2 and PostgreSQL.
$DB['SCHEMA'] = getenv('DB_SERVER_SCHEMA');

$ZBX_SERVER      = getenv('ZBX_SERVER_HOST');
$ZBX_SERVER_PORT = getenv('ZBX_SERVER_PORT');
$ZBX_SERVER_NAME = getenv('ZBX_SERVER_NAME');

$IMAGE_FORMAT_DEFAULT = IMAGE_FORMAT_PNG;
?>
