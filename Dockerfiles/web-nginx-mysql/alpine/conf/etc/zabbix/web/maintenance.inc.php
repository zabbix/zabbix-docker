<?php
/*
** Copyright (C) 2001-2025 Zabbix SIA
**
** This program is free software: you can redistribute it and/or modify it under the terms of
** the GNU Affero General Public License as published by the Free Software Foundation, version 3.
**
** This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
** without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
** See the GNU Affero General Public License for more details.
**
** You should have received a copy of the GNU Affero General Public License along with this program.
** If not, see <https://www.gnu.org/licenses/>.
**/


// Maintenance mode.
if (getenv('ZBX_DENY_GUI_ACCESS') == 'true') {
    define('ZBX_DENY_GUI_ACCESS', 1);

    // Array of IP addresses, which are allowed to connect to frontend (optional).
    $ip_range = str_replace("'","\"",getenv('ZBX_GUI_ACCESS_IP_RANGE'));
    $ZBX_GUI_ACCESS_IP_RANGE = (json_decode($ip_range)) ? json_decode($ip_range, true) : array();

    // Message shown on warning screen (optional).
    $ZBX_GUI_ACCESS_MESSAGE = getenv('ZBX_GUI_WARNING_MSG');
}
