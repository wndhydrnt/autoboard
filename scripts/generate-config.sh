#! /usr/bin/env bash

set -e

config=$(cat pkg/config/config.yaml)

cat << EOF > pkg/config/default.go
package config

var defaultConfig = \`
${config}
\`
EOF
