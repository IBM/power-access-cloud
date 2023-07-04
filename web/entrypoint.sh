#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

# Generates the env-config.js file
/usr/share/nginx/html/env.sh

# check for REACT_APP_PAC_GO_SERVER_TARGET environment variable set or empty
if [ -z "${REACT_APP_PAC_GO_SERVER_TARGET}" ]; then
  echo "REACT_APP_PAC_GO_SERVER_TARGET environment variable not set.."
  exit 1
fi

envsubst '$REACT_APP_PAC_GO_SERVER_TARGET' < /etc/nginx/conf.d/pac.conf.template > /etc/nginx/conf.d/default.conf
nginx -g "daemon off;"
