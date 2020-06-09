package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Development        bool
	Debug              bool
	APIListenAddr      string
	InternalListenAddr string
	MetricsPath        string
}

// NewCmdConfig returns a new command configuration.
func NewCmdConfig(args []string) (*CmdConfig, error) {
	c := &CmdConfig{}
	app := kingpin.New("rollify", "Online dice rolling application.")
	app.Version(Version)

	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)
	app.Flag("development", "Enable development mode.").BoolVar(&c.Development)
	app.Flag("api-listen-address", "the address where the HTTP API server will be listening.").Default(":8080").StringVar(&c.APIListenAddr)
	app.Flag("internal-listen-address", "the address where the HTTP internal data (metrics, pprof...) server will be listening.").Default(":8081").StringVar(&c.InternalListenAddr)
	app.Flag("metrics-path", "the path where Prometehus metrics will be served.").Default("/metrics").StringVar(&c.MetricsPath)

	_, err := app.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
