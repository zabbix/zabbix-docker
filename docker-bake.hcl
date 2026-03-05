// =========================
// docker-bake.hcl (3-phase)
// =========================

// ----- Variables -----
variable "OS"                   {
   type        = string
   default     = "alpine"
   description = "Base Operating System for building images. Allowed: alpine|centos|ol|ubuntu"
}
variable "ZBX_VERSION"          {
   type        = string
   default     = "7.4"
   description = "Zabbix branch or exact version to build"
}
variable "OS_BASE_IMAGE"        {
   type        = string
   default     = "alpine:3.23"
   description = "Base image for images. Passed to Dockerfiles as OS_BASE_IMAGE"
}
variable "ZBX_IMAGE_TAG"        {
   type        = string
   default     = "${OS}-${ZBX_VERSION}-local"
   description = "Image tag for all images. For example, \"alpine-7.4-local\""
}
variable "PLATFORMS"            {
   type        = string
   default     = ""
   description = "Target platform list. For example, \"linux/amd64,linux/arm64\""
}
variable "ZBX_IMAGE_NAMESPACE"  {
   type        = string
   default     = ""
   description = "Zabbix registry and namespace. For example \"zabbix/\""
}
variable "ZBX_IMAGE_PREFIX"     {
   type = string
   default = "zabbix-"
   description = "Prefix for Zabbix images"
}

// ----- Groups (3-phase) -----
group "base" {
  targets = ["build-base"]
}

group "builder-mysql" {
  targets = ["build-mysql"]
}

group "builder-pgsql" {
  targets = ["build-pgsql"]
}

group "builders-sqlite3" {
  targets = ["build-sqlite3"]
}

group "runtime-mysql-all" {
  targets = [
    // common
    "agent-mysql",
    "agent2-mysql",
    "java-gateway-mysql",
    "web-service-mysql",

    "server-mysql",
    "web-nginx-mysql",
    "web-apache-mysql",

    // db-specific fixed components
    "proxy-mysql",

    // No dependencies from build images
    "snmptraps"
  ]
}

group "runtime-pgsql-all" {
  targets = [
    // common
    "agent-pgsql",
    "agent2-pgsql",
    "java-gateway-pgsql",
    "web-service-pgsql",

    "server-pgsql",
    "web-nginx-pgsql",
    "web-apache-pgsql",

    // No dependencies from build images
    "snmptraps"
  ]
}

group "runtime-mysql-minimal" {
  targets = [
    // common
    "agent-mysql",

    "server-mysql",
    "web-nginx-mysql"
  ]
}

group "runtime-pgsql-minimal" {
  targets = [
    // common
    "agent-pgsql",

    "server-pgsql",
    "web-nginx-pgsql"
  ]
}

group "runtime-mysql-apache" {
  targets = [
    // common
    "agent-mysql",

    "server-mysql",
    "web-apache-mysql"
  ]
}

group "runtime-pgsql-apache" {
  targets = [
    // common
    "agent-pgsql",

    "server-pgsql",
    "web-apache-pgsql"
  ]
}

group "runtime-sqlite3" {
  targets = [
    // common
    "agent-mysql",

    "proxy-sqlite3"
  ]
}

// Keep default minimal to avoid surprises
group "default" { targets = ["base"] }

// ----- Common templates -----
target "_common" {
  args = {
    OS_BASE_IMAGE         = OS_BASE_IMAGE
    BUILDKIT_INLINE_CACHE = "1"
  }
  platforms = notequal(PLATFORMS, "") ? split(",", replace(PLATFORMS, " ", "")) : null
}

target "_builder_common" {
  inherits = ["_common"]
  contexts = {
    config_templates = "config_templates",
    sources          = "sources"
  }
  dockerfile = "Dockerfile"
  args = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}zabbix-build-base:${ZBX_IMAGE_TAG}"
  }
}

// For runtime images that depend on DB-flavored builder env
target "_runtime_db_common_mysql" {
  inherits = ["_common"]
  contexts = {
    builder = "target:build-mysql"
  }
}

target "_runtime_db_common_pgsql" {
  inherits = ["_common"]
  contexts = {
    builder = "target:build-pgsql"
  }
}

// For runtime images that do not depend on DB builder
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

// "Common" runtime
target "agent-mysql" {
  description = "Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/agent/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent:${ZBX_IMAGE_TAG}"]
}

target "agent2-mysql" {
  description = "Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/agent2/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent2:${ZBX_IMAGE_TAG}"]
}

target "java-gateway-mysql" {
  description = "Zabbix Java Gateway performs native support for monitoring JMX applications"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/java-gateway/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}java-gateway:${ZBX_IMAGE_TAG}"]
}

target "web-service-mysql" {
  description = "Zabbix web service for performing various tasks using headless web browser"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/web-service/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-service:${ZBX_IMAGE_TAG}"]
}

target "agent-pgsql" {
  description = "Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/agent/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent:${ZBX_IMAGE_TAG}"]
}

target "agent2-pgsql" {
  description = "Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/agent2/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}agent2:${ZBX_IMAGE_TAG}"]
}

target "java-gateway-pgsql" {
  description = "Zabbix Java Gateway performs native support for monitoring JMX applications"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/java-gateway/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}java-gateway:${ZBX_IMAGE_TAG}"]
}

target "web-service-pgsql" {
  description = "Zabbix web service for performing various tasks using headless web browser"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/web-service/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-service:${ZBX_IMAGE_TAG}"]
}

// Server/Web choose subfolder by DB
target "server-mysql" {
  description = "Zabbix server with MySQL database support"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/server-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}server-mysql:${ZBX_IMAGE_TAG}"]
}

target "web-nginx-mysql" {
  description = "Zabbix web-interface based on Nginx web server with MySQL database support"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/web-nginx-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-nginx-mysql:${ZBX_IMAGE_TAG}"]
}

target "web-apache-mysql" {
  description = "Zabbix web-interface based on Apache web server with MySQL database support"
  inherits    = ["_runtime_db_common_mysql"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/web-apache-mysql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-apache-mysql:${ZBX_IMAGE_TAG}"]
}

target "server-pgsql" {
  description = "Zabbix server with PostgreSQL database support"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/server-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}server-pgsql:${ZBX_IMAGE_TAG}"]
}

target "web-nginx-pgsql" {
  description = "Zabbix web-interface based on Nginx web server with PostgreSQL database support"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/web-nginx-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-nginx-pgsql:${ZBX_IMAGE_TAG}"]
}

target "web-apache-pgsql" {
  description = "Zabbix web-interface based on Apache web server with PostgreSQL database support"
  inherits    = ["_runtime_db_common_pgsql"]
  depends_on  = ["build-pgsql"]
  context     = "Dockerfiles/web-apache-pgsql/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}web-apache-pgsql:${ZBX_IMAGE_TAG}"]
}

// =========================
// Runtime (fixed DB targets)
// =========================

// proxy-mysql: always uses mysql builder (independent of DB variable, depends on build-sqlite3 build image)
target "proxy-mysql" {
  description = "Zabbix proxy with MySQL database support"
  inherits    = ["_common"]
  depends_on  = ["build-mysql"]
  context     = "Dockerfiles/proxy-mysql/${OS}"
  args        = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-mysql:${ZBX_IMAGE_TAG}"
  }
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}proxy-mysql:${ZBX_IMAGE_TAG}"]
}

// proxy-sqlite3: always uses sqlite3 builder (independent of DB variable, depends on build-sqlite3 build image)
target "proxy-sqlite3" {
  description = "Zabbix proxy with SQLite3 database support"
  inherits    = ["_common"]
  depends_on  = ["build-sqlite3"]
  context     = "Dockerfiles/proxy-sqlite3/${OS}"
  args        = {
    BUILD_BASE_IMAGE = "${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}build-sqlite3:${ZBX_IMAGE_TAG}"
  }
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}proxy-sqlite3:${ZBX_IMAGE_TAG}"]
}

// =========================
// Runtime (no build images dependency)
// =========================
target "snmptraps" {
  description = "Zabbix SNMP traps receiver"
  inherits    = ["_runtime_nodb"]
  context     = "Dockerfiles/snmptraps/${OS}"
  tags        = ["${ZBX_IMAGE_NAMESPACE}${ZBX_IMAGE_PREFIX}snmptraps:${ZBX_IMAGE_TAG}"]
}
