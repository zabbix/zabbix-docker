RemoteIPInternalProxy ${WEB_REAL_IP_FROM}
RemoteIPHeader ${WEB_REAL_IP_HEADER}

<LocationMatch "^/(ping|status)$">
    Require all granted
    SetHandler "proxy:unix:/tmp/php-fpm.sock|fcgi://localhost"
</LocationMatch>

<Directory "/usr/share/zabbix">
    Options FollowSymLinks
    AllowOverride None
    Require all granted

    <FilesMatch "\.php$">
        SetHandler "proxy:unix:/tmp/php-fpm.sock|fcgi://localhost"
    </FilesMatch>

    ExpiresActive On
    ExpiresByType image/x-icon "access plus 1 year"
    ExpiresByType text/css "access plus 14 days"
    ExpiresByType application/javascript "access plus 14 days"
    ExpiresByType text/javascript "access plus 14 days"
    ExpiresByType image/png "access plus 14 days"
    ExpiresByType image/jpeg "access plus 14 days"
    ExpiresByType image/gif "access plus 14 days"
    ExpiresByType text/plain "access plus 14 days"
    ExpiresByType application/xml "access plus 14 days"

    ExpiresByType image/svg+xml "access plus 14 days"
    ExpiresByType font/woff "access plus 30 days"
    ExpiresByType font/woff2 "access plus 30 days"
    ExpiresByType font/ttf "access plus 30 days"
    ExpiresByType font/otf "access plus 30 days"

    <FilesMatch "\.(ico|js|css|png|jpg|jpeg|gif|xml|txt|svg|woff|woff2|ttf|otf)$">
        Header append Cache-Control "public"
    </FilesMatch>
</Directory>

<DirectoryMatch "^/usr/share/zabbix/(conf|app|include|local|locale|vendor)$">
    Require all denied
    <Files "*.php">
        Require all denied
    </Files>
</DirectoryMatch>

<Location "^/usr/share/zabbix/robots.txt">
    Require all granted
</Location>

<Files "^/usr/share/zabbix/data/top_passwords.txt">
    Require all denied
</Files>

<LocationMatch "^/\.">
    Require all denied
</LocationMatch>
