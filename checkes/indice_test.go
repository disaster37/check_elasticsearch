package checkes

import (
	"context"
	"strings"

	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/stretchr/testify/assert"
)

func (s *CheckESTestSuite) TestCheckIndiceLocked() {

	checkES := s.monitorES.(*CheckES)

	// Create index with lock settings
	checkES.client.API.Indices.Create(
		"lock",
		checkES.client.API.Indices.Create.WithContext(context.Background()),
		checkES.client.API.Indices.Create.WithBody(strings.NewReader(`
			{
				"settings": {
					"index": {
						"blocks": {
							"read_only_allow_delete": true
						}
					}
				}
			}
		`)),
	)

	// When check all indices
	monitoringData, err := s.monitorES.CheckIndiceLocked("_all")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())

	// When check only one indice
	checkES.client.API.Indices.Create(
		"bar",
		checkES.client.API.Indices.Create.WithContext(context.Background()),
	)
	monitoringData, err = s.monitorES.CheckIndiceLocked("bar")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check indice that not exist
	monitoringData, err = s.monitorES.CheckIndiceLocked("foo")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_UNKNOWN, monitoringData.Status())

	// When indice is locked and only one indice
	monitoringData, err = s.monitorES.CheckIndiceLocked("lock")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())
}
