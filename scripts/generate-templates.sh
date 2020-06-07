#! /usr/bin/env bash

set -e

dashboard=$(cat templates/dashboard.json.mustache)
graph=$(cat templates/graph.json.mustache)
row=$(cat templates/row.json.mustache)
singlestat=$(cat templates/singlestat.json.mustache)

cat << EOF > pkg/config/templates.go
package config

var dashboardTplDefault = \`
${dashboard}
\`

var graphTplDefault = \`
${graph}
\`

var rowTplDefault = \`
${row}
\`

var singlestatTplDefault = \`
${singlestat}
\`
EOF
