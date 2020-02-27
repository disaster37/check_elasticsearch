package checkes

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/disaster37/go-nagios"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// IndiceSettingResponse is API response
type IndiceSettingResponse struct {
	Settings *IndiceSettings `json:"settings"`
}

// IndiceSettings is API response
type IndiceSettings struct {
	Indice *IndiceSetting `json:"index"`
}

// IndiceSetting is API response
type IndiceSetting struct {
	Blocks *IndiceSettingBlock `json:"blocks,omitempty"`
}

// IndiceSettingBlock is API response
type IndiceSettingBlock struct {
	Read                string `json:"read,omitempty"`
	ReadOnlyAllowDelete string `json:"read_only_allow_delete,omitempty"`
	ReadOnly            string `json:"read_only,omitempty"`
	Write               string `json:"write,omitempty"`
}

// CheckIndiceLocked wrap command line to check
func CheckIndiceLocked(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	if c.String("indice") == "" {
		return errors.New("You must set --indice parameter")
	}

	monitoringData, err := monitorES.CheckIndiceLocked(c.String("indice"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckIndiceLocked check that there are indice locked by security (read_only_allow_delete)
func (h *CheckES) CheckIndiceLocked(indiceName string) (*nagiosPlugin.Monitoring, error) {

	if indiceName == "" {
		return nil, errors.New("IndiceName can't be empty")
	}
	log.Debugf("IndiceName: %s", indiceName)
	monitoringData := nagiosPlugin.NewMonitoring()

	// Query the indice settings
	res, err := h.client.API.Indices.GetSettings(
		h.client.API.Indices.GetSettings.WithContext(context.Background()),
		h.client.API.Indices.GetSettings.WithPretty(),
		h.client.API.Indices.GetSettings.WithIndex(indiceName),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Indice %s not found", indiceName)
			return monitoringData, nil
		}
		return nil, errors.Errorf("Error when get indice %s: %s", indiceName, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get indice setting %s successfully:\n%s", indiceName, string(b))
	indicesSettingResponse := map[string]IndiceSettingResponse{}
	err = json.Unmarshal(b, &indicesSettingResponse)
	if err != nil {
		return nil, err
	}
	log.Debugf("Index settings: %+v", indicesSettingResponse)

	// Check if there are index that are read only by security
	brokenIndices := make([]string, 0)
	nbIndice := 0
	for indiceName, indiceSetting := range indicesSettingResponse {
		log.Debugf("%s: %+v", indiceName, indiceSetting)
		if indiceSetting.Settings.Indice.Blocks != nil && indiceSetting.Settings.Indice.Blocks.ReadOnlyAllowDelete == "true" {
			brokenIndices = append(brokenIndices, indiceName)
		}
		nbIndice++
	}

	if len(brokenIndices) > 0 {
		monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
		monitoringData.AddMessage("There are some indice locked (%d/%d)", nbIndice-len(brokenIndices), nbIndice)
		for _, indiceName := range brokenIndices {
			monitoringData.AddMessage("\tIndice %s", indiceName)
		}

	} else {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No indice locked (%d/%d)", nbIndice, nbIndice)
	}

	monitoringData.AddPerfdata("nbIndices", nbIndice, "")
	monitoringData.AddPerfdata("nbIndicesLocked", len(brokenIndices), "")

	return monitoringData, nil
}
