package v1

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/pkg/textparse"
)

// A MetricConverter contains the logic to convert a Metric to a Panel.
type MetricConverter interface {
	// Can indicates that a MetricConverter can handle a Metric.
	Can(metric Metric) bool
	// Do contains the the code that turns a Metric into one or more Panels.
	Do(metric Metric, options Options) []Panel
}

// Options are passed to the Do() method of each MetricConverter.
type Options struct {
	CounterChangeFunc string
	Datasource        string
	Labels            []string
	TimeRange         string
}

// CounterConverter handles metrics of type Counter.
// It constructs a query that applies a function, rate() or increase() to it.
type CounterConverter struct{}

// Can implements MetricConverter.
func (cc *CounterConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeCounter
}

// Do implements MetricConverter.
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
	g.Legend = strings.Join(legend, " ")
	g.Title = fmt.Sprintf("%s %s over %s", string(m.Name), o.CounterChangeFunc, o.TimeRange)
	g.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s(%s%s[%s])", o.CounterChangeFunc, m.Name, labelSelectors(o.Labels), o.TimeRange)},
	}
	return []Panel{g}
}

// GaugeConverter handles metrics of type Gauge.
// It returns one Singlestat Panel and because of that will only handle metrics without any labels.
type GaugeConverter struct{}

// Can implements MetricConverter.
func (gc *GaugeConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && len(m.LabelKeys) == 0
}

// Do implements MetricConverter.
func (gc *GaugeConverter) Do(m Metric, o Options) []Panel {
	s := Singlestat{}
	s.Datasource = o.Datasource
	s.Description = string(m.Help)
	s.Format = FindFormat(m.Name)
	s.Query = m.Name + labelSelectors(o.Labels)
	s.Title = m.Name
	s.ValueName = "current"
	return []Panel{s}
}

// GaugeDerivConverter handles Metrics of type Gauge that can be converted into Panels of type Graph.
// The deriv() function is applied to the Metric.
// A conversion is only done if the suffix of the Metric suggests that calculating per-second derivative is of interest.
type GaugeDerivConverter struct{}

// Can implements MetricConverter.
func (gd *GaugeDerivConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && strings.HasSuffix(m.Name, "_bytes")
}

// Do implements MetricConverter.
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
	g.Legend = strings.Join(legend, " ")
	g.Title = fmt.Sprintf("%s %s over %s", string(m.Name), "deriv", o.TimeRange)
	g.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s(%s%s[%s])", "deriv", m.Name, labelSelectors(o.Labels), o.TimeRange)},
	}
	return []Panel{g}
}

// GaugeTimestampConverter handles Metrics whose suffix suggests that their value is a unix timestamp.
// It returns a Singlestat Panel and because of that will only handle metrics without any labels.
type GaugeTimestampConverter struct{}

// Can implements MetricConverter.
func (gt *GaugeTimestampConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge &&
		(strings.HasSuffix(m.Name, "_timestamp_seconds") || strings.HasSuffix(m.Name, "_timestamp")) &&
		len(m.LabelKeys) == 0
}

// Do implements MetricConverter.
func (gt *GaugeTimestampConverter) Do(m Metric, o Options) []Panel {
	query := m.Name + labelSelectors(o.Labels)
	if strings.HasSuffix(m.Name, "_timestamp_seconds") {
		query = query + " * 1000"
	}

	s := Singlestat{}
	s.Datasource = o.Datasource
	s.Description = string(m.Help)
	s.Format = FindFormat(m.Name)
	s.Query = query
	s.Title = m.Name
	s.ValueName = "current"
	return []Panel{s}
}

// GaugeInfoConverter handles Metrics whose suffix suggests that information is stored in labels not in the value.
// Examples of such metrics are "go_info" or "prometheus_build_info".
// It returns one Singlestat Panel per label key.
type GaugeInfoConverter struct{}

// Can implements MetricConverter.
func (gi *GaugeInfoConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && strings.HasSuffix(m.Name, "_info")
}

// Do implements MetricConverter.
func (gi *GaugeInfoConverter) Do(m Metric, o Options) []Panel {
	panels := []Panel{}
	for _, lk := range m.LabelKeys {
		s := Singlestat{}
		s.Datasource = o.Datasource
		s.Description = string(m.Help)
		s.Format = defaultFormat
		s.Legend = fmt.Sprintf("{{%s}}", lk)
		s.Query = m.Name + labelSelectors(o.Labels)
		s.Title = fmt.Sprintf("%s - %s", string(m.Name), lk)
		s.ValueName = "name"
		panels = append(panels, s)
	}

	return panels
}

// HistogramConverter handles metrics of type Histogram.
// It returns four Panels of type Graph, one for avg, p50, p95 and p99.
type HistogramConverter struct{}

// Can implements MetricConverter.
func (h *HistogramConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeHistogram
}

// Do implements MetricConverter.
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

	selectors := labelSelectors(o.Labels)

	avg := Graph{}
	avg.Datasource = o.Datasource
	avg.Description = string(m.Help)
	avg.Format = FindFormat(m.Name)
	avg.HasLegend = hasLegend
	avg.Legend = legendFormatted
	avg.Title = fmt.Sprintf("%s avg", string(m.Name))
	avg.Queries = []GraphQuery{
		{Query: fmt.Sprintf("%s_sum%s / %s_count%s", m.Name, selectors, m.Name, selectors)},
	}

	p50 := Graph{}
	p50.Datasource = o.Datasource
	p50.Description = string(m.Help)
	p50.Format = FindFormat(m.Name)
	p50.HasLegend = hasLegend
	p50.Legend = legendFormatted
	p50.Title = fmt.Sprintf("%s p50", string(m.Name))
	p50.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.5, rate(%s_bucket%s[%s]))", m.Name, selectors, o.TimeRange)},
	}

	p90 := Graph{}
	p90.Datasource = o.Datasource
	p90.Description = string(m.Help)
	p90.Format = FindFormat(m.Name)
	p90.HasLegend = hasLegend
	p90.Legend = legendFormatted
	p90.Title = fmt.Sprintf("%s p90", string(m.Name))
	p90.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.9, rate(%s_bucket%s[%s]))", m.Name, selectors, o.TimeRange)},
	}

	p99 := Graph{}
	p99.Datasource = o.Datasource
	p99.Description = string(m.Help)
	p99.Format = FindFormat(m.Name)
	p99.HasLegend = hasLegend
	p99.Legend = legendFormatted
	p99.Title = fmt.Sprintf("%s p99", string(m.Name))
	p99.Queries = []GraphQuery{
		{Query: fmt.Sprintf("histogram_quantile(0.99, rate(%s_bucket%s[%s]))", m.Name, selectors, o.TimeRange)},
	}

	return []Panel{avg, p50, p90, p99}
}

// GaugeWithLabelsConverter handles metrics of type Gauge that define labels.
// It returns one Panel of type Graph.
type GaugeWithLabelsConverter struct{}

// Can implements MetricConverter.
func (gl *GaugeWithLabelsConverter) Can(m Metric) bool {
	return m.Type == textparse.MetricTypeGauge && len(m.LabelKeys) > 0
}

// Do implements MetricConverter.
func (gl *GaugeWithLabelsConverter) Do(m Metric, o Options) []Panel {
	legend := []string{}
	for _, lk := range m.LabelKeys {
		legend = append(legend, fmt.Sprintf("{{%s}}", lk))
	}

	query := m.Name + labelSelectors(o.Labels)
	if strings.HasSuffix(m.Name, "_timestamp_seconds") {
		query = query + " * 1000"
	}

	g := Graph{}
	g.Datasource = o.Datasource
	g.Description = string(m.Help)
	g.Format = FindFormat(m.Name)
	g.HasLegend = true
	g.Legend = strings.Join(legend, " ")
	g.Title = m.Name
	g.Queries = []GraphQuery{
		{Query: query},
	}
	return []Panel{g}
}

func labelSelectors(labels []string) string {
	if len(labels) == 0 {
		return ""
	}

	selectors := []string{}
	for _, l := range labels {
		selectors = append(selectors, fmt.Sprintf(`%s=\"$%s\"`, l, l))
	}

	return "{" + strings.Join(selectors, ",") + "}"
}
