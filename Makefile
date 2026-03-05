# =========================
# Zabbix images build + run helper (bake v3-phase)
# =========================
.DEFAULT_GOAL := up

COMPOSE_PROFILES ?=

# -------- User-facing knobs (override via CLI) --------
OS ?= alpine
DB ?= mysql
ZBX_VERSION ?= 7.4

# Remote defaults (Official image registry)
REMOTE_IMAGE_PREFIX ?= zabbix/
REMOTE_ZBX_TAG ?= $(OS)-$(ZBX_VERSION)-latest

# Local defaults (what bake builds)
LOCAL_IMAGE_PREFIX ?=
LOCAL_ZBX_TAG ?= $(OS)-$(ZBX_VERSION)-local

# Optional multi-arch platforms: "linux/amd64,linux/arm64"
PLATFORMS ?=

# ---- Base images per OS ----
ALPINE_BASE_IMAGE ?= alpine:3.23
CENTOS_BASE_IMAGE ?= quay.io/centos/centos:stream10-minimal
OL_BASE_IMAGE ?= container-registry.oracle.com/os/oraclelinux:10-slim
UBUNTU_BASE_IMAGE ?= ubuntu:noble
RHEL_BASE_IMAGE ?= registry.access.redhat.com/ubi10/ubi-minimal:10.1

# Auto-select base image by OS (unless explicitly overridden)
ifeq ($(origin OS_BASE_IMAGE), undefined)
  ifeq ($(OS),alpine)
        OS_BASE_IMAGE := $(ALPINE_BASE_IMAGE)
  else ifeq ($(OS),centos)
        OS_BASE_IMAGE := $(CENTOS_BASE_IMAGE)
  else ifeq ($(OS),ol)
        OS_BASE_IMAGE := $(OL_BASE_IMAGE)
  else ifeq ($(OS),ubuntu)
        OS_BASE_IMAGE := $(UBUNTU_BASE_IMAGE)
  else ifeq ($(OS),rhel)
        OS_BASE_IMAGE := $(RHEL_BASE_IMAGE)
  else
        $(error Unsupported OS: $(OS))
  endif
endif

# Compose
COMPOSE ?= docker compose
ENV_FILE ?= .env

# -------- Bake group names (as in docker-bake.hcl) --------
BAKE_BASE_GROUP              := base
BAKE_BUILDERS_MYSQL_GROUP    := builder-mysql
BAKE_BUILDERS_PGSQL_GROUP    := builder-pgsql
BAKE_BUILDERS_SQLITE3_GROUP  := builder-sqlite3

BAKE_RUNTIME_MYSQL_ALL       := runtime-mysql-all
BAKE_RUNTIME_MYSQL_MINIMAL   := runtime-mysql-minimal
BAKE_RUNTIME_PGSQL_ALL       := runtime-pgsql-all
BAKE_RUNTIME_PGSQL_MINIMAL   := runtime-pgsql-minimal

# Export for sub-make / shells
export OS DB MAJOR_VERSION ZBX_VERSION OS_BASE_IMAGE PLATFORMS
export LOCAL_IMAGE_PREFIX LOCAL_ZBX_TAG REMOTE_IMAGE_PREFIX REMOTE_ZBX_TAG
export COMPOSE_PROFILES

# Pick compose file based on DB
ifeq ($(DB),mysql)
  COMPOSE_FILE := compose.yaml
else ifeq ($(DB),pgsql)
  COMPOSE_FILE := compose_pgsql.yaml
else
  COMPOSE_FILE := compose.yaml
endif

# Compose env passthrough: ensure compose sees the same variables
# (useful if compose uses ${OS}/${DB}/${IMAGE_TAG}/${IMAGE_PREFIX})
define compose_env
OS="$(OS)" DB="$(DB)" ZBX_IMAGE_TAG="$(1)" ZBX_IMAGE_REGISTRY="$(2)" ZBX_IMAGE_NAMESPACE="" $(COMPOSE) -f "$(COMPOSE_FILE)" --env-file "$(ENV_FILE)"
endef

# Bake env passthrough (match docker-bake.hcl variable names!)
define bake_env
OS="$(OS)" \
ZBX_VERSION="$(ZBX_VERSION)" \
OS_BASE_IMAGE="$(OS_BASE_IMAGE)" \
ZBX_IMAGE_TAG="$(LOCAL_ZBX_TAG)" \
PLATFORMS="$(PLATFORMS)" \
ZBX_IMAGE_NAMESPACE="$(LOCAL_IMAGE_PREFIX)" \
docker buildx bake
endef

# Choose runtime groups by DB
ifeq ($(DB),mysql)
  BAKE_RUNTIME_ALL_GROUP     := $(BAKE_RUNTIME_MYSQL_ALL)
  BAKE_RUNTIME_MINIMAL_GROUP := $(BAKE_RUNTIME_MYSQL_MINIMAL)
  BAKE_BUILDERS_GROUP        := $(BAKE_BUILDERS_MYSQL_GROUP)
else ifeq ($(DB),pgsql)
  BAKE_RUNTIME_ALL_GROUP     := $(BAKE_RUNTIME_PGSQL_ALL)
  BAKE_RUNTIME_MINIMAL_GROUP := $(BAKE_RUNTIME_PGSQL_MINIMAL)
  BAKE_BUILDERS_GROUP        := $(BAKE_BUILDERS_PGSQL_GROUP)
else ifeq ($(DB),sqlite3)
  # No sqlite3 runtime groups listed in your target table; keep builders only.
  BAKE_BUILDERS_GROUP        := $(BAKE_BUILDERS_SQLITE3_GROUP)
  BAKE_RUNTIME_ALL_GROUP     :=
  BAKE_RUNTIME_MINIMAL_GROUP :=
endif

# ---- Guards ----
check-rhel-host:
	@if [ "$(OS)" = "rhel" ]; then \
	  if [ -r /etc/os-release ]; then \
	    . /etc/os-release; HOST_ID="$$ID"; \
	  else \
	    HOST_ID="$$(uname -s | tr A-Z a-z)"; \
	  fi; \
	  if [ "$$HOST_ID" != "rhel" ]; then \
	    echo "ERROR: Refusing to build Red Hat images on host '$$HOST_ID'."; \
	    echo "This build requires RHEL subscription repositories. Run the build on a Red Hat host."; \
	    exit 1; \
	  fi; \
	fi

# -------- Targets --------
.PHONY: help print-vars \
	base builders-mysql builders-pgsql builders-sqlite3 builders \
	runtime-mysql-all runtime-mysql-minimal runtime-pgsql-all runtime-pgsql-minimal \
	runtime-all runtime-minimal runtime \
	build build-all \
	up up-local down restart logs ps \
	bake-target clean

help:
	@echo "Usage:"
	@echo "  make base                       # build build-base"
	@echo "  make builders                    # build builders for DB=$(DB) (mysql/pgsql/sqlite3)"
	@echo "  make runtime-minimal             # build runtime-<db>-minimal (mysql/pgsql only)"
	@echo "  make runtime-all                 # build runtime-<db>-all (mysql/pgsql only)"
	@echo "  make build                       # base + builders + runtime-minimal (mysql/pgsql only)"
	@echo "  make build-all                   # base + builders + runtime-all (mysql/pgsql only)"
	@echo "  make bake-target TARGET=server-mysql   # build a single bake target by name"
	@echo ""
	@echo "Compose:"
	@echo "  make up                          # pull+up using REMOTE_* images"
	@echo "  make up-local                    # build (minimal) then up using LOCAL_* images"
	@echo ""
	@echo "Common overrides:"
	@echo "  make build DB=mysql"
	@echo "  make build DB=pgsql"
	@echo "  make builders DB=sqlite3"
	@echo "  make build OS=ubuntu OS_BASE_IMAGE=ubuntu:noble"
	@echo "  make build PLATFORMS=linux/amd64,linux/arm64"
	@echo "  make build LOCAL_IMAGE_PREFIX=ghcr.io/zabbix/"
	@echo ""
	@echo "Current config:"
	@$(MAKE) --no-print-directory print-vars

print-vars:
	@echo "OS=$(OS)"
	@echo "OS_BASE_IMAGE=$(OS_BASE_IMAGE)"
	@echo "DB=$(DB)"
	@echo "MAJOR_VERSION=$(MAJOR_VERSION)"
	@echo "ZBX_VERSION=$(ZBX_VERSION)"
	@echo "LOCAL_ZBX_TAG=$(LOCAL_ZBX_TAG)"
	@echo "LOCAL_IMAGE_PREFIX=$(LOCAL_IMAGE_PREFIX)"
	@echo "PLATFORMS=$(PLATFORMS)"
	@echo "ENV_FILE=$(ENV_FILE)"
	@echo "COMPOSE_FILE=$(COMPOSE_FILE)"
	@echo "BAKE_BUILDERS_GROUP=$(BAKE_BUILDERS_GROUP)"
	@echo "BAKE_RUNTIME_MINIMAL_GROUP=$(BAKE_RUNTIME_MINIMAL_GROUP)"
	@echo "BAKE_RUNTIME_ALL_GROUP=$(BAKE_RUNTIME_ALL_GROUP)"

# ---- Bake groups ----
base: check-rhel-host
	@echo "==> Bake group: $(BAKE_BASE_GROUP) (OS=$(OS))"
	@$(bake_env) $(BAKE_BASE_GROUP)

builders-mysql: check-rhel-host
	@echo "==> Bake group: $(BAKE_BUILDERS_MYSQL_GROUP) (OS=$(OS))"
	@$(bake_env) $(BAKE_BUILDERS_MYSQL_GROUP)

builders-pgsql: check-rhel-host
	@echo "==> Bake group: $(BAKE_BUILDERS_PGSQL_GROUP) (OS=$(OS))"
	@$(bake_env) $(BAKE_BUILDERS_PGSQL_GROUP)

builders-sqlite3: check-rhel-host
	@echo "==> Bake group: $(BAKE_BUILDERS_SQLITE3_GROUP) (OS=$(OS))"
	@$(bake_env) $(BAKE_BUILDERS_SQLITE3_GROUP)

builders: check-rhel-host
	@echo "==> Bake group: $(BAKE_BUILDERS_GROUP) (DB=$(DB), OS=$(OS))"
	@$(bake_env) $(BAKE_BUILDERS_GROUP)

runtime-mysql-all: check-rhel-host
	@echo "==> Bake group: $(BAKE_RUNTIME_MYSQL_ALL) (OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_MYSQL_ALL)

runtime-mysql-minimal: check-rhel-host
	@echo "==> Bake group: $(BAKE_RUNTIME_MYSQL_MINIMAL) (OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_MYSQL_MINIMAL)

runtime-pgsql-all: check-rhel-host
	@echo "==> Bake group: $(BAKE_RUNTIME_PGSQL_ALL) (OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_PGSQL_ALL)

runtime-pgsql-minimal: check-rhel-host
	@echo "==> Bake group: $(BAKE_RUNTIME_PGSQL_MINIMAL) (OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_PGSQL_MINIMAL)

runtime-all: check-rhel-host
	@if [ -z "$(BAKE_RUNTIME_ALL_GROUP)" ]; then \
	  echo "ERROR: runtime-all is not defined for DB=$(DB) (no runtime groups listed for sqlite3)."; \
	  exit 1; \
	fi
	@echo "==> Bake group: $(BAKE_RUNTIME_ALL_GROUP) (DB=$(DB), OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_ALL_GROUP)

runtime-minimal: check-rhel-host
	@if [ -z "$(BAKE_RUNTIME_MINIMAL_GROUP)" ]; then \
	  echo "ERROR: runtime-minimal is not defined for DB=$(DB) (no runtime groups listed for sqlite3)."; \
	  exit 1; \
	fi
	@echo "==> Bake group: $(BAKE_RUNTIME_MINIMAL_GROUP) (DB=$(DB), OS=$(OS))"
	@$(bake_env) $(BAKE_RUNTIME_MINIMAL_GROUP)

# Alias
runtime: runtime-minimal

# Convenience: full build
build: base builders runtime-minimal
build-all: base builders runtime-all

# Build a single bake target by name (not group)
bake-target: check-rhel-host
	@if [ -z "$(TARGET)" ]; then \
	  echo "ERROR: TARGET is required. Example: make bake-target TARGET=server-mysql"; \
	  exit 1; \
	fi
	@echo "==> Bake target: $(TARGET) (OS=$(OS), DB=$(DB), local tag=$(LOCAL_ZBX_TAG))"
	@$(bake_env) "$(TARGET)"

# ---- Compose helpers ----
up:
	@$(call compose_env,$(REMOTE_ZBX_TAG),$(REMOTE_IMAGE_PREFIX)) pull --ignore-pull-failures
	@$(call compose_env,$(REMOTE_ZBX_TAG),$(REMOTE_IMAGE_PREFIX)) up -d --pull always

up-local:
	@$(MAKE) --no-print-directory build
	@echo "==> up-local (local) tag=$(LOCAL_ZBX_TAG) prefix=$(LOCAL_IMAGE_PREFIX)"
	@$(call compose_env,$(LOCAL_ZBX_TAG),$(LOCAL_IMAGE_PREFIX)) up -d

# ---- Compose command sets ----
COMPOSE_CMDS := pull down ps logs config restart start stop

define compose_remote
    @$(call compose_env,$(REMOTE_ZBX_TAG),$(REMOTE_IMAGE_PREFIX)) $(1)
endef

define compose_local
    @$(call compose_env,$(LOCAL_ZBX_TAG),$(LOCAL_IMAGE_PREFIX)) $(1)
endef

.PHONY: $(COMPOSE_CMDS) l-$(COMPOSE_CMDS)

$(COMPOSE_CMDS):
	$(call compose_remote,$@ $(ARGS))

l-%:
	$(call compose_local,$* $(ARGS))

# ---- Cleanup ----
clean:
	@echo "==> Removing local images for OS=$(OS) tag=$(LOCAL_ZBX_TAG) (best-effort)"
	@docker image rm -f \
	  "$(LOCAL_IMAGE_PREFIX)build-base:$(LOCAL_ZBX_TAG)" \
	  "$(LOCAL_IMAGE_PREFIX)build-mysql:$(LOCAL_ZBX_TAG)" \
	  "$(LOCAL_IMAGE_PREFIX)build-pgsql:$(LOCAL_ZBX_TAG)" \
	  "$(LOCAL_IMAGE_PREFIX)build-sqlite3:$(LOCAL_ZBX_TAG)" \
	  2>/dev/null || true
