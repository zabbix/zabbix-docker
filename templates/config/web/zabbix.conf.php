<?php

$DB['TYPE']     = getenv('DB_SERVER_TYPE');
$DB['SERVER']   = getenv('DB_SERVER_HOST');
$DB['PORT']     = getenv('DB_SERVER_PORT');
$DB['DATABASE'] = getenv('DB_SERVER_DBNAME');
$DB['USER']     = (! getenv('VAULT_TOKEN') || ! getenv('ZBX_VAULTURL')) ? getenv('DB_SERVER_USER') : '';
$DB['PASSWORD'] = (! getenv('VAULT_TOKEN') || ! getenv('ZBX_VAULTURL')) ? getenv('DB_SERVER_PASS') : '';

$DB['SCHEMA'] = getenv('DB_SERVER_SCHEMA');

if (getenv('ZBX_SERVER_HOST')) {
    $ZBX_SERVER      = getenv('ZBX_SERVER_HOST');
    $ZBX_SERVER_PORT = getenv('ZBX_SERVER_PORT');
}
$ZBX_SERVER_NAME = getenv('ZBX_SERVER_NAME');

$DB['ENCRYPTION']               = getenv('ZBX_DB_ENCRYPTION') == 'true' ? true: false;
$DB['KEY_FILE']                 = getenv('ZBX_DB_KEY_FILE');
$DB['CERT_FILE']                = getenv('ZBX_DB_CERT_FILE');
$DB['CA_FILE']                  = getenv('ZBX_DB_CA_FILE');
$DB['VERIFY_HOST']              = getenv('ZBX_DB_VERIFY_HOST') == 'true' ? true: false;
$DB['CIPHER_LIST']              = getenv('ZBX_DB_CIPHER_LIST') ? getenv('ZBX_DB_CIPHER_LIST') : '';

$DB['VAULT']                    = getenv('ZBX_VAULT');
$DB['VAULT_URL']                = getenv('ZBX_VAULTURL');
$DB['VAULT_PREFIX']		= getenv('ZBX_VAULTPREFIX');
$DB['VAULT_DB_PATH']            = getenv('ZBX_VAULTDBPATH');
$DB['VAULT_TOKEN']              = getenv('VAULT_TOKEN');

if (file_exists('/etc/zabbix/web/certs/vault.crt')) {
   $DB['VAULT_CERT_FILE'] = '/etc/zabbix/web/certs/vault.crt';
}
elseif (file_exists(getenv('ZBX_VAULTCERTFILE'))) {
   $DB['VAULT_CERT_FILE'] = getenv('ZBX_VAULTCERTFILE');
}
else {
   $DB['VAULT_CERT_FILE'] = '';
}

if (file_exists('/etc/zabbix/web/certs/vault.key')) {
   $DB['VAULT_KEY_FILE'] = '/etc/zabbix/web/certs/vault.key';
}
elseif (file_exists(getenv('ZBX_VAULTKEYFILE'))) {
   $DB['VAULT_KEY_FILE'] = getenv('ZBX_VAULTKEYFILE');
}
else {
   $DB['VAULT_KEY_FILE'] = '';
}

$DB['VAULT_CACHE']              = getenv('ZBX_VAULTCACHE') == 'true' ? true: false;

$DB['DOUBLE_IEEE754']           = getenv('DB_DOUBLE_IEEE754') == 'true' ? true: false;

$IMAGE_FORMAT_DEFAULT  = IMAGE_FORMAT_PNG;

$history_providers = str_replace("'","\"",getenv('ZBX_HISTORYPROVIDERS'));

if (json_decode($history_providers)) {
   $HISTORY_PROVIDERS[] = json_decode($history_providers, true);
}

// Used for SAML authentication.
if (file_exists('/etc/zabbix/web/certs/sp.key')) {
   $SSO['SP_KEY'] = '/etc/zabbix/web/certs/sp.key';
}
elseif (file_exists(getenv('ZBX_SSO_SP_KEY'))) {
   $SSO['SP_KEY'] = getenv('ZBX_SSO_SP_KEY');
}
else {
   $SSO['SP_KEY'] = '';
}

if (file_exists('/etc/zabbix/web/certs/sp.crt')) {
   $SSO['SP_CERT'] = '/etc/zabbix/web/certs/sp.crt';
}
elseif (file_exists(getenv('ZBX_SSO_SP_CERT'))) {
   $SSO['SP_CERT'] = getenv('ZBX_SSO_SP_CERT');
}
else {
   $SSO['SP_CERT'] = '';
}

if (file_exists('/etc/zabbix/web/certs/idp.crt')) {
   $SSO['IDP_CERT'] = '/etc/zabbix/web/certs/idp.crt';
}
elseif (file_exists(getenv('ZBX_SSO_IDP_CERT'))) {
   $SSO['IDP_CERT'] = getenv('ZBX_SSO_IDP_CERT');
}
else {
   $SSO['IDP_CERT'] = '';
}

$sso_settings = str_replace("'","\"",getenv('ZBX_SSO_SETTINGS'));
$SSO['SETTINGS'] = (json_decode($sso_settings)) ? json_decode($sso_settings, true) : array();

$SSO['CERT_STORAGE']		= getenv('ZBX_CERT_STORAGE') ? getenv('ZBX_CERT_STORAGE') : 'database';

$ZBX_FEATURE_FLAGS['banners_enabled'] = getenv('ZBX_BANNERS_ENABLED') ? getenv('ZBX_BANNERS_ENABLED') : true;

$ZBX_FEATURE_FLAGS['http_auth_enabled'] = getenv('ZBX_HTTP_AUTH_ENABLED') ? getenv('ZBX_HTTP_AUTH_ENABLED') : true;

$ZBX_FEATURE_FLAGS['modules_config_enabled'] = getenv('ZBX_MODULES_CONFIG_ENABLED') ? getenv('ZBX_MODULES_CONFIG_ENABLED') : true;

$media_type_denylist = str_replace("'","\"",getenv('ZBX_MEDIA_TYPE_DENYLIST'));
$ZBX_FEATURE_FLAGS['media_type_denylist'] = (json_decode($media_type_denylist)) ? json_decode($media_type_denylist, true) : array();

$ZBX_SERVER_TLS['ACTIVE'] = getenv('ZBX_SERVER_TLS_ACTIVE') == 'true' ? true : false;
$ZBX_SERVER_TLS['CA_FILE'] = file_exists(getenv('ZBX_SERVER_TLS_CAFILE')) ? getenv('ZBX_SERVER_TLS_CAFILE') : '';
$ZBX_SERVER_TLS['KEY_FILE'] = file_exists(getenv('ZBX_SERVER_TLS_KEYFILE')) ? getenv('ZBX_SERVER_TLS_KEYFILE') : '';
$ZBX_SERVER_TLS['CERT_FILE'] = file_exists(getenv('ZBX_SERVER_TLS_CERTFILE')) ? getenv('ZBX_SERVER_TLS_CERTFILE') : '';
$ZBX_SERVER_TLS['CERTIFICATE_ISSUER']  = getenv('ZBX_SERVER_TLS_CERT_ISSUER');
$ZBX_SERVER_TLS['CERTIFICATE_SUBJECT'] = getenv('ZBX_SERVER_TLS_CERT_SUBJECT');
