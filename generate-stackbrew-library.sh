#!/usr/bin/env bash

set -eu

declare -a -r versions=( 3.0.28 4.0.11 4.2.5 )

self="${BASH_SOURCE##*/}"

# get the most recent commit which modified any of "$@"
fileCommit() {
        git log -1 --format='format:%H' HEAD -- "$@"
}

# prints "$2$1$3$1...$N"
join() {
    local sep="$1"; shift
    local out; printf -v out "${sep//%/%%}%s" "$@"
    echo "${out#$sep}"
}

cat <<-EOH
# this file is generated via https://github.com/zabbix/zabbix-docker/blob/$(fileCommit "$self")/$self

Maintainers: Alexey Pustovalov <alexey.pustovalov@zabbix.com> (@dotneft)
GitRepo: https://github.com/zabbix/zabbix-docker.git
EOH

for version in "${versions[@]}"; do
    major_version=${version%.*}
    commit=`git rev-list -n 1 "$version"`

    for component in agent java-gateway proxy-{mysql,sqlite3} server-{mysql,pgsql} web-{apache,nginx}-{mysql,pgsql}; do
        for variant in alpine; do
		dir="${component}/${variant}"
		[ -f "$dir/Dockerfile" ] || continue

		variantArches=( amd64 )

		echo
		cat <<-EOE
			Tags: $version-$component, $major_version-$component
			Architectures: $(join ', ' "${variantArches[@]}")
			GitCommit: $commit
			GitFetch: refs/heads/$major_version
			Directory: $dir
		EOE
        done
    done
done
