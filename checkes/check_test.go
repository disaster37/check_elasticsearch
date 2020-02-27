package checkes

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type CheckESTestSuite struct {
	suite.Suite
	monitorES MonitorES
}

func (s *CheckESTestSuite) SetupSuite() {

	// Init logger
	logrus.SetFormatter(new(prefixed.TextFormatter))
	logrus.SetLevel(logrus.DebugLevel)

	// Init client
	url := os.Getenv("ELASTICSEARCH_URL")
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")

	monitorES, err := NewCheckES(url, username, password, false)
	if err != nil {
		panic(err)
	}

	s.monitorES = monitorES

}

func (s *CheckESTestSuite) SetupTest() {

	// Do somethink before each test

}

func TestCheckESTestSuite(t *testing.T) {
	suite.Run(t, new(CheckESTestSuite))
}
