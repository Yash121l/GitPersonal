#!/bin/sh
set -eu

mkdir -p /data/repos /data/ssh /app
chown -R forge:forge /data /app

exec gosu forge /usr/local/bin/forge "$@"
