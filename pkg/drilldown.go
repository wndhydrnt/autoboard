package v1

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/textparse"
	log "github.com/sirupsen/logrus"
	"github.com/wndhydrnt/autoboard/pkg/config"
)

const (
	groupNameGeneral = "ab_general"
)

type Group struct {
	Metrics []Metric
	Name    string
}

type Metric struct {
	Help      string
	LabelKeys []string
	Name      string
	Type      textparse.MetricType
}

type Groups map[string][]Metric

type Drilldown struct {
	Converters []MetricConverter
}

func NewDrilldown() *Drilldown {
	return &Drilldown{
		Converters: []MetricConverter{
			&HistogramConverter{},
			&GaugeInfoConverter{},
			&GaugeDerivConverter{},
			&GaugeTimestampConverter{},
			&GaugeWithLabelsConverter{},
			&GaugeConverter{},
			&CounterConverter{},
		},
	}
}

func (d *Drilldown) Run(cfg config.Config, counterChangeFunc, endpoint string, groupLevel int, title, prefix, timeRange string) error {
	log.SetLevel(cfg.LogLevel)

	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Get(endpoint)
	if err != nil {
		return fmt.Errorf("reading Prometheus endpoint %s: %v", endpoint, err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body of Prometheus endpoint %s: %v", endpoint, err)
	}

	metrics := parseMetrics(b, resp.Header.Get("Content-Type"), prefix)
	groups := groupMetrics(metrics, groupLevel)
	panels := d.convertGroupsToPanels(groups, Options{counterChangeFunc, cfg.DatasourceDefault, timeRange})
	r := &Renderer{
		dashboardTpl:  cfg.TemplateDashboard,
		graphTpl:      cfg.TemplateGraph,
		rowTpl:        cfg.TemplateRow,
		singlestatTpl: cfg.TemplateSinglestat,
	}
	db := Dashboard{Title: title}
	s := r.Render(db, panels)
	gf := &Grafana{Address: cfg.GrafanaAddress}
	err = gf.CreateDashboard(s)
	if err != nil {
		return fmt.Errorf("create drilldown dashboard: %s", err)
	}

	return nil
}

func (d *Drilldown) convertGroupsToPanels(groups Groups, options Options) []Panel {
	panels := []Panel{}
	for name, metrics := range groups {
		if len(metrics) > 0 {
			panels = append(panels, Row{Title: name})
		}

		for _, m := range metrics {
			c := findConverter(m, d.Converters)
			if c == nil {
				log.Debugf("no converter found for metric %s (%s)", string(m.Name), m.Type)
				continue
			}

			newPanels := c.Do(m, options)
			panels = append(panels, newPanels...)
		}
	}

	return panels
}

func parseMetrics(b []byte, contentType string, prefix string) []Metric {
	metrics := []Metric{}
	cm := Metric{}
	p := textparse.New(b, contentType)
	for {
		et, err := p.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		switch et {
		case textparse.EntryType:
			n, t := p.Type()
			cm.Name = string(n)
			cm.Type = t
			continue
		case textparse.EntryHelp:
			n, h := p.Help()
			cm.Name = string(n)
			cm.Help = string(h)
			continue
		default:
		}

		if cm.Name != "" {
			var lset labels.Labels
			p.Metric(&lset)
			for _, l := range lset {
				if l.Name == labels.MetricName {
					continue
				}

				cm.LabelKeys = append(cm.LabelKeys, l.Name)
			}
			metrics = append(metrics, cm)
			cm = Metric{}
		}
	}

	result := []Metric{}
	for _, m := range metrics {
		if strings.HasPrefix(m.Name, prefix) {
			result = append(result, m)
		}
	}

	return result
}

func findConverter(m Metric, converters []MetricConverter) MetricConverter {
	for _, c := range converters {
		if c.Can(m) {
			return c
		}
	}

	return nil
}

func groupMetrics(metrics []Metric, level int) Groups {
	groups := Groups{
		groupNameGeneral: {},
	}
	for _, m := range metrics {
		parts := strings.Split(m.Name, "_")
		if len(parts) <= level || level == 0 {
			groups[groupNameGeneral] = append(groups[groupNameGeneral], m)
			continue
		}

		key := strings.Join(parts[:level], "_")
		_, exists := groups[key]
		if exists {
			groups[key] = append(groups[key], m)
		} else {
			groups[key] = []Metric{m}
		}
	}

	return groups
}
