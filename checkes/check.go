package checkes

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/disaster37/go-nagios"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CheckES is implementation of MonitorES
type CheckES struct {
	client *elastic.Client
}

// MonitorES is interface of elasticsearch monitoring
type MonitorES interface {
	CheckILMError(indiceName string, excludeIndices []string) (*nagiosPlugin.Monitoring, error)
	CheckILMStatus() (*nagiosPlugin.Monitoring, error)
	CheckSLMError(snapshotRepositoryName string) (*nagiosPlugin.Monitoring, error)
	CheckSLMStatus() (*nagiosPlugin.Monitoring, error)
	CheckIndiceLocked(indiceName string) (*nagiosPlugin.Monitoring, error)
}

func manageElasticsearchGlobalParameters(c *cli.Context) (MonitorES, error) {

	if c.String("url") == "" {
		return nil, errors.New("You must set --url parameter")
	}

	return NewCheckES(c.String("url"), c.String("user"), c.String("password"), c.Bool("self-signed-certificate"))

}

//NewCheckES permit to initialize connexion on Elasticsearch cluster
func NewCheckES(URL string, username string, password string, disableTLSVerification bool) (MonitorES, error) {

	if URL == "" {
		return nil, errors.New("URL can't be empty")
	}
	log.Debugf("URL: %s", URL)
	log.Debugf("User: %s", username)
	log.Debugf("Password: xxx")
	checkES := &CheckES{}

	cfg := elastic.Config{
		Addresses: []string{URL},
	}
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}
	if disableTLSVerification {
		cfg.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	res, err := client.API.Info(
		client.API.Info.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.Errorf("Error when connecting on Elasticsearch: %s", res.String())
	}

	checkES.client = client
	return checkES, nil
}
