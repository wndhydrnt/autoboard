package v1

import (
	"fmt"
	"math"
	"strings"

	"github.com/hoisie/mustache"
	pav1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/prometheus/promql/parser"
)

type Dashboard struct {
	Title       string
	SingleStat  string
	SingleStats []Singlestat
	Graph       string
	Graphs      []Graph
}

type Panel interface{}

type Graph struct {
	Datasource     string
	Format         string
	HasLegend      bool
	HasThreshold   bool
	Height         int
	Legend         string
	Queries        []GraphQuery
	PosX           int
	PosY           int
	ThresholdOP    string
	ThresholdValue string
	Title          string
	Width          int
}

type GraphQuery struct {
	Code    string
	HasMore bool
	Query   string
	RefID   string
}

type Singlestat struct {
	Datasource         string
	Format             string
	Height             int
	Query              string
	PosX               int
	PosY               int
	ThresholdInvertNo  bool
	ThresholdInvertYes bool
	ThresholdValue     string
	Title              string
	Width              int
}

const (
	defaultFormat          = "short"
	graphPanelsPerRow      = 2
	panelHeight            = 4
	panelWidth             = 6
	singleStatPanelsPerRow = 4
)

type Renderer struct {
	dashboard  *mustache.Template
	graph      *mustache.Template
	singlestat *mustache.Template
}

func (r *Renderer) Render(db Dashboard) string {
	graphs := []string{}
	for gi, g := range db.Graphs {
		g.PosX = int(24 * math.Abs(math.Remainder(float64(gi), graphPanelsPerRow)) / graphPanelsPerRow)
		g.PosY = int(math.Floor(float64(gi)/graphPanelsPerRow)) * panelHeight
		graphs = append(graphs, r.graph.Render(g))
	}

	db.Graph = strings.Join(graphs, ",")

	stats := []string{}
	singlestatStartY := int(math.Ceil(float64(len(db.Graphs))/graphPanelsPerRow)) * panelHeight
	for si, s := range db.SingleStats {
		s.PosX = int(24 * math.Abs(math.Remainder(float64(si), singleStatPanelsPerRow)) / singleStatPanelsPerRow)
		s.PosY = singlestatStartY + (int(math.Floor(float64(si)/singleStatPanelsPerRow)) * panelHeight)
		stats = append(stats, r.singlestat.Render(s))
	}

	db.SingleStat = strings.Join(stats, ",")
	return r.dashboard.Render(db)
}

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
