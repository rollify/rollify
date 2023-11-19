package main

import (
	"time"

	"github.com/alecthomas/kingpin/v2"
)

const (
	// StorageTypeMemory is the memory storage type.
	StorageTypeMemory = "memory"
	// StorageTypeMySQL is the mysql storage type.
	StorageTypeMySQL = "mysql"
	// EventSubsTypeMemory is the memory event subscription type.
	EventSubsTypeMemory = "memory"
	// EventSubsNATS is the NATS event subscription type.
	EventSubsNATS = "nats"
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
	MySQL              struct {
		Username        string
		Password        string
		Dial            string
		Address         string
		Database        string
		Params          string
		ConnMaxLifetime time.Duration
		MaxIdleConns    int
		MaxOpenConns    int
		OpTimeout       time.Duration
	}
	EventSubsType string
	NATS          struct {
		Username string
		Password string
		Address  string
	}
}

// NewCmdConfig returns a new command configuration.
func NewCmdConfig(args []string) (*CmdConfig, error) {
	c := &CmdConfig{}
	app := kingpin.New("rollify", "Online dice rolling application.")
	app.DefaultEnvars()
	app.Version(Version)

	// General.
	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)
	app.Flag("development", "Enable development mode.").BoolVar(&c.Development)

	// API.
	app.Flag("api-listen-address", "the address where the HTTP API server will be listening.").Default(":8080").StringVar(&c.APIListenAddr)

	// Internal.
	app.Flag("internal-listen-address", "the address where the HTTP internal data (metrics, pprof...) server will be listening.").Default(":8081").StringVar(&c.InternalListenAddr)
	app.Flag("metrics-path", "the path where Prometehus metrics will be served.").Default("/metrics").StringVar(&c.MetricsPath)
	app.Flag("health-check-path", "the path where the health check will be served.").Default("/status").StringVar(&c.HealthCheckPath)
	app.Flag("pprof-path", "the path where the pprof handlers will be served.").Default("/debug/pprof").StringVar(&c.PprofPath)

	// Repository.
	app.Flag("storage-type", "the storage type used on the application.").Default(StorageTypeMemory).EnumVar(&c.StorageType, StorageTypeMemory, StorageTypeMySQL)
	app.Flag("mysql.username", "the username for MySQL connection.").StringVar(&c.MySQL.Username)
	app.Flag("mysql.password", "the password for MySQL connection.").StringVar(&c.MySQL.Password)
	app.Flag("mysql.database", "the database for MySQL connection.").StringVar(&c.MySQL.Database)
	app.Flag("mysql.protocol", "the protocol for MySQL connection.").Default("tcp").StringVar(&c.MySQL.Dial)
	app.Flag("mysql.address", "the address for MySQL connection.").Default("localhost:3306").StringVar(&c.MySQL.Address)
	app.Flag("mysql.params", "the extra params for MySQL connection.").Default(`charset=utf8mb4&parseTime=True&loc=UTC`).StringVar(&c.MySQL.Params) // https://github.com/go-sql-driver/mysql#parameters
	app.Flag("mysql.conn-max-lifetime", "the max life time for MySQL connection.").Default("5m").DurationVar(&c.MySQL.ConnMaxLifetime)
	app.Flag("mysql.max-idle-conns", "the max iddle connections for MySQL.").Default("20").IntVar(&c.MySQL.MaxIdleConns)
	app.Flag("mysql.max-open-conns", "the max open connections for MySQL.").Default("25").IntVar(&c.MySQL.MaxOpenConns)
	app.Flag("mysql.operations-timeout", "timeout duration for MySQL operations.").Default("1s").DurationVar(&c.MySQL.OpTimeout)

	// Event subscription.
	app.Flag("event-subscription-type", "the event subscription type used on the application.").Default(EventSubsTypeMemory).EnumVar(&c.EventSubsType, EventSubsTypeMemory, EventSubsNATS)
	app.Flag("nats.username", "the username for NATS connection.").StringVar(&c.NATS.Username)
	app.Flag("nats.password", "the password for NATS connection.").StringVar(&c.NATS.Password)
	app.Flag("nats.address", "the address for NATS connection.").Default("localhost:4222").StringVar(&c.NATS.Address)

	_, err := app.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
