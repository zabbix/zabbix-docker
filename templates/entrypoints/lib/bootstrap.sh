: "${ENTRYPOINT_LIBS:=/usr/lib/docker-entrypoint}"

# Base helpers used by most entrypoints
source "${ENTRYPOINT_LIBS}/debug.sh"
source "${ENTRYPOINT_LIBS}/logging.sh"
source "${ENTRYPOINT_LIBS}/format.sh"
source "${ENTRYPOINT_LIBS}/config.sh"
source "${ENTRYPOINT_LIBS}/clear_env.sh"
