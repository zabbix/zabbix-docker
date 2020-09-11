<?php
/*
** Zabbix
** Copyright (C) 2001-2016 Zabbix SIA
**
** This program is free software; you can redistribute it and/or modify
** it under the terms of the GNU General Public License as published by
** the Free Software Foundation; either version 2 of the License, or
** (at your option) any later version.
**
** This program is distributed in the hope that it will be useful,
** but WITHOUT ANY WARRANTY; without even the implied warranty of
** MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
** GNU General Public License for more details.
**
** You should have received a copy of the GNU General Public License
** along with this program; if not, write to the Free Software
** Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
**/


// Maintenance mode
if (getenv('ZBX_DENY_GUI_ACCESS') == 'true') {
    define('ZBX_DENY_GUI_ACCESS', 1);

    // IP range, who are allowed to connect to FrontEnd
    $ip_range = str_replace("'","\"",getenv('ZBX_GUI_ACCESS_IP_RANGE'));
    $ZBX_GUI_ACCESS_IP_RANGE = (json_decode($ip_range)) ? json_decode($ip_range) : array();

    // MSG shown on Warning screen!
    $_REQUEST['warning_msg'] = getenv('ZBX_GUI_WARNING_MSG');
}
