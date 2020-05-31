package v1

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/pkg/textparse"
)

type Panel interface {
	Type() string
}

type MetricConverter interface {
	Can(metric Metric) bool
	Do(metric Metric, options Options) []Panel
}

type Options struct {
	CounterChangeFunc string
	Datasource        string
	TimeRange         string
}

type CounterConverter struct{}

func (cc *CounterConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeCounter
}

func (cc *CounterConverter) Do(m Metric, o Options) []Panel {
	legend := []string{"{{instance}}"}
	for _, l := range m.LabelKeys {
		legend = append(legend, fmt.Sprintf("{{%s}}", l))
	}

	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindRangeFormat(m.Name)
	g.HasLegend = true
	g.Height = panelHeight
	g.Legend = strings.Join(legend, " ")
	g.Title = fmt.Sprintf("%s %s over %s", string(m.Name), o.CounterChangeFunc, o.TimeRange)
	g.Width = panelWidth * 2
	g.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s(%s[%s])", o.CounterChangeFunc, m.Name, o.TimeRange)},
	}
	return []Panel{g}
}

type GaugeConverter struct{}

func (gc *GaugeConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && len(m.LabelKeys) == 0
}

func (gc *GaugeConverter) Do(m Metric, o Options) []Panel {
	s := Singlestat{}
	s.Datasource = o.Datasource
	s.Description = string(m.Help)
	s.Format = FindFormat(m.Name)
	s.Height = panelHeight
	s.Query = string(m.Name)
	s.Title = string(m.Name)
	s.Width = panelWidth
	return []Panel{s}
}

type GaugeDerivConverter struct{}

func (gd *GaugeDerivConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && bytes.HasSuffix(m.Name, []byte("_bytes"))
}

func (gd *GaugeDerivConverter) Do(m Metric, o Options) []Panel {
	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindRangeFormat(m.Name)
	g.HasLegend = true
	g.Height = panelHeight
	g.Legend = "{{instance}}"
	g.Title = fmt.Sprintf("%s %s over %s", string(m.Name), "deriv", o.TimeRange)
	g.Width = panelWidth * 2
	g.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s(%s[%s])", "deriv", m.Name, o.TimeRange)},
	}
	return []Panel{g}
}

type GaugeTimestampConverter struct{}

func (gt *GaugeTimestampConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && (bytes.HasSuffix(m.Name, []byte("_timestamp_seconds")) || bytes.HasSuffix(m.Name, []byte("_timestamp")))
}

func (gt *GaugeTimestampConverter) Do(m Metric, o Options) []Panel {
	query := string(m.Name)
	if bytes.HasSuffix(m.Name, []byte("_timestamp_seconds")) {
		query = query + " * 1000"
	}

	s := Singlestat{}
	s.Datasource = o.Datasource
	s.Description = string(m.Help)
	s.Format = FindFormat(m.Name)
	s.Height = panelHeight
	s.Query = query
	s.Title = string(m.Name)
	s.Width = panelWidth
	return []Panel{s}
}
