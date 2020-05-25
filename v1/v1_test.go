package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/autoboard/v1/config"
)

type grafanaSearchResponse struct {
	UID string `json:"uid"`
}

type grafanaGetDashboardResponse struct {
	Dashboard map[string]interface{} `json:"dashboard"`
}

func TestDashboardGeneration(t *testing.T) {
	cfg, err := config.Parse("./test/config.yml")
	require.NoError(t, err)
	err = RunAlert(cfg, []*regexp.Regexp{regexp.MustCompile(".*")})
	require.NoError(t, err)

	searchResp, err := http.Get("http://admin:admin@127.0.0.1:12959/api/search")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(searchResp.Body)
	require.NoError(t, err)
	defer searchResp.Body.Close()
	var search []grafanaSearchResponse
	err = json.Unmarshal(b, &search)
	require.NoError(t, err)
	require.Len(t, search, 1)

	dashboardResp, err := http.Get("http://admin:admin@127.0.0.1:12959/api/dashboards/uid/" + search[0].UID)
	require.NoError(t, err)
	dashboardPayload, err := ioutil.ReadAll(dashboardResp.Body)
	require.NoError(t, err)
	defer dashboardResp.Body.Close()
	var result grafanaGetDashboardResponse
	fmt.Println(string(dashboardPayload))
	err = json.Unmarshal(dashboardPayload, &result)
	require.NoError(t, err)
	// Reset fields dynamically set by Grafana
	result.Dashboard["uid"] = "abc"
	result.Dashboard["version"] = float64(1)

	expectedData, err := ioutil.ReadFile("./test/result.json")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err)

	require.Equal(t, expected, result.Dashboard)
}
