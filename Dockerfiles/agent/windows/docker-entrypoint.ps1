
# Script trace mode
if ($env:DEBUG_MODE -eq "true") {
    Set-PSDebug -trace 1
}

# Default Zabbix server host
if ([string]::IsNullOrWhitespace($env:ZBX_SERVER_HOST)) {
    $env:ZBX_SERVER_HOST="zabbix-server"
}
# Default Zabbix server port number
if ([string]::IsNullOrWhitespace($env:ZBX_SERVER_PORT)) {
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
        [string] $ZbxAgentConfig,
        [Parameter(Mandatory=$true, Position=1)]
        [ValidateNotNullOrEmpty()]
        [string] $VarName,
        [Parameter(Mandatory=$false, Position=2)]
        [string]$FileName = $null,
        [Parameter(Mandatory=$false, Position=3)]
        [string]$VarValue = $null
    )

    if (![string]::IsNullOrEmpty($VarValue)) {
        $VarValue | Set-Content "$ZabbixInternalEncDir\$VarName"
        $FileName="$ZabbixInternalEncDir\$VarName"
    }

    Update-Config-Var $ZbxAgentConfig "$VarName" "$FileName"
}


function Prepare-Zbx-Agent-Config {
    Write-Host "** Preparing Zabbix agent configuration file"

    $ZbxAgentConfig="$env:ZABBIX_CONF_DIR\zabbix_agentd.conf"

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
        Write-Host  "** Using '$env:ZBX_PASSIVESERVERS' servers for passive checks"
        Update-Config-Var $ZbxAgentConfig "Server" "$env:ZBX_PASSIVESERVERS"
    }
    else {
        Update-Config-Var $ZbxAgentConfig "Server"
    }

    if ([string]::IsNullOrWhitespace($env:ZBX_ACTIVE_ALLOW)) {
        $env:ZBX_ACTIVE_ALLOW="true"
    }

    if ($env:ZBX_ACTIVE_ALLOW -eq "true") {
        Write-Host "** Using '$env:ZBX_ACTIVESERVERS' servers for active checks"
        Update-Config-Var $ZbxAgentConfig "ServerActive" "$env:ZBX_ACTIVESERVERS"
    }
    else {
        Update-Config-Var $ZbxAgentConfig "ServerActive"
    }

    Update-Config-Var $ZbxAgentConfig "LogType" "console"
    Update-Config-Var $ZbxAgentConfig "LogFile"
    Update-Config-Var $ZbxAgentConfig "LogFileSize"
    Update-Config-Var $ZbxAgentConfig "DebugLevel" "$env:ZBX_DEBUGLEVEL"
    Update-Config-Var $ZbxAgentConfig "SourceIP" "$env:ZBX_SOURCEIP"
    Update-Config-Var $ZbxAgentConfig "EnableRemoteCommands" "$env:ZBX_ENABLEREMOTECOMMANDS"
    Update-Config-Var $ZbxAgentConfig "LogRemoteCommands" "$env:ZBX_LOGREMOTECOMMANDS"

    Update-Config-Var $ZbxAgentConfig "ListenPort" "$env:ZBX_LISTENPORT"
    Update-Config-Var $ZbxAgentConfig "ListenIP" "$env:ZBX_LISTENIP"
    Update-Config-Var $ZbxAgentConfig "ListenBacklog" "$env:ZBX_LISTENBACKLOG"
    Update-Config-Var $ZbxAgentConfig "StartAgents" "$env:ZBX_STARTAGENTS"

    Update-Config-Var $ZbxAgentConfig "HostInterface" "$env:ZBX_HOSTINTERFACE"
    Update-Config-Var $ZbxAgentConfig "HostInterfaceItem" "$env:ZBX_HOSTINTERFACEITEM"

    Update-Config-Var $ZbxAgentConfig "Hostname" "$env:ZBX_HOSTNAME"
    Update-Config-Var $ZbxAgentConfig "HostnameItem" "$env:ZBX_HOSTNAMEITEM"
    Update-Config-Var $ZbxAgentConfig "HostMetadata" "$env:ZBX_METADATA"
    Update-Config-Var $ZbxAgentConfig "HostMetadataItem" "$env:ZBX_METADATAITEM"
    Update-Config-Var $ZbxAgentConfig "RefreshActiveChecks" "$env:ZBX_REFRESHACTIVECHECKS"
    Update-Config-Var $ZbxAgentConfig "BufferSend" "$env:ZBX_BUFFERSEND"
    Update-Config-Var $ZbxAgentConfig "BufferSize" "$env:ZBX_BUFFERSIZE"
    Update-Config-Var $ZbxAgentConfig "MaxLinesPerSecond" "$env:ZBX_MAXLINESPERSECOND"
    # Please use include to enable Alias feature
#    update_config_multiple_var $ZBX_AGENT_CONFIG "Alias" $env:ZBX_ALIAS
    # Please use include to enable Perfcounter feature
#    update_config_multiple_var $ZBX_AGENT_CONFIG "PerfCounter" $env:ZBX_PERFCOUNTER
    Update-Config-Var $ZbxAgentConfig "Timeout" "$env:ZBX_TIMEOUT"
    Update-Config-Var $ZbxAgentConfig "Include" "$env:ZABBIX_CONF_DIR\zabbix_agentd.d\*.conf"
    Update-Config-Var $ZbxAgentConfig "UnsafeUserParameters" "$env:ZBX_UNSAFEUSERPARAMETERS"
    Update-Config-Var $ZbxAgentConfig "UserParameterDir" "$env:ZBX_USERPARAMETERDIR"
    Update-Config-Var $ZbxAgentConfig "TLSConnect" "$env:ZBX_TLSCONNECT"
    Update-Config-Var $ZbxAgentConfig "TLSAccept" "$env:ZBX_TLSACCEPT"
    File-Process-From-Env $ZbxAgentConfig "TLSCAFile" "$env:ZBX_TLSCAFILE" "$env:ZBX_TLSCA"
    File-Process-From-Env $ZbxAgentConfig "TLSCRLFile" "$env:ZBX_TLSCRLFILE" "$env:ZBX_TLSCRL"
    Update-Config-Var $ZbxAgentConfig "TLSServerCertIssuer" "$env:ZBX_TLSSERVERCERTISSUER"
    Update-Config-Var $ZbxAgentConfig "TLSServerCertSubject" "$env:ZBX_TLSSERVERCERTSUBJECT"
    File-Process-From-Env $ZbxAgentConfig "TLSCertFile" "$env:ZBX_TLSCERTFILE" "$env:ZBX_TLSCERT"
    Update-Config-Var $ZbxAgentConfig "TLSCipherAll" "$env:ZBX_TLSCIPHERALL"
    Update-Config-Var $ZbxAgentConfig "TLSCipherAll13" "$env:ZBX_TLSCIPHERALL13"
    Update-Config-Var $ZbxAgentConfig "TLSCipherCert" "$env:ZBX_TLSCIPHERCERT"
    Update-Config-Var $ZbxAgentConfig "TLSCipherCert13" "$env:ZBX_TLSCIPHERCERT13"
    Update-Config-Var $ZbxAgentConfig "TLSCipherPSK" "$env:ZBX_TLSCIPHERPSK"
    Update-Config-Var $ZbxAgentConfig "TLSCipherPSK13" "$env:ZBX_TLSCIPHERPSK13"
    File-Process-From-Env $ZbxAgentConfig "TLSKeyFile" "$env:ZBX_TLSKEYFILE" "$env:ZBX_TLSKEY"
    Update-Config-Var $ZbxAgentConfig "TLSPSKIdentity" "$env:ZBX_TLSPSKIDENTITY"
    File-Process-From-Env $ZbxAgentConfig "TLSPSKFile" "$env:ZBX_TLSPSKFILE" "$env:ZBX_TLSPSK"

    Update-Config-Multiple-Var $ZbxAgentConfig "DenyKey" "$env:ZBX_DENYKEY"
    Update-Config-Multiple-Var $ZbxAgentConfig "AllowKey" "$env:ZBX_ALLOWKEY"
}

function ClearZbxEnv() {
    if ($env:ZBX_CLEAR_ENV -eq "false") {
        return
    }

    $env_vars=Get-ChildItem env:* | Where-Object {$_.Name -match "^ZBX_.*" } | foreach { $_.Name }
    foreach ($env_var in $env_vars) {
        Set-Item env:$env_var -Value $null
    }
}

function PrepareAgent {
    Write-Host "** Preparing Zabbix agent"
    Prepare-Zbx-Agent-Config
    ClearZbxEnv
}

$commandArgs = $args

if ($args.length -gt 0 -And $args[0].StartsWith('-')) {
    $commandArgs = @("C:\zabbix\sbin\zabbix_agentd.exe") + $args
}

if ($commandArgs.length -gt 0 -And $commandArgs[0] -eq "C:\zabbix\sbin\zabbix_agentd.exe") {
    PrepareAgent
}

if ($commandArgs.length -gt 0) {
    $exe, $exeArgs = $commandArgs
    & $exe @exeArgs
    exit $LASTEXITCODE
}
