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

//ILMExplainResponse is the API response
type ILMExplainResponse struct {
	Indices map[string]ILMExplain `json:"indices,omitempty"`
}

// ILMExplain is the API response
type ILMExplain struct {
	Index    string    `json:"index,omitempty"`
	Policy   string    `json:"policy,omitempty"`
	StepInfo *StepInfo `json:"step_info,omitempty"`
}

// StepInfo is the API response
type StepInfo struct {
	Type       string `json:"type,omitempty"`
	Reason     string `json:"reason,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// ILMStatusResponse is the API response
type ILMStatusResponse struct {
	OperationMode string `json:"operation_mode,omitempty"`
}

// CheckILMError Wrap cli argument and call check
func CheckILMError(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	if c.String("indice") == "" {
		return errors.New("You must set --indice parameter")
	}

	monitoringData, err := monitorES.CheckILMError(c.String("indice"), c.StringSlice("exclude"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckILMStatus wrap cli with monitoring check
func CheckILMStatus(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckILMStatus()
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckILMError check that there are no ILM policy failed on indice name
func (h *CheckES) CheckILMError(indiceName string, excludeIndices []string) (*nagiosPlugin.Monitoring, error) {

	if indiceName == "" {
		return nil, errors.New("IndiceName can't be empty")
	}
	log.Debugf("IndiceName: %s", indiceName)
	log.Debugf("ExcludeIndices: %+v", excludeIndices)
	monitoringData := nagiosPlugin.NewMonitoring()

	// Query if there are ILM error
	res, err := h.client.API.ILM.ExplainLifecycle(
		indiceName,
		h.client.API.ILM.ExplainLifecycle.WithContext(context.Background()),
		h.client.API.ILM.ExplainLifecycle.WithOnlyErrors(true),
		h.client.API.ILM.ExplainLifecycle.WithOnlyManaged(true),
		h.client.API.ILM.ExplainLifecycle.WithPretty(),
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
		return nil, errors.Errorf("Error when get ILM explain on indice %s: %s", indiceName, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get lifecycle explain on index %s successfully:\n%s", indiceName, string(b))
	ilmExplainResponse := &ILMExplainResponse{}
	err = json.Unmarshal(b, ilmExplainResponse)
	if err != nil {
		return nil, err
	}

	// Check if there are some ILM polices that failed
	if ilmExplainResponse.Indices == nil || len(ilmExplainResponse.Indices) == 0 {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No error found on indice %s", indiceName)
		monitoringData.AddPerfdata("NbIndiceFailed", 0, "")
		return monitoringData, nil
	}

	// Remove exclude indices
	for _, indiceExcludeName := range excludeIndices {
		if _, ok := ilmExplainResponse.Indices[indiceExcludeName]; ok {
			log.Debugf("Indice %s is exclude", indiceExcludeName)
			delete(ilmExplainResponse.Indices, indiceExcludeName)
		}
	}

	// Compute error
	if len(ilmExplainResponse.Indices) == 0 {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No error found on indice %s", indiceName)
		monitoringData.AddPerfdata("NbIndiceFailed", 0, "")
		return monitoringData, nil
	}
	monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
	monitoringData.AddPerfdata("NbIndiceFailed", len(ilmExplainResponse.Indices), "")
	monitoringData.AddMessage("There are %d indices failed", len(ilmExplainResponse.Indices))
	for _, ilmExplain := range ilmExplainResponse.Indices {
		monitoringData.AddMessage("Indice %s (%s): %s", ilmExplain.Index, ilmExplain.Policy, ilmExplain.StepInfo.Reason)
	}

	return monitoringData, nil
}

// CheckILMStatus check the status of ILM is running
func (h *CheckES) CheckILMStatus() (*nagiosPlugin.Monitoring, error) {

	monitoringData := nagiosPlugin.NewMonitoring()

	// Check the ILM status
	res, err := h.client.API.ILM.GetStatus(
		h.client.API.ILM.GetStatus.WithContext(context.Background()),
		h.client.API.ILM.GetStatus.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("ILM Status not found")
			return monitoringData, nil
		}
		return nil, errors.Errorf("Error when get ILM status: %s", res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get ILM status successfully:\n%s", string(b))
	ilmStatusResponse := &ILMStatusResponse{}
	err = json.Unmarshal(b, ilmStatusResponse)
	if err != nil {
		return nil, err
	}

	if ilmStatusResponse.OperationMode == "RUNNING" {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("ILM is running")
		return monitoringData, nil
	}

	monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
	monitoringData.AddMessage("ILM is not running: %s", ilmStatusResponse.OperationMode)
	return monitoringData, nil
}
