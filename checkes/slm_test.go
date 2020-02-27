package checkes

import (
	"context"
	"strings"

	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/stretchr/testify/assert"
)

func (s *CheckESTestSuite) TestCheckSLMError() {

	// When reposiotry exist
	checkES := s.monitorES.(*CheckES)
	checkES.client.API.Snapshot.CreateRepository(
		"snapshot",
		strings.NewReader(`
			{
				"type": "fs",
  				"settings": {
    				"location": "/tmp",
    				"compress": true
  				}
			}
		`),
		checkES.client.API.Snapshot.CreateRepository.WithContext(context.Background()),
	)
	monitoringData, err := s.monitorES.CheckSLMError("snapshot")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When repository not exist
	monitoringData, err = s.monitorES.CheckSLMError("foo")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_UNKNOWN, monitoringData.Status())
}

func (s *CheckESTestSuite) TestCheckSLMStatus() {

	checkES := s.monitorES.(*CheckES)

	// When SLM is stopped
	checkES.client.API.SlmStop(
		checkES.client.API.SlmStop.WithContext(context.Background()),
	)
	monitoringData, err := s.monitorES.CheckSLMStatus()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())

	// When SLM is running
	checkES.client.API.SlmStart(
		checkES.client.API.SlmStart.WithContext(context.Background()),
	)
	monitoringData, err = s.monitorES.CheckSLMStatus()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

}
