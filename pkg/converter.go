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
	legend := []string{}
	for _, l := range m.LabelKeys {
		legend = append(legend, fmt.Sprintf("{{%s}}", l))
	}

	hasLegend := true
	if len(legend) == 0 {
		legend = append(legend, "{{instance}}")
		hasLegend = false

	}

	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindRangeFormat(m.Name)
	g.HasLegend = hasLegend
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
	s.ValueName = "current"
	s.Width = panelWidth
	return []Panel{s}
}

type GaugeDerivConverter struct{}

func (gd *GaugeDerivConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && bytes.HasSuffix(m.Name, []byte("_bytes"))
}

func (gd *GaugeDerivConverter) Do(m Metric, o Options) []Panel {
	legend := []string{}
	for _, lk := range m.LabelKeys {
		legend = append(legend, fmt.Sprintf("{{%s}}", lk))
	}

	hasLegend := true
	if len(legend) == 0 {
		legend = append(legend, "{{instance}}")
		hasLegend = false
	}

	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindRangeFormat(m.Name)
	g.HasLegend = hasLegend
	g.Height = panelHeight
	g.Legend = strings.Join(legend, " ")
	g.Title = fmt.Sprintf("%s %s over %s", string(m.Name), "deriv", o.TimeRange)
	g.Width = panelWidth * 2
	g.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s(%s[%s])", "deriv", m.Name, o.TimeRange)},
	}
	return []Panel{g}
}

type GaugeTimestampConverter struct{}

func (gt *GaugeTimestampConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge &&
		(bytes.HasSuffix(m.Name, []byte("_timestamp_seconds")) || bytes.HasSuffix(m.Name, []byte("_timestamp"))) &&
		len(m.LabelKeys) == 0
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
	s.ValueName = "current"
	s.Width = panelWidth
	return []Panel{s}
}

type GaugeInfoConverter struct{}

func (gi *GaugeInfoConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && bytes.HasSuffix(m.Name, []byte("_info"))
}

func (gi *GaugeInfoConverter) Do(m Metric, o Options) []Panel {
	panels := []Panel{}
	for _, lk := range m.LabelKeys {
		s := Singlestat{}
		s.Datasource = o.Datasource
		s.Description = string(m.Help)
		s.Format = defaultFormat
		s.Height = panelHeight
		s.Legend = fmt.Sprintf("{{%s}}", lk)
		s.Query = string(m.Name)
		s.Title = fmt.Sprintf("%s - %s", string(m.Name), lk)
		s.ValueName = "name"
		s.Width = panelWidth
		panels = append(panels, s)
	}

	return panels
}

type HistogramConverter struct{}

func (h *HistogramConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeHistogram
}

func (h *HistogramConverter) Do(m Metric, o Options) []Panel {
	legend := []string{}
	for _, lk := range m.LabelKeys {
		if lk == "le" {
			continue
		}

		legend = append(legend, fmt.Sprintf("{{%s}}", lk))
	}

	hasLegend := true
	if len(legend) == 0 {
		legend = append(legend, "{{instance}}")
		hasLegend = false
	}

	legendFormatted := strings.Join(legend, " ")

	avg := Graph{}
	avg.Datasource = o.Datasource
	avg.Description = string(m.Help)
	avg.Format = FindFormat(m.Name)
	avg.HasLegend = hasLegend
	avg.Height = panelHeight
	avg.Legend = legendFormatted
	avg.Title = fmt.Sprintf("%s avg", string(m.Name))
	avg.Width = panelWidth * 2
	avg.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s_sum / %s_count", m.Name, m.Name)},
	}

	p50 := Graph{}
	p50.Datasource = o.Datasource
	p50.Description = string(m.Help)
	p50.Format = FindFormat(m.Name)
	p50.HasLegend = hasLegend
	p50.Height = panelHeight
	p50.Legend = legendFormatted
	p50.Title = fmt.Sprintf("%s p50", string(m.Name))
	p50.Width = panelWidth * 2
	p50.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.5, rate(%s_bucket[%s]))", m.Name, o.TimeRange)},
	}

	p90 := Graph{}
	p90.Datasource = o.Datasource
	p90.Description = string(m.Help)
	p90.Format = FindFormat(m.Name)
	p90.HasLegend = hasLegend
	p90.Height = panelHeight
	p90.Legend = legendFormatted
	p90.Title = fmt.Sprintf("%s p90", string(m.Name))
	p90.Width = panelWidth * 2
	p90.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.9, rate(%s_bucket[%s]))", m.Name, o.TimeRange)},
	}

	p99 := Graph{}
	p99.Datasource = o.Datasource
	p99.Description = string(m.Help)
	p99.Format = FindFormat(m.Name)
	p99.HasLegend = hasLegend
	p99.Height = panelHeight
	p99.Legend = legendFormatted
	p99.Title = fmt.Sprintf("%s p99", string(m.Name))
	p99.Width = panelWidth * 2
	p99.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.99, rate(%s_bucket[%s]))", m.Name, o.TimeRange)},
	}

	return []Panel{avg, p50, p90, p99}
}

type GaugeWithLabelsConverter struct{}

func (gl *GaugeWithLabelsConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && len(m.LabelKeys) > 0
}

func (gl *GaugeWithLabelsConverter) Do(m Metric, o Options) []Panel {
	legend := []string{}
	for _, lk := range m.LabelKeys {
		legend = append(legend, fmt.Sprintf("{{%s}}", lk))
	}

	query := string(m.Name)
	if bytes.HasSuffix(m.Name, []byte("_timestamp_seconds")) {
		query = query + " * 1000"
	}

	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindFormat(m.Name)
	g.HasLegend = true
	g.Height = panelHeight
	g.Legend = strings.Join(legend, " ")
	g.Title = string(m.Name)
	g.Width = panelWidth * 2
	g.Queries = []GraphQuery{
		{Query: query},
	}
	return []Panel{g}
}
