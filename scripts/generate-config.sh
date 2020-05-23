#! /usr/bin/env bash

set -e

config=$(cat v1/config/config.yaml)

cat << EOF > v1/config/default.go
package config

var defaultConfig = \`
${config}
\`
EOF
