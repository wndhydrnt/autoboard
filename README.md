# autoboard

Automatically generate dashboards from Prometheus resources.

## Features

- Create a dashboard from an alert group in Prometheus.
- Create a dashboard from a metrics endpoint exposed by any service that supports Prometheus.
- Detect the type of panel to create based on the query of an alert or the metric type.
- Group panels into rows.
- Configure a panel via annotations of the alert in Prometheus.
- Set thresholds on panels based on the query of the alert.

## Commands

### `alert`

Create a dashboard from [alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
in Prometheus.

Usage: `autoboard alert -h`

### `drilldown`

Create a dashboard for all metrics exposed by a service at its scrape endpoint.

Usage: `autoboard drilldown -h`

## Roadmap

The [.plan file](./.plan.md) contains ideas for new features and completed tasks.

## Known Caveats

### `[[` and `]]` in legend format

The format of the legend of a Graph panel can be configured via the annotation `atd_legend`.
Instead of setting `{{method}} - {{status}}` as required by Grafana, the value of that annotation needs to be `[[method]] - [[status]]`.
