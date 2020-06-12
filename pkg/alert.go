package v1

import (
	"fmt"
	"regexp"
	"strings"

	pav1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/prometheus/promql/parser"
	log "github.com/sirupsen/logrus"
	"github.com/wndhydrnt/autoboard/pkg/config"
)

// ConvertAlertToPanel takes a Prometheus Alerting Rule as input and converts it to a Panel.
// A query that aggregates a value, e.g. by using sum() without any "by" clause, is converted to a Singlestat panel.
// A query is converted to a Graph panel otherwise.
// A threshold is set for Singlestat and Graph panels if one side of the query is a scalar value.
func ConvertAlertToPanel(alert pav1.AlertingRule, datasource string) (r interface{}, err error) {
	expr, err := parser.ParseExpr(alert.Query)
	if err != nil {
		return r, fmt.Errorf("parse query expression: %w", err)
	}

	be, ok := expr.(*parser.BinaryExpr)
	if !ok {
		return r, fmt.Errorf("query is not a binary expression")
	}

	format := settingString(alert, "format", defaultFormat)
	lhsHasGrouping := false
	lhsAggr, lhsIsAggregate := be.LHS.(*parser.AggregateExpr)
	if lhsIsAggregate {
		lhsHasGrouping = len(lhsAggr.Grouping) > 0
	}

	if lhsIsAggregate && be.RHS.Type() == parser.ValueTypeScalar && !lhsHasGrouping {
		ss := Singlestat{
			Datasource:         datasource,
			Format:             format,
			Query:              escapeQuery(be.LHS.String()),
			ThresholdInvertNo:  be.Op == parser.GTR || be.Op == parser.GTE,
			ThresholdInvertYes: be.Op == parser.LSS || be.Op == parser.LTE,
			ThresholdValue:     be.RHS.String(),
		}

		return ss, nil
	}

	rhsHasGrouping := false
	rhsAggr, rhsIsAggregate := be.RHS.(*parser.AggregateExpr)
	if rhsIsAggregate {
		rhsHasGrouping = len(rhsAggr.Grouping) > 0
	}

	if rhsIsAggregate && be.LHS.Type() == parser.ValueTypeScalar && !rhsHasGrouping {
		ss := Singlestat{
			Datasource:         datasource,
			Format:             format,
			Query:              escapeQuery(be.RHS.String()),
			ThresholdInvertNo:  be.Op == parser.LSS || be.Op == parser.LTE,
			ThresholdInvertYes: be.Op == parser.GTR || be.Op == parser.GTE,
			ThresholdValue:     be.LHS.String(),
		}

		return ss, nil
	}

	g := Graph{
		Datasource: datasource,
		Format:     format,
		Legend:     strings.ReplaceAll(strings.ReplaceAll(settingString(alert, "legend", ""), "[[", "{{"), "]]", "}}"),
	}
	g.HasLegend = g.Legend != ""
	gq := []GraphQuery{}
	if be.LHS.Type() == parser.ValueTypeScalar {
		if be.Op == parser.LSS || be.Op == parser.LTE {
			g.HasThreshold = true
			g.ThresholdOP = "lt"
			g.ThresholdValue = be.LHS.String()
		}

		if be.Op == parser.GTR || be.Op == parser.GTE {
			g.HasThreshold = true
			g.ThresholdOP = "gt"
			g.ThresholdValue = be.LHS.String()
		}
	} else {
		gq = append(gq, GraphQuery{Query: escapeQuery(be.LHS.String())})
	}

	if be.RHS.Type() == parser.ValueTypeScalar {
		if be.Op == parser.LSS || be.Op == parser.LTE {
			g.HasThreshold = true
			g.ThresholdOP = "lt"
			g.ThresholdValue = be.RHS.String()
		}

		if be.Op == parser.GTR || be.Op == parser.GTE {
			g.HasThreshold = true
			g.ThresholdOP = "gt"
			g.ThresholdValue = be.RHS.String()
		}
	} else {
		gq = append(gq, GraphQuery{Query: escapeQuery(be.RHS.String())})

		if len(gq) == 2 {
			gq[0].HasMore = true
		}
	}

	g.Queries = gq
	return g, nil
}

func escapeQuery(q string) string {
	return strings.ReplaceAll(q, `"`, `\"`)
}

// RunAlert is the entrypoint to create a dashboard from an alert.
func RunAlert(cfg config.Config, filters []*regexp.Regexp, promAddr string, settingPrefix string) error {
	SetPrefix(settingPrefix)
	log.SetLevel(cfg.LogLevel)
	promapi, err := NewPrometheusAPI(promAddr)
	if err != nil {
		return fmt.Errorf("init Prometheus API client: %w", err)
	}

	p := &Prometheus{
		DatasourceDefault: cfg.Datasource,
		Filters:           filters,
		PromAPI:           promapi,
	}
	alerts, err := p.ReadAlerts()
	if err != nil {
		return fmt.Errorf("read alerts from Prometheus: %w", err)
	}

	r := &Renderer{
		dashboardTpl:         cfg.TemplateDashboard,
		graphTpl:             cfg.TemplateGraph,
		panelHeight:          cfg.GrafanaPanelsHeight,
		panelWidthGraph:      cfg.GrafanaPanelsGraphWidth,
		panelWidthSinglestat: cfg.GrafanaPanelsSinglestatWidth,
		singlestatTpl:        cfg.TemplateSinglestat,
	}
	gf := &Grafana{Address: cfg.GrafanaAddress}
	for _, a := range alerts {
		s := r.Render(a.Dashboard, a.Panels)
		err := gf.CreateDashboard(s, cfg.GrafanaFolder)
		if err != nil {
			return fmt.Errorf("create board %s: %w", a.Dashboard.Title, err)
		}

		log.Infof("board %s created", a.Dashboard.Title)
	}

	return nil
}
