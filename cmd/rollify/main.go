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

	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/log"
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

	// Prepare our run entrypoints.
	var g run.Group

	// Serving API HTTP server.
	{
		logger := logger.WithKV(log.KV{
			"addr":  cmdCfg.APIListenAddr,
			"apiv1": true,
		})

		// API.
		apiv1Handler, err := apiv1.New(apiv1.Config{Logger: logger})
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
			"addr":    cmdCfg.InternalListenAddr,
			"metrics": true,
			"pprof":   true,
		})
		mux := http.NewServeMux()

		// Metrics.
		mux.Handle(cmdCfg.MetricsPath, promhttp.Handler())

		// Pprof.
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

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
