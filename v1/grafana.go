package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	grafanaUpdateMessage = "Updated by alert-to-dashboard"
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
