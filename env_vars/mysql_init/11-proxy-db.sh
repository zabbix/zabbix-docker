#!/usr/bin/env bash
set -e

docker_process_sql --database=mysql <<-EOSQL
CREATE DATABASE IF NOT EXISTS \`${ZBX_PROXY_DB_NAME}\`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_bin;
EOSQL
