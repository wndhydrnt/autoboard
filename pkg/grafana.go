package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hoisie/mustache"
	log "github.com/sirupsen/logrus"
)

const (
	defaultFormat        = "short"
	grafanaUpdateMessage = "Updated by autoboard"
)

type grafanaCreateDashboardRequest struct {
	Dashboard *json.RawMessage `json:"dashboard"`
	FolderID  int              `json:"folderId"`
	Message   string           `json:"message"`
	Overwrite bool             `json:"overwrite"`
}

type grafanaAPIFolder struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	UID   string `json:"uid"`
}

type grafanaAPIReadFoldersResponse []grafanaAPIFolder

// Grafana encapsulates all interactions with the Grafana API.
type Grafana struct {
	Address  string
	Password string
	Username string
}

// CreateDashboard creates a new dashboard via the Grafana API.
func (g *Grafana) CreateDashboard(d string, folder string) error {
	folderID := 0
	if folder != "" {
		log.Debugf("finding folder by name %s", folder)
		f, err := g.findFolderByName(folder)
		if err != nil {
			return fmt.Errorf("find folder: %s", err)
		}

		folderID = f.ID
	}

	dashboard := json.RawMessage([]byte(d))
	request := &grafanaCreateDashboardRequest{
		Dashboard: &dashboard,
		FolderID:  folderID,
		Message:   grafanaUpdateMessage,
		Overwrite: true,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = g.sendJSON("POST", "/api/dashboards/db", bytes.NewBuffer(payload), nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *Grafana) findFolderByName(name string) (gaf grafanaAPIFolder, _ error) {
	folders, err := g.readFolders()
	if err != nil {
		return gaf, fmt.Errorf("reading folders from grafana: %s", err)
	}

	for _, f := range folders {
		if f.Title == name {
			return f, nil
		}
	}

	return gaf, fmt.Errorf("folder %s not found", name)
}

func (g *Grafana) readFolders() (grafanaAPIReadFoldersResponse, error) {
	data := grafanaAPIReadFoldersResponse{}
	err := g.sendJSON("GET", "/api/folders?limit=10000", nil, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Grafana) sendJSON(method string, url string, body io.Reader, data interface{}) error {
	r, err := http.NewRequest(method, g.Address+url, body)
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")
	if g.Username != "" || g.Password != "" {
		r.SetBasicAuth(g.Username, g.Password)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("grafana API returned status code %d", resp.StatusCode)
	}

	if data == nil {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, data)
	if err != nil {
		return err
	}

	return nil
}

// Renderer contains all logic to create dashboard.
type Renderer struct {
	dashboardTpl         *mustache.Template
	datasource           string
	graphTpl             *mustache.Template
	panelHeight          int
	panelWidthGraph      int
	panelWidthSinglestat int
	rowTpl               *mustache.Template
	singlestatTpl        *mustache.Template
}

// Render takes templates, a Dashboard and a list of Panels and creates a JSON data model of dashboard as required by
// Grafana.
// Panels are placed on the dashboard in the order in which they are defined in the slice.
func (r *Renderer) Render(db Dashboard, panels []Panel) string {
	panelsRendered := []string{}
	posX := 0
	posY := 0
	for _, p := range panels {
		switch p.Type() {
		case PanelTypeRow:
			row := p.(Row)
			row.PosX = 0
			if posY != 0 {
				posY = posY + r.panelHeight
			}

			row.PosY = posY
			panelsRendered = append(panelsRendered, r.rowTpl.Render(row))
			posX = 0
			// A row always has a height of 1
			posY = posY + 1
		case PanelTypeGraph:
			graph := p.(Graph)
			if (posX + r.panelWidthGraph) > 24 {
				posX = 0
				posY = posY + r.panelHeight
			}

			graph.Datasource = r.datasource
			graph.HasDatasource = r.datasource != ""
			graph.Height = r.panelHeight
			graph.PosX = posX
			graph.PosY = posY
			graph.Width = r.panelWidthGraph
			panelsRendered = append(panelsRendered, r.graphTpl.Render(graph))
			posX = posX + r.panelWidthGraph
		case PanelTypeSinglestat:
			singlestat := p.(Singlestat)
			if (posX + r.panelWidthSinglestat) > 24 {
				posX = 0
				posY = posY + r.panelHeight
			}

			singlestat.Datasource = r.datasource
			singlestat.HasDatasource = r.datasource != ""
			singlestat.Height = r.panelHeight
			singlestat.PosX = posX
			singlestat.PosY = posY
			singlestat.Width = r.panelWidthSinglestat
			panelsRendered = append(panelsRendered, r.singlestatTpl.Render(singlestat))
			posX = posX + r.panelWidthSinglestat
		}
	}

	db.Panels = strings.Join(panelsRendered, ",")
	return r.dashboardTpl.Render(db)
}

var (
	PanelTypeGraph      = "graph"
	PanelTypeRow        = "row"
	PanelTypeSinglestat = "singlestat"
)

// A Panel is a data container.
type Panel interface {
	// Type helps identifying the underlying type of a Panel.
	Type() string
}

// Dashboard holds data used on the top level of the JSON model.
type Dashboard struct {
	Panels    string
	Title     string
	Variables []Variable
}

// Variable is rendered as a selector by Grafana.
type Variable struct {
	Datasource string
	HasMore    bool
	Query      string
	Name       string
}

func labelsToVariables(datasource string, labels []string, query string) []Variable {
	variables := []Variable{}
	for i, l := range labels {
		v := Variable{Datasource: datasource, Name: l}
		v.HasMore = i+1 < len(labels)
		v.Query = fmt.Sprintf("label_values(%s, %s)", query, l)
		variables = append(variables, v)
	}

	return variables
}

// A Row is rendered as a row by Grafana.
// Height and width are not configurable because a row in Grafana always has a height of "1" and a width of "24".
type Row struct {
	ID    int
	PosX  int
	PosY  int
	Title string
}

// Type implements Panel.
func (r Row) Type() string {
	return PanelTypeRow
}

// A Graph is rendered as a graph panel by Grafana.
type Graph struct {
	Datasource     string
	Description    string
	Format         string
	HasDatasource  bool
	HasLegend      bool
	HasThreshold   bool
	Height         int
	ID             int
	Legend         string
	Queries        []GraphQuery
	PosX           int
	PosY           int
	ThresholdOP    string
	ThresholdValue string
	Title          string
	Width          int
}

// Type implements Panel.
func (g Graph) Type() string {
	return PanelTypeGraph
}

// A GraphQuery is rendered as one query in a graph.
type GraphQuery struct {
	Code    string
	HasMore bool
	Query   string
	RefID   string
}

// A Singlestat is rendered as a singlestat panel by Grafana.
type Singlestat struct {
	Datasource         string
	Description        string
	Format             string
	HasDatasource      bool
	Height             int
	ID                 int
	Legend             string
	Query              string
	PosX               int
	PosY               int
	ThresholdInvertNo  bool
	ThresholdInvertYes bool
	ThresholdValue     string
	Title              string
	ValueName          string
	Width              int
}

// Type implements Panel.
func (s Singlestat) Type() string {
	return PanelTypeSinglestat
}
