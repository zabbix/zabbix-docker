# syntax=docker/dockerfile:1
# escape=`
ARG OS_BASE_IMAGE=mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022

ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5

ARG BUILD_BASE_IMAGE=zabbix-build-agent:ltsc2022-agent2-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV ZABBIX_USER_HOME_DIR="C:\zabbix" `
    ZABBIX_CONF_DIR="C:\zabbix\conf"

ENV ZBX_PERSISTENTBUFFERFILE="$ZABBIX_USER_HOME_DIR\buffer\agent2.db" `
    ZBX_USERPARAMETERDIR="${ZABBIX_USER_HOME_DIR}\user_scripts"

LABEL org.opencontainers.image.title="Zabbix agent 2" `
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" `
      org.opencontainers.image.vendor="Zabbix SIA" `
      org.opencontainers.image.url="https://zabbix.com/" `
      org.opencontainers.image.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" `
      org.opencontainers.image.licenses="AGPL v3.0" `
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" `
      org.opencontainers.image.version="${ZBX_VERSION}" `
      org.opencontainers.image.source="${ZBX_SOURCES}"

SHELL ["pwsh", "-Command", "$ErrorActionPreference = 'Stop';"]

WORKDIR C:\zabbix\

COPY --from=builder ["C:\\zabbix-${ZBX_VERSION}-output", "C:\\zabbix"]
COPY [".\\docker-entrypoint.ps1", "C:\\zabbix"]

USER ContainerAdministrator

RUN $env:PATH = [string]::Format('{0}\zabbix\sbin;{0}\zabbix\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    [Environment]::SetEnvironmentVariable('ZABBIX_USER_HOME_DIR', $env:ZABBIX_USER_HOME_DIR, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host ('Verifying Zabbix agent 2 {0} ...' -f $env:ZBX_VERSION); `
    zabbix_agent2.exe -V;

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_CONF_DIR | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_CONF_DIR\zabbix_agentd.d | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_CONF_DIR\zabbix_agent2.d | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_USER_HOME_DIR\enc | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_USER_HOME_DIR\enc_internal | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZABBIX_USER_HOME_DIR\buffer | Out-Null; `
    New-Item -Force -ItemType directory -Path $env:ZBX_USERPARAMETERDIR | Out-Null; `
    `
    net accounts /MaxPWAge:unlimited; `
    net user /add zabbix /expire:never /passwordreq:no; `
    $acl = Get-Acl -Path $env:ZABBIX_USER_HOME_DIR; `
    $ace = New-Object Security.AccessControl.FileSystemAccessRule ('zabbix', 'Modify', 'ContainerInherit, ObjectInherit', 'InheritOnly', 'Allow'); `
    $acl.AddAccessRule($ace); `
    Set-Acl -AclObject $acl -Path $env:ZABBIX_USER_HOME_DIR;

USER zabbix

EXPOSE 10050/TCP 31999/TCP

CMD C:\zabbix\docker-entrypoint.ps1 C:\zabbix\sbin\zabbix_agent2.exe -c C:\zabbix\conf\zabbix_agent2.conf -f
