package config

var defaultConfig = `
grafana:
  # Address of the Grafana server.
  # Env Var: AB_GRAFANA_ADDRESS=http://localhost:3000
  address: http://localhost:3000
  # Datasource to set in Grafana dashboards.
  # Env Var: AB_GRAFANA_DATASOURCE_DEFAULT=Prometheus
  datasource_default: ""

log:
  # Log level of the application.
  # Env Var: AB_LOG_LEVEL=error
  level: error

prometheus:
  # Address of the Prometheus server.
  # Env Var: AB_PROMETHEUS_ADDRESS=http://localhost:9090
  address: http://localhost:9090
  # Prefix used to detect annotations on an alert that configure a panel.
  # Env Var: AB_PROMETHEUS_SETTINGS_PREFIX=http://localhost:9090
  settings_prefix: ab_

templates:
  dashboard: ""
  graph: ""
  singlestat: ""
`