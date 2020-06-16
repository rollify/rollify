package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	// StorageTypeMemory is the memory storage type.
	StorageTypeMemory = "memory"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Development        bool
	Debug              bool
	APIListenAddr      string
	InternalListenAddr string
	MetricsPath        string
	HealthCheckPath    string
	PprofPath          string
	StorageType        string
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
	app.Flag("health-check-path", "the path where the health check will be served.").Default("/status").StringVar(&c.HealthCheckPath)
	app.Flag("pprof-path", "the path where the pprof handlers will be served.").Default("/debug/pprof").StringVar(&c.PprofPath)
	app.Flag("storage-type", "the storage type used on the application.").Default(StorageTypeMemory).EnumVar(&c.StorageType, StorageTypeMemory)

	_, err := app.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
