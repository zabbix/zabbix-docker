#!/bin/bash

set +e

if [ ! -f "Dockerfile" ]; then
    echo "Dockerfile is missing!"
    exit 1
fi

os=${PWD##*/}

version=$1
version=${version:-"latest"}

type=$2
type=${type:-"build"}

cd ../
app_component=${PWD##*/}
cd $os/

if [ "$app_component" == "zabbix-appliance" ]; then
    app_component="appliance"
fi

if [[ ! $version =~ ^[0-9]*\.[0-9]*\.[0-9]*$ ]] && [ "$version" != "latest" ]; then
    echo "Incorrect syntax of the version"
    exit 1
fi

if [ "$version" != "latest" ]; then
    VCS_REF=`git ls-remote https://git.zabbix.com/scm/zbx/zabbix.git  refs/tags/$version | cut -c1-10`
else
    MAJOR_VERSION=`cat Dockerfile | grep "ARG MAJOR_VERSION" | head -n1 | cut -f2 -d"="`
    MINOR_VERSION=`cat Dockerfile | grep "ARG ZBX_VERSION" | head -n1 | cut -f2 -d"."`

    VCS_REF=$MAJOR_VERSION.$MINOR_VERSION
fi

if hash docker 2>/dev/null; then
    exec_command='docker'
elif hash podman 2>/dev/null; then
    exec_command='podman'
else
    echo >&2 "Build command requires docker or podman.  Aborting.";
    exit 1;
fi

DOCKER_BUILDKIT=1 $exec_command build -t zabbix-$app_component:$os-$version --build-arg VCS_REF="$VCS_REF" --build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` -f Dockerfile .

if [ "$type" != "build" ]; then
    links=""
    env_vars=""

    if [[ $app_component =~ .*mysql.* ]]; then
        links="$links --link mysql-server:mysql"
        env_vars="$env_vars -e MYSQL_DATABASE=\"zabbix\" -e MYSQL_USER=\"zabbix\" -e MYSQL_PASSWORD=\"zabbix\" -e MYSQL_RANDOM_ROOT_PASSWORD=true"

        $exec_command rm -f mysql-server
        $exec_command run --name mysql-server -t $env_vars -d mysql:5.7
    fi

    if [ "$links" != "" ]; then
        sleep 5
    fi

    $exec_command rm -f zabbix-$app_component

    $exec_command run --name zabbix-$app_component -t -d $links $env_vars zabbix-$app_component:$os-$version
fi
