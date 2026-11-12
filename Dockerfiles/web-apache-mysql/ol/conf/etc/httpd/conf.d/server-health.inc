<LocationMatch "^/ping$">
    Require all granted
    SetHandler "proxy:unix:/tmp/php-fpm.sock|fcgi://localhost"
</LocationMatch>

<LocationMatch "^/status$">
    Require expr "%{QUERY_STRING} in { '', 'html', 'json', 'openmetrics', 'xml' }"
    SetHandler "proxy:unix:/tmp/php-fpm-status.sock|fcgi://localhost"
</LocationMatch>

<LocationMatch "^/(?!ping$|status$)">
    Require all denied
</LocationMatch>
