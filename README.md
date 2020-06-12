# autoboard

Turn Prometheus alerts into Grafana dashboards.

## Features

- Read alert groups from a Prometheus server and create a dashboard per group.
- Detect the type of panel to create based on the query of an alert. Graph and Singlestat panels are supported.
- Configure a panel via annotations of the alert in Prometheus.
- Set threshold on Graph panel based on the query of the alert.
- Set threshold on Singlestat panel based on the query of the alert.

## Commands

### `alert`

Create a dashboard from [alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
in Prometheus.

### `drilldown`

Create a dashboard for all metrics exposed by a service at its scrape endpoint.

## Roadmap

The [.plan file](./.plan.md) contains ideas for new features and completed tasks.

## Known Caveats

### `[[` and `]]` in legend format

The format of the legend of a Graph panel can be configured via the annotation `atd_legend`.
Instead of setting `{{method}} - {{status}}` as required by Grafana, the value of that annotation needs to be `[[method]] - [[status]]`.
