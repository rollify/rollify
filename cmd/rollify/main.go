package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nats-io/nats.go"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/event"
	eventmemory "github.com/rollify/rollify/internal/event/memory"
	eventnats "github.com/rollify/rollify/internal/event/nats"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/log"
	metrics "github.com/rollify/rollify/internal/metrics/prometheus"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/storage"
	storagememory "github.com/rollify/rollify/internal/storage/memory"
	"github.com/rollify/rollify/internal/storage/mysql"
	"github.com/rollify/rollify/internal/user"
)

var (
	// Version is set in compile time.
	Version = "dev"
)

// Run runs the main application.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	// Ensure our context will end if any of the func uses as the main context.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Load command flags and arguments.
	cmdCfg, err := NewCmdConfig(args)
	if err != nil {
		return fmt.Errorf("could not load command configuration: %w", err)
	}

	// Set up logger.
	logrusLog := logrus.New()
	logrusLog.Out = stderr // By default logger goes to stderr (so it can split stdout prints).
	logrusLogEntry := logrus.NewEntry(logrusLog)
	if cmdCfg.Debug {
		logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	}
	if !cmdCfg.Development {
		logrusLogEntry.Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger := log.NewLogrus(logrusLogEntry).WithKV(log.KV{
		"app":     "rollify",
		"version": Version,
	})

	// Set up metrics with default metrics recorder.
	metricsRecorder := metrics.NewRecorder(prometheus.DefaultRegisterer)

	// Create storage.
	var (
		roomRepo     storage.RoomRepository
		diceRollRepo storage.DiceRollRepository
		userRepo     storage.UserRepository
	)
	switch cmdCfg.StorageType {
	// Memory storage.
	case StorageTypeMemory:
		diceRollRepo = storagememory.NewDiceRollRepository()
		roomRepo = storagememory.NewRoomRepository()
		userRepo = storagememory.NewUserRepository()

	// MySQL storage.
	case StorageTypeMySQL:
		db, err := createMySQLConnection(*cmdCfg)
		if err != nil {
			return fmt.Errorf("could not create mysql connection: %w", err)
		}

		roomRepo, err = mysql.NewRoomRepository(mysql.RoomRepositoryConfig{
			DBClient: db,
			Logger:   logger,
		})
		if err != nil {
			return fmt.Errorf("could not create mysql room repository: %w", err)
		}

		userRepo, err = mysql.NewUserRepository(mysql.UserRepositoryConfig{
			DBClient: db,
			Logger:   logger,
		})
		if err != nil {
			return fmt.Errorf("could not create mysql user repository: %w", err)
		}

		diceRollRepo, err = mysql.NewDiceRollRepository(mysql.DiceRollRepositoryConfig{
			DBClient: db,
			Logger:   logger,
		})
		if err != nil {
			return fmt.Errorf("could not create mysql dice roll repository: %w", err)
		}

	// Unsuported storage type.
	default:
		return fmt.Errorf("storage type '%s' unknown", cmdCfg.StorageType)
	}

	// Wrap repositories.
	diceRollRepo = storage.NewMeasuredDiceRollRepository(cmdCfg.StorageType, metricsRecorder,
		storage.NewTimeoutDiceRollRepository(cmdCfg.MySQL.OpTimeout, diceRollRepo))
	roomRepo = storage.NewMeasuredRoomRepository(cmdCfg.StorageType, metricsRecorder,
		storage.NewTimeoutRoomRepository(cmdCfg.MySQL.OpTimeout, roomRepo))
	userRepo = storage.NewMeasuredUserRepository(cmdCfg.StorageType, metricsRecorder,
		storage.NewTimeoutUserRepository(cmdCfg.MySQL.OpTimeout, userRepo))

	// Roller.
	roller := dice.NewRandomRoller()
	roller = dice.NewMeasureRoller("random", metricsRecorder, roller)

	// Events.
	var notifier event.Notifier
	var subscriber event.Subscriber

	switch cmdCfg.EventSubsType {
	// Memory event subscriber.
	case EventSubsTypeMemory:
		hub := eventmemory.NewHub(logger)
		notifier = hub
		subscriber = hub

	// NATS event subscriber.
	case EventSubsNATS:
		natsConn, err := createNATSConnection(*cmdCfg)
		if err != nil {
			return fmt.Errorf("could not create NATS connnection: %w", err)
		}

		hub, err := eventnats.NewHub(eventnats.HubConfig{
			Ctx:        ctx,
			NATSClient: natsConn,
			Logger:     logger,
		})
		if err != nil {
			return fmt.Errorf("could not create a NATS event hub: %w", err)
		}
		notifier = hub
		subscriber = hub

	// Unsuported event subscriber type.
	default:
		return fmt.Errorf("event subscriber type '%s' unknown", cmdCfg.EventSubsType)
	}

	notifier = event.NewMeasuredNotifier(cmdCfg.EventSubsType, metricsRecorder, notifier)
	subscriber = event.NewMeasuredSubscriber(cmdCfg.EventSubsType, metricsRecorder, subscriber)

	// Create app services.
	diceAppService, err := dice.NewService(dice.ServiceConfig{
		DiceRollRepository: diceRollRepo,
		RoomRepository:     roomRepo,
		UserRepository:     userRepo,
		Roller:             roller,
		EventNotifier:      notifier,
		EventSubscriber:    subscriber,
		Logger:             logger,
	})
	if err != nil {
		return fmt.Errorf("could not create dice application service: %w", err)
	}
	diceAppService = dice.NewMeasureService(metricsRecorder, diceAppService)

	roomAppService, err := room.NewService(room.ServiceConfig{
		RoomRepository: roomRepo,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("could not create room application service: %w", err)
	}
	roomAppService = room.NewMeasureService(metricsRecorder, roomAppService)

	userAppService, err := user.NewService(user.ServiceConfig{
		UserRepository: userRepo,
		RoomRepository: roomRepo,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("could not create user application service: %w", err)
	}
	userAppService = user.NewMeasureService(metricsRecorder, userAppService)

	// Prepare our main runner.
	var g run.Group

	// Serving API HTTP server.
	{
		logger := logger.WithKV(log.KV{
			"addr":  cmdCfg.APIListenAddr,
			"apiv1": true,
		})

		// API.
		apiv1Handler, err := apiv1.New(apiv1.Config{
			DiceAppService:  diceAppService,
			RoomAppService:  roomAppService,
			UserAppService:  userAppService,
			MetricsRecorder: metricsRecorder,
			Logger:          logger,
		})
		if err != nil {
			return fmt.Errorf("could not create apiv1 handler: %w", err)
		}

		// Create server.
		server := &http.Server{
			Addr:    cmdCfg.APIListenAddr,
			Handler: apiv1Handler,
		}

		g.Add(
			func() error {
				logger.Infof("http server listening for requests")
				return server.ListenAndServe()
			},
			func(_ error) {
				logger.Infof("http server shutdown, draining connections...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("error shutting down server: %w", err)
				}

				logger.Infof("connections drained")
			},
		)
	}

	// Serving internal HTTP server.
	{
		logger := logger.WithKV(log.KV{
			"addr":         cmdCfg.InternalListenAddr,
			"metrics":      cmdCfg.MetricsPath,
			"health-check": cmdCfg.HealthCheckPath,
			"pprof":        cmdCfg.PprofPath,
		})
		mux := http.NewServeMux()

		// Metrics.
		mux.Handle(cmdCfg.MetricsPath, promhttp.Handler())

		// Pprof.
		mux.HandleFunc(cmdCfg.PprofPath+"/", pprof.Index)
		mux.HandleFunc(cmdCfg.PprofPath+"/cmdline", pprof.Cmdline)
		mux.HandleFunc(cmdCfg.PprofPath+"/profile", pprof.Profile)
		mux.HandleFunc(cmdCfg.PprofPath+"/symbol", pprof.Symbol)
		mux.HandleFunc(cmdCfg.PprofPath+"/trace", pprof.Trace)

		// Health check.
		mux.Handle(cmdCfg.HealthCheckPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"status":"ok"}`)) }))

		// Create server.
		server := &http.Server{
			Addr:    cmdCfg.InternalListenAddr,
			Handler: mux,
		}

		g.Add(
			func() error {
				logger.Infof("http server listening for requests")
				return server.ListenAndServe()
			},
			func(_ error) {
				logger.Infof("http server shutdown, draining connections...")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("error shutting down server: %w", err)
				}

				logger.Infof("connections drained")
			},
		)
	}

	// OS signals.
	{
		sigC := make(chan os.Signal, 1)
		exitC := make(chan struct{})
		signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)

		g.Add(
			func() error {
				select {
				case s := <-sigC:
					logger.Infof("signal %s received", s)
					return nil
				case <-exitC:
					return nil
				}
			},
			func(_ error) {
				close(exitC)
			},
		)
	}

	err = g.Run()
	if err != nil {
		return err
	}

	return nil
}

func createMySQLConnection(cfg CmdConfig) (*sql.DB, error) {
	conn := fmt.Sprintf("%s:%s@%s(%s)/%s?%s",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Dial,
		cfg.MySQL.Address,
		cfg.MySQL.Database,
		cfg.MySQL.Params,
	)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(cfg.MySQL.ConnMaxLifetime)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	return db, nil
}

func createNATSConnection(cfg CmdConfig) (*nats.Conn, error) {
	c, err := nats.Connect(cfg.NATS.Address, nats.UserInfo(cfg.NATS.Username, cfg.NATS.Password))
	if err != nil {
		return nil, err
	}

	return c, nil
}

func main() {
	ctx := context.Background()

	err := Run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
