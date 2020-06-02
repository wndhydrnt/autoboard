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

func ConvertAlertToPanel(alert pav1.AlertingRule, datasourceDef string) (r interface{}, err error) {
	expr, err := parser.ParseExpr(alert.Query)
	if err != nil {
		return r, fmt.Errorf("parse query expression: %w", err)
	}

	be, ok := expr.(*parser.BinaryExpr)
	if !ok {
		return r, fmt.Errorf("query is not a binary expression")
	}

	datasource := settingString(alert, "datasource", datasourceDef)
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
			Height:             panelHeight,
			Query:              escapeQuery(be.LHS.String()),
			ThresholdInvertNo:  be.Op == parser.GTR || be.Op == parser.GTE,
			ThresholdInvertYes: be.Op == parser.LSS || be.Op == parser.LTE,
			ThresholdValue:     be.RHS.String(),
			Width:              panelWidth,
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
			Height:             panelHeight,
			Query:              escapeQuery(be.RHS.String()),
			ThresholdInvertNo:  be.Op == parser.LSS || be.Op == parser.LTE,
			ThresholdInvertYes: be.Op == parser.GTR || be.Op == parser.GTE,
			ThresholdValue:     be.LHS.String(),
			Width:              panelWidth,
		}

		return ss, nil
	}

	g := Graph{
		Datasource: datasource,
		Format:     format,
		Height:     panelHeight,
		Legend:     strings.ReplaceAll(strings.ReplaceAll(settingString(alert, "legend", ""), "[[", "{{"), "]]", "}}"),
		Width:      panelWidth * 2,
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

func RunAlert(cfg config.Config, filters []*regexp.Regexp) error {
	SetPrefix(cfg.SettingsPrefix)
	log.SetLevel(cfg.LogLevel)
	promapi, err := NewPrometheusAPI(cfg.PrometheusAddress)
	if err != nil {
		return fmt.Errorf("init Prometheus API client: %w", err)
	}

	p := &Prometheus{
		DatasourceDefault: cfg.DatasourceDefault,
		Filters:           filters,
		PromAPI:           promapi,
	}
	boards, err := p.ReadAlerts()
	if err != nil {
		return fmt.Errorf("read alerts from Prometheus: %w", err)
	}

	r := &Renderer{
		dashboardTpl:  cfg.TemplateDashboard,
		graphTpl:      cfg.TemplateGraph,
		singlestatTpl: cfg.TemplateSinglestat,
	}
	gf := &Grafana{Address: cfg.GrafanaAddress}
	for _, d := range boards {
		s := r.Render(d)
		err := gf.CreateDashboard(s)
		if err != nil {
			return fmt.Errorf("create board %s: %w", d.Title, err)
		}

		log.Infof("board %s created", d.Title)
	}

	return nil
}
