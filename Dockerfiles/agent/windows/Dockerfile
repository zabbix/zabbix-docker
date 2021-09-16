# syntax=docker/dockerfile:1
# escape=`
ARG BUILD_BASE_IMAGE=mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022
ARG BASE_IMAGE=mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022
FROM $BUILD_BASE_IMAGE as builder_base

ARG PCRE_VERSION=8.45
ARG OPENSSL_VERSION=1.1.1l
ARG LIBMODBUS_VERSION=master
ARG BUILD_ARCH=x64
ARG CPU_MODEL=AMD64

ARG VS_BUILDTOOLS_URL=https://aka.ms/vs/16/release/vs_buildtools.exe
ARG NASM_URL=https://www.nasm.us/pub/nasm/releasebuilds/2.15.05/win64/nasm-2.15.05-installer-x64.exe
ARG PERL_URL=https://strawberryperl.com/download/5.32.1.1/strawberry-perl-5.32.1.1-64bit.msi
ARG GIT_URL=https://github.com/git-for-windows/git/releases/download/v2.33.0.windows.2/MinGit-2.33.0.2-busybox-64-bit.zip
ARG PCRE_URL=https://ftp.pcre.org/pub/pcre/pcre-$PCRE_VERSION.zip
ARG OPENSSL_URL=https://www.openssl.org/source/openssl-$OPENSSL_VERSION.tar.gz
ARG LIBMODBUS_URL=https://github.com/stephane/libmodbus.git

ENV BUILD_ARCH=$BUILD_ARCH CPU_MODEL=$CPU_MODEL `
    PCRE_VERSION=$PCRE_VERSION OPENSSL_VERSION=$OPENSSL_VERSION LIBMODBUS_VERSION=$LIBMODBUS_VERSION `
    GIT_URL=$GIT_URL NASM_URL=$NASM_URL PERL_URL=$PERL_URL LIBMODBUS_URL=$LIBMODBUS_URL PCRE_URL=$PCRE_URL OPENSSL_URL=$OPENSSL_URL

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Write-Host ('Downloading {0} ...' -f $env:GIT_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\git.zip -Uri $env:GIT_URL; `
    `
    $sha256 = '273f55e881094d00877d64f56570b0c997c4da5dedcb26738d56481033c1eba1'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\git.zip -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum GIT for Windows failed!'; `
        exit 1; `
    }; `
    `
    Write-Host 'Installing ...'; `
    Expand-Archive `
        -Path git.zip `
        -DestinationPath $env:SystemDrive\git\.; `
    Remove-Item -Force -Path $env:SystemDrive\git.zip; `
    $env:PATH = [string]::Format('{0}\git\cmd;{0}\git\mingw64\bin;{0}\git\usr\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("git version") ...'; `
    git version; `
    `
    Write-Host ('Downloading {0} ...' -f $env:NASM_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\nasm_installer.exe -Uri $env:NASM_URL; `
    $sha256 = 'a02325b9fe54f917f5d6a3036637b38dbb6addf6f7ba9d344d9b943a143fe7d0'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\nasm_installer.exe -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum NASM failed!'; `
        exit 1; `
    }; `
    Write-Host 'Installing ...'; `
    Start-Process `
        -FilePath $env:SystemDrive\nasm_installer.exe `
        -ArgumentList '/S' -Wait; `
    Remove-Item -Force -Path $env:SystemDrive\nasm_installer.exe; `
    `
    $env:PATH = [string]::Format('{0}\NASM;', ${env:ProgramFiles}) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("nasm -v") ...'; `
    nasm -v; `
    `
    Write-Host ('Downloading {0} ...' -f $env:PERL_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\perl_installer.msi -Uri $env:PERL_URL; `
    `
    $sha256 = '241a881670164feb0b91bb69d39fbbf84c981bec0d9f8c19959f8f48fd177768'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\perl_installer.msi -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum Strawberry Perl failed!'; `
        exit 1; `
    }; `
    Write-Host 'Installing ...'; `
    Start-Process `
        -FilePath 'msiexec.exe' `
        -ArgumentList """/i $env:SystemDrive\perl_installer.msi /qn /norestart INSTALLDIR=$env:SystemDrive\Strawberry""" `
        -Wait; `
    Remove-Item -Force -Path  $env:SystemDrive\perl_installer.msi; `
    $env:PATH = [string]::Format('{0}\Strawberry\perl\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("perl -V") ...'; `
    perl -V; `
    `
    Write-Host 'Installing Text::Template...'; `
    cpan Text::Template; `
	`
    Write-Host ('Downloading {0} ...' -f $env:VS_BUILDTOOLS_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\vs_buildtools.exe $env:VS_BUILDTOOLS_URL; `
    `
    Write-Host ('{0} - Visual Studio components installing...' -f $(Get-Date -format 'u')); `
    cmd /C start /w $env:SystemDrive\vs_buildtools.exe `
        --quiet `
        --wait `
        --norestart `
        --nocache modify `
        --installPath """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools""" `
        # https://docs.microsoft.com/en-us/visualstudio/install/workload-component-id-vs-build-tools?view=vs-2019
        --add Microsoft.VisualStudio.Component.Windows10SDK.19041 `
        --add Microsoft.VisualStudio.Component.VC.CMake.Project; `
    if ($err = dir $Env:TEMP -Filter dd_setup_*_errors.log | where Length -gt 0 | Get-Content) { `
        throw $err; `
    }; `
    Wait-Process -name msiexec; `
    Write-Host ('{0} - Visual Studio components installed' -f $(Get-Date -format 'u')); `
    `
    Write-Host 'Visual Studio components installation cleanup'; `
    Remove-Item -Force -Path $env:SystemDrive\vs_buildtools.exe; `
    Get-ChildItem -Path """${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer""" -Directory -Recurse | Remove-Item -Force –Recurse; `
    Remove-Item -Force -Recurse $env:TEMP\*; `
    Write-Host 'Build environment is ready...';

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Import-Module ('{0}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll' -f ${env:ProgramFiles(x86)} ); `
    Enter-VsDevShell -VsInstallPath ('{0}\Microsoft Visual Studio\2019\BuildTools' -f ${env:ProgramFiles(x86)}) -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    `
    Write-Host ('Downloading {0} ...' -f $env:PCRE_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\pcre.zip -Uri $env:PCRE_URL; `
    `
    $sha256 = '5b709aa45ea3b8bb73052947200ad187f651a2049158fb5bbfed329e4322a977'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\pcre.zip -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum PCRE library failed!'; `
        exit 1; `
    }; `
    Write-Host 'Extracting archive ...'; `
    Expand-Archive -Path $env:SystemDrive\pcre.zip -DestinationPath $env:SystemDrive; `
    Remove-Item -Force -Path $env:SystemDrive\pcre.zip; `
    Rename-Item -Path $env:SystemDrive\pcre-$env:PCRE_VERSION -NewName $env:SystemDrive\pcre_build; `
    New-Item -ItemType directory -Path "$env:SystemDrive\pcre_build\build" | Out-Null; `
    `
    Set-Location -Path $env:SystemDrive\pcre_build\build; `
    Write-Host 'Building PCRE library ...'; `
    cmake --log-level=ERROR `
        -G 'Visual Studio 16 2019' `
	-A $env:BUILD_ARCH `
        -DPCRE_SUPPORT_UNICODE_PROPERTIES=ON `
        -DPCRE_SUPPORT_UTF=ON `
        -DCMAKE_C_FLAGS_RELEASE:string="""/MT""" ..; `
    msbuild PCRE.sln `
        -maxcpucount:"""$env:NUMBER_OF_PROCESSORS""" `
        /verbosity:quiet `
        /property:Configuration=Release `
        /property:Platform=$env:BUILD_ARCH `
        /target:pcre; `
    `
    Write-Host 'PCRE is ready...';

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Import-Module ('{0}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll' -f ${env:ProgramFiles(x86)} ); `
    Enter-VsDevShell -VsInstallPath ('{0}\Microsoft Visual Studio\2019\BuildTools' -f ${env:ProgramFiles(x86)}) -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    `
    Write-Host ('Downloading {0} ...' -f $env:OPENSSL_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\openssl.tar.gz -Uri $env:OPENSSL_URL; `
    `
    $sha256 = '0b7a3e5e59c34827fe0c3a74b7ec8baef302b98fa80088d7f9153aa16fa76bd1'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\openssl.tar.gz -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum OpenSSL library failed!'; `
        exit 1; `
    }; `
    `
    Write-Host 'Extracting archive ...'; `
    tar -zxf $env:SystemDrive\openssl.tar.gz; `
    Remove-Item -Force -Path $env:SystemDrive\openssl.tar.gz; `
    Rename-Item -Path $env:SystemDrive\openssl-$env:OPENSSL_VERSION -NewName $env:SystemDrive\openssl_build; `
    `
    Write-Host 'Building OpenSSL library...'; `
    Set-Location -Path $env:SystemDrive\openssl_build; `
    perl $env:SystemDrive\openssl_build\Configure `
        VC-WIN64A `
        no-shared `
        no-ui-console `
        no-tests `
#        enable-capieng `
        no-capieng `
        --api=1.1.0 `
        --prefix=$env:SystemDrive\openssl_output `
        --openssldir=$env:SystemDrive\openssl_output_ssl; `
    set CL=/MP; `
    nmake /S; `
    nmake /S install_sw; `
    Write-Host 'OpenSSL is ready...';

COPY modbus.vs16.vcxproj c:\
COPY modbus.vs16.sln c:\
COPY modbus.vs16.vcxproj.filters c:\

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Import-Module ('{0}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll' -f ${env:ProgramFiles(x86)} ); `
    Enter-VsDevShell -VsInstallPath ('{0}\Microsoft Visual Studio\2019\BuildTools' -f ${env:ProgramFiles(x86)}) -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    `
    Write-Host ('Checkout GIT {0} repository ...' -f $env:LIBMODBUS_URL); `
    git -c advice.detachedHead=false clone $env:LIBMODBUS_URL --branch $env:LIBMODBUS_VERSION --depth 1 --single-branch $env:SystemDrive\libmodbus; `
    `
    Write-Host 'Building Libmodbus library...'; `
    Copy-Item -Path $env:SystemDrive\modbus.vs16.vcxproj $env:SystemDrive\libmodbus\src\win32\modbus.vs16.vcxproj; `
    Copy-Item -Path $env:SystemDrive\modbus.vs16.sln $env:SystemDrive\libmodbus\src\win32\modbus.vs16.sln; `
    Copy-Item -Path $env:SystemDrive\modbus.vs16.vcxproj.filters $env:SystemDrive\libmodbus\src\win32\modbus.vs16.vcxproj.filters; `
    Set-Location -Path $env:SystemDrive\libmodbus\src\win32; `
    cscript .\configure.js; `
    msbuild modbus.vs16.sln `
        -maxcpucount:"""$env:NUMBER_OF_PROCESSORS""" `
        /verbosity:quiet `
        /property:Configuration=Release `
        /property:Platform=$env:BUILD_ARCH; `
    Write-Host 'Libmodbus is ready...';

FROM builder_base as builder_zabbix

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG ZABBIX_VERSION_RC_NUM=2400
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_COMPONENT=all
ENV ZBX_SOURCES=$ZBX_SOURCES MAJOR_VERSION=$MAJOR_VERSION ZBX_VERSION=$ZBX_VERSION ZABBIX_VERSION_RC_NUM=$ZABBIX_VERSION_RC_NUM

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Import-Module ('{0}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll' -f ${env:ProgramFiles(x86)} ); `
    Enter-VsDevShell -VsInstallPath ('{0}\Microsoft Visual Studio\2019\BuildTools' -f ${env:ProgramFiles(x86)}) -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    `
    Write-Host ('Checkout GIT {0} (master) repository ...' -f $env:ZBX_SOURCES); `
    git -c advice.detachedHead=false clone $env:ZBX_SOURCES --branch master --depth 1 --single-branch $env:SystemDrive\zabbix-$env:ZBX_VERSION; `
    `
    Write-Host ('Building Zabbix {0} version ...' -f $env:ZBX_VERSION); `
    Set-Location -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION; `
    $ZbxRevision=(git rev-parse --short HEAD); `
    (Get-Content include/version.h).replace('{ZABBIX_REVISION}', $ZbxRevision) | Set-Content include/version.h; `
    Set-Location -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\build\win32\project; `
    set CL=/MP; `
    nmake /S -f Makefile `
        CPU=$env:CPU_MODEL `
        CFLAGS="""/D ZABBIX_VERSION_REVISION=$ZbxRevision /D HAVE_LIBMODBUS_STATIC""" `
        PCREINCDIR=$env:SystemDrive\pcre_build\build `
        PCRELIBDIR=$env:SystemDrive\pcre_build\build\Release TLS=openssl `
        RFLAGS="""/D ZABBIX_VERSION_REVISION=$ZbxRevision /D ZABBIX_VERSION_RC_NUM=$env:ZABBIX_VERSION_RC_NUM""" `
#        LIBS="Crypt32.lib" `
        TLSINCDIR=$env:SystemDrive\openssl_output\include `
        TLSLIBDIR=$env:SystemDrive\openssl_output\lib `
        MODBINCDIR=$env:SystemDrive\libmodbus\src `
        MODBLIBDIR=$env:SystemDrive\libmodbus\src\win32\$env:BUILD_ARCH\Release `
        $env:ZBX_COMPONENT; `
    `
    Write-Host 'Verifying build ("zabbix_agent2.exe -V") ...'; `
    & $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agentd.exe -V; `
    dumpbin /dependents $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agentd.exe; `
    `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix | Out-Null; `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix\sbin | Out-Null; `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix\conf | Out-Null; `
    `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agentd.exe $env:SystemDrive\zabbix\sbin; `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_get.exe $env:SystemDrive\zabbix\sbin; `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_sender.exe $env:SystemDrive\zabbix\sbin; `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\conf\zabbix_agentd.win.conf $env:SystemDrive\zabbix\conf; `
    Write-Host 'Zabbix binaries are compiled...';

FROM $BASE_IMAGE

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

LABEL org.opencontainers.image.title="Zabbix agent" `
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" `
      org.opencontainers.image.vendor="Zabbix LLC" `
      org.opencontainers.image.url="https://zabbix.com/" `
      org.opencontainers.image.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" `
      org.opencontainers.image.licenses="GPL v2.0" `
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" `
      org.opencontainers.image.version="${ZBX_VERSION}" `
      org.opencontainers.image.source="${ZBX_SOURCES}"

SHELL ["pwsh", "-Command", "$ErrorActionPreference = 'Stop';"]

WORKDIR C:\zabbix\

COPY --from=builder_zabbix ["C:\\zabbix", "C:\\zabbix"]
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

EXPOSE 10050/TCP

CMD C:\zabbix\docker-entrypoint.ps1 C:\zabbix\sbin\zabbix_agentd.exe -c C:\zabbix\conf\zabbix_agentd.win.conf -f
