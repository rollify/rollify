package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/memory"
	"github.com/rollify/rollify/internal/user"
)

var (
	// Version is set in compile time.
	Version = "dev"
)

// Run runs the main application.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
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

	// Create storage.
	var (
		roomRepo     storage.RoomRepository
		diceRollRepo storage.DiceRollRepository
		userRepo     storage.UserRepository
	)
	switch cmdCfg.StorageType {
	case StorageTypeMemory:
		roomRepo = memory.NewRoomRepository()
		diceRollRepo = memory.NewDiceRollRepository()
		userRepo = memory.NewUserRepository()
	default:
		return fmt.Errorf("storage type '%s' unknown", cmdCfg.StorageType)
	}

	// Create app services.
	diceAppService, err := dice.NewService(dice.ServiceConfig{
		DiceRollRepository: diceRollRepo,
		RoomRepository:     roomRepo,
		UserRepository:     userRepo,
		Roller:             dice.NewRandomRoller(),
		Logger:             logger,
	})
	if err != nil {
		return fmt.Errorf("could not create dice application service: %w", err)
	}

	roomAppService, err := room.NewService(room.ServiceConfig{
		RoomRepository: roomRepo,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("could not create room application service: %w", err)
	}

	userAppService, err := user.NewService(user.ServiceConfig{
		UserRepository: userRepo,
		RoomRepository: roomRepo,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("could not create user application service: %w", err)
	}

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
			DiceAppService: diceAppService,
			RoomAppService: roomAppService,
			UserAppService: userAppService,
			Logger:         logger,
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

func main() {
	ctx := context.Background()

	err := Run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
