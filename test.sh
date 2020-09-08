TAGS_ARRAY=()

IMAGE_NAME="zabbix/zabbix-agent"
RELEASE_VERSION="refs/tags/5.0.3"
RELEASE_VERSION=${RELEASE_VERSION:10}

GIT_BRANCH=${RELEASE_VERSION%.*}
echo "::debug::Release version ${RELEASE_VERSION}. Branch ${GIT_BRANCH}"
TAGS_ARRAY+=("$IMAGE_NAME:alpine-${RELEASE_VERSION}")

if [ "alpine" == "alpine" ] && [ "${LATEST_BRANCH}" == "${GIT_BRANCH}" ]; then
    TAGS_ARRAY+=("$IMAGE_NAME:latest")
  fi
TAGS=$(printf -- "--tag %s " "${TAGS_ARRAY[@]}")

echo $TAGS