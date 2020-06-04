package v1

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/textparse"
	log "github.com/sirupsen/logrus"
	"github.com/wndhydrnt/autoboard/pkg/config"
)

type Metric struct {
	Help      []byte
	LabelKeys []string
	Name      []byte
	Type      textparse.MetricType
}

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

func (d *Drilldown) Run(cfg config.Config, counterChangeFunc, endpoint, title, prefix, timeRange string) error {
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
	db := Dashboard{
		Title: title,
	}
	db.Panels = d.convertMetricsToPanels(metrics, Options{counterChangeFunc, cfg.DatasourceDefault, timeRange})
	r := &Renderer{
		dashboardTpl:  cfg.TemplateDashboard,
		graphTpl:      cfg.TemplateGraph,
		singlestatTpl: cfg.TemplateSinglestat,
	}
	s := r.Render(db)
	gf := &Grafana{Address: cfg.GrafanaAddress}
	err = gf.CreateDashboard(s)
	if err != nil {
		return fmt.Errorf("create drilldown dashboard: %s", err)
	}

	return nil
}

func (d *Drilldown) convertMetricsToPanels(metrics []Metric, options Options) []Panel {
	panels := []Panel{}
	for _, m := range metrics {
		c := findConverter(m, d.Converters)
		if c == nil {
			log.Debugf("no converter found for metric %s (%s)", string(m.Name), m.Type)
			continue
		}

		newPanels := c.Do(m, options)
		panels = append(panels, newPanels...)
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
			cm.Name, cm.Type = p.Type()
			continue
		case textparse.EntryHelp:
			cm.Name, cm.Help = p.Help()
			continue
		default:
		}

		if cm.Name != nil {
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
	prefixB := []byte(prefix)
	for _, m := range metrics {
		if bytes.HasPrefix(m.Name, prefixB) {
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
