package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"

	"github.com/hoisie/mustache"
)

const (
	grafanaUpdateMessage = "Updated by autoboard"
)

type GrafanaCreateDashboardRequest struct {
	Dashboard *json.RawMessage `json:"dashboard"`
	FolderID  int              `json:"folderId"`
	Message   string           `json:"message"`
	Overwrite bool             `json:"overwrite"`
}

type Grafana struct {
	Address         string
	FolderIDDefault int
}

func (g *Grafana) CreateDashboard(d string) error {
	dashboard := json.RawMessage([]byte(d))
	request := &GrafanaCreateDashboardRequest{
		Dashboard: &dashboard,
		FolderID:  g.FolderIDDefault,
		Message:   grafanaUpdateMessage,
		Overwrite: true,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := http.Post(g.Address+"/api/dashboards/db", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error reading response body: %w", err)
		}

		defer resp.Body.Close()
		return fmt.Errorf("grafana returned %d: %s", resp.StatusCode, b)
	}

	return nil
}

type Renderer struct {
	dashboardTpl  *mustache.Template
	graphTpl      *mustache.Template
	singlestatTpl *mustache.Template
}

func (r *Renderer) Render(db Dashboard) string {
	graphs := []string{}
	for gi, g := range db.Graphs {
		g.PosX = int(24 * (gi % graphPanelsPerRow) / graphPanelsPerRow)
		g.PosY = int(math.Floor(float64(gi)/graphPanelsPerRow)) * panelHeight
		graphs = append(graphs, r.graphTpl.Render(g))
	}

	db.Graph = strings.Join(graphs, ",")

	stats := []string{}
	singlestatStartY := int(math.Ceil(float64(len(db.Graphs))/graphPanelsPerRow)) * panelHeight
	for si, s := range db.SingleStats {
		s.PosX = int(24 * (si % singleStatPanelsPerRow) / singleStatPanelsPerRow)
		s.PosY = singlestatStartY + (int(math.Floor(float64(si)/singleStatPanelsPerRow)) * panelHeight)
		stats = append(stats, r.singlestatTpl.Render(s))
	}

	db.HasSinglestat = len(stats) > 0
	db.SingleStat = strings.Join(stats, ",")
	return r.dashboardTpl.Render(db)
}

var (
	PanelTypeGraph      = "graph"
	PanelTypeSinglestat = "singlestat"
)

type Dashboard struct {
	Graph         string
	Graphs        []Graph
	HasSinglestat bool
	SingleStat    string
	SingleStats   []Singlestat
	Title         string
}

type Graph struct {
	Datasource     string
	Description    string
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

func (g Graph) Type() string {
	return PanelTypeGraph
}

type GraphQuery struct {
	Code    string
	HasMore bool
	Query   string
	RefID   string
}

type Singlestat struct {
	Datasource         string
	Description        string
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

func (s Singlestat) Type() string {
	return PanelTypeSinglestat
}

const (
	defaultFormat          = "short"
	graphPanelsPerRow      = 2
	panelHeight            = 5
	panelWidth             = 6
	singleStatPanelsPerRow = 4
)
