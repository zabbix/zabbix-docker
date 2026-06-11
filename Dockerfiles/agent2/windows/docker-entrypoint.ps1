
# Script trace mode
if ($env:DEBUG_MODE -eq "true") {
    Set-PSDebug -trace 1
}

# Default Zabbix server host
if ([string]::IsNullOrEmpty($env:ZBX_SERVER_HOST)) {
    $env:ZBX_SERVER_HOST="zabbix-server"
}
# Default Zabbix server port number
if ([string]::IsNullOrEmpty($env:ZBX_SERVER_PORT)) {
    $env:ZBX_SERVER_PORT="10051"
}

# Default directories
# Internal directory for TLS related files, used when TLS*File specified as plain text values
$ZabbixInternalEncDir="$env:ZABBIX_USER_HOME_DIR/enc_internal"

function Update-Config-Var {
    Param (
        [Parameter(Mandatory=$true, Position=0)]
        [ValidateNotNullOrEmpty()]
        [string] $ConfigPath,
        [Parameter(Mandatory=$true, Position=1)]
        [ValidateNotNullOrEmpty()]
        [string]$VarName,
        [Parameter(Mandatory=$false, Position=2)]
        [string]$VarValue = $null,
        [Parameter(Mandatory=$false, Position=3)]
        [bool]$IsMultiple
    )

    $MaskList = "TLSPSKIdentity"

    if (-not(Test-Path -Path $ConfigPath -PathType Leaf)) {
        throw "**** Configuration file '$ConfigPath' does not exist"
    }

    if ($MaskList.Contains($VarName) -eq $true -And [string]::IsNullOrWhitespace($VarValue) -ne $true) {
        Write-Host -NoNewline "** Updating '$ConfigPath' parameter ""$VarName"": '****'. Enable DEBUG_MODE to view value ..."
    }
    else {
        Write-Host -NoNewline  "** Updating '$ConfigPath' parameter ""$VarName"": '$VarValue'..."
    }

    if ([string]::IsNullOrWhitespace($VarValue)) {
        if ((Get-Content $ConfigPath | %{$_ -match "^$VarName="}) -contains $true) {
            (Get-Content $ConfigPath) |
                Where-Object {$_ -notmatch "^$VarName=" } |
                Set-Content $ConfigPath
         }

        Write-Host "removed"
        return
    }

    if ($VarValue -eq '""') {
        (Get-Content $ConfigPath) | Foreach-Object { $_ -Replace "^($VarName=)(.*)", '$1' } | Set-Content $ConfigPath
        Write-Host "undefined"
        return
    }

    if ($VarName -match '^TLS.*File$') {
        $VarValue="$env:ZABBIX_USER_HOME_DIR\enc\$VarValue"
    }

    if ((Get-Content $ConfigPath | %{$_ -match "^$VarName="}) -contains $true -And $IsMultiple -ne $true) {
        (Get-Content $ConfigPath) | Foreach-Object { $_ -Replace "^$VarName=.+", "$VarName=$VarValue" } | Set-Content $ConfigPath

        Write-Host updated
    }
    elseif ((Get-Content $ConfigPath | select-string -pattern "^[#;] $VarName=").length -gt 1) {
        (Get-Content $ConfigPath) |
            Foreach-Object {
                $_
                if ($_ -match "^[#;] $VarName=$") {
                    "$VarName=$VarValue"
                }
            } | Set-Content $ConfigPath

        Write-Host "added first occurrence"
    }
    elseif ((Get-Content $ConfigPath | select-string -pattern "^[#;] $VarName=").length -gt 0) {
        (Get-Content $ConfigPath) |
            Foreach-Object {
                $_
                if ($_ -match "^[#;] $VarName=") {
                    "$VarName=$VarValue"
                }
            } | Set-Content $ConfigPath

        Write-Host "added"
    }
    else {
        Add-Content -Path $ConfigPath -Value "$VarName=$VarValue"
        Write-Host "added at the end"
    }
}

function Update-Config-Multiple-Var {
    Param (
        [Parameter(Mandatory=$true, Position=0)]
        [ValidateNotNullOrEmpty()]
        [string] $ConfigPath,
        [Parameter(Mandatory=$true, Position=1)]
        [ValidateNotNullOrEmpty()]
        [string]$VarName,
        [Parameter(Mandatory=$false, Position=2)]
        [string]$VarValue = $null
    )

    foreach ($value in $VarValue.split(',')) {
        Update-Config-Var $ConfigPath $VarName $value $true
    }
}

function File-Process-From-Env {
    Param (
        [Parameter(Mandatory=$true, Position=0)]
        [ValidateNotNullOrEmpty()]
        [string] $VarName,
        [Parameter(Mandatory=$false, Position=1)]
        [string]$FileName = $null,
        [Parameter(Mandatory=$false, Position=2)]
        [string]$VarValue = $null
    )

    if (![string]::IsNullOrEmpty($VarValue)) {
        $VarValue | Set-Content "$ZabbixInternalEncDir\$VarName"
        $FileName="$ZabbixInternalEncDir\$VarName"
    }

    if (![string]::IsNullOrEmpty($FileName)) {
        Set-Item env:$VarName -Value $FileName
    }

    $VarName=$VarName -replace 'FILE$'
    Set-Item env:$VarName -Value $null
}

function Prepare-Zbx-Agent-Config {
    if ([string]::IsNullOrEmpty($env:ZBX_PASSIVESERVERS)) {
        $env:ZBX_PASSIVESERVERS=""
    }
    if ([string]::IsNullOrEmpty($env:ZBX_ACTIVESERVERS)) {
        $env:ZBX_ACTIVESERVERS=""
    }

    if (![string]::IsNullOrEmpty($env:ZBX_SERVER_HOST) -And ![string]::IsNullOrEmpty($env:ZBX_PASSIVESERVERS)) {
        $env:ZBX_PASSIVESERVERS="$env:ZBX_SERVER_HOST,$env:ZBX_PASSIVESERVERS"
    }
    elseif (![string]::IsNullOrEmpty($env:ZBX_SERVER_HOST)) {
        $env:ZBX_PASSIVESERVERS=$env:ZBX_SERVER_HOST
    }

    if (![string]::IsNullOrEmpty($env:ZBX_SERVER_HOST)) {
        if (![string]::IsNullOrEmpty($env:ZBX_SERVER_PORT) -And $env:ZBX_SERVER_PORT -ne "10051") {
            $env:ZBX_SERVER_HOST="$env:ZBX_SERVER_HOST:$env:ZBX_SERVER_PORT"
        }
        if (![string]::IsNullOrEmpty($env:ZBX_ACTIVESERVERS)) {
            $env:ZBX_ACTIVESERVERS="$env:ZBX_SERVER_HOST,$env:ZBX_ACTIVESERVERS"
        }
        else {
            $env:ZBX_ACTIVESERVERS=$env:ZBX_SERVER_HOST
        }
    }

    if ([string]::IsNullOrWhitespace($env:ZBX_PASSIVE_ALLOW)) {
        $env:ZBX_PASSIVE_ALLOW="true"
    }

    if ($env:ZBX_PASSIVE_ALLOW -eq "true") {
        Write-Host "** Using '$env:ZBX_PASSIVESERVERS' servers for passive checks"
    }
    else {
        Set-Item env:ZBX_PASSIVESERVERS -Value $null
    }

    if ([string]::IsNullOrWhitespace($env:ZBX_ACTIVE_ALLOW)) {
        $env:ZBX_ACTIVE_ALLOW="true"
    }

    if ($env:ZBX_ACTIVE_ALLOW -eq "true") {
        Write-Host "** Using '$env:ZBX_ACTIVESERVERS' servers for active checks"
    }
    else {
        Set-Item env:ZBX_ACTIVESERVERS -Value $null
    }
    Set-Item env:ZBX_SERVER_HOST -Value $null
    Set-Item env:ZBX_SERVER_PORT -Value $null

    if ([string]::IsNullOrWhitespace($env:ZBX_ENABLEPERSISTENTBUFFER)) {
        $env:ZBX_ENABLEPERSISTENTBUFFER="true"
    }

    if ($env:ZBX_ENABLEPERSISTENTBUFFER -eq "true") {
        $env:ZBX_ENABLEPERSISTENTBUFFER="1"
    }
    else {
        Set-Item env:ZBX_ENABLEPERSISTENTBUFFER -Value $null
        Set-Item env:ZBX_PERSISTENTBUFFERFILE -Value $null
    }

    if ([string]::IsNullOrWhitespace($env:ZBX_ENABLESTATUSPORT)) {
        $env:ZBX_ENABLESTATUSPORT="true"
    }

    if ($env:ZBX_ENABLESTATUSPORT -eq "true") {
        $env:ZBX_STATUSPORT="31999"
    }

    Update-Config-Multiple-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2_item_keys.conf" "DenyKey" "$env:ZBX_DENYKEY"
    Update-Config-Multiple-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2_item_keys.conf" "AllowKey" "$env:ZBX_ALLOWKEY"

    File-Process-From-Env "ZBX_TLSCAFILE" "$env:ZBX_TLSCAFILE" "$env:ZBX_TLSCA"
    File-Process-From-Env "ZBX_TLSCRLFILE" "$env:ZBX_TLSCRLFILE" "$env:ZBX_TLSCRL"
    File-Process-From-Env "ZBX_TLSCERTFILE" "$env:ZBX_TLSCERTFILE" "$env:ZBX_TLSCERT"
    File-Process-From-Env "ZBX_TLSKEYFILE" "$env:ZBX_TLSKEYFILE" "$env:ZBX_TLSKEY"
    File-Process-From-Env "ZBX_TLSPSKFILE" "$env:ZBX_TLSPSKFILE" "$env:ZBX_TLSPSK"
}

function Prepare-Zbx-Agent-Plugins-Config {
    Write-Host "** Preparing Zabbix agent 2 (plugins) configuration files"

    Update-Config-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2.d\plugins.d\mongodb.conf" "Plugins.MongoDB.System.Path" "$env:ZABBIX_USER_HOME_DIR\zabbix-agent2-plugin\mongodb.exe"
    Update-Config-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2.d\plugins.d\postgresql.conf" "Plugins.PostgreSQL.System.Path" "$env:ZABBIX_USER_HOME_DIR\zabbix-agent2-plugin\postgresql.exe"
    Update-Config-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2.d\plugins.d\mssql.conf" "Plugins.MSSQL.System.Path" "$env:ZABBIX_USER_HOME_DIR\zabbix-agent2-plugin\mssql.exe"
    Update-Config-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2.d\plugins.d\ember.conf" "Plugins.EmberPlus.System.Path" "$env:ZABBIX_USER_HOME_DIR\zabbix-agent2-plugin\ember-plus.exe"
    if (Get-Command nvidia-smi.exe -errorAction SilentlyContinue) {
        Update-Config-Var "$env:ZABBIX_CONF_DIR\zabbix_agent2.d\plugins.d\nvidia.conf" "Plugins.NVIDIA.System.Path" "$env:ZABBIX_USER_HOME_DIR\zabbix-agent2-plugin\nvidia-gpu.exe"
    }
}

function ClearZbxEnv() {
    if ([string]::IsNullOrWhitespace($env:ZBX_CLEAR_ENV)) {
        return
    }

    $env_vars=Get-ChildItem env:* | Where-Object {$_.Name -match "^ZABBIX_.*" } | foreach { $_.Name }
    foreach ($env_var in $env_vars) {
        Set-Item env:$env_var -Value $null
    }
}

function PrepareAgent {
    Write-Host "** Preparing Zabbix agent 2"

    Prepare-Zbx-Agent-Config
    Prepare-Zbx-Agent-Plugins-Config
    ClearZbxEnv
}

$commandArgs = $args

if ($args.length -gt 0 -And $args[0].StartsWith('-')) {
    $commandArgs = @("C:\zabbix\sbin\zabbix_agent2.exe") + $args
}

if ($commandArgs.length -gt 0 -And $commandArgs[0] -eq "C:\zabbix\sbin\zabbix_agent2.exe") {
    PrepareAgent
}

if ($commandArgs.length -gt 0) {
    $exe, $exeArgs = $commandArgs
    & $exe @exeArgs
    exit $LASTEXITCODE
}
