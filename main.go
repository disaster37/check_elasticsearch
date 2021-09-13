package main

import (
	"os"
	"sort"

	"github.com/disaster37/check_elasticsearch/v7/checkes"
	"github.com/disaster37/go-nagios"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func run(args []string) error {

	// Logger setting
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true
	formatter.ForceFormatting = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Check Elasticsearch"
	app.Version = "develop"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "url",
			Usage:   "The Elasticsearch URL",
			EnvVars: []string{"ELASTICSEARCH_URL"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "user",
			Usage:   "The Elasticsearch user",
			EnvVars: []string{"ELASTICSEARCH_USER"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "password",
			Usage:   "The Elasticsearch password",
			EnvVars: []string{"ELASTICSEARCH_PASSWORD"},
		}),
		&cli.BoolFlag{
			Name:  "self-signed-certificate",
			Usage: "Disable the TLS certificate check",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Display debug output",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:     "check-ilm-indice",
			Usage:    "Check the ILM on specific indice. Set indice _all to check all ILM policies",
			Category: "ILM",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "indice",
					Usage: "The indice name",
				},
				&cli.StringSliceFlag{
					Name:  "exclude",
					Usage: "The indice name to exclude",
				},
			},
			Action: checkes.CheckILMError,
		},
		{
			Name:     "check-ilm-status",
			Usage:    "Check that ILM is running",
			Category: "ILM",
			Action:   checkes.CheckILMStatus,
		},
		{
			Name:     "check-repository-snapshot",
			Usage:    "Check snapshots state on repository",
			Category: "SLM",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "repository",
					Usage: "The repisitory name",
				},
			},
			Action: checkes.CheckSLMError,
		},
		{
			Name:     "check-slm-status",
			Usage:    "Check that SLM service is running",
			Category: "SLM",
			Action:   checkes.CheckSLMStatus,
		},
		{
			Name:     "check-indice-locked",
			Usage:    "Check if there are indice locked. You can use _all as indice name to check all indices",
			Category: "Indice",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "indice",
					Usage: "The indice name",
				},
			},
			Action: checkes.CheckIndiceLocked,
		},
		{
			Name:     "check-transform",
			Usage:    "Check that Transform have not in error state",
			Category: "Transform",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "The transform id or empty for check all transform",
				},
				&cli.StringSliceFlag{
					Name:  "exclude",
					Usage: "The transform id to exclude",
				},
			},
			Action: checkes.CheckTransformError,
		},
	}

	app.Before = func(c *cli.Context) error {

		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if c.String("config") != "" {
			before := altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
			return before(c)
		}
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(args)
	return err
}

func main() {
	err := run(os.Args)
	if err != nil {
		monitoringData := nagiosPlugin.NewMonitoring()
		monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
		monitoringData.AddMessage("Error appear during check: %s", err)
		monitoringData.ToSdtOut()
	}
}
