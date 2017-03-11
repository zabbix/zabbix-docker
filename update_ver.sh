#!/bin/bash

set -e

new_version=$1
is_trunk=${2:-"false"}

if [[ $new_version =~ ^[0-9]*\.[0-9]*.*$ ]] && [ "$is_trunk" == "true" ]; then
    echo "** Switching to trunk"
elif [[ ! $new_version =~ ^[0-9]*\.[0-9]*\.[0-9]*.*$ ]] && [ "$new_version" != "master" ]; then
    echo "Incorrect syntax of the version"
    exit 1
fi

major_version=$(echo $new_version | cut -d'.' -f 1,2)
minor_version=$(echo $new_version | cut -d'.' -f 3)

echo "Using $major_version.$minor_version version of Zabbix"

find ./ -type d | while read DIR; do
    if [ -d "$DIR" ] && [ -f "$DIR/run_zabbix_component.sh" ] && [ -f "$DIR/Dockerfile" ]; then
        echo "** Updating $DIR/Dockerfile"
        if [ "$is_trunk" == "true" ]; then
            sed -i -e "/^ARG MAJOR_VERSION=/s/=.*/=$major_version/" $DIR/Dockerfile
            sed -i -e "/^ARG ZBX_VERSION=/s/=.*/=\${MAJOR_VERSION}/" $DIR/Dockerfile
            sed -i -e "/^ARG ZBX_SOURCES=/s/=.*/=svn:\/\/svn.zabbix.com\/trunk\//" $DIR/Dockerfile
        else
            sed -i -e "/^ARG MAJOR_VERSION=/s/=.*/=$major_version/" $DIR/Dockerfile
            [ "$new_version" == "master" ] || sed -i -e "/^ARG ZBX_VERSION=/s/=.*/=\${MAJOR_VERSION}.$minor_version/" $DIR/Dockerfile
            [ "$new_version" == "master" ] && sed -i -e "/^ARG ZBX_VERSION=/s/=.*/=\${MAJOR_VERSION}/" $DIR/Dockerfile
            sed -i -e "/^ARG ZBX_SOURCES=/s/=.*/=svn:\/\/svn.zabbix.com\/tags\/\${ZBX_VERSION}\//" $DIR/Dockerfile
        fi
    fi
done

exit 0
