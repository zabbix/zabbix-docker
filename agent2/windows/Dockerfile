# escape=`
ARG BUILD_BASE_IMAGE=mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022
ARG BASE_IMAGE=mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022
FROM $BUILD_BASE_IMAGE as builder_base

ARG PCRE_VERSION=8.45
ARG OPENSSL_VERSION=1.1.1l
ARG ZLIB_VERSION=1.2.11
ARG GOLANG_VERSION=1.17
ARG SEVEN_ZIP_VERSION=1900

ARG VS_BUILDTOOLS_URL=https://aka.ms/vs/16/release/vs_buildtools.exe

ARG GIT_URL=https://github.com/git-for-windows/git/releases/download/v2.33.0.windows.2/Git-2.33.0.2-64-bit.exe
ARG MINGW_URL=https://deac-riga.dl.sourceforge.net/project/mingw-w64/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-win32/sjlj/x86_64-8.1.0-release-win32-sjlj-rt_v6-rev0.7z
ARG CYGWIN_URL=https://cygwin.com/setup-x86_64.exe

ARG VS_DEV_CMD_LOCATION="C:\Program Files (x86)\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\VsDevCmd.bat"

ENV BUILD_ARCH=x64 CPU_MODEL=AMD64 `
    PCRE_VERSION=$PCRE_VERSION OPENSSL_VERSION=$OPENSSL_VERSION ZLIB_VERSION=$ZLIB_VERSION `
    GOLANG_VERSION=$GOLANG_VERSION SEVEN_ZIP_VERSION=$SEVEN_ZIP_VERSION `
    GIT_URL=$GIT_URL MINGW_URL=$MINGW_URL CYGWIN_URL=$CYGWIN_URL `
    VS_DEV_CMD_LOCATION=$VS_DEV_CMD_LOCATION

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

RUN Set-Location -Path "$env:SystemDrive"; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\git-installer.exe" $env:GIT_URL; `
    $sha256 = 'a5704733c219e9a0c96bfeb0febef62bc2518bdd4e358bc9519dbc5e63a3b5fe'; `
    if ((Get-FileHash "$env:SystemDrive\git-installer.exe" -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host "Checksum failed!"; `
        exit 1; `
    }; `
    Start-Process -FilePath "$env:SystemDrive\git-installer.exe" -ArgumentList '/VERYSILENT /SUPPRESSMSGBOXES /NORESTART /NOCANCEL /SP- /LOG /COMPONENTS="icons,assoc,assoc_sh,ext,ext\shellhere,gitlfs' -wait; `
    Remove-Item -Force -Path "$env:SystemDrive\git-installer.exe"; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\go$env:GOLANG_VERSION.windows-amd64.msi" https://golang.org/dl/go$env:GOLANG_VERSION.windows-amd64.msi; `
    $sha256 = '705254e0a459edae2c6bf4c88be0b4a14ac1cbbf9607a379112235f0271e6c4b'; `
    if ((Get-FileHash "$env:SystemDrive\go$env:GOLANG_VERSION.windows-amd64.msi" -Algorithm sha256).Hash -ne $sha256) { `
        Write-Host "Checksum failed!"; `
        exit 1; `
    }; `
    Start-Process -FilePath "$env:SystemDrive\go$env:GOLANG_VERSION.windows-amd64.msi" -Wait -ArgumentList '/qn /norestart'; `
    Remove-Item -Force -Path "$env:SystemDrive\go$env:GOLANG_VERSION.windows-amd64.msi"; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\7z$env:SEVEN_ZIP_VERSION-$env:BUILD_ARCH.msi" https://www.7-zip.org/a/7z$env:SEVEN_ZIP_VERSION-$env:BUILD_ARCH.msi; `
    Start-Process -FilePath "$env:SystemDrive\7z$env:SEVEN_ZIP_VERSION-$env:BUILD_ARCH.msi" -Wait -ArgumentList '/qn /norestart'; `
    Remove-Item -Force -Path "$env:SystemDrive\7z$env:SEVEN_ZIP_VERSION-$env:BUILD_ARCH.msi"; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\mingw.7z" $env:MINGW_URL; `
    Start-Process -FilePath "${env:ProgramFiles}\7-Zip\7z.exe" -Wait -ArgumentList """x $env:SystemDrive\mingw.7z"""; `
    Remove-Item -Force -Path "$env:SystemDrive\mingw.7z"; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\vs_buildtools.exe" $env:VS_BUILDTOOLS_URL; `
    Write-Host "Visual Studio components installing..."; `
    cmd /C start /w "$env:SystemDrive\vs_buildtools.exe" --quiet --wait --norestart --nocache modify `
        --installPath """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools""" `
        # https://docs.microsoft.com/en-us/visualstudio/install/workload-component-id-vs-build-tools?view=vs-2019
        --add Microsoft.VisualStudio.Component.Windows10SDK.19041 `
        --add Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
        --add Microsoft.VisualStudio.Component.VC.CMake.Project `
        --add Microsoft.VisualStudio.Workload.MSBuildTools `
        --add Microsoft.VisualStudio.Workload.NetCoreBuildTools `
        --add Microsoft.Net.Component.4.8.SDK `
        --remove Microsoft.VisualStudio.Component.Windows10SDK.10240 `
        --remove Microsoft.VisualStudio.Component.Windows10SDK.10586 `
        --remove Microsoft.VisualStudio.Component.Windows10SDK.14393 `
        --remove Microsoft.VisualStudio.Component.Windows81SDK `
        --remove Microsoft.VisualStudio.Component.WebDeploy `
        --remove Microsoft.Component.ClickOnce.MSBuild; `
    if ($err = dir $Env:TEMP -Filter dd_setup_*_errors.log | where Length -gt 0 | Get-Content) { `
	throw $err; `
    }; `
    Write-Host "Visual Studio components installed"; `
    Remove-Item -Force -Path "$env:SystemDrive\vs_buildtools.exe"; `
    Get-ChildItem -Path """${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer""" -Directory -Recurse | Remove-Item -Force –Recurse; `
    Remove-Item -Force -Recurse "$env:TEMP\*";

RUN Set-Location -Path "$env:SystemDrive"; `
    Import-Module """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll"""; `
    Enter-VsDevShell -VsInstallPath """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools""" -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    $env:PATH+=""";$env:SystemDrive\mingw64\bin"""; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\pcre-$env:PCRE_VERSION.zip" https://ftp.pcre.org/pub/pcre/pcre-$env:PCRE_VERSION.zip; `
    Expand-Archive -Path "$env:SystemDrive\pcre-$env:PCRE_VERSION.zip" -DestinationPath $env:SystemDrive ; `
    Rename-Item -Path "$env:SystemDrive\pcre-$env:PCRE_VERSION" -NewName "$env:SystemDrive\pcre_build_mingw"; `
    Set-Location -Path "$env:SystemDrive\pcre_build_mingw"; `
    cmake `
		-G """MinGW Makefiles""" `
		-DBUILD_SHARED_LIBS=OFF `
		-DCMAKE_C_COMPILER=gcc `
		-DCMAKE_C_FLAGS="""-O2 -g""" `
		-DCMAKE_CXX_FLAGS="""-O2 -g""" `
		-DCMAKE_INSTALL_PREFIX="""$env:SystemDrive\pcre_output_mingw""" . ; `
    mingw32-make -j"""$env:NUMBER_OF_PROCESSORS"""; `
    mingw32-make -j"""$env:NUMBER_OF_PROCESSORS""" install; `
    Remove-Item -Force -Path "$env:SystemDrive\pcre-$env:PCRE_VERSION.zip"

RUN Set-Location -Path "$env:SystemDrive"; `
    (New-Object Net.WebClient).DownloadFile("""$env:CYGWIN_URL""", """$env:SystemDrive\setup-x86_64.exe"""); `
    Start-Process "$env:SystemDrive\setup-x86_64.exe" -ArgumentList """-qnNdO -R $env:SystemDrive\Cygwin64 --site http://cygwin.mirror.constant.com -l $env:SystemDrive\cygwin64\var\cache\setup --packages perl""" -Wait -NoNewWindow; `
    Invoke-WebRequest -OutFile "$env:SystemDrive\openssl-$env:OPENSSL_VERSION.tar.gz" https://www.openssl.org/source/openssl-$env:OPENSSL_VERSION.tar.gz; `
    tar -zxf $env:SystemDrive\openssl-$env:OPENSSL_VERSION.tar.gz; `
    Rename-Item -Path "$env:SystemDrive\openssl-$env:OPENSSL_VERSION" -NewName "$env:SystemDrive\openssl_build_mingw"; `
    Set-Location -Path "$env:SystemDrive\openssl_build_mingw"; `
    $env:PATH+=""";$env:SystemDrive\mingw64\bin;$env:SystemDrive\Cygwin64\bin"""; `
    perl Configure `
		mingw64 `
		no-shared `
		no-ui-console `
		no-tests `
#		enable-capieng `
		no-capieng `
		--api=1.1.0 `
		--prefix="$env:SystemDrive\openssl_output_mingw" `
		--openssldir="$env:SystemDrive\openssl_output_ssl_mingw"; `
    mingw32-make -j"""$env:NUMBER_OF_PROCESSORS"""; `
    mingw32-make -j"""$env:NUMBER_OF_PROCESSORS""" install_sw; `
    Remove-Item -Recurse -Force -Path "$env:SystemDrive\Cygwin64"; `
    Remove-Item -Force -Path "$env:SystemDrive\setup-x86_64.exe"

FROM builder_base as builder_zabbix

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG ZABBIX_VERSION_RC_NUM=2400
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_COMPONENT=all
ENV ZBX_SOURCES=$ZBX_SOURCES MAJOR_VERSION=$MAJOR_VERSION ZBX_VERSION=$ZBX_VERSION

RUN Set-Location -Path "$env:SystemDrive"; `
    Import-Module """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\Microsoft.VisualStudio.DevShell.dll"""; `
    Enter-VsDevShell -VsInstallPath """${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools""" -DevCmdArguments """-arch=$env:BUILD_ARCH"""; `
    Start-Process -FilePath "${env:ProgramFiles}\Git\bin\git.exe" -Wait -ArgumentList """-c advice.detachedHead=false clone $env:ZBX_SOURCES --branch master --depth 1 --single-branch zabbix-$env:ZBX_VERSION"""; `
    Write-Host "Building Zabbix $env:ZBX_VERSION version"; `
    Set-Location -Path "$env:SystemDrive\zabbix-$env:ZBX_VERSION"; `
    $ZbxRevision=(git rev-parse --short HEAD); `
    (Get-Content src\go\pkg\version\version.go).replace('{ZABBIX_REVISION}', $ZbxRevision) | Set-Content src\go\pkg\version\version.go; `
    $env:PATH+=""";$env:SystemDrive\mingw64\bin"""; `
    Set-Location -Path "$env:SystemDrive\zabbix-$env:ZBX_VERSION\build\mingw"; `
    mingw32-make -j"""$env:NUMBER_OF_PROCESSORS""" `
		PCRE="$env:SystemDrive\pcre_output_mingw" `
		OPENSSL="$env:SystemDrive\openssl_output_mingw" `
		RFLAGS="""-DZABBIX_VERSION_RC_NUM=$env:ZABBIX_VERSION_RC_NUM""" `
#		CGO_LDFLAGS="""-lssl -lcrypto -lcrypt32 -L$env:SystemDrive\openssl_output_mingw\lib -L$env:SystemDrive\pcre_output_mingw\lib""" `
		OPENSSL="$env:SystemDrive\openssl_output_mingw"; `
    New-Item -ItemType directory -Path "$env:SystemDrive\zabbix2" | Out-Null; `
    New-Item -ItemType directory -Path "$env:SystemDrive\zabbix2\sbin" | Out-Null; `
    New-Item -ItemType directory -Path "$env:SystemDrive\zabbix2\conf" | Out-Null; `
    & $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe -V; `
    dumpbin /dependents $env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe; `
    Copy-Item -Path "$env:SystemDrive\zabbix-$env:ZBX_VERSION\bin\win64\zabbix_agent2.exe" "$env:SystemDrive\zabbix2\sbin"; `
    Copy-Item -Path "$env:SystemDrive\zabbix-$env:ZBX_VERSION\src\go\conf\zabbix_agent2.win.conf" "$env:SystemDrive\zabbix2\conf"

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

RUN setx /M PATH """$env:PATH;C:\zabbix\sbin;C:\zabbix\bin;"""; `
	setx /M ZbxDir """C:\zabbix"""

RUN Set-Location -Path "$env:SystemDrive"; `
	New-Item -ItemType directory -Path "$env:ZbxDir\conf\zabbix_agentd.d" | Out-Null; `
	New-Item -ItemType directory -Path "$env:ZbxDir\enc" | Out-Null; `
	New-Item -ItemType directory -Path "$env:ZbxDir\modules" | Out-Null; `
	New-Item -ItemType directory -Path "$env:ZbxDir\buffer" | Out-Null; `
	net user /add zabbix; `
	$acl = Get-Acl -Path $env:ZbxDir; `
	$ace = New-Object Security.AccessControl.FileSystemAccessRule ('zabbix', 'Modify', 'ContainerInherit, ObjectInherit', 'InheritOnly', 'Allow'); `
	$acl.AddAccessRule($ace); `
	Set-Acl -AclObject $acl -Path $env:ZbxDir

USER zabbix

EXPOSE 10050/TCP 31999/TCP

CMD C:\zabbix\docker-entrypoint.ps1 C:\zabbix\sbin\zabbix_agent2.exe -c c:\zabbix\conf\zabbix_agent2.win.conf -f
