#!/bin/bash

os=ubuntu

version=$1
version=${version:-"latest"}

app_component=agent

if [[ ! $version =~ ^[0-9]*\.[0-9]*\.[0-9]*$ ]] && [ "$version" != "latest" ]; then
    echo "Incorrect syntax of the version"
    exit 1
fi

docker build -t zabbix-$app_component:$os-$version -f Dockerfile .

#docker rm -f zabbix-$app_component
#docker run --name zabbix-$app_component -t -d --link zabbix-server:zabbix-server zabbix-$app_component:$os-$version
