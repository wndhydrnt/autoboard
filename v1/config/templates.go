package config

var dashboardTplDefault = `
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "links": [],
  "panels": [
{{{Graph}}}
{{#HasSinglestat}},{{/HasSinglestat}}
{{{SingleStat}}}
  ],
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "{{{Title}}}"
}
`

var graphTplDefault = `
{
  "aliasColors": {},
  "bars": false,
  "dashLength": 10,
  "dashes": false,
  "datasource": "{{{Datasource}}}",
  "description": "{{Description}}",
  "fill": 1,
  "fillGradient": 0,
  "gridPos": {
    "h": {{{Height}}},
    "w": {{{Width}}},
    "x": {{PosX}},
    "y": {{PosY}}
  },
  "hiddenSeries": false,
  "legend": {
    "avg": false,
    "current": false,
    "hideEmpty": true,
    "hideZero": true,
    "max": false,
    "min": false,
{{#HasLegend}}
    "show": true,
    "alignAsTable": true,
    "total": true,
    "values": true
{{/HasLegend}}
{{^HasLegend}}
    "show": false,
    "alignAsTable": false,
    "total": false,
    "values": false
{{/HasLegend}}
  },
  "lines": true,
  "linewidth": 1,
  "links": [],
  "nullPointMode": "null",
  "options": {
    "dataLinks": []
  },
  "percentage": false,
  "pointradius": 2,
  "points": false,
  "renderer": "flot",
  "seriesOverrides": [],
  "spaceLength": 10,
  "stack": false,
  "steppedLine": false,
  "targets": [
{{#Queries}}
    {
      "expr": "{{{Query}}}",
      "format": "time_series",
      "intervalFactor": 1,
      "legendFormat": "{{{Legend}}}",
      "refId": "A"
    }{{#HasMore}},{{/HasMore}}
{{/Queries}}
  ],
  "thresholds": [
{{#HasThreshold}}
    {
      "colorMode": "critical",
      "fill": true,
      "line": true,
      "op": "{{{ThresholdOP}}}",
      "value": {{{ThresholdValue}}},
      "yaxis": "left"
    }
{{/HasThreshold}}
  ],
  "timeFrom": null,
  "timeRegions": [],
  "timeShift": null,
  "title": "{{{Title}}}",
  "tooltip": {
    "shared": true,
    "sort": 2,
    "value_type": "individual"
  },
  "type": "graph",
  "xaxis": {
    "buckets": null,
    "mode": "time",
    "name": null,
    "show": true,
    "values": []
  },
  "yaxes": [
    {
      "format": "{{Format}}",
      "label": null,
      "logBase": 1,
      "max": null,
      "min": null,
      "show": true
    },
    {
      "format": "short",
      "label": null,
      "logBase": 1,
      "max": null,
      "min": null,
      "show": true
    }
  ],
  "yaxis": {
    "align": false,
    "alignLevel": null
  }
}
`

var singlestatTplDefault = `
{
  "cacheTimeout": null,
  "colorBackground": false,
  "colorValue": true,
  "colors": [
{{#ThresholdInvertYes}}
    "#d44a3a","rgba(237, 129, 40, 0.89)","#299c46"
{{/ThresholdInvertYes}}
{{#ThresholdInvertNo}}
    "#299c46","rgba(237, 129, 40, 0.89)","#d44a3a"
{{/ThresholdInvertNo}}
  ],
  "datasource": "{{{Datasource}}}",
  "description": "{{Description}}",
  "format": "{{{Format}}}",
  "gauge": {
    "maxValue": 100,
    "minValue": 0,
    "show": false,
    "thresholdLabels": false,
    "thresholdMarkers": true
  },
  "gridPos": {
    "h": {{{Height}}},
    "w": {{{Width}}},
    "x": {{PosX}},
    "y": {{PosY}}
  },
  "interval": null,
  "links": [],
  "mappingType": 1,
  "mappingTypes": [
    {
      "name": "value to text",
      "value": 1
    },
    {
      "name": "range to text",
      "value": 2
    }
  ],
  "maxDataPoints": 100,
  "nullPointMode": "connected",
  "nullText": null,
  "options": {},
  "postfix": "",
  "postfixFontSize": "50%",
  "prefix": "",
  "prefixFontSize": "50%",
  "rangeMaps": [
    {
      "from": "null",
      "text": "N/A",
      "to": "null"
    }
  ],
  "sparkline": {
    "fillColor": "rgba(31, 118, 189, 0.18)",
    "full": false,
    "lineColor": "rgb(31, 120, 193)",
    "show": false
  },
  "tableColumn": "",
  "targets": [
    {
      "expr": "{{{Query}}}",
      "format": "time_series",
      "instant": true,
      "intervalFactor": 1,
      "legendFormat": "{{{Legend}}}",
      "refId": "A"
    }
  ],
  "thresholds": "{{{ThresholdValue}}},{{{ThresholdValue}}}",
  "timeFrom": null,
  "timeShift": null,
  "title": "{{{Title}}}",
  "type": "singlestat",
  "valueFontSize": "80%",
  "valueMaps": [
    {
      "op": "=",
      "text": "N/A",
      "value": "null"
    }
  ],
  "valueName": "{{{ValueName}}}"
}
`
