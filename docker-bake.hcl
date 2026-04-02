// =========================
// docker-bake.hcl
// =========================

// ----- Variables -----
variable "OS" {
  type        = string
  default     = "alpine"
  description = "Base Operating System for building images. Allowed: alpine|centos|ol|ubuntu"
}

variable "ZBX_VERSION" {
  type        = string
  default     = "8.0"
  description = "Zabbix branch or exact version to build"
}

variable "OS_BASE_IMAGE" {
  type        = string
  default     = "alpine:3.23"
  description = "Base image for images. Passed to Dockerfiles as OS_BASE_IMAGE"
}

variable "ZBX_IMAGE_TAG" {
  type        = string
  default     = "${OS}-${ZBX_VERSION}-local"
  description = "Image tag for all images. For example, \"alpine-8.0-local\""
}

variable "PLATFORMS" {
  type        = string
  default     = ""
  description = "Target platform list. For example, \"linux/amd64,linux/arm64\""
}

variable "ZBX_IMAGE_NAMESPACE" {
  type        = string
  default     = ""
  description = "Zabbix registry and namespace. For example \"zabbix/\""
}

variable "ZBX_IMAGE_PREFIX" {
  type        = string
  default     = "zabbix-"
  description = "Prefix for Zabbix images"
}

// ----- Groups -----
group "base" {
  targets = ["build-base"]
}

group "builder-mysql" {
  targets = ["build-mysql"]
}

group "builder-pgsql" {
  targets = ["build-pgsql"]
}

group "builder-sqlite3" {
  targets = ["build-sqlite3"]
}

group "runtime-mysql-all" {
  targets = [
    "agent-mysql",
    "agent2-mysql",
    "java-gateway-mysql",
    "web-service-mysql",
    "server-mysql",
    "web-nginx-mysql",
    "web-apache-mysql",
    "proxy-mysql",
    "snmptraps",
  ]
}

group "runtime-pgsql-all" {
  targets = [
    "agent-pgsql",
    "agent2-pgsql",
    "java-gateway-pgsql",
    "web-service-pgsql",
    "server-pgsql",
    "web-nginx-pgsql",
    "web-apache-pgsql",
    "snmptraps",
  ]
}

group "runtime-mysql-minimal" {
  targets = [
    "agent-mysql",
    "server-mysql",
    "web-nginx-mysql",
  ]
}

group "runtime-pgsql-minimal" {
  targets = [
    "agent-pgsql",
    "server-pgsql",
    "web-nginx-pgsql",
  ]
}

group "runtime-mysql-apache" {
  targets = [
    "agent-mysql",
    "server-mysql",
    "web-apache-mysql",
  ]
}

group "runtime-pgsql-apache" {
  targets = [
    "agent-pgsql",
    "server-pgsql",
    "web-apache-pgsql",
  ]
}

group "runtime-sqlite3" {
  targets = [
    "agent-mysql",
    "proxy-sqlite3",
  ]
}

// Keep default minimal to avoid surprises
group "default" {
  targets = ["base"]
}

// ----- Common templates -----
target "_common" {
  args = {
    OS_BASE_IMAGE         = OS_BASE_IMAGE
    BUILDKIT_INLINE_CACHE = "1"
  }

  platforms = notequal(PLATFORMS, "") ? split(",", replace(PLATFORMS, " ", "")) : null
}

target "_builder_common" {
  inherits   = ["_common"]
  dockerfile = "Dockerfile"

  contexts = {
    config_templates = "templates/config"
    sources          = "sources"
  }

  args = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}zabbix-build-base:${ZBX_IMAGE_TAG}"
  }
}

target "_runtime_common" {
  inherits = ["_common"]

  contexts = {
    entrypoints = "templates/entrypoints"
  }
}

target "_runtime_mysql" {
  inherits   = ["_runtime_common"]
  depends_on = ["build-mysql"]

  contexts = {
    builder     = "target:build-mysql"
  }
}

target "_runtime_pgsql" {
  inherits   = ["_runtime_common"]
  depends_on = ["build-pgsql"]

  contexts = {
    builder     = "target:build-pgsql"
  }
}

target "_runtime_nodb" {
  inherits = ["_common"]
}

// =========================
// Phase 1: build-base
// =========================
target "build-base" {
  description = "Zabbix build base image contains all required packages to build Zabbix images"
  inherits    = ["_common"]
  context     = "Dockerfiles/build-base/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-base:${ZBX_IMAGE_TAG}"]
}

// =========================
// Phase 2: builders per DB
// =========================
target "build-mysql" {
  description = "Zabbix build base for MySQL based images"
  inherits    = ["_builder_common"]
  depends_on  = ["build-base"]
  context     = "Dockerfiles/build-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-mysql:${ZBX_IMAGE_TAG}"]
}

target "build-pgsql" {
  description = "Zabbix build base for PostgreSQL based images"
  inherits    = ["_builder_common"]
  depends_on  = ["build-base"]
  context     = "Dockerfiles/build-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-pgsql:${ZBX_IMAGE_TAG}"]
}

target "build-sqlite3" {
  description = "Zabbix build base for SQLite3 based images"
  inherits    = ["_builder_common"]
  depends_on  = ["build-base"]
  context     = "Dockerfiles/build-sqlite3/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-sqlite3:${ZBX_IMAGE_TAG}"]
}

// =========================
// Phase 3: runtime (DB-flavored)
// =========================
target "agent-mysql" {
  description = "Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/agent/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent:${ZBX_IMAGE_TAG}"]
}

target "agent2-mysql" {
  description = "Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/agent2/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent2:${ZBX_IMAGE_TAG}"]
}

target "java-gateway-mysql" {
  description = "Zabbix Java Gateway performs native support for monitoring JMX applications"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/java-gateway/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}java-gateway:${ZBX_IMAGE_TAG}"]
}

target "web-service-mysql" {
  description = "Zabbix web service for performing various tasks using headless web browser"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/web-service/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-service:${ZBX_IMAGE_TAG}"]
}

target "agent-pgsql" {
  description = "Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/agent/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent:${ZBX_IMAGE_TAG}"]
}

target "agent2-pgsql" {
  description = "Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/agent2/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent2:${ZBX_IMAGE_TAG}"]
}

target "java-gateway-pgsql" {
  description = "Zabbix Java Gateway performs native support for monitoring JMX applications"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/java-gateway/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}java-gateway:${ZBX_IMAGE_TAG}"]
}

target "web-service-pgsql" {
  description = "Zabbix web service for performing various tasks using headless web browser"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/web-service/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-service:${ZBX_IMAGE_TAG}"]
}

target "server-mysql" {
  description = "Zabbix server with MySQL database support"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/server-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}server-mysql:${ZBX_IMAGE_TAG}"]
}

target "web-nginx-mysql" {
  description = "Zabbix web-interface based on Nginx web server with MySQL database support"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/web-nginx-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-nginx-mysql:${ZBX_IMAGE_TAG}"]
}

target "web-apache-mysql" {
  description = "Zabbix web-interface based on Apache web server with MySQL database support"
  inherits    = ["_runtime_mysql"]
  context     = "Dockerfiles/web-apache-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-apache-mysql:${ZBX_IMAGE_TAG}"]
}

target "server-pgsql" {
  description = "Zabbix server with PostgreSQL database support"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/server-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}server-pgsql:${ZBX_IMAGE_TAG}"]
}

target "web-nginx-pgsql" {
  description = "Zabbix web-interface based on Nginx web server with PostgreSQL database support"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/web-nginx-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-nginx-pgsql:${ZBX_IMAGE_TAG}"]
}

target "web-apache-pgsql" {
  description = "Zabbix web-interface based on Apache web server with PostgreSQL database support"
  inherits    = ["_runtime_pgsql"]
  context     = "Dockerfiles/web-apache-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-apache-pgsql:${ZBX_IMAGE_TAG}"]
}

// =========================
// Runtime (fixed DB targets)
// =========================
target "proxy-mysql" {
  description = "Zabbix proxy with MySQL database support"
  inherits    = ["_runtime_common"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/proxy-mysql/${OS}"

  args = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-mysql:${ZBX_IMAGE_TAG}"
  }

  tags = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}proxy-mysql:${ZBX_IMAGE_TAG}"]
}

target "proxy-sqlite3" {
  description = "Zabbix proxy with SQLite3 database support"
  inherits    = ["_runtime_common"]
  depends_on  = ["build-sqlite3"]
  context     = "Dockerfiles/proxy-sqlite3/${OS}"

  args = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-sqlite3:${ZBX_IMAGE_TAG}"
  }

  tags = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}proxy-sqlite3:${ZBX_IMAGE_TAG}"]
}

// =========================
// Runtime (no build image dependency)
// =========================
target "snmptraps" {
  description = "Zabbix SNMP traps receiver"
  inherits    = ["_runtime_nodb"]
  context     = "Dockerfiles/snmptraps/${OS}"
  contexts = {
    config_templates = "templates/config/snmptraps"
    scripts          = "templates/scripts/snmptraps"
  }
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}snmptraps:${ZBX_IMAGE_TAG}"]
}
