package checkes

import (
	"encoding/json"
	"io/ioutil"

	"github.com/disaster37/go-nagios"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// TransformStatData is the transform stats data response
type TransformStatData struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

// TransformStatsData is the transform stats data response
type TransformStatsData struct {
	TransformStats []TransformStatData `json:"transforms"`
}

// CheckILMError Wrap cli argument and call check
func CheckTransformError(c *cli.Context) error {

	monitorES, err := manageElasticsearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckTransformError(c.String("name"), c.StringSlice("exclude"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckTransformError check that there are no transform failed
func (h *CheckES) CheckTransformError(transformName string, excludeTransforms []string) (*nagiosPlugin.Monitoring, error) {

	if transformName == "" {
		transformName = "_all"
	}
	log.Debugf("TransformName: %s", transformName)
	log.Debugf("ExcludeTransform: %+v", excludeTransforms)
	monitoringData := nagiosPlugin.NewMonitoring()

	// Query if there are Transform error
	res, err := h.client.API.TransformGetTransformStats(
		transformName,
		h.client.API.TransformGetTransformStats.WithSize(1000),
		h.client.API.TransformGetTransformStats.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Transform %s not found", transformName)
			return monitoringData, nil
		}
		return nil, errors.Errorf("Error when get Transform stats %s: %s", transformName, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Get Transform %s successfully:\n%s", transformName, string(b))

	transformStats := &TransformStatsData{}
	err = json.Unmarshal(b, transformStats)
	if err != nil {
		return nil, err
	}

	// Handle not found transform when id is provided
	if len(transformStats.TransformStats) == 0 && transformName != "_all" && transformName != "*" {
		monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
		monitoringData.AddMessage("Transform %s not found", transformName)
		return monitoringData, nil
	}

	// Loop over index and exclude transform if needed
	var isExclude bool
	var nbTransformStarted int
	var nbTranformFailed int
	var nbTransformStopped int
	for _, transformStat := range transformStats.TransformStats {
		isExclude = false
		for _, excludeTransform := range excludeTransforms {
			if transformStat.ID == excludeTransform {
				isExclude = true
				log.Debugf("Transform %s is exclude", transformStat.ID)
				break
			}
		}
		if !isExclude {
			if transformStat.State == "indexing" || transformStat.State == "started" {
				nbTransformStarted++
				continue
			} else if transformStat.State == "stopped" || transformStat.State == "stopping" {
				nbTransformStopped++
				continue
			} else {
				nbTranformFailed++
				monitoringData.SetStatus(nagiosPlugin.STATUS_CRITICAL)
				monitoringData.AddMessage("Transform %s %s: %s", transformStat.ID, transformStat.State, transformStat.Reason)
				continue
			}
		}

	}

	monitoringData.AddPerfdata("nbTransformFailed", nbTranformFailed, "")
	monitoringData.AddPerfdata("nbTransformStopped", nbTransformStopped, "")
	monitoringData.AddPerfdata("nbTransformStarted", nbTransformStarted, "")

	if monitoringData.Status() == nagiosPlugin.STATUS_OK {
		if transformName == "_all" || transformName == "*" {
			monitoringData.AddMessage("All transform works fine")
		} else {
			monitoringData.AddMessage("Transform %s works fine", transformName)
		}
	}

	return monitoringData, nil
}
