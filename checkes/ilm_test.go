package checkes

import (
	"context"

	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/stretchr/testify/assert"
)

func (s *CheckESTestSuite) TestCheckILMError() {

	checkES := s.monitorES.(*CheckES)

	// When check all indices
	monitoringData, err := s.monitorES.CheckILMError("_all", []string{})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check all indices with exclude
	monitoringData, err = s.monitorES.CheckILMError("_all", []string{"foo"})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check only one indice
	checkES.client.API.Indices.Create(
		"bar",
		checkES.client.API.Indices.Create.WithContext(context.Background()),
	)
	monitoringData, err = s.monitorES.CheckILMError("bar", []string{})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check indice that not exist
	monitoringData, err = s.monitorES.CheckILMError("foo", []string{})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_UNKNOWN, monitoringData.Status())

}

func (s *CheckESTestSuite) TestCheckILMStatus() {

	checkES := s.monitorES.(*CheckES)

	// When ILM is stopped
	checkES.client.API.ILM.Stop(
		checkES.client.API.ILM.Stop.WithContext(context.Background()),
	)
	monitoringData, err := s.monitorES.CheckILMStatus()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())

	// When ILM is started
	checkES.client.API.ILM.Start(
		checkES.client.API.ILM.Start.WithContext(context.Background()),
	)
	monitoringData, err = s.monitorES.CheckILMStatus()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())
}
