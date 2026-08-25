<?php

function env_string(string $name, string $default = ''): string {
    $value = getenv($name);

    return ($value === false) ? $default : $value;
}

function env_int(string $name, int $default = 0): int {
    $value = getenv($name);

    return ($value === false || $value === '') ? $default : (int) $value;
}

function env_bool(string $name, bool $default = false): bool {
    $value = getenv($name);

    if ($value === false || $value === '') {
        return $default;
    }

    return filter_var($value, FILTER_VALIDATE_BOOLEAN);
}

function env_json(string $name, array $default = []): array {
    $value = getenv($name);

    if ($value === false || $value === '') {
        return $default;
    }

    $decoded = json_decode(str_replace("'", '"', $value), true);

    return is_array($decoded) ? $decoded : $default;
}

function resolve_file(string $default_path, string $env_name): string {
    if (file_exists($default_path)) {
        return $default_path;
    }

    $path = getenv($env_name);

    return ($path && file_exists($path)) ? $path : '';
}

$DB['TYPE']     = env_string('DB_SERVER_TYPE');
$DB['SERVER']   = env_string('DB_SERVER_HOST');
$DB['PORT']     = env_int('DB_SERVER_PORT');
$DB['DATABASE'] = env_string('DB_SERVER_DBNAME');
$DB['SCHEMA']   = env_string('DB_SERVER_SCHEMA');

if (!getenv('ZBX_VAULT')) {
    $DB['USER']     = env_string('DB_SERVER_USER');
    $DB['PASSWORD'] = env_string('DB_SERVER_PASS');
}
else {
    $DB['USER']     = '';
    $DB['PASSWORD'] = '';
}

if (getenv('ZBX_SERVER_HOST')) {
    $ZBX_SERVER      = env_string('ZBX_SERVER_HOST');
    $ZBX_SERVER_PORT = env_string('ZBX_SERVER_PORT');
}

$ZBX_SERVER_NAME = env_string('ZBX_SERVER_NAME');

$DB['ENCRYPTION']  = env_bool('ZBX_DB_ENCRYPTION');
$DB['VERIFY_HOST'] = env_bool('ZBX_DB_VERIFY_HOST');
$DB['KEY_FILE']    = env_string('ZBX_DB_KEY_FILE');
$DB['CERT_FILE']   = env_string('ZBX_DB_CERT_FILE');
$DB['CA_FILE']     = env_string('ZBX_DB_CA_FILE');
$DB['CIPHER_LIST'] = env_string('ZBX_DB_CIPHER_LIST');

$DB['VAULT']               = env_string('ZBX_VAULT');
$DB['VAULT_URL']           = env_string('ZBX_VAULTURL');
$DB['VAULT_PREFIX']        = env_string('ZBX_VAULTPREFIX');
$DB['VAULT_DB_PATH']       = env_string('ZBX_VAULTDBPATH');
$DB['VAULT_TOKEN']         = env_string('VAULT_TOKEN');
$DB['VAULT_APP_ROLE_ID']   = env_string('ZBX_VAULTAPPROLEID');
$DB['VAULT_APP_SECRET_ID'] = env_string('ZBX_VAULTAPPSECRETID');

$DB['VAULT_CERT_FILE'] = resolve_file('/etc/zabbix/web/certs/vault.crt', 'ZBX_VAULTCERTFILE');
$DB['VAULT_KEY_FILE'] = resolve_file('/etc/zabbix/web/certs/vault.key', 'ZBX_VAULTKEYFILE');

$DB['VAULT_CACHE']    = env_bool('ZBX_VAULTCACHE');

$DB['DOUBLE_IEEE754'] = env_bool('ZBX_DB_DOUBLE_IEEE754');

$IMAGE_FORMAT_DEFAULT = IMAGE_FORMAT_PNG;

$history_providers = env_json('ZBX_HISTORYPROVIDERS');
if ($history_providers !== []) {
    $HISTORY_PROVIDERS[] = $history_providers;
}

$SSO['SP_KEY'] = resolve_file('/etc/zabbix/web/certs/sp.key', 'ZBX_SSO_SP_KEY');
$SSO['SP_CERT'] = resolve_file('/etc/zabbix/web/certs/sp.crt', 'ZBX_SSO_SP_CERT');
$SSO['IDP_CERT'] = resolve_file('/etc/zabbix/web/certs/idp.crt', 'ZBX_SSO_IDP_CERT');

$SSO['SETTINGS'] = env_json('ZBX_SSO_SETTINGS');
$SSO['CERT_STORAGE'] = env_string('ZBX_CERT_STORAGE', 'database');

$ZBX_FEATURE_FLAGS['banners_enabled'] = env_bool('ZBX_BANNERS_ENABLED', true);
$ZBX_FEATURE_FLAGS['http_auth_enabled'] = env_bool('ZBX_HTTP_AUTH_ENABLED', true);
$ZBX_FEATURE_FLAGS['modules_config_enabled'] = env_bool('ZBX_MODULES_CONFIG_ENABLED', true);
$ZBX_FEATURE_FLAGS['media_type_denylist'] = env_json('ZBX_MEDIA_TYPE_DENYLIST');

$ZBX_SERVER_TLS['ACTIVE'] = env_bool('ZBX_SERVER_TLS_ACTIVE');
$ZBX_SERVER_TLS['CA_FILE'] = resolve_file('', 'ZBX_SERVER_TLS_CAFILE');
$ZBX_SERVER_TLS['KEY_FILE'] = resolve_file('', 'ZBX_SERVER_TLS_KEYFILE');
$ZBX_SERVER_TLS['CERT_FILE'] = resolve_file('', 'ZBX_SERVER_TLS_CERTFILE');
$ZBX_SERVER_TLS['CERTIFICATE_ISSUER'] = env_string('ZBX_SERVER_TLS_CERT_ISSUER');
$ZBX_SERVER_TLS['CERTIFICATE_SUBJECT'] = env_string('ZBX_SERVER_TLS_CERT_SUBJECT');
