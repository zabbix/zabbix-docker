# syntax=docker/dockerfile:1
# escape=`
ARG BUILD_BASE_IMAGE=mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022
ARG BASE_IMAGE=mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022
FROM $BUILD_BASE_IMAGE as builder_base

ARG PCRE_VERSION=8.45
ARG OPENSSL_VERSION=1.1.1l
ARG ZLIB_VERSION=1.2.11
ARG GOLANG_VERSION=1.17.1
ARG SEVEN_ZIP_VERSION=1900
ARG BUILD_ARCH=x64
ARG CPU_MODEL=AMD64

ARG VS_BUILDTOOLS_URL=https://aka.ms/vs/16/release/vs_buildtools.exe

ARG GIT_URL=https://github.com/git-for-windows/git/releases/download/v2.33.0.windows.2/MinGit-2.33.0.2-busybox-64-bit.zip
ARG MINGW_URL=https://deac-riga.dl.sourceforge.net/project/mingw-w64/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-win32/sjlj/x86_64-8.1.0-release-win32-sjlj-rt_v6-rev0.7z
ARG GOLANG_URL=https://golang.org/dl/go$GOLANG_VERSION.windows-amd64.msi
ARG CYGWIN_URL=https://cygwin.com/setup-x86_64.exe
ARG PCRE_URL=https://ftp.pcre.org/pub/pcre/pcre-$PCRE_VERSION.zip
ARG OPENSSL_URL=https://www.openssl.org/source/openssl-$OPENSSL_VERSION.tar.gz
ARG SEVEN_ZIP_URL=https://www.7-zip.org/a/7z$SEVEN_ZIP_VERSION-$BUILD_ARCH.msi

ENV BUILD_ARCH=$BUILD_ARCH CPU_MODEL=$CPU_MODEL `
    PCRE_VERSION=$PCRE_VERSION OPENSSL_VERSION=$OPENSSL_VERSION ZLIB_VERSION=$ZLIB_VERSION `
    GOLANG_VERSION=$GOLANG_VERSION SEVEN_ZIP_VERSION=$SEVEN_ZIP_VERSION `
    GIT_URL=$GIT_URL MINGW_URL=$MINGW_URL CYGWIN_URL=$CYGWIN_URL GOLANG_URL=$GOLANG_URL SEVEN_ZIP_URL=$SEVEN_ZIP_URL `
	PCRE_URL=$PCRE_URL OPENSSL_URL=$OPENSSL_URL

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
    Expand-Archive -Path git.zip -DestinationPath $env:SystemDrive\git\.; `
    Remove-Item -Force -Path $env:SystemDrive\git.zip; `
    $env:PATH = [string]::Format('{0}\git\cmd;{0}\git\mingw64\bin;{0}\git\usr\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("git version") ...'; `
    git version; `
    `
    Write-Host ('Downloading {0} ...' -f $env:GOLANG_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\go-amd64.msi -Uri $env:GOLANG_URL; `
    $sha256 = '7a360967708350354ea479500a8eada6a032e07eac5bd43142367ee5b0ab1df9'; `
    `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\go-amd64.msi -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum Go Lang failed!'; `
        exit 1; `
    }; `
    Write-Host 'Installing ...'; `
    Start-Process `
		-FilePath $env:SystemDrive\go-amd64.msi `
		-Wait `
		-ArgumentList '/qn /norestart'; `
    Remove-Item -Force -Path $env:SystemDrive\go-amd64.msi; `
    `
    $env:PATH = [string]::Format('{0}\Go\bin;', ${env:ProgramFiles}) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    Write-Host 'Verifying install ("go version") ...'; `
    go version; `
    `
    Write-Host ('Downloading {0} ...' -f $env:SEVEN_ZIP_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\7z.msi -Uri $env:SEVEN_ZIP_URL; `
    `
    $sha256 = 'a7803233eedb6a4b59b3024ccf9292a6fffb94507dc998aa67c5b745d197a5dc'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\7z.msi -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum 7-zip failed!'; `
        exit 1; `
    }; `
    `
    Write-Host 'Installing ...'; `
    Start-Process `
		-FilePath $env:SystemDrive\7z.msi `
		-Wait `
		-ArgumentList '/qn /norestart'; `
    Remove-Item -Force -Path $env:SystemDrive\7z.msi; `
    $env:PATH = [string]::Format('{0}\7-Zip;', ${env:ProgramFiles}) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("7z -h") ...'; `
    7z -h | Select -first 2; `
    `
    Write-Host ('Downloading {0} ...' -f $env:MINGW_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-WebRequest -OutFile $env:SystemDrive\mingw.7z -Uri $env:MINGW_URL; `
    `
    $sha256 = 'e8c65ddc655534b0330f66f7b480565621e8617cda9937d76ba141a22bf3b2fa'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\mingw.7z -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum Mingw-w64 failed!'; `
        exit 1; `
    }; `
    `
    Write-Host 'Installing ...'; `
    7z x $env:SystemDrive\mingw.7z; `
    Remove-Item -Force -Path $env:SystemDrive\mingw.7z; `
    $env:PATH = [string]::Format('{0}\mingw64\bin;', $env:SystemDrive) + $env:PATH; `
    [Environment]::SetEnvironmentVariable('PATH', $env:PATH, [EnvironmentVariableTarget]::Machine); `
    `
    Write-Host 'Verifying install ("mingw32-make -v") ...'; `
    mingw32-make -v; `
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
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    $sha256 = '5b709aa45ea3b8bb73052947200ad187f651a2049158fb5bbfed329e4322a977'; `
    if ((Get-FileHash $env:SystemDrive\pcre.zip -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum PCRE library failed!'; `
        exit 1; `
    }; `
    Write-Host 'Extracting archive ...'; `
    Expand-Archive -Path $env:SystemDrive\pcre.zip -DestinationPath $env:SystemDrive; `
    Remove-Item -Force -Path $env:SystemDrive\pcre.zip; `
    Rename-Item -Path $env:SystemDrive\pcre-$env:PCRE_VERSION -NewName $env:SystemDrive\pcre_build; `
    `
    Set-Location -Path $env:SystemDrive\pcre_build; `
    Write-Host 'Building PCRE library ...'; `
    cmake --log-level=ERROR `
        -G 'MinGW Makefiles' `
        -DBUILD_SHARED_LIBS=OFF `
        -DCMAKE_C_COMPILER=gcc `
        -DCMAKE_C_FLAGS='-O2 -g' `
        -DCMAKE_CXX_FLAGS='-O2 -g' `
        -DCMAKE_INSTALL_PREFIX="""$env:SystemDrive\pcre_output""" . ; `
    mingw32-make -s -j"""$env:NUMBER_OF_PROCESSORS"""; `
    mingw32-make -s -j"""$env:NUMBER_OF_PROCESSORS""" install; `
    Write-Host 'PCRE is ready...';

RUN Set-Location -Path $env:SystemDrive\.; `
    `
    Write-Host ('Downloading {0} ...' -f $env:CYGWIN_URL); `
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    (New-Object Net.WebClient).DownloadFile("""$env:CYGWIN_URL""", """$env:SystemDrive\setup.exe"""); `
    `
    $sha256 = 'b9219acd1241ffa4d38e19587f1ccc2854f951e451f3858efc9d2e1fe19d375c'; `
    Write-Host ('Verifying SHA256 ({0}) ...' -f $sha256); `
    if ((Get-FileHash $env:SystemDrive\setup.exe -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host 'Checksum Cygwin failed!'; `
        exit 1; `
    }; `
    Write-Host 'Installing ...'; `
    Start-Process $env:SystemDrive\setup.exe `
        -ArgumentList """-qnNdO -a x86_64 -R $env:SystemDrive\cygwin --site http://cygwin.mirror.constant.com -l $env:SystemDrive\cygwin\var\cache\setup --packages perl""" `
        -Wait `
        -NoNewWindow; `
    `
    Remove-Item -Force -Path $env:SystemDrive\setup.exe; `
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
    $env:PATH+=""";$env:SystemDrive\cygwin\bin"""; `
    perl Configure `
        mingw64 `
        no-shared `
        no-ui-console `
        no-tests `
#        enable-capieng `
        no-capieng `
        --api=1.1.0 `
        --prefix=$env:SystemDrive\openssl_output `
        --openssldir=$env:SystemDrive\openssl_output_ssl; `
    mingw32-make -s -j"""$env:NUMBER_OF_PROCESSORS"""; `
    mingw32-make -s -j"""$env:NUMBER_OF_PROCESSORS""" install_sw; `
    Write-Host 'OpenSSL is ready...'; `
    Remove-Item -Recurse -Force -Path $env:SystemDrive\cygwin;

FROM builder_base as builder_zabbix

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG ZABBIX_VERSION_RC_NUM=2400
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_COMPONENT=all
ENV ZBX_SOURCES=$ZBX_SOURCES MAJOR_VERSION=$MAJOR_VERSION ZBX_VERSION=$ZBX_VERSION

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
    (Get-Content src\go\pkg\version\version.go).replace('{ZABBIX_REVISION}', $ZbxRevision) | Set-Content src\go\pkg\version\version.go; `
    Set-Location -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\build\mingw; `
    mingw32-make -s -j"""$env:NUMBER_OF_PROCESSORS""" `
#        CGO_LDFLAGS="""-lssl -lcrypto -lcrypt32 -L$env:SystemDrive\openssl_output\lib -L$env:SystemDrive\pcre_output\lib""" `
        PCRE=$env:SystemDrive\pcre_output `
        OPENSSL=$env:SystemDrive\openssl_output `
        RFLAGS="""-DZABBIX_VERSION_RC_NUM=$env:ZABBIX_VERSION_RC_NUM"""; `
    `
    Write-Host 'Verifying build ("zabbix_agent2.exe -V") ...'; `
    & $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe -V; `
    dumpbin /dependents $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe; `
    `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix2 | Out-Null; `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix2\sbin | Out-Null; `
    New-Item -ItemType directory -Path $env:SystemDrive\zabbix2\conf | Out-Null; `
    `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe $env:SystemDrive\zabbix2\sbin; `
    Copy-Item -Path $env:SystemDrive\zabbix-$env:ZBX_VERSION\src\go\conf\zabbix_agent2.win.conf $env:SystemDrive\zabbix2\conf; `
    Write-Host 'Zabbix binaries are compiled...';

FROM $BASE_IMAGE

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
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

COPY --from=builder_zabbix ["C:\\zabbix2", "C:\\zabbix"]
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
