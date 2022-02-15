# syntax=docker/dockerfile:1
# escape=`
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG BUILD_BASE_IMAGE=zabbix-build-agent:ltsc2022-agent2-${ZBX_VERSION}

ARG BASE_IMAGE=mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022

FROM ${BUILD_BASE_IMAGE} as builder

FROM $BASE_IMAGE

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

LABEL org.opencontainers.image.title="Zabbix agent 2" `
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" `
      org.opencontainers.image.vendor="Zabbix LLC" `
      org.opencontainers.image.url="https://zabbix.com/" `
      org.opencontainers.image.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" `
      org.opencontainers.image.licenses="GPL v2.0" `
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" `
      org.opencontainers.image.version="${ZBX_VERSION}" `
      org.opencontainers.image.source="${ZBX_SOURCES}"

SHELL ["pwsh", "-Command", "$ErrorActionPreference = 'Stop';"]

WORKDIR C:\zabbix\

COPY --from=builder ["C:\\zabbix2", "C:\\zabbix"]
COPY [".\\docker-entrypoint.ps1", "C:\\zabbix"]

USER ContainerAdministrator

RUN $env:PATH = [string]::Format('{0}\zabbix\sbin;{0}\zabbix\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    $env:ZBX_HOME = [string]::Format('{0}\zabbix', $env:SystemDrive); `
    [Environment]::SetEnvironmentVariable('ZBX_HOME', $env:ZBX_HOME, [EnvironmentVariableTarget]::Machine);

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    New-Item -ItemType directory -Path $env:ZBX_HOME\conf\zabbix_agentd.d | Out-Null; `
    New-Item -ItemType directory -Path $env:ZBX_HOME\enc | Out-Null; `
    New-Item -ItemType directory -Path $env:ZBX_HOME\modules | Out-Null; `
    New-Item -ItemType directory -Path $env:ZBX_HOME\buffer | Out-Null; `
    `
    net user /add zabbix; `
    $acl = Get-Acl -Path $env:ZBX_HOME; `
    $ace = New-Object Security.AccessControl.FileSystemAccessRule ('zabbix', 'Modify', 'ContainerInherit, ObjectInherit', 'InheritOnly', 'Allow'); `
    $acl.AddAccessRule($ace); `
    Set-Acl -AclObject $acl -Path $env:ZBX_HOME;

USER zabbix

EXPOSE 10050/TCP 31999/TCP

CMD C:\zabbix\docker-entrypoint.ps1 C:\zabbix\sbin\zabbix_agent2.exe -c C:\zabbix\conf\zabbix_agent2.win.conf -f
