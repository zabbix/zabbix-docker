<?php
// Zabbix GUI configuration file.
global $DB, $HISTORY;

$DB['TYPE']     = 'POSTGRESQL';
$DB['SERVER']   = '{DB_SERVER_HOST}';
$DB['PORT']     = '{DB_SERVER_PORT}';
$DB['DATABASE'] = '{DB_SERVER_DBNAME}';
$DB['USER']     = '{DB_SERVER_USER}';
$DB['PASSWORD'] = '{DB_SERVER_PASS}';

// Schema name. Used for IBM DB2 and PostgreSQL.
$DB['SCHEMA'] = '{DB_SERVER_SCHEMA}';

$ZBX_SERVER      = '{ZBX_SERVER_HOST}';
$ZBX_SERVER_PORT = '{ZBX_SERVER_PORT}';
$ZBX_SERVER_NAME = '{ZBX_SERVER_NAME}';

// Used for TLS connection.
$DB['ENCRYPTION']		= {ZBX_DB_ENCRYPTION};
$DB['KEY_FILE']			= '{ZBX_DB_KEY_FILE}';
$DB['CERT_FILE']		= '{ZBX_DB_CERT_FILE}';
$DB['CA_FILE']			= '{ZBX_DB_CA_FILE}';
$DB['VERIFY_HOST']		= {ZBX_DB_VERIFY_HOST};
$DB['CIPHER_LIST']		= '{ZBX_DB_CIPHER_LIST}';

// Use IEEE754 compatible value range for 64-bit Numeric (float) history values.
// This option is enabled by default for new Zabbix installations.
// For upgraded installations, please read database upgrade notes before enabling this option.
$DB['DOUBLE_IEEE754']	= true;


$IMAGE_FORMAT_DEFAULT	= IMAGE_FORMAT_PNG;

// Elasticsearch url (can be string if same url is used for all types).
$HISTORY['url']   = '{ZBX_HISTORYSTORAGEURL}';
// Value types stored in Elasticsearch.
$HISTORY['types'] = {ZBX_HISTORYSTORAGETYPES};
