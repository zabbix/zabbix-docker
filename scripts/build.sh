#!/bin/bash

fn_help() {
  echo -e "Usage: $0 -c <app_component> -g <zabbix_gid> -o <os> -v <version> -u <zabbix_uid>\n"
  echo "Defaults"
  echo "  -c: agent"
  echo "  -g: 20000"
  echo "  -o: ubuntu"
  echo "  -v: latest"
  echo -e "  -u: 20000\n"
  echo "Example #1: Build a new image and tag it as alpine os:"
  echo -e "            $0 -o alpine\n"
  echo "Example #2: Build a new image, tag it as alpine os and set zabbix user id inside container as 10000:"
  echo -e "            $0 -o alpine -u 10000\n"
  exit 1
}

fn_error() {
  echo -e "\nERROR\n"
  exit 1
}

os=ubuntu
version="latest"
zabbix_uid=20000
zabbix_gid=20000
app_component=agent

while getopts ":c:g:o:hv:u:" opt; do
  case $opt in
    c)
      app_component="${OPTARG}"
      ;;
    g)
      zabbix_gid="${OPTARG}"
      ;;
    h)
      fn_help
      ;;
    o)
      os="${OPTARG}"
      ;;
    v)
      version="${OPTARG}"
      ;;
    u)
      zabbix_uid="${OPTARG}"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      fn_help
      ;;
    :)
      echo "Option -$OPTARG requires an argument" >&2
      fn_help
      ;;
    *)
      echo "Unimplemented option: -$OPTARG" >&2
      fn_help
      ;;
  esac
done

if [[ ! $version =~ ^[0-9]*\.[0-9]*\.[0-9]*$ ]] && [ "$version" != "latest" ]; then
    echo "Incorrect syntax of the version"
    exit 1
fi

docker build \
  --build-arg DOCKER_GID=`getent group docker | cut -d: -f3` \
  --build-arg ZABBIX_UID=$zabbix_uid \
  --build-arg ZABBIX_GID=$zabbix_gid \
  -t zabbix-$app_component:$os-$version \
  -f Dockerfile .

#docker rm -f zabbix-$app_component
#docker run --name zabbix-$app_component -t -d --link zabbix-server:zabbix-server zabbix-$app_component:$os-$version
