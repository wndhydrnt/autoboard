# autoboard

Turn Prometheus alerts into Grafana dashboards.

## Features

- Read alert groups from a Prometheus server and create a dashboard per group.
- Detect the type of panel to create based on the query of an alert. Graph and Singlestat panels are supported.
- Configure a panel via annotations of the alert in Prometheus.
- Set threshold on Graph panel based on the query of the alert.
- Set threshold on Singlestat panel based on the query of the alert.

## Usage

## Roadmap

- Update a dashboard only if it changes.
- Improve detection of more complex queries.

## Known Caveats

### `[[` and `]]` in legend format

The format of the legend of a Graph panel can be configured via the annotation `atd_legend`.
Instead of setting `{{method}} - {{status}}` as required by Grafana, the value of that annotation needs to be `[[method]] - [[status]]`.
