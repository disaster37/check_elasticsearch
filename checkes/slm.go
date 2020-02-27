package checkes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/disaster37/go-nagios"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SnapshotsResponse is the API response
type SnapshotsResponse struct {
	Snaphots []SnapshotResponse `json:"snapshots"`
}

// SnapshotResponse is the API response
type SnapshotResponse struct {
	Snapshot           string            `json:"snapshot,omitempty"`
	Indices            []string          `json:"indices,omitempty"`
	IncludeGlobalState bool              `json:"include_global_state,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	State              string            `json:"state,omitempty"`
	Failures           []SnapshotFailure `json:"failures,omitempty"`
	StartTime          time.Time         `json:"start_time,omitempty"`
	EndTime            time.Time         `json:"end_time,omitempty"`
}

// SnapshotFailure is the API response
type SnapshotFailure struct {
	NodeID string `json:"node_id,omitempty"`
	Indice string `json:"index,omitempty"`
	Reason string `json:"reason,omitempty"`
	Status string `json:"status,omitempty"`
}

// SLMStatusResponse is the API response
type SLMStatusResponse struct {
	OperationMode string `json:"operation_mode,omitempty"`
}

// CheckSLMError wrap command line to check
func CheckSLMError(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	if c.String("repository") == "" {
		return errors.New("You must set --repository parameter")
	}

	monitoringData, err := monitorES.CheckSLMError(c.String("repository"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckSLMStatus wrap command line to check
func CheckSLMStatus(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckSLMStatus()
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckSLMError check that there are no ILM policy failed on indice name
func (h *CheckES) CheckSLMError(snapshotRepositoryName string) (*nagiosPlugin.Monitoring, error) {

	if snapshotRepositoryName == "" {
		return nil, errors.New("SnapshotRepositoryName can't be empty")
	}
	log.Debugf("snapshotRepositoryName: %s", snapshotRepositoryName)
	monitoringData := nagiosPlugin.NewMonitoring()

	// Query if there are snapshot error
	res, err := h.client.API.Snapshot.Get(
		snapshotRepositoryName,
		[]string{"_all"},
		h.client.API.Snapshot.Get.WithContext(context.Background()),
		h.client.API.Snapshot.Get.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Repository %s not found", snapshotRepositoryName)
			return monitoringData, nil
		}
		return nil, errors.Errorf("Error when get snapshots on repository %s: %s", snapshotRepositoryName, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get snapshots on repository %s successfully:\n%s", snapshotRepositoryName, string(b))
	snapshotsResponse := &SnapshotsResponse{}
	err = json.Unmarshal(b, snapshotsResponse)
	if err != nil {
		return nil, err
	}

	// Check if there are some snapshot failed
	if (snapshotsResponse.Snaphots == nil) || (len(snapshotsResponse.Snaphots) == 0) {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No snapshot on repository %s", snapshotRepositoryName)
		monitoringData.AddPerfdata("NbSnapshot", 0, "")
		monitoringData.AddPerfdata("NbSnapshotFailed", 0, "")
		return monitoringData, nil
	}

	nbSnapshot := 0
	snapshotsFailed := make([]SnapshotResponse, 0)
	monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
	for _, snapshotResponse := range snapshotsResponse.Snaphots {
		nbSnapshot++
		if snapshotResponse.State != "SUCCESS" && snapshotResponse.State != "IN_PROGRESS" {
			monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
			snapshotsFailed = append(snapshotsFailed, snapshotResponse)
		}
	}
	if len(snapshotsFailed) > 0 {
		monitoringData.AddMessage("Some snapshots failed (%d/%d)", nbSnapshot-len(snapshotsFailed), nbSnapshot)
		for _, snapshotFailed := range snapshotsFailed {

			var errorMsg strings.Builder
			for _, failure := range snapshotFailed.Failures {
				errorMsg.WriteString(fmt.Sprintf("\n\tIndice %s on node %s failed with status %s: %s", failure.Indice, failure.NodeID, failure.Status, failure.Reason))
			}

			monitoringData.AddMessage("Snapshot %s failed (%s - %s) with status %s: %s", snapshotFailed.Snapshot, snapshotFailed.StartTime, snapshotFailed.EndTime, snapshotFailed.State, errorMsg.String())
		}
	} else {
		monitoringData.AddMessage("All snapshots are ok (%d/%d)", nbSnapshot, nbSnapshot)
	}

	monitoringData.AddPerfdata("NbSnapshot", nbSnapshot, "")
	monitoringData.AddPerfdata("NbSnapshotFailed", len(snapshotsFailed), "")

	return monitoringData, nil
}

// CheckSLMStatus check that SLM service is running
func (h *CheckES) CheckSLMStatus() (*nagiosPlugin.Monitoring, error) {

	monitoringData := nagiosPlugin.NewMonitoring()

	res, err := h.client.API.SlmGetStatus(
		h.client.API.SlmGetStatus.WithContext(context.Background()),
		h.client.API.SlmGetStatus.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("SLM status not found")
			return monitoringData, nil
		}
		return nil, errors.Errorf("Error when get SLM status: %s", res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get SLM status successfully:\n%s", string(b))
	slmStatusResponse := &SLMStatusResponse{}
	err = json.Unmarshal(b, slmStatusResponse)
	if err != nil {
		return nil, err
	}

	if slmStatusResponse.OperationMode == "RUNNING" {
		monitoringData.SetStatus(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("SLM status is running")
		return monitoringData, nil
	}

	monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
	monitoringData.AddMessage("SLM status is not running: %s", slmStatusResponse.OperationMode)

	return monitoringData, nil
}
